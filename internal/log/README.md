# Eph Logging Package

The `internal/log` package provides structured logging for the Eph project using Go's standard `slog` package. It supports both JSON logging for production and pretty-printed colored output for development.

## Features

- 🔧 **Environment-based configuration** - JSON by default, pretty printing via `LOG_PRETTY=true`
- 🎨 **Pretty console output** - Colored, human-readable logs for development
- 🔒 **Automatic redaction** - Sensitive data is automatically redacted using `slog.LogValuer`
- 🏷️ **Context propagation** - Automatic inclusion of request IDs, environment IDs, etc.
- ⚡ **High performance** - Minimal allocations, built on Go's efficient `slog`
- 🧵 **Thread-safe** - Safe for concurrent use

## Quick Start

```go
import "github.com/ephlabs/eph/internal/log"

// Use the default logger
log.Info("Server started", "port", 8080)

// With context propagation
ctx := log.WithRequestID(context.Background(), "req-123")
log.Info(ctx, "Processing request", "method", "GET", "path", "/api/v1/status")

// Structured logging with groups
logger := log.Default().WithGroup("database")
logger.Info("Connected to database", "host", "localhost", "port", 5432)
```

## Configuration

Configure logging behavior via environment variables:

```bash
# Set log level (debug, info, warn, error)
export LOG_LEVEL=info

# Enable pretty printing for development
export LOG_PRETTY=true
```

## Logging Patterns

### Basic Logging

```go
// Simple messages
log.Info("User logged in", "user_id", "123")
log.Error("Failed to connect", "error", err)

// With custom logger
logger := log.New()
logger.Debug("Detailed information", "data", complexStruct)
```

### Context-Aware Logging

The package provides helpers to attach common fields via context:

```go
// Add environment context
ctx := log.WithEnvironment(ctx, "env-123", "pr-42-staging")

// Add PR information
ctx = log.WithPR(ctx, "org/repo", 42)

// Add request ID (typically in middleware)
ctx = log.WithRequestID(ctx, uuid.New().String())

// Log with context - fields are automatically included
log.Info(ctx, "Environment created")
// Output includes: environment_id="env-123" environment_name="pr-42-staging" pr_number=42
```

### Sensitive Data Handling

Use the provided types for automatic redaction:

```go
// Define sensitive fields
type Config struct {
    APIKey log.APIKey `json:"api_key"`
    Token  log.Token  `json:"-"`
}

config := Config{
    APIKey: log.APIKey("sk_test_1234567890abcdef"),
    Token:  log.Token("secret-auth-token"),
}

log.Info("Config loaded", "config", config)
// Output:
// JSON: {"msg":"Config loaded","config":{"api_key":"sk_t...cdef"}}
// Pretty: Config loaded config={api_key="sk_t...cdef"}

// Direct logging of sensitive values
log.Info("Authentication", "token", log.Token("secret-123"))
// Output: {"msg":"Authentication","token":"[REDACTED_TOKEN]"}
```

### Custom LogValuer Types

Implement `slog.LogValuer` for custom redaction:

```go
type CreditCard string

func (c CreditCard) LogValue() slog.Value {
    if len(c) >= 4 {
        return slog.StringValue("****-****-****-" + string(c[len(c)-4:]))
    }
    return slog.StringValue("[INVALID_CARD]")
}
```

## HTTP Middleware Integration

Example middleware that adds request context:

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract or generate request ID
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        // Add to context
        ctx := log.WithRequestID(r.Context(), requestID)

        // Log request
        log.Info(ctx, "HTTP request",
            "method", r.Method,
            "path", r.URL.Path,
            "remote_addr", r.RemoteAddr)

        // Continue with enriched context
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Output Examples

### JSON Format (Production)

```json
{
  "time": "2024-01-15T10:30:45.123Z",
  "level": "INFO",
  "msg": "Environment created",
  "environment_id": "env-123",
  "environment_name": "pr-42-staging",
  "pr_number": 42,
  "request_id": "req-456",
  "duration": 1.234
}
```

### Pretty Format (Development)

```
[10:30:45.123] INFO  Environment created environment_id="env-123" environment_name="pr-42-staging" pr_number=42 request_id="req-456" duration=1.234s
```

With colors:
- Timestamp: gray
- Level: green (INFO), yellow (WARN), red (ERROR), gray (DEBUG)
- Keys: cyan
- Values: white (strings), yellow (numbers), green/red (booleans)

## Testing

The package includes comprehensive tests:

```bash
# Run tests
go test ./internal/log/...

# With coverage
go test -cover ./internal/log/...

# Verbose output
go test -v ./internal/log/...
```

## Best Practices

1. **Use context propagation** - Pass context through your call stack for consistent logging
2. **Structure your logs** - Use key-value pairs instead of formatted strings
3. **Use appropriate levels** - Debug for development, Info for notable events, Warn for recoverable issues, Error for failures
4. **Group related fields** - Use `WithGroup` for logical grouping
5. **Avoid logging sensitive data** - Use the provided redaction types or implement `LogValuer`

## Performance Considerations

- Attributes are lazily evaluated - expensive operations in log values are only executed if the log level is enabled
- Use `Debug` level sparingly in production
- The pretty handler adds overhead - use JSON in production
- Context propagation has minimal overhead - values are only extracted when logging

## Integration with Eph

This logging package is designed to integrate seamlessly with Eph's architecture:

- **Environment tracking** - Logs automatically include environment context
- **Request correlation** - HTTP requests are tracked with unique IDs
- **Provider operations** - Infrastructure operations include provider context
- **Webhook processing** - PR events are logged with repository and PR number

## Migration from Standard Log

Replace standard library usage:

```go
// Before
log.Printf("Starting server on port %d", port)

// After
log.Info("Starting server", "port", port)

// Before
log.Fatal("Failed to start:", err)

// After
log.Error("Failed to start", "error", err)
os.Exit(1)
```
