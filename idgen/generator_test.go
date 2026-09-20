package idgen

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/ildoo-app/go-packages/logger"
)

func TestNewRejectsNilLogger(t *testing.T) {
	t.Parallel()

	_, err := New(nil)
	if err == nil || !strings.Contains(err.Error(), "logger is nil") {
		t.Fatalf("got %v, want logger is nil", err)
	}
}

func TestNewV7(t *testing.T) {
	t.Parallel()

	gen, err := New(testLogger(t))
	if err != nil {
		t.Fatal(err)
	}

	id := gen.NewV7()
	if id == uuid.Nil {
		t.Fatal("expected non-nil uuid")
	}
	if id.Version() != 7 {
		t.Fatalf("got version %d, want 7", id.Version())
	}
}

func TestNewV7FallsBackToV4(t *testing.T) {
	origV7, origV4 := newV7, newV4
	t.Cleanup(func() {
		newV7, newV4 = origV7, origV4
	})

	fallback := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	newV7 = func() (uuid.UUID, error) {
		return uuid.Nil, errors.New("boom")
	}
	newV4 = func() uuid.UUID {
		return fallback
	}

	var buf bytes.Buffer
	log, err := logger.New(&logger.Config{Writer: &buf, Format: logger.FormatJSON})
	if err != nil {
		t.Fatal(err)
	}

	gen, err := New(log)
	if err != nil {
		t.Fatal(err)
	}

	got := gen.NewV7()
	if got != fallback {
		t.Fatalf("got %v, want %v", got, fallback)
	}
	if !bytes.Contains(buf.Bytes(), []byte(`falling back to v4`)) {
		t.Fatalf("expected warning log, got %q", buf.String())
	}
	if !bytes.Contains(buf.Bytes(), []byte(`boom`)) {
		t.Fatalf("expected error in log, got %q", buf.String())
	}
}

func testLogger(t *testing.T) logger.Logger {
	t.Helper()
	log, err := logger.New(&logger.Config{Writer: io.Discard})
	if err != nil {
		t.Fatal(err)
	}
	return log
}
