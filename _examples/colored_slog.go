package _examples

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"

	"github.com/hedzr/is/term/color"

	"github.com/hedzr/go-socketlib/net"
)

type PrettyHandlerOptions struct {
	SlogOpts slog.HandlerOptions
}

type PrettyHandler struct {
	slog.Handler
	l *log.Logger
}

func levelToString(l slog.Level) string {
	if t, ok := net.LevelNames[l]; ok {
		return t
	}
	return l.String()
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	levelStr := levelToString(r.Level) + ":"

	switch r.Level {
	case slog.LevelDebug:
		levelStr = color.ToColor(color.FgMagenta, levelStr)
	case slog.LevelInfo:
		levelStr = color.ToColor(color.FgBlue, levelStr)
	case slog.LevelWarn:
		levelStr = color.ToColor(color.FgYellow, levelStr)
	case slog.LevelError:
		levelStr = color.ToColor(color.FgRed, levelStr)
	}

	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})

	b, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return err
	}

	timeStr := r.Time.Format("15:05:05.000")
	msg := color.ToColor(color.FgCyan, r.Message)

	h.l.Println(timeStr, levelStr, msg, color.ToColor(color.FgDarkColor, string(b)))

	return nil
}

func NewPrettyHandler(
	out io.Writer,
	opts PrettyHandlerOptions,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, &opts.SlogOpts),
		l:       log.New(out, "", 0),
	}

	return h
}
