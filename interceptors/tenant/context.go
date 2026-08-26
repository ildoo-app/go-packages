package tenant

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type ctxKey struct{}

// WithContext returns a child context carrying tenantID.
func WithContext(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKey{}, tenantID)
}

// FromContext returns the tenant ID stored in ctx, if any.
func FromContext(ctx context.Context) (uuid.UUID, error) {
	if ctx == nil {
		return uuid.Nil, errors.New("context is nil")
	}
	id, ok := ctx.Value(ctxKey{}).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid or missing tenant ID in context")
	}
	return id, nil
}

func MustFromContext(ctx context.Context) uuid.UUID {
	id, err := FromContext(ctx)
	if err != nil {
		panic(err)
	}
	return id
}
