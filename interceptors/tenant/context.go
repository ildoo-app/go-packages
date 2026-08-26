package tenant

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct{}

// WithContext returns a child context carrying tenantID.
func WithContext(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKey{}, tenantID)
}

// FromContext returns the tenant ID stored in ctx, if any.
func FromContext(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}
	id, ok := ctx.Value(ctxKey{}).(uuid.UUID)
	return id, ok && id != uuid.Nil
}
