package lifecycle

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ildoo-app/go-packages/logger"
)

// BootContext is cancelled when the process receives SIGINT or SIGTERM. It is
// for startup work only: dialing databases, discovery, cache warmup. It is not
// a context.Context so a component that outlives startup cannot retain it.
type BootContext struct {
	ctx context.Context
}

// Context returns the underlying signalled context for short-lived startup work.
func (b BootContext) Context() context.Context { return b.ctx }

// Lifecycle owns the process signal context and bounds teardown with a hard deadline.
type Lifecycle struct {
	log  logger.Logger
	cfg  *Config
	ctx  context.Context
	stop context.CancelFunc
	exit func(int)
}

// New installs signal handling. The returned cleanup is registered by Wire and
// runs last, after every other cleanup, so a teardown that never reaches it
// still gets its signal disposition restored by beginShutdown.
func New(log logger.Logger, cfg *Config) (*Lifecycle, func(), error) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	return newLifecycle(log, cfg, ctx, stop, os.Exit)
}

func newLifecycle(
	log logger.Logger,
	cfg *Config,
	ctx context.Context,
	stop context.CancelFunc,
	exit func(int),
) (*Lifecycle, func(), error) {
	if err := cfg.validate(); err != nil {
		stop()
		return nil, nil, err
	}
	cfg.withDefaults()

	if log == nil {
		stop()
		return nil, nil, errors.New("lifecycle: logger is nil")
	}
	if exit == nil {
		exit = os.Exit
	}

	l := &Lifecycle{
		log:  log,
		cfg:  cfg,
		ctx:  ctx,
		stop: stop,
		exit: exit,
	}

	return l, func() { l.stop() }, nil
}

// BootContextOf extracts the startup-only BootContext from a Lifecycle.
// Intended as a Wire provider alongside New.
func BootContextOf(l *Lifecycle) BootContext {
	return BootContext{ctx: l.ctx}
}

// Run starts each component in its own goroutine and blocks until a signal is
// received or a component fails. Components are stopped by the Wire cleanup
// chain after Run returns; do not join them here or shutdown deadlocks.
func (l *Lifecycle) Run(components ...Component) error {
	errCh := make(chan error, len(components))
	for _, c := range components {
		go func(c Component) { errCh <- c.Start() }(c)
	}
	select {
	case <-l.ctx.Done():
		l.beginShutdown("signal received")
		return nil
	case err := <-errCh:
		l.beginShutdown("component failed")
		return err
	}
}

func (l *Lifecycle) beginShutdown(reason string) {
	l.stop() // re-arm default signal disposition before the long teardown
	l.log.Info("Shutting down", "reason", reason, "grace", l.cfg.GraceTimeout)
	// The binary is PID 1 in the container, so a repeat SIGTERM is ignored by
	// the kernel. This watchdog is the only escape from a wedged teardown.
	time.AfterFunc(l.cfg.GraceTimeout, func() {
		l.log.Error("Teardown exceeded grace period, forcing exit")
		l.exit(1)
	})
}
