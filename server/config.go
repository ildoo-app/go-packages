package server

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

const (
	defaultPort              = 8080
	defaultShutdownTimeout   = 15 * time.Second
	defaultReadHeaderTimeout = 10 * time.Second
	defaultIdleTimeout       = 60 * time.Second
)

// Config holds HTTP server settings.
// Callers map their own env vars into this struct.
type Config struct {
	// Host is the bind address. Empty means all interfaces.
	Host string

	// Port is the TCP port to listen on. Defaults to 8080.
	Port int

	// ShutdownTimeout bounds graceful shutdown. Defaults to 15s.
	ShutdownTimeout time.Duration

	// ReadHeaderTimeout bounds the time to read request headers.
	// Defaults to 10s.
	ReadHeaderTimeout time.Duration

	// IdleTimeout is the keep-alive idle period. Defaults to 60s.
	IdleTimeout time.Duration

	// ReadTimeout and WriteTimeout are optional. Zero means no timeout.
	// Leave WriteTimeout unset for long-lived Connect/gRPC streams.
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (c *Config) validate() error {
	if c == nil {
		return errors.New("server: config is nil")
	}
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("server: Port %d is out of range", c.Port)
	}
	return nil
}

func (c *Config) withDefaults() {
	if c.Port == 0 {
		c.Port = defaultPort
	}
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = defaultShutdownTimeout
	}
	if c.ReadHeaderTimeout == 0 {
		c.ReadHeaderTimeout = defaultReadHeaderTimeout
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = defaultIdleTimeout
	}
}

func (c *Config) addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
