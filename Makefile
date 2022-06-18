.PHONY: build dev persistence clean test test-api deploy-ui deploy-server docker-server docker-ui docker

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=edhgo
BINARY_UNIX=$(BINARY_NAME)_unix
BUILD_TAG=latest #tag all releases as latest for watchtower detection

all: test build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
short:
	$(GOTEST) -v -short ./...
test-api:
	$(GOTEST) -v ./server/... -race
test: test-api
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run:
	$(GOCMD) run ./
generate:
	$(GOCMD) run github.com/99designs/gqlgen
# Migrate will run migrations at your env's EDHGO_PG_URL value.
# This is how we run prod migrations, so BE CAREFUL ABOUT RUNNING THIS COMMAND.
# ALWAYS TEST MIGRATIONS LOCALLY FIRST.
migrate-prod: confirm
	migrate -path ./persistence/migrations -database $(EDHGO_PG_URL) up
docker: docker-ui docker-server
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
docker-ui:
	docker build -f ./frontend/Dockerfile -t openmtg/edhgo-ui:$(BUILD_TAG) ./frontend
docker-server:
	docker build -t openmtg/edhgo-server:$(BUILD_TAG) .
deploy: confirm deploy-server deploy-ui
deploy-ui: confirm docker-ui
	docker push openmtg/edhgo-ui:$(BUILD_TAG)
deploy-server: confirm docker-server
	docker push openmtg/edhgo-server:$(BUILD_TAG)
persistence:
	docker-compose -f dev.docker-compose.yml up postgres redis
confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]