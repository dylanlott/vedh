package telemetry

import (
	"time"

	"github.com/openmtg/edh-go/pkg/deckimport"
	"github.com/openmtg/edh-go/pkg/ratelimit"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Role identifies which side of a create/join or board-activation flow an
// event describes. The PRD's vocabulary table names a "user role" field
// without enumerating its values; RoleHost and RoleInvitee are this plan's
// discretionary choice, matching the two roles the PRD's funnel groups by.
type Role string

const (
	// RoleHost identifies the player who created the game.
	RoleHost Role = "host"
	// RoleInvitee identifies a player who joined via an invite.
	RoleInvitee Role = "invitee"
)

// AllowedLabelNames is the closed set of Prometheus label names any
// vedh_-prefixed collector may declare. No collector may declare a label
// for a username, session ID, user ID, game ID, or attribution value:
// PostgreSQL is the source for product-funnel analysis; Prometheus/Grafana
// cover technical SLIs only.
var AllowedLabelNames = map[string]struct{}{
	"source":   {},
	"outcome":  {},
	"role":     {},
	"reason":   {},
	"provider": {},
	"surface":  {},
}

// previewDeckBuckets is tuned for a DB-bound preview that should land in
// tens of milliseconds, keeping resolution where the mass actually is while
// still bounding the tail (RESEARCH Pattern 6).
var previewDeckBuckets = []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 4, 8}

// deckProviderFetchBuckets places a boundary at exactly 3 and 8 seconds —
// the locked connect and total timeout budgets — so a pile-up at the
// connect timeout and one at the total timeout are separately readable.
// The default Prometheus bucket set has neither boundary and cannot show
// that distinction.
var deckProviderFetchBuckets = []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 3, 5, 8, 10}

// Collectors holds every vedh_-prefixed Prometheus collector declared by
// this phase. Construct with NewCollectors — never by declaring
// package-level vars built directly with promauto against
// prometheus.DefaultRegisterer — so tests can pass a private registry and
// so server/metrics.go has exactly one registration site.
type Collectors struct {
	deckImportTotal    *prometheus.CounterVec
	deckImportDuration *prometheus.HistogramVec

	guestSessionTotal    *prometheus.CounterVec
	guestSessionDuration *prometheus.HistogramVec

	gameCreateTotal    *prometheus.CounterVec
	gameCreateDuration *prometheus.HistogramVec
	gameJoinTotal      *prometheus.CounterVec
	gameJoinDuration   *prometheus.HistogramVec

	boardActivationTotal    *prometheus.CounterVec
	boardActivationDuration *prometheus.HistogramVec

	productEventsWritten *prometheus.CounterVec
	productEventsDropped *prometheus.CounterVec

	// deckSuggestionDuration and deckSuggestionTruncated instrument
	// Pattern 3's bounded eager suggestion lookup (D-02/D-03/D-04), plan
	// 01-05's own path rather than one of the four ROADMAP criterion-4
	// families above: they measure the *cost* of the suggestion path, not
	// a funnel step.
	deckSuggestionDuration  *prometheus.HistogramVec
	deckSuggestionTruncated prometheus.Counter

	// rateLimitTotal instruments plan 01-06's per-surface, per-client
	// limiter (server/ratelimit.go). It is incremented on both the
	// allowed and the limited outcome — never only on rejection — so the
	// ratio between them is readable in Grafana before a too-tight limit
	// shows up as a funnel dip.
	rateLimitTotal *prometheus.CounterVec

	// deckProviderFetchTotal and deckProviderFetchDuration instrument
	// plan 01-06's provider-agnostic secure outbound fetch client
	// (server/deck_providers.go). Declared here, ahead of the D-14
	// checkpoint that selects (or declines) a provider, exactly like the
	// four criterion-4 families above: the name, label set, and bucket
	// boundaries are closed now, and stay unobserved until plan 01-07
	// wires an emit site.
	deckProviderFetchTotal    *prometheus.CounterVec
	deckProviderFetchDuration *prometheus.HistogramVec
}

// NewCollectors registers every collector family into reg via
// promauto.With(reg) and returns the handle used to observe them.
//
// All four criterion-4 families are declared here — guest session, deck
// import, game create/join, and board activation — plus the two
// product-event counters, twelve collectors in total. Only the import and
// product-event families are observed in this plan.
//
// Declaration alone does not close ROADMAP Phase 1 criterion 4: a
// CounterVec or HistogramVec with no observation exports no child series,
// so /prometheus will not list a vedh_guest_session_total sample until
// ACT-005 (Phase 2) records one, nor vedh_game_create_total /
// vedh_game_join_total until ACT-006 (Phase 2) / ACT-008 (Phase 3), nor
// vedh_board_activation_total / vedh_board_activation_duration_seconds
// until ACT-009 (Phase 3). What this plan closes is the name, the label
// set, the bucket boundaries, and the cardinality proof; criterion 4
// closes when those emit sites land.
func NewCollectors(reg prometheus.Registerer) *Collectors {
	factory := promauto.With(reg)

	return &Collectors{
		deckImportTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "vedh_deck_import_total",
			Help: "Deck import attempts by source type and outcome.",
		}, []string{"source", "outcome"}),
		deckImportDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "vedh_deck_import_duration_seconds",
			Help:    "Deck import latency by source type and outcome.",
			Buckets: previewDeckBuckets,
		}, []string{"source", "outcome"}),

		guestSessionTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "vedh_guest_session_total",
			Help: "Guest session creation attempts by outcome. First observed by ACT-005 (Phase 2).",
		}, []string{"outcome"}),
		guestSessionDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name: "vedh_guest_session_duration_seconds",
			Help: "Guest session creation latency by outcome. First observed by ACT-005 (Phase 2).",
		}, []string{"outcome"}),

		// Create and join are separate families rather than one family
		// with a label, because the two have different denominators in
		// the PRD's funnel and a shared name would make the ratio
		// unreadable.
		gameCreateTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "vedh_game_create_total",
			Help: "Game create attempts by outcome. First observed by ACT-006 (Phase 2).",
		}, []string{"outcome"}),
		gameCreateDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name: "vedh_game_create_duration_seconds",
			Help: "Game create latency by outcome. First observed by ACT-006 (Phase 2).",
		}, []string{"outcome"}),
		gameJoinTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "vedh_game_join_total",
			Help: "Game join attempts by outcome. First observed by ACT-008 (Phase 3).",
		}, []string{"outcome"}),
		gameJoinDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name: "vedh_game_join_duration_seconds",
			Help: "Game join latency by outcome. First observed by ACT-008 (Phase 3).",
		}, []string{"outcome"}),

		boardActivationTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "vedh_board_activation_total",
			Help: "Board activation attempts by role and outcome. First observed by ACT-009 (Phase 3).",
		}, []string{"role", "outcome"}),
		boardActivationDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name: "vedh_board_activation_duration_seconds",
			Help: "Time to board activation by role. First observed by ACT-009 (Phase 3).",
		}, []string{"role"}),

		productEventsWritten: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "vedh_product_events_written_total",
			Help: "Product events successfully written, by outcome.",
		}, []string{"outcome"}),
		// Required, not optional: D-17 accepts that a failed write
		// silently under-counts the funnel, and this counter is the only
		// thing that lets a real activation dip be told apart from a
		// measurement gap.
		productEventsDropped: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "vedh_product_events_dropped_total",
			Help: "Product events dropped before or during write, by bounded reason.",
		}, []string{"reason"}),

		// Pattern 3 (RESEARCH.md): instrumenting the two bounds turns "is
		// 25 the right cap" and "is 750ms the right sub-budget" into
		// questions the Grafana panels answer after beta, rather than
		// numbers defended in review.
		deckSuggestionDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "vedh_deck_suggestion_duration_seconds",
			Help:    "Batched deck-import suggestion lookup latency by outcome (success|timeout|error).",
			Buckets: previewDeckBuckets,
		}, []string{"outcome"}),
		deckSuggestionTruncated: factory.NewCounter(prometheus.CounterOpts{
			Name: "vedh_deck_suggestion_truncated_total",
			Help: "Times the eager suggestion lookup truncated its needle set to the first 25 distinct unresolved names in one preview.",
		}),

		rateLimitTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "vedh_rate_limit_total",
			Help: "Rate limiter decisions on a public surface, by surface and outcome (allowed|limited).",
		}, []string{"surface", "outcome"}),

		// First observed by plan 01-07, once the D-14 checkpoint's selected
		// branch (or its no-go fallback) wires an emit site onto the
		// secure client server/deck_providers.go builds in this plan.
		deckProviderFetchTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "vedh_deck_provider_fetch_total",
			Help: "Outbound deck-provider fetch attempts by provider and outcome. First observed by plan 01-07.",
		}, []string{"provider", "outcome"}),
		deckProviderFetchDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "vedh_deck_provider_fetch_duration_seconds",
			Help:    "Outbound deck-provider fetch latency by provider and outcome, with boundaries at the 3s connect and 8s total timeout budgets. First observed by plan 01-07.",
			Buckets: deckProviderFetchBuckets,
		}, []string{"provider", "outcome"}),
	}
}

// ObserveDeckImport records one deck import attempt. source is a
// deckimport.SourceType and outcome is a small, server-controlled string
// ("success" | "failure") — a client-controlled value can never reach
// WithLabelValues here because the compiler requires the enum type.
func (c *Collectors) ObserveDeckImport(source deckimport.SourceType, outcome string, duration time.Duration) {
	c.deckImportTotal.WithLabelValues(string(source), outcome).Inc()
	c.deckImportDuration.WithLabelValues(string(source), outcome).Observe(duration.Seconds())
}

// ObserveGuestSession records one guest session creation attempt.
// First observed by ACT-005 in Phase 2.
func (c *Collectors) ObserveGuestSession(outcome string, duration time.Duration) {
	c.guestSessionTotal.WithLabelValues(outcome).Inc()
	c.guestSessionDuration.WithLabelValues(outcome).Observe(duration.Seconds())
}

// ObserveGameCreate records one game create attempt.
// First observed by ACT-006 in Phase 2.
func (c *Collectors) ObserveGameCreate(outcome string, duration time.Duration) {
	c.gameCreateTotal.WithLabelValues(outcome).Inc()
	c.gameCreateDuration.WithLabelValues(outcome).Observe(duration.Seconds())
}

// ObserveGameJoin records one game join attempt.
// First observed by ACT-008 in Phase 3.
func (c *Collectors) ObserveGameJoin(outcome string, duration time.Duration) {
	c.gameJoinTotal.WithLabelValues(outcome).Inc()
	c.gameJoinDuration.WithLabelValues(outcome).Observe(duration.Seconds())
}

// ObserveBoardActivation records one board activation. role is a
// telemetry.Role — again a compile-time-enforced enum, never a raw
// client-controlled string. First observed by ACT-009 in Phase 3.
func (c *Collectors) ObserveBoardActivation(role Role, outcome string, duration time.Duration) {
	c.boardActivationTotal.WithLabelValues(string(role), outcome).Inc()
	c.boardActivationDuration.WithLabelValues(string(role)).Observe(duration.Seconds())
}

// RecordProductEventWritten increments the written counter for one
// successful product_events insert.
func (c *Collectors) RecordProductEventWritten(outcome string) {
	c.productEventsWritten.WithLabelValues(outcome).Inc()
}

// RecordProductEventDropped increments the dropped counter for one refused
// or failed product_events write. reason is a Rejection, whose bounded
// String() form is the only thing ever passed to WithLabelValues here —
// the offending event name or key, which is attacker-chosen, must never
// reach this call.
func (c *Collectors) RecordProductEventDropped(reason Rejection) {
	c.productEventsDropped.WithLabelValues(reason.String()).Inc()
}

// The three accessors below return a single already-labelled
// prometheus.Counter rather than the underlying CounterVec, so a caller
// (production or test) can read or observe one specific, compiler-checked
// label combination with prometheus/testutil, without ever reaching
// WithLabelValues with an arbitrary string. This is the read-side analog of
// "give Collectors typed observation methods rather than exposing the
// vectors": the label values are pinned by this method's typed parameters,
// not supplied freely by the caller.

// DeckImportCounter returns the vedh_deck_import_total counter for one
// source/outcome pair, for use with prometheus/testutil.ToFloat64.
func (c *Collectors) DeckImportCounter(source deckimport.SourceType, outcome string) prometheus.Counter {
	return c.deckImportTotal.WithLabelValues(string(source), outcome)
}

// ProductEventsWrittenCounter returns the vedh_product_events_written_total
// counter for one outcome, for use with prometheus/testutil.ToFloat64.
func (c *Collectors) ProductEventsWrittenCounter(outcome string) prometheus.Counter {
	return c.productEventsWritten.WithLabelValues(outcome)
}

// ProductEventsDroppedCounter returns the vedh_product_events_dropped_total
// counter for one bounded reason, for use with prometheus/testutil.ToFloat64.
func (c *Collectors) ProductEventsDroppedCounter(reason Rejection) prometheus.Counter {
	return c.productEventsDropped.WithLabelValues(reason.String())
}

// ObserveDeckSuggestion records one batched eager-suggestion lookup
// (Pattern 3). outcome is a small, server-controlled string
// ("success" | "timeout" | "error") — never a client-controlled value.
func (c *Collectors) ObserveDeckSuggestion(outcome string, duration time.Duration) {
	c.deckSuggestionDuration.WithLabelValues(outcome).Observe(duration.Seconds())
}

// IncDeckSuggestionTruncated increments vedh_deck_suggestion_truncated_total
// once each time Pattern 3 bound #3 fires: more than maxSuggestionNeedles
// distinct unresolved names appeared in one preview.
func (c *Collectors) IncDeckSuggestionTruncated() {
	c.deckSuggestionTruncated.Inc()
}

// DeckSuggestionTruncatedCounter returns the
// vedh_deck_suggestion_truncated_total counter, for use with
// prometheus/testutil.ToFloat64.
func (c *Collectors) DeckSuggestionTruncatedCounter() prometheus.Counter {
	return c.deckSuggestionTruncated
}

// ObserveRateLimit increments vedh_rate_limit_total for one allowRequest
// decision. surface is a ratelimit.Surface and outcome is a small,
// server-controlled string ("allowed" | "limited") — a client-controlled
// value can never reach WithLabelValues here because the compiler requires
// the enum type for surface, and outcome is derived only from the
// registry's own boolean decision in server/ratelimit.go, never from
// caller input.
func (c *Collectors) ObserveRateLimit(surface ratelimit.Surface, outcome string) {
	c.rateLimitTotal.WithLabelValues(string(surface), outcome).Inc()
}

// RateLimitCounter returns the vedh_rate_limit_total counter for one
// surface/outcome pair, for use with prometheus/testutil.ToFloat64.
func (c *Collectors) RateLimitCounter(surface ratelimit.Surface, outcome string) prometheus.Counter {
	return c.rateLimitTotal.WithLabelValues(string(surface), outcome)
}

// ObserveDeckProviderFetch records one outbound deck-provider fetch
// attempt. provider and outcome are both small, bounded strings — never a
// caller-supplied host or error string. First observed by plan 01-07.
func (c *Collectors) ObserveDeckProviderFetch(provider, outcome string, duration time.Duration) {
	c.deckProviderFetchTotal.WithLabelValues(provider, outcome).Inc()
	c.deckProviderFetchDuration.WithLabelValues(provider, outcome).Observe(duration.Seconds())
}

// DeckProviderFetchCounter returns the vedh_deck_provider_fetch_total
// counter for one provider/outcome pair, for use with
// prometheus/testutil.ToFloat64.
func (c *Collectors) DeckProviderFetchCounter(provider, outcome string) prometheus.Counter {
	return c.deckProviderFetchTotal.WithLabelValues(provider, outcome)
}
