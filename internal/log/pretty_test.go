package log

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"
)

func TestPrettyHandler_DifferentLevels(t *testing.T) {
	var buf bytes.Buffer
	handler := NewPrettyHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	if !bytes.Contains(buf.Bytes(), []byte("DEBUG")) {
		t.Error("Debug level not found in output")
	}
	if !bytes.Contains(buf.Bytes(), []byte("INFO")) {
		t.Error("Info level not found in output")
	}
	if !bytes.Contains(buf.Bytes(), []byte("WARN")) {
		t.Error("Warn level not found in output")
	}
	if !bytes.Contains(buf.Bytes(), []byte("ERROR")) {
		t.Error("Error level not found in output")
	}
}

func TestPrettyHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	handler := NewPrettyHandler(&buf, nil)
	logger := slog.New(handler).With("key1", "value1")

	logger.Info("test message", "key2", "value2")

	if !bytes.Contains(buf.Bytes(), []byte("key1")) {
		t.Error("Pre-attached attribute not found")
	}
	if !bytes.Contains(buf.Bytes(), []byte("value1")) {
		t.Error("Pre-attached value not found")
	}
	if !bytes.Contains(buf.Bytes(), []byte("key2")) {
		t.Error("Record attribute not found")
	}
	if !bytes.Contains(buf.Bytes(), []byte("value2")) {
		t.Error("Record value not found")
	}
}

func TestPrettyHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	handler := NewPrettyHandler(&buf, nil)
	logger := slog.New(handler).WithGroup("group1")

	logger.Info("test message", "key", "value")

	if !bytes.Contains(buf.Bytes(), []byte("group1.key")) {
		t.Error("Grouped attribute not found")
	}
}

func TestPrettyHandler_WithSource(t *testing.T) {
	var buf bytes.Buffer
	handler := NewPrettyHandler(&buf, &slog.HandlerOptions{AddSource: true})
	logger := slog.New(handler)

	logger.Info("test message")

	if !bytes.Contains(buf.Bytes(), []byte(".go:")) {
		t.Error("Source information not found")
	}
}

func TestPrettyHandler_DifferentValueTypes(t *testing.T) {
	var buf bytes.Buffer
	handler := NewPrettyHandler(&buf, nil)
	logger := slog.New(handler)

	now := time.Now()
	duration := 5 * time.Second
	err := errors.New("test error")

	logger.Info("test message",
		"string", "test",
		"int", 42,
		"float", 3.14,
		"bool_true", true,
		"bool_false", false,
		"time", now,
		"duration", duration,
		"error", err,
	)

	expectedSubstrings := []string{
		"string", "test",
		"int", "42",
		"float", "3.14",
		"bool_true", "true",
		"bool_false", "false",
		"duration", "5s",
		"error", "test error",
	}

	for _, expected := range expectedSubstrings {
		if !bytes.Contains(buf.Bytes(), []byte(expected)) {
			t.Errorf("Expected substring %q not found in output", expected)
		}
	}
}

func TestPrettyHandler_NestedGroups(t *testing.T) {
	var buf bytes.Buffer
	handler := NewPrettyHandler(&buf, nil)
	logger := slog.New(handler)

	logger.Info("test message",
		slog.Group("outer",
			slog.String("inner_key", "inner_value"),
			slog.Group("inner",
				slog.String("deep_key", "deep_value"),
			),
		),
	)

	if !bytes.Contains(buf.Bytes(), []byte("outer.inner_key")) {
		t.Error("Nested group key not found")
	}
	if !bytes.Contains(buf.Bytes(), []byte("outer.inner.deep_key")) {
		t.Error("Deeply nested group key not found")
	}
}

func TestPrettyHandler_Enabled(t *testing.T) {
	tests := []struct {
		name    string
		level   slog.Level
		check   slog.Level
		enabled bool
	}{
		{"debug handler allows debug", slog.LevelDebug, slog.LevelDebug, true},
		{"info handler blocks debug", slog.LevelInfo, slog.LevelDebug, false},
		{"info handler allows info", slog.LevelInfo, slog.LevelInfo, true},
		{"warn handler allows error", slog.LevelWarn, slog.LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewPrettyHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: tt.level})
			if handler.Enabled(context.Background(), tt.check) != tt.enabled {
				t.Errorf("Expected enabled=%v for level %v with handler level %v", tt.enabled, tt.check, tt.level)
			}
		})
	}
}

func TestPrettyHandler_NilOptions(t *testing.T) {
	var buf bytes.Buffer
	handler := NewPrettyHandler(&buf, nil)
	logger := slog.New(handler)

	logger.Info("test message")

	if buf.Len() == 0 {
		t.Error("Expected output with nil options")
	}
}
