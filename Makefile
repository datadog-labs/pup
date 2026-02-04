# Makefile for Pup - Datadog API CLI Wrapper
# Uses goreleaser for repeatable builds

.PHONY: help build build-all build-snapshot test test-verbose lint clean install deps check-goreleaser version

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME := pup

# Get version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -s -w \
	-X github.com/DataDog/pup/internal/version.Version=$(VERSION) \
	-X github.com/DataDog/pup/internal/version.Commit=$(COMMIT) \
	-X github.com/DataDog/pup/internal/version.Date=$(DATE) \
	-X github.com/DataDog/pup/internal/version.BuiltBy=makefile

# Go build flags
GO_BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

help: ## Show this help message
	@echo "Pup - Datadog API CLI Wrapper"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: check-goreleaser ## Build binary for current platform using goreleaser (fast, local development)
	@echo "Building $(BINARY_NAME) for current platform..."
	@goreleaser build --snapshot --clean --single-target
	@echo "Binary available at: dist/pup_*/$(BINARY_NAME)"

build-all: check-goreleaser ## Build binaries for all platforms using goreleaser
	@echo "Building $(BINARY_NAME) for all platforms..."
	@goreleaser build --snapshot --clean
	@echo "Binaries available in dist/ directory"

build-snapshot: check-goreleaser ## Create a full snapshot release (archives, checksums, SBOMs) without publishing
	@echo "Creating snapshot release..."
	@goreleaser release --snapshot --clean --skip=publish,sign
	@echo "Release artifacts available in dist/ directory"

build-quick: ## Quick build using go build (no goreleaser, fastest for development)
	@echo "Building $(BINARY_NAME) with go build..."
	@CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -o $(BINARY_NAME) .
	@echo "Binary available at: ./$(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	@go test ./...

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	@go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linters (requires golangci-lint)
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it from: https://golangci-lint.run/"; \
		exit 1; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf dist/
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

install: build-quick ## Install binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install $(GO_BUILD_FLAGS) .
	@echo "Installed to: $$(go env GOPATH)/bin/$(BINARY_NAME)"

deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify
	@echo "Dependencies ready"

tidy: ## Tidy go.mod and go.sum
	@echo "Tidying dependencies..."
	@go mod tidy

check-goreleaser: ## Check if goreleaser is installed
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Error: goreleaser is not installed"; \
		echo ""; \
		echo "Install goreleaser:"; \
		echo "  brew install goreleaser/tap/goreleaser  # macOS"; \
		echo "  go install github.com/goreleaser/goreleaser@latest"; \
		echo ""; \
		echo "Or visit: https://goreleaser.com/install/"; \
		exit 1; \
	fi

version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

run: build-quick ## Build and run the binary
	@./$(BINARY_NAME)

dev: ## Build and install for development, then show version
	@$(MAKE) install
	@$(BINARY_NAME) version 2>/dev/null || echo "Note: 'pup version' command not yet implemented"

validate: ## Run all validation checks (test, lint, build)
	@echo "Running validation checks..."
	@$(MAKE) test
	@$(MAKE) lint
	@$(MAKE) build-quick
	@echo "✓ All validation checks passed"

release-test: check-goreleaser ## Test release process without publishing (requires clean git state)
	@echo "Testing release process..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: working directory is not clean. Commit or stash changes first."; \
		exit 1; \
	fi
	@goreleaser release --snapshot --clean --skip=publish
	@echo "✓ Release test successful. Artifacts in dist/"

.PHONY: help build build-all build-snapshot build-quick test test-verbose test-coverage lint fmt clean install deps tidy check-goreleaser version run dev validate release-test
