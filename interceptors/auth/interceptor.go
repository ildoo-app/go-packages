package auth

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/ildoo-app/go-packages/interceptors/tenant"
)

const HeaderAuthorization = "Authorization"

// Claims is the verified access-token payload exposed to handlers.
type Claims struct {
	Subject    string
	SessionID  string
	MerchantID string
	Audience   string
}

// TokenVerifier validates a bearer access token and returns claims.
type TokenVerifier interface {
	VerifyAccessToken(ctx context.Context, rawToken string) (*Claims, error)
}

// Interceptor verifies Authorization bearer tokens and, for tenant-scoped
// requests, asserts claims.tid matches X-Tenant-ID.
type Interceptor struct {
	verifier TokenVerifier
	public   map[string]struct{}
}

func New(verifier TokenVerifier, publicProcedures ...string) *Interceptor {
	public := make(map[string]struct{}, len(publicProcedures))
	for _, procedure := range publicProcedures {
		public[procedure] = struct{}{}
	}
	return &Interceptor{verifier: verifier, public: public}
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			return next(ctx, req)
		}
		if _, ok := i.public[req.Spec().Procedure]; ok {
			return next(ctx, req)
		}
		claims, err := i.authenticate(ctx, req.Header().Get(HeaderAuthorization))
		if err != nil {
			return nil, err
		}
		return next(WithContext(ctx, claims), req)
	}
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if _, ok := i.public[conn.Spec().Procedure]; ok {
			return next(ctx, conn)
		}
		claims, err := i.authenticate(ctx, conn.RequestHeader().Get(HeaderAuthorization))
		if err != nil {
			return err
		}
		return next(WithContext(ctx, claims), conn)
	}
}

func (i *Interceptor) authenticate(ctx context.Context, header string) (*Claims, error) {
	if header == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing authorization header"))
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid authorization header"))
	}

	claims, err := i.verifier.VerifyAccessToken(ctx, strings.TrimPrefix(header, prefix))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid access token"))
	}

	if tenantID, err := tenant.FromContext(ctx); err == nil {
		if claims.MerchantID == "" || claims.MerchantID != tenantID.String() {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("token tenant mismatch"))
		}
	}

	return claims, nil
}
