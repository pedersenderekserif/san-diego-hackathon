.PHONY: help build start run test fmt tidy clean

APP_NAME := api-server
BIN_DIR := bin
BIN_PATH := $(BIN_DIR)/$(APP_NAME)

help:
	@echo "Available targets:"
	@echo "  make build  - Build the API binary"
	@echo "  make start  - Start the API server (go run)"
	@echo "  make run    - Run the built binary"
	@echo "  make test   - Run all tests"
	@echo "  make fmt    - Format Go source files"
	@echo "  make tidy   - Tidy Go module dependencies"
	@echo "  make clean  - Remove build artifacts"

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_PATH) ./cmd/server

start:
	go run ./cmd/server

run: build
	./$(BIN_PATH)

test:
	go test ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)