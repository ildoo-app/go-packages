package tenant

import "context"

type ctxKey struct{}

// WithContext returns a child context carrying tenantID.
func WithContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, tenantID)
}

// FromContext returns the tenant ID stored in ctx, if any.
func FromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	id, ok := ctx.Value(ctxKey{}).(string)
	return id, ok && id != ""
}
