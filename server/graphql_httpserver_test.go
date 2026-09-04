package server

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestWithMaxGraphQLBody_RejectsOversizedPOSTBeforeDecode(t *testing.T) {
	s := &graphQLServer{}
	handler := s.withMaxGraphQLBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		var tooLarge *http.MaxBytesError
		if !errors.As(err, &tooLarge) {
			t.Fatalf("expected MaxBytesError, got %v", err)
		}
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	}))

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(bytes.Repeat([]byte("x"), int(maxGraphQLBodyBytes)+1)))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("unexpected status: got %d want %d", res.Code, http.StatusRequestEntityTooLarge)
	}
}
