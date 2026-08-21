package auth

import "context"

type ctxKey struct{}

func WithContext(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, claims)
}

func FromContext(ctx context.Context) (*Claims, bool) {
	if ctx == nil {
		return nil, false
	}
	claims, ok := ctx.Value(ctxKey{}).(*Claims)
	return claims, ok && claims != nil
}
