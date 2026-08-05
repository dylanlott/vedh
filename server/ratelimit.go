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

// clientKeyFor derives the per-request rate-limit key. CR-02 fix (code
// review, phase 01): sessionID is a caller-supplied, unauthenticated field
// (app/src/services/productEvents.ts's newSessionID) — never trust it
// alone as a rate-limit key, or any caller can mint an unlimited number of
// distinct buckets simply by sending a fresh random sessionID on every
// request, defeating the limiter entirely. The server-observed remote
// address (server/graphql.go's withClientAddr, sourced from
// r.RemoteAddr — the TCP peer address, never a client-supplied header) is
// the sole bucketing key whenever it is available, precisely so that a
// single IP rotating session IDs cannot escape into a fresh bucket per
// request: appending sessionID onto the address (e.g. "addr|session")
// would NOT achieve this, since a caller who varies sessionID while
// keeping the same address would still produce a distinct key each time.
// sessionID is consulted only as a last-resort fallback when no
// server-observed address is present in context (e.g. a direct call in a
// test, or a future non-HTTP transport) — it is never a substitute for the
// address component when both are available. No per-session fairness is
// layered on top today; if that becomes a goal, it must be implemented as
// an additional, independent limit alongside the address-anchored one, not
// by folding sessionID into this key. A forwarded-for header is never
// consulted here — see withClientAddr's own comment for why trusting one
// by default would let an unauthenticated caller mint unlimited distinct
// keys.
func clientKeyFor(ctx context.Context, sessionID string) string {
	if addr, ok := remoteAddrFromContext(ctx); ok {
		if trimmed := strings.TrimSpace(addr); trimmed != "" {
			return "addr:" + trimmed
		}
	}
	if trimmed := strings.TrimSpace(sessionID); trimmed != "" {
		return "session:" + trimmed
	}
	return "unknown"
}
