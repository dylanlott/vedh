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
- Go v1.17
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
VUE_APP_WEBSOCKET_URL="ws://127.0.0.1:8080/graphql"
VUE_APP_BASE_URL="http://127.0.0.1:8080/graphql"
```

### Persistence

You can quickly start the persistence dependencies by running `make persistence`

This will boot up Postgres database.

Then run the server with our Makefile by running `make run`

The server will attempt to run all migrations and then start up.  If it can't run migrations, it will rollback the database and noisily fail.

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
- `npm test` to run the frontend unit and helper tests

## Testing

Use the smallest layer that proves the change you made:

- **Frontend unit/helper tests** (`cd app && npm test`): proves Vue components, stores, and small browser helpers behave correctly in isolation. Requires Node and frontend deps installed; no server or database.
- **Backend integration tests** (`make test-api`): proves the Go API works against its real persistence and GraphQL paths. Requires local Postgres on `localhost:5432` and any test fixture/config expected by the current Go tests.
- **API smoke** (`cd app && npm run test:smoke`): proves a create/join flow works against a running API with minimal end-to-end setup. Requires the frontend deps plus a locally running server configured for the smoke script.
- **Browser E2E** (`cd app && npm run test:e2e` or `npm run test:e2e:headed`): proves the browser experience works through real UI flows. Requires frontend deps, a running app/API target for Playwright, and browser binaries installed.

If you only need fast feedback, start with unit/helper tests. Reach for backend integration, smoke, or browser E2E when you need confidence across process boundaries.

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

## Stack

- Postgres stores application and card data.
- Golang for the server
- Migrate [CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) for managing database migrations.
- GraphQL as a realtime API layer

## Deployment (Dokku)

The current production deploys use Dokku on `dokku@192.241.142.53`.

The frontend Dokku app is `app`.

Ensure your local deploy remote targets that app:

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

Full `./server` tests currently expect:

- local Postgres on `localhost:5432`
- the card import fixture path (`../All Printings.json` by default)

For a fast listener/origin/metrics smoke check that does **not** go through `server/main_test.go`, run:

```sh
cd /root/.openclaw/workspace/vedh
/usr/local/go/bin/go test $(find server -maxdepth 1 -name '*.go' ! -name '*_test.go' | sort) server/graphql_metrics_test.go server/graphql_origin_test.go -run 'TestGraphQLServer_|TestParseAllowedOrigins' -v
```

Use full `make test-api` only when the local DB + card fixture prerequisites are available.
