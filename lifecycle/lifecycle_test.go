package lifecycle

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ildoo-app/go-packages/logger"
)

func TestNewRejectsNilConfig(t *testing.T) {
	t.Parallel()

	_, _, err := New(testLogger(t), nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNewRejectsNilLogger(t *testing.T) {
	t.Parallel()

	_, _, err := New(nil, &Config{})
	if err == nil {
		t.Fatal("expected logger is nil")
	}
}

func TestBootContextCancelsOnStop(t *testing.T) {
	t.Parallel()

	ctx, stop := context.WithCancel(context.Background())
	lc, cleanup, err := newLifecycle(testLogger(t), &Config{}, ctx, stop, func(int) {})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	boot := BootContextOf(lc)
	if err := boot.Context().Err(); err != nil {
		t.Fatalf("boot context already done: %v", err)
	}

	stop()
	select {
	case <-boot.Context().Done():
	case <-time.After(time.Second):
		t.Fatal("boot context did not cancel")
	}
}

func TestRunReturnsNilOnSignal(t *testing.T) {
	t.Parallel()

	ctx, stop := context.WithCancel(context.Background())
	lc, cleanup, err := newLifecycle(testLogger(t), &Config{GraceTimeout: time.Hour}, ctx, stop, func(int) {})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	started := make(chan struct{})
	comp := componentFunc(func() error {
		close(started)
		<-ctx.Done()
		return nil
	})

	errCh := make(chan error, 1)
	go func() { errCh <- lc.Run(comp) }()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("component did not start")
	}

	stop()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("got %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not return")
	}
}

func TestRunReturnsComponentError(t *testing.T) {
	t.Parallel()

	ctx, stop := context.WithCancel(context.Background())
	lc, cleanup, err := newLifecycle(testLogger(t), &Config{GraceTimeout: time.Hour}, ctx, stop, func(int) {})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	want := errors.New("boom")
	comp := componentFunc(func() error { return want })

	got := lc.Run(comp)
	if !errors.Is(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestWatchdogFiresPastGraceTimeout(t *testing.T) {
	t.Parallel()

	var exited atomic.Int32
	ctx, stop := context.WithCancel(context.Background())
	lc, cleanup, err := newLifecycle(
		testLogger(t),
		&Config{GraceTimeout: 20 * time.Millisecond},
		ctx,
		stop,
		func(code int) { exited.Store(int32(code)) },
	)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	comp := componentFunc(func() error {
		<-ctx.Done()
		time.Sleep(time.Hour)
		return nil
	})

	errCh := make(chan error, 1)
	go func() { errCh <- lc.Run(comp) }()
	stop()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("got %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not return")
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if exited.Load() == 1 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("watchdog did not fire")
}

type componentFunc func() error

func (f componentFunc) Start() error { return f() }

func testLogger(t *testing.T) logger.Logger {
	t.Helper()
	log, err := logger.New(&logger.Config{Writer: io.Discard})
	if err != nil {
		t.Fatal(err)
	}
	return log
}
