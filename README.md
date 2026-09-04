# vedh

> A Magic: The Gathering boardstate tracker built with GraphQL, Vue, and Go.

```text
██╗   ██╗███████╗██████╗ ██╗  ██╗
██║   ██║██╔════╝██╔══██╗██║  ██║
██║   ██║█████╗  ██║  ██║███████║
╚██╗ ██╔╝██╔══╝  ██║  ██║██╔══██║
 ╚████╔╝ ███████╗██████╔╝██║  ██║
  ╚═══╝  ╚══════╝╚═════╝ ╚═╝  ╚═╝
```

> Pronounced "vee-dee-aych"
> The virtual expressive deck handler.

## Development

Prerequisites:

- Make
- Go v1.27.0
- PostgreSQL 14.15.0
- Node 16

### Server

You will need to configure `.vedh.env` and `.pg.env` environment files at your project root.

```sh
# .vedh.env example
# Database connection URI
DATABASE_URL=""
# Optional listener port
PORT="8080"
# Authentication secret
JWT_SECRET=""

# Allowed browser origins for HTTP + websocket GraphQL traffic
ALLOWED_ORIGINS="http://localhost:5173,http://127.0.0.1:5173"

# Optional Prometheus exposure. Metrics stay disabled unless both are set.
METRICS_ENABLED="false"
METRICS_TOKEN=""

# Optional: structured logging
# Levels: debug, info, warn, error
LOG_LEVEL="info"
```

```sh
# .pg.env example
POSTGRESQL_USERNAME=""
POSTGRESQL_PASSWORD=""
POSTGRESQL_DATABASE=""
```

### FrontEnd

The front end uses Node 16 and won't build with any other version.
I recommend NVM to manage the environment for it.
The front end also supports an environment file, `.env.local`.

```sh
NODE_ENV="development"
VITE_GRAPHQL_WS="ws://127.0.0.1:8080/graphql"
VITE_GRAPHQL_HTTP="http://127.0.0.1:8080/graphql"
```

### Persistence

You can quickly start the persistence dependencies by running `make persistence`

This will boot up Postgres database.

Then run the server with our Makefile by running `make run`

The server will attempt to run all migrations and then start up. If it can't run migrations, it will rollback the database and noisily fail.

You can run the server as if it's in prod with this same config, so you can switch between local and prod as long as you've configured your environment variables correctly.

A copy of the server environment file for development is included in this repository.

### Web App

The front end is a Vue & Apollo GraphQL application that is statically served in production.

To run the current web app:

```sh
> cd ./app
> npm install
> npm run dev
```

#### App Frontend

The current frontend lives in `app/`.

- `npm run dev` to run the dev server
- `npm test` to run the frontend unit and helper tests once
- `npm run test:watch` to run Vitest in watch mode

## Testing

Use the smallest layer that proves the change you made:

- **Frontend unit/helper tests** (`cd app && npm test`): proves Vue components, stores, and small browser helpers behave correctly in isolation. Runs once and exits cleanly. Requires Node and frontend deps installed; no server or database.
- **Backend integration tests** (`make test-api`): proves the Go API works against its real persistence and GraphQL paths. Requires a Postgres role with `CREATE DATABASE` permission. The harness creates and drops a unique `edhgo_test_*` scratch database and uses the committed minimal card fixture; it never resets the database named by `DATABASE_URL`.
- **API smoke (Rust, preferred)** (`make test-smoke-rust` or `cargo run --locked --manifest-path tools/smoke/Cargo.toml --`): proves a create/join flow works against a running GraphQL API with minimal end-to-end setup. Defaults to `http://127.0.0.1:8080/graphql` and respects `VEDH_GRAPHQL_URL` / `VEDH_SMOKE_TIMEOUT_MS`. For production, point it at `https://api.vedh.xyz/graphql`.
- **API smoke (legacy JS)** (`cd app && npm run test:smoke`): older create/join smoke runner against a running API. Keep only as fallback while the Rust path settles.
- **Browser E2E** (`cd app && npm run test:e2e` or `npm run test:e2e:headed`): proves the browser experience works through real UI flows. Requires frontend deps, a running app/API target for Playwright, and browser binaries installed.

CI prepares its isolated browser-test database with `make prepare-test-db`.
That target applies the committed test migrations to `DATABASE_URL`; only use it
with a disposable database.

If you only need fast feedback, start with unit/helper tests. Reach for backend integration, the Rust smoke runner, or browser E2E when you need confidence across process boundaries.

## Observability

### Logging

The server emits structured JSON logs to stdout.

- Control log verbosity with `LOG_LEVEL` (defaults to `info`).
- Each HTTP request gets an `X-Request-Id`:
  - If you send `X-Request-Id`, the server will reuse it.
  - Otherwise, the server generates one and returns it in the response.
  - Logs include `request_id`, `method`, `path`, `status`, and `duration_ms`.

```sh
LOG_LEVEL=debug make run
```

For local stack runs via `dev.docker-compose.yml`, add basic container log controls:

- `web`, `server`, and `postgres` now use the `json-file` logging driver.
- log rotation is set to `10m` per file with `3` files retained.

Read logs per service:

```sh
docker compose -f dev.docker-compose.yml logs -f server
docker compose -f dev.docker-compose.yml logs -f web
docker compose -f dev.docker-compose.yml logs -f postgres
```

## Observability stack (local)

Use this stack to validate that the API is exporting `/prometheus` and that Grafana is wired to it.

1. Ensure API metrics are enabled and token-protected in the API env.

```sh
cd /root/.openclaw/workspace/vedh
sed -i 's/^METRICS_ENABLED=.*/METRICS_ENABLED=true/' .vedh.env
sed -i 's/^METRICS_TOKEN=.*/METRICS_TOKEN=dev-metrics-token/' .vedh.env
grep -q '^METRICS_ENABLED=' .vedh.env || echo 'METRICS_ENABLED=true' >> .vedh.env
grep -q '^METRICS_TOKEN=' .vedh.env || echo 'METRICS_TOKEN=dev-metrics-token' >> .vedh.env
```

2. Configure observability env.

```sh
cd /root/.openclaw/workspace/vedh/monitoring
cp .env.observability.example .env.observability
```

Set these values in `.env.observability`:

- `PROMETHEUS_TARGET` to the API host:port your monitoring stack should scrape
  - default example: `host.docker.internal:8081` for the API in `dev.docker-compose.yml`
- `PROMETHEUS_BEARER_TOKEN` to match `METRICS_TOKEN` from `.vedh.env`
- `GRAFANA_ADMIN_USER` / `GRAFANA_ADMIN_PASSWORD` for UI login

3. Start the stack.

```sh
cd /root/.openclaw/workspace/vedh
make monitoring-up
```

If the API stack is already running, restart it after changing metrics env vars:

```sh
docker compose -f dev.docker-compose.yml up -d --force-recreate server
```

4. Verify Prometheus is running and scraping the API.

```sh
curl -sf http://localhost:9090/-/ready
curl -sf http://localhost:9090/api/v1/targets?state=any | grep -q '"job":"vedh-api"' && echo "scrape target configured"
curl -sf "http://localhost:9090/api/v1/query?query=up%7Bjob%3D%22vedh-api%22%7D" | grep -q '"value":\[' && echo "metrics query returned"
```

5. Verify Grafana has the Prometheus datasource and is connected.

```sh
export GRAFANA_ADMIN_USER=admin
export GRAFANA_ADMIN_PASSWORD=admin

curl -s -u "$GRAFANA_ADMIN_USER:$GRAFANA_ADMIN_PASSWORD" \
  http://localhost:3000/api/health | grep -q '"database": "ok"' && echo "grafana up"

curl -s -u "$GRAFANA_ADMIN_USER:$GRAFANA_ADMIN_PASSWORD" \
  "http://localhost:3000/api/datasources/name/Prometheus" | grep -q '"url":"http://prometheus:9090"' && echo "prometheus datasource configured"
```

Open `http://localhost:3000` and run an Explore query such as:

```txt
up{job="vedh-api"}
```

Stop the monitoring stack when done:

```sh
cd /root/.openclaw/workspace/vedh
make monitoring-down
```

## Stack

- Postgres stores application and card data.
- Golang for the server
- Migrate [CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) for managing database migrations.
- GraphQL as a realtime API layer

## Deployment (Dokku)

The current production deploys use Dokku on `dokku@192.241.142.53`.

Every successful `CI` run for a push to `main` triggers the `Deploy to Dokku`
workflow. That workflow deploys the tested revision to both Dokku apps and then
checks the frontend and API routes. Configure these GitHub production-environment
values before enabling automatic deployment:

- Repository variables: `DOKKU_HOST`, `DOKKU_API_APP`, `DOKKU_WEB_APP`
- Environment secrets: `DOKKU_SSH_PRIVATE_KEY`, `DOKKU_KNOWN_HOSTS`

`DOKKU_KNOWN_HOSTS` must contain a host key verified independently by the
operator; the workflow deliberately does not trust a runtime `ssh-keyscan`.

Production is split across two Dokku apps:

- `app` serves the frontend SPA on `https://vedh.xyz` (also `www.vedh.xyz` / `app.vedh.xyz`)
- `vedh-api` serves the Go GraphQL API on `https://api.vedh.xyz/graphql`

The frontend is built with these production env vars:

```sh
VITE_GRAPHQL_HTTP="https://api.vedh.xyz/graphql"
VITE_GRAPHQL_WS="wss://api.vedh.xyz/graphql"
```

Ensure your local deploy remote targets the frontend Dokku app when shipping `app/` changes:

```sh
git remote remove dokku-app 2>/dev/null || true
git remote add dokku-app dokku@192.241.142.53:app
```

### vedh-api (server)

```sh
git add $CHANGES
git commit -m "server changes"
git push dokku main
```

### app (Vite/Vue)

```sh
git add $CHANGES
git commit -m "app changes"
git subtree push --prefix app dokku-app main
```

## Documentation & Resources

- [How to connect to a postgres instance inside of docker](https://stackoverflow.com/questions/37694987/connecting-to-postgresql-in-a-docker-container-from-outside)
- [How to import an SQL dump into Postgres](https://stackoverflow.com/questions/6842393/import-sql-dump-into-postgresql-database)
- [Make sure when you rows.Scan() you don't point it at a nil value](https://stackoverflow.com/questions/44670212/scan-sql-null-values-in-golang/46753197)

## Environments

`app/.env.local` sets local environemnt variables and is used when `yarn start` is run.
`app/.env.production` sets production environment variables and it used for `yarn build`.

A copy of the frontend environment file for development is included in this repository.

## Runtime knobs that exist today

The current Go server reads these env vars in code:

- `DATABASE_URL`
- `PORT`
- `ALLOWED_ORIGINS`
- `METRICS_ENABLED`
- `METRICS_TOKEN`
- `LOG_LEVEL`

### Metrics hardening note

`/prometheus` is only exposed when **both** of these are set:

- `METRICS_ENABLED=true`
- `METRICS_TOKEN` is non-empty

Even then, prefer to keep the route behind ingress/network restriction instead of relying on bearer auth alone.

## Lightweight server smoke verification

Full `./server` tests require a reachable Postgres role that can create a
temporary database. Test migrations seed the small committed card fixture, so
the multi-hundred-megabyte `All Printings.json` file is not required.

For a fast listener/origin/metrics smoke check that does **not** go through `server/main_test.go`, run:

```sh
cd /path/to/vedh
/usr/local/go/bin/go test $(find server -maxdepth 1 -name '*.go' ! -name '*_test.go' | sort) server/graphql_metrics_test.go server/graphql_origin_test.go -run 'TestGraphQLServer_|TestParseAllowedOrigins' -v
```

Use full `make test-api` only when the local DB + card fixture prerequisites are available.

## Production route sanity notes

A few route expectations that matter when debugging prod:

- `https://vedh.xyz` is the frontend SPA, not the GraphQL API
- `https://api.vedh.xyz/graphql` is the real live GraphQL endpoint
- `GET https://api.vedh.xyz/graphql` may return a GraphQL validation error when no operation is supplied; that is expected
- `POST https://api.vedh.xyz/graphql` is the meaningful API smoke target
- `https://api.vedh.xyz/playground` should load the GraphQL playground
- there is currently no dedicated `/health` endpoint in the Go app
