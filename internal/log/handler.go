package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
)

// Handler is a slog.Handler that writes user-friendly messages to stderr.
//
// Output format:
//
//	Info:  message
//	Warn:  [WARNING] message
//	Error: [ERROR] message
//
// Attributes are displayed on separate lines with 2-space indent:
//
//	Registration successful
//	  registration_id = abc-123
//
// The [WARNING] and [ERROR] prefixes are colorized unless the NO_COLOR
// environment variable is set (see https://no-color.org).
type Handler struct {
	noColor bool
}

// NewHandler creates a new Handler.
func NewHandler() *Handler {
	_, noColor := os.LookupEnv("NO_COLOR")
	return &Handler{noColor: noColor}
}

func (h *Handler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder

	if prefix := h.prefix(r.Level); prefix != "" {
		b.WriteString(prefix)
		b.WriteByte(' ')
	}

	b.WriteString(r.Message)

	r.Attrs(func(a slog.Attr) bool {
		b.WriteString("\n  ")
		b.WriteString(a.Key)
		b.WriteString(" = ")
		b.WriteString(a.Value.String())
		return true
	})

	_, err := fmt.Fprintln(os.Stderr, b.String())
	return err
}

func (h *Handler) WithAttrs(_ []slog.Attr) slog.Handler { return h }

func (h *Handler) WithGroup(_ string) slog.Handler { return h }

func (h *Handler) prefix(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return h.colorize(colorRed, "[ERROR]")
	case level >= slog.LevelWarn:
		return h.colorize(colorYellow, "[WARNING]")
	default:
		return ""
	}
}

func (h *Handler) colorize(color, text string) string {
	if h.noColor {
		return text
	}
	return color + text + colorReset
}
