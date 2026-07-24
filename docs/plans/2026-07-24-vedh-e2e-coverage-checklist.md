# vEDH E2E Coverage Checklist Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Expand vEDH’s true end-to-end coverage from one happy-path browser test plus API smoke into a small, trustworthy suite that covers auth, score rendering, negative join behavior, CI execution, and a safe policy for live/prod smoke.

**Architecture:** Keep the current split between browser E2E (Playwright) and API smoke (Rust) rather than replacing it. Add a few high-value Playwright scenarios around the existing Vue 3 app routes, wire Playwright into CI with an ephemeral local stack, and document a production-safe smoke policy so future live checks do not casually mutate `vedh.xyz`.

**Tech Stack:** Vue 3, Vite, Playwright, Vitest, Go 1.24, GraphQL, Rust smoke CLI, GitHub Actions.

---

## Checklist

- [x] 1. Add login + auth redirect Playwright coverage
- [x] 2. Add score view Playwright coverage
- [x] 3. Add bad-join / not-found Playwright coverage
- [x] 4. Run Playwright in CI with a local ephemeral stack
- [x] 5. Document and enforce a production-safe smoke policy

### Task 1: Login + auth redirect Playwright coverage

**Files:**
- Modify: `vedh/app/e2e/create-and-join-game.spec.ts`
- Or create: `vedh/app/e2e/auth-redirect.spec.ts`
- Reference: `vedh/app/src/router/index.ts`
- Reference: `vedh/app/src/views/LoginView.vue`
- Reference: `vedh/app/src/views/GamesView.vue`

**Step 1: Write the failing browser test**
- Add a Playwright scenario that visits `/games` while logged out.
- Assert the user is redirected to `/login?redirect=%2Fgames`.
- Sign up or log in with a fresh test user.
- Assert the app redirects back to `/games`.

**Step 2: Run only the new test**

```bash
cd /root/.openclaw/workspace/vedh
npm --prefix app run test:e2e -- --grep "auth redirect"
```

**Expected:** one failing test at first if the redirect flow is not yet asserted correctly.

**Step 3: Implement the minimal fix if needed**
- Only touch router/login flow if the test exposes a real bug.
- Prefer fixing route/query handling over broad auth-store churn.

**Step 4: Re-run the focused Playwright test**

```bash
cd /root/.openclaw/workspace/vedh
npm --prefix app run test:e2e -- --grep "auth redirect"
```

**Expected:** PASS.

**Step 5: Commit**

```bash
git add app/e2e app/src/router/index.ts app/src/views/LoginView.vue app/src/views/GamesView.vue
git commit -m "test: cover vedh auth redirect flow"
```

### Task 2: Score view Playwright coverage

**Files:**
- Modify: `vedh/app/e2e/create-and-join-game.spec.ts`
- Or create: `vedh/app/e2e/score-view.spec.ts`
- Reference: `vedh/app/src/views/ScoreView.vue`
- Reference: `vedh/app/src/views/BoardView.vue`

**Step 1: Write the failing browser test**
- Reuse the existing create/join fixture flow to create a real game.
- Navigate to `/games/:id/score`.
- Assert both player names render.
- Assert life totals are visible.
- Assert the battlefield/hand/graveyard chips render for each player.

**Step 2: Run only the score-view test**

```bash
cd /root/.openclaw/workspace/vedh
npm --prefix app run test:e2e -- --grep "score view"
```

**Expected:** FAIL first if selectors or route state handling are incomplete.

**Step 3: Implement the minimal fix if needed**
- Touch `ScoreView.vue` only if the test exposes a real route-loading or rendering bug.
- Do not redesign the page during this task.

**Step 4: Re-run the focused Playwright test**

```bash
cd /root/.openclaw/workspace/vedh
npm --prefix app run test:e2e -- --grep "score view"
```

**Expected:** PASS.

**Step 5: Commit**

```bash
git add app/e2e app/src/views/ScoreView.vue
git commit -m "test: cover vedh score view flow"
```

### Task 3: Bad-join / not-found Playwright coverage

**Files:**
- Create: `vedh/app/e2e/join-error.spec.ts`
- Reference: `vedh/app/src/views/JoinGameView.vue`
- Reference: `vedh/app/src/views/GameDoesNotExistView.vue`
- Reference: `vedh/app/src/router/index.ts`

**Step 1: Write the failing browser test**
- Visit `/join/not-a-real-game-id` as an authenticated user.
- Assert the app shows the correct error or not-found state instead of silently hanging.
- Add one direct route check for `/games/404` or unmatched route behavior.

**Step 2: Run only the join-error test**

```bash
cd /root/.openclaw/workspace/vedh
npm --prefix app run test:e2e -- --grep "bad join"
```

**Expected:** FAIL first if the UI lacks a stable error state.

**Step 3: Implement the minimal fix if needed**
- Keep the fix narrow: route fallback, empty state copy, or explicit not-found handling.

**Step 4: Re-run the focused Playwright test**

```bash
cd /root/.openclaw/workspace/vedh
npm --prefix app run test:e2e -- --grep "bad join"
```

**Expected:** PASS.

**Step 5: Commit**

```bash
git add app/e2e app/src/views/JoinGameView.vue app/src/views/GameDoesNotExistView.vue app/src/router/index.ts
git commit -m "test: cover vedh bad join flow"
```

### Task 4: Run Playwright in CI with a local ephemeral stack

**Files:**
- Create or modify: `vedh/.github/workflows/playwright.yml`
- Or modify: `vedh/.github/workflows/test.yml`
- Reference: `vedh/.github/workflows/smoke-rust.yml`
- Reference: `vedh/app/playwright.config.ts`
- Reference: `vedh/dev.docker-compose.yml`
- Reference: `vedh/app/package.json`

**Step 1: Write the workflow in the smallest useful shape**
- Check out code.
- Set up Node.
- Install frontend deps with `npm --prefix app ci`.
- Start the local app/API stack needed for Playwright.
- Wait for the target URL to become ready.
- Run `npm --prefix app run test:e2e`.

**Step 2: Validate workflow syntax locally as far as practical**

```bash
cd /root/.openclaw/workspace/vedh
sed -n '1,240p' .github/workflows/playwright.yml
```

**Expected:** file is readable and mirrors the existing Rust smoke workflow style.

**Step 3: Run the same startup/test commands locally**

```bash
cd /root/.openclaw/workspace/vedh
npm --prefix app run test:e2e -- --list
```

**Expected:** suite enumerates cleanly and the workflow commands are grounded in real repo scripts.

**Step 4: Commit**

```bash
git add .github/workflows app/playwright.config.ts app/package.json dev.docker-compose.yml
git commit -m "ci: run vedh playwright smoke"
```

### Task 5: Document and enforce a production-safe smoke policy

**Files:**
- Modify: `vedh/README.md`
- Modify or create: `vedh/docs/plans/2026-05-16-vedh-test-strategy.md`
- Optional: `vedh/app/scripts/smoke-create-join.mjs`
- Optional: `vedh/tools/smoke/src/main.rs`

**Step 1: Document the policy**
- State clearly that current smoke/E2E flows create real users and games.
- Mark them as local/staging-safe by default.
- Require explicit env targeting for any live/prod run.
- Document cleanup expectations or the lack thereof.

**Step 2: Add light enforcement if appropriate**
- If safe and low-risk, require an explicit opt-in env var before allowing known live/prod hosts.
- Keep this minimal; documentation alone is acceptable if code enforcement would slow down the other tasks.

**Step 3: Verify docs/scripts stay aligned**

```bash
cd /root/.openclaw/workspace/vedh
grep -RIn "production-safe\|live/prod\|VEDH_GRAPHQL_URL\|VEDH_APP_BASE_URL" README.md docs app/scripts tools/smoke | sed -n '1,240p'
```

**Expected:** docs and smoke entrypoints describe the same policy.

**Step 4: Commit**

```bash
git add README.md docs app/scripts tools/smoke
git commit -m "docs: define vedh smoke safety policy"
```

## Done definition

This checklist is done when all of the following are true:
- Playwright has coverage for auth redirect, score view, and bad-join/not-found behavior.
- Browser E2E is wired into GitHub Actions.
- The Rust/API smoke path still exists and remains the lightweight API-level proof.
- The README/test-strategy docs explain what is actually covered and how to run it safely.
- The checklist above is fully checked off.
