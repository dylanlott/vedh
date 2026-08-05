---
phase: 01-measured-deck-import-foundation
plan: 01
subsystem: api
tags: [gqlgen, graphql, postgresql, prometheus, promauto, product-analytics, telemetry]

# Dependency graph
requires: []
provides:
  - "GraphQL mutations previewDeck and trackProductEvent, plus DeckPreview/DeckPreviewEntry/DeckImportIssue/DeckSuggestion types and InputDeckImport/InputProductEvent/InputProductEventMeta inputs — the complete Phase 1 contract; no later plan in this phase edits server/schema.graphql"
  - "product_events table with two funnel indexes and the PG14 COALESCE partial-unique dedup index, mirrored byte-identically into persistence/migrations_test/"
  - "pkg/telemetry: closed 15-event vocabulary (D-20) with per-event metadata key allowlists (D-21), ValidateEvent, and all 12 collectors for the four ROADMAP criterion-4 families plus the two product-event counters"
  - "pkg/deckimport: ParsedDeck/ParsedEntry/Warning/BlockingError/Section value types, the D-23/D-24 SourceType enum, and the minimal `<digits><space><name>` scanner"
  - "server/product_events.go's never-fail recordProductEvent + insertProductEvent, and server/deck_import.go's PreviewDeck resolver — the single canonical deckimport.Parse call site"
  - "withScratchMigrationDB (server/product_events_test.go) — reusable scratch-database helper for migration up/down/up proofs, reserved for plan 01-04 task 1's reuse"
  - "exerciseAndGather (pkg/telemetry/metrics_test.go) — shared private-registry Prometheus exercise-then-gather helper, reserved for plan 01-06's extension"
  - "docs/analytics/product-event-vocabulary.md and docs/analytics/product-event-funnel-example.sql — the documented vocabulary with per-key provenance and the ACT-001 example funnel query"
affects: [01-02 (pkg/deckimport expansion), 01-04 (card-name search migration reuses withScratchMigrationDB), 01-06 (rate limiting and provider adapter extend exerciseAndGather and the metrics/vocabulary files), Phase 2 ACT-005/ACT-006 and Phase 3 ACT-008/ACT-009 (first observers of the declared-but-unobserved collector families)]

# Actuals (#2632)
actuals:
  tokens: 53258
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Never-fail product-event write mirroring server/gamelog_helpers.go's logEvent — validate, then write, and on any rejection or DB error increment a bounded-reason drop counter and return without touching the database or failing the caller"
    - "Named, reusable within-package test helpers (withScratchMigrationDB, exerciseAndGather) declared once and reused across plans/tasks rather than re-derived inline"
    - "Prometheus collectors as typed observation methods (Collectors.Observe*) rather than exposed vectors, so a client-controlled string can never reach WithLabelValues — the compiler is the control"

key-files:
  created:
    - persistence/migrations/20260804120000_product_events.up.sql
    - persistence/migrations/20260804120000_product_events.down.sql
    - persistence/migrations_test/20260804120000_product_events.up.sql
    - persistence/migrations_test/20260804120000_product_events.down.sql
    - pkg/deckimport/result.go
    - pkg/deckimport/scanner.go
    - pkg/deckimport/sourcetype.go
    - pkg/telemetry/vocabulary.go
    - pkg/telemetry/metrics.go
    - pkg/telemetry/vocabulary_test.go
    - pkg/telemetry/metrics_test.go
    - server/product_events.go
    - server/product_events_test.go
    - server/metrics.go
    - server/deck_import.go
    - server/deck_import_test.go
    - docs/analytics/product-event-vocabulary.md
    - docs/analytics/product-event-funnel-example.sql
  modified:
    - server/schema.graphql
    - server/generated.go
    - server/models_gen.go
    - server/schema.resolvers.go
    - app/src/types/generated.ts
    - .planning/phases/01-measured-deck-import-foundation/01-VALIDATION.md

key-decisions:
  - "previewDeck placed on Mutation (never Query), per the locked api-contract — the writer of the schema block and this plan's own verify block both assert its absence from type Query."
  - "DeckPreview.BlockingErrors is a deliberate additive extension beyond the locked contract block; recorded as a dated addendum in .planning/intel/constraints.md and in ROADMAP Phase 1's Plans block, not silently added here."
  - "The dedup index's COALESCE wrappers are load-bearing on PostgreSQL 14 (no UNIQUE NULLS NOT DISTINCT until PG15); TestProductEvents_AuthoritativeDedup's both-NULL-keys subtest is the one that would catch a future 'simplification' removing them."
  - "Two senses of 'authoritative' are deliberately kept distinct and both documented: EventSpec.Authoritative marks the six server-owned event names; the migration's dedup predicate names a different four-member set. Identifiers keep the older, ambiguous word for continuity with 01-RESEARCH.md/01-VALIDATION.md rather than trading it for a stale cross-reference."
  - "pkg/telemetry.Vocabulary carries every PRD field that already has a dedicated product_events column (session_id, game_id, role, source, duration_ms) on that column, never duplicated into metadata — a deliberate departure from 01-RESEARCH.md Pattern 5's [ASSUMED] table, which duplicated those fields into metadata on nearly every row. Recorded with full row-by-row discrepancy detail in docs/analytics/product-event-vocabulary.md."
  - "invite_copied's share_method metadata key has no PRD provenance; it is included because REQ-A4 requires recording which share mechanism succeeded, and its provenance is recorded as REQ-A4, not the PRD table."
  - "invite_viewed's four attribution keys were added even though 01-RESEARCH.md Pattern 5 omitted them for this event — the PRD row does require 'campaign source' here, and Pattern 5's assumption missed it."

patterns-established:
  - "withScratchMigrationDB: creates a scratch database off DATABASE_URL, hands its URL to a test closure, terminates lingering backends before DROP DATABASE (with a short retry loop), and never touches the shared TestMain database — the pattern every future migration up/down/up proof in this codebase should use."
  - "exerciseAndGather: builds a private prometheus.NewRegistry(), constructs Collectors, exercises every collector once (including families not yet observed in production), and returns Gather() — the only way four separate TestMetrics_* functions can share one exercised registry."

requirements-completed: [REQ-ACT-001, REQ-ACT-002]

coverage:
  - id: D1
    description: "End-to-end tracer: a pasted '1 Sol Ring' through previewDeck returns a correct DeckPreview, writes one deck_import_succeeded product_events row, and moves the vedh_deck_import_total counter by exactly one"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/deck_import_test.go#TestTracer_PreviewDeckEmitsMeasuredEvent"
        status: pass
      - kind: integration
        ref: "server/product_events_test.go#TestProductEvents_WriteFailureIsNonFatal"
        status: pass
    human_judgment: false
  - id: D2
    description: "The PostgreSQL 14 COALESCE dedup index actually deduplicates (both-NULL keys, concurrent in-flight collision, client/server replay shapes), a repeatable-event control is never deduplicated, occurred_at ties order deterministically and collapse correctly in a min()-per-session funnel query, the allowlist refuses five case types with an exact bounded-counter delta and writes no row, and the product_events migration pair applies/reverses/re-applies cleanly in scratch databases for both migration directories"
    requirement: "REQ-ACT-001"
    verification:
      - kind: integration
        ref: "server/product_events_test.go#TestProductEvents_AuthoritativeDedup"
        status: pass
      - kind: integration
        ref: "server/product_events_test.go#TestProductEvents_ConcurrentInsert"
        status: pass
      - kind: integration
        ref: "server/product_events_test.go#TestProductEvents_OccurredAtTieOrdering"
        status: pass
      - kind: integration
        ref: "server/product_events_test.go#TestProductEvents_Allowlist"
        status: pass
      - kind: integration
        ref: "server/product_events_test.go#TestMigrations_ProductEvents"
        status: pass
    human_judgment: false
  - id: D3
    description: "The 15-event vocabulary is closed and matches D-20 literally, no allowlisted metadata key names a forbidden concept, exactly the six PRD-Server events are marked Authoritative, every ValidateEvent rejection case returns its own bounded Rejection value, every vedh_-prefixed Prometheus label is in AllowedLabelNames, all four criterion-4 families exist with a counter and histogram, and source/role/reason label values are bounded to their declared enums"
    requirement: "REQ-ACT-001"
    verification:
      - kind: unit
        ref: "pkg/telemetry/vocabulary_test.go#TestVocabulary_IsClosedAtFifteen"
        status: pass
      - kind: unit
        ref: "pkg/telemetry/vocabulary_test.go#TestVocabulary_NoForbiddenKeys"
        status: pass
      - kind: unit
        ref: "pkg/telemetry/vocabulary_test.go#TestVocabulary_AuthoritativeSetMatchesSpec"
        status: pass
      - kind: unit
        ref: "pkg/telemetry/vocabulary_test.go#TestValidateEvent_Rejections"
        status: pass
      - kind: unit
        ref: "pkg/telemetry/metrics_test.go#TestMetrics_LabelAllowlist"
        status: pass
      - kind: unit
        ref: "pkg/telemetry/metrics_test.go#TestMetrics_AllCriterionFourFamiliesExist"
        status: pass
      - kind: unit
        ref: "pkg/telemetry/metrics_test.go#TestMetrics_SourceLabelValuesAreEnumMembers"
        status: pass
      - kind: unit
        ref: "pkg/telemetry/metrics_test.go#TestMetrics_ReasonLabelValuesAreBounded"
        status: pass
    human_judgment: false

duration: ~6h10m wall-clock across two sessions (Task 1 committed 15:42 local; this continuation resumed at ~21:35 local and completed Tasks 2–3 by 21:48 local — active execution time for Tasks 2–3 was under 15 minutes)
completed: 2026-08-04
status: complete
---

# Phase 1 Plan 1: Measured Deck Import Foundation — Tracer, Product Events, and Vocabulary Summary

**A pasted decklist becomes a measured `DeckPreview` end to end — through a new `pkg/deckimport` grammar, a PG14-dedup'd `product_events` table proven to actually deduplicate under concurrency, and a closed 15-event Prometheus/PostgreSQL telemetry vocabulary proven private and low-cardinality by tests that run in CI without a database.**

## Performance

- **Duration:** Task 1 (previous session, checkpoint-gated on user tracer verification) + this continuation for Tasks 2–3 (~13 min active execution)
- **Started:** 2026-08-04 (Task 1); continuation resumed 2026-08-04T21:35 local
- **Completed:** 2026-08-04T21:47:59-06:00
- **Tasks:** 3/3
- **Files modified:** 24 (17 created, 7 modified — see `key-files` above; full list in the plan's `files_modified` frontmatter)

## Accomplishments

- The complete Phase 1 GraphQL contract exists and is generated: `previewDeck`/`trackProductEvent` mutations, `DeckPreview` and its nested types, and the two new input types — verified absent from `type Query` and present in `type Mutation`.
- `product_events` ships as a real, proven table: the PG14 `COALESCE` dedup index is proven to deduplicate (not merely present in SQL text) on the both-NULL-keys case, under two-goroutine concurrency with `-race`, and across three client/server replay shapes; a repeatable-event control (`deck_import_succeeded`) is proven to *not* dedupe; `occurred_at` ties are proven to order deterministically and collapse correctly in a funnel query; and the migration pair is proven to apply/reverse/re-apply cleanly in scratch databases for both `persistence/migrations/` and `persistence/migrations_test/`.
- `pkg/telemetry`'s 15-event vocabulary, per-event key allowlists, and all 12 Prometheus collectors (covering all four ROADMAP criterion-4 families) are proven closed, private (no forbidden-substring key), and low-cardinality (no label outside `AllowedLabelNames`, all `source`/`role`/`reason` values bounded to their enums) — entirely by DB-free tests CI actually runs.
- The vocabulary is documented with per-key provenance in `docs/analytics/product-event-vocabulary.md`, including every discrepancy found against `01-RESEARCH.md` Pattern 5's `[ASSUMED]` table, and `docs/analytics/product-event-funnel-example.sql` ships the ACT-001 example host-activation query.
- Two reusable test helpers land for later plans to reuse rather than re-derive: `withScratchMigrationDB` (plan 01-04) and `exerciseAndGather` (plan 01-06).

## Task Commits

Each task was committed atomically:

1. **Task 1: End-to-end "a pasted card becomes a measured preview" — one path only** - `0144ecc` (feat) — completed in a prior session; user verified the tracer at the plan's checkpoint before this continuation began.
2. **Task 2: Prove the dedup index deduplicates, the write path is ordered, and the migration reverses** - `9374ed0` (test)
3. **Task 3: Prove the event vocabulary is closed, private, and low-cardinality** - `286ffed` (test)

**Plan metadata:** pending (this commit, immediately following)

## Files Created/Modified

- `persistence/migrations/20260804120000_product_events.{up,down}.sql` + `persistence/migrations_test/` mirror — the `product_events` table, two funnel indexes, and the PG14 `COALESCE` dedup index
- `server/schema.graphql`, `server/generated.go`, `server/models_gen.go`, `server/schema.resolvers.go`, `app/src/types/generated.ts` — the complete Phase 1 GraphQL contract, generated via `make generate`
- `pkg/deckimport/{result,scanner,sourcetype}.go` — parsed-deck value types, the minimal scanner, and the complete D-24 `SourceType` enum
- `pkg/telemetry/{vocabulary,metrics}.go` — the closed 15-event vocabulary and all 12 Prometheus collectors
- `pkg/telemetry/{vocabulary_test,metrics_test}.go` — the full structural test suite (Task 3): closed-at-15, no-forbidden-keys, authoritative-set, 12-case `ValidateEvent` table, and the shared `exerciseAndGather` helper backing four label/cardinality tests
- `server/{product_events,metrics,deck_import}.go` — the never-fail writer, the single collector registration site, and the `previewDeck` resolver
- `server/deck_import_test.go` — the tracer end-to-end test
- `server/product_events_test.go` — six test functions (Task 1's non-fatal-write smoke test plus Task 2's five): dedup, concurrency, ordering, allowlist behavior, and migration reversal, including the new `withScratchMigrationDB` helper
- `docs/analytics/product-event-vocabulary.md`, `docs/analytics/product-event-funnel-example.sql` — the documented vocabulary and example funnel query
- `.planning/phases/01-measured-deck-import-foundation/01-VALIDATION.md` — reconciled: every row this plan's three tasks made real flipped to ✅ green

## Decisions Made

- `previewDeck` lives on `Mutation`, never `Query`, per the locked api-contract; `DeckPreview.BlockingErrors` is a recorded additive extension beyond that contract.
- The dedup index's `COALESCE` wrappers are load-bearing on PostgreSQL 14; the both-NULL-keys subtest is the regression guard for a future "simplification" that would remove them.
- "Authoritative" (the six server-owned event names) and "deduplicated" (the four names in the dedup predicate) are deliberately kept distinct, with both senses documented wherever the ambiguous shared identifiers (`product_events_authoritative_once`, `TestProductEvents_AuthoritativeDedup`) appear.
- `pkg/telemetry.Vocabulary` carries every PRD field with a dedicated `product_events` column on that column, never duplicated into metadata — correcting `01-RESEARCH.md` Pattern 5's `[ASSUMED]` table, which duplicated those fields into metadata on nearly every row. Full discrepancy detail is in `docs/analytics/product-event-vocabulary.md`.
- `invite_copied`'s `share_method` key is provenanced to REQ-A4, not the PRD (which lists no metadata key for this event); `invite_viewed`'s four attribution keys were added per the PRD despite Pattern 5 omitting them.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `withScratchMigrationDB`'s cleanup raced a lingering backend and could leave orphaned scratch databases**
- **Found during:** Task 2 (`TestMigrations_ProductEvents`)
- **Issue:** The first implementation opened the scratch database's `*sql.DB` for each up/down/up step, closed it, then immediately ran `DROP DATABASE`. In practice the drop intermittently failed with `pq: database "..." is being accessed by other users` — a lingering Postgres backend from a step just closed on the Go side. The helper only logged this failure (per the plan's "tolerate an already-dropped database" instruction), which meant a failed drop silently left an orphaned database behind rather than actually cleaning up. Manual verification confirmed three `edhgo_mig_*` databases had accumulated from earlier runs.
- **Fix:** Added a `pg_terminate_backend` sweep against `pg_stat_activity` for the scratch database's name before attempting `DROP DATABASE`, plus a short retry loop (3 attempts, 100ms apart) around the drop itself.
- **Files modified:** `server/product_events_test.go`
- **Verification:** Re-ran `TestMigrations_ProductEvents` three times with no cleanup warnings; manually confirmed via `psql`/`pg_database` that no `edhgo_mig_*` databases remain after a run; manually dropped the three databases that had accumulated from the pre-fix runs.
- **Committed in:** `9374ed0` (Task 2 commit)

---

**2. [Rule 1 - Bug] Did not mark REQ-ACT-001/REQ-ACT-002 complete in REQUIREMENTS.md, despite the standard state-update step calling for it**
- **Found during:** state-update step, after Task 3
- **Issue:** This plan's frontmatter lists `requirements: [REQ-ACT-001, REQ-ACT-002]`, and the standard workflow step runs `requirements mark-complete` against that list. Running it flipped both requirements' checkboxes and traceability-table rows to "Complete" in `.planning/REQUIREMENTS.md`. But both are multi-plan requirements in this same phase: `REQ-ACT-002` ("Canonical deck parser and preview API...quantity/name, `1x`, spaced CSV, quoted CSV, headers, blank lines, sideboard/maybeboard") is also claimed by plans 01-02, 01-04, and 01-05 (only the minimal `<digits><space><name>` grammar exists after this plan); `REQ-ACT-001` ("...frontend event service with a persisted random session ID...") is also claimed by plans 01-03 and 01-06 (no frontend session-ID service exists yet). Marking either "Complete" after only plan 01-01 would misrepresent the project's actual requirements-traceability state to every later reader, including future planning agents that scan `REQUIREMENTS.md` for what remains.
- **Fix:** Reverted `.planning/REQUIREMENTS.md` with `git checkout -- .planning/REQUIREMENTS.md` and did not call `requirements mark-complete`. `requirements-completed` in this SUMMARY's frontmatter still lists both IDs per the template's instruction ("copy ALL requirement IDs from this plan's requirements frontmatter field") — that field documents what this plan *contributed to*, not a project-wide completion claim. The actual `REQUIREMENTS.md` checkboxes for `REQ-ACT-001`/`REQ-ACT-002` should be flipped by whichever of plans 01-02 through 01-06 is the last to land its piece of each requirement.
- **Files modified:** none (revert only; no file left in a changed state)
- **Verification:** `git diff .planning/REQUIREMENTS.md` is empty after the revert.
- **Committed in:** not applicable — reverted before any commit

---

**Total deviations:** 2 auto-fixed (1 bug, 1 process correction)
**Impact on plan:** Both fixes are correctness requirements. The scratch-database fix prevents orphaned databases from accumulating on the shared PostgreSQL instance; the REQUIREMENTS.md revert prevents two multi-plan requirements from being reported as done four-to-five plans before they actually are. No scope creep; the plan's stated cleanup tolerance ("tolerating an already-dropped database so a failed run does not leave the cleanup itself failing") was preserved as the outer safety net, with the terminate-and-retry logic added so that tolerance is rarely needed in practice.

## Issues Encountered

None beyond the deviation above.

## Known Issues (pre-existing, out of scope for this plan)

`/prometheus` over real HTTP returns 401 even with a correct `METRICS_TOKEN` bearer, because `server/graphql.go`'s `withAuthContext` wraps the whole mux and parses any `Authorization` header as a JWT before `withMetricsAuth` runs. This is pre-existing (not introduced by this plan), out of scope for plan 01-01, and the user explicitly chose not to fix it here. All of this plan's Prometheus proofs use a private `prometheus.NewRegistry()` and `promhttp`-free assertions, so this pre-existing issue does not affect any test in this plan.

## User Setup Required

None — no external service configuration required. PostgreSQL 14 and the MTGJSON snapshot required by `server/main_test.go`'s `TestMain` were already available in the execution environment.

## Next Phase Readiness

- Plan 01-02 can build the full parser grammar directly on top of `pkg/deckimport`'s value types and `SourceType` enum without touching the GraphQL contract, the migration, or the vocabulary — all three are closed for the rest of the phase.
- Plan 01-04 task 1 has `withScratchMigrationDB` ready to reuse by name for `TestMigrations_CardNameSearch`.
- Plan 01-06 has `exerciseAndGather` ready to extend for the rate-limit and deck-provider collector proofs.
- Phases 2 and 3 have three fully declared-but-unobserved collector families (guest session, game create/join, board activation) waiting only for an `Observe*` call at their respective emit sites — the label sets, bucket boundaries, and cardinality proofs are already closed.
- No blockers. `01-VALIDATION.md` carries no remaining `TBD` in the Task ID column for this plan's rows, and every row this plan's tasks made real is green.

---
*Phase: 01-measured-deck-import-foundation*
*Completed: 2026-08-04*

## Self-Check: PASSED

All 19 files listed under `key-files` (created/modified) exist on disk, and all three task commit
hashes (`0144ecc`, `9374ed0`, `286ffed`) are present in `git log --oneline --all`.
