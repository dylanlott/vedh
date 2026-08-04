---
phase: 1
slug: measured-deck-import-foundation
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-04
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Derived from `01-RESEARCH.md` §Validation Architecture (line 1161).
> Per-task map reconciled against the written plan set on 2026-08-04; plan 01-01 task 3 owns any
> further reconciliation discovered during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework (Go)** | `go test` + `github.com/stretchr/testify v1.11.1` + `github.com/matryer/is v1.4.0` (go.mod:21,23) |
| **Framework (frontend)** | `vitest ^1.1.9` + `jsdom ^22.1.0` + `@vue/test-utils ^2.4.6` |
| **Config file (Go)** | none — standard `go test`; `server/main_test.go` provides `TestMain` |
| **Config file (frontend)** | inline `"vitest"` block in `app/package.json` — **there is no `vitest.config.*` file** |
| **Quick run command** | `go test ./pkg/... -race` |
| **Full suite command** | `make test-api` (= `go test -v ./server/... -race`) then `cd app && npm test` |
| **Estimated runtime** | quick ~seconds (DB-free); full requires local Postgres **and** a ~500 MB `All Printings.json` |

### ⚠ CI reality — this drives the whole strategy

CI runs `make test-unit` = `go test -v ./pkg/... -race` **only**. `./server/...` is in no
workflow and cannot run without live Postgres plus an MTGJSON dump absent from the repo
(reproduced by the researcher: `test setup failed: dial tcp [::1]:5432`).

**Consequence:** the deterministic parser grammar must live in a new `pkg/deckimport/`
package, not in `server/`. That is the only way REQ-A2's acceptance criteria gate every
PR at zero CI cost. Anything placed in `server/` is effectively untested in CI today.

---

## Sampling Rate

- **After every task commit:** `go test ./pkg/... -race` — DB-free, seconds, and exactly what CI runs
- **After every plan wave:** `go test ./pkg/... -race` + `go test ./server/... -race` + `cd app && npm test` + `cd app && npm run type-check`
- **Before `/gsd-verify-work`:** full suite green **and** `make generate` produces no uncommitted diff
- **Max feedback latency:** ~30 seconds for the quick path

---

## Per-Task Verification Map

Reconciled against the written plan set on 2026-08-04. `Task ID` is `{plan}-T{task}`; `Wave` is the
executing plan's frontmatter wave. Every row has an automated command, so no task implementing these
behaviors may ship with a manual-only verify. Rows marked **DB** run in `package server`, whose
`TestMain` migrates Postgres and imports an MTGJSON snapshot before any test runs — those tasks carry
a `<precondition>` naming both, and none of them is CI coverage (see the CI-reality note above).

**Terminology, because "authoritative" means two different things in this phase.**
`EventSpec.Authoritative` marks the **six server-owned** event names a client may not submit. The
migration's dedup predicate names a different, smaller set: the **four deduplicated** names
`game_created`, `player_joined`, `guest_session_created`, `account_claimed`. `deck_import_succeeded`
and `deck_import_failed` are server-owned but deliberately *outside* the predicate, so they persist on
every emission. Two identifiers keep the older word for continuity with `01-RESEARCH.md` — the index
`product_events_authoritative_once` and the test `TestProductEvents_AuthoritativeDedup` — and both mean
"the deduplicated four". Renaming them was considered and rejected at plan time: the strings appear in
`01-RESEARCH.md` line 1235 as well as here, so a rename would trade one ambiguous name for one stale
cross-reference. `01-01-PLAN.md` task 2 carries the same note where the test is specified.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 01-02-T1 | 01-02 | 2 | REQ-ACT-002 | — | Six required syntaxes parse identically | unit | `go test ./pkg/deckimport -run TestScanner_RequiredSyntaxes` | ❌ W0 | ⬜ pending |
| 01-02-T1 | 01-02 | 2 | REQ-ACT-002 | — | Comma-name truncation regression cannot return | unit | `go test ./pkg/deckimport -run TestScanner_RequiredSyntaxes` | ❌ W0 | ⬜ pending |
| 01-02-T1 | 01-02 | 2 | REQ-ACT-002 | — | `//` is a face separator mid-line, a comment at line start | unit | `go test ./pkg/deckimport -run TestScanner_RequiredSyntaxes` | ❌ W0 | ⬜ pending |
| 01-02-T1 | 01-02 | 2 | REQ-ACT-002 | — | D-07 printing metadata extracted, name left clean | unit | `go test ./pkg/deckimport -run TestScanner_RequiredSyntaxes` | ❌ W0 | ⬜ pending |
| 01-02-T2 | 01-02 | 2 | REQ-ACT-002 | — | D-11 sideboard dropped with count; D-12 commander header preselects | unit | `go test ./pkg/deckimport -run TestSections_DropAndPreselect` | ❌ W0 | ⬜ pending |
| 01-02-T1 | 01-02 | 2 | REQ-ACT-002 | T-01-14 | **No nonblank row disappears** — accounting invariant over all fixtures | unit (property) | `go test ./pkg/deckimport -run TestScanner_NoRowDisappears` | ❌ W0 | ⬜ pending |
| 01-02-T1 | 01-02 | 2 | REQ-ACT-002 | T-01-05 | Byte and line caps refuse oversized input before scanning; parsing is pure and race-free | unit | `go test ./pkg/deckimport -race -run 'TestScanner_InputBounds\|TestScanner_IsPure\|TestScanner_ConcurrentParseIsSafe'` | ❌ W0 | ⬜ pending |
| 01-02-T2 | 01-02 | 2 | REQ-ACT-002 | — | Golden fixtures parse to a checked-in expected `ParsedDeck`, so a grammar change that moves a fixture fails loudly | unit (golden) | `go test ./pkg/deckimport -run TestGolden_Fixtures` | ❌ W0 | ⬜ pending |
| 01-02-T2 | 01-02 | 2 | REQ-ACT-002 | — | D-23 source detection never alters parsed entries | unit (differential) | `go test ./pkg/deckimport -run TestSourceDetection_DoesNotAffectParse` | ❌ W0 | ⬜ pending |
| 01-02-T2 | 01-02 | 2 | REQ-ACT-001 | T-01-02 | D-24 detection returns only enum members (parallel-vocabulary guard) | unit | `go test ./pkg/deckimport -run TestSourceDetection_ReturnsOnlyEnumMembers` | ❌ W0 | ⬜ pending |
| 01-04-T1 | 01-04 | 3 | REQ-ACT-002 | T-01-21 | Name-search indexes exist; up/down/up cycle passes both directories; extension survives down | integration **DB** | `go test ./server -run TestMigrations_CardNameSearch` | ❌ W0 | ⬜ pending |
| 01-04-T2 | 01-04 | 3 | REQ-ACT-002 | T-01-21 | D-09/D-10 printing disambiguation, missing-printing warning, index-usable lookup | integration **DB** | `go test ./server -run TestDeckImport_PrintingDisambiguation` | ❌ W0 | ⬜ pending |
| 01-05-T1 | 01-05 | 4 | REQ-ACT-002 | — | D-02/D-03 ≤3 ranked candidates, nearest returned below cutoff | integration **DB** | `go test ./server -run TestDeckImport_Suggestions` | ❌ W0 | ⬜ pending |
| 01-05-T2 | 01-05 | 4 | REQ-ACT-002 | — | D-05/D-06 unresolved excluded from count and library; `100 - commanders` preserved | integration **DB** | `go test ./server -run TestDeckImport_UnresolvedAccounting` | ❌ W0 | ⬜ pending |
| 01-05-T2 | 01-05 | 4 | REQ-ACT-002 | — | Preview and create/join consume one `ParsedDeck` | integration **DB** | `go test ./server -run TestDeckImport_SingleParse` | ❌ W0 | ⬜ pending |
| 01-01-T1 | 01-01 | 1 | REQ-ACT-001, REQ-ACT-002 | — | Tracer: pasted card ⇒ preview + one event row + counter increment | integration **DB** | `go test ./server -run TestTracer_PreviewDeckEmitsMeasuredEvent` | ❌ W0 | ⬜ pending |
| 01-01-T1 | 01-01 | 1 | REQ-ACT-001 | — | D-17 write failure does not fail the caller | integration **DB** | `go test ./server -run TestProductEvents_WriteFailureIsNonFatal` | ❌ W0 | ⬜ pending |
| 01-01-T2 | 01-01 | 1 | REQ-ACT-001 | T-01-01 | Unknown names/keys/oversized/client-authoritative/blank-session dropped: no row, counter +1 exactly, no neighbouring reason moved | integration **DB** | `go test ./server -run TestProductEvents_Allowlist` | ❌ W0 | ⬜ pending |
| 01-01-T2 | 01-01 | 1 | REQ-ACT-001 | T-01-04 | Duplicate **deduplicated-set** events (the four in the index predicate) collapse to one row incl. both-NULL keys; a **repeatable** control (`deck_import_succeeded`) persists twice; a client replay of a server-owned `game_created` leaves the server-side original as the single row | integration **DB** | `go test ./server -run TestProductEvents_AuthoritativeDedup` | ❌ W0 | ⬜ pending |
| 01-01-T2 | 01-01 | 1 | REQ-ACT-001 | T-01-04 | Two concurrent in-flight inserts of one **deduplicated** event (`game_created`) resolve to one row | integration **DB** | `go test ./server -run TestProductEvents_ConcurrentInsert` | ❌ W0 | ⬜ pending |
| 01-01-T2 | 01-01 | 1 | REQ-ACT-001 | — | Identical `occurred_at` orders by ascending id; funnel collapses to earliest per session | integration **DB** | `go test ./server -run TestProductEvents_OccurredAtTieOrdering` | ❌ W0 | ⬜ pending |
| 01-01-T2 | 01-01 | 1 | REQ-ACT-001 | — | Migration up/down/up passes for prod **and** test schemas, in scratch databases | integration **DB** | `go test ./server -run TestMigrations_ProductEvents` | ❌ W0 | ⬜ pending |
| 01-01-T3 | 01-01 | 1 | REQ-ACT-001 | T-01-01 | **No allowlisted key names a forbidden concept** (structural privacy proof) | unit (DB-free, CI) | `go test ./pkg/telemetry -run TestVocabulary_NoForbiddenKeys` | ❌ W0 | ⬜ pending |
| 01-01-T3 | 01-01 | 1 | REQ-ACT-001 | — | D-20 vocabulary closed at 15; D-21 per-event keys; six authoritative events | unit (DB-free, CI) | `go test ./pkg/telemetry -run 'TestVocabulary_IsClosedAtFifteen|TestVocabulary_AuthoritativeSetMatchesSpec'` | ❌ W0 | ⬜ pending |
| 01-01-T3 | 01-01 | 1 | REQ-ACT-001 | T-01-01 | Every rejection path returns its own bounded `Rejection` — byte equality, oversized value, client-authoritative, missing session | unit (DB-free, CI) | `go test ./pkg/telemetry -run TestValidateEvent_Rejections` | ❌ W0 | ⬜ pending |
| 01-01-T3 | 01-01 | 1 | REQ-ACT-001 | T-01-02 | **No metric carries a high-cardinality label** (structural proof) | unit (DB-free, CI) | `go test ./pkg/telemetry -run TestMetrics_LabelAllowlist` | ❌ W0 | ⬜ pending |
| 01-01-T3 | 01-01 | 1 | REQ-ACT-001 | T-01-02 | All four criterion-4 families exist with a counter and a histogram; every label value bounded | unit (DB-free, CI) | `go test ./pkg/telemetry -run 'TestMetrics_AllCriterionFourFamiliesExist|TestMetrics_SourceLabelValuesAreEnumMembers|TestMetrics_ReasonLabelValuesAreBounded'` | ❌ W0 | ⬜ pending |
| 01-03-T1 | 01-03 | 2 | REQ-ACT-001 | — | D-18 session ID stable across reloads; attribution allowlisted | unit (vitest) | `cd app && npx vitest --run __tests__/productEvents.spec.ts` | ❌ W0 | ⬜ pending |
| 01-06-T1 | 01-06 | 5 | REQ-ACT-001 | T-01-13, T-01-27 | Burst exhaustion, per-key and per-surface isolation, idle eviction, registry returns to zero entries | unit (DB-free, CI) | `go test ./pkg/ratelimit -race` | ❌ W0 | ⬜ pending |
| 01-06-T1 | 01-06 | 5 | REQ-ACT-001 | T-01-13 | Both limiter outcomes counted exactly once; a limited preview names no limit value | integration **DB** | `go test ./server -run TestRateLimit` | ❌ W0 | ⬜ pending |
| 01-06-T2 | 01-06 | 5 | REQ-ACT-003 | T-01-06, T-01-07 | Non-allowlisted host, non-HTTPS scheme, off-allowlist redirect, redirect bound, case-only host difference | unit (hermetic) **DB** | `go test ./server -run TestSafeClient_HostAndRedirect` | ❌ W0 | ⬜ pending |
| 01-06-T2 | 01-06 | 5 | REQ-ACT-003 | T-01-06 | Private/reserved/link-local/CGNAT/metadata IPs rejected incl. `::ffff:` forms | unit (table) **DB** | `go test ./server -run TestSafeControl_DeniedAddresses` | ❌ W0 | ⬜ pending |
| 01-06-T2 | 01-06 | 5 | REQ-ACT-003 | T-01-06 | An address inside two deny prefixes names the first in declared order | unit **DB** | `go test ./server -run TestSafeControl_ReportsFirstMatchingPrefix` | ❌ W0 | ⬜ pending |
| 01-06-T2 | 01-06 | 5 | REQ-ACT-003 | T-01-25 | DNS rebinding fails closed, proving the hook is wired into the client | unit **DB** | `go test ./server -run TestSafeClient_RebindingFailsClosed` | ❌ W0 | ⬜ pending |
| 01-06-T2 | 01-06 | 5 | REQ-ACT-003 | T-01-06 | Empty host and non-secure scheme refused before any name resolution | unit **DB** | `go test ./server -run TestSafeClient_RejectsBeforeResolution` | ❌ W0 | ⬜ pending |
| 01-06-T2 | 01-06 | 5 | REQ-ACT-003 | T-01-09 | Over-1 MiB response fails closed; a lying `Content-Length` does not help | unit **DB** | `go test ./server -run TestSafeClient_BodyCap` | ❌ W0 | ⬜ pending |
| 01-06-T2 | 01-06 | 5 | REQ-ACT-003 | T-01-26 | Connect and total timeouts fail closed | unit **DB** | `go test ./server -run TestSafeClient_Timeouts` | ❌ W0 | ⬜ pending |
| 01-06-T2 | 01-06 | 5 | REQ-ACT-003 | T-01-11 | A zero-byte provider body is a normalized provider error, not an empty deck | unit **DB** | `go test ./server -run TestSafeClient_ZeroByteBodyIsProviderError` | ❌ W0 | ⬜ pending |
| 01-06-T3 | 01-06 | 5 | REQ-ACT-003 | T-01-28 | Kill switch off ⇒ normalized error + paste fallback, zero dial attempts | unit **DB** | `go test ./server -run TestProvider_KillSwitch` | ❌ W0 | ⬜ pending |
| 01-06-T3 | 01-06 | 5 | REQ-ACT-003 | T-01-28 | Flag set with an empty allowlist is treated as disabled | unit **DB** | `go test ./server -run TestProvider_KillSwitchRequiresAllowlist` | ❌ W0 | ⬜ pending |
| 01-06-T3 | 01-06 | 5 | REQ-ACT-003 | — | Pasted import is identical with the flag set and unset | unit **DB** | `go test ./server -run TestProvider_PasteUnaffectedByFlag` | ❌ W0 | ⬜ pending |
| 01-07-T3 | 01-07 | 6 | REQ-ACT-003 | T-01-32 | Fixture contract detects a provider response-shape change | unit **DB** | `go test ./server -run TestProvider_FixtureContract` | ❌ W0 (provider branch only, post-checkpoint) | ⬜ pending |
| 01-07-T3 | 01-07 | 6 | REQ-ACT-003 | T-01-31 | A private deck is unsupported; no user third-party credential is ever requested | unit **DB** | `go test ./server -run TestProvider_PrivateDeckIsUnsupported` | ❌ W0 (provider branch only) | ⬜ pending |
| 01-07-T3 | 01-07 | 6 | REQ-ACT-003 | T-01-33 | No-go branch: the deck-URL path stays a paste-fallback error in either flag state | unit **DB** | `go test ./server -run TestProvider_NoGoPathIsStable` | ❌ W0 (no-go branch only) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

### Renames and relocations recorded during planning

| Was | Is | Why |
|-----|----|-----|
| `TestSafeControl_Rebinding` | `TestSafeClient_RebindingFailsClosed` (plan 01-06 task 2) | The assertion proves the dial-control hook is *wired into the constructed client*, not that the control is correct in isolation; the `SafeClient_` prefix marks the tests that go through the client. |
| `TestEventVocabulary_NoForbiddenKeys` in `server/` | `TestVocabulary_NoForbiddenKeys` in `pkg/telemetry` (plan 01-01 task 3) | Moved to the only package CI runs. A DB-free assertion in `package server` never executes, because `TestMain` hard-exits without Postgres and the MTGJSON snapshot. |
| `TestMetrics_LabelAllowlist` in `server/` | `TestMetrics_LabelAllowlist` in `pkg/telemetry` (plan 01-01 task 3) | Same reason; the test now gathers from a private registry it constructs and exercises itself. |
| `server/metrics_test.go` (Wave 0 item) | `pkg/telemetry/metrics_test.go` | Same reason. |
| Migration up/down coverage (unnamed Wave 0 item) | `TestMigrations_ProductEvents` (plan 01-01 task 2) and `TestMigrations_CardNameSearch` (plan 01-04 task 1) | Both named, both using a scratch database per migration directory so no test migrates the shared database down. |

Further renames or relocations discovered during execution are appended here by plan 01-01 task 3,
each with a reason, and reported in that plan's summary.

---

## Wave 0 Requirements

- [ ] `pkg/deckimport/` package skeleton — `scanner.go`, `sections.go`, `sourcetype.go`, `result.go`. **Highest-leverage item**: it puts REQ-A2's acceptance criteria inside the only test target CI actually runs.
- [ ] `pkg/deckimport/scanner_test.go` — the required-syntax corpus table (ACT-002 / REQ-A2)
- [ ] `pkg/deckimport/testdata/` — Moxfield-shaped, Archidekt-shaped, and generic golden exports plus expected `ParsedDeck` files
- [ ] `pkg/telemetry/vocabulary_test.go` — structural forbidden-key test, closed-at-15 test, authoritative-set test, and the full `ValidateEvent` rejection table including the missing-session case
- [ ] `pkg/telemetry/metrics_test.go` — label-allowlist registry walk, the criterion-4 family-existence check, the D-24 source enum-value check, and the bounded `reason`/`role` value checks
- [ ] `pkg/ratelimit/limiter_test.go` — burst exhaustion, per-key and per-surface isolation, idle eviction via an injected clock, and bounded registry growth
- [ ] `server/product_events_test.go` — non-fatal-write test, allowlist *behaviour* (no row plus an exact counter delta), dedup-predicate coverage incl. both-NULL keys, concurrent insert, `occurred_at` tie ordering, and `TestMigrations_ProductEvents` with its named `withScratchMigrationDB` helper (which plan 01-04 task 1 reuses)
- [ ] `server/deck_import_test.go` — the tracer end-to-end test, DB-backed resolution, suggestions, D-05/D-06 accounting, single-parse proof, `TestMigrations_CardNameSearch`
- [ ] `server/ratelimit_test.go` — the wrapper's counter-on-both-paths assertion only; the registry logic is covered in `pkg/ratelimit`
- [ ] `server/deck_providers_test.go` + `server/testdata/deck_providers/` — SSRF table tests and fixture contract. The safe-client tests are provider-agnostic and land **before** the D-14 checkpoint; only the fixture contract waits for it.
- [ ] `app/__tests__/productEvents.spec.ts` — session-ID stability across simulated reloads, attribution allowlisting, fire-and-forget error swallowing
- [ ] Framework install: **none required** — Go stdlib testing, testify, `matryer/is`, and vitest are all already present

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| ACT-003 provider feasibility outcome | REQ-ACT-003 | D-14 makes this a human decision at a `checkpoint:decision`. Which provider clears the gate — or whether both fail and paste-only ships — is judgement over spike evidence, not an assertion. | Review `docs/research/deck-provider-feasibility.md`. Confirm it names a provider **or** documents a reasoned no-go; confirm the D-15 hard condition (stable response shape) was evaluated; confirm no SSRF or reliability control was weakened to force a provider through. Approve or reject at the checkpoint. |
| Per-event metadata key transcription | REQ-ACT-001 | RESEARCH.md marks this **Assumption A1 — LOW confidence**: the per-event key table is `[ASSUMED]` for 13 of 15 events. Adopting the assumed table would silently drop real events in production with only a counter to show for it. | Read `docs/product/2026-07-23-deck-to-game-activation-prd.md` §Product event vocabulary and transcribe the authoritative per-event keys. **Now an explicit plan task, not a review step:** plan 01-01 task 1 §3 transcribes from the PRD table and forbids adopting the `[ASSUMED]` research table; plan 01-01 task 3 re-checks the transcription row by row and records every PRD-vs-research discrepancy in `docs/analytics/product-event-vocabulary.md` with per-key provenance. What remains manual is one human reading the two tables side by side at review. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s on the quick path
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
