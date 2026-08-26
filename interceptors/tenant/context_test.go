package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestWithContextAndFromContext(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	ctx := WithContext(context.Background(), id)
	got, ok := FromContext(ctx)
	if !ok {
		t.Fatal("expected tenant ID in context")
	}
	if got != id {
		t.Fatalf("FromContext = %q, want %q", got, id)
	}
}

func TestFromContextMissing(t *testing.T) {
	t.Parallel()

	if _, ok := FromContext(context.Background()); ok {
		t.Fatal("expected no tenant ID")
	}
	if _, ok := FromContext(nil); ok {
		t.Fatal("expected no tenant ID for nil ctx")
	}
	if _, ok := FromContext(WithContext(context.Background(), uuid.Nil)); ok {
		t.Fatal("expected nil tenant ID to be treated as missing")
	}
}
