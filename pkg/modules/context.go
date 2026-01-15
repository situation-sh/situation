package modules

import (
	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/store"
)

const (
	CONTEXT_AGENT   = "agent"
	CONTEXT_STORAGE = "storage"
	CONTEXT_LOGGER  = "logger"
)

func dummyLogger() logrus.FieldLogger {
	// dummy logger
	l := logrus.New()
	l.Out = nil
	return l
}

func fallbackStorage(agent string) *store.BunStorage {
	s, err := store.NewSQLiteBunStorage(":memory:", agent, func(err error) {})
	if err != nil {
		// here we prefer panic since we must have a storage
		panic(err)
	}
	return s
}

func SituationContext(ctx context.Context, agent string, storage *store.BunStorage, logger logrus.FieldLogger) context.Context {
	ctx = context.WithValue(ctx, CONTEXT_AGENT, agent)
	ctx = context.WithValue(ctx, CONTEXT_STORAGE, storage)
	ctx = context.WithValue(ctx, CONTEXT_LOGGER, logger)
	return ctx
}

func getLogger(ctx context.Context, m Module) logrus.FieldLogger {
	// dummy logger
	if ctx == nil {
		return dummyLogger()
	}
	switch l := ctx.Value(CONTEXT_LOGGER).(type) {
	case logrus.FieldLogger:
		// we suppose a base logger is provided so we append the module field
		return l.WithField("module", m.Name())
	default:
		return dummyLogger()
	}
}

func getAgent(ctx context.Context) string {
	if ctx == nil {
		return uuid.New().String()
	}
	switch a := ctx.Value(CONTEXT_AGENT).(type) {
	case string:
		return a
	default:
		return uuid.New().String()
	}
}

func getStorage(ctx context.Context) *store.BunStorage {
	if ctx == nil {
		return fallbackStorage(getAgent(ctx))
	}
	switch s := ctx.Value(CONTEXT_STORAGE).(type) {
	case *store.BunStorage:
		return s
	default:
		return fallbackStorage(getAgent(ctx))
	}
}
