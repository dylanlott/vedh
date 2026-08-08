---
phase: 01-measured-deck-import-foundation
reviewed: 2026-08-08T07:36:26Z
iteration: 2
depth: standard
files_reviewed: 4
files_reviewed_list:
  - server/deck_providers.go
  - server/deck_providers_test.go
  - server/deck_import.go
  - server/deck_import_test.go
findings:
  critical: 0
  warning: 1
  info: 3
  total: 4
status: issues_found
---

# Phase 1: Code Review Report (Delta re-review, iteration 2 — plans 01-08/01-09/01-10 + review-fix commits)

**Reviewed:** 2026-08-08T07:36:26Z
**Depth:** standard
**Files Reviewed:** 4
**Status:** issues_found

> **Scope note.** This is iteration 2 of the delta review of the same four
> files. It **supersedes** both the 2026-08-05 full-phase review (preserved
> at git commit `a92d8cc`) and iteration 1 of this delta review (2026-08-08,
> `01-REVIEW.md` prior revision), which found CR-01, CR-02, WR-01, WR-02,
> WR-03, IN-01 and were then fixed across four commits (`d2d8df6`, `53d9f60`,
> `3e5786e`, `e628f39`). This pass's job was to adversarially re-examine
> those fixes themselves, not merely confirm they exist.

## Summary

All five in-scope findings from iteration 1 (CR-01, CR-02, WR-01, WR-02,
WR-03) were re-verified directly against `pkg/deckimport/scanner.go` and
`pkg/deckimport/sections.go` (read fresh for this pass, not inferred from
the adapter's own comments), and against the actual committed code, not the
patch descriptions:

- `archidektNameIsSafe`'s four checks (embedded `\n`/`\r` anywhere; a
  leading `"`; a trailing `` ` `` or `]`; a boundary-preceded trailing
  `#tag`) were traced rune-by-rune against `splitLines`,
  `resolveNameAndMetadata`, and `extractTrailingAnnotations`'s real,
  iterative hash-stripping loop. Every case constructed to find a gap
  (multi-tag trailing sequences, a `#` at index 0, a tag containing a
  backtick, IPv6-style Unicode line separators U+0085/U+2028/U+2029, a
  leading digit in the name, a name ending in a multi-word parenthetical
  like `Erase (Not the Urza's Legacy One)`) turned out to be either already
  covered by an existing check or structurally unreachable (e.g. `splitLines`
  is byte-oriented and never treats U+0085/U+2028/U+2029 as terminators, and
  the quantity prefix `fmt.Sprintf("%d %s", ...)` means a name's own leading
  digit is never re-consumed by `extractQuantity`, which only scans once at
  true line-start). No gap was found.
- A dedicated probe (`TestProbe_RealCardNamesWithParens`, run against this
  checkout and removed afterward — not committed) fed six real, tricky
  Magic card names through `archidektNameIsSafe` and a full
  normalize-then-parse round trip: `B.O.B. (Bevy of Beebles)`,
  `Erase (Not the Urza's Legacy One)`, `Look at Me, I'm the DCI`,
  `Ach! Hans, Run!`, `Kongming, "Sleeping Dragon"`, and
  `Yargle and Multani`. All six pass the guard and round-trip byte-identical
  through `Parse` — no false positive was found (see also this repo's own
  `TestArchidekt_InternalQuoteInNameSurvives`, which already covers the
  `Kongming` case).
- `archidektTokenRoundTrips` is applied to **both** `setCode` and
  `collectorNumber` with `&&`, and `formatArchidektDeckLine` has exactly one
  branch — there is no second code path where the suffix could be emitted
  with only one side validated.
- `go test ./server/...` (all `TestArchidekt_*`, `TestSafeControl_*`,
  `TestSafeClient_*`, `TestProvider_*`) passes clean on this checkout.

No new Critical or Blocker was found this pass. One Warning below revisits
the "reject the whole deck on one bad row" failure mode (iteration 1's
`IN-01`, at Info level, on a narrower trigger) now that WR-01 has
substantially widened the set of inputs that can trip it — this iteration's
own explicit ask (point 4) — and three Info items note smaller
comment-accuracy and test-rigor gaps in the new code.

## Warnings

### WR-01: `errArchidektUnsafeCardName` fails the *entire* deck on one row, and the guard surface that can trigger it just grew substantially

**File:** `server/deck_providers.go:509-514`, `server/deck_providers.go:600-620`

**Issue:** `normalizeToDeckText` returns immediately —
`return "", errArchidektUnsafeCardName` — the moment any single `cards[]`
row's name fails `archidektNameIsSafe`. This is not "reject the row" (the
function's own comment's wording, line 598: "Rejecting the whole row when
any of these trip"); it is reject the *entire deck normalization*: every
other row, including ones already appended to `mainLines`/`commanderLines`
in a prior loop iteration, is discarded, and `previewDeckURL`
(`server/deck_import.go:275-280`) turns that into the same generic
"We couldn't load that deck link right now. Paste your decklist as text
instead." message used for a network failure or malformed JSON —
`CanContinue = false`, no partial preview, and no indication of which row
or which character was the problem.

Iteration 1 flagged this same fail-whole-deck shape at Info level
(`IN-01`), but only for the single, narrow trigger that existed then (an
empty `oracleCard.name` — an unambiguous provider malformation signal).
WR-01's fix widened `archidektNameIsSafe` to four independent trigger
conditions (embedded `\n`/`\r`, leading `"`, trailing `` ` ``/`]`, trailing
`#tag`) across every row of a deck that can have hundreds of rows. None of
the six real card names probed above trips it, so today's practical risk
stays low — but the failure mode itself is now reachable by a much larger
input surface than the case iteration 1 accepted, and the asymmetry with
this same file's own Maybeboard-exclusion design (D-11: drop the *one* bad
row, count it, keep going) is more visible now that there is a real
per-row-exclusion precedent sitting right next to a per-row-rejection path
that instead aborts everything.

**Fix:** Not a required change (the fail-loud tradeoff is a defensible,
previously-accepted design decision, and this finding does not block
shipping), but worth a deliberate re-decision now that the trigger surface
has grown: consider treating an unsafe name the same way an unresolved
card name is already treated — emit the row as a `DeckImportIssue` /
`Warning` naming the reason ("this card's name could not be safely
imported") and continue with the rest of the deck — rather than aborting
the whole response. If the all-or-nothing choice is kept, at minimum fix
the misleading "the whole row" wording in the comment at
`server/deck_providers.go:598` to say what actually happens (the whole
response/deck), so a future reader isn't misled about the blast radius.

## Info

### IN-01: Adversarial rejection tests assert only `err != nil`, not the specific sentinel

**File:** `server/deck_providers_test.go:1021-1054`, `server/deck_providers_test.go:1186-1215`

**Issue:** `TestArchidekt_UnsafeCardNameRejected` and
`TestArchidekt_UnsafeNameGrammarConflictRejected` both check
`if err == nil { t.Fatalf(...) }` but never assert
`errors.Is(err, errArchidektUnsafeCardName)`. For the inputs these two
tests use this is not currently exploitable (no other code path in
`normalizeToDeckText` returns a non-nil error for a non-empty name and
otherwise-valid JSON), so the tests are not tautological today — but they
would silently keep "passing" if a future edit accidentally routed these
same inputs through a *different* error path (e.g. if `archidektNameIsSafe`
were folded into the empty-name check and started returning
`errArchidektMissingCardName` instead), which would mask exactly the kind
of regression these tests exist to catch.

**Fix:** Add `if !errors.Is(err, errArchidektUnsafeCardName) { t.Fatalf(...) }`
alongside the existing `err == nil` check in both tests, matching the
precision `TestArchidekt_MalformedInputFailsClosed` and
`TestProvider_ArchidektAdapterMalformedBody` already use elsewhere in this
file.

### IN-02: `archidektTokenRoundTrips` is stricter than the grammar it's modeling for `setCode`, silently dropping recoverable D-07 metadata

**File:** `server/deck_providers.go:640-650`

**Issue:** `archidektTokenRoundTrips` requires ASCII alnum-only content for
both `setCode` and `collectorNumber`. That is the exact requirement
`extractTrailingAnnotations`'s bare-token rule (`isAlnumToken`) imposes on
a trailing collector number, so applying it there is precise. But the
*set-code* rule in the real grammar (`extractTrailingAnnotations`'s
parenthesised-group check) only forbids internal whitespace —
`!strings.ContainsAny(inner, " \t")` — and otherwise accepts any
character, including a hyphen. A hypothetical `setCode` like `"foo-bar"`
would round-trip fine through the real parser but is rejected by
`archidektTokenRoundTrips`, needlessly omitting the whole D-07
printing-metadata suffix (this direction is always safe — it degrades
rather than corrupts — so this is not a correctness bug). Real Archidekt
`editioncode` values are ordinarily short pure-alnum strings (`cmr`,
`mh1`), so the practical impact today is minimal.

**Fix:** No action required; noting for awareness. If a future Archidekt
response is observed with a punctuated edition code that this drops
unnecessarily, `archidektTokenRoundTrips` could be split into two
predicates (one alnum-only for `collectorNumber`, one alnum-plus-select-
punctuation for `setCode`) to match the two different underlying grammar
rules exactly instead of reusing one predicate for both.

### IN-03: A zero or negative `quantity` from the provider degrades to a silent per-row warning with no adapter-level test

**File:** `server/deck_providers.go:515`, `server/deck_providers.go:663-669`

**Issue:** `formatArchidektDeckLine` interpolates `row.Quantity` directly
with no validation. If Archidekt ever returns `quantity: 0` or a negative
value for a row, the generated line (e.g. `"0 Sol Ring"`) is handled
correctly downstream — `extractQuantity` (scanner.go) rejects a
non-positive value and turns the row into a `Warning`, which
`AssertAccounting` still balances — so this is not a correctness bug. It
is, however, untested at the adapter level: none of the
`TestArchidekt_*`/`archidektSyntheticResponse`-based tests construct a row
with `Quantity: 0` or negative to confirm the adapter's output degrades
exactly this way rather than, say, silently coercing to `1`.

**Fix:** Optional; a small synthetic-response test
(`archidektSyntheticResponse` with `Quantity: 0`) asserting the resulting
`ParsedDeck` carries one `Warning` and zero `Entries` for that row would
close this gap cheaply, consistent with `TestArchidekt_UnsafeSetCodeOmitsSuffix`'s
existing pattern of testing one synthetic anomaly at a time.

---

_Reviewed: 2026-08-08T07:36:26Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
