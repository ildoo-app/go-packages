package tenant

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
)

type testMsg struct{}

func TestInterceptorUnaryExtractsTenantID(t *testing.T) {
	t.Parallel()

	interceptor := NewInterceptor()
	var got string
	next := connect.UnaryFunc(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		id, ok := FromContext(ctx)
		if !ok {
			t.Fatal("expected tenant ID in handler context")
		}
		got = id
		return connect.NewResponse(&testMsg{}), nil
	})

	req := connect.NewRequest(&testMsg{})
	req.Header().Set(HeaderTenantID, "tenant-123")

	if _, err := interceptor.WrapUnary(next)(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "tenant-123" {
		t.Fatalf("tenant ID = %q, want %q", got, "tenant-123")
	}
}

func TestInterceptorUnaryMissingHeader(t *testing.T) {
	t.Parallel()

	interceptor := NewInterceptor()
	next := connect.UnaryFunc(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
		t.Fatal("next should not be called")
		return nil, nil
	})

	req := connect.NewRequest(&testMsg{})
	_, err := interceptor.WrapUnary(next)(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing header")
	}
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}
	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want %v", connectErr.Code(), connect.CodeInvalidArgument)
	}
}

func TestInterceptorUnarySkipsClient(t *testing.T) {
	t.Parallel()

	interceptor := NewInterceptor()
	called := false
	next := connect.UnaryFunc(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		called = true
		if _, ok := FromContext(ctx); ok {
			t.Fatal("client path should not inject tenant ID")
		}
		return connect.NewResponse(&testMsg{}), nil
	})

	req := connect.NewRequest(&testMsg{})
	clientReq := &clientAnyRequest{Request: req}
	if _, err := interceptor.WrapUnary(next)(context.Background(), clientReq); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected next to be called")
	}
}

// clientAnyRequest marks Spec.IsClient so the interceptor treats the call as outbound.
type clientAnyRequest struct {
	*connect.Request[testMsg]
}

func (r *clientAnyRequest) Spec() connect.Spec {
	spec := r.Request.Spec()
	spec.IsClient = true
	return spec
}
