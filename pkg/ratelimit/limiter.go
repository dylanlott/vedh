// Package ratelimit provides a DB-free, per-surface, per-client-key
// token-bucket registry with bounded growth. It imports nothing from
// github.com/openmtg/edh-go/server and nothing from
// github.com/openmtg/edh-go/pkg/telemetry, so it stays a leaf package that
// `go test ./pkg/... -race` — the only test target CI actually runs — gates
// on every pull request, per 01-VALIDATION.md's "CI reality" note. The
// counting and client-key derivation that make this registry's decision
// visible to an operator live in server/ratelimit.go, one layer up.
package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Surface identifies which limited operation a bucket belongs to. It is a
// named type, deliberately unlike a raw string, so a Prometheus label built
// from it (server/ratelimit.go) is bounded by the compiler rather than by
// review — the same reasoning pkg/deckimport.SourceType documents for its
// own label type.
//
// One constant is declared per surface this phase actually limits.
// Guest creation and public invite lookup (later phases, per
// .planning/intel/constraints.md's "Rate limiting on public surfaces") add
// their own constants here without changing Registry's shape at all: the
// composite (Surface, client key) map key already isolates any number of
// surfaces from one another.
type Surface string

const (
	// SurfaceDeckImport identifies the previewDeck mutation.
	SurfaceDeckImport Surface = "deck_import"
	// SurfaceProductEvent identifies the trackProductEvent mutation.
	SurfaceProductEvent Surface = "product_event"
)

// idleWindow is how long a (surface, client key) bucket may go untouched
// before Registry evicts it. It is deliberately a package constant, not a
// third environment variable alongside the two rate-budget envconfig
// fields on server.Conf: the bounded-growth guarantee this value protects
// must not be weakenable from the deployment environment.
const idleWindow = 10 * time.Minute

// bucketKey is the composite map key that keeps two distinct client keys,
// or two distinct surfaces, from ever sharing a bucket.
type bucketKey struct {
	surface   Surface
	clientKey string
}

// bucketEntry pairs one client's token-bucket limiter with the last time it
// was consulted, so an idle sweep can find and evict it.
type bucketEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Registry is a mutex-guarded, in-memory token-bucket registry keyed by
// (Surface, client key). It never touches a database, a request, or a
// resolver — the whole point of splitting it out of server/ratelimit.go is
// that this file's tests run in CI with no Postgres and no MTGJSON
// snapshot.
type Registry struct {
	mu        sync.Mutex
	perMinute int
	burst     int
	entries   map[bucketKey]*bucketEntry

	// now is the injected clock, defaulting to time.Now. Tests advance it
	// directly (this package's tests are white-box, package ratelimit) to
	// prove idle eviction without sleeping for idleWindow.
	now func() time.Time
}

// NewRegistry constructs a Registry whose buckets refill at perMinute
// tokens per minute up to burst tokens. Both values apply uniformly across
// every surface and client key this Registry serves.
func NewRegistry(perMinute, burst int) *Registry {
	return &Registry{
		perMinute: perMinute,
		burst:     burst,
		entries:   make(map[bucketKey]*bucketEntry),
		now:       time.Now,
	}
}

// Allow is the whole public decision surface: it reports whether one
// request for surface from clientKey should proceed, creating a fresh
// bucket for a (surface, clientKey) pair never seen before, and sweeping
// every idle entry first so the map never grows without bound.
func (r *Registry) Allow(surface Surface, clientKey string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	r.sweepLocked(now)

	key := bucketKey{surface: surface, clientKey: clientKey}
	e, ok := r.entries[key]
	if !ok {
		e = &bucketEntry{
			// perMinute tokens per minute is perMinute/60 tokens per
			// second, which is what rate.Limit expects.
			limiter: rate.NewLimiter(rate.Limit(float64(r.perMinute))/60, r.burst),
		}
		r.entries[key] = e
	}
	e.lastSeen = now
	return e.limiter.AllowN(now, 1)
}

// sweepLocked evicts every entry untouched for longer than idleWindow. The
// caller must hold r.mu. This is the bounded-growth control: without it, an
// unauthenticated caller minting new client keys (or a deployment serving
// many distinct surfaces) grows this map forever.
func (r *Registry) sweepLocked(now time.Time) {
	for k, e := range r.entries {
		if now.Sub(e.lastSeen) > idleWindow {
			delete(r.entries, k)
		}
	}
}
