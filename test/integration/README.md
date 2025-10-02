# MCP Integration Tests

Comprehensive integration tests for all MCP protocol transport implementations.

## Test Coverage

### Protocol Tests (`protocol_test.go`)
Tests core MCP operations across different transports:
- ✅ **Initialization**: Protocol handshake and capability negotiation
- ✅ **Tools**: List and call tools
- ✅ **Resources**: List and read resources
- ✅ **Prompts**: List and get prompts
- ✅ **Error Handling**: Invalid requests, missing entities
- ✅ **Concurrency**: Multiple simultaneous requests
- ✅ **Benchmarks**: Performance comparison

### Transport Comparison Tests (`transport_comparison_test.go`)
Compares behavior and features across transports:
- ✅ **Consistency**: Same operations work across all transports
- ✅ **HTTP Transport**: Header validation, reconnection
- ✅ **Streamable HTTP**: Session management, SSE stream resumption
- ✅ **Transport-Specific**: Features unique to each transport

### Compliance Tests (`compliance_test.go`)
Ensures MCP specification compliance:
- ✅ **JSON-RPC 2.0**: Message format, error codes
- ✅ **MCP Protocol**: Version negotiation, capability exchange
- ✅ **Content Types**: Proper MIME types and encoding
- ✅ **Session Management**: ID generation, persistence (streamable HTTP)

## Running Tests

### Run All Integration Tests
```bash
go test ./test/integration/... -v
```

### Run Specific Test Suite
```bash
# Protocol tests only
go test ./test/integration -run TestHTTPTransport -v

# Streamable HTTP tests
go test ./test/integration -run TestStreamableHTTP -v

# Comparison tests
go test ./test/integration -run TestTransportComparison -v

# Compliance tests
go test ./test/integration -run TestCompliance -v
```

### Run With Coverage
```bash
go test ./test/integration/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Benchmarks
```bash
go test ./test/integration -bench=. -benchmem
```

### Short Mode (Skip Slow Tests)
```bash
go test ./test/integration/... -short
```

## Test Transports

| Transport | Protocol | Status | Features |
|-----------|----------|--------|----------|
| **stdio** | Standard I/O | ⚠️ Partial | Process-based communication |
| **HTTP** | HTTP POST | ✅ Full | Simple request/response |
| **Streamable HTTP** | HTTP POST + SSE | ✅ Full | Bi-directional, session management |

## Test Structure

```
test/integration/
├── README.md                          # This file
├── protocol_test.go                   # Core protocol operations
├── transport_comparison_test.go       # Transport feature comparison
└── compliance_test.go                 # MCP specification compliance
```

## Writing New Tests

### Adding Protocol Tests

```go
func TestNewFeature(t *testing.T) {
    srv := createTestServer(t)
    // ... setup transport ...
    c := client.New(conn)

    testProtocolOperations(t, c)

    // Add your specific tests
}
```

### Adding Transport-Specific Tests

```go
func TestTransportSpecificFeature(t *testing.T) {
    // Test feature unique to one transport
    // e.g., session resumption for streamable HTTP
}
```

### Common Helpers

- `createTestServer(t)`: Creates a server with test tools/resources/prompts
- `testProtocolOperations(t, c)`: Runs standard protocol operation tests
- `httptest.NewServer()`: Creates test HTTP server
- `context.WithTimeout()`: Ensures tests don't hang

## CI/CD Integration

### GitHub Actions Example
```yaml
- name: Run Integration Tests
  run: |
    go test ./test/integration/... -v -race -coverprofile=coverage.txt

- name: Upload Coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.txt
```

### Pre-commit Hook
```bash
#!/bin/bash
go test ./test/integration/... -short
```

## Debugging Failed Tests

### Enable Verbose Output
```bash
go test ./test/integration -v -run TestFailingTest
```

### Run Single Test
```bash
go test ./test/integration -run TestHTTPTransport/specific_subtest -v
```

### Check Test Timeout
```bash
# Increase timeout for slow tests
go test ./test/integration -timeout 30s -v
```

## Performance Benchmarks

Expected benchmark results (on Apple M1):

```
BenchmarkHTTPTransport-8                    5000    250000 ns/op
BenchmarkStreamableHTTPTransport-8          4000    300000 ns/op
```

Streamable HTTP has ~20% overhead due to SSE connection management.

## Known Issues

### Stdio Transport
- Requires external process management
- Tests are simplified and may not cover all cases
- Mark as skipped in short mode

### SSE Stream Testing
- SSE events are async, may need `time.Sleep()` in tests
- Use buffered channels for event collection
- Test server keeps connections open

## Contributing

When adding new features:
1. ✅ Add protocol tests for core functionality
2. ✅ Add transport-specific tests if behavior differs
3. ✅ Update compliance tests if spec changes
4. ✅ Add benchmarks for performance-sensitive code
5. ✅ Update this README with new test coverage

## References

- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [JSON-RPC 2.0 Spec](https://www.jsonrpc.org/specification)
- [SSE Specification](https://html.spec.whatwg.org/multipage/server-sent-events.html)
