---
phase: 01-measured-deck-import-foundation
fixed_at: 2026-08-08T05:08:38Z
review_path: .planning/phases/01-measured-deck-import-foundation/01-REVIEW.md
iteration: 1
findings_in_scope: 5
fixed: 5
skipped: 0
status: all_fixed
---

# Phase 1: Code Review Fix Report

**Fixed at:** 2026-08-08T05:08:38Z
**Source review:** .planning/phases/01-measured-deck-import-foundation/01-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 5 (fix_scope: critical_warning — CR-01, CR-02, WR-01, WR-02, WR-03; IN-01 excluded as info-only)
- Fixed: 5
- Skipped: 0

All work was done in an isolated git worktree (`gsd-reviewfix/01-<pid>` branched from `main`),
fast-forwarded back onto `main` on completion — no rebasing or history rewriting of the user's
branch. Verification (build/vet/tests) ran inside that same worktree, with a symlink to the
repo-root `All Printings.json` (gitignored, not checked out into new worktrees) so the package's
`TestMain` DB-seeding step could run identically to the main checkout.

## Fixed Issues

### CR-01: Unsanitized embedded newline/CR in card name injects lines into the decklist grammar

**Files modified:** `server/deck_providers.go`, `server/deck_providers_test.go`
**Commit:** `d2d8df6` (source fix), `53d9f60` (regression test)
**Applied fix:** Added `archidektNameIsSafe`, called in the `cards[]` loop right after the existing
empty-name check, returning a new static sentinel `errArchidektUnsafeCardName` when a name contains
an embedded `\n` or `\r`. This mirrors the existing fail-closed pattern for
`errArchidektMissingCardName` rather than inventing a parallel mechanism. Regression coverage added
in `TestArchidekt_UnsafeCardNameRejected` (synthetic rows with `\n` and `\r`), independently
confirmed to fail against the pre-fix code (reverted the source file to the pre-fix commit, ran the
new test, observed both subtests fail with the exact corrupted text the review/verification
reports describe, then restored the fix).

### CR-02: `setCode` has no round-trip validation, unlike `collectorNumber`

**Files modified:** `server/deck_providers.go`, `server/deck_providers_test.go`
**Commit:** `d2d8df6` (source fix, same commit as CR-01 — same helper functions, same fix location,
fixed together per the orchestrator's guidance), `53d9f60` (regression test)
**Applied fix:** Generalized `archidektCollectorNumberRoundTrips` into `archidektTokenRoundTrips`
(same alnum-only grammar check, now documented for both fields) and updated
`formatArchidektDeckLine` to require **both** `setCode` and `collectorNumber` to round-trip before
emitting the D-07 printing-metadata suffix — omitting the whole suffix, never emitting it with only
one side correct, exactly as the existing comment already promised for `collectorNumber` alone.
Regression coverage added in `TestArchidekt_UnsafeSetCodeOmitsSuffix` (setCode `"SET CODE"` with a
round-tripping collector number), independently confirmed to fail against the pre-fix code (parsed
entry name carried the corrupted `"Sol Ring (SET CODE)"` suffix, exactly as reported).

### WR-01: No sanitization against the grammar's other special characters in `name`

**Files modified:** `server/deck_providers.go`, `server/deck_providers_test.go`
**Commit:** `e628f39`
**Applied fix:** Widened `archidektNameIsSafe` (rather than enumerating characters one bug report at
a time, per the review's own framing) to reject a name that: starts with `"` (triggers the CSV-style
quoted-name rule); or ends with `` ` ``, `]`, or a token-boundary-preceded `#tag`-shaped suffix
(triggers `extractTrailingAnnotations`'s trailing-category rules). Deliberately **position-specific**
rather than a blanket "reject any occurrence" rule: an internal, non-leading quote is a legal
component of a real Magic card name (e.g. `Kongming, "Sleeping Dragon"`), and a naive
contains-check would have incorrectly excluded it. Added
`TestArchidekt_UnsafeNameGrammarConflictRejected` (four synthetic adversarial cases: leading quote,
trailing backtick category, trailing bracket category, trailing hash tag) and
`TestArchidekt_InternalQuoteInNameSurvives` (proving the real `Kongming, "Sleeping Dragon"` name is
NOT rejected and round-trips intact). All four rejection cases independently confirmed to fail
against the pre-fix code; the internal-quote survival case passes both before and after, confirming
it is not a false positive introduced by an overly broad check.

### WR-02: Deck-level category membership tests use exact string equality, inconsistent with `EqualFold`

**Files modified:** `server/deck_providers.go`, `server/deck_providers_test.go`
**Commit:** `3e5786e`
**Applied fix:** Folded both `excludedCategories` (map keys and lookups) and the commander-category
comparison to lower-case, consistent with the `strings.EqualFold` already used to *find* the
Commander category. A case mismatch between a row's category tag and the deck-level
`categories[].name` entry now can no longer silently include an excluded (maybeboard) card, nor
silently miss the commander. Added `TestArchidekt_CategoryMembershipIsCaseFolded` (a synthetic
response with `"maybeboard"`/`"Maybeboard"` and `"COMMANDER"`/`"Commander"` case mismatches),
independently confirmed to fail against the pre-fix code (the case-mismatched maybeboard row landed
in `Entries` instead of `DroppedSectionRows`).

### WR-03: Contract tests do not cover adversarial/malformed field content

**Files modified:** `server/deck_providers_test.go`
**Commit:** `53d9f60` (initial CR-01/CR-02 regression tests), extended by `3e5786e` and `e628f39`
(WR-02/WR-01 regression tests)
**Applied fix:** Added five new fixture-independent tests exercising synthetic, adversarial
`cards[]` rows (as opposed to every other test in this file, which loads the real committed
fixture): `TestArchidekt_UnsafeCardNameRejected`, `TestArchidekt_UnsafeSetCodeOmitsSuffix`,
`TestArchidekt_CategoryMembershipIsCaseFolded`, `TestArchidekt_UnsafeNameGrammarConflictRejected`,
`TestArchidekt_InternalQuoteInNameSurvives`. Every test that asserts a defect was rejected was
verified to genuinely FAIL against the pre-fix code (not merely pass either way) by temporarily
reverting the corresponding source commit, re-running the new test, observing the failure with the
exact corrupted output the review/verification reports describe, then restoring the fix.

## Skipped Issues

None — all 5 in-scope findings were fixed.

(IN-01 — the deliberate all-or-nothing fail-closed design for a malformed row — was excluded from
scope per `fix_scope: critical_warning`; it is Info-severity and the review explicitly frames it as
a documented, deliberate tradeoff, not a defect requiring a change.)

## Verification

All commands run inside the isolated worktree (fast-forwarded onto `main` afterward), with a
symlinked `All Printings.json` at the worktree root to satisfy `server`'s `TestMain` DB-seed step
(the real file is 623 MB and gitignored, so it is not present in a fresh worktree checkout by
default — this is a test-harness environment detail, not a change to tracked source).

| Command | Result |
|---|---|
| `go build ./...` | Clean, no errors |
| `go vet ./server` | Clean, no errors |
| `go test ./server -count=1` | `ok  github.com/openmtg/edh-go/server  32.335s` — all tests pass, including the 5 new adversarial regression tests and the full pre-existing `TestArchidekt_*`/`TestProvider_*`/`TestDeckImport_ArchidektURLPath` suite |

Each of the 4 commits was independently verified before being made permanent: `go build ./...` and
`go vet ./server` after every edit, the targeted `TestArchidekt_*` subset re-run after every edit,
and — for every new adversarial test — a temporary revert-and-rerun cycle proving the test fails
against the pre-fix code before restoring the fix and moving on. No existing assertion was weakened
to make a test pass.

---

_Fixed: 2026-08-08T05:08:38Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
