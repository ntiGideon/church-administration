package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"time"
)

// Color constants
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[90m"
)

// ColoredHandler is a custom slog.Handler that adds color to log output
type ColoredHandler struct {
	handler slog.Handler
	writer  io.Writer
}

func NewColoredHandler(w io.Writer, opts *slog.HandlerOptions) *ColoredHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	baseHandler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level:       opts.Level,
		AddSource:   opts.AddSource,
		ReplaceAttr: opts.ReplaceAttr,
	})

	return &ColoredHandler{
		handler: baseHandler,
		writer:  w,
	}
}

func (h *ColoredHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *ColoredHandler) Handle(ctx context.Context, r slog.Record) error {
	// Get the appropriate color for the log level
	var levelColor string
	switch r.Level {
	case slog.LevelDebug:
		levelColor = colorGray
	case slog.LevelInfo:
		levelColor = colorGreen
	case slog.LevelWarn:
		levelColor = colorYellow
	case slog.LevelError:
		levelColor = colorRed
	default:
		levelColor = colorWhite
	}

	// Format the time
	timeStr := r.Time.Format(time.DateTime)

	// Format the level with color
	levelStr := levelColor + r.Level.String() + colorReset

	// Format the message
	msg := colorCyan + r.Message + colorReset

	// Format source if available
	var source string
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		source = fmt.Sprintf("%s:%d", f.File, f.Line)
	}

	// Combine all parts
	var output string
	if source != "" {
		output = fmt.Sprintf("%s %s %s %s\n",
			colorGray+timeStr+colorReset,
			levelStr,
			msg,
			colorPurple+source+colorReset)
	} else {
		output = fmt.Sprintf("%s %s %s\n",
			colorGray+timeStr+colorReset,
			levelStr,
			msg)
	}

	// Write to the output writer
	_, err := h.writer.Write([]byte(output))
	return err
}

func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ColoredHandler{
		handler: h.handler.WithAttrs(attrs),
		writer:  h.writer,
	}
}

func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	return &ColoredHandler{
		handler: h.handler.WithGroup(name),
		writer:  h.writer,
	}
}
