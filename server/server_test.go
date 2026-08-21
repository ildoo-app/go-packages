package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ildoo-app/go-packages/logger"
)

func TestNewRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	log := testLogger(t)
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

	_, _, err := New(nil, log, "/svc/", handler)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNewRejectsNilLogger(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	_, _, err := New(&Config{}, nil, "/svc/", handler)
	if err == nil || !strings.Contains(err.Error(), "logger is nil") {
		t.Fatalf("got %v, want logger is nil", err)
	}
}

func TestNewRejectsNilHandler(t *testing.T) {
	t.Parallel()

	_, _, err := New(&Config{}, testLogger(t), "/svc/", nil)
	if err == nil || !strings.Contains(err.Error(), "handler is nil") {
		t.Fatalf("got %v, want handler is nil", err)
	}
}

func TestNewRejectsBadServicePath(t *testing.T) {
	t.Parallel()

	log := testLogger(t)
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

	tests := []struct {
		path    ServicePath
		wantErr string
	}{
		{path: "", wantErr: "service path is empty"},
		{path: "svc/", wantErr: "must start with /"},
	}

	for _, tt := range tests {
		t.Run(string(tt.path), func(t *testing.T) {
			t.Parallel()
			_, _, err := New(&Config{}, log, tt.path, handler)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestStartAndShutdown(t *testing.T) {
	t.Parallel()

	port := freePort(t)
	cfg := &Config{
		Host:            "127.0.0.1",
		Port:            port,
		ShutdownTimeout: time.Second,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	s, cleanup, err := New(cfg, testLogger(t), "/svc/", handler)
	if err != nil {
		t.Fatal(err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
	}()

	waitForServer(t, port)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/svc/", port))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	cleanup()
	if err := <-errCh; err != nil {
		t.Fatalf("start: %v", err)
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

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
	return port
}

func waitForServer(t *testing.T, port int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not start")
}
