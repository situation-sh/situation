package cmd

import (
	"github.com/getsentry/sentry-go"
	"github.com/situation-sh/situation/agent/config"
	"github.com/situation-sh/situation/pkg/modules"
)

func initSentry(dsn string) error {
	return sentry.Init(
		sentry.ClientOptions{
			Dsn:              dsn,
			EnableTracing:    true,
			TracesSampleRate: 1.0,
			EnableLogs:       true,
			ServerName:       config.AgentString(),
			Release:          config.Version,
			Dist:             config.Commit,
		},
	)
}

type sentrySupervisor struct {
	span *sentry.Span
}

func (s *sentrySupervisor) StartChild(name string) modules.SchedulerSupervisor {
	return &sentrySupervisor{span: s.span.StartChild(name)}
}

func (s *sentrySupervisor) Finish() {
	s.span.Finish()
}

func (s *sentrySupervisor) SetStatus(err error) {
	if err != nil {
		s.span.Status = sentry.SpanStatusInternalError
	} else {
		s.span.Status = sentry.SpanStatusOK
	}
}

func newSentrySupervisor(span *sentry.Span) modules.SchedulerSupervisor {
	return &sentrySupervisor{span: span}
}
