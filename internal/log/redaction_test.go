package log

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestTokenRedactionInActualLogging(t *testing.T) {
	token := Token("super-secret-token")

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	logger.Info("user authenticated", "token", token)

	output := buf.String()
	if !strings.Contains(output, "[REDACTED_TOKEN]") {
		t.Errorf("Token not redacted in log output: %s", output)
	}
	if strings.Contains(output, "super-secret-token") {
		t.Errorf("Actual token value leaked in log output: %s", output)
	}
}

func TestAPIKeyRedactionInActualLogging(t *testing.T) {
	longKey := APIKey("abcd1234567890xyz")
	shortKey := APIKey("short")
	emptyKey := APIKey("")

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	logger.Info("api call", "long_key", longKey, "short_key", shortKey, "empty_key", emptyKey)

	output := buf.String()

	if !strings.Contains(output, "abcd...0xyz") {
		t.Errorf("Long key not properly redacted: %s", output)
	}
	if !strings.Contains(output, "[REDACTED_KEY]") {
		t.Errorf("Short key not redacted: %s", output)
	}
	if strings.Contains(output, "1234567890") {
		t.Errorf("Long key middle portion leaked: %s", output)
	}
}

func TestEnvironmentLoggingHidesToken(t *testing.T) {
	env := Environment{
		ID:       "env-123",
		Name:     "test-env",
		URL:      "https://test.example.com",
		Provider: "kubernetes",
		Token:    Token("secret-token"),
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	logger.Info("environment created", "env", env)

	output := buf.String()

	expectedFields := []string{"env-123", "test-env", "https://test.example.com", "kubernetes"}
	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Expected field %q not found in output: %s", field, output)
		}
	}

	if strings.Contains(output, "secret-token") {
		t.Errorf("Environment token leaked in log output: %s", output)
	}
}
