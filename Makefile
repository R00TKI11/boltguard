.PHONY: build test clean install lint fmt run help

BINARY_NAME=boltguard
VERSION?=0.1.0
BUILD_DIR=./bin
GO=go

# detect OS for platform-specific builds
ifeq ($(OS),Windows_NT)
	BINARY_EXT=.exe
else
	BINARY_EXT=
endif

help: ## show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GO) build -ldflags="-X 'main.version=$(VERSION)'" -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) ./cmd/boltguard
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)"

install: ## install to $GOPATH/bin
	$(GO) install ./cmd/boltguard

test: ## run tests
	$(GO) test -v ./...

test-coverage: ## run tests with coverage
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint: ## run linters
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, install from https://golangci-lint.run" && exit 1)
	golangci-lint run

fmt: ## format code
	$(GO) fmt ./...
	@which goimports > /dev/null && goimports -w . || true

clean: ## remove build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

run: build ## build and run with example
	@echo "Running against alpine:latest..."
	$(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) alpine:latest

# cross-compile for common platforms
build-linux: ## build for linux amd64
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-X 'main.version=$(VERSION)'" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/boltguard

build-darwin: ## build for macOS amd64
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="-X 'main.version=$(VERSION)'" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/boltguard

build-windows: ## build for windows amd64
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="-X 'main.version=$(VERSION)'" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/boltguard

build-all: build-linux build-darwin build-windows ## build for all platforms

deps: ## download dependencies
	$(GO) mod download
	$(GO) mod tidy

vendor: ## vendor dependencies
	$(GO) mod vendor
