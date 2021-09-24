# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=edhgo
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build
build:
		$(GOBUILD) -o $(BINARY_NAME) -v
short:
		$(GOTEST) -v -short ./...
test-api:
		$(GOTEST) -v ./server/... -race
clean:
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
		rm -f $(BINARY_UNIX)
run:
		$(GOCMD) run ./
generate:
		$(GOCMD) generate ./...
# Migrate will run migrations at your env's DATABASE_URL value.
# This is how we run prod migrations, so BE CAREFUL ABOUT RUNNING THIS COMMAND.
# ALWAYS TEST MIGRATIONS LOCALLY FIRST.
migrate:
	migrate -path ./persistence/migrations -database $(DATABASE_URL) up
# dev target requires watchexec to be installed
dev:
	watchexec $(GOCMD) run ./
# builds a heroku compatible go binary
build-heroku:
	go build -o bin/edh-go -v .
build-linux:
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
# builds the docker image for the vue app and then tags it as latest.
build-ui:
	docker build -f ./frontend/Dockerfile -t dylanlott/edhgo-ui:latest ./frontend
build-server:
	docker build -f ./Dockerfile -t dylanlott/edhgo:server .
# pushes the most recently built image up to docker hub.
# watchtower will detect the pushed container and pull it down so we should
# only push tested and vetted containers.
deploy-ui:
	docker push dylanlott/edhgo-ui