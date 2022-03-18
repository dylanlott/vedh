# EDH-Go
> A Magic: The Gathering board state tracker built with GraphQL and Go.
 [![Build Status](https://travis-ci.org/dylanlott/edh-go.svg?branch=master)](https://travis-ci.org/dylanlott/edh-go)

## Running

### Go Server

1. Prerequisites:
- Make
- Go v1.17
- Redis
- PostgreSQL

Then run the server with our Makefile.
The server will attempt to run all migrations and then start up. 
If it can't run migrations, it will rollback the database and noisily fail. 

`make run`

You can quickly start the persistence dependencies by running

`$ make persistence`

This will boot up Postgres and Redis development servers.

You can run the server as if it's in prod with this same config, so you 
can switch between local and prod as long as you've configured your environment
variables correctly.

### Front End 

Run Vue app:

```
$ cd frontend
$ npm start
```

#### Vue Tests

`npm run test` will run the boardstate unit tests.

## Testing 

### Prerequisites 
- Postgres instance running locally `localhost@5432`
- Redis instance running at `localhost:6379`

To run the API tests
```
$ make test-api
```

Once you have a Redis and Postgres instance running locally, you can run your 
tests. You may need to change the config values in `games_test.go` to start 
your tests or configure your environment to fit with the provided 

## Stack
Postgres stores application and card data.
Redis persists games and boardstates.

## Documentation & Resources:

- [How to connect to a postgres instance inside of docker](https://stackoverflow.com/questions/37694987/connecting-to-postgresql-in-a-docker-container-from-outside)
- [How to import an SQL dump into Postgres](https://stackoverflow.com/questions/6842393/import-sql-dump-into-postgresql-database)
- [Make sure when you rows.Scan() you don't point it at a nil value](https://stackoverflow.com/questions/44670212/scan-sql-null-values-in-golang/46753197)

# Deployment 

We run our deployments through docker-compose using `vtec2/watchtower`.
This means that we must always tag our containers with `latest` so that watchtower detects them.

## Environments

`frontend/.env.local` sets local environemnt variables and is used when `npm start` is run.
`frontend/.env.production` sets production environment variables and it used for `npm run build`.

## Deploying with Make 

`make deploy` will deploy a new version of both the server and the UI.
`make deploy-ui` and `make deploy-server` will deploy them each individually.

We have a nifty `confirm` script that requires user confirmation before running deployment targets.
After running any of the deployment targets, you'll be prompted with a yes / no before proceeding.

> Note: Your SSH key must be registered on the production server in order to deploy.


