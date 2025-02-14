# Project metadata
APP_NAME := andmerada
BUILD_DIR := bin
MAIN_FILE := cmd/$(APP_NAME)/main.go

EXECUTABLE := $(BUILD_DIR)/$(APP_NAME)

GOLANG_BIN := $(shell go env GOPATH)/bin
GOLANGCI_LINT_VERSION := v1.62.2


.PHONY: all ci run build clean fmt test test-with-race lint check-fmt install-lint


all: lint lint-yml lint-sql lint-docker test build
ci: check-fmt lint test-with-race


install-lint-ci:
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOLANG_BIN) $(GOLANGCI_LINT_VERSION)
	@echo "golangci-lint $(GOLANGCI_LINT_VERSION) installed."


run:
	@echo "Running the application..."
	go run $(MAIN_FILE)


build: clean
	@echo "Building the application..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-s -w" -o $(EXECUTABLE) $(MAIN_FILE)
	@echo "Build complete! Executable at $(EXECUTABLE)"


fmt:
	@echo "Formatting code..."
	go fmt ./...


check-fmt:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not properly formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	else \
		echo "All files are properly formatted!"; \
	fi


lint:
	@echo "Running golangci-lint..."
	PATH=$(PATH):$(GOLANG_BIN) golangci-lint run


lint-yml:
	@echo "Running yamllint on YAML files..."
	docker run --rm -v $(PWD)/internal:/data:Z cytopia/yamllint .


lint-sql:
	@docker build -t squawk-linter -f Dockerfile.squawk .
	@docker run --rm -v $(PWD):/lint:Z squawk-linter internal/migrator/sqlres/*.sql


lint-docker:
	@docker pull docker.io/hadolint/hadolint
	@docker run --rm -i hadolint/hadolint < Dockerfile.squawk


test-with-race:
	@echo "Running tests..."
	go test -shuffle on -timeout=60s -race ./...


test:
	@echo "Running tests..."
	go test -shuffle on -timeout=60s ./...


clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@echo "Cleanup complete!"


.PHONY: postgres
postgres:
	@echo "Starting PostgreSQL in a temporary container..."
	@docker run --replace --rm -d --name andmerada-db \
	  -e POSTGRES_USER=andmerada \
	  -e POSTGRES_PASSWORD=andmerada \
	  -e POSTGRES_DB=andmerada \
	  -p 5432:5432 \
	  postgres:16

	@echo "Waiting for PostgreSQL to be ready..."

	@until docker exec andmerada-db pg_isready -h localhost -U andmerada; do \
	    sleep 3; \
	done

	@echo "✅ PostgreSQL is ready!"
	@echo "Connect using: postgres://andmerada:andmerada@localhost:5432/andmerada?sslmode=disable"
