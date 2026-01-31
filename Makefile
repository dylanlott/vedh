.PHONY: build dev persistence clean test test-api deploy-ui deploy-server docker-server docker-ui docker

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=edhgo
BINARY_UNIX=$(BINARY_NAME)_unix
BUILD_TAG=latest #tag releases as latest by default

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

test-api:
	$(GOTEST) -v ./server/... -race

test: test-api

test-unit:
	$(GOTEST) -v ./pkg/... -race

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

run:
	$(GOCMD) run ./

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
	docker-compose -f dev.docker-compose.yml up -d postgres

confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
