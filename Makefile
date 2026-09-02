# Copyright 2026 Harness Inc. All rights reserved.
# Use of this source code is governed by the PolyForm Free Trial 1.0.0 license
# that can be found in the licenses directory at the root of this repository.

.PHONY: test test-short help

GOTEST ?= go test

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

test: ## Run all tests with race detection
	$(GOTEST) -v -race -count=1 ./...

test-short: ## Run short tests with race detection
	$(GOTEST) -v -race -short -count=1 ./...
