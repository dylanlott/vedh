---
phase: 01-measured-deck-import-foundation
plan: 02
subsystem: api
tags: [decklist-parser, deckimport, mtg, golang]

# Dependency graph
requires:
  - phase: 01-measured-deck-import-foundation (plan 01)
    provides: "pkg/deckimport's ParsedDeck/ParsedEntry/Warning/BlockingError/Section value types, the D-23/D-24 SourceType enum, and the minimal <digits><space><name> tracer scanner this plan widens"
provides:
  - "The complete canonical decklist grammar in pkg/deckimport/scanner.go: all six required syntaxes (1, 1x, 1X, 1,, 1, , quoted, unquoted-comma), double-faced \"//\" names surviving unquoted, and D-07 printing metadata (set code, collector number, category via bracket/backtick/hash-tag), all right-to-left-extracted with set-code case preserved"
  - "ParseWithSource — an exported seam that forces ParsedDeck.Source to any SourceType so a test (or a later phase's URL-host detector) can prove detection never influences parsing"
  - "AssertAccounting — the in-code, exported form of the \"no nonblank row disappears\" invariant, appending a BlockingError on mismatch instead of returning a silently short deck"
  - "pkg/deckimport/sections.go: real D-11 Sideboard/Maybeboard row-dropping with a per-section summary warning, and D-12 Commander-header preselection via SectionCommander"
  - "pkg/deckimport/sourcetype.go's real paste-shape heuristics: uppercase (SET) NUMBER majority → moxfield, trailing bracket/backtick category or a header's parenthesised count → archidekt, no marker → plain_text, markers from more than one provider → unknown"
  - "Three whole-file golden fixtures (testdata/{moxfield,archidekt,plain_text}_export.txt) plus checked-in expected ParsedDeck JSON and a -update regeneration flag that refuses to run under CI=true"
affects: [01-04 (card-name search/suggestions consumes these ParsedEntry.Name values), 01-05 (createLibraryFromDecklist extraction reuses this grammar as the single canonical parse site), 01-06 (rate limiting/provider adapter — SourceType values this plan's detector produces become the second detector's peer)]

# Actuals (#2632)
actuals:
  tokens: 20649
  tasks: 2
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Explicit numbered ordered-rule line scanner (no regex, no CSV reader) — each of the 8 grammar rules is a separate, commented step in Parse, so rule order is reviewable rather than inferred"
    - "AssertAccounting as an exported, directly-testable invariant checker: a struct's own internal consistency rule lives in code, not only in a test, and is provable against a deliberately-inconsistent value constructed without going through the normal constructor path"
    - "ParseWithSource as a forced-value test seam for a value that's normally computed internally (DetectPasteSource) — proves structural independence between two computations that happen to run over the same input, without needing mocks or interfaces"
    - "sectionTracker as a per-call (not package-level) state machine threaded through a single scan loop, closing out a still-open section either on the next header or explicitly at EOF"

key-files:
  created:
    - pkg/deckimport/scanner_test.go
    - pkg/deckimport/sections.go
    - pkg/deckimport/sections_test.go
    - pkg/deckimport/sourcetype_test.go
    - pkg/deckimport/golden_test.go
    - pkg/deckimport/testdata/moxfield_export.txt
    - pkg/deckimport/testdata/moxfield_export.expected.json
    - pkg/deckimport/testdata/archidekt_export.txt
    - pkg/deckimport/testdata/archidekt_export.expected.json
    - pkg/deckimport/testdata/plain_text_export.txt
    - pkg/deckimport/testdata/plain_text_export.expected.json
  modified:
    - pkg/deckimport/scanner.go
    - pkg/deckimport/sourcetype.go

key-decisions:
  - "Task 1's scanner.go rewrite deliberately did NOT depend on sections.go: header words are recognized well enough to never become an Entry, but with no drop/warning/preselect semantics, so Task 1 and Task 2 remain independently committable and testable — Task 1's own test suite (TestScanner_*) never references sections.go, and re-ran green both before and after Task 2 wired the real tracker in."
  - "The dropped-section header row is accounted for by attributing D-11's one summary warning to that header's own SourceLine, rather than by adding it to DroppedSectionRows. This is what makes the accounting invariant (Entries+Warnings+BlockingErrors+DroppedSectionRows == NonBlankLines) balance exactly: the header row becomes the Warning, and each row beneath it becomes a DroppedSectionRow, together covering every nonblank line the section contains. Commander/Deck/Companion headers and comment lines, having no natural warning, are excluded from NonBlankLines entirely (treated like blank lines) rather than invented one."
  - "extractQuantity treats a leading '-' followed by digits as a failed quantity attempt (one Warning, zero entries), not as \"no leading digits\" (which would silently swallow the whole line, including the minus sign, into the name). This distinction isn't spelled out verbatim in the plan's rule-4 prose but is required to satisfy the '-1 Sol Ring → one warning' corpus row without contradicting 'absent leading digits, default the quantity to 1.'"
  - "DetectPasteSource's majority threshold for moxfield is a strict majority (moxfieldLines*2 > entryLines), matching the plan's 'majority of entry lines' wording literally rather than requiring 100%."

patterns-established:
  - "Right-to-left trailing-annotation extraction as a sequence of independent HasSuffix checks over a shrinking string, each conditionally consuming and re-trimming — reusable for any future 'strip trailing decorations from a free-text field' need in this package."
  - "splitLines: a single explicit pass treating \\r, \\n, and \\r\\n as identical terminators, replacing the ad-hoc TrimSuffix(line, \"\\r\") from the plan 01-01 tracer — any future line-oriented parser in this package should reuse it rather than re-deriving CRLF handling."

requirements-completed: [REQ-ACT-002]

coverage:
  - id: D1
    description: "All six required decklist syntaxes (1, 1x, 1X, 1,, 1, , quoted-comma, unquoted-comma) parse to identical (quantity, name) pairs, and the comma-truncation regression (1,Atraxa, Praetors' Voice) yields the full name instead of truncating to \"Atraxa\""
    requirement: "REQ-ACT-002"
    verification:
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_RequiredSyntaxes"
        status: pass
    human_judgment: false
  - id: D2
    description: "A double slash inside a line is a double-faced card-name separator preserved in the name (verified against the repo's own test/decklists/jarad.csv fixtures); a double slash at the start of a line is a comment producing no entry"
    requirement: "REQ-ACT-002"
    verification:
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_RequiredSyntaxes/unquoted_double-faced_name"
        status: pass
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_RequiredSyntaxes/double_slash_comment"
        status: pass
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_NoRowDisappears"
        status: pass
    human_judgment: false
  - id: D3
    description: "D-07 printing metadata (set code, collector number, category via bracket/backtick/hash-tag) is extracted into its own fields with the remaining name clean, and set-code letter case is preserved exactly as written"
    requirement: "REQ-ACT-002"
    verification:
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_RequiredSyntaxes/printing_metadata_set_and_collector"
        status: pass
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_RequiredSyntaxes/printing_metadata_set_collector_and_bracket_category"
        status: pass
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_RequiredSyntaxes/backtick_category"
        status: pass
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_RequiredSyntaxes/hash_tag_categories"
        status: pass
    human_judgment: false
  - id: D4
    description: "The 'no nonblank row disappears' accounting invariant (Entries+Warnings+BlockingErrors+DroppedSectionRows == NonBlankLines) is asserted mechanically in code (AssertAccounting), not only in a test, and is proven over the required-syntax corpus and both real test/decklists/*.csv fixtures"
    requirement: "REQ-ACT-002"
    verification:
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_NoRowDisappears"
        status: pass
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_AccountingInvariantCatchesInconsistency"
        status: pass
    human_judgment: false
  - id: D5
    description: "Parse is pure (two calls on the same text are deeply equal) and holds no mutable package-level state (safe under -race from concurrent goroutines); MaxDecklistBytes/MaxDecklistLines are each enforced exactly at their boundary with no panic"
    requirement: "REQ-ACT-002"
    verification:
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_IsPure"
        status: pass
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_ConcurrentParseIsSafe"
        status: pass
      - kind: unit
        ref: "pkg/deckimport/scanner_test.go#TestScanner_InputBounds"
        status: pass
    human_judgment: false
  - id: D6
    description: "A Sideboard/Maybeboard header drops its rows with one summary warning stating the count; a Commander header preselects candidates via SectionCommander; a card name beginning with a header word (\"1 Commander's Sphere\") stays an entry, never a header"
    requirement: "REQ-ACT-002"
    verification:
      - kind: unit
        ref: "pkg/deckimport/sections_test.go#TestSections_DropAndPreselect"
        status: pass
    human_judgment: false
  - id: D7
    description: "Paste-shape source detection (moxfield/archidekt/plain_text/unknown) never influences what a line parses to: forcing every SourceType via ParseWithSource across all three golden fixtures produces deeply equal ParsedDecks apart from the Source field"
    requirement: "REQ-ACT-002"
    verification:
      - kind: unit
        ref: "pkg/deckimport/sourcetype_test.go#TestSourceDetection_DoesNotAffectParse"
        status: pass
      - kind: unit
        ref: "pkg/deckimport/sourcetype_test.go#TestSourceDetection_ReturnsOnlyEnumMembers"
        status: pass
      - kind: unit
        ref: "pkg/deckimport/sourcetype_test.go#TestSourceDetection_ShapesDetectDistinctly"
        status: pass
    human_judgment: false
  - id: D8
    description: "Three whole-file golden fixtures (Moxfield-, Archidekt-, plain-text-shaped) round-trip against their checked-in expected ParsedDeck JSON, each additionally passing the accounting invariant; the -update flag refuses to run when CI=true"
    requirement: "REQ-ACT-002"
    verification:
      - kind: unit
        ref: "pkg/deckimport/golden_test.go#TestGolden_Fixtures"
        status: pass
    human_judgment: false

duration: ~37min
completed: 2026-08-04
status: complete
---

# Phase 1 Plan 2: Full Decklist Grammar, Sections, and Source Detection Summary

**pkg/deckimport/scanner.go now implements the complete eight-rule canonical decklist grammar — all six required syntaxes, the comma-truncation regression fixed, double-faced names, D-07 printing metadata, D-11 sideboard dropping, D-12 commander preselection, and provably-inert D-23 source detection — backed by a mechanical "no row disappears" invariant and three golden fixtures.**

## Performance

- **Duration:** ~37 min
- **Started:** 2026-08-04T21:52:29-06:00 (plan 01-01 completion)
- **Completed:** 2026-08-04T22:29:26-06:00
- **Tasks:** 2/2
- **Files modified:** 13 (11 created, 2 modified)

## Accomplishments

- All six required decklist syntaxes (`1 Sol Ring`, `1x Sol Ring`, `1,Sol Ring`, `1, Sol Ring`, `1,"Atraxa, Praetors' Voice"`, `1 Atraxa, Praetors' Voice`) parse to identical `(quantity, name)` pairs, and the comma-truncation regression (`1,Atraxa, Praetors' Voice`, which today silently truncates to `Atraxa` in `server/games.go`) now yields the full name.
- Double-faced card names written with `" // "` survive unquoted (verified against the repo's own `test/decklists/jarad.csv`), while a line-initial `//` is correctly read as a comment — the two never collide because the comment check is anchored strictly at index 0.
- D-07 printing metadata — set code, collector number, and category via bracketed, backtick-delimited, or hash-tag forms — is extracted right-to-left off the unquoted remainder, with set-code letter case preserved exactly as typed (`C21` and `c21` both round-trip unchanged).
- The "no nonblank row disappears" invariant is enforced in code, not only asserted by a test: `AssertAccounting` appends a `BlockingError` on any mismatch between `NonBlankLines` and the sum of `Entries`+`Warnings`+`BlockingErrors`+`DroppedSectionRows`, and is itself unit-tested against a deliberately inconsistent `ParsedDeck`.
- `pkg/deckimport/sections.go` gives Sideboard/Maybeboard headers real D-11 semantics (rows dropped, one summary warning naming the count) and Commander headers real D-12 semantics (`SectionCommander` preselection) — while proving `"1 Commander's Sphere"` stays an entry because the header rule requires the header word to be the *whole* line.
- `DetectPasteSource` now runs real paste-shape heuristics instead of the 01-01 tracer's single branch, and `ParseWithSource` is a new exported seam that lets a test force every `SourceType` and prove the rest of the parsed result never changes — the D-23 structural-isolation guarantee, checked against all three golden fixtures.
- Three whole-file golden fixtures ship in `pkg/deckimport/testdata/`, each ≥30 lines with a comma-bearing name, an unquoted double-faced name drawn from `test/decklists/jarad.csv`, a blank line, and a section header, plus checked-in expected `ParsedDeck` JSON and a `-update` flag that refuses to run under `CI=true` (demonstrated during development).

## Task Commits

Each task was committed atomically:

1. **Task 1: The full line grammar — six syntaxes, comma names, double faces, printing metadata** - `d734e25` (feat)
2. **Task 2: Sections, source detection, and the golden fixture corpus** - `69cb8f3` (test)
3. **Follow-up: number scanner.go's rule 7 explicitly** - `3396d0f` (docs) — a same-task acceptance-criterion fixup (all eight grammar rules now carry an explicit numbered comment), committed separately per the "always create a new commit" rule rather than amending Task 2.

**Plan metadata:** pending (this commit, immediately following)

## Files Created/Modified

- `pkg/deckimport/scanner.go` — rewritten `Parse`/`ParseWithSource` implementing the full eight-rule grammar; `AssertAccounting`, `splitLines`, `extractQuantity`, `resolveNameAndMetadata`, `extractTrailingAnnotations` are the new supporting functions
- `pkg/deckimport/scanner_test.go` — `TestScanner_RequiredSyntaxes` (24-row corpus), `TestScanner_NoRowDisappears`, `TestScanner_AccountingInvariantCatchesInconsistency`, `TestScanner_IsPure`, `TestScanner_ConcurrentParseIsSafe`, `TestScanner_InputBounds`
- `pkg/deckimport/sections.go` — `matchSectionHeader`, `sectionTracker` (D-11 drop bookkeeping, D-12 preselection)
- `pkg/deckimport/sections_test.go` — `TestSections_DropAndPreselect` (8 subtests covering drop count, preselection, the `Commander's Sphere` disambiguation, and dropped-row accounting)
- `pkg/deckimport/sourcetype.go` — real `DetectPasteSource` heuristics (`hasUppercaseSetAndCollector`, `hasBracketOrBacktickCategory`, `hasParenthesizedCount`)
- `pkg/deckimport/sourcetype_test.go` — `TestSourceDetection_ReturnsOnlyEnumMembers`, `TestSourceDetection_ShapesDetectDistinctly`, `TestSourceDetection_DoesNotAffectParse`
- `pkg/deckimport/golden_test.go` — `TestGolden_Fixtures` with a `-update` flag and a `CI=true` refusal guard
- `pkg/deckimport/testdata/{moxfield,archidekt,plain_text}_export.txt` + `.expected.json` — the three golden fixtures

## Decisions Made

- Task 1's `scanner.go` was deliberately written to not depend on `sections.go` (which Task 2 creates): header words are recognized well enough to never become an `Entry`, with real drop/warning/preselect semantics added only in Task 2. This kept the two tasks independently buildable and testable — Task 1's own suite never references `sections.go`, and stayed green both before and after Task 2 wired the real tracker in.
- The dropped-section header row is accounted for via D-11's one summary warning (attributed to the header's own `SourceLine`), not via `DroppedSectionRows`. This is the only assignment that makes the accounting invariant balance exactly: the header row becomes the `Warning`, the rows beneath it become `DroppedSectionRows`. Commander/Deck/Companion headers and comment lines, having no natural warning, are excluded from `NonBlankLines` entirely (treated like blank lines).
- A leading `-` followed by digits is read as a failed quantity attempt (one `Warning`, zero entries) rather than "no leading digits" (which would otherwise silently swallow the minus sign into the name) — required to satisfy the `-1 Sol Ring` corpus row.
- `DetectPasteSource`'s Moxfield threshold is a strict majority (`moxfieldLines*2 > entryLines`), matching "majority of entry lines" literally.
- Per 01-01-SUMMARY.md's precedent, `REQ-ACT-002` in `.planning/REQUIREMENTS.md` was deliberately NOT marked complete during the state-update step, even though it is this plan's sole `requirements` frontmatter entry. It is also claimed by plans 01-04 and 01-05 (card-name search/suggestions, and the `createLibraryFromDecklist` single-canonical-parse-site extraction), neither of which exists yet. Marking it complete now would misrepresent the requirement's actual traceability state to any later reader. `requirements-completed` in this SUMMARY's frontmatter still lists `REQ-ACT-002` per the template's instruction to record what this plan *contributed to*.

## Deviations from Plan

None — plan executed as written. The only mid-execution adjustment was structural (Task 1 not depending on Task 2's `sections.go`, and Task 2 replacing Task 1's minimal placeholder with the real tracker), which is scoping/sequencing already implied by the plan's own file-list split across the two tasks, not a change to any acceptance criterion, behavior, or test requirement. A one-line docs-only follow-up commit (`3396d0f`) added an explicit "Rule 7" comment that had been present in substance (folded into a "Rules 6-7" comment) but not literally numbered, to unambiguously satisfy the acceptance criterion "a numbered comment for each of the eight rules."

## Issues Encountered

None.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- `pkg/deckimport` now exposes the complete, pure, deterministic decklist grammar this phase's REQ-ACT-002 required, with a mechanical no-row-disappears guarantee gated in `go test ./pkg/... -race` — the only Go target CI runs today.
- Plan 01-04 (card-name search/suggestions) can consume `ParsedEntry.Name` as-is; nothing here changes name casing or normalization beyond what 01-04 will need to match against.
- Plan 01-05 (extracting `createLibraryFromDecklist` as the single canonical parse site in `server/games.go`) can call `deckimport.Parse` directly — this plan did not touch `server/`, so that extraction remains fully unstarted and unblocked.
- Plan 01-06's provider adapter, when it lands a URL-host detector per D-23, can reuse `ParseWithSource` unchanged to prove its own detection is equally inert.
- No blockers. `go vet ./pkg/deckimport` is clean, `go test ./pkg/... -race` is green, and all three golden fixtures round-trip against their checked-in expected `ParsedDeck`.

---
*Phase: 01-measured-deck-import-foundation*
*Completed: 2026-08-04*

## Self-Check: PASSED

All 13 files listed under `key-files` (created/modified) exist on disk, and all four
commit hashes (`d734e25`, `69cb8f3`, `3396d0f`, `3336efe`) are present in
`git log --oneline --all`.
