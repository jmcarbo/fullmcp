.PHONY: test build lint clean coverage examples all test-all

# Default target
all: lint test build

# Run tests (excluding examples which have no test files)
test:
	@echo "Running tests..."
	@go list ./... | grep -v examples | xargs go test -v

# Run ALL tests including examples (will show build warnings)
test-all:
	@echo "Running all tests including examples..."
	@go test ./... -v

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go list ./... | grep -v examples | xargs go test -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | tail -1

# Build all packages
build:
	@echo "Building all packages..."
	@go build ./...

# Build examples
examples:
	@echo "Building examples..."
	@for dir in examples/*/; do \
		if [ -f "$${dir}main.go" ]; then \
			echo "Building $${dir}..."; \
			go build -o /dev/null "./$${dir}" || exit 1; \
		fi \
	done
	@echo "All examples built successfully!"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@go clean ./...
	@rm -f coverage.out coverage_core.out coverage.html
	@find . -name "*.test" -delete

# Quick check (lint + test + examples)
check: lint test examples

# Help target
help:
	@echo "Available targets:"
	@echo "  all         - Run lint, test, and build (default)"
	@echo "  test        - Run tests (excluding examples)"
	@echo "  test-all    - Run all tests including examples"
	@echo "  coverage    - Run tests with coverage report"
	@echo "  build       - Build all packages"
	@echo "  examples    - Build all example programs"
	@echo "  lint        - Run golangci-lint"
	@echo "  check       - Run lint, test, and build examples"
	@echo "  clean       - Remove build artifacts"
	@echo "  help        - Show this help message"
