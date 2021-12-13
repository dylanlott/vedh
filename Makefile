# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=edhgo
BINARY_UNIX=$(BINARY_NAME)_unix

all: test-api build-ui build-server
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
	$(GOCMD) generate ./...
# Migrate will run migrations at your env's DATABASE_URL value.
# This is how we run prod migrations, so BE CAREFUL ABOUT RUNNING THIS COMMAND.
# ALWAYS TEST MIGRATIONS LOCALLY FIRST.
migrate:
	migrate -path ./persistence/migrations -database $(DATABASE_URL) up
build: build-ui build-server
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
build-ui:
	docker build -f ./frontend/Dockerfile -t dylanlott/edhgo-ui:latest ./frontend
build-server:
	docker build -f ./Dockerfile -t dylanlott/edhgo-server:latest .
deploy: deploy-server deploy-ui
deploy-ui:
	docker push dylanlott/edhgo-ui:latest 
deploy-server:
	docker push dylanlott/edhgo-server:latest
