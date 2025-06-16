# Structured Logging Implementation Plan

## Overview

This document outlines the implementation plan for adding structured logging to the Eph project using Go's standard `slog` package. The implementation will replace the current basic `log` package usage with a comprehensive structured logging system that supports both development and production environments.

**Key Decisions**:
- The logging package is implemented in `internal/log/` since this is application-specific configuration and not intended as a reusable library for other projects.
- Simplified environment configuration: JSON logging by default, pretty printing only when explicitly enabled via `LOG_PRETTY=true`
- Leverage slog's built-in LogValuer interface for redaction instead of custom implementation
- Use slog.Handler interface for pretty printing (avoid reinventing wheels)
- Utilize slog's existing JSON/Text handlers as foundation components

## Leveraging slog's Built-in Capabilities

Rather than reinventing functionality, this implementation leverages slog's existing features:

### What slog Provides Out of the Box:
- **JSONHandler**: Production-ready structured JSON logging
- **TextHandler**: Human-readable key=value format (alternative to custom pretty handler)
- **LogValuer Interface**: Built-in mechanism for custom value representation and redaction
- **Handler Interface**: Standard interface for custom log processing
- **Context Integration**: Native support for context-aware logging
- **Performance Optimizations**: Minimal allocations, lazy evaluation support
- **Attribute Grouping**: Built-in support for nested log attributes

### What We Need to Implement:
- **Context Propagation Helpers**: Domain-specific context keys (environment_id, pr_number, etc.)
- **Pretty Handler**: Colored console output for development (using Handler interface)
- **LogValuer Types**: Specific types for our sensitive data (Token, Password, User, etc.)
- **Configuration Layer**: Environment-based handler selection and setup

### What We Don't Need to Build:
- ❌ Custom redaction system (use LogValuer)
- ❌ JSON serialization (use JSONHandler)
- ❌ Thread safety mechanisms (built into slog)
- ❌ Performance optimizations (slog handles this)
- ❌ Attribute handling (slog provides this)

## Current State Analysis

### Existing Logging Touchpoints
Based on codebase analysis, current logging occurs in:

1. **`cmd/ephd/main.go:10,13`** - Server startup and failure logging
2. **`internal/server/server.go:60,82,87,102`** - Server lifecycle and JSON encoding errors
3. **`internal/server/middleware.go:26,64`** - HTTP request logging and panic recovery

### Current Infrastructure
- **`internal/log/`** exists but only contains a README placeholder (will be updated with full implementation)
- **`internal/server/middleware.go`** has basic middleware chain with request ID generation
- **HTTP server** uses standard `net/http` with custom middleware stack
- **Dependencies** include `github.com/google/uuid` for request ID generation
- **Documentation** in various READMEs and CLAUDE.md references `internal/log/` (needs updating)

### Gaps vs Issue Requirements
The issue description appears to be comprehensive but needs these updates based on current codebase and requirements:
- Request ID middleware already exists but doesn't integrate with context logging
- Pretty handler implementation needed for development
- No context propagation system currently in place
- Simplified environment variable strategy needed (JSON by default)
- Redaction functionality missing from original issue but critical for security
- Cleanup tasks not addressed in original issue (remove internal/log, update docs)

## Implementation Strategy

### Phase 1: Core Logging Package
Create the structured logging foundation in `internal/log/`:

**Files to create/modify:**
- `internal/log/logger.go` - Core logging functionality with slog wrapper
- `internal/log/pretty.go` - Development-friendly pretty printer (custom Handler)
- `internal/log/context.go` - Context propagation helpers
- `internal/log/types.go` - LogValuer implementations for sensitive types
- `internal/log/logger_test.go` - Comprehensive test suite

**Key features:**
- JSON logging by default, pretty printing via `LOG_PRETTY=true`
- Log level configuration via `LOG_LEVEL` environment variable
- Context value extraction and propagation
- Thread-safe concurrent logging
- Redaction via slog.LogValuer interface (no custom implementation needed)

### Phase 2: Server Integration
Update existing server middleware and add request context:

**Files to modify:**
- `internal/server/middleware.go` - Enhance request ID middleware with context integration
- `internal/server/server.go` - Replace `log` calls with structured logging
- `cmd/ephd/main.go` - Update startup logging

**Key changes:**
- Request ID middleware adds to context (not just headers)
- All HTTP requests logged with structured format
- Server lifecycle events use structured logging
- JSON response errors logged with context

### Phase 3: Context Propagation System
Implement context keys and helpers for domain-specific logging:

**Context keys to implement:**
- `environment_id` / `environment_name` - For environment operations
- `pr_number` / `repository` - For PR-related operations
- `request_id` - For HTTP request tracing
- `provider` - For infrastructure provider operations

### Phase 4: Cleanup & Documentation
Remove old stubs and update documentation:

**Cleanup tasks:**
- Update `CLAUDE.md` project structure section if needed
- Update various package READMEs to reference `internal/log/` consistently
- Ensure comprehensive `internal/log/README.md` with usage examples

### Phase 5: Testing & Validation
Comprehensive testing and integration:

**Testing approach:**
- Unit tests for all logging functions including redaction
- JSON output format validation
- Pretty output format validation
- Log level filtering tests
- Context propagation tests
- Concurrent logging tests
- Redaction functionality tests

## Detailed Implementation

### 1. Core Logger Structure

```go
// internal/log/logger.go
package log

import (
    "context"
    "log/slog"
    "os"
    "strings"
)

var defaultLogger *slog.Logger

// Environment-based configuration
func New() *slog.Logger {
    level := parseLogLevel(os.Getenv("LOG_LEVEL"))

    opts := &slog.HandlerOptions{
        Level: level,
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

func shouldUsePretty() bool {
    return os.Getenv("LOG_PRETTY") == "true"
}
```

### 2. Context Integration

```go
// internal/log/context.go
type contextKey string

const (
    environmentIDKey   contextKey = "environment_id"
    environmentNameKey contextKey = "environment_name"
    prNumberKey       contextKey = "pr_number"
    repositoryKey     contextKey = "repository"
    requestIDKey      contextKey = "request_id"
    providerKey       contextKey = "provider"
)

// Context helpers
func WithEnvironment(ctx context.Context, envID, envName string) context.Context
func WithPR(ctx context.Context, repo string, prNumber int) context.Context
func WithRequestID(ctx context.Context, requestID string) context.Context
func WithProvider(ctx context.Context, provider string) context.Context

// Logger extraction
func FromContext(ctx context.Context) *slog.Logger
```

### 3. Middleware Enhancement

```go
// internal/server/middleware.go (modifications)
func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        w.Header().Set("X-Request-ID", requestID)
        ctx := log.WithRequestID(r.Context(), requestID)

        // Log incoming request
        log.Info(ctx, "HTTP request",
            "method", r.Method,
            "path", r.URL.Path,
            "remote_addr", r.RemoteAddr)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 4. Pretty Handler for Development

Custom slog.Handler implementation for colored console output:

```go
// internal/log/pretty.go
type PrettyHandler struct {
    handler slog.Handler // Wrap JSONHandler to leverage attribute processing
    out     io.Writer
    mu      sync.Mutex
}

func NewPrettyHandler(out io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
    if opts == nil {
        opts = &slog.HandlerOptions{}
    }
    return &PrettyHandler{
        handler: slog.NewJSONHandler(io.Discard, opts), // Use for attribute processing
        out:     out,
    }
}

// Implement slog.Handler interface
func (h *PrettyHandler) Enabled(ctx context.Context, level slog.Level) bool {
    return h.handler.Enabled(ctx, level)
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    return &PrettyHandler{
        handler: h.handler.WithAttrs(attrs),
        out:     h.out,
    }
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
    return &PrettyHandler{
        handler: h.handler.WithGroup(name),
        out:     h.out,
    }
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
    h.mu.Lock()
    defer h.mu.Unlock()

    // Format: [15:04:05] INFO  Message key=value key=value
    // With ANSI colors for levels and structured formatting

    level := colorizeLevel(r.Level)
    time := r.Time.Format("15:04:05")

    fmt.Fprintf(h.out, "[%s] %s %s",
        colorize(time, colorGray),
        level,
        r.Message)

    // Process attributes using the wrapped handler's logic
    var buf bytes.Buffer
    tempHandler := slog.NewJSONHandler(&buf, nil)
    if err := tempHandler.Handle(ctx, r); err != nil {
        return err
    }

    // Parse JSON and format as pretty key=value pairs
    formatAttributes(&buf, h.out)

    fmt.Fprintln(h.out)
    return nil
}

// Color constants
const (
    colorReset  = "\033[0m"
    colorRed    = "\033[31m"
    colorGreen  = "\033[32m"
    colorYellow = "\033[33m"
    colorBlue   = "\033[34m"
    colorGray   = "\033[37m"
)

func colorize(text, color string) string {
    return color + text + colorReset
}

func colorizeLevel(level slog.Level) string {
    switch level {
    case slog.LevelDebug:
        return colorize("DEBUG", colorGray)
    case slog.LevelInfo:
        return colorize("INFO ", colorGreen)
    case slog.LevelWarn:
        return colorize("WARN ", colorYellow)
    case slog.LevelError:
        return colorize("ERROR", colorRed)
    default:
        return level.String()
    }
}
```

### 5. Redaction Using slog.LogValuer Interface

Leverage slog's built-in LogValuer interface for redaction (no custom implementation needed):

```go
// internal/log/types.go - Custom types that implement slog.LogValuer

// Token represents a sensitive authentication token
type Token string

// LogValue implements slog.LogValuer for automatic redaction
func (Token) LogValue() slog.Value {
    return slog.StringValue("[REDACTED_TOKEN]")
}

// Password represents a user password
type Password string

// LogValue implements slog.LogValuer for automatic redaction
func (Password) LogValue() slog.Value {
    return slog.StringValue("[REDACTED_PASSWORD]")
}

// Email with partial redaction for debugging
type Email string

// LogValue implements slog.LogValuer with partial visibility
func (e Email) LogValue() slog.Value {
    email := string(e)
    if len(email) == 0 {
        return slog.StringValue("")
    }
    if at := strings.Index(email, "@"); at > 0 {
        return slog.StringValue(email[:1] + "***" + email[at:])
    }
    return slog.StringValue("[REDACTED_EMAIL]")
}

// User struct with sensitive fields that implement LogValuer
type User struct {
    ID       string   `json:"id"`
    Name     string   `json:"name"`
    Email    Email    `json:"email"`
    Password Password `json:"-"` // Never serialize
    Token    Token    `json:"-"` // Never serialize
}

// LogValue groups public fields only
func (u User) LogValue() slog.Value {
    return slog.GroupValue(
        slog.String("id", u.ID),
        slog.String("name", u.Name),
        slog.Any("email", u.Email), // Will call Email.LogValue()
        // Password and Token are intentionally omitted
    )
}

// Usage examples (automatic redaction via LogValuer):
user := User{
    ID:       "user-123",
    Name:     "John Doe",
    Email:    Email("john@example.com"),
    Password: Password("secret123"),
    Token:    Token("jwt-token-here"),
}

log.Info(ctx, "User login", "user", user)
// Output: ... user={"id":"user-123","name":"John Doe","email":"j***@example.com"}

log.Info(ctx, "Token validation", "token", user.Token)
// Output: ... token="[REDACTED_TOKEN]"
```

## Migration Plan

### Step 1: Package Implementation
- Create `internal/log/logger.go` with core functionality
- Create `internal/log/pretty.go` with custom slog.Handler for development
- Create `internal/log/types.go` with LogValuer implementations for sensitive types
- Add comprehensive tests
- Create detailed `internal/log/README.md` with usage examples

### Step 2: Server Integration
- Update `internal/server/middleware.go` to use structured logging
- Update `internal/server/server.go` to replace `log` calls
- Update `cmd/ephd/main.go` for startup logging

### Step 3: Context Propagation & LogValuer Types
- Add context helpers for environment, PR, provider tracking
- Implement and test LogValuer types for sensitive data
- Integration points will be added as other packages are developed

### Step 4: Cleanup & Documentation
- Update project documentation to reflect final structure
- Update CLAUDE.md project structure if needed
- Update all READMEs to reference internal/log consistently
- Update Makefile to set LOG_PRETTY=true for development commands

### Step 5: Testing & Validation
- Unit tests for all logging functionality including LogValuer implementations
- Integration tests for middleware behavior
- Manual testing of pretty vs JSON output formats
- Security testing of LogValuer redaction functionality

## Environment Configuration

```bash
# Production (default)
LOG_LEVEL=info
# JSON logging by default

# Development
LOG_LEVEL=debug
LOG_PRETTY=true  # Enable pretty colored output
```

## Makefile Integration

Local development commands should automatically enable pretty logging:

```makefile
# In Makefile
run-daemon:
	@LOG_PRETTY=true LOG_LEVEL=debug go run cmd/ephd/main.go

run-cli:
	@LOG_PRETTY=true LOG_LEVEL=debug go run cmd/eph/main.go
```

## Testing Strategy

### Unit Tests
- **JSON output validation**: Parse JSON logs and verify field presence
- **Pretty output validation**: Check for color codes and readable format
- **Log level filtering**: Ensure debug/info/warn/error levels work correctly
- **Context propagation**: Verify all context keys appear in logs
- **Concurrent safety**: Multiple goroutines logging simultaneously

### Integration Tests
- **HTTP middleware**: Request ID propagation and HTTP request logging
- **Server lifecycle**: Startup, shutdown, error scenarios
- **Environment switching**: JSON vs pretty format based on environment

## Dependencies

No new external dependencies required:
- Uses Go 1.24+ standard `log/slog` package (project already uses Go 1.24.3)
- Existing `github.com/google/uuid` for request IDs
- All color handling in pretty printer uses ANSI escape codes
- LogValuer interface is built into slog (no custom redaction library needed)

## Future Extensions

### Phase 5: Advanced Features (Post-MVP)
- Log sampling for high-traffic scenarios
- Log forwarding to external systems
- Metrics integration (request duration, error rates)
- Custom log attributes for specific operations

### Integration Points
As the project develops, structured logging will integrate with:
- **Environment Controller**: Environment lifecycle logging
- **Provider System**: Infrastructure operation logging
- **Webhook Handlers**: PR event processing logging
- **Worker System**: Background job logging

## Success Criteria

1. **Functional Requirements**
   - JSON logging in production environments
   - Pretty colored logging in development
   - Context propagation across HTTP requests
   - Configurable log levels
   - Request ID correlation

2. **Non-Functional Requirements**
   - Thread-safe concurrent logging
   - Minimal performance overhead
   - Clear, readable log output in both formats
   - Comprehensive test coverage (>90%)

3. **Integration Requirements**
   - All existing `log` package usage replaced
   - HTTP middleware enhanced with structured logging
   - Server lifecycle events properly logged
   - No breaking changes to existing functionality

## Risks & Mitigations

### Risk: Performance Impact
**Mitigation**: Use structured logging efficiently, avoid expensive operations in hot paths

### Risk: Log Volume in Debug Mode
**Mitigation**: Clear level configuration and selective debug logging

### Risk: Pretty Handler Complexity
**Mitigation**: Keep implementation simple, focus on readability over features

### Risk: Context Propagation Overhead
**Mitigation**: Use lightweight context keys, avoid deep copying

This implementation plan provides a foundation for robust, observable logging throughout the Eph application while maintaining simplicity and performance.
