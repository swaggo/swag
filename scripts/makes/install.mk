.PHONY: sync-core
sync-core: ## sync custom core packages with shadcn_builder
	@echo "Syncing custom core packages"
	@GITHUB_TOKEN=$$(gh auth token) shadcn_builder add -all;

.PHONY: sync-core-force
sync-core-force: ## sync custom core packages with shadcn_builder
	@echo "Syncing custom core packages"
	@GITHUB_TOKEN=$$(gh auth token) shadcn_builder add -all -force;

.PHONY: check-sync-core
check-sync-core: ## sync custom core packages with shadcn_builder
	@echo "Syncing custom core packages"
	@GITHUB_TOKEN=$$(gh auth token) shadcn_builder check -all;
	

.PHONY: install-core
install-core: ## Install private Go modules
	@bash -c '\
		export GOPRIVATE=github.com/griffnb/core/*; \
		export GH_TOKEN=$$(gh auth token); \
		go get github.com/griffnb/core/lib@latest \
	'


.PHONY: install-private
install-private: ## Install private Go modules
	@bash -c '\
		export GOPRIVATE=github.com/griffnb/core/*; \
		export GH_TOKEN=$$(gh auth token); \
		go get $(filter-out $@,$(MAKECMDGOALS)); \
	'
# Prevent Make from interpreting the args as targets
%:
	@:

.PHONY: install-codegen
install-codegen: ## Install latest code generation from go-core
	@bash -c '\
		export GOPRIVATE=github.com/griffnb/core/*; \
		export GH_TOKEN=$$(gh auth token); \
		go install github.com/griffnb/core/core_gen@latest; \
	'


.PHONY: install-lint
install-lint: ## Install golangci-lint
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	$$(go env GOPATH)/bin/golangci-lint --version


.PHONY: install-air
install-air: ## Install air
	go install github.com/air-verse/air@latest

.PHONY: install-gh
install-gh: ## Install GitHub CLI
	brew install gh
	gh auth login


.PHONY: install-deadcode
install-deadcode: ## Install deadcode
	go install golang.org/x/tools/cmd/deadcode@latest


.PHONY: stripe-install
stripe-install: ## Install Stripe CLI
	brew install stripe/stripe-cli/stripe
	stripe login