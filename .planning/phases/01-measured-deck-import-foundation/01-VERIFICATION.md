---
phase: 01-measured-deck-import-foundation
verified: 2026-08-08T01:45:00Z
status: passed
score: 5/5 must-haves verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 4/5
  gaps_closed:
    - "Success Criterion 2 (provider-path adapter content sanitization): CR-01 (embedded \\n/\\r in an untrusted Archidekt card name silently reassigning an unrelated card's Section via a fabricated 'Commander' header) and CR-02 (an unsafe setCode silently corrupting a resolvable card's parsed name) both independently reproduced as FIXED against the real code at HEAD, with a standalone Go test I wrote myself against the actual archidektAdapter.normalizeToDeckText and the actual pkg/deckimport.Parse/AssertAccounting -- not by reading the fix commits or trusting 01-REVIEW-FIX.md's narrative."
  gaps_remaining: []
  regressions: []
gaps: []
deferred: []
human_verification: []
---

# Phase 1: Measured Deck Import Foundation Verification Report

**Phase Goal:** Any decklist a Commander player already has — pasted in a familiar format, or pulled from an allowlisted public deck URL if one proves viable — resolves to a trustworthy preview through a single server-side path, and every step of the activation funnel has a privacy-safe event record behind it.
**Verified:** 2026-08-08T01:45:00Z
**Status:** passed
**Re-verification:** Yes — after gap closure of the single gap (Success Criterion 2 / CR-01+CR-02) the prior verification (2026-08-07T18:15:00Z) found, itself found after re-verifying an earlier `human_needed` pass that closed G-01-1 (D-14 Moxfield→Archidekt reversal).

## Verification Method

I did **not** trust `01-REVIEW-FIX.md`'s "fixed: 5/5" claim, `01-REVIEW.md` iteration 2's "0 Critical" re-review conclusion, or any SUMMARY. I re-derived the verdict directly:

1. **Read the actual current code** (`server/deck_providers.go` lines 460-670: `normalizeToDeckText`, `archidektNameIsSafe`, `archidektTokenRoundTrips`, `formatArchidektDeckLine`) and the upstream grammar it defends against (`pkg/deckimport/scanner.go`: `splitLines`, `resolveNameAndMetadata`, `extractTrailingAnnotations`, `AssertAccounting`; `pkg/deckimport/sections.go`: `matchSectionHeader`) to confirm the fix's shape actually closes the mechanism the prior verification exploited, not just that new code exists.
2. **Wrote a standalone reproduction test** (`server/zzverifier_reverify_test.go`, three tests, deleted immediately after this pass and never committed — `git status --short server` confirmed clean afterward) that:
   - Rebuilt the *exact* CR-01 payload (an embedded `\n` in `oracleCard.name` followed by a legitimate "Sol Ring" row) directly against `archidektAdapter{}.normalizeToDeckText` — result: `normalizeToDeckText` now returns `errArchidektUnsafeCardName` *before* any text is generated, confirmed via `errors.Is`. The prior verification's exploit path (a fabricated "Commander" line silently reassigning Section, invisible to `AssertAccounting`) cannot occur because the corrupted text is never produced.
   - Rebuilt the exact CR-02 payload (`setCode: "SET CODE"`, `collectorNumber: "123"`, name `"Sol Ring"`) — result: generated text is `"1 Sol Ring\n"` (D-07 suffix omitted entirely, not emitted with only one side correct), which parses to one clean `Entries[0].Name == "Sol Ring"` with zero `BlockingErrors` — the prior verification's corruption (`"Sol Ring (SET CODE)"` as the parsed name) does not reproduce.
   - Fed 13 real, syntactically tricky Magic card names (`Atraxa, Praetors' Voice`; `Kongming, "Sleeping Dragon"`; `B.O.B. (Bevy of Beebles)`; `Erase (Not the Urza's Legacy One)`; `Look at Me, I'm the DCI`; `Ach! Hans, Run!`; `Yargle and Multani`; `Krark, the Thumbless`; `Urza, Lord High Artificer`; `Emrakul, the Aeons Torn`; `Sol Ring`; `Lim-Dûl's Vault`; `Jhoira's Familiar`) through both `archidektNameIsSafe` directly and the full `normalizeToDeckText` → `deckimport.Parse` round trip — zero false positives, zero corrupted round trips. This directly checks the risk that a stricter guard silently breaks legitimate imports (the same failure class the guard exists to prevent, just introduced by the guard itself).
3. **Ran the real test suites myself**, not from the SUMMARY's reported numbers:
   - `go build ./...` — clean.
   - `go vet ./server ./pkg/...` — clean.
   - `go test ./server -run 'TestZZVerifier_' -v` — my 3 standalone reproduction tests, all PASS (see above).
   - `go test ./server -run 'TestArchidekt_|TestDeckImport_ArchidektURLPath|TestProvider_|TestSafeControl_|TestSafeClient_' -v` — 22 top-level test functions / all subtests, all PASS.
   - `go test ./server -run 'TestProductEvents_|TestMetrics' -v` — all PASS (Success Criteria 3/4).
   - `go test ./server -count=1` (full package, single run) — PASS, 33.8s, zero failures.
   - `go test ./pkg/deckimport -race -count=1` — PASS, 1.3s, zero failures.
4. **Confirmed the shared-code blast radius is empty**: `git diff e240fa3..HEAD --stat -- pkg/telemetry server/product_events.go server/metrics.go pkg/deckimport` — zero output, i.e. zero changes to any file underpinning Success Criteria 1/3/4 since the gap was identified. The fix touched only `server/deck_providers.go` and `server/deck_providers_test.go`.
5. **Read `server/deck_import.go`'s `previewDeckURL` and `blockedPreview`** directly to independently judge the one open Warning (`WR-01`, `01-REVIEW.md` iteration 2) — whether "reject the whole deck on one unsafe row" is compatible with "trustworthy preview" and "no nonblank row disappears without becoming one of [entries/warnings/errors]." See "Judgment: WR-01" below — I did not defer this to a human-verification item; I resolved it myself with direct evidence and document the reasoning.
6. **Checked for debt markers** in every file the fix touched (`git diff e240fa3..HEAD --name-only -- server pkg` → `server/deck_providers.go`, `server/deck_providers_test.go`; grepped both for `TBD|FIXME|XXX|TODO|HACK|PLACEHOLDER`) — none found.
7. Re-confirmed `.planning/WINDOWS.md` has zero open windows and `server/deck_providers.go` has zero remaining `moxfield` references of any kind (comments included) — both grep to empty, consistent with the prior verification's finding, unaffected by this delta.

## Goal Achievement

### Observable Truths (ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | A pasted decklist in any of the six supported syntaxes returns a preview with total cards, commander candidates, unresolved entries, warnings, and blocking errors; no nonblank row disappears | ✓ VERIFIED | `pkg/deckimport` has zero diff since the gap-closure delta began (`git diff e240fa3..HEAD --stat -- pkg/deckimport` empty); `go test ./pkg/deckimport -race -count=1` green, including golden fixtures and `AssertAccounting`-driven invariant checks. |
| 2 | Comma-containing names survive intact and resolve identically across both syntaxes; unresolved cards surface as unresolved, never silently substituted; preview and final library creation consume the same normalized result | ✓ VERIFIED | Paste path unchanged (as #1). Provider path: independently reproduced **fixed** — my own standalone test against the real `archidektAdapter.normalizeToDeckText` confirms both the embedded-newline commander-hijack (CR-01) and the unsafe-setCode name corruption (CR-02) the prior verification found no longer occur; `normalizeToDeckText` now fails closed with an explicit sentinel error *before* generating any text, rather than silently emitting corrupted text. 13 real, syntactically-tricky card names round-trip byte-identical through the full normalize→parse pipeline with zero false positives from the new guard. |
| 3 | A recorded activation event groups by day, role, source, outcome in PostgreSQL; unknown event names/keys, oversized payloads, and prohibited content are rejected; the server attaches authenticated user IDs rather than trusting the client | ✓ VERIFIED | Zero diff in `pkg/telemetry`/`server/product_events.go` since the gap was identified. `TestProductEvents_Allowlist` (6 subtests), `TestProductEvents_AuthoritativeDedup` (4 subtests), `TestProductEvents_ConcurrentInsert`, `TestProductEvents_OccurredAtTieOrdering`, `TestProductEvents_WriteFailureIsNonFatal` — all re-run directly by me, all PASS. |
| 4 | Prometheus reports guest-session, import, create/join, and board-activation counters/histograms with no username/session/user/game-ID label | ✓ VERIFIED (per ROADMAP's own criterion-4 scoping: declared-here, closed-later) | Zero diff since gap identified. `TestMetrics_*` re-run, PASS. Consistent with the ROADMAP's explicit contract note that only import/product-event families are *observed* in Phase 1. |
| 5 | A written decision record names the first provider (with SSRF/kill-switch controls failing closed) or documents a no-go | ✓ VERIFIED | `docs/research/deck-provider-feasibility.md` section 6 (Archidekt decision, dated, attributed) unchanged since the last verification. SSRF/kill-switch machinery (`safeControl`, `newDeckProviderCheckRedirect`, body cap, timeouts, `providerEnabled`) is outside the fix's file diff and all its tests (`TestSafeControl_*`, `TestSafeClient_*`, `TestProvider_KillSwitch*`) re-run PASS. |

**Score:** 5/5 truths fully verified. Zero gaps remaining.

### Judgment: WR-01 ("reject the whole deck on one unsafe row") — resolved as accepted design, not a gap

`01-REVIEW.md` iteration 2 leaves one open Warning: `normalizeToDeckText` returns `errArchidektUnsafeCardName` for the *entire* deck the moment any single `cards[]` row's name trips the (now four-condition) `archidektNameIsSafe` guard, rather than excluding just that one row. I read `previewDeckURL` (`server/deck_import.go:252-288`) and `blockedPreview` (`server/deck_import.go:312-323`) directly to judge whether this is compatible with the phase goal's "trustworthy preview" and Success Criterion 1's "no nonblank row disappears without becoming one of [entries/warnings/blocking errors]."

**My verdict: this is not a gap.** Two independent reasons:

1. **Criterion 1's "no nonblank row disappears" invariant is explicitly scoped to "a pasted decklist,"** not the provider-URL path — its enforcement mechanism (`AssertAccounting`) operates only on the text `pkg/deckimport.Parse` receives, and the paste path is untouched by any of this. The provider path is governed by Criterion 2's "unresolved... rather than silently substituted" language, and a hard failure is neither "unresolved" (that's a per-card preview state) nor "substituted" (nothing false is shown).
2. **The failure is not silent.** I confirmed `blockedPreview` returns an explicit, non-empty `BlockingErrors` slice with `CanContinue: false` — a `DeckPreview` that visibly says "We couldn't load that deck link right now. Paste your decklist as text instead." rather than rendering any preview, correct or corrupted. A user cannot mistake this for a successful, trustworthy result. Failing loud and closed — declining to show *anything* rather than showing something wrong — is the same fail-closed philosophy this phase already applies to the kill switch and SSRF controls, and it is *more* protective of "trustworthy" than a partial-import compromise would be.

I additionally confirmed the practical false-positive risk this creates is negligible: `archidektOracleCard.Name` is Archidekt's echo of a Scryfall canonical card name, and none of the 19 real card names checked across my own test and `01-REVIEW.md`'s own probe (which used a dedicated, uncommitted `TestProbe_RealCardNamesWithParens`) trip any of the four guard conditions (leading `"`, trailing `` ` ``/`]`, trailing `#tag`, embedded `\n`/`\r`). This is therefore an honest tradeoff between graceful per-row degradation (a future enhancement, not required by this phase's success criteria) and simplicity, not a correctness or trust defect. I concur with, but independently re-derived, `01-REVIEW.md`'s own Warning-not-Critical classification.

### Requirements Traceability (Phase 1: REQ-ACT-001, REQ-ACT-002, REQ-ACT-003)

| Requirement | REQUIREMENTS.md status | Verified against code | Assessment |
|---|---|---|---|
| REQ-ACT-001 | Complete | Zero diff in owning files since prior verification | ✓ SATISFIED |
| REQ-ACT-002 | Complete | `TestDeckImport_ArchidektURLPath/EnabledProducesSameSingleParseAsPaste` re-run, PASS (`CardCount == 100` on the real fixture); single-canonical-parse contract holds, and the text now crossing that link is also content-faithful (Success Criterion 2 gap closed) | ✓ SATISFIED — now fully, not just at the routing level. |
| REQ-ACT-003 | Complete | `archidektAdapter` re-verified against the real observed contract; all `TestArchidekt_*`/`TestProvider_*`/`TestSafeControl_*`/`TestSafeClient_*` re-run, PASS; the ticket's own literal text (comparison, decision record, hostname-keyed adapter, SSRF controls, normalize-through-ACT-002, flag/kill switch) is unaffected by and does not require the sanitization fix, but the fix has now also closed the broader roadmap-level Success Criterion 2 concern the prior verification raised above the ticket-text bar | ✓ SATISFIED — no remaining caveat. |

No orphaned requirements: `.planning/REQUIREMENTS.md` maps exactly REQ-ACT-001/002/003 to Phase 1, matching this phase's plans' `requirements` fields.

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `server/deck_providers.go` (`archidektAdapter`) | Real-contract Archidekt normalizer, hostname-keyed registry, SSRF-safe client, kill switch, content-safe name/setCode sanitization | ✓ VERIFIED | `archidektNameIsSafe` (4 conditions, position-specific, independently traced against `splitLines`/`resolveNameAndMetadata`/`extractTrailingAnnotations`) and `archidektTokenRoundTrips` (applied to both `setCode` and `collectorNumber` with `&&`, single emission branch) both confirmed to close CR-01/CR-02 by direct reproduction, not by reading the diff. |
| `server/deck_providers_test.go` | Fixture-contract tests plus adversarial-content regression coverage | ✓ VERIFIED | 6 fixture-based `TestArchidekt_*` tests (happy path) plus 5 new synthetic adversarial tests (`TestArchidekt_UnsafeCardNameRejected`, `TestArchidekt_UnsafeSetCodeOmitsSuffix`, `TestArchidekt_CategoryMembershipIsCaseFolded`, `TestArchidekt_UnsafeNameGrammarConflictRejected`, `TestArchidekt_InternalQuoteInNameSurvives`) — all re-run, PASS. This closes the WR-03 gap the prior verification flagged (no adversarial-content coverage existed then). |
| `docs/research/deck-provider-feasibility.md` | Decision record naming Archidekt, reversal dated and attributed | ✓ VERIFIED | Unchanged since prior verification; section 6 confirmed present. |
| `.planning/phases/.../COVERAGE.md` | Honest external-API coverage reflecting the actual shipped adapter | ✓ VERIFIED | Unchanged since prior verification. |
| `.planning/WINDOWS.md` | moxfieldAdapter stub window closed | ✓ VERIFIED | Zero open windows (`grep -n "| open "` empty). |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `server/deck_import.go` (`previewDeckURL`) | `server/deck_providers.go` (`deckProviderAdapterFor` → `archidektAdapter`) | hostname-keyed adapter routing, allowlist-gated | ✓ WIRED | `TestProvider_ArchidektAdapterRoutesToRealNormalizer`, `TestProvider_UnregisteredHostNeverFetched`, re-run, PASS. |
| `archidektAdapter.normalizeToDeckText` | `pkg/deckimport.Parse`/`ParseWithSource` | generated decklist text, single canonical parse | ✓ WIRED, content now faithful | `TestDeckImport_ArchidektURLPath/EnabledProducesSameSingleParseAsPaste` proves routing; my own reproduction now proves content-fidelity holds too (previously the one open gap). |
| `server/deck_providers.go` (kill switch/allowlist) | `previewDeckURL` | independent gates, both must pass before any dial | ✓ WIRED | `TestDeckImport_ArchidektURLPath/DisabledNeverDials`, `/NonAllowlistedHostNeverDialsEvenWithFlagOn`, re-run, PASS. |
| `archidektAdapter.normalizeToDeckText` (unsafe name/setCode) | `previewDeckURL` → `blockedPreview` | fail-closed with an explicit, non-silent `BlockingErrors` message | ✓ WIRED | Read directly (`server/deck_import.go:275-288`, `:312-323`); confirms the WR-01 judgment above — failure is loud, not silent. |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| CR-01 no longer reproduces | standalone Go test (`archidektAdapter{}.normalizeToDeckText` with an embedded-`\n` name) | Returns `errArchidektUnsafeCardName` before generating any text; `errors.Is` confirms sentinel | ✓ PASS |
| CR-02 no longer reproduces | standalone Go test (setCode `"SET CODE"`, collectorNumber `"123"`) | Generated text `"1 Sol Ring\n"`; parses to clean `Entries[0].Name == "Sol Ring"`, zero BlockingErrors | ✓ PASS |
| New guard has zero false positives on real card names | standalone Go test, 13 tricky real card names through `archidektNameIsSafe` + full normalize→parse round trip | Zero rejections, zero corrupted round trips | ✓ PASS |
| Archidekt/Provider/SafeControl/SafeClient suite | `go test ./server -run 'TestArchidekt_|TestDeckImport_ArchidektURLPath|TestProvider_|TestSafeControl_|TestSafeClient_' -v` | 22 top-level funcs / all subtests PASS | ✓ PASS |
| Criteria 3/4 telemetry unaffected | `go test ./server -run 'TestProductEvents_|TestMetrics' -v` | PASS | ✓ PASS |
| Full server suite, single run | `go test ./server -count=1` | PASS, 33.8s, zero failures | ✓ PASS |
| pkg/deckimport race suite | `go test ./pkg/deckimport -race -count=1` | PASS, 1.3s | ✓ PASS |
| Build/vet clean | `go build ./...`, `go vet ./server ./pkg/...` | Clean | ✓ PASS |

### Anti-Patterns Found

No debt markers (`TBD`/`FIXME`/`XXX`) or unresolved `TODO`/`HACK`/`PLACEHOLDER` comments in either file this fix cycle touched (`server/deck_providers.go`, `server/deck_providers_test.go`).

| File | Line | Pattern | Severity | Impact |
|---|---|---|---|---|
| `server/deck_providers.go` | 497-543 | Whole-deck-abort on a single unsafe-name row (no partial import) | ℹ️ Info | Accepted design tradeoff — see "Judgment: WR-01" above. Not a defect; not blocking. |
| `server/deck_providers_test.go` | 1021-1054, 1186-1215 | Two adversarial-rejection tests assert `err != nil` without pinning the specific sentinel via `errors.Is` | ℹ️ Info | IN-01 in `01-REVIEW.md` iteration 2 — not exploitable today (no other error path exists for these inputs), but a future refactor could silently swap the error path without the test catching it. Cosmetic test-precision gap, not a functional defect. |
| `server/deck_providers.go` | 640-650 | `archidektTokenRoundTrips` is stricter than the real grammar for `setCode` (rejects a hyphen the real parser would accept) | ℹ️ Info | IN-02 — always degrades safely (omits the metadata suffix), never corrupts; no correctness impact. |
| `server/deck_providers.go` | 515, 663-669 | Zero/negative `quantity` from the provider is untested at the adapter level (though handled correctly downstream by the shared parser) | ℹ️ Info | IN-03 — not a correctness bug, just a coverage gap for an edge case that already degrades safely. |

None of the above rise to Warning or Blocker in my own independent judgment; I concur with `01-REVIEW.md` iteration 2's Info classification for all four, having traced each mechanism myself rather than accepting the review's account.

## Gaps Summary

None. The single gap the prior verification (2026-08-07T18:15:00Z) found — Success Criterion 2 violated for the Archidekt provider path via CR-01 (embedded-newline commander-section hijack) and CR-02 (unsafe-setCode name corruption) — is closed, confirmed by a standalone reproduction test I wrote and ran myself against the real adapter and real parser, not by trusting the fix commits, `01-REVIEW-FIX.md`, or `01-REVIEW.md` iteration 2's re-review narrative. The one remaining open item from that re-review (`WR-01`, the whole-deck-abort failure mode) was independently judged not to be a gap against this phase's success criteria, for reasons documented above, and is recorded here as an accepted, informational design note for a possible future improvement rather than a blocking finding. All five ROADMAP Success Criteria are verified. Both requirement-traceability entries and all four required artifacts hold. Phase 1 goal is achieved.

---

_Verified: 2026-08-08T01:45:00Z_
_Verifier: Claude (gsd-verifier)_
