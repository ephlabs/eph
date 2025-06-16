package log

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLogLevelParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"", slog.LevelInfo},
		{"invalid", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestJSONLogging(t *testing.T) {
	oldLevel := os.Getenv("LOG_LEVEL")
	oldPretty := os.Getenv("LOG_PRETTY")
	defer func() {
		os.Setenv("LOG_LEVEL", oldLevel)
		os.Setenv("LOG_PRETTY", oldPretty)
	}()

	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_PRETTY", "false")

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	logger.Info("test message", "key", "value", "number", 42)

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}
	if result["msg"] != "test message" {
		t.Errorf("Expected msg='test message', got %v", result["msg"])
	}
	if result["key"] != "value" {
		t.Errorf("Expected key='value', got %v", result["key"])
	}
	if result["number"] != float64(42) {
		t.Errorf("Expected number=42, got %v", result["number"])
	}
	if result["level"] != "INFO" {
		t.Errorf("Expected level='INFO', got %v", result["level"])
	}
}

func TestPrettyLogging(t *testing.T) {
	var buf bytes.Buffer
	handler := NewPrettyHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	// Test various log levels
	logger.Debug("debug message", "key", "value")
	logger.Info("info message", "count", 42)
	logger.Warn("warning message", "error", "something went wrong")
	logger.Error("error message", "fatal", true)

	output := buf.String()

	// Check for expected patterns
	if !strings.Contains(output, "DEBUG") {
		t.Error("Expected DEBUG in output")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Expected INFO in output")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("Expected WARN in output")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Expected ERROR in output")
	}

	// Check for color codes
	if !strings.Contains(output, "\033[") {
		t.Error("Expected ANSI color codes in output")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")
	logger.Error("error message")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines (warn and error), got %d", len(lines))
	}

	// Verify only warn and error were logged
	for _, line := range lines {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}
		level := result["level"].(string)
		if level != "WARN" && level != "ERROR" {
			t.Errorf("Unexpected log level: %s", level)
		}
	}
}

func TestContextLogging(t *testing.T) {
	// Save original logger
	originalLogger := Default()
	defer SetDefault(originalLogger)

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))
	SetDefault(logger)

	ctx := context.Background()
	ctx = WithEnvironment(ctx, "env-123", "pr-42-test")
	ctx = WithPR(ctx, "org/repo", 42)
	ctx = WithRequestID(ctx, "req-456")
	ctx = WithProvider(ctx, "kubernetes")

	Info(ctx, "test with context")

	// Parse output - take first line only
	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Fatal("Expected log output, but buffer is empty")
	}
	lines := strings.Split(output, "\n")
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify context values
	if result["environment_id"] != "env-123" {
		t.Errorf("Expected environment_id='env-123', got %v", result["environment_id"])
	}
	if result["environment_name"] != "pr-42-test" {
		t.Errorf("Expected environment_name='pr-42-test', got %v", result["environment_name"])
	}
	if result["pr_number"] != float64(42) {
		t.Errorf("Expected pr_number=42, got %v", result["pr_number"])
	}
	if result["repository"] != "org/repo" {
		t.Errorf("Expected repository='org/repo', got %v", result["repository"])
	}
	if result["request_id"] != "req-456" {
		t.Errorf("Expected request_id='req-456', got %v", result["request_id"])
	}
	if result["provider"] != "kubernetes" {
		t.Errorf("Expected provider='kubernetes', got %v", result["provider"])
	}
}

func TestLogValuerTypes(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))

	// Test Token redaction
	token := Token("super-secret-token")
	logger.Info("token test", "token", token)

	// Test APIKey partial visibility
	apiKey := APIKey("sk_test_1234567890abcdef")
	logger.Info("api key test", "api_key", apiKey)

	// Test Environment struct
	env := Environment{
		ID:       "env-123",
		Name:     "test-env",
		URL:      "https://test.example.com",
		Provider: "kubernetes",
		Token:    Token("secret-token"),
	}
	logger.Info("environment test", "env", env)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")

	// Check token redaction
	var tokenResult map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &tokenResult); err != nil {
		t.Fatalf("Failed to parse token JSON: %v", err)
	}
	if tokenResult["token"] != "[REDACTED_TOKEN]" {
		t.Errorf("Expected token to be redacted, got %v", tokenResult["token"])
	}

	// Check API key partial visibility
	var apiKeyResult map[string]interface{}
	if err := json.Unmarshal([]byte(lines[1]), &apiKeyResult); err != nil {
		t.Fatalf("Failed to parse API key JSON: %v", err)
	}
	expectedKey := "sk_t...cdef"
	if apiKeyResult["api_key"] != expectedKey {
		t.Errorf("Expected api_key='%s', got %v", expectedKey, apiKeyResult["api_key"])
	}

	// Check environment struct - token should not appear
	var envResult map[string]interface{}
	if err := json.Unmarshal([]byte(lines[2]), &envResult); err != nil {
		t.Fatalf("Failed to parse environment JSON: %v", err)
	}
	envMap := envResult["env"].(map[string]interface{})
	if envMap["id"] != "env-123" {
		t.Errorf("Expected env.id='env-123', got %v", envMap["id"])
	}
	if envMap["name"] != "test-env" {
		t.Errorf("Expected env.name='test-env', got %v", envMap["name"])
	}
	if _, exists := envMap["token"]; exists {
		t.Error("Environment token should not be logged")
	}
}

func TestConcurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			logger.Info("concurrent log", "goroutine", n, "timestamp", time.Now().UnixNano())
		}(i)
	}
	wg.Wait()

	// Verify all logs were written
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 100 {
		t.Errorf("Expected 100 log lines, got %d", len(lines))
	}

	// Verify each line is valid JSON
	seen := make(map[int]bool)
	for _, line := range lines {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			t.Errorf("Failed to parse JSON: %v", err)
		}
		if goroutine, ok := result["goroutine"].(float64); ok {
			seen[int(goroutine)] = true
		}
	}

	// Verify all goroutines logged
	if len(seen) != 100 {
		t.Errorf("Expected logs from 100 goroutines, got %d", len(seen))
	}
}

func TestWithHelpers(t *testing.T) {
	var buf bytes.Buffer
	baseLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))
	SetDefault(baseLogger)

	// Test With helper
	logger := With("service", "ephd", "version", "1.0.0")
	logger.Info("with test")

	// Test WithGroup helper
	groupLogger := WithGroup("request")
	groupLogger.Info("group test", "method", "GET", "path", "/api/v1/environments")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")

	// Check With attributes
	var withResult map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &withResult); err != nil {
		t.Fatalf("Failed to parse with result JSON: %v", err)
	}
	if withResult["service"] != "ephd" {
		t.Errorf("Expected service='ephd', got %v", withResult["service"])
	}
	if withResult["version"] != "1.0.0" {
		t.Errorf("Expected version='1.0.0', got %v", withResult["version"])
	}

	// Check WithGroup
	var groupResult map[string]interface{}
	if err := json.Unmarshal([]byte(lines[1]), &groupResult); err != nil {
		t.Fatalf("Failed to parse group result JSON: %v", err)
	}
	if request, ok := groupResult["request"].(map[string]interface{}); ok {
		if request["method"] != "GET" {
			t.Errorf("Expected request.method='GET', got %v", request["method"])
		}
		if request["path"] != "/api/v1/environments" {
			t.Errorf("Expected request.path='/api/v1/environments', got %v", request["path"])
		}
	} else {
		t.Error("Expected 'request' group in output")
	}
}

func TestPrettyHandlerAttributes(t *testing.T) {
	var buf bytes.Buffer
	handler := NewPrettyHandler(&buf, &slog.HandlerOptions{})
	logger := slog.New(handler)

	// Test with various attribute types
	logger.With("service", "ephd").Info("test message",
		"string", "value",
		"int", 42,
		"float", 3.14,
		"bool_true", true,
		"bool_false", false,
		"duration", 5*time.Second,
		slog.Group("nested",
			"key1", "value1",
			"key2", 2,
		),
	)

	output := buf.String()

	// Strip ANSI color codes for easier testing
	stripAnsi := func(str string) string {
		// Simple ANSI code removal - remove \033[...m sequences
		result := str
		for {
			start := strings.Index(result, "\033[")
			if start == -1 {
				break
			}
			end := strings.Index(result[start:], "m")
			if end == -1 {
				break
			}
			result = result[:start] + result[start+end+1:]
		}
		return result
	}

	cleanOutput := stripAnsi(output)

	// Check for expected content
	if !strings.Contains(cleanOutput, "service=") {
		t.Error("Expected 'service=' in output")
	}
	if !strings.Contains(cleanOutput, `string="value"`) {
		t.Error("Expected quoted string value in output")
	}
	if !strings.Contains(cleanOutput, "int=42") {
		t.Error("Expected 'int=42' in output")
	}
	if !strings.Contains(cleanOutput, "float=3.14") {
		t.Error("Expected 'float=3.14' in output")
	}
	if !strings.Contains(cleanOutput, "duration=5s") {
		t.Error("Expected 'duration=5s' in output")
	}
	if !strings.Contains(cleanOutput, "nested.key1") {
		t.Error("Expected nested group attributes in output")
	}
}
