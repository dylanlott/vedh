.PHONY: build dev persistence clean test test-api test-unit test-persistence test-frontend test-smoke-rust prepare-test-db monitoring-up monitoring-down

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=edhgo
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

test-api:
	$(GOTEST) -v ./server/... -race -count=1

test: test-unit test-persistence test-api test-frontend

test-smoke-rust:
	cargo run --locked --manifest-path tools/smoke/Cargo.toml -- --graphql-url $${VEDH_GRAPHQL_URL:-http://127.0.0.1:8080/graphql}

test-unit:
	$(GOTEST) -v ./pkg/... -race -count=1

test-persistence:
	$(GOTEST) -v ./persistence/... -race -count=1

test-frontend:
	cd app && npm test && npm run build

prepare-test-db:
	$(GOCMD) run ./tools/testdb

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

run:
	@set -eu; set -a; . ./.env; set +a; $(GOCMD) run ./

# Run the local API and Vite application together. The API configuration lives
# in the ignored root .env file; Vite proxies /graphql to the configured API port.
dev:
	@set -eu; \
	set -a; . ./.env; set +a; \
	$(GOCMD) run ./ & api_pid=$$!; \
	(cd app && npm run dev -- --host 127.0.0.1) & web_pid=$$!; \
	trap 'kill $$api_pid $$web_pid 2>/dev/null || true' EXIT INT TERM; \
	wait $$api_pid

generate:
	$(GOCMD) run github.com/99designs/gqlgen

# Migrate will run migrations at your env's DB_URL value.
# This is how we run prod migrations, so BE CAREFUL ABOUT RUNNING THIS COMMAND.
# ALWAYS TEST MIGRATIONS LOCALLY FIRST.
migrate-prod: confirm
	migrate -path ./persistence/migrations -database $(VEDH_DB_URL) up

# Run migrations directory against your local environment.
migrate-local: confirm
	migrate \
		-database $(DB_URL) \
		-source "file://./persistence/migrations/" up

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

# import-db assumes that the desired cards.csv is present in the root directory.
# refresh the CSV file here: https://mtgjson.com/downloads/all-files
import-db:
	$(GOCMD) run scripts/db_import.go

# import-allprintings assumes that AllPrintings.sql is present in the root directory.
# Download from: https://mtgjson.com/downloads/all-files
import-allprintings:
	$(GOCMD) run scripts/db_import.go -sql AllPrintings.sql -verbose

# import-csv assumes that AllPrintings.csv is present in the root directory.
import-csv:
	$(GOCMD) run scripts/db_import.go -csv cards.csv -verbose

persistence:
	docker compose --env-file .env -f dev.docker-compose.yml up -d postgres

monitoring-up:
	docker compose --env-file .env -f monitoring/docker-compose.yml up -d

monitoring-down:
	docker compose --env-file .env -f monitoring/docker-compose.yml down

confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
