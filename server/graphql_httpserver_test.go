package server

import (
	"net/http"
	"testing"
)

func TestNewHTTPServer_SetsSecurityTimeoutsAndLimits(t *testing.T) {
	h := http.NewServeMux()
	srv := newHTTPServer(8080, h)

	if srv == nil {
		t.Fatalf("newHTTPServer returned nil")
	}
	if srv.Addr != ":8080" {
		t.Fatalf("unexpected addr: got %q want %q", srv.Addr, ":8080")
	}
	if srv.Handler != h {
		t.Fatalf("handler mismatch")
	}
	if srv.ReadHeaderTimeout != readHeaderTimeout {
		t.Fatalf("unexpected ReadHeaderTimeout: got %v want %v", srv.ReadHeaderTimeout, readHeaderTimeout)
	}
	if srv.ReadTimeout != readTimeout {
		t.Fatalf("unexpected ReadTimeout: got %v want %v", srv.ReadTimeout, readTimeout)
	}
	if srv.WriteTimeout != writeTimeout {
		t.Fatalf("unexpected WriteTimeout: got %v want %v", srv.WriteTimeout, writeTimeout)
	}
	if srv.IdleTimeout != idleTimeout {
		t.Fatalf("unexpected IdleTimeout: got %v want %v", srv.IdleTimeout, idleTimeout)
	}
	if srv.MaxHeaderBytes != maxHeaderBytes {
		t.Fatalf("unexpected MaxHeaderBytes: got %d want %d", srv.MaxHeaderBytes, maxHeaderBytes)
	}
}
