# Project metadata
APP_NAME := andmerada
BUILD_DIR := bin
MAIN_FILE := cmd/$(APP_NAME)/main.go

EXECUTABLE := $(BUILD_DIR)/$(APP_NAME)


.PHONY: all ci run build clean fmt test lint check-fmt


all: lint test build
ci: check-fmt lint test


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
	PATH=$(PATH):$(shell go env GOPATH)/bin golangci-lint run


test:
	@echo "Running tests..."
	go test -shuffle on -timeout=30s -race ./...


clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@echo "Cleanup complete!"
