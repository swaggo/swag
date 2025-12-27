.PHONY: pr
pr: ## Create a pull request (Usage: make pr title='Your PR title')
	@if [ -z "$(title)" ]; then \
		echo "Usage: make pr title='Fixes this bug'"; \
		exit 1; \
	fi; \
	current_branch=$$(git rev-parse --abbrev-ref HEAD); \
	gh pr create --base development --head $$current_branch --title "$(title)" --body "$(title)"

.PHONY: hotfix
hotfix: ## Create hotfix PRs for main and development (Usage: make hotfix title='Your hotfix title')
	@if [ -z "$(title)" ]; then \
		echo "Usage: make hotfix title='Fixes this bug'"; \
		exit 1; \
	fi; \
	current_branch=$$(git rev-parse --abbrev-ref HEAD); \
	gh pr create --base main --head $$current_branch --title "HOTFIX PROD: $(title)" --body "$(title)"; \
	gh pr create --base development --head $$current_branch --title "HOTFIX STAGE: $(title)" --body "$(title)"



.PHONY: gh-login
gh-login: ## Login to GitHub CLI
	gh auth login
