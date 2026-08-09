---
phase: 2
slug: guest-host-activation
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-08
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Seeded by `/gsd-plan-phase 2` from `02-RESEARCH.md` § Validation Architecture.
> The Per-Task Verification Map is filled in once PLAN.md task IDs exist.

---

## Test Infrastructure

This phase spans two test stacks. Both must stay green.

| Property | Backend (Go) | Frontend (Vue) |
|----------|--------------|----------------|
| **Framework** | Go `testing` + `-race` | Vitest + `@vue/test-utils` (jsdom) |
| **Config file** | `Makefile` (`test-api`) | `app/package.json` (`"test": "vitest --run"`) |
| **Quick run command** | `go test ./server -run <TestName> -race` | `npx vitest run app/__tests__/<File>.spec.ts` |
| **Full suite command** | `make test-api` | `npm test` (from `app/`) |
| **Test DB** | `persistence/migrations_test/` against real Postgres (Docker), reset per `server/test.go:76-79` | n/a |
| **Estimated runtime** | ~60–120s (real Postgres, `-race`) | ~10–20s |

**Additional full-suite gate:** `npm run type-check` — the generated GraphQL types must stay in
sync after this phase's schema changes (`display_name`, `guestSession`, `claimGuestAccount`).

**E2E is out of scope this phase.** `app/e2e/create-and-join-game.spec.ts` exists, but guest
host/join Playwright coverage is REQ-ACT-013 (Phase 5). This phase builds the resolvers and
routes Phase 5 will drive.

---

## Sampling Rate

- **After every task commit:** Run the narrowest `go test -run <Name> -race` or
  `npx vitest run <file>` for the touched area.
- **After every plan wave:** Run `make test-api` AND `npm test` (from `app/`) AND
  `npm run type-check` — all three, both stacks, full suites.
- **Before `/gsd-verify-work`:** Full suite green on both stacks, plus a manual dev-server
  smoke run confirming `previewDeck` is reachable from `/play` (no automated E2E until Phase 5).
- **Max feedback latency:** 120 seconds (backend full suite is the long pole).

---

## Per-Task Verification Map

*Seeded from RESEARCH.md's Phase Requirements → Test Map. Task IDs are assigned when PLAN.md
files are written; `validate-phase` completes this table.*

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 02-01 T1 | 02-01 | 1 | REQ-ACT-005 / 006 | T-02-01, T-02-04 | Tracer: logged-out paste reaches a live board; guest row created once, name unique, event written once | integration (Go) + component | `go test ./server -run TestGuestHost_Tracer -race` and `npm --prefix app run test -- __tests__/QuickStartView.spec.ts` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-01 T1 | 02-01 | 1 | REQ-ACT-005 | — | Generated names unique & collision-retried; guest creation needs no username/password | unit/integration | `go test ./server -run TestGuestUsers_Create -race` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-01 T1 | 02-01 | 1 | REQ-ACT-005 | T-02-01, T-02-02 | Guest creation rate-limited behind a kill switch | unit | `go test ./server -run 'TestGuestUsers_RateLimit\|TestGuestUsers_KillSwitch' -race` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-01 T1 | 02-01 | 1 | REQ-ACT-005 | — | Migration applies cleanly on both prod and test schemas, files byte-identical | migration | `go test ./server -run TestMigrations_GuestUsers -race` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-01 T2 | 02-01 | 1 | REQ-ACT-005 | T-02-03 | Guest credential stored under its own key, absent from the auth profile and the analytics ID | unit | `npm --prefix app run test -- __tests__/authGuest.spec.ts` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-02 T1 | 02-02 | 2 | REQ-ACT-005 | T-02-08 | Guest cannot log in via password endpoint before claiming | unit | `go test ./server -run TestUsers_LoginRejectsGuest -race` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-02 T1 | 02-02 | 2 | REQ-ACT-005 | T-02-11, T-02-14 | Closed error-code set leaks no internals; display name capped by rune and stripped of control chars | unit | `go test ./server -run 'TestActivationErrors\|TestGuestUsers_DisplayNameValidation' -race` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-02 T2 | 02-02 | 2 | REQ-ACT-005 | T-02-09 | `expires_at` gates minting a NEW token only (hand-marked row — see note); silent re-issue works | unit | `go test ./server -run 'TestGuestUsers_Refresh\|TestGuestUsers_ExpiredRow' -race` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-02 T3 | 02-02 | 2 | REQ-ACT-005 | T-02-10, T-02-12 | Claim preserves the row and its games; cleanup never deletes a referenced guest | integration | `go test ./server -run 'TestGuestUsers_Claim\|TestGuestUsers_Cleanup' -race` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-03 T1 | 02-03 | 2 | REQ-ACT-004 | T-02-16, T-02-17 | Paste → preview → correct, every UI state, four failure codes, no raw output rendered | component | `npm --prefix app run test -- __tests__/DeckImportPanel.spec.ts __tests__/activationErrors.spec.ts` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-03 T2 | 02-03 | 2 | REQ-ACT-004 | T-02-18, T-02-20 | Commander review via commanderPartner.ts; draft survives refresh and failed nav, cleared on success | component | `npm --prefix app run test -- __tests__/CommanderReview.spec.ts __tests__/quickStartDraft.spec.ts` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-03 T3 | 02-03 | 2 | REQ-ACT-004 | — | Host and join submit the same normalized deck through one component | integration | `npm --prefix app run test -- __tests__/FormCreateGame.integration.spec.ts __tests__/JoinGame.integration.spec.ts` | ✅ (updated by the task) | ⬜ pending |
| 02-04 T1 | 02-04 | 3 | REQ-ACT-006 / Criterion 5 | T-02-21, T-02-22, T-02-26 | `game_created` is server-written; a retried call does not double-count a conversion | integration | `go test ./server -run 'TestGames_Create\|TestProductEvents_AuthoritativeDedup' -race` | ✅ (extended by the task) | ⬜ pending |
| 02-04 T2 | 02-04 | 3 | REQ-ACT-006 | T-02-24 | `/play` public while every previously-protected route still redirects | router unit | `npm --prefix app run test -- __tests__/router.spec.ts` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-04 T2 | 02-04 | 3 | REQ-ACT-006 | T-02-25 | Guest and authenticated branches; error preservation across all 4 failure classes; exactly 3 client events | component/store | `npm --prefix app run test -- __tests__/QuickStartView.spec.ts` | ✅ (extended by the task) | ⬜ pending |
| 02-05 T1 | 02-05 | 4 | REQ-ACT-005 | — | Display name captured onto the game record at write time; legacy payloads still read | integration | `go test ./server -run TestGames_ -race` | ✅ (extended by the task) | ⬜ pending |
| 02-05 T2 | 02-05 | 4 | REQ-ACT-005 | T-02-28, T-02-29, T-02-30 | Every display site uses the fallback; every identity site still compares on Username; every document selects DisplayName | unit/component | `npm --prefix app run test -- __tests__/displayName.spec.ts` | ❌ W0 (authored by the task) | ⬜ pending |
| 02-06 T1 | 02-06 | 5 | REQ-ACT-004 / 005 / 006 | T-02-31, T-02-32, T-02-33 | The four manual-only verifications below | manual | *(blocking human checkpoint — see Manual-Only Verifications)* | n/a | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

**Note on the expired-guest test (CONTEXT.md § Specific Ideas):** under D-2.1 the backing guest
row never expires, so criterion 3's "stop working when the backing guest expires" clause can
never fire in a live scenario. D-2.2 keeps `expires_at` enforced in authorization, so the check
is still built and still tested — but **against a hand-marked row**, not an organically expired
one. Verification must not expect to observe this rejection end-to-end.

---

## Wave 0 Requirements

Every gap below is closed by the task that needs it: each of these tasks carries `tdd="true"` with a
`<behavior>` block, so the test file is authored (and fails first) inside the same task that
implements against it. There is no separate Wave 0 plan.

- [ ] `server/guest_users_test.go` — REQ-ACT-005's create / collision / rate-limit / kill-switch /
      expiry / login-rejection / claim / conflict / relationship-preservation clauses.
      *Owned by 02-01 T1 (create, collision, rate limit, kill switch, dedup, migration) and
      02-02 T1/T2/T3 (validation, refresh, expiry, claim).*
- [ ] Migration test for the `20260809120000_guest_users` pair — model on
      `TestMigrations_CardNameSearch` (`server/deck_import_test.go:150`). *Owned by 02-01 T1.*
- [ ] `server/guest_cleanup_test.go` — cleanup preservation, idempotency, and the no-scheduled-work
      proof. *Owned by 02-02 T3.*
- [ ] `server/activation_errors_test.go` — closed code set, no leaked internals. *Owned by 02-02 T1.*
- [ ] `app/__tests__/DeckImportPanel.spec.ts` — *Owned by 02-03 T1.*
- [ ] `app/__tests__/activationErrors.spec.ts` — *Owned by 02-03 T1.*
- [ ] `app/__tests__/CommanderReview.spec.ts` and `app/__tests__/quickStartDraft.spec.ts` —
      *Owned by 02-03 T2.*
- [ ] `app/__tests__/QuickStartView.spec.ts` — created by 02-01 T1 (tracer path), extended by
      02-03 T2 (draft survival, responsive) and 02-04 T2 (events, branches, four failure classes).
- [ ] `app/__tests__/authGuest.spec.ts` — guest credential storage separation. *Owned by 02-01 T2.*
- [ ] `app/__tests__/router.spec.ts` — `/play` never redirects; every previously-protected route
      still does. *Owned by 02-04 T2.*
- [ ] `app/__tests__/displayName.spec.ts` — the display-versus-identity invariant and the
      GraphQL-document completeness assertion. *Owned by 02-05 T2.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `previewDeck` reachable from `/play` end-to-end | REQ-ACT-006 | No Playwright coverage for the guest path until Phase 5 (REQ-ACT-013) | Dev server, logged out, open `/play`, paste a real decklist, confirm preview renders and continue reaches `/games/:id` |
| ≥768px breakpoint renders correctly on tablet | D-2.10 | First `@media` query in the codebase; visual, not assertable in jsdom | Resize to 768px and 1024px; confirm activation components reflow without overflow |
| Phone heads-up notice appears below 768px | D-2.11 | Visual copy check, no hard bounce to assert | Load `/play` at <768px; confirm notice is visible and the import flow still functions |
| Generated guest names are never offensive or absurd | D-2.5 | Curation is human judgment, not assertable | Review the full curated adjective × noun product before merge |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags (`vitest --run`, never bare `vitest`)
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
