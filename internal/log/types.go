package log

import (
	"log/slog"
	"strings"
)

// Token represents a sensitive authentication token
type Token string

// LogValue implements slog.LogValuer for automatic redaction
func (Token) LogValue() slog.Value {
	return slog.StringValue("[REDACTED_TOKEN]")
}

// String implements fmt.Stringer
func (t Token) String() string {
	return string(t)
}

// Password represents a user password
type Password string

// LogValue implements slog.LogValuer for automatic redaction
func (Password) LogValue() slog.Value {
	return slog.StringValue("[REDACTED_PASSWORD]")
}

// String implements fmt.Stringer
func (p Password) String() string {
	return string(p)
}

// Secret represents any generic secret value
type Secret string

// LogValue implements slog.LogValuer for automatic redaction
func (Secret) LogValue() slog.Value {
	return slog.StringValue("[REDACTED_SECRET]")
}

// String implements fmt.Stringer
func (s Secret) String() string {
	return string(s)
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
		// Show first character and domain
		return slog.StringValue(email[:1] + "***" + email[at:])
	}
	return slog.StringValue("[INVALID_EMAIL]")
}

// String implements fmt.Stringer
func (e Email) String() string {
	return string(e)
}

// APIKey represents an API key with partial visibility
type APIKey string

// LogValue implements slog.LogValuer showing only prefix
func (k APIKey) LogValue() slog.Value {
	key := string(k)
	if len(key) == 0 {
		return slog.StringValue("")
	}
	if len(key) > 8 {
		return slog.StringValue(key[:4] + "..." + key[len(key)-4:])
	}
	return slog.StringValue("[REDACTED_KEY]")
}

// String implements fmt.Stringer
func (k APIKey) String() string {
	return string(k)
}

// User represents a user with sensitive fields
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

// Environment represents an ephemeral environment
type Environment struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Provider string `json:"provider"`
	Token    Token  `json:"-"` // Access token, never serialize
}

// LogValue provides safe logging for environments
func (e Environment) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", e.ID),
		slog.String("name", e.Name),
		slog.String("url", e.URL),
		slog.String("provider", e.Provider),
		// Token is intentionally omitted
	)
}

// DatabaseConfig represents database configuration with credentials
type DatabaseConfig struct {
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	Database string   `json:"database"`
	Username string   `json:"username"`
	Password Password `json:"-"` // Never serialize
}

// LogValue provides safe logging for database config
func (d DatabaseConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", d.Host),
		slog.Int("port", d.Port),
		slog.String("database", d.Database),
		slog.String("username", d.Username),
		// Password is intentionally omitted
	)
}
