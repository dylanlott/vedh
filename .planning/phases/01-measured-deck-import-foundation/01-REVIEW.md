---
phase: 01-measured-deck-import-foundation
reviewed: 2026-08-05T21:28:31Z
depth: standard
files_reviewed: 48
files_reviewed_list:
  - app/__tests__/productEvents.spec.ts
  - app/src/graphql/mutations.ts
  - app/src/services/productEvents.ts
  - docs/analytics/product-event-funnel-example.sql
  - docs/analytics/product-event-vocabulary.md
  - docs/research/deck-provider-feasibility.md
  - go.mod
  - persistence/import_all_printings_json.go
  - persistence/migrations_test/20260804120000_product_events.down.sql
  - persistence/migrations_test/20260804120000_product_events.up.sql
  - persistence/migrations_test/20260804120100_card_name_search.down.sql
  - persistence/migrations_test/20260804120100_card_name_search.up.sql
  - persistence/migrations/20260804120000_product_events.down.sql
  - persistence/migrations/20260804120000_product_events.up.sql
  - persistence/migrations/20260804120100_card_name_search.down.sql
  - persistence/migrations/20260804120100_card_name_search.up.sql
  - pkg/deckimport/golden_test.go
  - pkg/deckimport/result.go
  - pkg/deckimport/scanner_test.go
  - pkg/deckimport/scanner.go
  - pkg/deckimport/sections_test.go
  - pkg/deckimport/sections.go
  - pkg/deckimport/sourcetype_test.go
  - pkg/deckimport/sourcetype.go
  - pkg/ratelimit/limiter_test.go
  - pkg/ratelimit/limiter.go
  - pkg/telemetry/metrics_test.go
  - pkg/telemetry/metrics.go
  - pkg/telemetry/vocabulary_test.go
  - pkg/telemetry/vocabulary.go
  - server/cards_test.go
  - server/cards.go
  - server/deck_import_test.go
  - server/deck_import.go
  - server/deck_providers_test.go
  - server/deck_providers.go
  - server/formats.go
  - server/games_test.go
  - server/games.go
  - server/graphql.go
  - server/metrics.go
  - server/product_events_test.go
  - server/product_events.go
  - server/ratelimit_test.go
  - server/ratelimit.go
  - server/schema.graphql
  - server/schema.resolvers.go
  - server/test.go
findings:
  critical: 2
  warning: 4
  info: 1
  total: 7
status: issues_found
---

# Phase 01: Code Review Report

**Reviewed:** 2026-08-05T21:28:31Z
**Depth:** standard
**Files Reviewed:** 48
**Status:** issues_found

## Summary

This phase's telemetry allowlist (`pkg/telemetry.Vocabulary`/`ValidateEvent`), Prometheus
cardinality discipline, and outbound SSRF controls (`server/deck_providers.go`) are all
implemented carefully and match their documentation closely — I traced the client-controlled
paths into `product_events.metadata`, the Prometheus label set, and the outbound fetch
client's redirect/address checks and did not find a way to defeat any of the three headline
guarantees (closed vocabulary, bounded labels, private-address denial) as stated.

Two real defects did surface, both severe enough to affect production correctness/security
today:

1. `server/games.go`'s `JoinGame` never received the `defaultLifeForAll` life-total fix that
   `CreateGame` got in this same phase, so a player who joins without an explicit life total
   starts the game already counted as eliminated (`game_finish.go`'s `alivePlayerNames`
   treats `Life == 0` as dead), and the very next `UpdateBoardState` call from anyone finishes
   the game as a win for the other side.
2. `server/ratelimit.go`'s `clientKeyFor` derives the token-bucket key from the caller-supplied,
   unauthenticated `sessionID` field before falling back to remote address, so any caller can
   defeat both the `previewDeck` and `trackProductEvent` rate limits simply by sending a fresh
   random `sessionID` on every request — the exact "client key spoofable" failure mode called
   out for review, just via a body field rather than a header.

Additionally, the outbound deck-provider fetch client's operator-facing host allowlist
(`DECK_PROVIDER_ALLOWED_HOSTS`) is checked for non-emptiness only; it is never consulted for
the *initial* request's host, only for redirects. Today this has no attacker-reachable blast
radius because the adapter registry (`deckProviderAdapters`) is hardcoded to a single host, but
it means the "hostname allowlist" control documented in `deck_providers.go`'s own comments does
not do what its comments claim for the request that actually matters.

Test-infrastructure risk (area of interest 7) is real but consistent with this codebase's
explicitly documented "CI reality" (several packages' own comments state `go test ./server/...`
is not part of the CI-gated target): `server/test.go`'s `testAPI` hardcodes a DSN, ignores
`DATABASE_URL`, and calls `t.Skipf` on every DB-unavailability path, so a misconfigured or
unreachable database makes the entire `server` package's test suite report as passed-with-skips
rather than failed.

## Structural Findings (fallow)

None provided for this review — no `<structural_findings>` block was supplied.

## Narrative Findings (AI reviewer)

## Critical Issues

### CR-01: `JoinGame` has no life-total default, so an omitted life silently starts a player "already eliminated"

**File:** `server/games.go:541-554` (boardstate construction inside `JoinGame`), contrasted with
`server/games.go:657-676` (the `defaultLifeForAll` fix added to `CreateGame` in this same phase)

**Issue:** This phase's own fix comment at `games.go:657-670` explains the ambiguity precisely:
`InputBoardState.Life` is a required, non-pointer `Int!` (`server/schema.graphql:196,247`), so
there is no wire-level way to distinguish "the caller didn't set a life total" from "the caller
explicitly wants 0." `CreateGame` was given a heuristic (`defaultLifeForAll`) to resolve this:
when *no* player in the call specifies a nonzero life, every player defaults to
`format.StartingLife`.

`JoinGame` was never given the equivalent treatment. At `games.go:546` the joining player's
boardstate is built with `Life: input.BoardState.Life` verbatim — no defaulting at all. A
client that omits (or zero-values) the life field when constructing a join request — exactly
the same ambiguity the `CreateGame` fix exists to solve — creates a player whose
`Boardstate.Life == 0` from the moment they join.

`server/game_finish.go`'s `alivePlayerNames` (consulted from `UpdateBoardState`,
`server/boardstates.go:89-99`) counts a player with `Boardstate.Life > 0` as alive and anyone
else — including this newly-joined player — as already dead. In a 2-player game, the very next
`UpdateBoardState` call from *either* player (not necessarily the affected one) computes
`alive == 1` and calls `finalizeGame(game, GameResultWin, []string{alive[0]}, nil)`, ending the
game as an immediate win for the opponent before the newly-joined player ever took an action.

This is exactly the risk flagged in the review brief: removing `ensureFormatRules`'s blanket
per-player life defaulting (a deliberate, well-reasoned fix for the `CreateGame`/general-load
regression this phase's SUMMARY documents) left `JoinGame` depending on a defaulting behavior
that no longer exists anywhere in its own path. `server/games_test.go`'s `TestJoinGame` and
`TestMultipleSubscriptions` never exercise an omitted/zero life value for `JoinGame` — every
call in the test suite passes `Life: 40` — so this gap is untested as well as unfixed.

**Fix:** Apply the same `defaultLifeForAll`-style heuristic to `JoinGame`, defaulting the
joining player's life to the game's own format's `StartingLife` (via `formatFromRules(game.Rules)`)
whenever the caller did not supply a nonzero value:

```go
format := formatFromRules(game.Rules)
life := input.BoardState.Life
if life == 0 {
    life = format.StartingLife
}
user := &User{
    Username: input.BoardState.User,
    ID:       input.BoardState.UserID,
    Boardstate: &BoardState{
        User: input.BoardState.User,
        Life: life,
        // ...
    },
}
```
(A caller that genuinely wants a joining player to start at 0 has no other affected players in
this call to signal intent against, unlike `CreateGame`'s multi-player heuristic — but that
tradeoff already exists for `CreateGame` and is documented; the important fix is that `JoinGame`
must not silently default to "already eliminated.")

---

### CR-02: The rate-limit client key is entirely client-controlled, letting any caller bypass `previewDeck`/`trackProductEvent` limits at will

**File:** `server/ratelimit.go:54-62` (`clientKeyFor`)

**Issue:** `clientKeyFor` derives the token-bucket key primarily from `sessionID`, the
caller-supplied field on `InputDeckImport`/`InputProductEvent`, falling back to remote address
only when `sessionID` is blank:

```go
func clientKeyFor(ctx context.Context, sessionID string) string {
	if trimmed := strings.TrimSpace(sessionID); trimmed != "" {
		return "session:" + trimmed
	}
	if addr, ok := remoteAddrFromContext(ctx); ok && strings.TrimSpace(addr) != "" {
		return "addr:" + addr
	}
	return "unknown"
}
```

`sessionID` is never authenticated, signed, or tied to anything server-controlled — it is a
plain string the client generates client-side (`app/src/services/productEvents.ts`'s
`getSessionID`/`newSessionID`) and sends verbatim. Both `previewDeck` (which runs a
suggestion query against the database on every unresolved card) and `trackProductEvent` (which
performs a database write) are public, unauthenticated mutations. An attacker who sends a
freshly-generated random `sessionID` on every request — trivial to script, and requiring no
special access — gets a brand-new token bucket every time, defeating the limiter entirely for
both surfaces. This is precisely the "can the client key be spoofed" question the review brief
raised about headers; here it is spoofable through an ordinary mutation input field, which is if
anything easier to exploit than a header, since no proxy or client library needs to be
convinced to forward an unusual header.

`server/graphql.go`'s `withClientAddr` comment explicitly reasons about *why* trusting a
client-supplied value (there, `X-Forwarded-For`) for the rate-limit key would be unsafe — but
`clientKeyFor` prefers exactly such a client-supplied value (`sessionID`) over the one value
(`remoteAddr`) that withClientAddr deliberately keeps trustworthy.

**Fix:** Do not let a caller-chosen value alone gate a rate limit meant to bound abuse. At
minimum, always incorporate the remote address into the key (e.g. `addr + "|" + session`, so a
single IP rotating session IDs still shares one bucket per surface), or drop the session-based
key entirely in favor of address-based limiting for these two public surfaces:

```go
func clientKeyFor(ctx context.Context, sessionID string) string {
	addr, _ := remoteAddrFromContext(ctx)
	addr = strings.TrimSpace(addr)
	session := strings.TrimSpace(sessionID)
	switch {
	case addr != "" && session != "":
		return "addr:" + addr + "|session:" + session
	case addr != "":
		return "addr:" + addr
	case session != "":
		return "session:" + session
	default:
		return "unknown"
	}
}
```
If per-session fairness (not spoofing resistance) is the actual goal, that should be stated
explicitly and layered on top of an address-anchored limit, not substituted for it.

## Warnings

### WR-01: The outbound deck-provider host allowlist is enforced only on redirects, never on the initial request

**File:** `server/deck_providers.go:217-256` (`newSafeProviderClient`/`newSafeProviderClientWithOptions`),
`server/deck_import.go:248-292` (`previewDeckURL`, `deckProviderClient`)

**Issue:** `s.deckProviderAllowedHosts` (parsed from `DECK_PROVIDER_ALLOWED_HOSTS`) is passed
into `newSafeProviderClient` and used only inside `newDeckProviderCheckRedirect`
(`deck_providers.go:194-208`), which Go's `http.Client` invokes solely before following a
*redirect* hop — never for the first request of a chain. `fetchDeckProviderURL`
(`deck_providers.go:282-313`) validates only that the initial URL's scheme is `https` and its
host is non-empty; it never checks the host against `allowedHosts`.

`providerEnabled()` (`deck_providers.go:322-324`) only checks that the allowlist is
*non-empty*, not that it contains the host actually about to be dialled. The only thing
constraining which host the *initial* request reaches is `deckProviderAdapterFor`
(`deck_providers.go:412-428`), which consults the hardcoded `deckProviderAdapters` map — today
containing exactly one entry, `moxfieldHost`. The comment on `deckProviderAdapters`
(`deck_providers.go:399-407`) states "the allowlist bounds what the secure client may dial" —
that is not what the code does: the allowlist bounds what a *redirect* may dial to; the initial
dial target is bounded only by the adapter registry and by `safeControl`'s private/reserved
address denial (which does not restrict *public* hosts at all).

There is no exploitable path for an external attacker today, because `deckProviderAdapters` is
compiled into the binary and not influenced by request input or by `DECK_PROVIDER_ALLOWED_HOSTS`.
But the control does not do what its own documentation says: an operator who configures
`DECK_PROVIDER_ALLOWED_HOSTS` believing it is the authoritative list of hosts this codebase can
reach is wrong for the first hop of every request, and the gap becomes live risk the moment a
second adapter is registered without someone separately re-deriving and cross-checking the
initial-request host against the allowlist by hand.

**Fix:** Check the initial request's host against `allowedHosts` before dialing, the same way
`newDeckProviderCheckRedirect` does for later hops — either inside `fetchDeckProviderURL` (which
would need `allowedHosts` added as a parameter) or in `previewDeckURL` before calling
`deckProviderFetch`:

```go
func fetchDeckProviderURL(ctx context.Context, client *http.Client, allowedHosts map[string]struct{}, rawURL string) ([]byte, error) {
	parsed, err := url.Parse(rawURL)
	// ... existing scheme/empty-host checks ...
	host := strings.ToLower(parsed.Hostname())
	if _, ok := allowedHosts[host]; !ok {
		return nil, fmt.Errorf("deck provider: blocked host %q", host)
	}
	// ...
}
```

### WR-02: `server/test.go`'s `testAPI` masks a broken or misconfigured database as a pass, not a failure

**File:** `server/test.go:16-69`

**Issue:** `testAPI` hardcodes `postgres://edhgo:edhgo@localhost:5432/edhgo?...` and never
consults `DATABASE_URL` (or any environment override) at all — a deployment or CI environment
using a different host, port, credentials, or database name for its test Postgres instance has
no way to point `testAPI` at it. Worse, every failure mode along the way —
`ensurePostgresReachable` failing, `ForceCleanMigrations` failing, `NewPostgres` failing — calls
`t.Skipf`, not `t.Fatalf`. A `go test ./server/...` run against a database that is unreachable,
has a permissions problem, or has a broken migration set will report every one of those tests as
skipped rather than failed. Combined with several other packages in this same phase explicitly
documenting that CI only gates `go test ./pkg/... -race` (see `pkg/ratelimit/limiter.go:1-8`,
`pkg/telemetry/metrics.go:1-7`, `pkg/deckimport/sourcetype.go:1-6`), this means the entire
`server` package's correctness — including this review's CR-01 finding above, and every other
resolver-level behavior in this phase — has no CI-enforced safety net at all: a broken database
locally or in a dev environment silently downgrades "tests fail" to "tests skipped," and if
`server` tests are ever run in a CI job that also lacks Postgres, the same silent downgrade
would happen there too, reporting green.

**Fix:** At minimum, read `DATABASE_URL` with a documented, safe local fallback (mirroring
production's own `envconfig` default) so a differently-configured environment is not silently
unreachable:

```go
dsn := os.Getenv("DATABASE_URL")
if dsn == "" {
    dsn = "postgres://edhgo:edhgo@localhost:5432/edhgo?sslmode=disable&connect_timeout=3"
}
cfg := Conf{PostgresURL: dsn, ...}
```
And consider distinguishing "Postgres binary/service simply isn't present in this environment"
(a legitimate `t.Skip`) from "Postgres is present but migrations or connection failed" (which
should `t.Fatal`, since that is much more likely to indicate a real regression than an absent
dependency).

### WR-03: `pkg/telemetry.Collectors.ObserveDeckProviderFetch` accepts unbounded strings, unlike every other typed observation method in the same file

**File:** `pkg/telemetry/metrics.go:335-341`

**Issue:** Every other `Observe*` method on `Collectors` accepts a compiler-enforced enum for
its label-bound parameter — `deckimport.SourceType` (`ObserveDeckImport`), `Role`
(`ObserveBoardActivation`), `ratelimit.Surface` (`ObserveRateLimit`) — specifically so, per this
file's own comments, "a client-controlled value can never reach `WithLabelValues`... because the
compiler requires the enum type." `ObserveDeckProviderFetch(provider, outcome string, ...)`
breaks that pattern: both parameters are plain `string`. The doc comment asserts "provider and
outcome are both small, bounded strings — never a caller-supplied host or error string," but
nothing in the type system enforces that claim the way it is enforced everywhere else in this
file.

This method is not called from any production code path in this phase
(`grep -rn ObserveDeckProviderFetch` outside test files returns only the declaration) — the
`vedh_deck_provider_fetch_total`/`_duration_seconds` families are declared but not yet wired to
an emit site, contrary to the doc comment's "First observed by plan 01-07" (no plan-01-07 code
calls it). There is no live cardinality bug today, but the moment a future author wires this up,
nothing stops a hostname or an error string from being passed as `provider`, silently opening
exactly the unbounded-label hole `pkg/telemetry/metrics_test.go`'s `TestMetrics_SourceLabelValuesAreEnumMembers`
and its siblings are designed to catch for every other family — and no equivalent test exists
for `provider`/`outcome` on this specific counter, because there is nothing enum-typed to test
against yet.

**Fix:** Give `provider` a named, closed type (e.g. a small `DeckProvider` enum with one value
per registered adapter, mirroring `deckimport.SourceType`) before this method is ever called from
production code, and add a `TestMetrics_ProviderLabelValuesAreEnumMembers`-style test alongside
the existing source/role/reason/surface checks.

### WR-04: `documentation vs. implementation mismatch` — `docs/analytics/product-event-vocabulary.md` describes `vedh_deck_provider_fetch_total` as first observed by "plan 01-07," but plan 01-07's shipped code never observes it

**File:** `docs/analytics/product-event-vocabulary.md:145` and `pkg/telemetry/metrics.go:202-213`,
cross-referenced against `server/deck_providers.go`/`server/deck_import.go` (no call site)

**Issue:** The vocabulary doc's metric table states the deck-provider-fetch family is "First
observed" by this plan (01-07 in the doc's row), matching `metrics.go`'s own comment ("First
observed by plan 01-07, once the D-14 checkpoint's selected branch... wires an emit site"). Given
the actual D-14 outcome (Moxfield selected, but its adapter's `normalizeToDeckText` deliberately
never succeeds, and `previewDeckURL` never calls `collectors.ObserveDeckProviderFetch` at any
point in its success or failure paths — see `server/deck_import.go:248-285`), this metric family
still exports no child series after this phase, contradicting the doc/comment's stated
first-observed milestone. This is a low-severity documentation-drift finding, not a functional
defect, since the underlying behavior (declared-but-unobserved) is otherwise handled correctly
and consistently elsewhere in this same file for the Phase 2/3 families.

**Fix:** Update the doc table's "First observed" column for the deck-provider-fetch row (and the
corresponding comment in `metrics.go`) to reflect that this remains unobserved pending a future
task that actually calls `ObserveDeckProviderFetch` from `previewDeckURL`, consistent with how
the doc already correctly describes the guest-session/game-create/game-join/board-activation rows.

## Info

### IN-01: `server/schema.resolvers.go`'s generated `previewDeck`/`trackProductEvent` stubs are unreachable dead code

**File:** `server/schema.resolvers.go:57-65`

**Issue:** gqlgen's generated boilerplate added `mutationResolver.PreviewDeck` and
`mutationResolver.TrackProductEvent` methods that unconditionally `panic("not implemented")`.
These are never invoked: `graphQLServer.Mutation()` (`server/graphql.go:435-437`) returns `s`
itself, and `*graphQLServer` already implements both methods directly
(`server/deck_import.go:186`, `server/product_events.go:151`), which is what the generated
schema actually binds to at runtime. This is standard, low-risk gqlgen scaffold noise rather
than a defect introduced by this phase, and it is consistent with every other resolver stub
already in this file — flagging only for completeness since the file was in scope.

**Fix:** None required; this is expected gqlgen output given this project's pattern of
implementing resolvers directly on `*graphQLServer` rather than on the generated
`mutationResolver`/`queryResolver` wrapper types.

---

_Reviewed: 2026-08-05T21:28:31Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
