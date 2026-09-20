package idgen

import (
	"github.com/google/uuid"
)

// Generator creates unique identifiers.
// The value returned by New satisfies it.
type Generator interface {
	NewV7() uuid.UUID
}
