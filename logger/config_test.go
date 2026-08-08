package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *Config
		wantErr string
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: "config is nil",
		},
		{
			name: "empty format ok",
			cfg:  &Config{},
		},
		{
			name: "json",
			cfg:  &Config{Format: FormatJSON},
		},
		{
			name: "text",
			cfg:  &Config{Format: FormatText},
		},
		{
			name: "case insensitive format",
			cfg:  &Config{Format: "JSON"},
		},
		{
			name:    "bad format",
			cfg:     &Config{Format: "yaml"},
			wantErr: "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cfg.validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in      string
		want    slog.Level
		wantErr bool
	}{
		{in: "debug", want: slog.LevelDebug},
		{in: "INFO", want: slog.LevelInfo},
		{in: "", want: slog.LevelInfo},
		{in: "warn", want: slog.LevelWarn},
		{in: "warning", want: slog.LevelWarn},
		{in: "error", want: slog.LevelError},
		{in: "trace", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, err := ParseLevel(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseLevel(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestNewJSON(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log, err := New(&Config{
		Level:   slog.LevelInfo,
		Format:  FormatJSON,
		Writer:  &buf,
		Service: "test-svc",
	})
	if err != nil {
		t.Fatal(err)
	}

	log.Info("hello", slog.String("k", "v"))

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json %q: %v", buf.String(), err)
	}
	if payload["msg"] != "hello" {
		t.Fatalf("msg = %v, want hello", payload["msg"])
	}
	if payload["service"] != "test-svc" {
		t.Fatalf("service = %v, want test-svc", payload["service"])
	}
	if payload["k"] != "v" {
		t.Fatalf("k = %v, want v", payload["k"])
	}
}

func TestNewText(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log, err := New(&Config{
		Format: FormatText,
		Writer: &buf,
	})
	if err != nil {
		t.Fatal(err)
	}

	log.Info("hello")
	if !strings.Contains(buf.String(), "hello") {
		t.Fatalf("expected message in output: %q", buf.String())
	}
}

func TestNewRespectsLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log, err := New(&Config{
		Level:  slog.LevelWarn,
		Format: FormatJSON,
		Writer: &buf,
	})
	if err != nil {
		t.Fatal(err)
	}

	log.Info("suppressed")
	log.Warn("visible")

	out := buf.String()
	if strings.Contains(out, "suppressed") {
		t.Fatalf("info should be filtered: %q", out)
	}
	if !strings.Contains(out, "visible") {
		t.Fatalf("warn should be present: %q", out)
	}
}

func TestNewRejectsNilConfig(t *testing.T) {
	t.Parallel()

	if _, err := New(nil); err == nil {
		t.Fatal("expected error for nil config")
	}
}
