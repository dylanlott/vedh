---
phase: 01-measured-deck-import-foundation
plan: 04
subsystem: database
tags: [postgresql, pg_trgm, gist, golang-migrate, gqlgen, deck-import]

# Dependency graph
requires:
  - phase: 01-measured-deck-import-foundation (plan 01)
    provides: "server/deck_import.go's PreviewDeck resolver and buildDeckPreview assembly, the D-07 SetCode/CollectorNumber/ScryfallID/Category fields already declared on the generated Card and DeckPreviewEntry GraphQL types, and withScratchMigrationDB (server/product_events_test.go) for this plan's own up/down/up migration proof"
  - phase: 01-measured-deck-import-foundation (plan 02)
    provides: "pkg/deckimport.ParsedEntry's SetCode/CollectorNumber/Category fields (D-07), populated by the full scanner grammar this plan's resolution reads"
provides:
  - "persistence/migrations/20260804120100_card_name_search.{up,down}.sql (mirrored byte-identically into persistence/migrations_test/): pg_trgm extension, the card_names distinct-name projection with a GiST trigram index, and three cards expression indexes (lower-cased name, lower-cased facename, and the D-09 name/setcode/number composite)"
  - "persistence/import_all_printings_json.go's refreshCardNames: replays the card_names projection insert after every snapshot import, tolerating an absent relation and logging (never failing) on any other error"
  - "server/cards.go's Cards() batch lookup: lower-cased, index-usable WHERE clause; SELECT widened to carry SetCode/CollectorNumber/ScryfallID onto the returned Card"
  - "server/deck_import.go's D-09 two-stage printing resolution (resolveDeckEntries, exactPrintingQuery, nameOnlyFallbackQuery) and D-10 missing-printing warning; DeckPreviewEntry.SetCode/CollectorNumber/Category now populated from the parsed entry regardless of resolution outcome"
  - "server/deck_import.go's lowConfidenceCutoff constant and isLowConfidence(score float64) — the Go-only D-03 presentation cutoff, applied to a score never expressed in SQL"
affects: [01-05 (suggestion ranking consumes card_names/card_names_trgm_gist and lowConfidenceCutoff/isLowConfidence; createLibraryFromDecklist extraction reuses resolveDeckEntries's printing-aware resolution), 01-06 (rate limiting sits in front of previewDeck, which this plan's resolveDeckEntries now serves)]

# Actuals (#2632)
actuals:
  tokens: 10784
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "unnest(...) WITH ORDINALITY for a batched exact-match lookup: map a query result row back to its originating slice index via an ordinal column, with no second per-row query and no dependence on row-arrival order"
    - "Query text documents its own determinism: nameOnlyFallbackQuery's ORDER BY clause carries a comment naming the ordering key (lower(setcode), number, id) so 'which printing wins' is a decision visible in the SQL, not tribal knowledge"
    - "A Go constant cutoff applied to an already-returned score, explicitly never expressed in SQL (lowConfidenceCutoff/isLowConfidence), so a presentation threshold can be retuned without a migration"

key-files:
  created:
    - persistence/migrations/20260804120100_card_name_search.up.sql
    - persistence/migrations/20260804120100_card_name_search.down.sql
    - persistence/migrations_test/20260804120100_card_name_search.up.sql
    - persistence/migrations_test/20260804120100_card_name_search.down.sql
  modified:
    - persistence/import_all_printings_json.go
    - server/cards.go
    - server/cards_test.go
    - server/deck_import.go
    - server/deck_import_test.go

key-decisions:
  - "card_names is a plain table populated by INSERT ... SELECT DISTINCT with ON CONFLICT DO NOTHING, not a materialized view — per 01-RESEARCH.md Open Question 4 — because the refresh hook in persistence/import_all_printings_json.go needs an incremental, transactional insert on every snapshot import, not a periodic REFRESH MATERIALIZED VIEW."
  - "GiST (gist_trgm_ops), not GIN, on card_names.name_lower — decided by correctness (D-03's 'always show nearest matches below cutoff' cannot be served by GIN's boolean % operator), not benchmarked at this table's ~33k-row size."
  - "Collector number is compared as stored text, never lower-cased; only the name and set code columns are lower-cased in both the index and the query, matching D-09's literal wording ('set-code matching is case-insensitive')."
  - "An entry with a set code but no collector number is treated as a valid stage-one candidate: exactPrintingQuery's `(req.number = '' OR c.number = req.number)` clause lets it match any printing in that set, broken by the same number/id ordering documented on nameOnlyFallbackQuery, rather than being pushed straight to the name-only fallback."
  - "DeckPreviewEntry.SetCode/CollectorNumber/Category are now populated from the parsed entry for every entry, not only ones that hit a D-10 warning — a Rule 2 fix (see Deviations) for a gap left by the 01-01 tracer, where these already-declared GraphQL fields were never wired up."
  - "resolveDeckEntries deliberately duplicates server/cards.go's Cards() column list and scan shape (cardResultColumns/scanCardResultRow) rather than calling Cards() directly, continuing the QueryContext-vs-Query deviation server/deck_import.go's original lookupCardsByName already established: previewDeck's public request path should be able to propagate a caller deadline, which cards.go's Cards() (using Query, not QueryContext) cannot."

patterns-established:
  - "assertCardNameSearchObjectsPresent / assertTrgmExtensionInstalled (server/deck_import_test.go): reusable present/absent assertions over pg_tables/pg_indexes/pg_extension for a *sql.DB, mirroring plan 01-01's assertProductEventsObjectsPresent shape — the pattern any future migration-parity test in this package should follow rather than re-deriving pg_catalog queries."
  - "seedPrintingTestCards + previewCardTextLine (server/deck_import_test.go): a named test-data seam for inserting synthetic multi-printing rows under a name guaranteed not to collide with the real MTGJSON snapshot ('Test Printing Card'), reusable by plan 01-05's suggestion-ranking tests."

requirements-completed: [REQ-ACT-002]

coverage:
  - id: D1
    description: "The cards table carries indexes on lower-cased name, lower-cased face name, and the composite lower-cased name/set-code/collector-number key; a card_names projection holds one row per distinct searchable name with a trigram GiST index; the migration pair applies/reverses/re-applies cleanly in scratch databases for both migration directories without ever touching the shared TestMain database"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/deck_import_test.go#TestMigrations_CardNameSearch"
        status: pass
    human_judgment: false
  - id: D2
    description: "Importing the MTGJSON snapshot repopulates card_names without a migration, tolerating the projection table being absent by logging and continuing"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/main_test.go#TestMain (drives persistence.ImportAllPrintingsJSON, which calls refreshCardNames, ahead of every server test in this package)"
        status: pass
      - kind: integration
        ref: "server/deck_import_test.go#TestMigrations_CardNameSearch (asserts card_names is non-empty for a name present in the imported snapshot)"
        status: pass
    human_judgment: false
  - id: D3
    description: "A batch card lookup with a lower-cased needle matches a stored name of any letter case, and the lookup query lower-cases both the column and the parameter so the expression indexes are usable"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/cards_test.go#TestCards_LowerCasedNeedleMatchesMixedCaseName"
        status: pass
      - kind: integration
        ref: "server/cards_test.go#Test_graphQLServer_Cards (pre-existing suite, unchanged, still green)"
        status: pass
    human_judgment: false
  - id: D4
    description: "An entry naming a set code and collector number resolves to that exact printing when it exists, and set-code matching is case-insensitive so both spellings of the same set select the same row"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/deck_import_test.go#TestDeckImport_PrintingDisambiguation (subtests XYZ and xyz)"
        status: pass
    human_judgment: false
  - id: D5
    description: "An entry naming a set code absent from the snapshot resolves to some available printing of that name, carries a warning naming the requested printing, keeps the parsed set code on the entry, and is not reported as unresolved"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/deck_import_test.go#TestDeckImport_MissingPrintingWarns"
        status: pass
    human_judgment: false
  - id: D6
    description: "An entry naming no set code or collector number resolves by name alone, and which printing is returned is decided by a documented deterministic ordering rather than by row arrival order"
    requirement: "REQ-ACT-002"
    verification:
      - kind: integration
        ref: "server/deck_import_test.go#TestDeckImport_NameOnlyResolutionIsDeterministic"
        status: pass
    human_judgment: false
  - id: D7
    description: "A trigram score exactly equal to the low-confidence cutoff is treated as confident, and a score one step below it is marked low-confidence; the cutoff is applied in Go over the returned score and appears nowhere in the SQL"
    requirement: "REQ-ACT-002"
    verification:
      - kind: unit
        ref: "server/deck_import_test.go#TestDeckImport_LowConfidenceCutoffBoundary"
        status: pass
    human_judgment: false
  - id: D8
    description: "No SQL string in server/deck_import.go or server/cards.go is composed with string formatting, and go vet ./server plus the full go test ./server -run 'TestCards|TestDeckImport_|TestMigrations_' -race suite passes in one invocation without a snapshot reimport"
    requirement: "REQ-ACT-002"
    verification:
      - kind: automated
        ref: "grep -nE 'Sprintf\\(`?[[:space:]]*(SELECT|INSERT|UPDATE|DELETE)' server/deck_import.go server/cards.go (zero matches)"
        status: pass
      - kind: automated
        ref: "go vet ./server"
        status: pass
      - kind: integration
        ref: "go test ./server -run 'TestCards|TestDeckImport_|TestMigrations_' -race"
        status: pass
    human_judgment: false

duration: ~31min
completed: 2026-08-04
status: complete
---

# Phase 1 Plan 4: Name-Search Migration and Printing-Aware Card Resolution Summary

**The `cards` table gains three expression indexes and a trigram-indexed `card_names` projection, and `previewDeck`'s card lookup becomes lower-cased, index-usable, and printing-aware: an exact `(name, setcode, number)` match wins when the paste supplies one, a documented deterministic name-only ordering wins otherwise, and a printing missing from the snapshot warns instead of silently shrinking the deck.**

## Performance

- **Duration:** ~31 min
- **Started:** 2026-08-04T22:33:00-06:00 (approx.)
- **Completed:** 2026-08-04T23:04:00-06:00
- **Tasks:** 2/2
- **Files modified:** 9 (4 created, 5 modified)

## Accomplishments

- `persistence/migrations/20260804120100_card_name_search.{up,down}.sql` (byte-identical in `persistence/migrations_test/`) installs `pg_trgm`, creates the `card_names` distinct-name projection with a GiST trigram index (`card_names_trgm_gist`), and adds the three `cards` expression indexes the batch lookup and printing disambiguation both need (`cards_name_lower_idx`, `cards_facename_lower_idx`, `cards_name_set_num_idx`). The down migration is a true inverse that deliberately never drops the shared `pg_trgm` extension. `TestMigrations_CardNameSearch` proves the full up/down/up cycle against scratch databases for both migration directories, reusing plan 01-01's `withScratchMigrationDB` by name rather than declaring a second helper.
- `persistence/import_all_printings_json.go` now calls `refreshCardNames` after every snapshot import's final batch commit, replaying the same distinct-name projection insert the migration performs — tolerating an absent `card_names` relation as "nothing to refresh" and logging (never failing) on any other error. This is what makes the imported test database's `card_names` non-empty at all, since `server/main_test.go`'s `TestMain` runs migrations before the snapshot import.
- `server/cards.go`'s `Cards()` batch lookup lower-cases both its `WHERE` clause and its bound needles (usable by the new expression indexes), and widens its `SELECT` to carry `SetCode`/`CollectorNumber`/`ScryfallID` onto the returned `Card` — strictly wider than the previous case-sensitive comparison, proven by a new test asserting a fully lower-cased needle now matches a mixed-case stored name.
- `server/deck_import.go` replaces the tracer's name-only resolution with D-09's two-stage resolve: stage one is one batched, ordinality-preserving `unnest(...)` query over every entry supplying a set code or collector number, matching exactly on `(lower(name), lower(setcode), number)`, with upper- and lower-case set-code spellings proven to select the same row. Stage two is a deterministic name-only fallback — explicitly ordered by `lower(setcode)`, then `number`, then `id`, with that ordering documented directly on the query — for everything stage one misses, including a set code simply absent from the snapshot, which now resolves with a D-10 warning naming the requested printing rather than disappearing from the deck or being reported as unresolved.
- `DeckPreviewEntry.SetCode`/`CollectorNumber`/`Category` — GraphQL fields that existed since the 01-01 tracer but were never populated — now carry the parsed printing metadata through for every entry regardless of resolution outcome, closing that gap (see Deviations).
- `lowConfidenceCutoff` and `isLowConfidence(score float64) bool` land in `server/deck_import.go` as the Go-only D-03 presentation cutoff (0.3, mirroring `pg_trgm.similarity_threshold`'s own default): a score exactly at the cutoff is confident, one step below is low-confidence, and the cutoff is never expressed in SQL. Plan 01-05's suggestion ranking is the first intended consumer.

## Task Commits

Each task was committed atomically:

1. **Task 1: Name-search migration and a snapshot-import refresh for the projection** - `3faa9a1` (feat)
2. **Task 2: Index-usable batch lookup and printing-aware resolution** - `318fa7e` (feat)

**Plan metadata:** pending (this commit, immediately following)

## Files Created/Modified

- `persistence/migrations/20260804120100_card_name_search.{up,down}.sql` + `persistence/migrations_test/` mirror — `pg_trgm`, `card_names`, `card_names_trgm_gist`, and the three `cards` expression indexes
- `persistence/import_all_printings_json.go` — `refreshCardNames`, called after every snapshot import's final batch; a local `isMissingRelation` copy (this package cannot import `server`'s)
- `server/cards.go` — `Cards()`'s widened, lower-cased SELECT/WHERE
- `server/cards_test.go` — `TestCards_LowerCasedNeedleMatchesMixedCaseName`
- `server/deck_import.go` — `lowConfidenceCutoff`, `isLowConfidence`, `resolveDeckEntries`, `exactPrintingQuery`, `nameOnlyFallbackQuery`, `cardResultColumns`/`scanCardResultRow`, `stringPtrOrNil`; `buildDeckPreview` rewired to the two-stage resolver and to populate `DeckPreviewEntry`'s printing fields
- `server/deck_import_test.go` — `TestMigrations_CardNameSearch` (+ `assertCardNameSearchObjectsPresent`/`assertTrgmExtensionInstalled` helpers), `TestDeckImport_PrintingDisambiguation`, `TestDeckImport_MissingPrintingWarns`, `TestDeckImport_NameOnlyResolutionIsDeterministic`, `TestDeckImport_LowConfidenceCutoffBoundary` (+ `seedPrintingTestCards`/`previewCardTextLine` test-data helpers)

## Decisions Made

- `card_names` is a plain table refreshed by an incremental `INSERT ... ON CONFLICT DO NOTHING`, not a materialized view, because the snapshot-import refresh hook needs a transactional insert on every import rather than a periodic `REFRESH MATERIALIZED VIEW` (01-RESEARCH.md Open Question 4).
- GiST (`gist_trgm_ops`), not GIN, on `card_names.name_lower` — decided by D-03's correctness requirement (nearest matches must show even below the similarity cutoff, which GIN's boolean `%` operator cannot do), not by benchmarking at this table's size.
- Collector number is compared as stored text; only name and set code are lower-cased, matching D-09's literal "set-code matching is case-insensitive" wording (collector numbers were never claimed to need case-folding).
- An entry with a set code but no collector number is a valid stage-one candidate (`exactPrintingQuery`'s `req.number = ''` short-circuit), rather than being pushed straight to the name-only fallback — it still gets the benefit of an exact-set match, just not narrowed to one collector number.
- `resolveDeckEntries` deliberately duplicates `server/cards.go`'s column list and scan shape rather than calling `Cards()` directly, continuing the pre-existing `QueryContext`-vs-`Query` deviation this file's tracer already established: `previewDeck` is on the public request path and should be able to propagate a caller deadline.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] `DeckPreviewEntry.SetCode`/`CollectorNumber`/`Category` were declared in the GraphQL schema since plan 01-01 but never populated**
- **Found during:** Task 2, while reading `buildDeckPreview` before extending it for the two-stage resolver
- **Issue:** `server/models_gen.go`'s `DeckPreviewEntry` has had `SetCode`, `CollectorNumber`, and `Category` fields since plan 01-01's tracer (they mirror `pkg/deckimport.ParsedEntry`'s own D-07 fields), but the tracer's `buildDeckPreview` never set them when constructing each `DeckPreviewEntry`. This plan's own acceptance criterion — "an entry ... keeps the parsed set code on the entry" (D-10) — is unsatisfiable as a player-visible, testable fact without wiring these fields, since the alternative (verifying only the internal `deckimport.ParsedEntry`, which was never mutated) would prove nothing a client-facing test could observe.
- **Fix:** `buildDeckPreview` now sets `SetCode: stringPtrOrNil(pe.SetCode)`, `CollectorNumber: stringPtrOrNil(pe.CollectorNumber)`, `Category: stringPtrOrNil(pe.Category)` for every constructed `DeckPreviewEntry`, using a new `stringPtrOrNil` helper (the `*string`-from-`""`-or-value analog of the existing `nullStringPtr`).
- **Files modified:** `server/deck_import.go`
- **Verification:** `TestDeckImport_MissingPrintingWarns` asserts `entry.SetCode != nil && *entry.SetCode == "ZZZ"` directly on the GraphQL-shaped `DeckPreviewEntry`.
- **Committed in:** `318fa7e` (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 missing critical functionality)
**Impact on plan:** The fix is required for this plan's own D-10 acceptance criterion to be testable and true at the GraphQL boundary the player actually sees. No scope creep — it wires up fields the schema already declared for exactly this purpose.

## Issues Encountered

None beyond the deviation above.

## User Setup Required

None — no external service configuration required. PostgreSQL 14 and the MTGJSON snapshot required by `server/main_test.go`'s `TestMain` were already available in the execution environment.

## Next Phase Readiness

- Plan 01-05's suggestion ranking has everything it needs waiting: `card_names` and `card_names_trgm_gist` for the KNN `<->` nearest-name query, and `lowConfidenceCutoff`/`isLowConfidence` for classifying the returned scores — neither needs a migration or a new query pattern, only a call site.
- `resolveDeckEntries` is written as a self-contained, index-aware resolver `buildDeckPreview` calls; plan 01-05's `createLibraryFromDecklist` extraction (making `previewDeck` and final library creation share one canonical parse-and-resolve site) can call it directly rather than re-deriving the two-stage logic.
- `seedPrintingTestCards`/`previewCardTextLine` (test-only helpers in `server/deck_import_test.go`) are ready for plan 01-05's suggestion tests to reuse for multi-printing fixtures.
- No blockers. `go build ./...`, `go vet ./server`, `go test ./pkg/... -race`, and `go test ./server -run 'TestCards|TestDeckImport_|TestMigrations_' -race` are all green; a full `go test ./server -race` run shows exactly the three pre-existing, out-of-scope failures documented at this plan's start (`TestCreateGame`'s three subtests, `TestUpdateBoardState_AutoFinishRules`'s one subtest, `TestAdvancePhase`) and no new failures.

---
*Phase: 01-measured-deck-import-foundation*
*Completed: 2026-08-04*

## Self-Check: PASSED

All 9 files listed under `key-files` (created/modified) exist on disk, and both task commit
hashes (`3faa9a1`, `318fa7e`) are present in `git log --oneline --all`.
