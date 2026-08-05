package server

import (
	"context"
	"strings"

	"github.com/openmtg/edh-go/pkg/ratelimit"
)

// rateLimitOutcomeAllowed and rateLimitOutcomeLimited are the only two
// values allowRequest ever passes to collectors.ObserveRateLimit's outcome
// parameter — derived solely from the registry's own boolean decision,
// never from caller input, closing the cardinality hole a client-supplied
// string would otherwise open on this label.
const (
	rateLimitOutcomeAllowed = "allowed"
	rateLimitOutcomeLimited = "limited"
)

// allowRequest is the thin instrumented wrapper this plan's must_haves
// require: it owns no bucket logic of its own (that lives in
// pkg/ratelimit.Registry, the CI-gated half of this feature), and it
// increments vedh_rate_limit_total on both the allowed and the limited
// path — never only on rejection — so the ratio between them is readable
// in Grafana before a too-tight limit shows up as a funnel dip. surface is
// a declared ratelimit.Surface constant, never a caller-supplied string,
// so the label stays bounded by the compiler rather than by review.
//
// A nil s.limiter (a graphQLServer built as a bare struct literal by a
// test that does not construct one through NewGraphQLServer) allows every
// request: rate limiting is a production concern this method adds, never a
// precondition a test must satisfy to call PreviewDeck or
// TrackProductEvent at all.
func (s *graphQLServer) allowRequest(ctx context.Context, surface ratelimit.Surface, clientKey string) bool {
	allowed := true
	if s != nil && s.limiter != nil {
		allowed = s.limiter.Allow(surface, clientKey)
	}

	outcome := rateLimitOutcomeAllowed
	if !allowed {
		outcome = rateLimitOutcomeLimited
	}
	collectors.ObserveRateLimit(surface, outcome)
	return allowed
}

// clientKeyFor derives the per-request rate-limit key: the caller-supplied
// session identifier when present, falling back to the request's remote
// address (server/graphql.go's withClientAddr) when it is not. A
// forwarded-for header is never consulted here — see withClientAddr's own
// comment for why trusting one by default would let an unauthenticated
// caller mint unlimited distinct keys.
func clientKeyFor(ctx context.Context, sessionID string) string {
	if trimmed := strings.TrimSpace(sessionID); trimmed != "" {
		return "session:" + trimmed
	}
	if addr, ok := remoteAddrFromContext(ctx); ok && strings.TrimSpace(addr) != "" {
		return "addr:" + addr
	}
	return "unknown"
}
