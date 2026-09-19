package server

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGraphQLServer_ShouldExposeMetrics(t *testing.T) {
	s := &graphQLServer{cfg: Conf{MetricsEnabled: true, MetricsToken: "secret-token"}}
	if !s.shouldExposeMetrics() {
		t.Fatalf("expected metrics to be enabled")
	}
	s.cfg.MetricsEnabled = false
	if s.shouldExposeMetrics() {
		t.Fatalf("expected metrics to be disabled")
	}
}

func TestGraphQLServer_ShouldExposeMetrics_RequiresTokenWhenEnabled(t *testing.T) {
	s := &graphQLServer{cfg: Conf{MetricsEnabled: true, MetricsToken: ""}}
	if s.shouldExposeMetrics() {
		t.Fatalf("expected metrics to stay disabled without a token")
	}
}

func TestGraphQLServer_WithMetricsAuth_DisabledWhenTokenEmpty(t *testing.T) {
	s := &graphQLServer{cfg: Conf{MetricsToken: ""}}
	called := false
	h := s.withMetricsAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/prometheus", nil)
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", resp.Code, http.StatusOK)
	}
	if !called {
		t.Fatalf("expected wrapped handler to be called")
	}
}

func TestGraphQLServer_WithMetricsAuth_RejectsMissingOrWrongBearer(t *testing.T) {
	s := &graphQLServer{cfg: Conf{MetricsToken: "secret-token"}}
	h := s.withMetricsAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, tc := range []struct {
		name string
		auth string
	}{
		{name: "missing", auth: ""},
		{name: "wrong", auth: "Bearer wrong-token"},
		{name: "wrong scheme", auth: "Basic abc"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/prometheus", nil)
			if tc.auth != "" {
				req.Header.Set("Authorization", tc.auth)
			}
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			if resp.Code != http.StatusUnauthorized {
				t.Fatalf("unexpected status: got %d want %d", resp.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestGraphQLServer_WithMetricsAuth_AllowsValidBearer(t *testing.T) {
	s := &graphQLServer{cfg: Conf{MetricsToken: "secret-token"}}
	called := false
	h := s.withMetricsAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/prometheus", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", resp.Code, http.StatusOK)
	}
	if !called {
		t.Fatalf("expected wrapped handler to be called")
	}
}

func TestGraphQLServer_WithAuthContext_DoesNotTouchUserActivityForPrometheusScrape(t *testing.T) {
	s := testAPI(t)
	var logs bytes.Buffer
	s.logger = slog.New(slog.NewTextHandler(&logs, nil))

	token, err := newAuthToken(&User{ID: "prometheus", Username: "prometheus"}, time.Hour)
	if err != nil {
		t.Fatalf("failed to mint auth token: %v", err)
	}

	h := s.withAuthContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/prometheus", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	h.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: got %d want %d", resp.Code, http.StatusNoContent)
	}
	if strings.Contains(logs.String(), "failed to touch authenticated user activity") {
		t.Fatalf("expected prometheus scrape to skip user activity touch, got logs: %s", logs.String())
	}
}
