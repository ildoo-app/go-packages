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
	got, err := FromContext(ctx)
	if err != nil {
		t.Fatalf("expected tenant ID in context: %v", err)
	}
	if got != id {
		t.Fatalf("FromContext = %q, want %q", got, id)
	}
}

func TestFromContextMissing(t *testing.T) {
	t.Parallel()

	if _, err := FromContext(context.Background()); err == nil {
		t.Fatal("expected error when tenant ID is missing")
	}
	if _, err := FromContext(nil); err == nil {
		t.Fatal("expected error for nil ctx")
	}
}

func TestMustFromContext(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	if got := MustFromContext(WithContext(context.Background(), id)); got != id {
		t.Fatalf("MustFromContext = %q, want %q", got, id)
	}
}

func TestMustFromContextPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	MustFromContext(context.Background())
}
