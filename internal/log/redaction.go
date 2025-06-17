package log

import (
	"log/slog"
)

func (Token) LogValue() slog.Value {
	return slog.StringValue("[REDACTED_TOKEN]")
}

func (t Token) String() string {
	return string(t)
}

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

func (k APIKey) String() string {
	return string(k)
}

func (e Environment) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", e.ID),
		slog.String("name", e.Name),
		slog.String("url", e.URL),
		slog.String("provider", e.Provider),
	)
}
