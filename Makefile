.PHONY: build clean install test fmt vet deps help

# Default target
all: build

# Build the application
build:
	@echo "Building easy-cli..."
	@go build -o easy-cli .

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@GOOS=linux GOARCH=amd64 go build -o dist/easy-cli-linux-amd64 .
	@GOOS=darwin GOARCH=amd64 go build -o dist/easy-cli-darwin-amd64 .
	@GOOS=windows GOARCH=amd64 go build -o dist/easy-cli-windows-amd64.exe .
	@echo "Built binaries in dist/"

# Install the application
install:
	@echo "Installing easy-cli..."
	@go install .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f easy-cli
	@rm -rf dist/

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Created .env file from .env.example"
	@echo "Please edit .env with your actual configuration values"

# Quick development build and run
dev: build
	@echo "Running development build..."
	@./easy-cli

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  install    - Install the application"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  deps       - Download dependencies"
	@echo "  lint       - Run linter (requires golangci-lint)"
	@echo "  dev-setup  - Set up development environment"
	@echo "  dev        - Quick development build and run"
	@echo "  help       - Show this help message"