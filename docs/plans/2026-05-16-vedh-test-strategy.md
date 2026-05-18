# vEDH Smoke and E2E Test Strategy Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a fast, repeatable smoke path plus one real browser E2E flow for vEDH without depending on the fragile full local backend test harness.

**Architecture:** Keep the current Go and Vitest tests as the lower layers, add a cheap API smoke script that exercises signup/create/join against a running GraphQL endpoint, then add a minimal Playwright browser flow against the frontend preview or staging app. This gives us one CI-friendly smoke gate and one true user-path E2E test without trying to solve the whole fixture-heavy backend test problem first.

**Tech Stack:** Go tests, Vitest, Vue 3 + Vite, GraphQL, Playwright, optional local Postgres for legacy server tests

---

## Current repo state

### Current pyramid
- **Unit-ish frontend tests**
  - `app/__tests__/commanderPartner.spec.ts`
  - `app/__tests__/formatRegistry.spec.ts`
  - `app/__tests__/stackUtils.spec.ts`
- **Frontend integration tests hitting a real backend**
  - `app/__tests__/FormCreateGame.integration.spec.ts`
  - `app/__tests__/JoinGame.integration.spec.ts`
- **Frontend component coverage gap**
  - `app/__tests__/Card.spec.ts` is skipped because SFC mount setup is not ready
- **Backend tests**
  - `server/*.go` tests cover resolvers, authz, HTTP server config, metrics, origins, search, etc.
  - `pkg/games/games_test.go`
  - `persistence/redis_test.go`
- **Cheap smoke path that already exists**
  - `server/graphql_metrics_test.go`
  - `server/graphql_origin_test.go`
  - run via the documented `go test` command in `README.md`

### Current blockers / weaknesses
- No dedicated browser E2E runner or `e2e` script in `app/package.json`
- Existing frontend integration tests are not hermetic; they require a live backend at `http://localhost:8080/graphql`
- Full backend suite is heavy:
  - `server/main_test.go` expects local Postgres
  - imports `../All Printings.json` unless `ALL_PRINTINGS_JSON_PATH` is set
- No single command today that means “prove the product basically works end-to-end”

### Recommendation
1. **First:** add a fast GraphQL smoke command that creates users, creates a game, and joins a game against a configurable endpoint.
2. **Second:** add one Playwright happy-path browser E2E for “signup/create game/join game”.
3. **Later:** decide whether to keep or retire the current Vitest integration specs once the Playwright path is stable.

---

### Task 1: Document the current test layers and the new commands

**Files:**
- Modify: `README.md`
- Modify: `app/package.json`
- Test: n/a

**Step 1: Add explicit testing sections to `README.md`**

Document four layers:
- unit/frontend helpers
- backend integration tests
- API smoke
- browser E2E

**Step 2: Add placeholder npm scripts in `app/package.json`**

Add scripts we will fill in during later tasks:

```json
{
  "scripts": {
    "test:smoke": "node ./scripts/smoke-create-join.mjs",
    "test:e2e": "playwright test",
    "test:e2e:headed": "playwright test --headed"
  }
}
```

**Step 3: Verify package JSON stays valid**

Run: `cd app && npm run`
Expected: new script names appear

**Step 4: Commit**

```bash
git add README.md app/package.json
git commit -m "docs: define vedh smoke and e2e test layers"
```

---

### Task 2: Add a standalone GraphQL smoke test runner

**Files:**
- Create: `app/scripts/smoke-create-join.mjs`
- Modify: `app/package.json`
- Test: run script directly

**Step 1: Write the failing smoke runner**

Create a script that reads:
- `VEDH_GRAPHQL_URL` (default `http://127.0.0.1:8080/graphql`)
- optional `VEDH_SMOKE_TIMEOUT_MS`

It should:
- sign up user A
- create a game as user A
- sign up user B
- join the same game as user B
- exit nonzero on any failed assertion

**Step 2: Run it to verify it fails cleanly when the server is absent**

Run:

```bash
cd app
node ./scripts/smoke-create-join.mjs
```

Expected: nonzero exit with a clear connection error

**Step 3: Write minimal implementation using `fetch`**

Use raw GraphQL POSTs instead of Apollo/browser globals so the script is CI-friendly and has no DOM dependency.

Example structure:

```js
const signupMutation = `mutation Signup($username: String!, $password: String!) { ... }`
const createGameMutation = `mutation CreateGame($input: GameInput!) { ... }`
const joinGameMutation = `mutation JoinGame($input: JoinGameInput!) { ... }`
```

**Step 4: Run against a running local server**

Run:

```bash
cd app
VEDH_GRAPHQL_URL=http://127.0.0.1:8080/graphql node ./scripts/smoke-create-join.mjs
```

Expected: PASS output including created game id and joined player count

**Step 5: Commit**

```bash
git add app/scripts/smoke-create-join.mjs app/package.json
git commit -m "test: add vedh graphql smoke runner"
```

---

### Task 3: Make the existing frontend integration specs optional wrappers around the smoke path

**Files:**
- Modify: `app/__tests__/FormCreateGame.integration.spec.ts`
- Modify: `app/__tests__/JoinGame.integration.spec.ts`
- Test: `cd app && npm test`

**Step 1: Decide whether to keep or demote the current specs**

Preferred minimal path:
- keep them
- gate them behind an env var like `VEDH_RUN_LIVE_INTEGRATION=1`
- otherwise skip them with a clear reason

**Step 2: Implement the gating**

Use `describe.runIf(...)` or early `it.skip(...)` so local unit runs do not accidentally hit a backend.

**Step 3: Run Vitest**

Run:

```bash
cd app
npm test
```

Expected: unit tests run green; live integration specs skip unless explicitly enabled

**Step 4: Commit**

```bash
git add app/__tests__/FormCreateGame.integration.spec.ts app/__tests__/JoinGame.integration.spec.ts
git commit -m "test: gate live frontend integration specs"
```

---

### Task 4: Add Playwright as the first real browser E2E layer

**Files:**
- Modify: `app/package.json`
- Create: `app/playwright.config.ts`
- Create: `app/e2e/create-and-join-game.spec.ts`
- Create: `app/e2e/helpers/testData.ts`
- Test: `cd app && npm run test:e2e`

**Step 1: Add the failing Playwright setup**

Install and wire:
- `@playwright/test`

Create config that reads:
- `VEDH_APP_BASE_URL` (default `http://127.0.0.1:5173`)

**Step 2: Write one failing happy-path spec**

Flow:
- open app
- sign up or log in as user A
- create game
- open a second browser context as user B
- join the same game
- assert both players appear

**Step 3: Add the smallest needed test selectors**

If selectors are unstable, add `data-testid` only where needed in:
- `app/src/components/games/FormCreateGame.vue`
- related join-game UI file(s)

**Step 4: Run against local preview or staging**

Run:

```bash
cd app
VEDH_APP_BASE_URL=http://127.0.0.1:5173 npm run test:e2e
```

Expected: one passing browser flow

**Step 5: Commit**

```bash
git add app/package.json app/playwright.config.ts app/e2e app/src/components/games/FormCreateGame.vue
git commit -m "test: add first vedh browser e2e flow"
```

---

### Task 5: Add a CI-friendly smoke target before broadening coverage

**Files:**
- Modify: `README.md`
- Create: `.github/workflows/vedh-smoke.yml` (or repo-equivalent CI config if GitHub Actions is not used)
- Test: dry-run commands locally

**Step 1: Add a smoke-first CI job**

The first job should only require:
- app install
- a reachable GraphQL endpoint, or
- a booted local stack in CI if affordable

**Step 2: Run the cheap gates in order**

Suggested order:
1. frontend unit tests
2. Go listener/origin/metrics smoke
3. GraphQL smoke script
4. browser E2E only if environment is available

**Step 3: Document the contract**

In `README.md`, define:
- what each command proves
- what infra it needs
- which command is the release gate

**Step 4: Commit**

```bash
git add README.md .github/workflows/vedh-smoke.yml
git commit -m "ci: add vedh smoke test lane"
```

---

### Task 6: Revisit the heavy backend harness later, not first

**Files:**
- Modify: `server/main_test.go`
- Modify: `server/test.go`
- Test: `make test-api`

**Step 1: Make prerequisites explicit, not magical**

If we return to this layer, improve:
- fixture path configuration
- Postgres bootstrap docs
- skip/error messaging

**Step 2: Only then consider reducing DB/card-fixture cost**

Possible follow-up improvements:
- dedicated minimal card fixture
- dockerized test DB helper
- split pure resolver tests from card-import/search integration tests

**Step 3: Commit**

```bash
git add server/main_test.go server/test.go README.md
git commit -m "test: clarify backend harness prerequisites"
```

---

## Fastest path summary

If the goal is **"one command that proves vedh basically works"**, do this in order:

1. keep the existing cheap Go smoke as-is
2. add `app/scripts/smoke-create-join.mjs`
3. add one Playwright happy-path spec
4. gate old live Vitest integration specs so they stop surprising people

That gets vEDH from **“some integration tests exist”** to **“we have a real smoke command and one browser E2E flow”** with the smallest possible implementation surface.
