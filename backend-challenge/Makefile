# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=main
BINARY_PATH=./cmd/api
DOCKER_IMAGE=github.com/ilyulev/kart-challenge/backend-api

.PHONY: help build run test test-coverage clean docker-build docker-run integration-test

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	$(GOBUILD) -o $(BINARY_NAME) -v $(BINARY_PATH)

run: ## Run the application
	$(GOCMD) run $(BINARY_PATH)/main.go

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

integration-test: ## Run integration tests
	./scripts/integration-test.sh

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -f deployments/Dockerfile -t $(DOCKER_IMAGE) .

docker-run: docker-build ## Run Docker container
	docker run -p 8080:8080 $(DOCKER_IMAGE)

docker-compose-up: ## Run with docker-compose
	docker-compose -f deployments/docker-compose.yml up --build

docker-compose-down: ## Stop docker-compose
	docker-compose -f deployments/docker-compose.yml down

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

fmt: ## Format code
	$(GOCMD) fmt ./...

vet: ## Run go vet
	$(GOCMD) vet ./...

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	golangci-lint run

all: fmt vet test build ## Run fmt, vet, test, and build

dev: ## Run in development mode with hot reload (requires air)
	air -c .air.toml