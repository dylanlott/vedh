package server

import (
	"net/http"
	"testing"
)

func TestParseAllowedOrigins_NormalizesAndFilters(t *testing.T) {
	allowed := parseAllowedOrigins(" http://localhost:5173,https://app.example.com/path,ftp://bad,,not-a-url ")
	if _, ok := allowed["http://localhost:5173"]; !ok {
		t.Fatalf("expected localhost origin to be allowed")
	}
	if _, ok := allowed["https://app.example.com"]; !ok {
		t.Fatalf("expected normalized https origin to be allowed")
	}
	if _, ok := allowed["ftp://bad"]; ok {
		t.Fatalf("unexpected ftp origin in allowlist")
	}
}

func TestGraphQLServer_IsAllowedOrigin(t *testing.T) {
	s := &graphQLServer{allowedOrigins: map[string]struct{}{"https://app.example.com": {}}}
	if !s.isAllowedOrigin("https://app.example.com/some/path") {
		t.Fatalf("expected origin to be allowed")
	}
	if s.isAllowedOrigin("https://evil.example.com") {
		t.Fatalf("did not expect origin to be allowed")
	}
}

func TestGraphQLServer_IsAllowedWebSocketOrigin(t *testing.T) {
	s := &graphQLServer{allowedOrigins: map[string]struct{}{"https://app.example.com": {}}}

	reqNoOrigin, _ := http.NewRequest(http.MethodGet, "http://localhost/graphql", nil)
	if !s.isAllowedWebSocketOrigin(reqNoOrigin) {
		t.Fatalf("expected request without origin to be allowed")
	}

	reqAllowed, _ := http.NewRequest(http.MethodGet, "http://localhost/graphql", nil)
	reqAllowed.Header.Set("Origin", "https://app.example.com")
	if !s.isAllowedWebSocketOrigin(reqAllowed) {
		t.Fatalf("expected allowed origin")
	}

	reqDenied, _ := http.NewRequest(http.MethodGet, "http://localhost/graphql", nil)
	reqDenied.Header.Set("Origin", "https://evil.example.com")
	if s.isAllowedWebSocketOrigin(reqDenied) {
		t.Fatalf("expected denied origin")
	}
}
