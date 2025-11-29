.DEFAULT_GOAL := help

BINARY_NAME := bin/qo
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/kiki-ki/go-qo/cmd.version=$(VERSION)"

.PHONY: build
build: ## build the binary
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/qo

.PHONY: test
test: ## run all tests
	go test ./... -v

.PHONY: test-coverage
test-coverage: ## run tests with coverage
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

.PHONY: fmt
fmt: ## format code
	go fmt ./...

.PHONY: lint
lint: ## run linter
	go tool golangci-lint run

.PHONY: tidy
tidy: ## run go mod tidy
	go mod tidy

.PHONY: check
check: tidy fmt lint ## tidy, format and lint code

.PHONY: clean
clean: ## remove build artifacts
	rm -f $(BINARY_NAME) coverage.out coverage.html

.PHONY: install-tool
install-tool: ## install development tools
	go install tool

.PHONY: releaser-check
releaser-check: ## run goreleaser check
	go tool goreleaser check

.PHONY: doc
doc: ## open godoc in browser
	go tool godoc & (sleep 1 && open http://localhost:6060/pkg/github.com/kiki-ki/go-qo/)

.PHONY: help
help: ## list up commands
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
