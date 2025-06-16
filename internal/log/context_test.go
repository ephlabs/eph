package log

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestWithEnvironment(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	ctx := context.Background()
	ctx = WithEnvironment(ctx, "env-123", "test-env")
	ctx = WithLogger(ctx, logger)

	Info(ctx, "test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logEntry["environment_id"] != "env-123" {
		t.Errorf("Expected environment_id 'env-123', got %v", logEntry["environment_id"])
	}
	if logEntry["environment_name"] != "test-env" {
		t.Errorf("Expected environment_name 'test-env', got %v", logEntry["environment_name"])
	}
}

func TestWithPR(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	ctx := context.Background()
	ctx = WithPR(ctx, "owner/repo", 42)
	ctx = WithLogger(ctx, logger)

	Info(ctx, "test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logEntry["repository"] != "owner/repo" {
		t.Errorf("Expected repository 'owner/repo', got %v", logEntry["repository"])
	}
	if logEntry["pr_number"] != float64(42) {
		t.Errorf("Expected pr_number 42, got %v", logEntry["pr_number"])
	}
}

func TestWithRequestID(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithLogger(ctx, logger)

	Info(ctx, "test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logEntry["request_id"] != "req-123" {
		t.Errorf("Expected request_id 'req-123', got %v", logEntry["request_id"])
	}
}

func TestWithProvider(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	ctx := context.Background()
	ctx = WithProvider(ctx, "kubernetes")
	ctx = WithLogger(ctx, logger)

	Info(ctx, "test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logEntry["provider"] != "kubernetes" {
		t.Errorf("Expected provider 'kubernetes', got %v", logEntry["provider"])
	}
}

func TestWithLogger(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	ctx = WithLogger(ctx, logger)

	retrieved := FromContext(ctx)
	if retrieved == nil {
		t.Error("Expected logger from context")
	}
}

func TestFromContext_WithExistingLogger(t *testing.T) {
	ctx := context.Background()
	originalLogger := slog.Default()
	ctx = WithLogger(ctx, originalLogger)

	logger := FromContext(ctx)
	if logger == nil {
		t.Error("Expected logger from context")
	}
}

func TestFromContext_WithoutLogger(t *testing.T) {
	ctx := context.Background()

	logger := FromContext(ctx)
	if logger == nil {
		t.Error("Expected default logger from context")
	}
}

func TestFromContext_WithAllContextValues(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	ctx := context.Background()
	ctx = WithEnvironment(ctx, "env-123", "test-env")
	ctx = WithPR(ctx, "owner/repo", 42)
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithProvider(ctx, "kubernetes")
	ctx = WithLogger(ctx, logger)

	Info(ctx, "test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	expected := map[string]interface{}{
		"environment_id":   "env-123",
		"environment_name": "test-env",
		"repository":       "owner/repo",
		"pr_number":        float64(42),
		"request_id":       "req-123",
		"provider":         "kubernetes",
	}

	for key, expectedValue := range expected {
		if logEntry[key] != expectedValue {
			t.Errorf("Expected %s '%v', got %v", key, expectedValue, logEntry[key])
		}
	}
}

func TestFromContext_WithEmptyValues(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	ctx := context.Background()
	ctx = WithEnvironment(ctx, "", "")
	ctx = WithPR(ctx, "", 0)
	ctx = WithRequestID(ctx, "")
	ctx = WithProvider(ctx, "")
	ctx = WithLogger(ctx, logger)

	Info(ctx, "test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Empty values should not be added to the log
	for _, key := range []string{"environment_id", "environment_name", "repository", "pr_number", "request_id", "provider"} {
		if _, exists := logEntry[key]; exists {
			t.Errorf("Expected empty value for %s to not appear in log, but it was present with value %v", key, logEntry[key])
		}
	}
}

func TestContextLogFunctions(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-req")
	ctx = WithLogger(ctx, logger)

	tests := []struct {
		name    string
		logFunc func(context.Context, string, ...any)
		level   string
		message string
	}{
		{"Info", Info, "INFO", "test info message"},
		{"Debug", Debug, "DEBUG", "test debug message"},
		{"Warn", Warn, "WARN", "test warn message"},
		{"Error", Error, "ERROR", "test error message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(ctx, tt.message)

			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}

			if logEntry["level"] != tt.level {
				t.Errorf("Expected level '%s', got %v", tt.level, logEntry["level"])
			}
			if logEntry["msg"] != tt.message {
				t.Errorf("Expected message '%s', got %v", tt.message, logEntry["msg"])
			}
			if logEntry["request_id"] != "test-req" {
				t.Errorf("Expected request_id 'test-req', got %v", logEntry["request_id"])
			}
		})
	}
}
