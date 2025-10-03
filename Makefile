.PHONY: test build lint clean coverage examples all test-all test-verbose mcpclient build-examples format

# Default target
all: format lint test build

# Run tests (excluding examples which have no test files)
test:
	@echo "Running tests..."
	@go list ./... | grep -v examples | xargs go test

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	@go list ./... | grep -v examples | xargs go test -v

# Run ALL tests including examples (disable vet for example files with format directives)
test-all:
	@echo "Running all tests including examples..."
	@go test -vet=off ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go list ./... | grep -v examples | xargs go test -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | tail -1

# Build all packages
build:
	@echo "Building all packages..."
	@go build ./...

# Build examples (verification only, no output)
examples:
	@echo "Building examples..."
	@for dir in examples/*/; do \
		if [ -f "$${dir}main.go" ]; then \
			echo "Building $${dir}..."; \
			go build -o /dev/null "./$${dir}" || exit 1; \
		fi \
	done
	@echo "All examples built successfully!"

# Build examples to bin folder
build-examples:
	@echo "Building examples to bin/..."
	@mkdir -p bin
	@for dir in examples/*/; do \
		if [ -f "$${dir}main.go" ]; then \
			name=$$(basename "$${dir}"); \
			echo "Building $${name}..."; \
			go build -o "bin/$${name}" "./$${dir}" || exit 1; \
		fi \
	done
	@echo "All examples built to bin/"

# Build mcpcli client
mcpclient:
	@echo "Building mcpcli..."
	@mkdir -p bin
	@go build -o bin/mcpcli ./cmd/mcpcli
	@echo "mcpcli built successfully in bin/mcpcli"

# Format code with gofumpt
format:
	@echo "Formatting code with gofumpt..."
	@gofumpt -l -w .

# Run linter
lint: format
	@echo "Running linter..."
	@golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@go clean ./...
	@rm -f coverage.out coverage_core.out coverage.html
	@rm -rf bin
	@find . -name "*.test" -delete

# Quick check (lint + test + examples)
check: lint test examples

# Help target
help:
	@echo "Available targets:"
	@echo "  all           - Run lint, test, and build (default)"
	@echo "  test          - Run tests (excluding examples)"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-all      - Run all tests including examples"
	@echo "  coverage      - Run tests with coverage report"
	@echo "  build         - Build all packages"
	@echo "  examples      - Build all example programs (verification)"
	@echo "  build-examples- Build all examples to bin/ folder"
	@echo "  mcpclient     - Build mcpcli binary"
	@echo "  lint          - Run golangci-lint"
	@echo "  check         - Run lint, test, and build examples"
	@echo "  clean         - Remove build artifacts"
	@echo "  help          - Show this help message"
