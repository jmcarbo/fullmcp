# FullMCP - Production-Ready Golang MCP Implementation

[![Go Reference](https://pkg.go.dev/badge/github.com/jmcarbo/fullmcp.svg)](https://pkg.go.dev/github.com/jmcarbo/fullmcp)
[![Go Report Card](https://goreportcard.com/badge/github.com/jmcarbo/fullmcp)](https://goreportcard.com/report/github.com/jmcarbo/fullmcp)
[![Coverage](https://img.shields.io/badge/coverage-95.8%25-brightgreen)](https://github.com/jmcarbo/fullmcp)

A comprehensive, production-ready Golang implementation of the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) with full support for tools, resources, prompts, and multiple transport mechanisms.

## Features

### Core Protocol
- ✅ **Full MCP 2025-06-18 Support**: Complete implementation of latest specification
- ✅ **Tool Output Schemas**: Define expected output structure for better type safety
- ✅ **Elicitation**: Servers can request structured user input with JSON Schema validation
- ✅ **Resource Metadata**: _meta fields for version tracking and audience targeting
- ✅ **Title Fields**: Human-friendly display names for better UX
- ✅ **Type-Safe**: Leverages Go's static typing with automatic JSON schema generation
- ✅ **Concurrent**: Thread-safe operations designed for high-concurrency environments
- ✅ **Idiomatic Go**: Follows Go best practices and conventions

### Transports
- ✅ **stdio**: Standard input/output transport
- ✅ **HTTP**: RESTful HTTP transport with authentication
- ✅ **WebSocket**: Full-duplex real-time communication
- ✅ **SSE**: Server-Sent Events for streaming

### Advanced Features
- ✅ **Authentication**: API Key, JWT, and OAuth 2.0 (Google, GitHub, Azure)
- ✅ **Middleware**: Composable middleware chain for logging, recovery, etc.
- ✅ **Proxy Server**: Forward requests to backend MCP servers
- ✅ **Server Composition**: Mount multiple servers under namespaces
- ✅ **Resource Templates**: Parameterized resources with URI templates
- ✅ **Lifecycle Hooks**: Startup and shutdown management
- ✅ **Builder Pattern**: Fluent APIs for easy configuration

### Developer Experience
- ✅ **CLI Tool**: `mcpcli` for testing and debugging MCP servers
- ✅ **95.8% Test Coverage**: Comprehensive test suite
- ✅ **Performance Benchmarks**: Measure and optimize operations
- ✅ **Integration Tests**: End-to-end scenario testing

## Installation

```bash
go get github.com/jmcarbo/fullmcp
```

### Install CLI Tool

```bash
go install github.com/jmcarbo/fullmcp/cmd/mcpcli@latest
```

## Quick Start

### Basic Server

```go
package main

import (
    "context"
    "log"

    "github.com/jmcarbo/fullmcp/builder"
    "github.com/jmcarbo/fullmcp/server"
)

type AddArgs struct {
    A int `json:"a" jsonschema:"required,description=First number"`
    B int `json:"b" jsonschema:"required,description=Second number"`
}

func main() {
    // Create server
    srv := server.New("math-server")

    // Add tool with automatic schema generation
    addTool, _ := builder.NewTool("add").
        Description("Add two numbers").
        Handler(func(ctx context.Context, args AddArgs) (int, error) {
            return args.A + args.B, nil
        }).
        Build()

    srv.AddTool(addTool)

    // Serve on stdio
    log.Fatal(srv.Run(context.Background()))
}
```

### Client

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/jmcarbo/fullmcp/client"
    "github.com/jmcarbo/fullmcp/transport/stdio"
)

func main() {
    // Create client with stdio transport
    transport := stdio.New()
    c := client.New(transport)

    ctx := context.Background()
    if err := c.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    // List available tools
    tools, err := c.ListTools(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d tools\n", len(tools))

    // Call a tool
    result, err := c.CallTool(ctx, "add", json.RawMessage(`{"a":5,"b":3}`))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Result: %v\n", result)
}
```

## Documentation

### Guides

- [Architecture Overview](./docs/architecture.md)
- [Building Tools](./docs/tools.md)
- [Managing Resources](./docs/resources.md)
- [Using Prompts](./docs/prompts.md)
- [Authentication](./docs/authentication.md)
- [Transports](./docs/transports.md)
- [Middleware](./docs/middleware.md)
- [CLI Tool Usage](./cmd/mcpcli/README.md)

### Examples

Comprehensive examples are available in the [`examples/`](./examples/) directory:

- **[basic-server](./examples/basic-server/)**: Simple math server with stdio
- **[advanced-server](./examples/advanced-server/)**: Middleware, lifecycle hooks, resource templates
- **[http-server](./examples/http-server/)**: HTTP transport with API key authentication
- **[websocket-server](./examples/websocket-server/)**: WebSocket transport with real-time communication

### API Reference

Full API documentation is available at [pkg.go.dev](https://pkg.go.dev/github.com/jmcarbo/fullmcp).

## Features in Detail

### Tools

Tools are functions that can be invoked by clients. Schema is automatically generated from Go structs.

```go
type CalculateArgs struct {
    Operation string  `json:"op" jsonschema:"required,enum=add|subtract|multiply|divide"`
    A         float64 `json:"a" jsonschema:"required"`
    B         float64 `json:"b" jsonschema:"required"`
}

calcTool, _ := builder.NewTool("calculate").
    Description("Perform mathematical operations").
    Handler(func(ctx context.Context, args CalculateArgs) (float64, error) {
        switch args.Operation {
        case "add":
            return args.A + args.B, nil
        case "subtract":
            return args.A - args.B, nil
        case "multiply":
            return args.A * args.B, nil
        case "divide":
            if args.B == 0 {
                return 0, fmt.Errorf("division by zero")
            }
            return args.A / args.B, nil
        default:
            return 0, fmt.Errorf("invalid operation")
        }
    }).
    Build()

srv.AddTool(calcTool)
```

### Resources

Resources provide read-only access to data.

```go
// Static resource
configResource := builder.NewResource("config://app").
    Name("Application Config").
    Description("Main application configuration").
    MimeType("application/json").
    Reader(func(ctx context.Context) ([]byte, error) {
        config := map[string]interface{}{
            "debug": true,
            "port":  8080,
        }
        return json.Marshal(config)
    }).
    Build()

srv.AddResource(configResource)

// Resource template (parameterized)
fileTemplate := builder.NewResourceTemplate("file:///{path}").
    Name("File Reader").
    Description("Read files from the filesystem").
    MimeType("text/plain").
    ReaderSimple(func(ctx context.Context, path string) ([]byte, error) {
        return os.ReadFile(path)
    }).
    Build()

srv.AddResourceTemplate(fileTemplate)
```

### Prompts

Prompts are reusable message templates.

```go
greetingPrompt := builder.NewPrompt("greeting").
    Description("Generate a personalized greeting").
    Argument("name", "Person's name", true).
    Argument("time", "Time of day (morning/afternoon/evening)", false).
    Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
        name := args["name"].(string)
        timeOfDay := "day"
        if t, ok := args["time"].(string); ok {
            timeOfDay = t
        }

        return []*mcp.PromptMessage{{
            Role: "user",
            Content: []mcp.Content{
                &mcp.TextContent{
                    Type: "text",
                    Text: fmt.Sprintf("Good %s, %s!", timeOfDay, name),
                },
            },
        }}, nil
    }).
    Build()

srv.AddPrompt(greetingPrompt)
```

### Authentication

#### API Key Authentication

```go
import "github.com/jmcarbo/fullmcp/auth/apikey"

authProvider := apikey.New()
authProvider.AddKey("secret-key-123", auth.Claims{
    Subject: "user-1",
    Email:   "user@example.com",
    Scopes:  []string{"read", "write"},
})

// Use in HTTP server
handler := authProvider.Middleware()(httpHandler)
```

#### JWT Authentication

```go
import "github.com/jmcarbo/fullmcp/auth/jwt"

key, _ := jwt.GenerateRandomKey(32)
jwtProvider := jwt.New(key,
    jwt.WithIssuer("mcp-server"),
    jwt.WithExpiration(24*time.Hour),
)

// Create token
token, _ := jwtProvider.CreateToken("user123", "user@example.com",
    []string{"read", "write"}, nil)

// Validate token
claims, _ := jwtProvider.ValidateToken(ctx, token)

// Use as middleware
handler := jwtProvider.Middleware()(httpHandler)
```

#### OAuth 2.0

```go
import "github.com/jmcarbo/fullmcp/auth/oauth"

// Google OAuth
provider := oauth.New(
    oauth.Google,
    "client-id",
    "client-secret",
    "http://localhost:8080/callback",
    []string{"email", "profile"},
)

// Generate auth URL
authURL := provider.AuthCodeURL("state")

// Handle callback
http.HandleFunc("/callback", provider.HandleCallback())
```

### Transports

#### HTTP Transport

```go
import "github.com/jmcarbo/fullmcp/transport/http"

// Client
transport := http.New("http://localhost:8080")
c := client.New(transport)

// Server
httpServer := http.NewServer(":8080", srv)
log.Fatal(httpServer.ListenAndServe())
```

#### WebSocket Transport

```go
import "github.com/jmcarbo/fullmcp/transport/websocket"

// Client
transport := websocket.New("ws://localhost:8080")
c := client.New(transport)

// Server
wsServer := websocket.NewServer(":8080", func(ctx context.Context, msg []byte) ([]byte, error) {
    var mcpMsg mcp.Message
    json.Unmarshal(msg, &mcpMsg)
    response := srv.HandleMessage(ctx, &mcpMsg)
    return json.Marshal(response)
})
log.Fatal(wsServer.ListenAndServe())
```

### Middleware

```go
import "github.com/jmcarbo/fullmcp/server"

// Create custom middleware
loggingMiddleware := func(next server.Handler) server.Handler {
    return func(ctx context.Context, req *server.Request) (*server.Response, error) {
        log.Printf("Request: %s", req.Method)
        resp, err := next(ctx, req)
        log.Printf("Response: error=%v", err)
        return resp, err
    }
}

// Apply middleware
srv := server.New("my-server",
    server.WithMiddleware(
        server.RecoveryMiddleware(),
        loggingMiddleware,
    ),
)
```

### Proxy Server

Forward requests to a backend MCP server:

```go
import "github.com/jmcarbo/fullmcp/server/proxy"

// Create backend client
backendTransport := stdio.New()
backendClient := client.New(backendTransport)

// Create proxy
proxyServer, _ := proxy.New("proxy-server", backendClient)

// Serve proxy
log.Fatal(proxyServer.Run(context.Background()))
```

### Server Composition

Mount multiple servers under namespaces:

```go
cs := server.NewCompositeServer("main")

apiServer := server.New("api-server")
adminServer := server.New("admin-server")

cs.Mount("api", apiServer)
cs.Mount("admin", adminServer)

// Tools are namespaced: "api/list-users", "admin/restart-service"
```

## CLI Tool

Test and debug MCP servers with `mcpcli`:

```bash
# Test connection
mcpcli ping

# List tools
mcpcli list-tools

# Call a tool
mcpcli call-tool add --args '{"a":5,"b":3}'

# List resources
mcpcli list-resources

# Read a resource
mcpcli read-resource config://app

# Get server info
mcpcli info
```

## Testing

### Run Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Verbose
go test -v ./...
```

### Run Benchmarks

```bash
go test -bench=. -benchmem ./server ./client ./builder
```

### Run Integration Tests

```bash
go test -v -run=TestIntegration ./...
```

## Performance

Performance benchmarks (on Apple M-series):

| Operation | Time | Memory |
|-----------|------|--------|
| Tool call | ~33 ns | 40 B |
| Resource read | ~22 ns | 16 B |
| Message handling | ~1.3 μs | 1.2 KB |
| Client connection | ~25 μs | 18 KB |

See [benchmarks](./docs/benchmarks.md) for detailed performance analysis.

## Project Status

**Production Ready** ✅

- ✅ 95.8% test coverage
- ✅ Zero linter errors
- ✅ Comprehensive integration tests
- ✅ Performance benchmarked
- ✅ Used in production environments

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

[MIT License](./LICENSE)

## Acknowledgments

- Inspired by [fastmcp](https://github.com/jlowin/fastmcp)
- Implements [Model Context Protocol](https://modelcontextprotocol.io/) specification
