package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/ildoo-app/go-packages/logger"
)

type server struct {
	cfg        *Config
	log        logger.Logger
	httpServer *http.Server
}

// New builds an HTTP server that serves handler at servicePath.
// The caller owns the Server and must call the returned cleanup function on process exit.
func New(
	cfg *Config,
	log logger.Logger,
	servicePath ServicePath,
	handler http.Handler,
) (Server, func(), error) {
	if err := cfg.validate(); err != nil {
		return nil, nil, err
	}
	cfg.withDefaults()

	if log == nil {
		return nil, nil, errors.New("server: logger is nil")
	}
	if handler == nil {
		return nil, nil, errors.New("server: handler is nil")
	}
	if err := servicePath.validate(); err != nil {
		return nil, nil, err
	}

	mux := http.NewServeMux()
	mux.Handle(servicePath.String(), handler)

	s := &server{
		cfg: cfg,
		log: log,
		httpServer: &http.Server{
			Addr:              cfg.addr(),
			Handler:           mux,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			IdleTimeout:       cfg.IdleTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
		},
	}

	return s, func() {
		s.log.Info("Stopping server...")
		ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.log.Error("Failed to shutdown server", "error", err)
		}
	}, nil
}

func (s *server) Start() error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("server: listen: %w", err)
	}

	s.log.Info("Starting server", "addr", ln.Addr().String())
	err = s.httpServer.Serve(ln)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server: serve: %w", err)
	}
	return nil
}
