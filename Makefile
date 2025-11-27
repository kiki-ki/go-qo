.DEFAULT_GOAL := help

BINARY_NAME := bin/qo
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/kiki-ki/go-qo/cmd.version=$(VERSION)"

.PHONY: build
build: ## Build the binary
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/qo

.PHONY: test
test: ## Run all tests
	go test ./... -v

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: check
check: tidy fmt lint ## Tidy, format and lint code

.PHONY: clean
clean: ## Remove build artifacts
	rm -f $(BINARY_NAME) coverage.out coverage.html

.PHONY: help
help: ## list up commands
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
