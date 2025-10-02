// Package main demonstrates Streamable HTTP transport in MCP (2025-03-26 specification)
package main

import (
	"fmt"
)

func main() {
	fmt.Println("MCP Streamable HTTP Transport Example")
	fmt.Println("======================================")
	fmt.Println()

	// Example 1: What is Streamable HTTP?
	fmt.Println("💡 Streamable HTTP Overview")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Println("Streamable HTTP is a new transport mechanism introduced in MCP 2025-03-26")
	fmt.Println("that replaces the HTTP+SSE transport from 2024-11-05. It provides:")
	fmt.Println()
	fmt.Println("  • Bi-directional communication over HTTP")
	fmt.Println("  • POST for client-to-server messages")
	fmt.Println("  • GET for server-to-client SSE stream")
	fmt.Println("  • Session management with unique IDs")
	fmt.Println("  • Stream resumption with event IDs")
	fmt.Println("  • Chunked transfer encoding support")
	fmt.Println()

	// Example 2: How It Works
	fmt.Println("🔄 How It Works")
	fmt.Println("===============")
	fmt.Println()
	fmt.Println("Client-to-Server (POST):")
	fmt.Println("  1. Client sends POST request with JSON-RPC message")
	fmt.Println("  2. Headers: Content-Type: application/json")
	fmt.Println("  3. Headers: Mcp-Session-Id: <session-id>")
	fmt.Println("  4. Server processes and returns response or 202 Accepted")
	fmt.Println()
	fmt.Println("Server-to-Client (GET):")
	fmt.Println("  1. Client opens GET request to same endpoint")
	fmt.Println("  2. Headers: Accept: text/event-stream")
	fmt.Println("  3. Headers: Mcp-Session-Id: <session-id>")
	fmt.Println("  4. Server streams SSE events with messages")
	fmt.Println()

	// Example 3: Session Management
	fmt.Println("🔑 Session Management")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("During initialization:")
	fmt.Print(`
  POST /mcp HTTP/1.1
  Content-Type: application/json

  {"jsonrpc":"2.0","method":"initialize","id":1,...}

  HTTP/1.1 200 OK
  Mcp-Session-Id: a1b2c3d4e5f6...
  Content-Type: application/json

  {"jsonrpc":"2.0","result":{...},"id":1}
`)
	fmt.Println("Server assigns session ID in Mcp-Session-Id header")
	fmt.Println("Client includes this ID in all subsequent requests")
	fmt.Println()

	// Example 4: POST Request Examples
	fmt.Println("📤 POST Request Examples")
	fmt.Println("========================")
	fmt.Println()
	fmt.Println("Single Request:")
	fmt.Print(`
  POST /mcp HTTP/1.1
  Content-Type: application/json
  Mcp-Session-Id: a1b2c3d4e5f6...
  Accept: application/json, text/event-stream

  {"jsonrpc":"2.0","method":"tools/list","id":2}
`)
	fmt.Println("Batch Request:")
	fmt.Print(`
  POST /mcp HTTP/1.1
  Content-Type: application/json
  Mcp-Session-Id: a1b2c3d4e5f6...

  [
    {"jsonrpc":"2.0","method":"tools/list","id":2},
    {"jsonrpc":"2.0","method":"resources/list","id":3}
  ]
`)
	fmt.Println()
	fmt.Println("Notification (no response expected):")
	fmt.Print(`
  POST /mcp HTTP/1.1
  Content-Type: application/json
  Mcp-Session-Id: a1b2c3d4e5f6...

  {"jsonrpc":"2.0","method":"notifications/cancelled","params":{...}}

  HTTP/1.1 202 Accepted
`)

	// Example 5: GET SSE Stream
	fmt.Println("📥 GET SSE Stream")
	fmt.Println("=================")
	fmt.Println()
	fmt.Println("Opening SSE connection:")
	fmt.Print(`
  GET /mcp HTTP/1.1
  Accept: text/event-stream
  Cache-Control: no-cache
  Mcp-Session-Id: a1b2c3d4e5f6...

  HTTP/1.1 200 OK
  Content-Type: text/event-stream
  Cache-Control: no-cache
  Connection: keep-alive

  id: 1
  data: {"jsonrpc":"2.0","method":"sampling/createMessage",...}

  id: 2
  data: {"jsonrpc":"2.0","method":"notifications/progress",...}
`)

	// Example 6: Stream Resumption
	fmt.Println("🔄 Stream Resumption")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("If connection drops, client can resume:")
	fmt.Print(`
  GET /mcp HTTP/1.1
  Accept: text/event-stream
  Mcp-Session-Id: a1b2c3d4e5f6...
  Last-Event-ID: 42

  Server replays events after event ID 42
`)
	fmt.Println()
	fmt.Println("Benefits:")
	fmt.Println("  • No message loss on reconnection")
	fmt.Println("  • Reliable delivery over unreliable networks")
	fmt.Println("  • Better mobile/flaky connection support")
	fmt.Println()

	// Example 7: Server Implementation
	fmt.Println("🖥️  Server Implementation")
	fmt.Println("=========================")
	fmt.Println()
	fmt.Println("Basic server setup:")
	fmt.Print(`
  import "github.com/jmcarbo/fullmcp/transport/streamhttp"

  // Create handler
  handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      // Process MCP messages
  })

  // Create server
  server := streamhttp.NewServer(":8080", handler,
      streamhttp.WithAllowedOrigin("http://localhost:3000"),
  )

  // Start server
  log.Fatal(server.ListenAndServe())
`)
	fmt.Println()

	// Example 8: Client Implementation
	fmt.Println("🔌 Client Implementation")
	fmt.Println("========================")
	fmt.Println()
	fmt.Println("Basic client setup:")
	fmt.Print(`
  import "github.com/jmcarbo/fullmcp/transport/streamhttp"

  // Create transport
  transport := streamhttp.New("http://localhost:8080/mcp")

  // Connect
  conn, err := transport.Connect(ctx)
  if err != nil {
      log.Fatal(err)
  }
  defer conn.Close()

  // Session ID is automatically managed after initialization
`)
	fmt.Println()

	// Example 9: Security Considerations
	fmt.Println("🔒 Security Considerations")
	fmt.Println("==========================")
	fmt.Println()
	fmt.Println("Origin Validation:")
	fmt.Println("  • Validate Origin header on all requests")
	fmt.Println("  • Reject requests from untrusted origins")
	fmt.Println("  • Use CORS policies appropriately")
	fmt.Println()
	fmt.Println("Session IDs:")
	fmt.Println("  • Must be cryptographically secure")
	fmt.Println("  • Should be globally unique (UUID, JWT, hash)")
	fmt.Println("  • Only visible ASCII characters (0x21-0x7E)")
	fmt.Println("  • Implement session timeouts")
	fmt.Println()
	fmt.Println("Localhost Binding:")
	fmt.Println("  • Bind to localhost when possible")
	fmt.Println("  • Use authentication for public endpoints")
	fmt.Println("  • Consider rate limiting")
	fmt.Println()

	// Example 10: Advantages over HTTP+SSE
	fmt.Println("✨ Advantages over HTTP+SSE")
	fmt.Println("===========================")
	fmt.Println()

	comparisons := []struct {
		aspect  string
		httpSSE string
		stream  string
	}{
		{"Endpoints", "2 separate (POST + SSE)", "1 unified endpoint"},
		{"Session Management", "Manual correlation", "Built-in with Mcp-Session-Id"},
		{"Resumption", "Not standardized", "Last-Event-ID support"},
		{"Batching", "Limited", "Full JSON-RPC batch support"},
		{"Complexity", "Higher", "Lower (single endpoint)"},
		{"Scalability", "Good", "Better (unified state)"},
	}

	for _, c := range comparisons {
		fmt.Printf("  %-20s | %-25s | %s\n", c.aspect, c.httpSSE, c.stream)
	}
	fmt.Println()

	// Example 11: Use Cases
	fmt.Println("💼 Use Cases")
	fmt.Println("============")
	fmt.Println()

	useCases := []struct {
		title       string
		description string
	}{
		{"Web Applications", "Browser-based clients with full duplex communication"},
		{"Mobile Apps", "Resilient communication with automatic resumption"},
		{"Cloud Services", "Scalable server-to-server MCP communication"},
		{"Real-time Updates", "Server-initiated notifications and progress updates"},
		{"Long-running Operations", "Bidirectional streaming for complex workflows"},
	}

	for i, uc := range useCases {
		fmt.Printf("   %d. %s\n", i+1, uc.title)
		fmt.Printf("      %s\n", uc.description)
		fmt.Println()
	}

	// Example 12: Best Practices
	fmt.Println("📋 Best Practices")
	fmt.Println("=================")
	fmt.Println()
	fmt.Println("For Clients:")
	fmt.Println("  ✓ Store and include Mcp-Session-Id from initialization")
	fmt.Println("  ✓ Handle connection drops gracefully")
	fmt.Println("  ✓ Use Last-Event-ID for resumption")
	fmt.Println("  ✓ Implement exponential backoff for reconnection")
	fmt.Println("  ✓ Monitor connection health with keep-alives")
	fmt.Println()
	fmt.Println("For Servers:")
	fmt.Println("  ✓ Generate secure, unique session IDs")
	fmt.Println("  ✓ Implement session timeouts")
	fmt.Println("  ✓ Store events for resumption (buffer recent events)")
	fmt.Println("  ✓ Send keep-alive comments on SSE stream")
	fmt.Println("  ✓ Validate Origin header for security")
	fmt.Println("  ✓ Handle graceful session termination")
	fmt.Println()

	// Example 13: Error Handling
	fmt.Println("⚠️  Error Handling")
	fmt.Println("==================")
	fmt.Println()
	fmt.Println("HTTP Status Codes:")
	fmt.Println("  200 OK           - Request with response")
	fmt.Println("  202 Accepted     - Notification (no response)")
	fmt.Println("  400 Bad Request  - Invalid request/missing session ID")
	fmt.Println("  403 Forbidden    - Invalid origin")
	fmt.Println("  404 Not Found    - Session not found")
	fmt.Println("  500 Internal     - Server error")
	fmt.Println()
	fmt.Println("Connection Errors:")
	fmt.Println("  • Network timeout: Retry with exponential backoff")
	fmt.Println("  • Session expired: Re-initialize")
	fmt.Println("  • Stream closed: Reopen with Last-Event-ID")
	fmt.Println()

	// Example 14: Performance Characteristics
	fmt.Println("📊 Performance Characteristics")
	fmt.Println("==============================")
	fmt.Println()
	fmt.Println("Latency:")
	fmt.Println("  • POST requests: Single round-trip")
	fmt.Println("  • Server messages: Instant (SSE push)")
	fmt.Println("  • Reconnection: ~1-3 seconds typical")
	fmt.Println()
	fmt.Println("Throughput:")
	fmt.Println("  • Batch requests reduce overhead")
	fmt.Println("  • SSE streaming for high-frequency updates")
	fmt.Println("  • Connection reuse improves efficiency")
	fmt.Println()
	fmt.Println("Resource Usage:")
	fmt.Println("  • One long-lived connection per session")
	fmt.Println("  • Minimal memory for session state")
	fmt.Println("  • Efficient for many concurrent clients")
	fmt.Println()

	fmt.Println("✨ Streamable HTTP demonstration complete!")
	fmt.Println()
	fmt.Println("Key Takeaways:")
	fmt.Println("  1. Single unified endpoint for bi-directional communication")
	fmt.Println("  2. Session management built into the protocol")
	fmt.Println("  3. Stream resumption prevents message loss")
	fmt.Println("  4. More scalable than HTTP+SSE")
	fmt.Println("  5. Better security with origin validation")
	fmt.Println("  6. Ideal for web and mobile applications")
}
