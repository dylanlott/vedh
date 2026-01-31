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
# Authentication secret
JWT_SECRET=""

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

To run the Vue app:

```sh
> cd ./frontend
> yarn install
> yarn start
```

#### V2 Frontend

The v2 frontend exists in frontend-v2/ and is in active development.

`npm run dev` to run the dev server.
`npm test` to run all the tests.

#### Vue Tests

`node@14` is required to build the frontend. Later versions are not supported.

`yarn test` executes the boardstate unit tests.

## Testing

### Prerequisites

- Postgres instance running locally `localhost@5432`

To run the API tests

```sh
> make test-api
```

Once you have a Postgres instance running locally, you can run your tests. You may need to change the config values in `games_test.go` to start your tests or configure your environment to fit with the provided.

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

### vedh-api (server)

```sh
git add $CHANGES
git commit -m "server changes"
git push dokku main
```

### frontend-v2 (Vite/Vue)

```sh
git add $CHANGES
git commit -m "frontend-v2 changes"
git subtree push --prefix frontend-v2 dokku@192.241.142.53:frontend-v2 main
```

## Documentation & Resources

- [How to connect to a postgres instance inside of docker](https://stackoverflow.com/questions/37694987/connecting-to-postgresql-in-a-docker-container-from-outside)
- [How to import an SQL dump into Postgres](https://stackoverflow.com/questions/6842393/import-sql-dump-into-postgresql-database)
- [Make sure when you rows.Scan() you don't point it at a nil value](https://stackoverflow.com/questions/44670212/scan-sql-null-values-in-golang/46753197)

## Deployment (Docker Hub, legacy)

Deployments are run using `vtec2/watchtower` to watch for container updates to Docker Hub. New builds are tested and then tagged `latest` and pushed to Docker Hub so that Watchtower detects them on the EDH-Go production server.

## Environments

`frontend/.env.local` sets local environemnt variables and is used when `yarn start` is run.
`frontend/.env.production` sets production environment variables and it used for `yarn build`.

A copy of the frontend environment file for development is included in this repository.

## Deploying with Make

`scripts/sync.sh` must be run before loading any migrations. It syncs the local git ls-files output to the production remote. This allows the server to access migration files at runtime.

- `make deploy` will deploy a new version of both the server and the UI.
- `make deploy-ui` and `make deploy-server` will deploy them each individually.
The Makefile contains a `confirm` script that requires user confirmation before running deployment targets.

After running any of the deployment targets, you'll be prompted with a yes / no before proceeding.

> Note: Your SSH key must be registered on the production server in order to deploy.
