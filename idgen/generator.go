package idgen

import (
	"errors"

	"github.com/google/uuid"
	"github.com/ildoo-app/go-packages/logger"
)

// Swappable for tests.
var (
	newV7 = uuid.NewV7
	newV4 = uuid.New
)

type generator struct {
	log logger.Logger
}

// New builds a Generator that logs through log.
func New(log logger.Logger) (Generator, error) {
	if log == nil {
		return nil, errors.New("idgen: logger is nil")
	}
	return &generator{log: log}, nil
}

// NewV7 returns a UUID version 7, falling back to version 4 if generation fails.
func (g *generator) NewV7() uuid.UUID {
	id, err := newV7()
	if err != nil {
		g.log.Warn("failed to generate uuid v7, falling back to v4", "error", err)
		return newV4()
	}
	return id
}
