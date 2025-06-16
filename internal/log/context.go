package log

import (
	"context"
	"log/slog"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

// ContextKey represents a context key type for external use
type ContextKey = contextKey

// Context keys for common fields
const (
	environmentIDKey   contextKey = "environment_id"
	environmentNameKey contextKey = "environment_name"
	prNumberKey        contextKey = "pr_number"
	repositoryKey      contextKey = "repository"
	requestIDKey       contextKey = "request_id"
	providerKey        contextKey = "provider"
	loggerKey          contextKey = "logger"

	// Exported context keys for external use
	RequestIDKey ContextKey = requestIDKey
)

// WithEnvironment adds environment information to the context
func WithEnvironment(ctx context.Context, envID, envName string) context.Context {
	ctx = context.WithValue(ctx, environmentIDKey, envID)
	ctx = context.WithValue(ctx, environmentNameKey, envName)
	return ctx
}

// WithPR adds pull request information to the context
func WithPR(ctx context.Context, repo string, prNumber int) context.Context {
	ctx = context.WithValue(ctx, repositoryKey, repo)
	ctx = context.WithValue(ctx, prNumberKey, prNumber)
	return ctx
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithProvider adds the infrastructure provider name to the context
func WithProvider(ctx context.Context, provider string) context.Context {
	return context.WithValue(ctx, providerKey, provider)
}

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext extracts the logger from context, creating one with context values if needed
func FromContext(ctx context.Context) *slog.Logger {
	// Check if there's already a logger in the context
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return enrichLogger(ctx, logger)
	}

	// Otherwise use the default logger and enrich it
	return enrichLogger(ctx, Default())
}

// enrichLogger adds context values to the logger
func enrichLogger(ctx context.Context, logger *slog.Logger) *slog.Logger {
	var attrs []any

	// Extract all known context values
	if envID, ok := ctx.Value(environmentIDKey).(string); ok && envID != "" {
		attrs = append(attrs, "environment_id", envID)
	}

	if envName, ok := ctx.Value(environmentNameKey).(string); ok && envName != "" {
		attrs = append(attrs, "environment_name", envName)
	}

	if prNumber, ok := ctx.Value(prNumberKey).(int); ok && prNumber > 0 {
		attrs = append(attrs, "pr_number", prNumber)
	}

	if repo, ok := ctx.Value(repositoryKey).(string); ok && repo != "" {
		attrs = append(attrs, "repository", repo)
	}

	if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID != "" {
		attrs = append(attrs, "request_id", requestID)
	}

	if provider, ok := ctx.Value(providerKey).(string); ok && provider != "" {
		attrs = append(attrs, "provider", provider)
	}

	// Only create a new logger if we have attributes to add
	if len(attrs) > 0 {
		return logger.With(attrs...)
	}

	return logger
}

// Helper functions that extract logger from context and log

// Info logs at info level using logger from context
func Info(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

// Debug logs at debug level using logger from context
func Debug(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).DebugContext(ctx, msg, args...)
}

// Warn logs at warn level using logger from context
func Warn(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).WarnContext(ctx, msg, args...)
}

// Error logs at error level using logger from context
func Error(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).ErrorContext(ctx, msg, args...)
}
