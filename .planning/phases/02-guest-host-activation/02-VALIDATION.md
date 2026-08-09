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
| TBD | TBD | TBD | REQ-ACT-004 | — | Paste → preview → correct → select commander → continue, state preserved across failures | component | `npx vitest run app/__tests__/DeckImportPanel.spec.ts` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | REQ-ACT-004 | — | Host and join submit the same normalized deck | integration | `npx vitest run app/__tests__/FormCreateGame.integration.spec.ts app/__tests__/JoinGame.integration.spec.ts` | ✅ (needs update) | ⬜ pending |
| TBD | TBD | TBD | REQ-ACT-005 | — | Guest creation requires no username/password; generated names unique & escaped | unit/integration | `go test ./server -run TestGuestUsers_Create -race` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | REQ-ACT-005 | Guest token vs password `Login` (EoP) | Guest cannot log in via password endpoint before claiming | unit | `go test ./server -run TestUsers_LoginRejectsGuest -race` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | REQ-ACT-005 | Mass guest creation (DoS) | Guest creation rate-limited behind kill switch | unit | `go test ./server -run TestGuestUsers_RateLimit -race` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | REQ-ACT-005 | — | Cleanup never deletes a guest referenced by an active game | integration | `go test ./server -run TestGuestUsers_CleanupPreservesActiveGames -race` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | REQ-ACT-005 | — | `expires_at` gates minting a NEW token only (test against hand-marked row — see note) | unit | `go test ./server -run TestGuestUsers_ExpiredRowRejected -race` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | REQ-ACT-005 | — | Migration applies cleanly on both prod and test schemas | migration | `go test ./server -run TestMigrations_GuestUsers -race` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | REQ-ACT-005 | XSS via `display_name` (Tampering) | Server caps length + strips control chars independent of client | unit | `go test ./server -run TestGuestUsers_DisplayNameValidation -race` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | REQ-ACT-006 | — | `/play` public while unrelated private routes remain protected | router unit | `npx vitest run app/__tests__/router.spec.ts` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | REQ-ACT-006 | — | Guest and authenticated branches; error preservation across all 4 failure classes | component/store | `npx vitest run app/__tests__/QuickStartView.spec.ts` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | Criterion 5 | — | Retried client call does not double-count a conversion | integration | `go test ./server -run TestProductEvents_AuthoritativeDedup -race` | ✅ (needs guest case) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

**Note on the expired-guest test (CONTEXT.md § Specific Ideas):** under D-2.1 the backing guest
row never expires, so criterion 3's "stop working when the backing guest expires" clause can
never fire in a live scenario. D-2.2 keeps `expires_at` enforced in authorization, so the check
is still built and still tested — but **against a hand-marked row**, not an organically expired
one. Verification must not expect to observe this rejection end-to-end.

---

## Wave 0 Requirements

- [ ] `server/guest_users_test.go` — REQ-ACT-005's create / collision / rate-limit / expiry /
      login-rejection / claim / conflict / relationship-preservation acceptance clauses
- [ ] Migration test for the new `{TS}_guest_users` pair — model on `TestMigrations_CardNameSearch`
      (`server/deck_import_test.go:150`)
- [ ] `app/__tests__/DeckImportPanel.spec.ts` — no component test exists (component not yet built)
- [ ] `app/__tests__/CommanderReview.spec.ts` — no component test exists (component not yet built)
- [ ] `app/__tests__/QuickStartView.spec.ts` — no integration test exercises a public-but-stateful
      route branch (guest vs. authenticated)
- [ ] `app/__tests__/router.spec.ts` — no router test found in `app/__tests__`; needs one asserting
      `/play` never redirects regardless of `auth.isAuthenticated`

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
