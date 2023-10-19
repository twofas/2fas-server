.PHONY: help

USERID=$(shell id -u)

help:
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| sort \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


migration: ## create database migrations file
	docker compose run -u ${USERID}:${USERID} --rm api migrate create new_migration sql


migration-up: ## apply all available migrations
	docker compose run -u ${USERID}:${USERID} --rm api migrate up


up: ## run all applications in stack
	docker compose build
	docker compose up -d


test: ## run unit tests
	go test ./internal/...


tests-e2e: ## run end to end tests
    ## There is some race condition when running tests as go test -count=1 ./tests/... Come back at some point and fix it
	go test ./tests/browser_extension/... -count=1
	go test ./tests/icons/... -count=1
	go test ./tests/mobile/... -count=1
	go test ./tests/support/... -count=1
	go test ./tests/system/... -count=1


vendor-licenses: ## report vendor licenses
	go-licenses report ./cmd/api --template licenses.tpl > licenses.json 2> licenses-errors