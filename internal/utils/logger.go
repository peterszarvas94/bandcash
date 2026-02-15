package utils

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bandcash/internal/config"
)

// ANSI color codes
const (
	reset  = "\033[0m"
	red    = "\033[31m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
	gray   = "\033[90m"
	green  = "\033[32m"
)

// ColoredHandler outputs colored logs to terminal
type ColoredHandler struct {
	w     io.Writer
	level slog.Level
	attrs []slog.Attr
	group string
}

func NewColoredHandler(w io.Writer, level slog.Level) *ColoredHandler {
	return &ColoredHandler{w: w, level: level}
}

func (h *ColoredHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *ColoredHandler) Handle(_ context.Context, r slog.Record) error {
	// Timestamp in gray
	ts := gray + r.Time.Format(time.DateTime) + reset

	// Level with color
	var levelStr string
	switch r.Level {
	case slog.LevelDebug:
		levelStr = green + "DBG" + reset
	case slog.LevelInfo:
		levelStr = blue + "INF" + reset
	case slog.LevelWarn:
		levelStr = yellow + "WRN" + reset
	case slog.LevelError:
		levelStr = red + "ERR" + reset
	default:
		levelStr = r.Level.String()
	}

	// Message
	msg := r.Message

	// Attributes
	var attrs strings.Builder

	// Add handler-level attrs first
	for _, a := range h.attrs {
		writeAttr(&attrs, h.group, a)
	}

	// Add record attrs
	r.Attrs(func(a slog.Attr) bool {
		writeAttr(&attrs, h.group, a)
		return true
	})

	_, _ = fmt.Fprintf(h.w, "%s %s %s%s\n", ts, levelStr, msg, attrs.String())
	return nil
}

func writeAttr(b *strings.Builder, group string, a slog.Attr) {
	if a.Equal(slog.Attr{}) {
		return
	}

	key := a.Key
	if group != "" {
		key = group + "." + key
	}

	b.WriteString(" ")
	b.WriteString(cyan)
	b.WriteString(key)
	b.WriteString(reset)
	b.WriteString("=")
	_, _ = fmt.Fprintf(b, "%v", a.Value.Any())
}

func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ColoredHandler{
		w:     h.w,
		level: h.level,
		attrs: newAttrs,
		group: h.group,
	}
}

func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	newGroup := name
	if h.group != "" {
		newGroup = h.group + "." + name
	}
	return &ColoredHandler{
		w:     h.w,
		level: h.level,
		attrs: h.attrs,
		group: newGroup,
	}
}

// MultiHandler writes to multiple slog handlers
type MultiHandler struct {
	handlers []slog.Handler
}

// New creates a multi-handler logger with colored output to stdout and JSON to a file
func New(logFile io.Writer, level slog.Level) *slog.Logger {
	coloredHandler := NewColoredHandler(os.Stdout, level)
	jsonHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(&MultiHandler{handlers: []slog.Handler{coloredHandler, jsonHandler}})
}

func SetupLogger() {
	cfg := config.Load()

	err := os.MkdirAll(filepath.Dir(cfg.LogFile), 0755)
	if err != nil {
		slog.Error("failed to create logs directory", "err", err)
		os.Exit(1)
	}

	timestamp := time.Now().Format("2006-01-02_150405")
	ext := filepath.Ext(cfg.LogFile)
	base := strings.TrimSuffix(cfg.LogFile, ext)
	timestampedLogFile := fmt.Sprintf("%s_%s%s", base, timestamp, ext)

	logFile, err := os.OpenFile(timestampedLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("failed to open log file", "err", err)
		os.Exit(1)
	}

	log := New(logFile, cfg.LogLevel)
	slog.SetDefault(log)
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			err := h.Handle(ctx, r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: handlers}
}
