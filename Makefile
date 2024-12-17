# Project metadata
APP_NAME := andmerada
BUILD_DIR := bin
MAIN_FILE := cmd/$(APP_NAME)/main.go

EXECUTABLE := $(BUILD_DIR)/$(APP_NAME)


.PHONY: init all run build clean fmt test lint


all: lint test build


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
