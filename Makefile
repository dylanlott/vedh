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
# Migrate will run migrations at your env's EDHGO_PG_URL value.
# This is how we run prod migrations, so BE CAREFUL ABOUT RUNNING THIS COMMAND.
# ALWAYS TEST MIGRATIONS LOCALLY FIRST.
migrate-prod: confirm
	migrate -path ./persistence/migrations -database $(EDHGO_PG_URL) up
docker: docker-ui docker-server
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
docker-ui:
	docker buildx build \
		-f ./frontend/Dockerfile \
		-t openmtg/edhgo-ui:$(BUILD_TAG) \
		--build-arg NODE_ENV=production \
		--build-arg VUE_APP_WEBSOCKET_URL=wss://api.edhgo.com/graphql \
		--build-arg VUE_APP_BASE_URL=https://api.edhgo.com/graphql \
		--platform linux/amd64 ./frontend
docker-server:
	docker build -t openmtg/edhgo-server:$(BUILD_TAG) .
sync:
	@echo "syncing git files to remote server"
	scripts/sync.sh
deploy: confirm sync deploy-server deploy-ui
	@echo "----- edhgo deployed 🚀 -----"
deploy-ui: confirm docker-ui
	docker push openmtg/edhgo-ui:$(BUILD_TAG)
deploy-server: confirm docker-server
	docker push openmtg/edhgo-server:$(BUILD_TAG)
persistence:
	docker-compose -f dev.docker-compose.yml up -d postgres
confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
