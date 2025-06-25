# Makefile for NATS Go Examples

# Variables
BINARY_NAME=nats-example
GO=go
GOFLAGS=-ldflags="-s -w"
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S')
PACKAGES=$(shell $(GO) list ./... | grep -v /vendor/)
DOCKER_COMPOSE=docker-compose

# Directories
CMD_DIR=./cmd
BIN_DIR=./bin
NATS_DOCKER_DIR=./nats-docker

# Output binaries
PUBLISHER_BINARY=$(BIN_DIR)/publisher
SUBSCRIBER_BINARY=$(BIN_DIR)/subscriber

# Default target
.PHONY: all
all: clean deps lint test build

# Setup development environment
.PHONY: setup
setup: deps
	mkdir -p $(BIN_DIR)

# Install dependencies
.PHONY: deps
deps:
	$(GO) mod tidy
	$(GO) mod download

# Build all applications
.PHONY: build
build: build-publisher build-subscriber

# Build publisher
.PHONY: build-publisher
build-publisher:
	mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(PUBLISHER_BINARY) $(CMD_DIR)/publisher

# Build subscriber
.PHONY: build-subscriber
build-subscriber:
	mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(SUBSCRIBER_BINARY) $(CMD_DIR)/subscriber

# Run tests
.PHONY: test
test:
	$(GO) test -v -race -cover $(PACKAGES)

# Run short tests (for CI)
.PHONY: test-short
test-short:
	$(GO) test -v -short -cover $(PACKAGES)

# Run linter
.PHONY: lint
lint:
	$(GO) vet $(PACKAGES)
	golangci-lint run

# Clean up
.PHONY: clean
clean:
	$(GO) clean
	rm -rf $(BIN_DIR)

# Start NATS server
.PHONY: nats-start
nats-start:
	cd $(NATS_DOCKER_DIR) && $(DOCKER_COMPOSE) up -d

# Stop NATS server
.PHONY: nats-stop
nats-stop:
	cd $(NATS_DOCKER_DIR) && $(DOCKER_COMPOSE) down

# Run publisher
.PHONY: run-publisher
run-publisher: build-publisher
	$(PUBLISHER_BINARY) -config configs/app.json

# Run subscriber
.PHONY: run-subscriber
run-subscriber: build-subscriber
	$(SUBSCRIBER_BINARY) -config configs/app.json

# Generate coverage report
.PHONY: coverage
coverage:
	$(GO) test -coverprofile=coverage.out $(PACKAGES)
	$(GO) tool cover -html=coverage.out -o coverage.html

# Build Docker image
.PHONY: docker-build
docker-build:
	docker build -t nats-go-examples:$(VERSION) .

# Show help
.PHONY: help
help:
	@echo "NATS Go Examples Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all           Build everything (default)"
	@echo "  setup         Setup development environment"
	@echo "  deps          Install dependencies"
	@echo "  build         Build all binaries"
	@echo "  test          Run tests"
	@echo "  test-short    Run short tests (for CI)"
	@echo "  lint          Run linters"
	@echo "  clean         Clean up build artifacts"
	@echo "  nats-start    Start NATS server with Docker"
	@echo "  nats-stop     Stop NATS server"
	@echo "  run-publisher Run publisher application"
	@echo "  run-subscriber Run subscriber application"
	@echo "  coverage      Generate test coverage report"
	@echo "  docker-build  Build Docker image"
	@echo "  help          Show this help message"
