package log

import (
	"context"
	"log/slog"
)

type contextKey string

type ContextKey = contextKey

const (
	environmentIDKey   contextKey = "environment_id"
	environmentNameKey contextKey = "environment_name"
	prNumberKey        contextKey = "pr_number"
	repositoryKey      contextKey = "repository"
	requestIDKey       contextKey = "request_id"
	providerKey        contextKey = "provider"
	loggerKey          contextKey = "logger"

	RequestIDKey ContextKey = requestIDKey
)

func WithEnvironment(ctx context.Context, envID, envName string) context.Context {
	ctx = context.WithValue(ctx, environmentIDKey, envID)
	ctx = context.WithValue(ctx, environmentNameKey, envName)
	return ctx
}

func WithPR(ctx context.Context, repo string, prNumber int) context.Context {
	ctx = context.WithValue(ctx, repositoryKey, repo)
	ctx = context.WithValue(ctx, prNumberKey, prNumber)
	return ctx
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func WithProvider(ctx context.Context, provider string) context.Context {
	return context.WithValue(ctx, providerKey, provider)
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	// Check if there's already a logger in the context
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return enrichLogger(ctx, logger)
	}

	// Otherwise use the default logger and enrich it
	return enrichLogger(ctx, Default())
}

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

func Info(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).DebugContext(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).WarnContext(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).ErrorContext(ctx, msg, args...)
}
