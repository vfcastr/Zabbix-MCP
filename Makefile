SHELL := /usr/bin/env bash -euo pipefail -c

BINARY_NAME ?= ./bin/zabbix-mcp-server
BASENAME := $(shell basename $(BINARY_NAME))
VERSION ?= $(if $(shell printenv VERSION),$(shell printenv VERSION),dev)

GO=go
DOCKER=docker

DOCKER_REGISTRY ?= docker.io
IMAGE_NAME = $(DOCKER_REGISTRY)/$(BASENAME):$(VERSION)

TARGET_DIR ?= $(CURDIR)/dist

# Build flags
LDFLAGS=-ldflags="-s -w -X github.com/vfcastr/Zabbix-MCP/version.Version=$(VERSION) -X github.com/vfcastr/Zabbix-MCP/version.GitCommit=$(shell git rev-parse HEAD 2>/dev/null || echo 'unknown') -X github.com/vfcastr/Zabbix-MCP/version.BuildDate=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')"

.PHONY: all build test clean deps docker-build run-http run-stdio help

# Default target
all: build

# Build the binary
# Get local ARCH; on Intel Mac, 'uname -m' returns x86_64 which we turn into amd64.
# Always use CGO_ENABLED=0 to ensure a statically linked binary is built
ARCH     = $(shell A=$$(uname -m); [ $$A = x86_64 ] && A=amd64; echo $$A)
OS       = $(shell uname | tr [[:upper:]] [[:lower:]])
build:
	CGO_ENABLED=0 GOARCH=$(ARCH) GOOS=$(OS) $(GO) build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/$(BASENAME)

# Build for Windows
build-windows:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows $(GO) build $(LDFLAGS) -o $(BINARY_NAME).exe ./cmd/$(BASENAME)

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux $(GO) build $(LDFLAGS) -o $(BINARY_NAME)-linux ./cmd/$(BASENAME)

# Run tests
test:
	$(GO) test -v ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe $(BINARY_NAME)-linux
	$(GO) clean

# Download dependencies
deps:
	$(GO) mod download

# Tidy dependencies
tidy:
	$(GO) mod tidy

# Build docker image
docker-build:
	$(DOCKER) build --build-arg VERSION=$(VERSION) -t $(IMAGE_NAME) .

# Push docker image
docker-push: docker-build
	$(DOCKER) push $(IMAGE_NAME)

# Run stdio server locally
run-stdio: build
	./$(BINARY_NAME) stdio

# Run HTTP server locally
run-http: build
	./$(BINARY_NAME) streamable-http --transport-port 8080

# Run HTTP server with security settings
run-http-secure: build
	MCP_ALLOWED_ORIGINS="http://localhost:3000,https://example.com" MCP_CORS_MODE="development" $(BINARY_NAME) streamable-http --transport-port 8080 --transport-host 0.0.0.0

# Run HTTP server in Docker
docker-run-http: docker-build
	$(DOCKER) run -p 8080:8080 --rm -e TRANSPORT_MODE=http $(IMAGE_NAME)

# Test HTTP endpoint
test-http:
	@echo "Testing StreamableHTTP server health endpoint..."
	@curl -f http://localhost:8080/health || echo "Health check failed - make sure server is running with 'make run-http'"
	@echo "StreamableHTTP MCP endpoint available at: http://localhost:8080/mcp"

# Show help
help:
	@echo "Available targets:"
	@echo "  all              - Build the binary (default)"
	@echo "  build            - Build the binary for current OS"
	@echo "  build-windows    - Build the binary for Windows"
	@echo "  build-linux      - Build the binary for Linux"
	@echo "  test             - Run all tests"
	@echo "  clean            - Remove build artifacts"
	@echo "  deps             - Download dependencies"
	@echo "  tidy             - Tidy go.mod dependencies"
	@echo "  docker-build     - Build docker image"
	@echo "  docker-push      - Push docker image to registry"
	@echo "  run-stdio        - Run stdio server locally"
	@echo "  run-http         - Run StreamableHTTP server locally on port 8080"
	@echo "  run-http-secure  - Run StreamableHTTP server with security settings"
	@echo "  docker-run-http  - Run StreamableHTTP server in Docker on port 8080"
	@echo "  test-http        - Test StreamableHTTP health endpoint"
	@echo "  help             - Show this help message"
