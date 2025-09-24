
# Build all by default, even if it's not first
.DEFAULT_GOAL := all

.PHONY: all
all: tidy gen add-copyright format lint cover build

# ==============================================================================
# Build options

ROOT_PACKAGE=github.com/strayca7/siam
VERSION_PACKAGE=github.com/strayca7/component-base/pkg/version

# ==============================================================================
# Includes

include scripts/make-rules/common.mk
include scripts/make-rules/golang.mk
include scripts/make-rules/tools.mk

# Targets

## test: Run unit test.
.PHONY: test
test:
	@$(MAKE) go.test

## format: Gofmt (reformat) package sources (exclude vendor dir if existed).
.PHONY: format
format: tools.verify.golines tools.verify.goimports tidy
	@echo "===========> Formating codes"
	@$(FIND) -type f -name '*.go' | $(XARGS) gofmt -s -w
	@$(FIND) -type f -name '*.go' | $(XARGS) goimports -w -local $(ROOT_PACKAGE)
	@$(FIND) -type f -name '*.go' | $(XARGS) golines -w --max-len=120 --reformat-tags --shorten-comments --ignore-generated .
	@$(GO) mod edit -fmt

## tools: install dependent tools.
.PHONY: tools
tools:
	@$(MAKE) tools.install

.PHONY: tidy
tidy:
	@$(GOIMPORTS) -w .
	@$(GO) mod tidy
