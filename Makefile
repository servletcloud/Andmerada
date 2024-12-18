# Project metadata
APP_NAME := andmerada
BUILD_DIR := bin
MAIN_FILE := cmd/$(APP_NAME)/main.go

EXECUTABLE := $(BUILD_DIR)/$(APP_NAME)

GOLANG_BIN := $(shell go env GOPATH)/bin
GOLANGCI_LINT_VERSION := v1.62.2


.PHONY: all ci run build clean fmt test lint check-fmt install-lint


all: lint test build
ci: check-fmt lint test


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
	go build -o $(EXECUTABLE) $(MAIN_FILE)
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


test:
	@echo "Running tests..."
	go test -shuffle on -timeout=30s -race ./...


clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@echo "Cleanup complete!"
