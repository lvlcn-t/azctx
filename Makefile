.DEFAULT_GOAL := help
SHELL := /bin/bash
export GOEXPERIMENT := jsonv2

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
LDFLAGS := -s -w -X main.version=$(VERSION)
ARGS ?=

.PHONY: help
help: ## Display this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "\033[36m%-20s\033[0m- %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dev
dev: build ## Build and run the CLI, you can pass subcommands using `make dev ARGS="subcommand args"`
	@./bin/azctx $(ARGS)

.PHONY: build
build: ## Build the CLI binary
	@go build -ldflags="$(LDFLAGS)" -o bin/azctx main.go

.PHONY: test
test: ## Run all tests
	@go test -race -cover -count=1 ./...

.PHONY: docs
docs: ## Generate CLI reference documentation
	@go generate ./internal/gendoc/...
