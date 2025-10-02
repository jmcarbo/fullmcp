.PHONY: test test-all test-http test-streamhttp test-compliance test-comparison clean coverage bench

# Run all stable tests (excluding concurrent and slow tests)
test:
	@echo "Running stable integration tests..."
	go test -v -run 'Test[^C]' -timeout 15s

# Run all tests including potentially unstable ones
test-all:
	@echo "Running all integration tests..."
	go test -v -timeout 30s

# Run HTTP transport tests only
test-http:
	@echo "Running HTTP transport tests..."
	go test -v -run TestHTTPTransport

# Run Streamable HTTP tests only
test-streamhttp:
	@echo "Running Streamable HTTP tests..."
	go test -v -run TestStreamable

# Run compliance tests only
test-compliance:
	@echo "Running MCP compliance tests..."
	go test -v -run TestCompliance -timeout 10s

# Run transport comparison tests
test-comparison:
	@echo "Running transport comparison tests..."
	go test -v -run TestTransport -timeout 10s

# Run with coverage
coverage:
	@echo "Generating coverage report..."
	go test -cover -coverprofile=coverage.out -timeout 15s -run 'Test[^C]'
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -run=^$$

# Clean test artifacts
clean:
	rm -f coverage.out coverage.html
	rm -f *.test

# Help
help:
	@echo "Available targets:"
	@echo "  test            - Run stable tests (default)"
	@echo "  test-all        - Run all tests including unstable"
	@echo "  test-http       - Run HTTP transport tests"
	@echo "  test-streamhttp - Run Streamable HTTP tests"
	@echo "  test-compliance - Run MCP spec compliance tests"
	@echo "  test-comparison - Run transport comparison tests"
	@echo "  coverage        - Generate coverage report"
	@echo "  bench           - Run benchmarks"
	@echo "  clean           - Clean test artifacts"
