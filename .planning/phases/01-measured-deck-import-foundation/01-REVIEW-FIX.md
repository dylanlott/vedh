---
phase: 01-measured-deck-import-foundation
fixed_at: 2026-08-05T22:14:39Z
review_path: .planning/phases/01-measured-deck-import-foundation/01-REVIEW.md
iteration: 1
findings_in_scope: 6
fixed: 6
skipped: 0
status: all_fixed
---

# Phase 01: Code Review Fix Report

**Fixed at:** 2026-08-05T22:14:39Z
**Source review:** .planning/phases/01-measured-deck-import-foundation/01-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 6 (CR-01, CR-02, WR-01, WR-02, WR-03, WR-04 — fix_scope: critical+warning)
- Fixed: 6
- Skipped: 0
- Out of scope by instruction: IN-01 (expected gqlgen scaffold noise, not fixed per orchestrator instruction)

**Isolation:** All edits and commits were made in an isolated git worktree at
`/tmp/sv-01-reviewfix-HvlSLB` on temp branch `gsd-reviewfix/01-55831`, created from
`main` at `a92d8cc`. The cleanup tail (fast-forward `main`, remove worktree, delete
temp branch, remove recovery sentinel) runs immediately after this report is written.

**Verification environment:** Go tests (`go test ./server -count=1`, `go test
./pkg/... -race`, `go mod verify`) ran inside the isolated worktree, pointing at the
same local Postgres (`postgres://edhgo:edhgo@localhost:5432/edhgo`) and the same
MTGJSON snapshot (`/Users/oberon/code/vedh/All Printings.json`, via
`ALL_PRINTINGS_JSON_PATH`) the main checkout uses — these results are reproducible
from either tree. `cd app && npm test` ran in the **main checkout**
(`/Users/oberon/code/vedh/app`) instead, because the worktree has no `node_modules`
(worktrees are not `npm install`-ed) and no file under `app/` was touched by any of
these six fixes — the frontend suite result is identical regardless of which tree it
runs in, so this is not a `main`-only-not-worktree discrepancy, just a practical
constraint that avoided a slow reinstall.

## Fixed Issues

### CR-01: `JoinGame` has no life-total default, so an omitted life silently starts a player "already eliminated"

**Files modified:** `server/games.go`, `server/games_test.go`
**Commit:** `ad7f344`
**Applied fix:** Added a defaulting step in `JoinGame`, mirroring `CreateGame`'s
`defaultLifeForAll` reasoning but adapted to `JoinGame`'s single-player-per-call
shape: `life := input.BoardState.Life; if life <= 0 { life = formatFromRules(game.Rules).StartingLife }`.
`formatFromRules` never returns nil (falls back to `DefaultFormat()`), and
`game.Rules` is already normalized by the preceding `ensureGameDefaults(game)`
call, so the format lookup is always well-formed. Treated `Life <= 0` (not just
`== 0`) as "unspecified" on the reasoning that a negative life on a joining player
is equally nonsensical and should not be preserved verbatim. Added
`TestJoinGame_DefaultsLifeWhenOmitted`, a new regression test that joins a game
with `BoardState.Life` omitted and asserts the resulting player's
`Boardstate.Life` equals `formatFromRules(game.Rules).StartingLife` (40 for the
seeded EDH format), not 0. Verified the new test passes against a real Postgres
instance and the full `go test ./server` suite (77.8s) stayed green afterward.

### CR-02: The rate-limit client key is entirely client-controlled, letting any caller bypass `previewDeck`/`trackProductEvent` limits at will

**Files modified:** `server/ratelimit.go`, `server/ratelimit_test.go`
**Commit:** `115c24e`
**Applied fix:** Rewrote `clientKeyFor` so the server-observed remote address
(`remoteAddrFromContext`, sourced from `server/graphql.go`'s `withClientAddr` →
`r.RemoteAddr`, the TCP peer address — confirmed never derived from a
client-supplied header) is the **sole** bucketing key whenever present;
`sessionID` is consulted only as a fallback when no address is in context.
**Deviation from the review's literal suggested code:** the review's own example
fix (`"addr:" + addr + "|session:" + session`) was tried first and found to
*not* close the hole — concatenating a varying `sessionID` onto a fixed address
still produces a distinct map key per session, so an attacker rotating
`sessionID` from one IP still gets a fresh bucket every time. I verified this
empirically: a test asserting "5 requests with 5 different sessionIDs from the
same address must not grant extra budget" failed under the concatenation
approach and passes under the address-only approach. Implemented the review's
own stated alternative instead ("drop the session-based key entirely in favor
of address-based limiting for these two public surfaces"), since per-session
fairness was never a stated goal for these surfaces. Rewrote
`TestRateLimit_ClientKeyForPrefersSessionOverAddress` (which asserted the
old, vulnerable precedence) as
`TestRateLimit_ClientKeyForAlwaysAnchorsOnAddress`, and added
`TestRateLimit_VaryingSessionIDAloneDoesNotGrantMoreBudget`, which exhausts a
burst-1 registry from one address, then proves five different
never-before-seen `sessionID`s from that same address are all still refused.
Verified against `go test ./server -run TestRateLimit` and the two production
call sites (`deck_import.go`, `product_events.go`) needed no changes since
`clientKeyFor`'s signature is unchanged. Full `go test ./server` suite stayed
green.

### WR-01: The outbound deck-provider host allowlist is enforced only on redirects, never on the initial request

**Files modified:** `server/deck_import.go`, `server/deck_providers.go`, `server/deck_providers_test.go`
**Commit:** `cd5ae3f`
**Applied fix:** Added an `allowedHosts map[string]struct{}` parameter to
`fetchDeckProviderURL`, checked immediately after the scheme/empty-host checks
and before the request is built (so before any dial or name resolution).
Updated the one call site (`previewDeckURL`) to pass
`s.deckProviderAllowedHosts`. Updated `withDeckProviderFetchSpy`'s closure
signature and every direct `fetchDeckProviderURL(...)` test call site
(`TestSafeClient_RejectsBeforeResolution`, `TestSafeClient_BodyCap`'s three
subtests, `TestSafeClient_ZeroByteBodyIsProviderError`) to supply the new
argument — added an `allowedHostsForServer(t, srv)` test helper that derives an
allowlist from an `httptest.Server`'s own loopback host so those hermetic tests
keep exercising body-cap/timeout/empty-body behavior rather than being blocked
by the new host check. Added
`TestSafeClient_InitialRequestHostEnforced`, a new regression test following
the existing `RedirectToUnallowlistedHostNeverInvoked` pattern: a target
server's handler sets an `atomic.Bool` if invoked; `fetchDeckProviderURL` is
called directly with an allowlist that deliberately excludes the target's
host; the test asserts an error containing "blocked host" and that the
handler was never invoked — proving zero dial attempts on the *initial*
request, not merely that an error was returned. Full `go test ./server`
suite (all `TestSafeClient*`/`TestSafeControl*`/`TestProvider*` and the full
package) verified green afterward.

### WR-02: `server/test.go`'s `testAPI` masks a broken or misconfigured database as a pass, not a failure

**Files modified:** `server/test.go`
**Commit:** `c22e8c5`
**Applied fix:** `testAPI` now reads `DATABASE_URL` with a fallback to the
previous hardcoded DSN (extracted to a named `testAPIDefaultDSN` constant).
Split the three failure branches per the orchestrator's explicit split: a
plain TCP-reachability failure (`ensurePostgresReachable`) remains `t.Skipf`
— this codebase documents `go test ./server/...` (the `make test-api` target)
as NOT part of the CI-gated workflow (`.github/workflows/test.yml`'s "Test"
job runs only `make test-unit`, i.e. `go test ./pkg/... -race`; confirmed by
reading the workflow file directly), and README.md's "Backend integration
tests" section documents this target as requiring a local Postgres — so "no
Postgres running at all" is a documented, supported developer-machine state.
Once that check has passed, Postgres is confirmed present and listening, so a
subsequent `ForceCleanMigrations` or `NewPostgres` failure now uses
`t.Fatalf` instead of `t.Skipf` — those indicate a real regression (bad
credentials, corrupted migrations, permissions), not an absent dependency.
**Verification limitation, noted honestly:** `server/main_test.go`'s
pre-existing `TestMain` already performs its own DB-reachability/migration
setup and calls `os.Exit(1)` on any failure, gating the entire test binary
before any individual test's `testAPI()` call runs. This means I could not
empirically reproduce the "Postgres reachable but migrations/connection
broken" branch live in this environment — doing so would require either a
Postgres instance reachable but misconfigured in a way `TestMain` itself
tolerates (none exists) or refactoring `TestMain`, which is out of scope for
this finding (`server/test.go:16-69` only). I verified this change by code
reading (the three branches are correctly ordered and match the reviewer's
stated "reachable vs. not reachable" distinction) and by confirming the full
`go test ./server` suite stays green with `DATABASE_URL` both unset and
explicitly set to the same DSN.

### WR-03: `pkg/telemetry.Collectors.ObserveDeckProviderFetch` accepts unbounded strings, unlike every other typed observation method in the same file

**Files modified:** `pkg/telemetry/metrics.go`, `pkg/telemetry/metrics_test.go`
**Commit:** `0e8eff0`
**Applied fix:** Added a new `DeckProvider` named string type in
`pkg/telemetry`, with one constant (`DeckProviderMoxfield`) mirroring the "one
constant per adapter actually registered" discipline `ratelimit.Surface`
documents for itself (only `moxfieldHost` is registered in
`server/deck_providers.go`'s `deckProviderAdapters` today), plus an
`AllDeckProviders()` helper mirroring `deckimport.AllSourceTypes()`. Changed
`ObserveDeckProviderFetch`'s and `DeckProviderFetchCounter`'s `provider`
parameter from `string` to `DeckProvider`. Updated the one test call site
(`exerciseAndGather`, which previously passed the literal `"archidekt"` — a
provider with no registered adapter at all) to pass `DeckProviderMoxfield`
instead. Added `TestMetrics_ProviderLabelValuesAreEnumMembers`, matching the
existing `TestMetrics_SourceLabelValuesAreEnumMembers` pattern, asserting every
`provider`-labeled sample on any `vedh_`-prefixed family is a member of
`AllDeckProviders()`. Confirmed via `grep` that no production code calls
`ObserveDeckProviderFetch` (matching the review's own finding), so this is a
purely additive type-safety change with zero production call-site impact.
Verified with `go test ./pkg/... -race` (full package) and `go build ./...`.

### WR-04: documentation vs. implementation mismatch — `docs/analytics/product-event-vocabulary.md` describes `vedh_deck_provider_fetch_total` as first observed by "plan 01-07," but plan 01-07's shipped code never observes it

**Files modified:** `docs/analytics/product-event-vocabulary.md`, `pkg/telemetry/metrics.go`
**Commit:** `adc8830`
**Applied fix:** **Adapted from the review's literal suggestion.** The review
cited `docs/analytics/product-event-vocabulary.md:145` as containing the
drifted "First observed by plan 01-07" text for the deck-provider-fetch
family — but on inspection, the "Metric families and where they are observed"
table (lines 143–151 as reviewed) has **no row at all** for
`vedh_deck_provider_fetch_total`/`vedh_deck_provider_fetch_duration_seconds`;
line 145 is the unrelated "Deck import" row. I confirmed this against both the
worktree and the main checkout at the same commit the review was run against,
so this is not a case of the file having changed since review — the cited
row simply does not exist in this file. The `pkg/telemetry/metrics.go`
comment half of this finding, however, IS real and verified: lines 202-213
(pre-fix) say "First observed by plan 01-07" for both collectors, and
`git log` confirms `deckProviderFetchTotal`/`deckProviderFetchDuration` were
actually declared by plan **01-06** (`f1ea608`), not 01-07, and `grep -rn
ObserveDeckProviderFetch` outside test files still returns nothing, so
"first observed" is inaccurate on both counts (wrong plan number, and not
actually observed at all). I fixed both real problems: (1) reworded the
`metrics.go` comments and `Help` strings to say "Declared by plan 01-06; not
yet observed by any emit site" instead of claiming a false "first observed"
milestone; (2) added a new table row to the vocabulary doc for the
deck-provider-fetch family (which the doc's own `AllowedLabelNames` cross-
reference at the bottom already implies should exist, since it lists
`provider` as an allowed label with no corresponding table row) marked "Not
yet observed," plus a paragraph explaining why — unlike the four
Phase-2/3-scheduled families already documented as unobserved, this one has
no scheduled landing milestone at all, pending an authorized Moxfield
response sample. This satisfies the finding's actual intent (doc accurately
reflects implementation reality) via addition rather than correction of a
nonexistent row. Verified with `go build ./pkg/...`, `go test ./pkg/... -race`,
and a re-read of both files.

## Skipped Issues

None — all 6 in-scope findings (CR-01, CR-02, WR-01, WR-02, WR-03, WR-04) were fixed.

IN-01 (`server/schema.resolvers.go`'s generated `previewDeck`/`trackProductEvent`
stubs are unreachable dead code) was explicitly excluded from `fix_scope` by the
orchestrator's instructions (expected gqlgen scaffold noise, not a defect) and was
not attempted.

## Full Verification Run (after all 6 fixes)

- `go build ./...` — clean
- `go test ./server -count=1` — `ok` (77.8s), zero failures
- `go test ./pkg/... -race` — `ok` across `deckimport`, `games`, `ratelimit`, `telemetry`
- `go mod verify` — all modules verified
- `cd app && npm test` (run in the main checkout, not the worktree — see
  "Verification environment" above) — 28 passed, 3 skipped (pre-existing,
  unrelated to these fixes), 0 failed
- `go vet ./...` — clean on every touched package

No test was weakened, disabled, or skipped to make the suite pass. Two new
production-behavior regression tests were added (`TestJoinGame_DefaultsLifeWhenOmitted`,
`TestRateLimit_VaryingSessionIDAloneDoesNotGrantMoreBudget`), one existing test was
rewritten because it asserted the vulnerable behavior CR-02 removed
(`TestRateLimit_ClientKeyForPrefersSessionOverAddress` →
`TestRateLimit_ClientKeyForAlwaysAnchorsOnAddress`), and one new regression test each
was added for WR-01 (`TestSafeClient_InitialRequestHostEnforced`) and WR-03
(`TestMetrics_ProviderLabelValuesAreEnumMembers`).

---

_Fixed: 2026-08-05T22:14:39Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
