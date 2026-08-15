package tenant

import (
	"context"
	"errors"

	"connectrpc.com/connect"
)

// HeaderTenantID is the HTTP header carrying the tenant identifier.
const HeaderTenantID = "X-Tenant-ID"

// Interceptor extracts X-Tenant-ID from inbound Connect requests and stores it
// on the request context for handlers to read via FromContext.
type Interceptor struct{}

// NewInterceptor returns a Connect interceptor that extracts X-Tenant-ID.
func NewInterceptor() *Interceptor {
	return &Interceptor{}
}

// WrapUnary implements connect.Interceptor.
func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			return next(ctx, req)
		}
		ctx, err := contextWithTenantID(ctx, req.Header().Get(HeaderTenantID))
		if err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

// WrapStreamingClient implements connect.Interceptor.
func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler implements connect.Interceptor.
func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, err := contextWithTenantID(ctx, conn.RequestHeader().Get(HeaderTenantID))
		if err != nil {
			return err
		}
		return next(ctx, conn)
	}
}

func contextWithTenantID(ctx context.Context, tenantID string) (context.Context, error) {
	if tenantID == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("missing X-Tenant-ID header"),
		)
	}
	return WithContext(ctx, tenantID), nil
}
