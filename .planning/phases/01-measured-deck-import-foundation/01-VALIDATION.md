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

Task IDs are assigned when PLAN.md files are written; this table seeds the requirement →
test mapping the planner must honor. Every row below has an automated command, so no task
implementing these behaviors may ship with a manual-only verify.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| TBD | TBD | 1 | REQ-ACT-002 | — | Six required syntaxes parse identically | unit | `go test ./pkg/deckimport -run TestScanner_RequiredSyntaxes` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-002 | — | Comma-name truncation regression cannot return | unit | `go test ./pkg/deckimport -run TestScanner_CommaNames` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-002 | — | `//` is a face separator mid-line, a comment at line start | unit | `go test ./pkg/deckimport -run TestScanner_DoubleFacedVsComment` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-002 | — | D-07 printing metadata extracted, name left clean | unit | `go test ./pkg/deckimport -run TestScanner_PrintingMetadata` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-002 | — | D-11 sideboard dropped with count; D-12 commander header preselects | unit | `go test ./pkg/deckimport -run TestScanner_Sections` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-002 | — | **No nonblank row disappears** — accounting invariant over all fixtures | unit (property) | `go test ./pkg/deckimport -run TestScanner_NoRowDisappears` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-002 | — | D-23 source detection never alters parsed entries | unit (differential) | `go test ./pkg/deckimport -run TestSourceDetection_DoesNotAffectParse` | ❌ W0 | ⬜ pending |
| TBD | TBD | 2 | REQ-ACT-002 | — | D-05/D-06 unresolved excluded from count and library; `100 - commanders` preserved | integration | `go test ./server -run TestDeckImport_UnresolvedAccounting` | ❌ W0 | ⬜ pending |
| TBD | TBD | 2 | REQ-ACT-002 | — | D-02/D-03 ≤3 ranked candidates, nearest returned below cutoff | integration | `go test ./server -run TestDeckImport_Suggestions` | ❌ W0 | ⬜ pending |
| TBD | TBD | 2 | REQ-ACT-002 | — | Preview and create/join consume one `ParsedDeck` | integration | `go test ./server -run TestDeckImport_SingleParse` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-001 | T-01 | Unknown event names / keys dropped, logged, counted — never errored | unit | `go test ./server -run TestProductEvents_Allowlist` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-001 | T-01 | **No allowlisted key names a forbidden concept** (structural privacy proof) | unit (DB-free) | `go test ./server -run TestEventVocabulary_NoForbiddenKeys` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-001 | T-02 | **No metric carries a high-cardinality label** (structural proof) | unit (DB-free) | `go test ./server -run TestMetrics_LabelAllowlist` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-001 | — | D-17 write failure does not fail the caller | unit | `go test ./server -run TestProductEvents_WriteFailureIsNonFatal` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-001 | — | Duplicate authoritative events deduplicate (PG14 `COALESCE` index) | integration | `go test ./server -run TestProductEvents_AuthoritativeDedup` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-001 | — | Migration up/down passes for prod **and** test schemas | integration | `go test ./server -run TestMigrations_ProductEvents` | ❌ W0 | ⬜ pending |
| TBD | TBD | 1 | REQ-ACT-001 | — | D-18 session ID stable across reloads; attribution allowlisted | unit (vitest) | `cd app && npx vitest --run __tests__/productEvents.spec.ts` | ❌ W0 | ⬜ pending |
| TBD | TBD | 3 | REQ-ACT-003 | T-03 | Non-allowlisted host, non-HTTPS scheme, off-allowlist redirect, redirect loop | unit | `go test ./server -run TestSafeClient_HostAndRedirect` | ❌ W0 | ⬜ pending |
| TBD | TBD | 3 | REQ-ACT-003 | T-03 | Private/reserved/link-local/CGNAT/metadata IPs rejected incl. `::ffff:` forms | unit (table) | `go test ./server -run TestSafeControl_DeniedAddresses` | ❌ W0 | ⬜ pending |
| TBD | TBD | 3 | REQ-ACT-003 | T-04 | DNS rebinding fails closed | unit | `go test ./server -run TestSafeControl_Rebinding` | ❌ W0 | ⬜ pending |
| TBD | TBD | 3 | REQ-ACT-003 | T-05 | Over-1 MiB response fails closed; a lying `Content-Length` does not help | unit | `go test ./server -run TestSafeClient_BodyCap` | ❌ W0 | ⬜ pending |
| TBD | TBD | 3 | REQ-ACT-003 | T-05 | Connect and total timeouts fail closed | unit | `go test ./server -run TestSafeClient_Timeouts` | ❌ W0 | ⬜ pending |
| TBD | TBD | 3 | REQ-ACT-003 | — | Kill switch off ⇒ normalized error + paste fallback, pasted import unaffected | unit | `go test ./server -run TestProvider_KillSwitch` | ❌ W0 | ⬜ pending |
| TBD | TBD | 3 | REQ-ACT-003 | — | Fixture contract detects a provider response-shape change | unit | `go test ./server -run TestProvider_FixtureContract` | ❌ W0 (post-checkpoint) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `pkg/deckimport/` package skeleton — `scanner.go`, `sections.go`, `sourcetype.go`, `result.go`. **Highest-leverage item**: it puts REQ-A2's acceptance criteria inside the only test target CI actually runs.
- [ ] `pkg/deckimport/scanner_test.go` — the required-syntax corpus table (ACT-002 / REQ-A2)
- [ ] `pkg/deckimport/testdata/` — Moxfield-shaped, Archidekt-shaped, and generic golden exports plus expected `ParsedDeck` files
- [ ] `server/product_events_test.go` — allowlist behavior, structural forbidden-key test, closed-at-15 test, non-fatal-write test
- [ ] `server/metrics_test.go` — label-allowlist registry walk and the D-24 enum-value check
- [ ] `server/deck_import_test.go` — DB-backed resolution, suggestions, D-05/D-06 accounting, single-parse proof
- [ ] `server/deck_providers_test.go` + `server/testdata/deck_providers/` — SSRF table tests and fixture contract. The safe-client tests are provider-agnostic and land **before** the D-14 checkpoint; only the fixture contract waits for it.
- [ ] `app/__tests__/productEvents.spec.ts` — session-ID stability across simulated reloads, attribution allowlisting, fire-and-forget error swallowing
- [ ] Migration up/down coverage for `product_events` across **both** `persistence/migrations/` and `persistence/migrations_test/`
- [ ] Framework install: **none required** — Go stdlib testing, testify, `matryer/is`, and vitest are all already present

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| ACT-003 provider feasibility outcome | REQ-ACT-003 | D-14 makes this a human decision at a `checkpoint:decision`. Which provider clears the gate — or whether both fail and paste-only ships — is judgement over spike evidence, not an assertion. | Review `docs/research/deck-provider-feasibility.md`. Confirm it names a provider **or** documents a reasoned no-go; confirm the D-15 hard condition (stable response shape) was evaluated; confirm no SSRF or reliability control was weakened to force a provider through. Approve or reject at the checkpoint. |
| Per-event metadata key transcription | REQ-ACT-001 | RESEARCH.md marks this **Assumption A1 — LOW confidence**: the per-event key table is `[ASSUMED]` for 13 of 15 events. Adopting the assumed table would silently drop real events in production with only a counter to show for it. | Read `docs/product/2026-07-23-deck-to-game-activation-prd.md` §Product event vocabulary and transcribe the authoritative per-event keys. This must be an explicit plan task, not a review step. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s on the quick path
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
