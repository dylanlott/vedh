---
phase: 01-measured-deck-import-foundation
verified: 2026-08-07T18:15:00Z
status: gaps_found
score: 4/5 must-haves verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: human_needed
  previous_score: 5/5
  gaps_closed:
    - "REQ-ACT-003 traceability sign-off (G-01-1): D-14 reversed from Moxfield to Archidekt; archidektAdapter implemented against the real observed contract, pinned with fixture-contract tests (01-08/01-09), and traceability independently re-audited clause-by-clause (01-10). Moxfield scaffold removed outright; WINDOWS.md stub window closed."
  gaps_remaining: []
  regressions:
    - "New gap introduced by the gap-closure work itself: archidektAdapter.normalizeToDeckText has no defense against control characters or unsafe setCode content in untrusted upstream fields, unlike the sibling collectorNumber field the same plan hardened (see gaps below). Not present in the prior verification because the Archidekt adapter did not exist at that time."
gaps:
  - truth: "Phase goal / Success Criterion 2: card resolution and section assignment for a decklist imported through the (allowlisted) provider path must be trustworthy — a resolvable card must never be silently reassigned to a different section or silently corrupted into unresolved with no warning."
    status: failed
    reason: >-
      Independently reproduced (not merely taken from 01-REVIEW.md's narrative) against the
      current archidektAdapter.normalizeToDeckText: a card row whose OracleCard.Name contains an
      embedded '\n' (e.g. "Fake Card\nCommander") is emitted verbatim into the generated decklist
      text with no sanitization. pkg/deckimport's splitLines treats '\n'/'\r' as line terminators,
      so the single corrupted row is silently split into two lines by the second, independent
      grammar the adapter's own text feeds into; the injected second line can exactly match
      matchSectionHeader ("Commander"), reassigning Section for every entry that follows it. I
      reproduced this directly with a standalone Go test against the real adapter and real parser:
      a following "Sol Ring" row lands in SectionCommander, with AssertAccounting reporting zero
      BlockingErrors and zero Warnings (entries=2, warnings=0, blockingerrors=0, dropped=0,
      nonblank=2 — the invariant is satisfied only because both sides of the equality are computed
      *after* the corruption, so it cannot see the corruption at all). This is a genuine,
      exploitable violation of "trustworthy preview": one attacker-or-data-quality-controlled field
      in a shared Archidekt deck link can hijack commander selection with zero surfaced signal
      anywhere in the preview. This is exactly the "silent mis-parse, not a cosmetic metadata loss"
      failure class the same file's own comment on archidektCollectorNumberRoundTrips (the 01-09
      Rule-1 bug fix for a sibling field) explicitly names as unacceptable — the fix was applied to
      collectorNumber only, not to name or setCode.
    artifacts:
      - path: "server/deck_providers.go"
        issue: "normalizeToDeckText (~line 499) trims but does not reject/sanitize control characters ('\\n'/'\\r') in row.Card.OracleCard.Name before handing it to formatArchidektDeckLine; formatArchidektDeckLine (~line 590) interpolates setCode directly into the generated line with no round-trip validation, unlike the collectorNumber parameter on the same line, which archidektCollectorNumberRoundTrips already guards"
    missing:
      - "Reject or sanitize a card name containing '\\n'/'\\r' before formatArchidektDeckLine (mirroring the existing empty-name rejection via errArchidektMissingCardName) — this is CR-01 in 01-REVIEW.md"
      - "Extend archidektCollectorNumberRoundTrips-style validation to setCode, omitting the printing-metadata suffix entirely when either setCode or collectorNumber would not round-trip — this is CR-02 in 01-REVIEW.md, and reproduced independently: a setCode of \"SET CODE\" with a round-tripping collectorNumber of \"123\" generates \"1 Sol Ring (SET CODE) 123\", which parses back to an entry named \"Sol Ring (SET CODE)\" — the real card silently fails resolution under a corrupted name"
      - "A regression test (fixture-independent, per WR-03 in 01-REVIEW.md) asserting a synthetic row with an embedded newline or unsafe setCode does not corrupt section membership or card resolution — none of the current TestArchidekt_* tests exercise adversarial/malformed field *content* within an otherwise well-formed row, only structural malformance"
deferred: []
human_verification: []
---

# Phase 1: Measured Deck Import Foundation Verification Report

**Phase Goal:** Any decklist a Commander player already has — pasted in a familiar format, or pulled from an allowlisted public deck URL if one proves viable — resolves to a trustworthy preview through a single server-side path, and every step of the activation funnel has a privacy-safe event record behind it.
**Verified:** 2026-08-07T18:15:00Z
**Status:** gaps_found
**Re-verification:** Yes — after gap closure (G-01-1: D-14 provider reversal from Moxfield to Archidekt, plans 01-08/01-09/01-10)

## Verification Method

Evidence-first, not SUMMARY-trust. I read the actual `archidektAdapter` implementation and its
tests, ran every relevant test suite myself, and — because a code review (`01-REVIEW.md`,
committed just before this verification) reported two Critical, unfixed findings in the delta this
gap-closure work produced — I independently reproduced both findings against the real code with a
standalone Go test (not the review's narrative alone), rather than accepting either the review's or
the SUMMARYs' account at face value:

- `go test ./server -run 'TestArchidekt_|TestDeckImport_ArchidektURLPath|TestProvider_' -v` — all
  green (17 test functions / subtests).
- `go test ./server -count=1` (full package suite, single run) — green, zero failures.
- `go test ./server -run 'TestProductEvents_|TestMetrics' -v` and `go test ./pkg/deckimport -race`
  — green, confirming criteria 3/4 (untouched by this delta, per `git diff d3a70e8..HEAD --stat --
  pkg/telemetry server/product_events.go`, zero changes) still hold.
- `go build ./...`, `go vet ./server ./pkg/...` — clean.
- Wrote and ran a standalone reproduction (`server/zzverifier_cr_repro_test.go`, deleted after
  verification, not committed) directly against `archidektAdapter.normalizeToDeckText` and
  `deckimport.ParseWithSource`/`AssertAccounting`: confirmed CR-01 (embedded-newline commander-
  section hijack) and CR-02 (unsafe-setCode name corruption) both reproduce exactly as
  `01-REVIEW.md` describes, on the code currently at HEAD.
- Read `server/deck_providers.go`, `server/deck_providers_test.go`, `server/deck_import.go`,
  `pkg/deckimport/scanner.go` (`splitLines`, `AssertAccounting`), `pkg/deckimport/sections.go`
  (`matchSectionHeader`) directly to confirm the mechanism the review describes is real, not
  theoretical.
- Read `docs/research/deck-provider-feasibility.md` section 6, `COVERAGE.md`, `.planning/WINDOWS.md`,
  `.planning/REQUIREMENTS.md` in full.
- Confirmed no debt markers (`TBD`/`FIXME`/`XXX`/`TODO`/`HACK`/`PLACEHOLDER`) in any file this
  phase's gap-closure plans touched.

## Goal Achievement

### Observable Truths (ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | A pasted decklist in any of the six supported syntaxes returns a preview with total cards, commander candidates, unresolved entries, warnings, and blocking errors; no nonblank row disappears | ✓ VERIFIED | Unchanged since the prior verification (`pkg/deckimport` untouched by this delta). `AssertAccounting` mechanically enforces the invariant on paste input; `go test ./pkg/deckimport -race` green including golden fixtures. |
| 2 | Comma-containing names survive intact and resolve identically across both syntaxes; unresolved cards surface as unresolved, never silently substituted; preview and final library creation consume the same normalized result | ✗ FAILED (provider path only) | Holds fully for the paste path (unchanged). For the newly-shipped Archidekt provider path, **independently reproduced**: an untrusted card name containing an embedded newline silently reassigns a *different*, unrelated card's Section to Commander with zero warning or blocking error — not "unresolved," not "disappeared," but silently misclassified, which is the same "silent substitution" failure class this criterion exists to prevent, just applied to a card's role rather than its name. A malformed `setCode` separately corrupts a resolvable card's name so it silently becomes unresolved for a reason the preview never surfaces. See Gaps below. |
| 3 | A recorded activation event groups by day, role, source, outcome in PostgreSQL; unknown event names/keys, oversized payloads, and prohibited content are rejected; the server attaches authenticated user IDs rather than trusting the client | ✓ VERIFIED | Unchanged since prior verification (zero diff in `pkg/telemetry`, `server/product_events.go` since `d3a70e8`). Re-ran `TestProductEvents_Allowlist` (6 subtests) directly — pass. |
| 4 | Prometheus reports guest-session, import, create/join, and board-activation counters/histograms with no username/session/user/game-ID label | ✓ VERIFIED (per ROADMAP's own criterion-4 scoping: declared-here, closed-later) | Unchanged since prior verification. `TestMetrics_AllCriterionFourFamiliesExist` re-run, pass. |
| 5 | A written decision record names the first provider (with SSRF/kill-switch controls failing closed) or documents a no-go; either satisfies the phase | ✓ VERIFIED | `docs/research/deck-provider-feasibility.md` section 6 (read in full) records the D-14 reversal to Archidekt, dated and attributed to the user at UAT, with section 5's original Moxfield decision left intact for the record. SSRF controls (`safeControl`, `newDeckProviderCheckRedirect`, body cap, timeouts) confirmed byte-for-byte unchanged since the 2026-08-05 review (`git diff d3a70e8..HEAD`) and still pass: `TestSafeControl_DeniedAddresses`, `TestSafeClient_HostAndRedirect`, `TestSafeClient_Timeouts`, `TestSafeClient_BodyCap`. Kill switch (`providerEnabled`) still fails closed and defaults off (`TestProvider_KillSwitch`, `TestProvider_KillSwitchRequiresAllowlist`, re-run, pass). This criterion is satisfied on its own terms — the SSRF/kill-switch controls it names are solid — even though the adapter's *content*-sanitization gap (criterion 2's failure) is a separate, real problem. |

**Score:** 4/5 truths fully verified; 1 failed (Success Criterion 2, provider-path adapter content sanitization).

### Reasoning on severity (why this is a gap, not a deferred follow-up)

The phase goal's own words are "resolves to a **trustworthy** preview." `01-REVIEW.md`'s two
Critical findings (CR-01, CR-02) are not style nits — they are reproducible, silent data-corruption
paths in the exact adapter this phase's three gap-closure plans (01-08/01-09/01-10) exist to
deliver, and I confirmed both by direct reproduction rather than trusting the review's prose. CR-01
in particular defeats the accounting invariant (`AssertAccounting`) that the rest of this phase's
architecture — and its own test suite — relies on to make mis-parses loud: the invariant's own two
operands (`NonBlankLines` and `Entries+Warnings+BlockingErrors+DroppedSectionRows`) are both
computed *from the corrupted text*, so a corruption introduced before that text is generated is
structurally invisible to it. That is a category of bug the codebase's own commentary (on
`archidektCollectorNumberRoundTrips`, the 01-09 fix for the *sibling* `collectorNumber` field)
explicitly identifies as unacceptable — "That is exactly the silent mis-parse this plan exists to
catch, not a cosmetic metadata loss" — just left open on the two adjacent fields (`name`, `setCode`)
carrying the identical risk on the identical code path.

The kill switch defaults off and requires a human to read Archidekt's terms before real deployment,
so this is not exploitable *today* in production. But the phase's deliverable is the adapter code
itself, intended to be enabled once that precondition clears — and shipping REQ-ACT-003 as
"Complete" on the strength of a normalizer with a demonstrated commander-hijack path, discovered by
this same verification pass reproducing the review's own findings, would be certifying "trustworthy"
against direct evidence to the contrary. This is why it is scored as a gap against the ROADMAP's
Success Criterion 2 rather than waved through as a deferred improvement: fixing it is a small,
well-scoped, already-drafted change (01-REVIEW.md gives working patches for both CR-01 and CR-02),
consistent in shape with the exact fix 01-09 already made for `collectorNumber`.

REQ-ACT-003's own literal ticket text (feasibility comparison, decision record, hostname-keyed
adapter, SSRF controls, normalize-through-ACT-002, feature flag/kill switch) does **not** mention
content sanitization, so REQ-ACT-003 remains correctly `Complete` in `.planning/REQUIREMENTS.md` —
that traceability claim is accurate on its own terms and 01-10's clause-by-clause audit is sound.
The gap identified here is at the ROADMAP goal/success-criterion level (Step 2a's authoritative
contract), which is a broader bar than any single SPEC ticket's text, exactly as this phase's own
prior verification already established `criterion 5` can be satisfied even when a stricter Layer-2
ticket bar is not. I am applying the same "roadmap contract stands even when narrower plan-level
success looks complete" principle here, in the other direction.

### Requirements Traceability (Phase 1: REQ-ACT-001, REQ-ACT-002, REQ-ACT-003)

| Requirement | REQUIREMENTS.md status | Verified against code | Assessment |
|---|---|---|---|
| REQ-ACT-001 | Complete | Unchanged since prior verification; zero diff in the owning files since `d3a70e8` | ✓ SATISFIED — no new evidence needed; nothing in this delta touches it. |
| REQ-ACT-002 | Complete | `TestDeckImport_ArchidektURLPath/EnabledProducesSameSingleParseAsPaste` (re-run, pass) proves the provider path still reuses the single canonical parse (`CardCount == 100` for the real fixture) | ✓ SATISFIED — the single-canonical-parse contract holds; note this assessment is about *which parser runs*, not about the fidelity of the text handed into it (see Success Criterion 2 gap above, which is a separate concern from ACT-002's text). |
| REQ-ACT-003 | Complete | `archidektAdapter` implemented against the real observed contract (`server/testdata/deck_providers/archidekt_deck_2026-08-05.json`), pinned by six `TestArchidekt_*` fixture-contract tests plus `TestDeckImport_ArchidektURLPath` (all re-run, pass); Moxfield scaffold fully removed (`grep -c moxfield server/deck_providers.go` → non-test-comment hits are zero; only historical references remain in test names/comments describing what was reversed); `.planning/WINDOWS.md` stub window closed; SSRF controls confirmed unchanged and passing | ✓ SATISFIED at the ticket-text level — every clause of REQ-ACT-003's own literal text (comparison, decision record, hostname-keyed adapter, SSRF controls, normalize-through-ACT-002, flag/kill switch) has direct, re-run, passing test evidence. **Not** a certification that the adapter is defect-free — see the Success Criterion 2 gap above, which sits above the ticket-text bar. |

No orphaned requirements: `.planning/REQUIREMENTS.md` maps exactly REQ-ACT-001/002/003 to Phase 1, matching this phase's own plans' `requirements` fields.

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `server/deck_providers.go` (`archidektAdapter`) | Real-contract Archidekt normalizer, hostname-keyed registry, SSRF-safe client, kill switch | ⚠️ VERIFIED WITH DEFECT | Exists, substantive, wired, and its happy-path field mapping is genuinely pinned by fixture tests (D1-D7, 01-09). SSRF/kill-switch machinery unchanged and solid. But `normalizeToDeckText`/`formatArchidektDeckLine` lack input sanitization for `name` (embedded newline/CR) and `setCode` (unsafe characters) — reproduced directly, see Gaps. |
| `server/deck_providers_test.go` | Fixture-contract tests pinning the field mapping | ⚠️ VERIFIED WITH GAP | Six real, disk-loaded, fixture-based `TestArchidekt_*` tests exist and pass, but per WR-03 (01-REVIEW.md, confirmed by reading the file) none exercises adversarial/malformed field *content* (newline, quote, bracket, embedded whitespace) — only structural malformance (missing fields). None of these tests would have caught or would catch a regression of CR-01/CR-02. |
| `docs/research/deck-provider-feasibility.md` | Decision record naming Archidekt, reversal dated and attributed | ✓ VERIFIED | Section 6 read in full; genuinely supersedes (not deletes) section 5's original Moxfield record; human ToS-read precondition explicitly tied to `DECK_PROVIDER_ALLOWED_HOSTS`/`DECK_PROVIDER_ENABLED`. |
| `.planning/phases/.../COVERAGE.md` | Honest external-API coverage reflecting the actual shipped adapter | ✓ VERIFIED | Rewritten for Archidekt; `read_public_deck`/`read_commander_designation` moved to INTEGRATE, each cited to a specific passing test; remaining 4 stay OPT-OUT with a current, non-stale reason. |
| `.planning/WINDOWS.md` | moxfieldAdapter stub window closed | ✓ VERIFIED | `status: fixed`, closed 2026-08-07T22:22:33Z; zero open windows remain (`grep -n "| open "` returns nothing). |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `server/deck_import.go` (`previewDeckURL`) | `server/deck_providers.go` (`deckProviderAdapterFor` → `archidektAdapter`) | hostname-keyed adapter routing, allowlist-gated | ✓ WIRED | `TestProvider_ArchidektAdapterRoutesToRealNormalizer`, `TestProvider_UnregisteredHostNeverFetched`, re-run, pass. |
| `archidektAdapter.normalizeToDeckText` | `pkg/deckimport.Parse`/`ParseWithSource` | generated decklist text, single canonical parse | ✓ WIRED (but see content-fidelity gap above) | `TestDeckImport_ArchidektURLPath/EnabledProducesSameSingleParseAsPaste` proves the routing and single-parse-site claim on the happy path; my own reproduction proves the *content* crossing this link is not always faithfully preserved. |
| `server/deck_providers.go` (kill switch/allowlist) | `previewDeckURL` | independent gates, both must pass before any dial | ✓ WIRED | `TestDeckImport_ArchidektURLPath/DisabledNeverDials`, `/NonAllowlistedHostNeverDialsEvenWithFlagOn`, re-run, pass. |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Archidekt fixture-contract suite passes | `go test ./server -run 'TestArchidekt_\|TestDeckImport_ArchidektURLPath\|TestProvider_' -v` | 17 test funcs/subtests, all PASS | ✓ PASS |
| Full server suite has no regressions | `go test ./server -count=1` | PASS, zero failures, 30.66s | ✓ PASS |
| Criteria 3/4 telemetry unaffected by this delta | `go test ./server -run 'TestProductEvents_\|TestMetrics' -v` | PASS | ✓ PASS |
| `go vet` clean | `go vet ./server ./pkg/...` | clean | ✓ PASS |
| CR-01 reproduction: embedded newline hijacks commander section | standalone Go test against `archidektAdapter{}.normalizeToDeckText` + `deckimport.ParseWithSource`/`AssertAccounting` | "Sol Ring" (a following, unrelated legitimate row) lands in `SectionCommander`; `BlockingErrors=[]`, `Warnings=[]`, accounting invariant satisfied (2 entries = 2 nonblank lines) despite the corruption | ✗ FAIL (confirms gap) |
| CR-02 reproduction: unsafe setCode corrupts card name | standalone Go test, same harness | Generated text `"1 Sol Ring (SET CODE) 123\n"` parses to an entry named `"Sol Ring (SET CODE)"`, not `"Sol Ring"` | ✗ FAIL (confirms gap) |

### Anti-Patterns Found

No debt markers (`TBD`/`FIXME`/`XXX`) or unresolved `TODO`/`HACK`/`PLACEHOLDER` comments in any file this phase's gap-closure plans (01-08/01-09/01-10) touched.

| File | Line | Pattern | Severity | Impact |
|---|---|---|---|---|
| `server/deck_providers.go` | ~499 (`normalizeToDeckText`) | Untrusted `name` passed to `formatArchidektDeckLine` with no control-character sanitization | 🛑 Blocker | CR-01 — reproduced; silently reassigns commander-section membership. |
| `server/deck_providers.go` | ~590 (`formatArchidektDeckLine`) | Untrusted `setCode` interpolated with no round-trip validation, unlike the adjacent `collectorNumber` parameter on the same line | ⚠️ Warning | CR-02 — reproduced; silently corrupts a resolvable card's name into "unresolved." Same fix location as CR-01; recommend fixing together. |
| `server/deck_providers.go` | ~499 | No defense against other grammar-special characters (`"`, `` ` ``, `[`, `]`, `#`) in `name` | ⚠️ Warning | WR-01 (01-REVIEW.md) — silently corrupts parsed names via quoted-name/trailing-annotation rules; does not defeat the accounting invariant the way CR-01 does. |
| `server/deck_providers.go` | 488-525 | Category membership checks use exact-case string equality, inconsistent with the `EqualFold` used for the "Commander" keyword | ⚠️ Warning | WR-02 (01-REVIEW.md) — could silently include an excluded (maybeboard) card if Archidekt ever emits inconsistent casing. |
| `server/deck_providers_test.go` | all `TestArchidekt_*` | No test exercises adversarial/malformed field *content*, only structural malformance | ⚠️ Warning | WR-03 (01-REVIEW.md) — confirms none of the current tests would catch a CR-01/CR-02 regression. |
| `server/deck_providers.go` | 497-502 | A single malformed row fails the whole deck import (no partial import) | ℹ️ Info | IN-01 (01-REVIEW.md) — deliberate, documented design choice; not a defect. |

## Gaps Summary

One roadmap-level gap: Success Criterion 2 ("unresolved cards surface as unresolved... never
silently substituted") is violated for the newly-shipped Archidekt provider path by two related,
independently-reproduced defects in `archidektAdapter` (CR-01, critical; CR-02, related and lower
severity) that a code review surfaced and this verification confirmed by direct reproduction, not
by trusting either the review's or the SUMMARYs' narrative. Both defects share a root cause
(untrusted upstream JSON fields interpolated into generated decklist text without the same
round-trip/content validation the adapter's own `collectorNumber` field already received in 01-09)
and a fix location (`formatArchidektDeckLine` / the `cards[]` loop in `normalizeToDeckText`), and
`01-REVIEW.md` already provides drafted patches for both. This does not block REQ-ACT-003's own
ticket-text traceability (correctly `Complete`), nor Success Criteria 1, 3, 4, or 5, all of which
remain independently verified. Recommend a small, targeted gap-closure plan before Phase 2 begins,
scoped to: (1) reject/sanitize control characters in `name`; (2) extend round-trip validation to
`setCode`; (3) add the fixture-independent adversarial-content regression tests WR-03 identifies as
currently missing.

---

_Verified: 2026-08-07T18:15:00Z_
_Verifier: Claude (gsd-verifier)_
