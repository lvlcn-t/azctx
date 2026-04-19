.DEFAULT_GOAL := help
SHELL := /bin/bash

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
LDFLAGS := -s -w -X main.version=$(VERSION)
ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))

.PHONY: help
help: ## Display this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "\033[36m%-20s\033[0m- %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dev
dev: build ## Build and run the CLI, you can pass subcommands like: make dev use or make dev list
	@./bin/azctx $(ARGS)

%: # Hack to allow passing arguments to the dev target without Make trying to find a target for them
	@:

.PHONY: build
build: ## Build the CLI binary
	@go build -ldflags="$(LDFLAGS)" -o bin/azctx main.go
