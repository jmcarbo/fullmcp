# Architecture Overview

FullMCP is designed with a modular, layered architecture that emphasizes type safety, concurrency, and extensibility.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│  (Tools, Resources, Prompts, Middleware, Lifecycle Hooks)   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      Server / Client                         │
│        (Message Routing, Request Handling, State)            │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      JSON-RPC Layer                          │
│    (MessageReader, MessageWriter, Request/Response IDs)      │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     Transport Layer                          │
│        (stdio, HTTP, WebSocket, SSE, Custom)                 │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### MCP Package (`mcp/`)

The foundation of the protocol implementation containing core types and error definitions:

- **types.go**: Protocol types (Tool, Resource, Prompt, Message, Content, Capabilities)
- **errors.go**: MCP error codes (InvalidRequest, MethodNotFound, InternalError, etc.)

All MCP types are strictly typed Go structs with JSON serialization support.

### Server Package (`server/`)

#### Server Core (`server.go`)

The main server implementation handles:
- JSON-RPC message routing
- Capability negotiation
- Request/response lifecycle
- Concurrent request handling

**Key Methods:**
```go
func New(name string, opts ...Option) *Server
func (s *Server) Run(ctx context.Context) error
func (s *Server) HandleMessage(ctx context.Context, msg *Message) *Message
```

#### Tool Manager (`tool.go`)

Manages tool registration and execution:
- Tool registry with concurrent-safe map
- Handler invocation with reflection
- Input validation via JSON schemas
- Error handling and recovery

```go
func (tm *ToolManager) AddTool(tool *Tool)
func (tm *ToolManager) CallTool(ctx context.Context, name string, args json.RawMessage) (interface{}, error)
```

#### Resource Manager (`resource.go`)

Handles static and templated resources:
- Resource registry
- URI template matching and parameter extraction
- Resource metadata (_meta fields)
- Content type handling

```go
func (rm *ResourceManager) AddResource(resource *Resource)
func (rm *ResourceManager) AddResourceTemplate(template *ResourceTemplate)
func (rm *ResourceManager) ReadResource(ctx context.Context, uri string) (*ResourceContent, error)
```

#### Prompt Manager (`prompt.go`)

Manages reusable message templates:
- Prompt registry
- Argument validation
- Message rendering
- Flexible argument handling

```go
func (pm *PromptManager) AddPrompt(prompt *Prompt)
func (pm *PromptManager) GetPrompt(ctx context.Context, name string, args map[string]interface{}) ([]*PromptMessage, error)
```

### Client Package (`client/`)

Asynchronous MCP client implementation using goroutines and channels:

- **Concurrent message handling**: Background goroutine routes responses
- **Pending request tracking**: Map of request IDs to response channels
- **Request ID generation**: Atomic counter for unique IDs
- **Transport abstraction**: Works with any io.ReadWriteCloser

```go
func New(transport io.ReadWriteCloser) *Client
func (c *Client) Connect(ctx context.Context) error
func (c *Client) CallTool(ctx context.Context, name string, args json.RawMessage) (json.RawMessage, error)
```

### Builder Package (`builder/`)

Fluent APIs for constructing MCP entities with automatic schema generation:

#### Tool Builder (`tool.go`)
- Reflection-based JSON schema generation using `jsonschema` package
- Type-safe handler registration
- Automatic input/output schema inference
- Support for tool hints (readonly, destructive, idempotent, open-world)

```go
func NewTool(name string) *ToolBuilder
func (tb *ToolBuilder) Handler(fn interface{}) *ToolBuilder
func (tb *ToolBuilder) Build() (*Tool, error)
```

### Transport Layer (`transport/`)

Multiple transport implementations:

#### stdio (`transport/stdio/`)
- Standard input/output communication
- Used for local process communication
- Default transport for CLI tools

#### HTTP (`transport/http/`)
- RESTful HTTP API
- Supports authentication middleware
- Standard HTTP status codes

#### WebSocket (`transport/websocket/`)
- Full-duplex real-time communication
- Connection lifecycle management
- Automatic reconnection support

#### SSE (`transport/sse/`)
- Server-Sent Events for streaming
- One-way server-to-client communication
- Event-based architecture

### Authentication (`auth/`)

Pluggable authentication providers:

#### API Key (`auth/apikey/`)
- Simple key-based authentication
- Claims storage
- Configurable scopes

#### JWT (`auth/jwt/`)
- Token-based authentication
- Signing and verification
- Expiration handling

#### OAuth 2.0 (`auth/oauth/`)
- Multiple providers (Google, GitHub, Azure)
- Authorization code flow
- Token management

### Internal Packages

#### JSON-RPC (`internal/jsonrpc/`)
- MessageReader: Reads JSON-RPC messages from transport
- MessageWriter: Writes JSON-RPC messages to transport
- Protocol compliance: Strict JSON-RPC 2.0 implementation

## Design Patterns

### Manager Pattern

Used for Tools, Resources, and Prompts:
- Centralized registry
- Thread-safe operations
- Consistent API across different entity types

### Builder Pattern

Fluent APIs for entity construction:
- Method chaining
- Optional parameters
- Validation at build time
- Type safety

### Middleware Chain

Composable middleware for cross-cutting concerns:
- Request/response interception
- Logging, metrics, authentication
- Error recovery
- Order-dependent execution

### Transport Abstraction

All transports implement `io.ReadWriteCloser`:
- Pluggable transport layer
- Consistent interface
- Easy testing with mock transports

## Concurrency Model

### Server Concurrency

- **Request handling**: Each request handled in a separate goroutine
- **Manager synchronization**: Concurrent-safe maps with sync.RWMutex
- **Context propagation**: All operations accept context.Context

### Client Concurrency

- **Response routing**: Background goroutine routes responses to waiting requests
- **Pending requests**: Map protected by mutex for thread-safe access
- **Request IDs**: Atomic counter ensures unique IDs

### Best Practices

1. Always pass context.Context for cancellation and timeouts
2. Use sync.RWMutex for read-heavy operations
3. Avoid shared mutable state
4. Prefer channels for goroutine communication

## Protocol Flow

### Server Initialization

1. Create server with configuration
2. Register tools, resources, prompts
3. Apply middleware
4. Register lifecycle hooks
5. Start transport listener

### Client Request Flow

1. Client sends JSON-RPC request with unique ID
2. Store pending response channel in map
3. Background goroutine receives response
4. Response routed to waiting channel
5. Result returned to caller

### Tool Execution Flow

1. Client calls `CallTool(name, args)`
2. Server routes to ToolManager
3. ToolManager validates input against schema
4. Handler invoked with reflection
5. Result serialized and returned

## Extension Points

### Custom Transports

Implement `io.ReadWriteCloser`:
```go
type CustomTransport struct {
    // ...
}

func (t *CustomTransport) Read(p []byte) (n int, err error) { }
func (t *CustomTransport) Write(p []byte) (n int, err error) { }
func (t *CustomTransport) Close() error { }
```

### Custom Middleware

Implement `server.Middleware`:
```go
func CustomMiddleware(next server.Handler) server.Handler {
    return func(ctx context.Context, req *server.Request) (*server.Response, error) {
        // Pre-processing
        resp, err := next(ctx, req)
        // Post-processing
        return resp, err
    }
}
```

### Custom Authentication

Implement `auth.Provider`:
```go
type CustomAuth struct { }

func (a *CustomAuth) Authenticate(ctx context.Context, token string) (*auth.Claims, error) { }
func (a *CustomAuth) Middleware() func(http.Handler) http.Handler { }
```

## Performance Considerations

### Memory Efficiency

- Streaming for large resources
- Pooling for frequent allocations
- Minimal copying of byte slices

### CPU Efficiency

- Lazy schema generation
- Efficient JSON parsing
- Minimal reflection overhead

### Network Efficiency

- Message batching (where applicable)
- Compression support
- Connection pooling

## Testing Strategy

### Unit Tests

- Individual component testing
- Mock dependencies
- Table-driven tests

### Integration Tests

- End-to-end scenarios
- Real transport communication
- Multi-server composition

### Benchmarks

- Tool call performance
- Resource read performance
- Message handling throughput
- Client connection overhead

## Security Model

### Defense in Depth

1. **Input validation**: JSON schema validation for all inputs
2. **Authentication**: Pluggable auth providers
3. **Authorization**: Scope-based access control
4. **Transport security**: TLS support for HTTP/WebSocket
5. **Rate limiting**: Middleware-based rate limiting
6. **Error sanitization**: No sensitive data in error messages

### Threat Mitigation

- **Injection attacks**: Schema validation prevents malformed input
- **DoS**: Rate limiting and request timeouts
- **Man-in-the-middle**: TLS encryption
- **Unauthorized access**: Authentication middleware

## Future Architecture

### Planned Enhancements

- gRPC transport
- Message queuing integration
- Distributed tracing
- OpenTelemetry integration
- Plugin system
- Hot reload support

See [IMPLEMENTATION_STATUS.md](../IMPLEMENTATION_STATUS.md) for detailed roadmap.
