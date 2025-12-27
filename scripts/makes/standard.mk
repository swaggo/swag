# Code quality targets
.PHONY: lint
PKG ?= ./...
lint: ## Run lint (Usage: make lint PKG=./internal/services/evidencing)
	@bash -cex '\
		golangci-lint run $(PKG)\
	'

.PHONY: fmt
PKG ?= ./...
fmt: ## Run fmt (Usage: make fmt PKG=./internal/services/evidencing)
	@bash -cex '\
		golangci-lint fmt $(PKG)\
	'

# Code quality targets
.PHONY: lint-fast
lint-fast: ## Run golangci-lint
	golangci-lint run --fast-only



# Dependency management
.PHONY: update-deps
update-deps: ## Update all Go modules
	go list -m -u all | awk '{print $$1}' | xargs -n 1 go get -u
	go mod tidy

.PHONY: tidy
tidy: ## Tidy up Go modules
	@bash -c '\
		export GOPRIVATE=github.com/griffnb/core/*; \
		export GH_TOKEN=$$(gh auth token); \
		git config --global url."https://$${GH_TOKEN}@github.com/".insteadOf "https://github.com/"; \
		go mod tidy; \
	'


.PHONY: unit_test
PKG ?= ./...
RUN ?=
EXTRA ?=

unit_test: ## Run tests (Usage: make test PKG=./internal/services/evidencing RUN='TestName/Subtest')
	@bash -cex '\
		go test $(PKG) -v -count=1 -timeout=30s -race $(if $(RUN),-run "^$(RUN)$$",) $(EXTRA) \
	'