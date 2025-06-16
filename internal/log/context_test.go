package log

import (
	"context"
	"log/slog"
	"testing"
)

func TestWithEnvironment(t *testing.T) {
	ctx := context.Background()
	ctx = WithEnvironment(ctx, "env-123", "test-env")

	logger := FromContext(ctx)
	if logger == nil {
		t.Error("Expected logger from context")
	}
}

func TestWithPR(t *testing.T) {
	ctx := context.Background()
	ctx = WithPR(ctx, "owner/repo", 42)

	logger := FromContext(ctx)
	if logger == nil {
		t.Error("Expected logger from context")
	}
}

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")

	logger := FromContext(ctx)
	if logger == nil {
		t.Error("Expected logger from context")
	}
}

func TestWithProvider(t *testing.T) {
	ctx := context.Background()
	ctx = WithProvider(ctx, "kubernetes")

	logger := FromContext(ctx)
	if logger == nil {
		t.Error("Expected logger from context")
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
	ctx := context.Background()
	ctx = WithEnvironment(ctx, "env-123", "test-env")
	ctx = WithPR(ctx, "owner/repo", 42)
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithProvider(ctx, "kubernetes")

	logger := FromContext(ctx)
	if logger == nil {
		t.Error("Expected enriched logger from context")
	}
}

func TestFromContext_WithEmptyValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithEnvironment(ctx, "", "")
	ctx = WithPR(ctx, "", 0)
	ctx = WithRequestID(ctx, "")
	ctx = WithProvider(ctx, "")

	logger := FromContext(ctx)
	if logger == nil {
		t.Error("Expected logger from context")
	}
}

func TestContextLogFunctions(_ *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "test-req")

	Info(ctx, "test info message")
	Debug(ctx, "test debug message")
	Warn(ctx, "test warn message")
	Error(ctx, "test error message")
}
