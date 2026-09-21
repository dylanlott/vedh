# Deck-to-game activation release

## Release contract

The release is eligible for a quiet beta only when the logged-out host, public
invitee, account claim, provider fallback, and existing authenticated flow all
pass against the same candidate commit. PostgreSQL `product_events` is the
funnel source of truth; Prometheus and Grafana are the technical-health layer.

Keep the two production Dokku applications (`vedh-api` and `app`) out of the
staging procedure. Use isolated applications and an isolated database:

- API: `vedh-api-staging`
- web: `vedh-app-staging`
- PostgreSQL: `vedh-api-staging-db`

If those targets do not exist, staging is **not configured**. Do not reinterpret
a successful local run as a staging run and do not link staging to
`vedh-api-db`.

## Candidate preflight

Record the candidate SHA and require a clean worktree before any push:

```sh
git rev-parse HEAD
git status --short
```

Run all local gates with PostgreSQL available at `DATABASE_URL`:

```sh
go test ./pkg/... -race -count=1
go test ./persistence/... -race -count=1
go test ./server/... -race -count=1

cd app
npm test
npm run build
npm run test:e2e
cd ..

cargo test --locked --manifest-path tools/smoke/Cargo.toml
node --test monitoring/grafana/dashboards/*.test.js
psql -v ON_ERROR_STOP=1 "$DATABASE_URL" \
  -f docs/analytics/deck-to-game-activation-fixture-test.sql
```

CI repeats these gates with deterministic card fixtures and without reaching a
live deck provider. The browser run must finish by executing
`tools/testdb/verify_activation_events.sql`; a nonzero result is a failed gate,
even when the user journey itself rendered successfully.

## One-time staging provisioning

These commands mutate Dokku and require operator authorization. Supply a fresh
staging-only JWT secret through the normal secret-management path; never copy
the production value into shell history or documentation.

```sh
ssh dokku@192.241.142.53 apps:create vedh-api-staging
ssh dokku@192.241.142.53 apps:create vedh-app-staging
ssh dokku@192.241.142.53 postgres:create vedh-api-staging-db
ssh dokku@192.241.142.53 postgres:link vedh-api-staging-db vedh-api-staging
```

Set the API configuration:

- `JWT_SECRET`: unique, at least 32 characters
- `ALLOWED_ORIGINS`: the exact staging web origin
- `GUEST_CREATION_ENABLED=true`
- `DECK_PROVIDER_ENABLED=false`
- `METRICS_ENABLED=true` only when a staging scrape token and target exist
- `PRODUCT_EVENT_RATE_PER_MINUTE=120`
- `PRODUCT_EVENT_RATE_BURST=60`

Configure the frontend build to use the staging API, not production:

```sh
ssh dokku@192.241.142.53 docker-options:add vedh-app-staging build \
  '--build-arg VITE_GRAPHQL_HTTP --build-arg VITE_GRAPHQL_WS'
```

Set `VITE_GRAPHQL_HTTP=https://<staging-api>/graphql` and
`VITE_GRAPHQL_WS=wss://<staging-api>/graphql` on `vedh-app-staging`. Confirm the
staging domains and TLS before admitting any tester.

Seed only deterministic public card fixtures in the isolated staging database.
Never clone production users, games, or product events into staging.

## Deploy and smoke

Deploy the API from the candidate commit. Deploy the web application from the
same commit's `app/` subtree so Dokku sees `app/Dockerfile` at repository root.
Record both deployed SHAs in the release note.

Run the following from two isolated browser contexts:

1. Logged-out host opens `/play`, pastes the deterministic Commander fixture,
   chooses the detected commander, enters a display name, and reaches a board.
2. Copy/share produces `/join/<game-id>`.
3. Logged-out invitee opens that URL, sees only the safe invite projection,
   imports the second fixture, joins, and reaches the same board.
4. Both browsers show both players.
5. Host claims the guest identity, refreshes, and retains the same game.
6. A simulated provider failure keeps the pasted deck and offers paste
   recovery. Do not enable or contact a live provider for this proof.
7. A separate authenticated signup/create/join run remains green.

Run the Rust smoke against staging:

```sh
cargo run --locked --manifest-path tools/smoke/Cargo.toml -- \
  --graphql-url https://<staging-api>/graphql
```

Query the run's unique `e2e-<run-id>-*` sessions with
`tools/testdb/verify_activation_events.sql`. The result must be `0`: every
expected client and server-authoritative event occurred exactly once.

## Go / hold / rollback

**Go** only when every local/CI/staging gate is green, the deployed SHA matches
the candidate, event verification is exact, and no product-event drops or
create/join/readiness regression is present.

**Hold** when staging is absent, domains/TLS are unverified, the provider terms
review is incomplete, the event verifier is nonzero, or the quiet-beta cohort
has not reached its minimum denominators.

**Paste-only mitigation:**

```sh
ssh dokku@192.241.142.53 config:set vedh-api-staging DECK_PROVIDER_ENABLED=false
```

**Stop new guest creation:**

```sh
ssh dokku@192.241.142.53 config:set vedh-api-staging GUEST_CREATION_ENABLED=false
```

**Application rollback:** redeploy the previously recorded API and web SHAs to
their matching staging applications. Migrations in this release are additive;
do not run a down migration as an incident response. Existing authenticated
play remains the recovery path if guest creation is disabled.

After a switch or rollback, repeat the two-user smoke and event verification,
then annotate the affected UTC interval before continuing the cohort described
in `docs/runbooks/deck-to-game-quiet-beta.md`.
