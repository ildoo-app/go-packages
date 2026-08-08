package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
)

func TestWithContextAndFromContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log, err := New(&Config{Writer: &buf, Format: FormatJSON})
	if err != nil {
		t.Fatal(err)
	}
	log = log.With(slog.String("req_id", "abc"))

	ctx := WithContext(context.Background(), log)
	FromContext(ctx).Info("from ctx")

	if !bytes.Contains(buf.Bytes(), []byte(`"req_id":"abc"`)) {
		t.Fatalf("expected req_id in output: %q", buf.String())
	}
}

func TestFromContextFallback(t *testing.T) {
	t.Parallel()

	if got := FromContext(context.Background()); got == nil {
		t.Fatal("expected slog.Default fallback")
	}
	if got := FromContext(nil); got == nil {
		t.Fatal("expected slog.Default for nil ctx")
	}
}
