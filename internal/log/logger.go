package log

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	defaultLogger *slog.Logger
	once          sync.Once
)

// New creates a new logger based on environment configuration
func New() *slog.Logger {
	level := parseLogLevel(os.Getenv("LOG_LEVEL"))

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}

	var handler slog.Handler
	if shouldUsePretty() {
		handler = NewPrettyHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// Default returns the default logger instance (singleton)
func Default() *slog.Logger {
	once.Do(func() {
		defaultLogger = New()
		slog.SetDefault(defaultLogger)
	})
	return defaultLogger
}

// SetDefault sets the default logger instance
func SetDefault(logger *slog.Logger) {
	defaultLogger = logger
	slog.SetDefault(logger)
}

// shouldUsePretty checks if pretty logging should be enabled
func shouldUsePretty() bool {
	return os.Getenv("LOG_PRETTY") == "true"
}

// parseLogLevel parses the log level from string
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Helper functions for direct logging using the default logger

// DebugContext logs at debug level with context using the default logger
func DebugContext(ctx context.Context, msg string, args ...any) {
	Default().DebugContext(ctx, msg, args...)
}

// InfoContext logs at info level with context using the default logger
func InfoContext(ctx context.Context, msg string, args ...any) {
	Default().InfoContext(ctx, msg, args...)
}

// WarnContext logs at warn level with context using the default logger
func WarnContext(ctx context.Context, msg string, args ...any) {
	Default().WarnContext(ctx, msg, args...)
}

// ErrorContext logs at error level with context using the default logger
func ErrorContext(ctx context.Context, msg string, args ...any) {
	Default().ErrorContext(ctx, msg, args...)
}

// With returns a logger with the given attributes attached
func With(args ...any) *slog.Logger {
	return Default().With(args...)
}

// WithGroup returns a logger with the given group name
func WithGroup(name string) *slog.Logger {
	return Default().WithGroup(name)
}
