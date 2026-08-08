---
phase: 01-measured-deck-import-foundation
reviewed: 2026-08-08T00:05:55Z
depth: standard
files_reviewed: 4
files_reviewed_list:
  - server/deck_providers.go
  - server/deck_providers_test.go
  - server/deck_import.go
  - server/deck_import_test.go
findings:
  critical: 2
  warning: 3
  info: 1
  total: 6
status: issues_found
---

# Phase 1: Code Review Report (Delta — plans 01-08/01-09/01-10)

**Reviewed:** 2026-08-08T00:05:55Z
**Depth:** standard
**Files Reviewed:** 4
**Status:** issues_found

> **Scope note.** This is a delta review, not a fresh full-phase review. The
> phase was already reviewed at 48-file scope on 2026-08-05; that report is
> preserved at git commit `a92d8cc`, and its 2 critical / 4 warning / 1 info
> findings were resolved (see `01-REVIEW-FIX.md`). This report **supersedes**
> that one and covers only what plans 01-08 (Archidekt provider reversal),
> 01-09 (contract-pinning tests), and 01-10 touched since commit `d3a70e8`:
> `server/deck_providers.go`'s `archidektAdapter`, its two helper functions
> (`formatArchidektDeckLine`, `archidektCollectorNumberRoundTrips`), and the
> new test files pinning them to the committed live capture.

## Summary

The Moxfield→Archidekt reversal itself is clean: the host allowlist, SSRF
address/port controls (`safeControl`), the redirect re-validation
(`newDeckProviderCheckRedirect`), the body cap, the timeouts, and the
fail-closed kill switch (`providerEnabled`) are all byte-for-byte unchanged
by this delta (confirmed against `git diff d3a70e8..HEAD`) — the swap did
not weaken anything the 2026-08-05 review verified there.

The new code is `archidektAdapter.normalizeToDeckText` and
`formatArchidektDeckLine`: the seam that turns untrusted third-party JSON
into text a second, independent grammar (`pkg/deckimport`) re-segments into
lines, quantities, names, and section headers. The adapter's own comments
show real awareness of this risk class — `archidektCollectorNumberRoundTrips`
exists specifically because an unvalidated collector number ("MH1-216")
silently folded into a card's parsed *name* and made the card vanish. That
same defense was applied to exactly one of the three fields flowing through
`formatArchidektDeckLine` (`collectorNumber`) and not to the other two
(`name`, `setCode`), and the gap is exploitable: a single corrupted or
adversarial-owner-controlled Archidekt card name can inject an extra
decklist line — including a fake `Commander` header that reassigns section
membership for every entry that follows — with **no** accounting-invariant
trip (verified by direct reproduction below), and a `setCode` containing a
space silently corrupts the card name so the entry fails to resolve. Both
are the same "silent mis-parse, not a cosmetic metadata loss" failure mode
the codebase's own commentary explicitly calls out as unacceptable, just on
fields the fix wasn't extended to.

## Critical Issues

### CR-01: Unsanitized embedded newline/CR in card name injects lines into the decklist grammar, silently reassigning sections

**File:** `server/deck_providers.go:499-503`, `server/deck_providers.go:590-596`

**Issue:** `normalizeToDeckText` builds each line as
`fmt.Sprintf("%d %s", quantity, name)` (`formatArchidektDeckLine`,
line 591) from `name := strings.TrimSpace(row.Card.OracleCard.Name)`
(line 499), an untrusted string read straight out of the fetched
third-party JSON. Neither the trim nor anything else checks `name` for an
embedded `\n` or `\r`. `pkg/deckimport`'s `splitLines` (the *second*,
independent grammar this text is fed into via `deckimport.ParseWithSource`)
treats `\n`/`\r` as line terminators — so a single `cards[]` row whose name
contains one of these characters is silently split into two or more lines
by the downstream parser, and the injected second "line" is then
re-interpreted from scratch: it can match `matchSectionHeader` exactly
(e.g. an embedded `"Commander"`), reassigning `Section` for every entry
that follows it, or match `isCommentLine` and vanish outright.

Reproduced directly against the real adapter and the real parser
(no mocks): a card named `"Fake Card\nCommander"` followed by a
legitimate `"Sol Ring"` row produces the generated text
`"1 Fake Card\nCommander\n1 Sol Ring\n"`, which parses to
`Sol Ring` being placed in `SectionCommander` — a commander-candidate
hijack triggered by one corrupted/malicious upstream field, with **no**
`AssertAccounting` violation and no warning of any kind. This is exactly
the "silently fails... vanishes/misassigns" class of bug this same file's
own comment on `archidektCollectorNumberRoundTrips` describes as
unacceptable ("That is exactly the silent mis-parse this plan exists to
catch, not a cosmetic metadata loss") — except unlike the collector-number
case, here it reaches production with no guard at all, and it defeats the
accounting invariant the rest of the architecture relies on to make
mis-parses loud rather than silent.

Whether Archidekt's *own* servers ever emit a control character in
`oracleCard.name` is beside the point: this adapter's stated job (see this
file's header comment, lines 341-372) is to normalize "untrusted
third-party JSON" defensively, and Archidekt is known to allow
user-created custom/proxy cards with attacker-chosen names in a public
deck — exactly the vector by which one Archidekt account holder could
corrupt another player's import of a deck link they share.

**Fix:** Reject or normalize control characters in `name` before handing it
to `formatArchidektDeckLine`, the same way a missing name is already
rejected via `errArchidektMissingCardName`:

```go
func archidektNameIsSafe(name string) bool {
	for _, c := range name {
		if c == '\n' || c == '\r' {
			return false
		}
	}
	return true
}

// in the cards[] loop:
name := strings.TrimSpace(row.Card.OracleCard.Name)
if name == "" {
	return "", errArchidektMissingCardName
}
if !archidektNameIsSafe(name) {
	return "", errArchidektMissingCardName // or a new, equally static sentinel
}
```

### CR-02: `setCode` has no round-trip validation, unlike `collectorNumber` — a space or other unsafe character silently corrupts the card name

**File:** `server/deck_providers.go:590-596`

**Issue:** `formatArchidektDeckLine` guards `collectorNumber` with
`archidektCollectorNumberRoundTrips` before emitting the `"(setcode)
collectornumber"` suffix, but applies **no equivalent check to `setCode`**
(`row.Card.Edition.EditionCode`) — it is interpolated directly:
`fmt.Sprintf("%s (%s) %s", line, setCode, collectorNumber)`. If `setCode`
contains a space or tab, `pkg/deckimport`'s
`extractTrailingAnnotations` (scanner.go) requires the parenthesised group
to contain no whitespace to be recognized as a set code
(`!strings.ContainsAny(inner, " \t")`); when that check fails, the
`"(SET CODE)"` group is **not** stripped and instead stays glued onto the
card name, corrupting it.

Reproduced directly: a card named `"Sol Ring"` with `editioncode: "SET
CODE"` and `collectorNumber: "123"` (a value that *does* round-trip)
produces generated text `"1 Sol Ring (SET CODE) 123\n"`, which parses to
an entry named `"Sol Ring (SET CODE)"` — not `"Sol Ring"` — so the entry
fails to resolve against the cards table and silently becomes an
"unresolved" row instead of the real Sol Ring the player asked for. This
is the identical failure mode `archidektCollectorNumberRoundTrips` exists
to prevent, on the field that check does not cover. A `setCode` containing
`\n`/`\r` compounds with CR-01 above.

**Fix:** Extend the round-trip check to `setCode`, or reuse the existing
alnum check (Archidekt edition codes are ordinarily short alnum strings
like `cmr`, `mh1`, `plst`):

```go
func formatArchidektDeckLine(quantity int, name, setCode, collectorNumber string) string {
	line := fmt.Sprintf("%d %s", quantity, name)
	if setCode != "" && archidektCollectorNumberRoundTrips(setCode) &&
		archidektCollectorNumberRoundTrips(collectorNumber) {
		line = fmt.Sprintf("%s (%s) %s", line, setCode, collectorNumber)
	}
	return line
}
```
(Rename the shared helper to something like `archidektTokenRoundTrips` once
it covers both fields, and omit the whole suffix — never emit a corrupt
one — when either fails, exactly as the current comment already promises
for `collectorNumber` alone.)

## Warnings

### WR-01: No sanitization against the grammar's other special characters in `name` (quotes, backticks, brackets, hash-tags)

**File:** `server/deck_providers.go:499-503`

**Issue:** Beyond the newline case (CR-01), `name` is passed through
`formatArchidektDeckLine` with no defense against any of
`pkg/deckimport`'s other line-initial/trailing special syntax: a name
starting with `"` triggers the quoted-name rule and silently discards
everything after the first closing quote (`resolveNameAndMetadata`,
scanner.go rule 5); a name ending in `` ` ``, `]`, or a `#tag`-shaped
suffix is parsed as a trailing category annotation and stripped from the
name (`extractTrailingAnnotations`). Unlike CR-01, these do not defeat the
accounting invariant (the row still becomes exactly one `Entry`), but they
do silently corrupt the parsed name, which can misresolve the entry to a
different card than the one Archidekt actually named, or make a
resolvable card spuriously "unresolved." Since no official Magic card name
contains these characters, the realistic trigger is the same
custom/proxy-card vector as CR-01.

**Fix:** Once CR-01/CR-02 are fixed, consider whether the same guard
should be widened to reject (or escape) any of `"`, `` ` ``, `[`, `]`, `#`
at the position where the grammar treats them specially, rather than
enumerating characters one bug report at a time.

### WR-02: Deck-level category membership tests use exact string equality with no defensive case-folding, inconsistent with the `EqualFold` used for the "Commander" keyword

**File:** `server/deck_providers.go:488-525`

**Issue:** `commanderCategory` is discovered with
`strings.EqualFold(cat.Name, "Commander")` (line 492), but the exclusion
map (`excludedCategories[cat.Name]`, line 490) and both membership checks
against a row's `Categories` (`excludedCategories[cat]`, line 507;
`cat == commanderCategory`, line 520) are exact, case-sensitive string
comparisons. This relies on an unstated assumption that Archidekt always
echoes a row's category tag with byte-identical casing to the matching
deck-level `categories[].name` entry. If that assumption is ever wrong for
even one deck (a data inconsistency on Archidekt's side, or a future API
revision), a maybeboard-tagged card would silently land in the main deck
instead of being dropped — the opposite of D-11's "never silently
discarding" intent, in the other direction (silently *including* what
should have been excluded).

**Fix:** Either document explicitly why exact-match is safe here (e.g.
cite the Archidekt API guarantee), or fold both sides consistently:

```go
excludedCategories[strings.ToLower(cat.Name)] = struct{}{}
...
if _, ok := excludedCategories[strings.ToLower(cat)]; ok { ... }
```

### WR-03: Contract tests do not cover adversarial/malformed field content — only the happy-path fixture

**File:** `server/deck_providers_test.go` (all `TestArchidekt_*` tests)

**Issue:** Every `TestArchidekt_*` test in this file loads the real,
clean fixture and asserts against its known-good shape. None of them
exercises a card name or edition code containing a newline, quote,
bracket, backtick, or embedded whitespace — precisely the class of input
CR-01/CR-02 above show is silently mishandled. `TestArchidekt_
MalformedInputFailsClosed` covers *structural* malformance (missing
`cards[]`, missing `oracleCard.name`) but not *content* malformance within
an otherwise well-formed row. As written, none of these tests would have
caught CR-01 or CR-02, and none would fail if a future edit reintroduced
either.

**Fix:** Add a small, fixture-independent unit test analogous to the one
`archidektCollectorNumberRoundTrips` already implies was needed for
collector numbers — feed `normalizeToDeckText` a synthetic row whose
`oracleCard.name` contains `\n` and assert either an error is returned or
the resulting parsed entry count/section membership is unaffected; do the
same for a `setCode` containing a space.

## Info

### IN-01: `errArchidektMissingCardName`-style hard failure on one bad row blocks the entire deck, by design

**File:** `server/deck_providers.go:497-502`

**Issue:** A single `cards[]` row with an empty name causes
`normalizeToDeckText` to fail the whole deck (no partial import), per the
explicit design intent documented on `TestArchidekt_
MalformedInputFailsClosed` ("never a partial deck"). This is a reasonable,
deliberate choice for a genuinely malformed API response, but it means a
single Archidekt-side data-quality issue on an otherwise-valid hundred-row
deck (e.g. one placeholder/incomplete custom card with no name set) blocks
the player from importing anything from that deck at all, with a generic
"paste your decklist as text instead" message that gives no hint which
row was the problem. Noting for awareness only — not asking for a change,
since the tradeoff (fail loud vs. guess silently) is explicitly a decision
this phase already made and documented.

---

_Reviewed: 2026-08-08T00:05:55Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
