test:
	@cd server && go test \
		-shuffle=on \
		-count=1 \
		-short \
		-timeout=5m \
		./... \
		-coverprofile=coverage.out
.PHONY: test

build-server:
	@mkdir -p bin
	@cd server && go build -o ../bin/server ./cmd/server
.PHONY: build-server

build-data-sync:
	@mkdir -p bin
	@cd server && go build -o ../bin/data-sync ./cmd/data-sync
.PHONY: build-data-sync

build:
	@mkdir -p bin
	@cd server && go build -o ../bin/server ./cmd/server
	@cd server && go build -o ../bin/data-sync ./cmd/data-sync
.PHONY: build

run-server:
	@cd server && go run ./cmd/server
.PHONY: run-server

run-data-sync:
	@cd server && go run ./cmd/data-sync
.PHONY: run-data-sync

start-docker:
	docker compose down -v && docker compose up
.PHONY: start-docker

stop-docker:
	docker compose down -v
.PHONY: stop-docker

integration-test:
	@chmod +x scripts/integration-test.sh
	@ENV=local ./scripts/integration-test.sh
.PHONY: integration-test

integration-test-dev:
	@chmod +x scripts/integration-test.sh
	@ENV=dev ./scripts/integration-test.sh
.PHONY: integration-test-dev

integration-test-pre:
	@chmod +x scripts/integration-test.sh
	@ENV=pre ./scripts/integration-test.sh
.PHONY: integration-test-pre

integration-test-pro:
	@chmod +x scripts/integration-test.sh
	@ENV=pro ./scripts/integration-test.sh
.PHONY: integration-test-pro

# lint uses the same linter as CI and tries to report the same results running
# locally. There is a chance that CI detects linter errors that are not found
# locally, but it should be rare.
lint:
	@cd server && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@cd server && golangci-lint run --config ../.golangci.yaml
.PHONY: lint

diff-check:
	@FINDINGS="$$(git status -s -uall)" ; \
		if [ -n "$${FINDINGS}" ]; then \
			echo "Changed files:\n\n" ; \
			echo "$${FINDINGS}\n\n" ; \
			echo "Diffs:\n\n" ; \
			git diff ; \
			git diff --cached ; \
			exit 1 ; \
		fi
.PHONY: diff-check

test-coverage:
	@cd server && go tool cover -func=./coverage.out
.PHONY: test-coverage

test-redis-integration:
	@cd server && go test -tags=integration ./internal/cache/... -v
.PHONY: test-redis-integration