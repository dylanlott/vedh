---
phase: 01-measured-deck-import-foundation
plan: 05
subsystem: api
tags: [postgresql, pg_trgm, gist, prometheus, promauto, deck-import, gqlgen]

# Dependency graph
requires:
  - phase: 01-measured-deck-import-foundation (plan 01)
    provides: "server/deck_import.go's PreviewDeck resolver/buildDeckPreview, the DeckSuggestion/DeckImportIssue GraphQL types this plan populates, and pkg/telemetry's Collectors/promauto registration site this plan extends"
  - phase: 01-measured-deck-import-foundation (plan 02)
    provides: "pkg/deckimport.Parse/ParsedDeck/ParsedEntry — the canonical grammar createLibraryFromDecklist now consumes directly instead of raw text"
  - phase: 01-measured-deck-import-foundation (plan 04)
    provides: "card_names/card_names_trgm_gist (the GiST trigram projection this plan's suggestion query reads), resolveDeckEntries (the shared D-09 resolver both preview and library creation now call), and lowConfidenceCutoff/isLowConfidence"
provides:
  - "server/deck_import.go's suggestFor: D-02/D-03/D-04's bounded eager suggestion lookup, with maxSuggestionNeedles/suggestionBudget/suggestionsPerEntry as the three instrumented bounds and dedupeNeedles as the shared bound-#2 helper"
  - "pkg/telemetry's vedh_deck_suggestion_duration_seconds (by outcome) and vedh_deck_suggestion_truncated_total, both exercised by the shared exerciseAndGather"
  - "server/games.go's createLibraryFromDecklist rebuilt to accept *deckimport.ParsedDeck (never raw text) with resolution and the commander-budget removal both preceding the 100-minus-commanders size check"
  - "server/deck_import.go's applyPrintingMetadata — D-07/D-08's printing metadata carried onto every persisted library Card"
  - "fix(01-05): removed a pre-existing, non-Phase-1 regression where ensureFormatRules silently reset any already-0-life player back to full life on every game load, and where AdvancePhase's normalizeTurnPhase call silently discarded any caller-supplied phase name outside the format's fixed PhaseSequence"
  - "server/test.go's testAPI now closes its *sql.DB pool via t.Cleanup, fixing a full-suite 'too many clients' failure this plan's own added tests exposed"
affects: [01-06 (rate limiting sits in front of previewDeck; the provider adapter's own resolution, if any, should follow resolveDeckEntries's shape rather than re-deriving it), Phase 2/3 (guest join and invite-preview flows call createLibraryFromDecklist's new signature)]

# Actuals (#2632)
actuals:
  tokens: 11527
  tasks: 2
  commits: 3

tech-stack:
  added: []
  patterns:
    - "budget-as-parameter for deterministic timeout testing: buildDeckPreviewWithSuggestionBudget/suggestFor take the sub-budget as an explicit parameter (production always passes the suggestionBudget constant) so a test can force expiry without racing a real query"
    - "KNN index usage is fragile to a second ORDER BY key: pg_trgm's GiST '<->' operator only accelerates a LATERAL's LIMIT when the ORDER BY is the distance expression alone — adding a display-name tiebreak inside the LATERAL silently downgrades it to a sequential scan plus a top-N sort; the tiebreak belongs in the cheap outer sort over the already-limited result set instead"
    - "defaultLifeForAll: when a required (non-pointer) field's zero value is structurally ambiguous between 'caller didn't set this' and 'caller explicitly wants zero', decide at the batch level (did ANY sibling in this call specify a nonzero value?) rather than per-field, when the alternative would make a legitimate zero value impossible to ever express"

key-files:
  created: []
  modified:
    - server/deck_import.go
    - server/deck_import_test.go
    - server/games.go
    - server/games_test.go
    - server/formats.go
    - server/test.go
    - pkg/telemetry/metrics.go
    - pkg/telemetry/metrics_test.go

key-decisions:
  - "The suggestion query's per-needle LATERAL orders by the GiST distance operator alone, never with a secondary tiebreak column — measured directly against the real ~35,831-row card_names table: adding cn.display ASC inside the LATERAL turned a 7.5ms index-ordered KNN scan into a 47ms sequential-scan-plus-sort, and the 25-needle worst case from ~190ms into ~1.2s, blowing the 750ms sub-budget. The tiebreak instead lives in the outer ORDER BY, over the already-limited ≤75-row result set, where it costs nothing."
  - "applyPrintingMetadata overwrites a persisted library Card's SetCode/CollectorNumber/Category with what the player actually typed (deckimport.ParsedEntry's D-07 fields) whenever the entry specifies them, rather than leaving resolveDeckEntries's resolved-printing values in place. This mirrors buildDeckPreview's existing 'carry the parsed value through regardless of resolution outcome' choice for DeckPreviewEntry, and follows D-10's own stated purpose (reconciliation on a future MTGJSON refresh or proxy-print feature) through to the persisted Card, not only the preview."
  - "createLibraryFromDecklist's commander-budget block and 100-minus-commanders comparison are moved verbatim (same logic, same order of operations relative to each other) — only their position relative to resolution changed, per Pitfall 5. The only textual differences are the field rebinding from a `{name, qty}` deckEntry to a `{card, qty}` one and an added (idempotent) strings.TrimSpace on the entry name, which the parser already trims before this function ever sees it."
  - "TestDeckImport_SingleParse deliberately passes no commander selection: previewDeck takes no commander input at all, so buildDeckPreview's CardCount and createLibraryFromDecklist's length are only guaranteed to agree when there is no commander-budget removal to make them diverge. This is the literal shape of the must_have truth ('a hundred-card paste... yields... the preview's card count equals the created library's length for the same input'), which likewise does not mention commanders."

patterns-established:
  - "seedCardNamesRow (server/deck_import_test.go): a direct card_names insert/cleanup helper for suggestion tests that need exact control over candidate scores, distinct from seedPrintingTestCards (plan 01-04, which seeds the cards table for exact-match/printing tests). Two pairs of these synthetic rows were empirically verified (via psql) to produce exact pg_trgm score ties before being hardcoded into the tiebreak test, rather than assumed."

requirements-completed: [REQ-ACT-002]

coverage:
  - id: D1
    description: "Every unresolved card carries up to three ranked candidates in the same previewDeck response, the below-cutoff case still returns the nearest matches marked low-confidence rather than an empty list, an equal-score tie between two candidates breaks on the displayed name ascending, and a repeated misspelling collapses to one needle before any query runs"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/deck_import_test.go#TestDeckImport_Suggestions"
        status: pass
    human_judgment: false
  - id: D2
    description: "Exactly 25 distinct unresolved needles produces no truncation warning and no counter movement; 26 produces exactly one warning naming both counts and exactly one increment of vedh_deck_suggestion_truncated_total"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/deck_import_test.go#TestDeckImport_SuggestionTruncation"
        status: pass
    human_judgment: false
  - id: D3
    description: "A suggestion query whose sub-budget has already expired returns zero suggestions plus one warning, and the surrounding preview still succeeds with the same CardCount and CanContinue as the same parsed input run with the real budget"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/deck_import_test.go#TestDeckImport_SuggestionTimeoutDegrades"
        status: pass
    human_judgment: false
  - id: D4
    description: "An unresolved entry counts toward neither the deck-size check nor the created library (a 100-card paste with 3 unmatched names yields a 97-card library and a 97 preview count), and a 101-card paste with 2 unmatched names is accepted because resolution now precedes the size check"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/deck_import_test.go#TestDeckImport_UnresolvedAccounting"
        status: pass
    human_judgment: false
  - id: D5
    description: "Exactly 100-minus-commanders resolved cards is accepted; one more is a blocking error naming both the actual and maximum count"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/deck_import_test.go#TestDeckImport_SizeBoundary"
        status: pass
    human_judgment: false
  - id: D6
    description: "Two rows naming the same card contribute their summed quantity to the library while remaining two separately-parsed entries, and a card named both as a selected commander and in the main deck has exactly the commander count removed from its quantity with the remainder kept"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/deck_import_test.go#TestDeckImport_DuplicateRowsAndCommanderOverlap"
        status: pass
    human_judgment: false
  - id: D7
    description: "The preview's card count equals the created library's length for the same parsed value (no commander selection, since previewDeck takes none) — the library is built from a parsed deck value, never from raw text, so it cannot be parsed a second time under different rules"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/deck_import_test.go#TestDeckImport_SingleParse"
        status: pass
      - kind: automated
        ref: "! grep -q 'encoding/csv' server/games.go (exits 0) && grep -q 'deckimport.ParsedDeck' server/games.go"
        status: pass
    human_judgment: false
  - id: D8
    description: "Assigned test fix: go test ./server -count=1 and go test ./server -race -count=1 both pass with zero failures — the three pre-existing format-registry failures are resolved (one stale test expectation, two real product regressions), and a full-suite connection-exhaustion failure this plan's own added tests exposed is fixed"
    verification:
      - kind: integration
        ref: "go test ./server -count=1"
        status: pass
      - kind: integration
        ref: "go test ./server -race -count=1"
        status: pass
    human_judgment: false

duration: ~55min
completed: 2026-08-04
status: complete
---

# Phase 1 Plan 5: Bounded Eager Suggestions and the Single Canonical Library Build Summary

**`previewDeck` now returns up to three ranked, GiST-KNN-ranked suggestions per unresolved card with an instrumented 25-needle/750ms/3-candidate bound, and `createLibraryFromDecklist` is rebuilt to consume the same parsed, resolved deck the preview reported — closing REQ-ACT-002 — plus a separately-committed fix for three pre-existing `server` test failures and a connection-exhaustion failure this plan's own tests exposed.**

## Performance

- **Duration:** ~55 min
- **Started:** 2026-08-04 (continuation from plan 01-04's completion, tip `becb880`)
- **Completed:** 2026-08-04T23:45:00-06:00 (approx.)
- **Tasks:** 2/2, plus one orchestrator-assigned fix
- **Files modified:** 8 (0 created, 8 modified)

## Accomplishments

- `server/deck_import.go` gains `suggestFor`, D-02/D-03/D-04's bounded eager suggestion lookup: one batched, parameter-bound query per preview over the deduplicated unresolved-name set, laterally joined against `card_names` (plan 01-04's trigram projection) ordered by the GiST `<->` distance operator alone — always returning the three nearest candidates, even below `lowConfidenceCutoff`, with the cutoff applied in Go via `isLowConfidence`. Three named, doc-commented, instrumented bounds (`maxSuggestionNeedles=25`, `suggestionBudget=750ms`, `suggestionsPerEntry=3`) cap the worst case at a constant regardless of deck size.
- `pkg/telemetry` gains `vedh_deck_suggestion_duration_seconds` (by outcome) and `vedh_deck_suggestion_truncated_total`, both exercised by the shared `exerciseAndGather` so the existing label-allowlist and bounded-value tests cover them for free.
- **Found and fixed a real query-plan bug while writing the SQL**: a display-name tiebreak added inside the per-needle `LATERAL`'s `ORDER BY` defeated the GiST index (`EXPLAIN ANALYZE` showed a sequential scan plus a top-N sort, 47ms/needle instead of 7.5ms), which alone pushed the 25-needle worst case from ~190ms to ~1.2s and blew the 750ms sub-budget — caught by `TestDeckImport_SuggestionTruncation` actually failing against the real, already-imported ~35,831-row `card_names` table, not a synthetic one. Moving the tiebreak to the outer sort (over the already-limited ≤75-row result set) restored the index scan and fixed the timing.
- `server/games.go`'s `createLibraryFromDecklist` now takes `*deckimport.ParsedDeck` instead of raw text, removing `encoding/csv` (and the `io`/`strconv` it pulled in) from the file entirely. Both production call sites (`JoinGame`, `CreateGame`) parse once and pass the result — a deck can no longer be parsed twice under divergent rules (Pitfall 6/D-06).
- The function body is resequenced per Pitfall 5: resolve via the shared `resolveDeckEntries` helper `previewDeck` already uses → drop unresolved entries from both the countable set and the library (deleting the branch that used to substitute a bare `&Card{Name:}` for an unmatched name and report nothing, D-05/D-06) → apply the commander-budget removal (moved verbatim) → compare against the `100 - commanders` cap. Resolution now precedes the size check, so a 101-card paste with 2 unmatched names is accepted instead of rejected for counting all 101 rows before lookup ran.
- Every surviving entry's card copies carry D-07/D-08's printing metadata (`SetCode`/`CollectorNumber`/`Category` from what the player actually typed, `SourceFormat` from the detected paste shape) via the new `applyPrintingMetadata` helper — persisted for free since `games.payload` is JSONB.
- **Assigned fix, committed separately**: resolved all three named pre-existing `server` test failures (one stale test expectation in `TestCreateGame`, two real product regressions in `TestUpdateBoardState_AutoFinishRules` and `TestAdvancePhase`), plus a full-suite Postgres connection-exhaustion failure (`pq: sorry, too many clients already`) that this plan's own ~10 added test functions tipped over — confirmed absent from the identical suite run against the pre-01-05 tree. `go test ./server -count=1` and `go test ./server -race -count=1` are both green with zero failures.

## Task Commits

Each task was committed atomically:

1. **Task 1: Bounded eager suggestions with instrumented limits** - `456150c` (feat)
2. **Task 2: One parsed deck feeds both the preview and the created library** - `cb362f6` (feat)
3. **Assigned deviation: stale format-registry test expectations and two real regressions** - `488b9e1` (fix) — committed separately from the two plan tasks, as directed.

**Plan metadata:** pending (this commit, immediately following)

## Files Created/Modified

- `server/deck_import.go` — `suggestFor`, `dedupeNeedles`, `suggestionQuery`, `maxSuggestionNeedles`/`suggestionBudget`/`suggestionsPerEntry`, `buildDeckPreviewWithSuggestionBudget` (with `buildDeckPreview` now a thin wrapper), `applyPrintingMetadata`
- `server/deck_import_test.go` — `seedCardNamesRow`; `TestDeckImport_Suggestions` (4 subtests), `TestDeckImport_SuggestionTruncation` (2 subtests), `TestDeckImport_SuggestionTimeoutDegrades`, `TestDeckImport_UnresolvedAccounting` (2 subtests), `TestDeckImport_SizeBoundary` (2 subtests), `TestDeckImport_DuplicateRowsAndCommanderOverlap`, `TestDeckImport_SingleParse`
- `server/games.go` — `createLibraryFromDecklist` rebuilt around `*deckimport.ParsedDeck`; both production call sites; `AdvancePhase`'s removed `normalizeTurnPhase` call; `CreateGame`'s `defaultLifeForAll` heuristic
- `server/games_test.go` — three pre-existing decklist-test call sites updated to the new signature; `starting_life` rule added to `TestCreateGame`'s three `want.Rules`
- `server/formats.go` — `ensureFormatRules`'s per-load `Turn.Phase` re-normalization and per-player `Life`-zero re-defaulting removed
- `server/test.go` — `testAPI` closes its `*sql.DB` pool via `t.Cleanup`
- `pkg/telemetry/metrics.go` — `deckSuggestionDuration`/`deckSuggestionTruncated` collectors, `ObserveDeckSuggestion`/`IncDeckSuggestionTruncated`/`DeckSuggestionTruncatedCounter`
- `pkg/telemetry/metrics_test.go` — `exerciseAndGather` extended to observe the two new collectors

## Decisions Made

- The suggestion query's per-needle `LATERAL` orders by the GiST distance operator alone; the display-name tiebreak lives only in the cheap outer sort. See key-decisions above for the measured before/after.
- `applyPrintingMetadata` overwrites the persisted library `Card`'s printing fields with what the player typed (when specified), not the resolved printing — extending D-10's reconciliation intent from the preview to the created library.
- `createLibraryFromDecklist`'s commander-budget block and size comparison are moved verbatim; only their position relative to resolution changed.
- `TestDeckImport_SingleParse` deliberately uses no commander selection, since `previewDeck` takes none — the only case where its `CardCount` and the library's length are guaranteed to agree.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Suggestion query's tiebreak column defeated the GiST index and blew the 750ms sub-budget**
- **Found during:** Task 1, running `TestDeckImport_SuggestionTruncation` against the real, already-imported `card_names` table
- **Issue:** The plan's own reference SQL (RESEARCH.md Pattern 2/3) orders the per-needle `LATERAL` by the distance operator alone; I added `cn.display ASC` as a second key to make ties deterministic at the `LIMIT` boundary. `EXPLAIN ANALYZE` showed this downgraded the plan from an index-ordered KNN scan (7.5ms) to a sequential scan plus a top-N sort (47ms) per needle — for 25 needles, ~1.2s total, exceeding the 750ms budget and making the "exactly 25 needles, no warning" acceptance criterion fail even at exactly the cap.
- **Fix:** Removed the tiebreak from the `LATERAL`'s `ORDER BY`; it now orders by the distance operator alone (restoring the index scan) and the tiebreak lives only in the outer `ORDER BY` over the already-limited ≤75-row result set, where it is free.
- **Files modified:** `server/deck_import.go`
- **Verification:** `TestDeckImport_SuggestionTruncation` passes at both 25 and 26 needles; `TestDeckImport_Suggestions`'s equal-score-tiebreak subtest still passes (the outer sort still produces the deterministic order).
- **Committed in:** `456150c` (Task 1 commit)

---

**2. [Rule 3 - Blocking] `go test ./server -count=1` hit Postgres connection exhaustion, unrelated to the three assigned failures**
- **Found during:** running the assigned fix's "zero failures" verification
- **Issue:** `testAPI(t)` opens a fresh `*sql.DB` connection pool on every call and nothing ever closed it. This plan added roughly 10 new test functions (several with subtests), each calling `testAPI` at least once; a full, unfiltered `go test ./server` run hit `pq: sorry, too many clients already` partway through. Confirmed via a temporary `git worktree` at the pre-01-05 tip (`becb880`) that the identical full-suite run there produces exactly the three known failures and zero connection errors — this is not one of the three assigned, pre-existing failures, and it is caused by this plan's own additions.
- **Fix:** Added `t.Cleanup(func() { _ = appDB.Close() })` in `testAPI`, right after the pool is created. Safe because this package's test logger discards all output, so a background goroutine that outlives its test and touches the closed pool afterward logs into the void rather than panicking or failing a completed test.
- **Files modified:** `server/test.go`
- **Verification:** `go test ./server -count=1` and `go test ./server -race -count=1` both pass with zero failures, run three times each without a single connection error.
- **Committed in:** `488b9e1` (assigned-fix commit)

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking issue), plus the orchestrator-assigned test fix documented below (not a deviation from *this* plan's own tasks, but tracked the same way per the assignment's instruction).

**Impact on plan:** Both fixes were required for this plan's own acceptance criteria and verification block to pass at all. No scope creep — the query-plan fix corrects code I wrote in this plan; the connection-cleanup fix corrects a latent defect this plan's own added tests were what exposed.

## Assigned Test Fix

Per the orchestrator's assignment, three pre-existing `server` test failures (dated to commit `b1ac894`, 2026-05-18, verified pre-existing and unrelated to Phase 1 via a clean run against `1ee003d`) were investigated and resolved, judging test-vs-code for each rather than blindly updating expectations:

1. **`TestCreateGame`'s three subtests — stale test expectation, code is correct.** `ensureFormatRules` (added by `b1ac894`) deliberately upserts a third `starting_life` rule alongside `format`/`deck_size`. The subtests' `want.Rules` literals were simply never updated to expect it. **Fix:** added `{Name: "starting_life", Value: "40"}` to all three `want.Rules` slices. `TestJoinGame`'s `want.Rules` has the identical staleness but only ever reaches `t.Logf`, never `t.Errorf` — it does not fail, so per the "zero failures" (not "zero staleness") mandate it was left alone.

2. **`TestUpdateBoardState_AutoFinishRules`'s multiplayer subtest — a real product regression, not a stale assertion.** `ensureFormatRules` ran a per-player `Life == 0 → format.StartingLife` default inside a function called by `ensureGameDefaults` on *every* load of an *existing* game (`JoinGame`, `UpdateBoardState`, `AdvancePhase`, `GetGame`, `UpdateGame`, `win_claim` — not only at creation). Concretely: a player who legitimately reached 0 life during play had it silently reset to full life the next time *anyone* called `UpdateBoardState`, *before* `alivePlayerNames` even ran — this undermines auto-finish for real gameplay, not only this test's setup. Traced the same defect to `Turn.Phase`: `ensureFormatRules` also called `normalizeTurnPhase` on every load, which is what caused failure 3 below. **Fix:** removed both blocks from `ensureFormatRules` entirely — it has no business re-applying "new game" defaults to a game that already has players and an in-progress turn. `CreateGame` still needs a rule for "`Life == 0` means unspecified" so `TestGameFormatDefaults`/`TestCreateGame_GenericDuelFormatRoundTrips` keep passing, but `InputBoardState.Life` is a required, non-pointer `Int` in the GraphQL schema — there is no wire-level way to tell "the caller didn't set this" from "the caller explicitly wants 0" per player. Replaced the per-player check with `defaultLifeForAll`: default every player's life only when *nobody* in this `CreateGame` call specified a nonzero life at all. Every existing passing test specifies either all-zero (wanting the default) or all-nonzero (specific totals) life across its players, so this is behavior-preserving for all of them and the actual fix for the one that was failing.

3. **`TestAdvancePhase` — also a real product regression.** `AdvancePhase` called `normalizeTurnPhase`, which coerces any phase name not literally present in the format's registered `PhaseSequence` back to `PhaseSequence[0]` — silently discarding `"COMBAT"`, `"END STEP"`, `"DISCARD"` (none of which are EDH `PhaseSequence` members) and returning `"pregame"` every time, regardless of what the player actually advanced to. `TestAdvancePhase` predates `b1ac894` by three months and uses exactly these free-form phase names; PROJECT.md is explicit that this tracker does not enforce turn structure ("Full Magic rules enforcement... not what a tracker is for"). **Fix:** removed the `normalizeTurnPhase` call from `AdvancePhase`. `normalizeTurnPhase`'s legitimate job — picking a sane initial phase for a brand-new game when none or an invalid one was supplied — still runs once, in `CreateGame`, where `TestCreateGame_GenericDuelFormatRoundTrips`'s `Phase: "invalid"` → `"draw"` case continues to pass.

Also investigated per the assignment's "worth knowing" note: `server/test.go`'s `testAPI` hardcodes the `edhgo` DSN (ignores `DATABASE_URL`) and calls `t.Skipf` on every DB failure mode, which is exactly how the three failures above stayed hidden from a misconfigured-database run (a green suite that is actually all-skipped looks identical to a green suite that actually ran). Not fixed, per the assignment's explicit "not required" — but a small, safe proposal: change every `t.Skipf` in `ensurePostgresReachable`'s and `persistence.ForceCleanMigrations`'s failure branches to `t.Fatalf` when an explicit `DATABASE_URL` (or `EDH_TEST_REQUIRE_DB=1`) environment variable is set, defaulting to today's skip-on-failure behavior otherwise. This would let CI (or a developer who explicitly wants strict behavior) opt into "a missing/misconfigured database fails loudly" without changing local-dev ergonomics for anyone who has not opted in.

## Issues Encountered

None beyond the two auto-fixed deviations and the three assigned-fix investigations above.

## User Setup Required

None — no external service configuration required. PostgreSQL 14 and the MTGJSON snapshot required by `server/main_test.go`'s `TestMain` were already available in the execution environment.

## Next Phase Readiness

- REQ-ACT-002 is now fully satisfied across plans 01-02 (grammar), 01-01 (contract), 01-04 (printing resolution), and this plan (suggestions + single canonical library build) — marked complete in `REQUIREMENTS.md`.
- `resolveDeckEntries` and `suggestFor` are both self-contained, ctx-aware helpers `server/deck_import.go` owns; plan 01-06's provider adapter (if it needs its own resolution) should call these rather than re-deriving them.
- `server/games.go` no longer has any raw-text decklist parsing path at all — `createLibraryFromDecklist` structurally cannot diverge from `previewDeck`'s parse.
- No blockers. `go build ./...`, `go vet ./server ./pkg/telemetry`, `go test ./pkg/... -race`, `go test ./server -count=1`, and `go test ./server -race -count=1` are all green.

---
*Phase: 01-measured-deck-import-foundation*
*Completed: 2026-08-04*

## Self-Check: PASSED

All 8 files listed under `key-files` (modified) exist on disk, and all three commit hashes
(`456150c`, `cb362f6`, `488b9e1`) are present in `git log --oneline --all`.
