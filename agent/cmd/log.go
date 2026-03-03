package cmd

import (
	"context"
	"log/slog"

	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/utils"
)

var logger = utils.NewLogger()

// slog handler that redirects to logrus
type slogHandler struct {
	entry *logrus.Entry
}

func newSlogHandler() *slogHandler {
	return &slogHandler{
		entry: logrus.NewEntry(logger),
	}
}

func (h *slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	switch h.entry.Level {
	case logrus.DebugLevel:
		return level >= slog.LevelDebug
	case logrus.InfoLevel:
		return level >= slog.LevelInfo
	case logrus.WarnLevel:
		return level >= slog.LevelWarn
	case logrus.ErrorLevel:
		return level >= slog.LevelError
	default:
		return false
	}
}

func (h *slogHandler) Handle(ctx context.Context, record slog.Record) error {
	switch record.Level {
	case slog.LevelDebug:
		h.entry.Debug(record.Message)
	case slog.LevelInfo:
		h.entry.Info(record.Message)
	case slog.LevelWarn:
		h.entry.Warn(record.Message)
	case slog.LevelError:
		h.entry.Error(record.Message)
	default:
		h.entry.Info(record.Message)
	}
	// reset entry
	h.entry = logrus.NewEntry(logger)
	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	for _, attr := range attrs {
		h.entry = h.entry.WithField(attr.Key, attr.Value.Any())
	}
	return h
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	h.entry = h.entry.WithField("group", name)
	return h
}
