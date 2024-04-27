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

.PHONY: up
up: ## run all applications in stack
	docker compose build
	docker compose up -d

.PHONY: unit-tests
unit-tests: ## run unit tests without e2e-tests directory.
	go test -race -count=1 `go list ./... | grep -v e2e-tests`

.PHONY: unit-tests-ci
unit-tests-ci: ## run unit tests without e2e-tests directory (multiple times to find race conditions).
	go test -race -count=50 -failfast `go list ./... | grep -v e2e-tests`

.PHONY: ci-e2e
ci-e2e: up
	go run ./e2e-tests/scripts/wait-ready/main.go -addr=':80;:8081;:8082'
	@$(MAKE) tests-e2e

.PHONY: tests-e2e
tests-e2e: ## run end to end tests
    ## There is some race condition when running tests as go test -count=1 ./tests/... Come back at some point and fix it
	go test ./e2e-tests/browser_extension/... -count=1
	go test ./e2e-tests/icons/... -count=1
	go test ./e2e-tests/mobile/... -count=1
	go test ./e2e-tests/support/... -count=1
	go test ./e2e-tests/system/... -count=1
	go test ./e2e-tests/pass/... -count=1

vendor-licenses: ## report vendor licenses
	go-licenses report ./cmd/api --template licenses.tpl > licenses.json 2> licenses-errors
