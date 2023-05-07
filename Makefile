.PHONY: build
build: ## Make build
	@docker build -t zeronethunter/tg-bot .
	echo "Build successfully"

.PHONY: run
run: ## Make run
	@docker-compose up -d
	echo "Run successfully"

.PHONY: stop
stop: ## Make stop
	@docker-compose down
	echo "Stop successfully"

.PHONY: push
push: ## Make push to docker hub
	@docker push zeronethunter/tg-bot
	echo "Push successfully"

.PHONY: lint
lint: ## Make linters
	@golangci-lint run -c configs/.golangci.yaml

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
