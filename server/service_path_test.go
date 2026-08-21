package server

import (
	"strings"
	"testing"
)

func TestServicePathValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    ServicePath
		wantErr string
	}{
		{
			name: "connect service path",
			path: "/ildoo.merchants.v1.MerchantService/",
		},
		{
			name: "root",
			path: "/",
		},
		{
			name:    "empty",
			path:    "",
			wantErr: "service path is empty",
		},
		{
			name:    "missing leading slash",
			path:    "svc/",
			wantErr: "must start with /",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.path.validate()
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

func TestServicePathString(t *testing.T) {
	t.Parallel()

	path := ServicePath("/ildoo.merchants.v1.MerchantService/")
	if got := path.String(); got != string(path) {
		t.Fatalf("String = %q, want %q", got, string(path))
	}
}
