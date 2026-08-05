package server

import (
	"context"
	"strings"
	"testing"

	"github.com/openmtg/edh-go/pkg/ratelimit"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// TestRateLimit_AllowRequestCountsBothOutcomes proves the wrapper's counting
// behavior directly, with no database: vedh_rate_limit_total increments by
// exactly one per allowRequest call, on both the allowed and the limited
// path, with the outcome label distinguishing them.
func TestRateLimit_AllowRequestCountsBothOutcomes(t *testing.T) {
	s := &graphQLServer{limiter: ratelimit.NewRegistry(60, 1)}
	ctx := context.Background()

	allowedBefore := testutil.ToFloat64(collectors.RateLimitCounter(ratelimit.SurfaceDeckImport, rateLimitOutcomeAllowed))
	limitedBefore := testutil.ToFloat64(collectors.RateLimitCounter(ratelimit.SurfaceDeckImport, rateLimitOutcomeLimited))

	if !s.allowRequest(ctx, ratelimit.SurfaceDeckImport, "client-a") {
		t.Fatal("expected the first request to be allowed")
	}
	if s.allowRequest(ctx, ratelimit.SurfaceDeckImport, "client-a") {
		t.Fatal("expected the second request to be limited")
	}

	allowedAfter := testutil.ToFloat64(collectors.RateLimitCounter(ratelimit.SurfaceDeckImport, rateLimitOutcomeAllowed))
	limitedAfter := testutil.ToFloat64(collectors.RateLimitCounter(ratelimit.SurfaceDeckImport, rateLimitOutcomeLimited))

	if delta := allowedAfter - allowedBefore; delta != 1 {
		t.Errorf("allowed counter delta = %v, want 1", delta)
	}
	if delta := limitedAfter - limitedBefore; delta != 1 {
		t.Errorf("limited counter delta = %v, want 1", delta)
	}
}

// TestRateLimit_NilLimiterAllowsEverything proves a graphQLServer built
// as a bare struct literal (no limiter constructed) never rate limits: this
// is what keeps every other test in this package, none of which cares
// about rate limiting, unaffected by this plan.
func TestRateLimit_NilLimiterAllowsEverything(t *testing.T) {
	s := &graphQLServer{}
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if !s.allowRequest(ctx, ratelimit.SurfaceDeckImport, "client-a") {
			t.Fatalf("request %d: expected a nil limiter to allow every request", i)
		}
	}
}

// TestRateLimit_PreviewDeckNamesNoLimitValue proves that a limited
// previewDeck call returns a product-language blocking error naming no
// limit value, window, or remaining count, and CanContinue false. It
// exhausts the registry's single-token burst directly (bypassing
// PreviewDeck) so PreviewDeck's own call is the one that observes the
// limit and never reaches s.db, which is nil here -- this test needs no
// database precisely because it never reaches one.
func TestRateLimit_PreviewDeckNamesNoLimitValue(t *testing.T) {
	s := &graphQLServer{limiter: ratelimit.NewRegistry(60, 1)}
	text := "1 Sol Ring"
	input := InputDeckImport{Text: &text, SessionID: "rate-limit-test-session"}

	clientKey := clientKeyFor(context.Background(), input.SessionID)
	if !s.limiter.Allow(ratelimit.SurfaceDeckImport, clientKey) {
		t.Fatal("expected the first Allow call against the registry to succeed")
	}

	preview, err := s.PreviewDeck(context.Background(), input)
	if err != nil {
		t.Fatalf("PreviewDeck() error = %v", err)
	}
	if preview.CanContinue {
		t.Fatal("expected CanContinue = false for a rate-limited call")
	}
	if len(preview.BlockingErrors) != 1 {
		t.Fatalf("expected exactly one blocking error, got %d: %v", len(preview.BlockingErrors), preview.BlockingErrors)
	}

	msg := strings.ToLower(preview.BlockingErrors[0])
	for _, forbidden := range []string{"limit", "minute", "burst", "30", "10", "quota", "window"} {
		if strings.Contains(msg, forbidden) {
			t.Errorf("blocking error %q names a limit-related term %q", preview.BlockingErrors[0], forbidden)
		}
	}
}

// TestRateLimit_TrackProductEventStillReturnsTrue proves a limited
// trackProductEvent call still returns true with no error, per D-19/D-22.
func TestRateLimit_TrackProductEventStillReturnsTrue(t *testing.T) {
	s := &graphQLServer{limiter: ratelimit.NewRegistry(60, 1)}
	input := InputProductEvent{Name: "quick_start_viewed", SessionID: "rate-limit-test-session"}

	clientKey := clientKeyFor(context.Background(), input.SessionID)
	if !s.limiter.Allow(ratelimit.SurfaceProductEvent, clientKey) {
		t.Fatal("expected the first Allow call against the registry to succeed")
	}

	ok, err := s.TrackProductEvent(context.Background(), input)
	if err != nil {
		t.Fatalf("TrackProductEvent() error = %v", err)
	}
	if !ok {
		t.Fatal("expected TrackProductEvent to return true even when rate limited")
	}
}

// TestRateLimit_ClientKeyForPrefersSessionOverAddress proves the derivation order:
// a non-empty session ID always wins over the remote address, and an
// absent session ID with no remote address in context falls back to a
// fixed "unknown" key rather than an empty string (which would collapse
// every unidentified caller onto the same key as a caller who explicitly
// supplied an empty one).
func TestRateLimit_ClientKeyForPrefersSessionOverAddress(t *testing.T) {
	ctxWithAddr := context.WithValue(context.Background(), remoteAddrContextKey{}, "203.0.113.5")

	if got := clientKeyFor(ctxWithAddr, "session-123"); got != "session:session-123" {
		t.Errorf("clientKeyFor() = %q, want a session-prefixed key", got)
	}
	if got := clientKeyFor(ctxWithAddr, ""); got != "addr:203.0.113.5" {
		t.Errorf("clientKeyFor() = %q, want an addr-prefixed key", got)
	}
	if got := clientKeyFor(context.Background(), ""); got != "unknown" {
		t.Errorf("clientKeyFor() = %q, want \"unknown\"", got)
	}
}
