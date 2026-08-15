package tenant

import (
	"context"
	"testing"
)

func TestWithContextAndFromContext(t *testing.T) {
	t.Parallel()

	ctx := WithContext(context.Background(), "tenant-abc")
	got, ok := FromContext(ctx)
	if !ok {
		t.Fatal("expected tenant ID in context")
	}
	if got != "tenant-abc" {
		t.Fatalf("FromContext = %q, want %q", got, "tenant-abc")
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
	if _, ok := FromContext(WithContext(context.Background(), "")); ok {
		t.Fatal("expected empty tenant ID to be treated as missing")
	}
}
