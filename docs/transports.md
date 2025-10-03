# Transports

Transports handle communication between MCP clients and servers. FullMCP supports multiple transport mechanisms with a unified interface.

## Table of Contents

- [Overview](#overview)
- [stdio Transport](#stdio-transport)
- [HTTP Transport](#http-transport)
- [WebSocket Transport](#websocket-transport)
- [SSE Transport](#sse-transport)
- [Custom Transports](#custom-transports)

## Overview

All transports implement the `io.ReadWriteCloser` interface, providing a consistent API across different communication mechanisms.

```go
type ReadWriteCloser interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    Close() error
}
```

**Supported Transports:**
- **stdio**: Standard input/output (local processes)
- **HTTP**: RESTful HTTP API
- **WebSocket**: Full-duplex real-time communication
- **SSE**: Server-Sent Events for streaming

## stdio Transport

Default transport for local process communication, perfect for CLI tools and testing.

### Server

```go
import (
    "context"
    "github.com/jmcarbo/fullmcp/server"
)

func main() {
    srv := server.New("my-server")

    // Add tools, resources, prompts...

    // Run on stdio (default)
    if err := srv.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### Client

```go
import (
    "github.com/jmcarbo/fullmcp/client"
    "github.com/jmcarbo/fullmcp/transport/stdio"
)

func main() {
    // Create stdio transport
    transport := stdio.New()

    // Create client
    c := client.New(transport)

    ctx := context.Background()
    if err := c.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    // Use client...
}
```

### Use Cases

- Local CLI tools
- Development and testing
- Process-to-process communication
- Integration with shell scripts

### Example: CLI Tool

```bash
# Server as a standalone binary
./mcp-server

# Client communicating via stdio
./mcp-client | ./mcp-server
```

## HTTP Transport

RESTful HTTP transport with support for authentication and middleware.

### Server

```go
import (
    "github.com/jmcarbo/fullmcp/server"
    "github.com/jmcarbo/fullmcp/transport/http"
)

func main() {
    // Create MCP server
    srv := server.New("http-server")

    // Add MCP entities...

    // Create HTTP server
    httpServer := http.NewServer(":8080", srv,
        http.WithReadTimeout(30*time.Second),
        http.WithWriteTimeout(30*time.Second),
        http.WithTLS("cert.pem", "key.pem"),
    )

    log.Printf("HTTP server listening on :8080")
    if err := httpServer.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
```

### Client

```go
import (
    "github.com/jmcarbo/fullmcp/client"
    "github.com/jmcarbo/fullmcp/transport/http"
)

func main() {
    // Create HTTP transport
    transport := http.New("http://localhost:8080",
        http.WithTimeout(30*time.Second),
        http.WithRetry(3),
    )

    // Create client
    c := client.New(transport)

    ctx := context.Background()
    if err := c.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    // Use client...
}
```

### With Authentication

#### Server

```go
import "github.com/jmcarbo/fullmcp/auth/apikey"

func main() {
    srv := server.New("secure-server")

    // Create auth provider
    authProvider := apikey.New()
    authProvider.AddKey("secret-key-123", auth.Claims{
        Subject: "user-1",
        Scopes:  []string{"read", "write"},
    })

    // HTTP server with auth middleware
    httpServer := http.NewServer(":8080", srv,
        http.WithMiddleware(authProvider.Middleware()),
    )

    log.Fatal(httpServer.ListenAndServe())
}
```

#### Client

```go
transport := http.New("http://localhost:8080",
    http.WithAPIKey("secret-key-123"),
)

c := client.New(transport)
```

### HTTPS/TLS

```go
// Server with TLS
httpServer := http.NewServer(":8443", srv,
    http.WithTLS("server.crt", "server.key"),
)

// Client with custom TLS config
tlsConfig := &tls.Config{
    InsecureSkipVerify: false,
    MinVersion:         tls.VersionTLS12,
}

transport := http.New("https://localhost:8443",
    http.WithTLSConfig(tlsConfig),
)
```

### Endpoints

The HTTP transport exposes these endpoints:

```
POST /mcp         - Main MCP message endpoint
GET  /health      - Health check
GET  /info        - Server info
```

### Custom HTTP Configuration

```go
httpServer := http.NewServer(":8080", srv,
    http.WithReadTimeout(30*time.Second),
    http.WithWriteTimeout(30*time.Second),
    http.WithIdleTimeout(120*time.Second),
    http.WithMaxHeaderBytes(1 << 20), // 1 MB
    http.WithCORS([]string{"https://example.com"}),
)
```

## WebSocket Transport

Full-duplex, real-time communication over WebSocket.

### Server

```go
import (
    "github.com/jmcarbo/fullmcp/server"
    "github.com/jmcarbo/fullmcp/transport/websocket"
)

func main() {
    srv := server.New("websocket-server")

    // Add MCP entities...

    // Create WebSocket server
    wsServer := websocket.NewServer(":8080", srv,
        websocket.WithReadTimeout(60*time.Second),
        websocket.WithWriteTimeout(10*time.Second),
        websocket.WithPingInterval(30*time.Second),
    )

    log.Printf("WebSocket server listening on :8080")
    if err := wsServer.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
```

### Client

```go
import (
    "github.com/jmcarbo/fullmcp/client"
    "github.com/jmcarbo/fullmcp/transport/websocket"
)

func main() {
    // Create WebSocket transport
    transport := websocket.New("ws://localhost:8080",
        websocket.WithReconnect(true),
        websocket.WithReconnectInterval(5*time.Second),
    )

    // Create client
    c := client.New(transport)

    ctx := context.Background()
    if err := c.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    // Use client...
}
```

### Connection Management

```go
// Automatic reconnection
transport := websocket.New("ws://localhost:8080",
    websocket.WithReconnect(true),
    websocket.WithReconnectInterval(5*time.Second),
    websocket.WithMaxReconnectAttempts(10),
)

// Connection lifecycle callbacks
transport.OnConnect(func() {
    log.Println("Connected to server")
})

transport.OnDisconnect(func(err error) {
    log.Printf("Disconnected: %v", err)
})

transport.OnReconnect(func(attempt int) {
    log.Printf("Reconnect attempt %d", attempt)
})
```

### WebSocket over TLS

```go
// WSS server
wsServer := websocket.NewServer(":8443", srv,
    websocket.WithTLS("server.crt", "server.key"),
)

// WSS client
transport := websocket.New("wss://localhost:8443")
```

### Use Cases

- Real-time dashboards
- Live notifications
- Collaborative editing
- Streaming data
- Gaming or interactive applications

## SSE Transport

Server-Sent Events for one-way server-to-client streaming.

### Server

```go
import (
    "github.com/jmcarbo/fullmcp/server"
    "github.com/jmcarbo/fullmcp/transport/sse"
)

func main() {
    srv := server.New("sse-server")

    // Create SSE server
    sseServer := sse.NewServer(":8080", srv,
        sse.WithRetryInterval(3*time.Second),
        sse.WithKeepaliveInterval(30*time.Second),
    )

    log.Printf("SSE server listening on :8080")
    if err := sseServer.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
```

### Client

```go
import (
    "github.com/jmcarbo/fullmcp/client"
    "github.com/jmcarbo/fullmcp/transport/sse"
)

func main() {
    // Create SSE transport
    transport := sse.New("http://localhost:8080/events",
        sse.WithReconnect(true),
    )

    // Create client
    c := client.New(transport)

    ctx := context.Background()
    if err := c.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    // Use client...
}
```

### Event Handling

```go
transport := sse.New("http://localhost:8080/events")

// Handle custom events
transport.OnEvent("notification", func(event sse.Event) {
    log.Printf("Notification: %s", event.Data)
})

transport.OnEvent("error", func(event sse.Event) {
    log.Printf("Error: %s", event.Data)
})
```

### Use Cases

- Server-pushed notifications
- Live feeds (news, social media)
- Progress updates
- Log streaming
- Monitoring dashboards

## Custom Transports

Implement custom transports for specialized communication needs.

### Transport Interface

```go
type Transport interface {
    io.ReadWriteCloser
}
```

### Example: Custom TCP Transport

```go
package mytransport

import (
    "net"
)

type TCPTransport struct {
    conn net.Conn
}

func New(address string) (*TCPTransport, error) {
    conn, err := net.Dial("tcp", address)
    if err != nil {
        return nil, err
    }

    return &TCPTransport{conn: conn}, nil
}

func (t *TCPTransport) Read(p []byte) (n int, err error) {
    return t.conn.Read(p)
}

func (t *TCPTransport) Write(p []byte) (n int, err error) {
    return t.conn.Write(p)
}

func (t *TCPTransport) Close() error {
    return t.conn.Close()
}
```

### Example: Custom Unix Socket Transport

```go
type UnixSocketTransport struct {
    conn net.Conn
}

func NewUnixSocket(socketPath string) (*UnixSocketTransport, error) {
    conn, err := net.Dial("unix", socketPath)
    if err != nil {
        return nil, err
    }

    return &UnixSocketTransport{conn: conn}, nil
}

func (t *UnixSocketTransport) Read(p []byte) (n int, err error) {
    return t.conn.Read(p)
}

func (t *UnixSocketTransport) Write(p []byte) (n int, err error) {
    return t.conn.Write(p)
}

func (t *UnixSocketTransport) Close() error {
    return t.conn.Close()
}
```

### Example: Buffered Transport

```go
type BufferedTransport struct {
    transport io.ReadWriteCloser
    reader    *bufio.Reader
    writer    *bufio.Writer
}

func NewBuffered(transport io.ReadWriteCloser) *BufferedTransport {
    return &BufferedTransport{
        transport: transport,
        reader:    bufio.NewReader(transport),
        writer:    bufio.NewWriter(transport),
    }
}

func (bt *BufferedTransport) Read(p []byte) (n int, err error) {
    return bt.reader.Read(p)
}

func (bt *BufferedTransport) Write(p []byte) (n int, err error) {
    n, err = bt.writer.Write(p)
    if err != nil {
        return n, err
    }
    return n, bt.writer.Flush()
}

func (bt *BufferedTransport) Close() error {
    bt.writer.Flush()
    return bt.transport.Close()
}
```

### Usage

```go
// Use custom transport with client
customTransport, err := mytransport.New("localhost:9000")
if err != nil {
    log.Fatal(err)
}

c := client.New(customTransport)
```

## Transport Comparison

| Feature | stdio | HTTP | WebSocket | SSE |
|---------|-------|------|-----------|-----|
| Bidirectional | ✅ | ✅ | ✅ | ❌ |
| Real-time | ✅ | ❌ | ✅ | ✅ |
| Stateful | ✅ | ❌ | ✅ | ✅ |
| Browser support | ❌ | ✅ | ✅ | ✅ |
| Firewall friendly | ✅ | ✅ | ⚠️ | ✅ |
| Reconnection | N/A | Built-in | Custom | Custom |
| Auth support | N/A | ✅ | ✅ | ✅ |
| Best for | CLI | APIs | Real-time | Streaming |

## Best Practices

### Choose the Right Transport

**stdio:**
- ✅ CLI tools, local development
- ❌ Remote clients, web browsers

**HTTP:**
- ✅ REST APIs, firewall-friendly, caching
- ❌ Real-time updates, streaming

**WebSocket:**
- ✅ Real-time apps, bidirectional streaming
- ❌ Simple request/response, HTTP caching

**SSE:**
- ✅ Server-to-client streaming, browser support
- ❌ Client-to-server streaming

### Timeouts

Always configure appropriate timeouts:

```go
// ✅ Good
transport := http.New("http://localhost:8080",
    http.WithTimeout(30*time.Second),
)

// ❌ Bad: No timeout
transport := http.New("http://localhost:8080")
```

### Error Handling

```go
transport := websocket.New("ws://localhost:8080")

transport.OnDisconnect(func(err error) {
    if err != nil {
        log.Printf("Connection error: %v", err)
        // Implement retry logic
    }
})
```

### Resource Cleanup

```go
defer func() {
    if err := c.Close(); err != nil {
        log.Printf("Error closing client: %v", err)
    }
}()
```

### Security

```go
// ✅ Good: Use TLS in production
transport := http.New("https://api.example.com",
    http.WithTLSConfig(&tls.Config{
        MinVersion: tls.VersionTLS12,
    }),
)

// ❌ Bad: Insecure in production
transport := http.New("http://api.example.com")
```

## Testing

### Mock Transport

```go
type MockTransport struct {
    readData  []byte
    writeData []byte
    readPos   int
}

func NewMock(data []byte) *MockTransport {
    return &MockTransport{readData: data}
}

func (mt *MockTransport) Read(p []byte) (n int, err error) {
    if mt.readPos >= len(mt.readData) {
        return 0, io.EOF
    }

    n = copy(p, mt.readData[mt.readPos:])
    mt.readPos += n
    return n, nil
}

func (mt *MockTransport) Write(p []byte) (n int, err error) {
    mt.writeData = append(mt.writeData, p...)
    return len(p), nil
}

func (mt *MockTransport) Close() error {
    return nil
}

// Usage in tests
func TestClient(t *testing.T) {
    mockTransport := NewMock([]byte(`{"result": "success"}`))
    c := client.New(mockTransport)
    // Test client...
}
```

## Related Documentation

- [Architecture Overview](./architecture.md)
- [Authentication](./authentication.md)
- [Middleware](./middleware.md)
