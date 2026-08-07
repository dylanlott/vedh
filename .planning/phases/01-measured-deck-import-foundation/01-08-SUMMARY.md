---
phase: 01-measured-deck-import-foundation
plan: 08
subsystem: api
tags: [go, deckimport, ssrf, archidekt, deck-provider]

# Dependency graph
requires:
  - phase: 01-measured-deck-import-foundation
    provides: "01-06's secure outbound fetch client (fetchDeckProviderURL, safeControl, newSafeProviderClient), providerEnabled() kill switch, deckProviderAdapter interface and registry, and 01-07's D-14 checkpoint/Moxfield scaffold that this plan reverses"
  - phase: 01-measured-deck-import-foundation
    provides: "01-02's canonical pkg/deckimport grammar (Parse/ParseWithSource, section headers, D-07 printing metadata, D-11 dropped-section accounting) that archidektAdapter emits text into"
provides:
  - "The reversed D-14 decision (Archidekt, not Moxfield), dated and attributed to the user at UAT, recorded without deleting the original Moxfield decision or its evidence"
  - "archidektAdapter.normalizeToDeckText, implemented against a real observed response contract rather than a guess, registered under the archidekt.com hostname"
  - "Closure of the moxfieldAdapter stub window in .planning/WINDOWS.md"
affects: ["01-09 (contract tests pinning archidektAdapter's field mapping)", "01-10 (traceability closure for REQ-ACT-003)"]

actuals:
  tokens: 9500
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Provider adapter normalizes into canonical decklist text (not hand-built ParsedEntry values), reusing pkg/deckimport's existing grammar and section-header accounting for a single canonical parse path"
    - "Adapter error hygiene via fixed static sentinels (errArchidektMalformedResponse, errArchidektNoCardRows, errArchidektMissingCardName), never wrapping raw response content, so logging normalizeToDeckText's error is safe by construction"

key-files:
  created: []
  modified:
    - server/deck_providers.go
    - server/deck_import.go
    - server/deck_providers_test.go
    - docs/research/deck-provider-feasibility.md
    - .planning/WINDOWS.md

key-decisions:
  - "D-14 reversed at UAT (gap G-01-1): Archidekt replaces Moxfield as the first supported provider, because Moxfield's response contract is unobtainable without authorized API access (blanket robots.txt disallow plus Cloudflare 403s), while Archidekt's contract was directly observed and captured"
  - "Moxfield scaffold removed outright (moxfieldAdapter, errMoxfieldContractUnverified, moxfieldHost) rather than left dormant, per explicit user direction recorded in G-01-1 -- re-addable later if authorized Moxfield access is obtained; deckimport.SourceMoxfield itself is untouched since the SourceType enum is closed at four values"
  - "Human precondition recorded, not resolved: enabling DECK_PROVIDER_ENABLED for archidekt.com in a real deployment requires a human to read https://archidekt.com/terms in a real browser first -- it is JS-rendered and has never been read by any agent; building/testing the adapter does not require this"
  - "Bucketing order in normalizeToDeckText checks exclusion (includedInDeck: false) before the commander check, because a card can carry a maybeboard tag and its eventual commander category simultaneously in Archidekt's data model"

patterns-established:
  - "Deck-level categories[] with includedInDeck is the authority on inclusion, not first-category-wins -- a card's full categories slice is checked against the excluded set"

requirements-completed: []

coverage:
  - id: D1
    description: "D-14 decision reversed: Archidekt recorded as first supported provider, attributed to the user at UAT, superseding (not deleting) the original Moxfield selection and its evidence"
    requirement: "REQ-ACT-003"
    verification:
      - kind: other
        ref: "grep -c Moxfield docs/research/deck-provider-feasibility.md (36, non-zero -- original decision survives); grep -n '^## 6' (new dated section present)"
        status: pass
      - kind: other
        ref: "gsd-tools windows fixed 1 -- moxfieldAdapter stub window closed, open_count: 0"
        status: pass
    human_judgment: false
  - id: D2
    description: "archidektAdapter.normalizeToDeckText decodes a real Archidekt response into canonical decklist text pkg/deckimport.Parse consumes, using only field paths present in the committed live capture, and the Moxfield scaffold (moxfieldAdapter, errMoxfieldContractUnverified, moxfieldHost) is fully removed"
    requirement: "REQ-ACT-003"
    verification:
      - kind: other
        ref: "go build ./... && go vet ./server; grep -c 'moxfield\\|Moxfield' server/deck_providers.go (0); grep -in moxfield server/deck_import.go (0); grep -n archidekt.com server/deck_providers.go (registered)"
        status: pass
      - kind: unit
        ref: "go test ./server -run Provider -v (TestProvider_ArchidektAdapterRoutesToRealNormalizer, TestProvider_ArchidektAdapterMalformedBody, TestProvider_DeckProviderAdapterFor, TestProvider_UnregisteredHostNeverFetched, TestProvider_KillSwitchStableAcrossAdapterChange -- all pass)"
        status: pass
      - kind: unit
        ref: "go test ./server -count=1 && go test ./pkg/... -race -count=1 (full suites, zero failures)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Field-mapping correctness against the real observed contract: a 126-total-quantity, 113-row Archidekt response yields exactly a 100-card main+commander deck once the includedInDeck:false Maybeboard category is dropped by the existing grammar's one-summary-warning accounting"
    requirement: "REQ-ACT-003"
    verification:
      - kind: other
        ref: "Ad-hoc uncommitted test against server/testdata/deck_providers/archidekt_deck_2026-08-05.json: entries=87 total_quantity=100 dropped=26 (Maybeboard warning), commander 'Cid, Timeless Artificer' present -- confirms the must-have numeric truth, but this is NOT a persisted automated check"
        status: pass
    human_judgment: true
    rationale: "Plan 01-09 is explicitly scoped to pin this exact field mapping with real, committed fixture-contract tests against server/testdata/deck_providers/archidekt_deck_2026-08-05.json. This plan deliberately does not do that work (see plan objective). The numeric result was hand-verified during execution and passed, but no permanent test in this codebase asserts it yet, so a human (or 01-09's own verification) should confirm before this is trusted long-term."
  - id: D4
    description: "Human precondition for enabling the kill switch on archidekt.com is recorded in the revised decision (archidekt.com/terms is JS-rendered and has never been read by any agent) -- this is a documentation deliverable, not a code gate, since providerEnabled() already defaults off regardless"
    verification: []
    human_judgment: true
    rationale: "Reading and evaluating https://archidekt.com/terms requires a human with a real browser; no agent working on this project has JavaScript execution capability. This is inherently a human action, not something automatable within this plan."

duration: 55min
completed: 2026-08-07
status: complete
---

# Phase 01 Plan 08: Archidekt adapter — decision reversal and normalizer against the observed contract Summary

**Reversed D-14 from Moxfield to Archidekt and implemented `archidektAdapter.normalizeToDeckText` against the real, committed 568KB Archidekt response capture, proving a 126-total-quantity response correctly yields a 100-card deck once the Maybeboard category is excluded — while removing the Moxfield scaffold outright.**

## Performance

- **Duration:** 55 min
- **Started:** 2026-08-07T22:12:17Z
- **Completed:** 2026-08-07T23:07:00Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Appended a dated, attributed section 6 to `docs/research/deck-provider-feasibility.md` recording the UAT reversal from Moxfield to Archidekt (gap G-01-1), leaving section 5's original Moxfield decision and evidence fully intact
- Closed the `moxfieldAdapter.normalizeToDeckText` stub window in `.planning/WINDOWS.md` (`gsd-tools windows fixed 1`)
- Implemented `archidektAdapter` against the observed contract (`categories[]`/`cards[]`), bucketing every card row into main, commander, or excluded (Maybeboard) text and letting `pkg/deckimport`'s existing grammar drop the excluded rows with its one counted summary warning
- Verified against the real committed fixture that the 113-row, 126-total-quantity response correctly resolves to 100 cards with the commander (`Cid, Timeless Artificer`) correctly identified
- Removed `moxfieldAdapter`, `errMoxfieldContractUnverified`, and `moxfieldHost` outright; rewrote the now-stale `previewDeckURL` routing doc comment so no clause of it is false against the shipped code
- Retargeted (not deleted) the kill-switch, allowlist, and routing-gate test coverage in `server/deck_providers_test.go` from `moxfield.com` to `archidekt.com`

## Task Commits

Each task was committed atomically:

1. **Task 1: Record the reversed decision** - `17cd1b8` (docs)
2. **Task 2: Implement archidektAdapter, remove the Moxfield scaffold** - `f119782` (feat)

## Files Created/Modified
- `docs/research/deck-provider-feasibility.md` - New dated section 6 recording the Archidekt reversal, Moxfield-still-viable-with-a-key note, carried-forward Archidekt risk, and the terms-of-service human precondition
- `.planning/WINDOWS.md` - `moxfieldAdapter` stub window closed (`status: fixed`)
- `server/deck_providers.go` - `archidektAdapter` (source, normalizeToDeckText, supporting structs, `formatArchidektDeckLine`), three static error sentinels, `archidektHost` constant, updated registry; `moxfieldAdapter`/`errMoxfieldContractUnverified`/`moxfieldHost` removed
- `server/deck_import.go` - Rewrote the stale `previewDeckURL` doc comment and its two inline comments that referenced the now-removed adapter, so every clause is true against the shipped code
- `server/deck_providers_test.go` - Retargeted kill-switch/allowlist/routing tests to `archidekt.com`; replaced the Moxfield-adapter-specific test with direct `archidektAdapter` sentinel-error tests; added a real routes-to-real-normalizer test

## Decisions Made
- D-14 reversed to Archidekt at UAT (gap G-01-1) — see Key Decisions in frontmatter for the full rationale
- Moxfield scaffold removed outright rather than left dormant, per the user's explicit direction in `01-UAT.md`'s `G-01-1.missing` list
- Bucketing checks category exclusion before the commander check, since a row can carry both tags simultaneously in Archidekt's data model
- Human precondition (reading `archidekt.com/terms`) recorded as a prerequisite for *enabling* the kill switch, not for building or testing the adapter

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Retargeted `server/deck_providers_test.go`, not listed in this plan's `<files>`**
- **Found during:** Task 2 (implementing `archidektAdapter`, removing `moxfieldAdapter`)
- **Issue:** Removing `moxfieldAdapter`, `errMoxfieldContractUnverified`, and `moxfieldHost` broke compilation of `server/deck_providers_test.go`, which referenced all three across seven tests. The plan's own `<verify>` step for Task 2 explicitly requires `go test ./server -count=1` to pass and explicitly instructs retargeting (not deleting) kill-switch/allowlist coverage.
- **Fix:** Retargeted `TestProvider_KillSwitch`, `TestProvider_KillSwitchRequiresAllowlist`, `TestProvider_PasteUnaffectedByFlag`, `TestProvider_DeckProviderAdapterFor`, and `TestProvider_UnregisteredHostNeverFetched` from `moxfield.com`/`moxfieldHost` to `archidekt.com`/`archidektHost`, preserving every assertion's original intent. Deleted `TestProvider_MoxfieldContractUnverified` (it tested a removed type with no analog) and replaced it with `TestProvider_ArchidektAdapterMalformedBody`, which asserts the same "static sentinel, never carries input bytes" property against the real adapter. Renamed `TestProvider_MoxfieldAdapterRoutesThenFailsClosed` to `TestProvider_ArchidektAdapterRoutesToRealNormalizer` (same routing-gate proof, now against a real normalizer that fails closed on the spy's nil body because that body isn't valid JSON, not because the normalizer is unimplemented). Renamed `TestProvider_MoxfieldScaffoldPathIsStable` to `TestProvider_KillSwitchStableAcrossAdapterChange`.
- **Files modified:** `server/deck_providers_test.go`
- **Verification:** `go build ./...`, `go vet ./server`, `go test ./server -count=1`, `go test ./pkg/... -race -count=1` all pass with zero failures
- **Committed in:** `f119782` (Task 2 commit)

**2. [Rule 2 - Missing Critical] Removed every literal mention of "Moxfield"/"moxfield" from `server/deck_providers.go` and `server/deck_import.go`**
- **Found during:** Task 2, after the first draft of the new comment block
- **Issue:** The plan's own `<verify>` step requires `grep -c "moxfield\|Moxfield" server/deck_providers.go` and `grep -in "moxfield" server/deck_import.go` to both be zero. My first draft of the historical-context comment named "Moxfield" repeatedly (for narrative clarity), which would have failed that exact verify step.
- **Fix:** Rewrote both comment blocks to describe the reversed decision without naming the previous provider by identifier, pointing readers to `docs/research/deck-provider-feasibility.md` section 6 for the full attributed account instead. No narrative content was lost — it now lives in the doc plan's own Task 1 wrote.
- **Files modified:** `server/deck_providers.go`, `server/deck_import.go`
- **Verification:** Both grep checks return zero
- **Committed in:** `f119782` (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 blocking test-compile fix, 1 verify-compliance comment rewrite)
**Impact on plan:** Both were necessary to satisfy this plan's own stated `<verify>` gates. No scope creep — no new product behavior was added beyond what Task 2 specified.

## Issues Encountered
None beyond the deviations above.

## User Setup Required

None for building or testing this plan's changes. **Before `DECK_PROVIDER_ENABLED`/`DECK_PROVIDER_ALLOWED_HOSTS` may be configured for `archidekt.com` in any real deployment**, a human must read `https://archidekt.com/terms` in a real browser — see `docs/research/deck-provider-feasibility.md` section 6, "Human precondition for enabling — not for building." No environment variable, dashboard step, or verification command is being deferred; this is a one-time human reading task with no machine-checkable proxy.

## Next Phase Readiness

This plan delivers the decision and the normalizer, honestly and completely for its own scope. It does NOT deliver:
- Plan 01-09's fixture-contract tests pinning the field mapping (deliverable D3 above is human-flagged for exactly this reason)
- Plan 01-10's traceability closure

**REQ-ACT-003 is deliberately left incomplete** in this SUMMARY's `requirements-completed` (empty array) — the requirement's substance (a working, contract-tested first-provider import) is not fully proven until 01-09 lands, following the same honesty precedent 01-07's SUMMARY set. The routing, normalizer, and error-hygiene invariants are real and tested; the fixture-contract pinning that makes the mapping trustworthy against future upstream drift is not yet in place.

No blockers for 01-09: `archidektAdapter`, the fixture, and the field paths it reads are all in place and stable for that plan to write tests against.

---
*Phase: 01-measured-deck-import-foundation*
*Completed: 2026-08-07*
