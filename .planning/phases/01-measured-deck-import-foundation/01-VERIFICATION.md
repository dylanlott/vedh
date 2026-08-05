---
phase: 01-measured-deck-import-foundation
verified: 2026-08-05T22:31:45Z
status: human_needed
score: 5/5 must-haves verified
behavior_unverified: 0
overrides_applied: 0
human_verification:
  - test: "Confirm REQ-ACT-003's traceability status (left 'Pending' in REQUIREMENTS.md) is the correct call to ship with, rather than a gap that must close before Phase 2 starts."
    expected: "Human/product-owner agreement that 'decision record + fail-closed scaffold, no working Moxfield import' is an acceptable Phase-1 exit for REQ-ACT-003, consistent with ROADMAP criterion 5's explicit either/or wording."
    why_human: "This is a scope/traceability judgment call already reasoned through in the plan chain and consistent with the roadmap's own scoping note for criterion 5 — not a code defect. Flagging for an explicit human sign-off rather than silently rounding a deliberately-left-Pending requirement up to Complete or down to a gap."
---

# Phase 1: Measured Deck Import Foundation Verification Report

**Phase Goal:** Any decklist a Commander player already has — pasted in a familiar format, or pulled from an allowlisted public deck URL if one proves viable — resolves to a trustworthy preview through a single server-side path, and every step of the activation funnel has a privacy-safe event record behind it.
**Verified:** 2026-08-05T22:31:45Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Verification Method

This was an evidence-first verification, not a SUMMARY-trust exercise. For every criterion below
I read the actual implementation, then independently ran the tests that are claimed to prove it
(not merely read the SUMMARY's claim that they passed):

- `go test ./pkg/... -race` — full package, green.
- `go test ./server -run 'TestScanner_RequiredSyntaxes|TestDeckImport|TestTracer_PreviewDeckEmitsMeasuredEvent' -v` — green (grammar + preview + suggestions + printing resolution).
- `go test ./server -run 'TestProductEvents_|TestMetrics' -v` — green (dedup, allowlist, ordering, migration reversal).
- `go test ./server -run 'TestSafeControl|TestSafeClient|TestProvider_' -race -v` — green (SSRF controls, kill switch, adapter routing).
- `go test ./server -count=1` (full suite, single run) — green, zero failures, matching the environment note.
- `go build ./...`, `go vet ./...` — clean.
- `cd app && npm test` — 28 passed / 3 skipped, matching the environment note.
- Read `server/games.go`, `server/ratelimit.go`, `server/deck_providers.go`, `pkg/deckimport/scanner.go`, `pkg/telemetry/{vocabulary,metrics}.go`, `server/schema.graphql`, `server/product_events.go` directly and confirmed the CR-01/CR-02/WR-01..04 review fixes are actually present in the code at HEAD (`e93920d`), not just claimed in `01-REVIEW-FIX.md`.
- Read `docs/research/deck-provider-feasibility.md` and `COVERAGE.md` in full.

## Goal Achievement

### Observable Truths (ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | A pasted decklist in any of the six supported syntaxes returns a preview with total cards, commander candidates, unresolved entries, warnings, and blocking errors; no nonblank row disappears | ✓ VERIFIED | `pkg/deckimport/scanner.go`'s `Parse`/`ParseWithSource` implement all 8 grammar rules; `AssertAccounting` mechanically enforces the no-row-disappears invariant on every return path. `server/schema.graphql`'s `DeckPreview` type carries `CardCount`, `CommanderCandidates`, `Unresolved`, `Warnings`, `CanContinue`, and the additive `BlockingErrors` field. Ran `go test ./pkg/deckimport -run TestScanner_RequiredSyntaxes -v` (24 subtests, all pass) and `go test ./server -run TestTracer_PreviewDeckEmitsMeasuredEvent -v` (pass) myself — not merely read the SUMMARY's claim. |
| 2 | Comma-containing names survive intact and resolve identically across both syntaxes; unresolved cards surface as unresolved, never silently substituted; preview and final library creation consume the same normalized result | ✓ VERIFIED | `TestScanner_RequiredSyntaxes/comma_truncation_regression` (ran, passes) proves `1,Atraxa, Praetors' Voice` yields the full name. `TestDeckImport_UnresolvedAccounting` (ran, passes) proves an unmatched name is dropped from both the countable set and the library, never substituted with a bare stub Card (the pre-existing `&Card{Name:}` substitution path was removed in plan 01-05). `grep` confirms `server/games.go` has zero `encoding/csv` references and `createLibraryFromDecklist` takes `*deckimport.ParsedDeck` directly — the same parsed/resolved value `previewDeck` builds, so a deck cannot be parsed twice under divergent rules. `TestDeckImport_SingleParse` (ran, passes) proves preview `CardCount` equals library length for the same input. |
| 3 | A recorded activation event groups by day, role, source, outcome in PostgreSQL; unknown event names/keys, oversized payloads, and prohibited content are rejected; the server attaches authenticated user IDs rather than trusting the client | ✓ VERIFIED | `pkg/telemetry/vocabulary.go`'s closed 15-event `Vocabulary` with per-event key allowlists, enforced by `ValidateEvent`; `forbiddenKeySubstrings`/`TestVocabulary_NoForbiddenKeys` structurally prevents any allowlisted key from ever naming a decklist, URL, password, token, JWT, IP, or clipboard value — so "prohibited content" cannot enter under any key name, and an attempt to send it under an unlisted key is rejected as `unknown_key`. `docs/analytics/product-event-funnel-example.sql` demonstrates the day/source/role/outcome grouping directly (read in full). `server/product_events.go`'s `TrackProductEvent` derives `UserID` exclusively from `authFromContext(ctx)`, never from client input — confirmed by reading the resolver; any user identifier the client sends is structurally ignored. Ran `go test ./server -run 'TestProductEvents_|TestMetrics' -v` myself: dedup (both-NULL, concurrent, replay-shape), ordering, allowlist rejections (6 subtests), migration up/down/up — all pass. |
| 4 | Prometheus reports guest-session, import, create/join, and board-activation counters/histograms with no username/session/user/game-ID label | ✓ VERIFIED (per ROADMAP's own criterion-4 scoping: declared-here, closed-later) | `pkg/telemetry/metrics.go`'s `AllowedLabelNames` is a closed 6-member set (`source`, `outcome`, `role`, `reason`, `provider`, `surface`) with an explicit comment ruling out username/session/user/game IDs; every `Observe*` method takes a compiler-enforced enum, never a bare string (WR-03's fix extended this to the deck-provider-fetch family too). `TestMetrics_AllCriterionFourFamiliesExist` (read directly) asserts all four families — guest session, deck import, game create, game join, board activation — exist with both a counter and histogram. `grep` confirms only `ObserveDeckImport` is called from production code today (plus the product-event counters); guest-session/create/join/board-activation are declared-but-unobserved by design, exactly as the ROADMAP's criterion-4 note requires and explicitly permits for this phase. No gap: the note says not to score this a gap on Phase 1 alone. |
| 5 | A written decision record names the first provider (with SSRF/kill-switch controls failing closed) or documents a no-go; either satisfies the phase | ✓ VERIFIED (per ROADMAP's own criterion-5 scoping: either outcome satisfies) | `docs/research/deck-provider-feasibility.md` (read in full) is a genuine, evidence-based, neutral six-criterion comparison; section 5 records the user's Moxfield selection, diverging from the document's own Archidekt recommendation, with two explicit open blockers. The SSRF/kill-switch controls are real and independently verified by running `go test ./server -run 'TestSafeControl|TestSafeClient|TestProvider_' -race -v` myself (22 subtests across denied-address, redirect/host, pre-resolution rejection, loopback-rebinding, body-cap, timeout, and kill-switch tests — all pass). `moxfieldAdapter.normalizeToDeckText` deliberately and verifiably always returns `errMoxfieldContractUnverified` (`TestProvider_MoxfieldContractUnverified`, ran, passes) rather than fabricating a field mapping — this is documented scaffolding, not a claimed working import, and `COVERAGE.md` honestly marks every Moxfield capability OPT-OUT. WR-01's fix (host allowlist now checked on the *initial* request, not only redirects) is confirmed present in `server/deck_providers.go`/`server/deck_import.go` and covered by `TestSafeClient_InitialRequestHostEnforced` (ran, passes). |

**Score:** 5/5 truths verified, 0 present-but-behavior-unverified.

### Requirements Traceability (Phase 1: REQ-ACT-001, REQ-ACT-002, REQ-ACT-003)

| Requirement | REQUIREMENTS.md status | Verified against code | Assessment |
|---|---|---|---|
| REQ-ACT-001 | Complete | `product_events` table + dedup index (migrations, `TestMigrations_ProductEvents`), closed vocabulary + `ValidateEvent`, all 12 Prometheus collectors, frontend `productEvents.ts` session-ID service, per-surface rate limiting (`pkg/ratelimit`) | ✓ SATISFIED — every clause of the ticket description (table, allowlists, six server-authoritative events, low-cardinality metrics, frontend session service, documented vocabulary/example query) has direct code and test evidence read and re-run above. |
| REQ-ACT-002 | Complete | `pkg/deckimport`'s full 8-rule grammar, `previewDeck`/`DeckPreview`, printing-aware two-stage resolution (`resolveDeckEntries`), bounded suggestions, single canonical parse site feeding both preview and `createLibraryFromDecklist` | ✓ SATISFIED — extraction from `createLibraryFromDecklist`, all required syntaxes, comma-name preservation, unresolved/warning/blocking-error reporting, and single-parse-site unification are all directly verified in code and by tests I ran myself. |
| REQ-ACT-003 | **Pending** (deliberately left open) | Feasibility spike + decision record (real, evidence-based); adapter interface + host-allowlist + DNS/IP validation + redirect revalidation + timeouts + 1 MiB cap + kill switch (all real, all independently re-tested); Moxfield `normalizeToDeckText` deliberately unimplemented, fails closed | **Honest, not a gap.** The requirement's own text requires "the one-day feasibility comparison... adapter interface... normalize through ACT-002" — a *working* first-provider import. That does not exist: no Moxfield response has ever been observed, and the normalizer is a documented stub. Leaving this Pending rather than marking it Complete is the accurate call; ROADMAP's own criterion 5 is explicitly satisfied by a no-go-or-scaffold outcome, but REQ-ACT-003 (Layer 2, the literal ticket) is a *stricter* bar than criterion 5, and the project's own traceability table correctly keeps it open. This is flagged in `human_verification` above only so a human explicitly signs off on shipping Phase 2 with this requirement still open, rather than the phase quietly rolling forward on an implicit assumption. |

No orphaned requirements: REQUIREMENTS.md maps exactly REQ-ACT-001/002/003 to Phase 1, matching this phase's own `requirements` field. REQ-A2/A6/A7 (Layer 1, partial) are correctly not claimed as complete anywhere in this phase's artifacts.

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `pkg/deckimport/{scanner,sections,sourcetype,result}.go` | Canonical grammar, sections, source detection | ✓ VERIFIED | Read in full; `go test ./pkg/deckimport -race` green including golden fixtures. |
| `pkg/telemetry/{vocabulary,metrics}.go` | Closed vocabulary, all criterion-4 collector families | ✓ VERIFIED | Read in full; `AllowedLabelNames` closed at 6; `go test ./pkg/telemetry -race` green. |
| `pkg/ratelimit/limiter.go` | Per-surface, per-client token-bucket registry | ✓ VERIFIED | Read; `go test ./pkg/ratelimit -race` green. |
| `server/deck_import.go` | `PreviewDeck` resolver, two-stage printing resolution, suggestions | ✓ VERIFIED | Read; all `TestDeckImport_*` pass against real Postgres + MTGJSON snapshot. |
| `server/product_events.go` | Never-fail write path, server-attached `UserID` | ✓ VERIFIED | Read; `UserID` sourced only from `authFromContext`. |
| `server/ratelimit.go` | `clientKeyFor` anchored on server-observed address | ✓ VERIFIED | CR-02 fix confirmed present in code (address-first, session-fallback only). |
| `server/games.go` | `JoinGame`/`CreateGame` life defaulting, single canonical parse site | ✓ VERIFIED | CR-01 fix confirmed present (`life <= 0` defaults to format `StartingLife`); zero `encoding/csv` references. |
| `server/deck_providers.go` | SSRF-safe client, kill switch, adapter registry | ✓ VERIFIED | WR-01 fix (initial-request host check) confirmed present; all security tests re-run and pass. |
| `docs/research/deck-provider-feasibility.md` | Neutral six-criterion comparison + decision | ✓ VERIFIED | Read in full; genuinely evidence-based, not fabricated. |
| `.planning/phases/.../COVERAGE.md` | Honest external-API coverage decision | ✓ VERIFIED | Every Moxfield capability OPT-OUT with a real stated reason. |
| `persistence/migrations*/2026080412*` | `product_events` + card-name-search migrations | ✓ VERIFIED | Migration up/down/up proofs re-run and pass against scratch databases. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `server/deck_import.go` (`previewDeck`) | `pkg/deckimport.Parse` | single call site | ✓ WIRED | Confirmed one canonical call site; no second grammar anywhere. |
| `server/games.go` (`createLibraryFromDecklist`) | `pkg/deckimport.ParsedDeck` | shared parsed value, not raw text | ✓ WIRED | Confirmed via grep + `TestDeckImport_SingleParse`. |
| `server/product_events.go` (`TrackProductEvent`) | `pkg/telemetry.ValidateEvent` | validation before any DB write | ✓ WIRED | Confirmed; rejection path never touches the database. |
| `server/ratelimit.go` (`allowRequest`) | `previewDeck` / `trackProductEvent` | wraps both public mutations | ✓ WIRED | Confirmed both resolvers call `allowRequest` before doing work. |
| `server/deck_import.go` (`previewDeckURL`) | `server/deck_providers.go` (`deckProviderAdapterFor` → `fetchDeckProviderURL`) | hostname-keyed adapter routing, allowlist-gated | ✓ WIRED | Confirmed via `TestProvider_MoxfieldAdapterRoutesThenFailsClosed`/`TestProvider_UnregisteredHostNeverFetched`, both re-run and passing. |

### Anti-Patterns Found

No blocking anti-patterns (TBD/FIXME/XXX with no follow-up reference, silent stub returns feeding a real path) were found in the files this phase modified. One deliberate, fully-documented stub exists:

| File | Line | Pattern | Severity | Impact |
|---|---|---|---|---|
| `server/deck_providers.go` | `moxfieldAdapter.normalizeToDeckText` | Always returns a static sentinel error, never a real field mapping | ℹ️ Info (documented, intentional) | Not a gap — this is the entire point of the D-14 checkpoint's "build skeleton, contract unverified" instruction, recorded in `01-07-SUMMARY.md`, `COVERAGE.md`, and the decision record, and is what keeps criterion 5 honest (no fabricated field mapping) rather than what would make it a defect. |

No other TODO/FIXME/XXX/placeholder markers were found in this phase's key files that lack a tracked follow-up. The `server/schema.resolvers.go` gqlgen stub `panic("not implemented")` noted by the code review (IN-01) is confirmed unreachable dead code (the real resolvers are on `*graphQLServer` directly) and was explicitly excluded from fix scope as expected scaffold noise, not a defect.

### Code Review Follow-Through

The phase's own `01-REVIEW.md` (2 critical + 4 warning findings) and `01-REVIEW-FIX.md` (all 6 fixed) were not taken on faith. I independently confirmed, by reading the current code at HEAD (`e93920d`) rather than the review-fix narrative alone:

- **CR-01** (`JoinGame` omitted-life "already eliminated" bug): fix present at `server/games.go:541-559`.
- **CR-02** (spoofable rate-limit key): fix present at `server/ratelimit.go:71-79`, address-anchored with session only as fallback.
- **WR-01** (host allowlist only checked on redirects): fix present — `fetchDeckProviderURL` now takes and checks `allowedHosts` before any request is built; `TestSafeClient_InitialRequestHostEnforced` re-run and passing.
- **WR-02** (`testAPI` DSN/skip masking): fix present in `server/test.go` (`DATABASE_URL` read with fallback; `t.Fatalf` on reachable-but-broken).
- **WR-03** (unbounded-string metric label): fix present — `DeckProvider` enum type now gates `ObserveDeckProviderFetch`.
- **WR-04** (doc/comment drift): fix present in `pkg/telemetry/metrics.go` comments and the vocabulary doc.

### Human Verification Required

1. **Confirm REQ-ACT-003's traceability status.**
   - **Test:** Review `docs/research/deck-provider-feasibility.md` section 5 and `COVERAGE.md`, and confirm with the product owner that shipping Phase 2 with REQ-ACT-003 still "Pending" (scaffolding + decision record, no working provider import) is the intended state, not an oversight.
   - **Expected:** Explicit sign-off that this is acceptable, OR a decision to open a follow-up ticket/phase note before Phase 2 begins.
   - **Why human:** This is a scope/traceability judgment call, not a code defect — the code and tests are exactly consistent with what the plans, SUMMARYs, and ROADMAP's own criterion-5 scoping note describe. Routing it to a human rather than silently defaulting to "passed" avoids quietly ratifying an intentional scope narrowing without an explicit decision.

## Gaps Summary

None. All five ROADMAP success criteria are verified against directly-read code and independently
re-run tests (not SUMMARY claims). All three phase requirements are accounted for, with
REQ-ACT-003 correctly left Pending for reasons that are honest and consistent with the ROADMAP's
own explicit scoping. The one open item is a traceability sign-off, not a defect, and is routed to
human verification rather than either a false PASS or an unwarranted gap.

---

_Verified: 2026-08-05T22:31:45Z_
_Verifier: Claude (gsd-verifier)_
