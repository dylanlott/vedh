# vEDH Productionization Checklist Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Close the highest-leverage vEDH production-readiness gaps identified in the repo audit: sensitive payload logging, metrics exposure defaults, stale deploy docs, and missing cheap verification guidance.

**Architecture:** Keep the first pass small and reversible. Harden runtime defaults in the Go server, remove log leakage instead of building a logging abstraction, and make the README match the deploy/testing reality that exists in this repo today. Prefer targeted tests that can run without the full Postgres/card-import harness when possible.

**Tech Stack:** Go, gqlgen GraphQL server, Prometheus, Dokku deploy flow, Markdown docs.

---

### Task 1: Remove sensitive game payload logging

**Files:**
- Modify: `vedh/server/games.go`
- Inspect: `vedh/security_best_practices_report.md`

**Step 1: Confirm the leaking log site**

Look at:
- `vedh/server/games.go`

Current risky line shape:

```go
s.logger.Debug("found game in database", "game_id", gameID, "payload", string(payload))
```

**Step 2: Replace it with non-sensitive structured logging**

Target shape:

```go
s.logger.Debug("found game in database", "game_id", gameID, "payload_bytes", len(payload))
```

**Step 3: Verify there are no other obvious raw payload logs in the server package**

Run:

```bash
grep -RIn 'payload", string(payload)' /root/.openclaw/workspace/vedh/server
```

Expected: no matches.

**Step 4: Commit**

```bash
git add vedh/server/games.go
git commit -m "fix: stop logging raw game payloads"
```

---

### Task 2: Make metrics exposure safe by default when enabled

**Files:**
- Modify: `vedh/server/graphql.go`
- Modify: `vedh/server/graphql_metrics_test.go`
- Modify: `vedh/README.md`

**Step 1: Write/adjust failing expectations**

Update the metrics tests so that:
- metrics are exposed only when `METRICS_ENABLED=true` **and** `METRICS_TOKEN` is non-empty
- if enabled without a token, exposure is denied by default

Add expectation shape like:

```go
func TestGraphQLServer_ShouldExposeMetrics_RequiresTokenWhenEnabled(t *testing.T) {
    s := &graphQLServer{cfg: Conf{MetricsEnabled: true, MetricsToken: ""}}
    if s.shouldExposeMetrics() {
        t.Fatalf("expected metrics to stay disabled without token")
    }
}
```

**Step 2: Run the targeted tests and confirm the old behavior fails**

Run:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test $(find server -maxdepth 1 -name '*.go' ! -name '*_test.go' | sort) server/graphql_metrics_test.go server/graphql_origin_test.go -run 'TestGraphQLServer_|TestParseAllowedOrigins' -v
```

Expected: failure on the new metrics exposure expectation.

**Step 3: Implement the minimal hardening**

In `vedh/server/graphql.go`, change the logic so that:

```go
func (s *graphQLServer) shouldExposeMetrics() bool {
    return s != nil && s.cfg.MetricsEnabled && strings.TrimSpace(s.cfg.MetricsToken) != ""
}
```

Also log a warning when `METRICS_ENABLED=true` but `METRICS_TOKEN` is empty.

**Step 4: Re-run the targeted tests**

Run the same command as Step 2.
Expected: PASS.

**Step 5: Commit**

```bash
git add vedh/server/graphql.go vedh/server/graphql_metrics_test.go vedh/README.md
git commit -m "fix: require auth token for metrics exposure"
```

---

### Task 3: Make deploy and runtime docs match reality

**Files:**
- Modify: `vedh/README.md`
- Inspect: `vedh/Makefile`
- Inspect: `vedh/main.go`
- Inspect: `vedh/.vedh.env`

**Step 1: Remove stale deploy claims**

Delete or rewrite outdated references to:
- `make deploy`
- `make deploy-ui`
- `make deploy-server`
- legacy Watchtower/Docker Hub language that no longer describes the primary deploy path

**Step 2: Replace them with the actual current commands**

Document the real split deploy paths:

```bash
# server
git push dokku main

# frontend
git subtree push --prefix app dokku-app main
```

Also note current runtime env knobs that actually exist in code:
- `DATABASE_URL`
- `PORT`
- `ALLOWED_ORIGINS`
- `METRICS_ENABLED`
- `METRICS_TOKEN`
- `LOG_LEVEL`

**Step 3: Add a clear production note for metrics**

Document that `/prometheus` only comes up when:
- `METRICS_ENABLED=true`
- `METRICS_TOKEN` is set

And recommend ingress/network restriction even then.

**Step 4: Inspect the README diff directly**

Run:

```bash
git diff -- vedh/README.md
```

Expected: stale make/deploy claims removed, current env/deploy guidance added.

**Step 5: Commit**

```bash
git add vedh/README.md
git commit -m "docs: align vedh deploy and runtime guidance"
```

---

### Task 4: Add a lightweight server verification path to the docs

**Files:**
- Modify: `vedh/README.md`
- Inspect: `vedh/server/main_test.go`

**Step 1: Capture the current test constraint explicitly**

Document that full `./server` tests currently depend on:
- local Postgres
- the card import fixture path (`../All Printings.json` by default)

**Step 2: Document the cheap targeted verification path that avoids `TestMain`**

Add a temporary targeted verification command:

```bash
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test $(find server -maxdepth 1 -name '*.go' ! -name '*_test.go' | sort) server/graphql_metrics_test.go server/graphql_origin_test.go -run 'TestGraphQLServer_|TestParseAllowedOrigins' -v
```

Document that this is the current fast smoke path for listener/origin/metrics hardening checks.

**Step 3: Inspect the README diff**

Run:

```bash
git diff -- vedh/README.md
```

Expected: cheap smoke path and full-test prerequisites are both clear.

**Step 4: Commit**

```bash
git add vedh/README.md
git commit -m "docs: add lightweight server smoke verification path"
```

---

### Task 5: Plan the auth-storage migration without blocking the above fixes

**Files:**
- Inspect: `vedh/app/src/stores/auth.ts`
- Optional Create: `vedh/docs/plans/2026-05-15-vedh-auth-storage-migration.md`

**Step 1: Confirm the current token storage behavior**

Check that `localStorage` still stores the bearer token in:
- `vedh/app/src/stores/auth.ts`

**Step 2: Write down the migration target**

Preferred next-state:
- `HttpOnly` cookie session, or
- short-lived in-memory token with refresh flow

**Step 3: Keep this as a follow-on task**

Do not block today’s productionization pass on the full auth migration.

**Step 4: Commit only if a separate doc is created**

```bash
git add vedh/docs/plans/2026-05-15-vedh-auth-storage-migration.md
git commit -m "docs: plan vedh auth storage migration"
```
