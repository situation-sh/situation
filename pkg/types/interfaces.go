package types

import (
	"github.com/asiffer/puzzle"
	"github.com/sirupsen/logrus"
	"github.com/situation-sh/situation/pkg/store"
)

type Configurable interface {
	Bind(config *puzzle.Config) error
}

type LogProducer interface {
	SetLogger(logrus.FieldLogger)
}

type StorageDemander interface {
	SetStore(store.Store)
}
