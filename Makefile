BINARY_NAME=flyby
BUILD_DIR=build
SOURCE=cmd/flyby/main.go

.PHONY: build run clean test help install

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SOURCE)
	@echo "Built $(BINARY_NAME) in $(BUILD_DIR)/"

# Run the application
run:
	@go run $(SOURCE)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -cover ./...

# Lint the code
lint:
	@echo "Running linter..."
	@golangci-lint run

# Format the code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(SOURCE)

# Display help
help:
	@echo "Available commands:"
	@echo "  build         - Build the binary"
	@echo "  run           - Run the application"
	@echo "  deps          - Install dependencies"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  install       - Install binary to GOPATH"
	@echo "  help          - Show this help message"