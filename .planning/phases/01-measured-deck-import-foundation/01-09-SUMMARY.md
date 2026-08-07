---
phase: 01-measured-deck-import-foundation
plan: 09
subsystem: api
tags: [go, deckimport, archidekt, deck-provider, testing]

# Dependency graph
requires:
  - phase: 01-measured-deck-import-foundation
    provides: "01-08's archidektAdapter (normalizeToDeckText, formatArchidektDeckLine) and the committed real-response fixture server/testdata/deck_providers/archidekt_deck_2026-08-05.json"
  - phase: 01-measured-deck-import-foundation
    provides: "01-02's canonical pkg/deckimport grammar (Parse/ParseWithSource, AssertAccounting, D-07 printing metadata, D-11 dropped-section accounting) that the fixture-contract tests parse the normalized text through"
provides:
  - "Fixture-contract tests (TestArchidekt_*) pinning archidektAdapter's field mapping to the real, committed live capture -- a future upstream shape change now fails a test instead of silently mis-parsing a player's deck"
  - "The 126-vs-100/26 contract-drift canary, asserted arithmetically rather than in prose"
  - "End-to-end proof (TestDeckImport_ArchidektURLPath) that the provider URL path reuses pkg/deckimport's single canonical parse, not a second one (ACT-002)"
  - "A real bug fix: archidektCollectorNumberRoundTrips, closing a silent card-drop defect in formatArchidektDeckLine that these same tests discovered"
affects: ["01-10 (traceability closure for REQ-ACT-003)"]

actuals:
  tokens: 4725
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Fixture-contract tests load the real committed capture from disk (os.ReadFile), never an inline literal, and run it through the real adapter and the real canonical parser -- never a hand-built shortcut -- so a future upstream response-shape change breaks a test instead of silently mis-parsing a deck"
    - "withDeckProviderFetchStub (sibling of 01-06/01-08's withDeckProviderFetchSpy) serves a canned real fixture through the stubbed fetch seam, proving the full previewDeck -> fetch -> normalize -> Parse -> buildDeckPreview chain end to end with zero real network dials"

key-files:
  created: []
  modified:
    - server/deck_providers_test.go
    - server/deck_import_test.go
    - server/deck_providers.go

key-decisions:
  - "Rule 1 bug fix, found via this plan's own fixture test: formatArchidektDeckLine unconditionally emitted printing metadata as \"(setcode) collectornumber\", but pkg/deckimport's bare-collector-number grammar (isAlnumToken) only recognizes ASCII alphanumeric tokens. Archidekt's real \"The List\" reprints carry hyphenated collector numbers (e.g. \"MH1-216\"), which folded into the parsed card NAME itself and silently failed to resolve -- 7 real cards vanished from a 100-card deck (CardCount = 93, not 100) before the fix. Fixed by gating the printing-metadata suffix on a new archidektCollectorNumberRoundTrips predicate: a non-round-trippable collector number is omitted so the card still imports by name (D-09's name-only fallback), rather than emitted and mis-parsed. This was exactly the class of silent mis-parse this plan's objective names as the failure mode that matters."
  - "TestDeckImport_ArchidektURLPath's non-allowlisted-host subtest deliberately runs the REAL, unstubbed deckProviderFetch (fetchDeckProviderURL) rather than a spy: fetchDeckProviderURL's own host-allowlist check runs before any request is built or any DNS lookup is attempted, so this is a genuine zero-dial proof at the previewDeck level, not merely a restatement of the already-covered unit-level proof (TestSafeClient_InitialRequestHostEnforced)"
  - "Chose 'Thirst for Knowledge' (collector \"103\", purely digits) as the printing-metadata sample entry rather than any of the 7 hyphenated-collector-number rows, since the plan only requires proving the property holds for at least one row, not that every row round-trips"

patterns-established:
  - "A provider adapter's printing-metadata emission must be gated on whether the value actually round-trips through pkg/deckimport's grammar, not merely on non-emptiness -- non-alphanumeric collector-number formats (seen in the wild from 'The List' reprints) exist and must degrade to name-only import rather than corrupt the card name"

requirements-completed: [REQ-ACT-003, REQ-ACT-002]

coverage:
  - id: D1
    description: "126-vs-100/26 contract-drift canary: the real committed fixture's 126 total quantity across 113 rows resolves to exactly 100 in-deck quantity once the 26-quantity Maybeboard (includedInDeck: false) category is excluded, with the exclusion counted via one summary warning naming 26, never silent"
    requirement: "REQ-ACT-003"
    verification:
      - kind: unit
        ref: "server/deck_providers_test.go#TestArchidekt_QuantityAccounting"
        status: pass
      - kind: unit
        ref: "server/deck_providers_test.go#TestArchidekt_AssertAccountingHolds"
        status: pass
    human_judgment: false
  - id: D2
    description: "Commander preselection survives the provider path with its comma intact: exactly one SectionCommander entry, named \"Cid, Timeless Artificer\" -- proven against a fixture where the same name also appears as an unrelated main-deck row, so the assertion genuinely filters on Section rather than counting name occurrences"
    requirement: "REQ-ACT-003"
    verification:
      - kind: unit
        ref: "server/deck_providers_test.go#TestArchidekt_CommanderPreselection"
        status: pass
    human_judgment: false
  - id: D3
    description: "D-07 printing metadata (set code, collector number) is carried through from edition.editioncode/collectorNumber for at least one real entry, with set-code case preserved exactly as the response wrote it; adapter.source() and the parsed deck's Source both attribute to deckimport.SourceArchidekt"
    requirement: "REQ-ACT-003"
    verification:
      - kind: unit
        ref: "server/deck_providers_test.go#TestArchidekt_PrintingMetadata"
        status: pass
      - kind: unit
        ref: "server/deck_providers_test.go#TestArchidekt_Source"
        status: pass
    human_judgment: false
  - id: D4
    description: "Malformed-input fail-closed: empty body, non-JSON bytes, valid JSON with no cards[], and a card row missing card.oracleCard.name each return an error (never a partial deck), and none of the four returned errors carries any byte of the input"
    requirement: "REQ-ACT-003"
    verification:
      - kind: unit
        ref: "server/deck_providers_test.go#TestArchidekt_MalformedInputFailsClosed (4 subtests)"
        status: pass
    human_judgment: false
  - id: D5
    description: "The provider URL path (previewDeckURL) reuses pkg/deckimport's single canonical parse, not a second one: enabled and allowlisted, the real fixture through the stubbed fetch seam yields CardCount == 100 and CanContinue true, and CardCount equals len(library) for the identical parsed value, the same invariant TestDeckImport_SingleParse proves for pasted text"
    requirement: "REQ-ACT-002"
    verification:
      - kind: unit
        ref: "server/deck_import_test.go#TestDeckImport_ArchidektURLPath/EnabledProducesSameSingleParseAsPaste"
        status: pass
    human_judgment: false
  - id: D6
    description: "The kill switch and the allowlist remain independent gates on the real Archidekt path: disabled (default) never dials and returns the ordinary paste-fallback error; a host absent from the allowlist is refused before any dial or DNS lookup even with the flag on"
    requirement: "REQ-ACT-002"
    verification:
      - kind: unit
        ref: "server/deck_import_test.go#TestDeckImport_ArchidektURLPath/DisabledNeverDials"
        status: pass
      - kind: unit
        ref: "server/deck_import_test.go#TestDeckImport_ArchidektURLPath/NonAllowlistedHostNeverDialsEvenWithFlagOn"
        status: pass
    human_judgment: false
  - id: D7
    description: "Rule 1 bug fix: archidektCollectorNumberRoundTrips closes a silent card-drop defect where a hyphenated Archidekt collector number (e.g. \"MH1-216\" from 'The List' reprints) folded into the parsed card name and failed to resolve, dropping 7 real cards from the 100-card fixture deck (measured CardCount = 93 before the fix)"
    requirement: "REQ-ACT-003"
    verification:
      - kind: unit
        ref: "server/deck_import_test.go#TestDeckImport_ArchidektURLPath/EnabledProducesSameSingleParseAsPaste (CardCount == 100, regression-proof for the fix)"
        status: pass
      - kind: other
        ref: "go build ./... && go vet ./server; go test ./server -count=1 -race; go test ./pkg/... -race -count=1 -- all clean, zero failures"
        status: pass
    human_judgment: false

duration: 40min
completed: 2026-08-07
status: complete
---

# Phase 01 Plan 09: Pin the Archidekt contract with fixture tests Summary

**Fixture-contract tests against the real committed Archidekt capture (TestArchidekt_*, TestDeckImport_ArchidektURLPath) pin 01-08's normalizer to the observed contract and, in doing so, caught and fixed a real silent-card-drop bug in its printing-metadata emission.**

## Performance

- **Duration:** 40 min
- **Started:** 2026-08-07T23:05:00Z (approx.)
- **Completed:** 2026-08-07T23:45:19Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- Added six `TestArchidekt_*` tests in `server/deck_providers_test.go`, all loading the real, committed 568KB fixture from disk (never inlined) and running it through the real `archidektAdapter` and `deckimport.ParseWithSource`: the 126-vs-100/26 contract-drift canary, the accounting-invariant re-derivation, comma-intact commander preselection, D-07 printing metadata, source attribution, and four-case malformed-input fail-closed coverage
- Added `TestDeckImport_ArchidektURLPath` in `server/deck_import_test.go` with three subtests proving the end-to-end provider URL path reuses `pkg/deckimport`'s single canonical parse (ACT-002), the kill switch and allowlist remain independent gates, and a non-allowlisted host is refused via the *real*, unstubbed `fetchDeckProviderURL` before any dial
- **Discovered and fixed a real defect** while writing Task 2's end-to-end test: `formatArchidektDeckLine` unconditionally emitted `(setcode) collectornumber`, but `pkg/deckimport`'s bare-collector-number grammar only recognizes ASCII alphanumeric tokens. Archidekt's real "The List" reprints (7 of the fixture's 87 in-deck rows) carry hyphenated collector numbers (e.g. `MH1-216`), which folded into the parsed card *name* and silently failed to resolve — dropping 7 real cards from the 100-card deck (measured `CardCount = 93` before the fix). Added `archidektCollectorNumberRoundTrips` to gate the printing-metadata suffix, so a non-round-trippable collector number is omitted rather than mis-parsed, and the card still imports by name.
- All of `go test ./server -count=1`, `go test ./server -count=1 -race`, and `go test ./pkg/... -race -count=1` pass with zero failures; `go build ./...` and `go vet ./server` are clean; no stray binary produced.

## Task Commits

Each task was committed atomically:

1. **Task 1: Contract tests over the captured response** - `16c4e0c` (test)
2. **Task 2: Prove the provider path is not a second parser** - `a728d49` (test, includes the Rule 1 bug fix)

## Files Created/Modified
- `server/deck_providers_test.go` - Six `TestArchidekt_*` fixture-contract tests pinning `archidektAdapter`'s field mapping to the real committed capture
- `server/deck_import_test.go` - `TestDeckImport_ArchidektURLPath` (three subtests) and `withDeckProviderFetchStub`, proving the provider URL path reuses the canonical single parse
- `server/deck_providers.go` - Added `archidektCollectorNumberRoundTrips`; gated `formatArchidektDeckLine`'s printing-metadata suffix on it (Rule 1 fix for the hyphenated-collector-number card-drop defect)

## Decisions Made
- Gated printing-metadata emission on round-trippability rather than fixing `pkg/deckimport`'s shared grammar to accept hyphens: the fix belongs in this plan's stated scope (01-08's normalizer), and widening the shared bare-collector-number grammar used by every provider and the pasted-text path is a broader change this plan does not own
- Used the real, unstubbed `fetchDeckProviderURL` (not a spy) for the non-allowlisted-host subtest, since that function's own host check runs before any dial or DNS lookup is attempted — a genuine end-to-end zero-dial proof rather than a restatement of the existing unit-level proof
- Chose "Thirst for Knowledge" (a purely-numeric collector number) as the printing-metadata sample row, since the plan requires proving the property for at least one row, not for every row in the fixture

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed a silent card-drop defect in `formatArchidektDeckLine`**
- **Found during:** Task 2 (`TestDeckImport_ArchidektURLPath/EnabledProducesSameSingleParseAsPaste`)
- **Issue:** The subtest's first run reported `CardCount = 93, want 100`. Investigation traced this to 7 fixture rows whose Archidekt `collectorNumber` contains a hyphen (Magic's "The List" reprint naming convention, e.g. `MH1-216`). `formatArchidektDeckLine` emitted these unconditionally in the `"%s (%s) %s"` printing-metadata form, but `pkg/deckimport`'s `extractTrailingAnnotations`/`isAlnumToken` only recognizes a bare trailing collector number when it consists solely of ASCII letters and digits. For a hyphenated value, that recognition step silently fails, and the unparsed `"(plst) MH1-216"` suffix folds into the parsed card *name* — no card by that inflated name exists, so the entry never resolves and vanishes from both the preview and the library, with no warning of any kind.
- **Fix:** Added `archidektCollectorNumberRoundTrips`, replicating `pkg/deckimport`'s exact ASCII-alnum predicate locally in `server/deck_providers.go`, and gated `formatArchidektDeckLine`'s metadata suffix on it. The 7 affected rows now emit as plain `quantity name` lines, carrying no D-07 printing metadata for that entry but resolving correctly by name via D-09's name-only fallback — a strictly smaller loss than silently dropping the card.
- **Files modified:** `server/deck_providers.go`
- **Verification:** `TestDeckImport_ArchidektURLPath/EnabledProducesSameSingleParseAsPaste` now reports `CardCount = 100`; full `go test ./server -count=1`, `go test ./server -count=1 -race`, `go test ./pkg/... -race -count=1` all pass; `go build ./...` and `go vet ./server` clean
- **Committed in:** `a728d49` (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** This is exactly the class of defect the plan's objective names as the failure mode that matters — "not a crash but a *silent* mis-parse that hands a player the wrong deck." The fix is scoped entirely to `archidektAdapter`'s own emission logic; no architectural change, no scope creep beyond correctness.

## Issues Encountered
None beyond the deviation above.

## User Setup Required

None. No external service configuration required.

## Next Phase Readiness

REQ-ACT-003's substance is now genuinely proven: `archidektAdapter`'s field mapping is pinned to the real observed contract with fixture tests that would fail on an upstream shape change, the 100-card exclusion arithmetic is asserted rather than trusted, and the provider URL path is proven to reuse the single canonical parse end to end. Both `REQ-ACT-003` and `REQ-ACT-002` are marked complete in this plan's frontmatter and in REQUIREMENTS.md.

No blockers for plan 01-10 (traceability closure): the Archidekt provider is implemented, contract-tested, and its one real defect found during that testing is already fixed and covered by a regression-proof assertion (`CardCount == 100`).

---
*Phase: 01-measured-deck-import-foundation*
*Completed: 2026-08-07*

## Self-Check: PASSED

All files listed under "Files Created/Modified" verified present on disk; both commit hashes (`16c4e0c`, `a728d49`) verified present in `git log --oneline --all`.
