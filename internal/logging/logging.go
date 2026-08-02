// Package logging configures a JSON slog logger suitable for shipping to an
// OTEL collector (and onward to ClickHouse or similar log stores).
package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// Init installs a JSON slog default logger that writes to w, tagged with the
// given service version and instance identity.
func Init(w io.Writer, version, instance string) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		ReplaceAttr: replaceAttr,
	})
	logger := slog.New(h).With(
		"service", "chaosbox",
		"version", version,
		"instance", instance,
	)
	slog.SetDefault(logger)
	return logger
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return a
	}
	switch a.Key {
	case slog.TimeKey:
		return slog.String("timestamp", a.Value.Time().UTC().Format(time.RFC3339Nano))
	case slog.LevelKey:
		level, ok := a.Value.Any().(slog.Level)
		if !ok {
			return a
		}
		return slog.String("level", strings.ToLower(level.String()))
	case slog.MessageKey:
		return slog.String("msg", a.Value.String())
	default:
		return a
	}
}

// Fatal logs an error-level structured event and exits the process.
func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
