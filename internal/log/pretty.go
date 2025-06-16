package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Color constants for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[37m"
	colorWhite  = "\033[97m"
)

// PrettyHandler implements slog.Handler for colored console output
type PrettyHandler struct {
	opts   *slog.HandlerOptions
	out    io.Writer
	mu     sync.Mutex
	attrs  []slog.Attr
	groups []string
}

// NewPrettyHandler creates a new pretty handler for development
func NewPrettyHandler(out io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &PrettyHandler{
		opts: opts,
		out:  out,
	}
}

// Enabled implements slog.Handler
func (h *PrettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle implements slog.Handler
func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Format timestamp
	timestamp := r.Time.Format("15:04:05.000")

	// Format level with color
	level := colorizeLevel(r.Level)

	// Format message
	message := r.Message

	// Build output
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[%s] %s %s",
		colorize(timestamp, colorGray),
		level,
		message)

	// Add pre-formatted attributes
	if len(h.attrs) > 0 {
		for _, attr := range h.attrs {
			formatAttr(&buf, attr, h.groups)
		}
	}

	// Add record attributes
	r.Attrs(func(attr slog.Attr) bool {
		formatAttr(&buf, attr, h.groups)
		return true
	})

	// Add source information if enabled
	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		if f.File != "" {
			// Show only relative path for readability
			file := f.File
			if idx := strings.LastIndex(file, "/eph/"); idx >= 0 {
				file = file[idx+5:]
			}
			fmt.Fprintf(&buf, " %s", colorize(fmt.Sprintf("%s:%d", file, f.Line), colorGray))
		}
	}

	buf.WriteByte('\n')

	_, err := h.out.Write(buf.Bytes())
	return err
}

// WithAttrs implements slog.Handler
func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &PrettyHandler{
		opts:   h.opts,
		out:    h.out,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

// WithGroup implements slog.Handler
func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	return &PrettyHandler{
		opts:   h.opts,
		out:    h.out,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

// formatAttr formats a single attribute with color coding
func formatAttr(buf *bytes.Buffer, attr slog.Attr, groups []string) {
	// Skip empty attributes
	if attr.Equal(slog.Attr{}) {
		return
	}

	// Handle special time attribute
	if attr.Key == slog.TimeKey {
		return // Already handled in main format
	}

	// Build full key with groups
	key := attr.Key
	if len(groups) > 0 {
		key = strings.Join(groups, ".") + "." + key
	}

	// Format based on value type
	switch v := attr.Value.Any().(type) {
	case string:
		if v != "" {
			fmt.Fprintf(buf, " %s=%s",
				colorize(key, colorCyan),
				colorize(fmt.Sprintf("%q", v), colorWhite))
		}
	case int, int64, uint, uint64:
		fmt.Fprintf(buf, " %s=%s",
			colorize(key, colorCyan),
			colorize(fmt.Sprintf("%v", v), colorYellow))
	case float32, float64:
		fmt.Fprintf(buf, " %s=%s",
			colorize(key, colorCyan),
			colorize(fmt.Sprintf("%v", v), colorYellow))
	case bool:
		color := colorRed
		if v {
			color = colorGreen
		}
		fmt.Fprintf(buf, " %s=%s",
			colorize(key, colorCyan),
			colorize(fmt.Sprintf("%v", v), color))
	case time.Time:
		fmt.Fprintf(buf, " %s=%s",
			colorize(key, colorCyan),
			colorize(v.Format(time.RFC3339), colorWhite))
	case time.Duration:
		fmt.Fprintf(buf, " %s=%s",
			colorize(key, colorCyan),
			colorize(v.String(), colorYellow))
	case error:
		fmt.Fprintf(buf, " %s=%s",
			colorize(key, colorCyan),
			colorize(fmt.Sprintf("%q", v.Error()), colorRed))
	case json.RawMessage:
		fmt.Fprintf(buf, " %s=%s",
			colorize(key, colorCyan),
			colorize(string(v), colorWhite))
	default:
		// Handle nested groups and other complex types
		if attr.Value.Kind() == slog.KindGroup {
			groupAttrs := attr.Value.Group()
			for _, gAttr := range groupAttrs {
				newGroups := make([]string, len(groups)+1)
				copy(newGroups, groups)
				newGroups[len(groups)] = attr.Key
				formatAttr(buf, gAttr, newGroups)
			}
			return
		}

		// Default formatting for other types
		fmt.Fprintf(buf, " %s=%s",
			colorize(key, colorCyan),
			colorize(fmt.Sprintf("%v", v), colorWhite))
	}
}

// colorize adds ANSI color codes to text
func colorize(text, color string) string {
	return color + text + colorReset
}

// colorizeLevel returns colored level string
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
		return fmt.Sprintf("%-5s", level.String())
	}
}
