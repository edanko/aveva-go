GO ?= go
BINARY ?= plate2dxf
CMD ?= ./cmd/plate2dxf
PKGS ?= ./...
GOLANGCI_LINT ?= golangci-lint

.DEFAULT_GOAL := help

# Self-documenting help: every target with a `## comment` is listed.
.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} \
		/^[a-zA-Z0-9_.-]+:.*?##/ { printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

##@ Build

.PHONY: build
build: ## Build the plate2dxf binary into ./bin
	$(GO) build -trimpath -o bin/$(BINARY) $(CMD)

.PHONY: install
install: ## Install plate2dxf into GOBIN
	$(GO) install $(CMD)

.PHONY: run
run: ## Run plate2dxf; pass args with ARGS="-i in -o out"
	$(GO) run $(CMD) $(ARGS)

##@ Test

.PHONY: test
test: ## Run the test suite
	$(GO) test $(PKGS)

.PHONY: test-race
test-race: ## Run tests with the race detector and shuffled order, as CI does
	$(GO) test -race -shuffle=on -count=1 $(PKGS)

.PHONY: test-cover
test-cover: ## Run tests and report total coverage
	$(GO) test -coverprofile=coverage.out -count=1 $(PKGS)
	$(GO) tool cover -func=coverage.out | tail -1

.PHONY: cover-html
cover-html: test-cover ## Open the coverage report in a browser
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "wrote coverage.html"

.PHONY: bench
bench: ## Run benchmarks
	$(GO) test -run '^$$' -bench=. -benchmem $(PKGS)

.PHONY: fuzz
fuzz: ## Fuzz the document reader; override with FUZZTIME=60s or FUZZ=FuzzReadProfileDocument
	$(GO) test -run '^$$' -fuzz=$(or $(FUZZ),FuzzReadDocument) -fuzztime=$(or $(FUZZTIME),30s) .

##@ Lint and format

.PHONY: lint
lint: ## Run golangci-lint
	$(GOLANGCI_LINT) run $(PKGS)

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with autofix
	$(GOLANGCI_LINT) run --fix $(PKGS)

.PHONY: lint-config
lint-config: ## Verify .golangci.yml is valid for the installed version
	$(GOLANGCI_LINT) config verify

.PHONY: fmt
fmt: ## Format with gofmt and goimports
	$(GOLANGCI_LINT) fmt $(PKGS)

.PHONY: fmt-check
fmt-check: ## Fail if any file is not gofmt-clean
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "not gofmt-clean:"; echo "$$out"; exit 1; fi
	@echo "gofmt clean"

.PHONY: vet
vet: ## Run go vet
	$(GO) vet $(PKGS)

##@ Dependencies and security

.PHONY: tidy
tidy: ## Tidy modules
	$(GO) mod tidy

.PHONY: tidy-check
tidy-check: ## Fail if go.mod/go.sum are not tidy (what CI enforces)
	@cp go.mod go.mod.bak && cp go.sum go.sum.bak
	@$(GO) mod tidy
	@if ! diff -q go.mod go.mod.bak >/dev/null || ! diff -q go.sum go.sum.bak >/dev/null; then \
		mv go.mod.bak go.mod; mv go.sum.bak go.sum; \
		echo "go.mod/go.sum are not tidy; run 'make tidy'"; exit 1; fi
	@rm -f go.mod.bak go.sum.bak
	@echo "modules tidy"

.PHONY: verify
verify: ## Verify module checksums and download dependencies
	$(GO) mod verify
	$(GO) mod download

.PHONY: vuln
vuln: ## Scan for known vulnerabilities (govulncheck)
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest $(PKGS)

##@ Aggregate

.PHONY: ci
ci: verify tidy-check build vet lint fmt-check test-race ## Run everything CI runs

.PHONY: check
check: build vet fmt-check test ## Fast local gate

.PHONY: clean
clean: ## Remove build and coverage artifacts
	rm -rf bin coverage.out coverage.html
