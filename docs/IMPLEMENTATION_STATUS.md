# Implementation Status

This document tracks the implementation status of features from the specification in `docs/golang-mcp-implementation-spec.md`.

## Completed Features ✅

### Phase 1: Core Protocol
- ✅ Core types and interfaces (`mcp/types.go`, `mcp/errors.go`)
- ✅ JSON-RPC layer (`internal/jsonrpc/`)
- ✅ Stdio transport (`transport/stdio/`)
- ✅ Basic server and client (`server/`, `client/`)
- ✅ Tool registration and execution (`server/tool.go`, `builder/tool.go`)
- ✅ Resource management (`server/resource.go`, `builder/resource.go`)
- ✅ Prompt management (`server/prompt.go`, `builder/prompt.go`)

### Phase 2: Transports
- ✅ Stdio transport (`transport/stdio/`)
- ✅ HTTP transport (`transport/http/`)
- ✅ SSE transport (`transport/sse/`)
- ✅ WebSocket transport (`transport/websocket/`)

### Phase 3: Advanced Features
- ✅ Resource templates (`server/resource.go` - `ResourceTemplateHandler`)
- ✅ Server composition/mounting (`server/composition.go` - `CompositeServer`)
- ✅ Proxy server (`server/proxy/` - forwards requests to backend MCP server)
- ✅ Middleware system (`server/middleware.go`)
- ✅ Context and dependencies (`server/context.go`)
- ✅ Lifecycle hooks (`server/lifecycle.go`)

### Phase 4: Authentication
- ✅ Auth interface (`auth/auth.go`)
- ✅ API key authentication (`auth/apikey/`)
- ✅ JWT authentication (`auth/jwt/`)
- ✅ OAuth providers (`auth/oauth/` - Google, GitHub, Azure)

### Phase 5: Developer Experience
- ✅ Builder APIs (`builder/tool.go`, `builder/resource.go`, `builder/prompt.go`)
- ✅ Reflection-based schema generation (via `invopop/jsonschema`)
- ✅ Comprehensive examples (`examples/basic-server/`, `examples/advanced-server/`, `examples/http-server/`)
- ✅ Documentation (comprehensive README.md)
- ✅ CLI tools (`cmd/mcpcli/`)

## New Features Implemented

### 1. Resource Templates
Location: `server/resource.go`

Allows parameterized resources with URI templates:

```go
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

### 2. Middleware Support
Location: `server/middleware.go`

Provides request/response middleware chain:

```go
srv := server.New("my-server",
    server.WithMiddleware(
        server.RecoveryMiddleware(),
        server.LoggingMiddleware(&SimpleLogger{}),
    ),
)
```

### 3. Lifecycle Hooks
Location: `server/lifecycle.go`

Manages server startup and shutdown:

```go
server.WithLifespan(func(ctx context.Context, s *server.Server) (context.Context, func(), error) {
    log.Println("Server starting up...")

    cleanup := func() {
        log.Println("Server shutting down...")
    }

    return ctx, cleanup, nil
})
```

### 4. Server Context
Location: `server/context.go`

Provides access to server capabilities from within handlers:

```go
sc := server.FromContext(ctx)
if sc != nil {
    data, err := sc.ReadResource(ctx, "config://app")
}
```

### 5. Builder Patterns
Locations: `builder/resource.go`, `builder/prompt.go`

Fluent API for creating resources and prompts:

```go
resource := builder.NewResource("config://app").
    Name("App Config").
    Description("Application configuration").
    MimeType("application/json").
    Reader(func(ctx context.Context) ([]byte, error) {
        return []byte(`{"debug": true}`), nil
    }).
    Build()
```

### 6. HTTP Transport
Location: `transport/http/`

HTTP-based transport for MCP:

```go
transport := http.New("http://localhost:8080")
client := client.New(transport)
```

### 7. Authentication Framework
Location: `auth/`

Pluggable authentication with API key support:

```go
authProvider := apikey.New()
authProvider.AddKey("secret-key-123", auth.Claims{
    Subject: "user-1",
    Email:   "user@example.com",
    Scopes:  []string{"read", "write"},
})

// Use in HTTP middleware
handler := authProvider.Middleware()(mux)
```

## Examples

### Basic Server
Location: `examples/basic-server/main.go`

Simple math server with tools, resources, and prompts using stdio transport.

### Advanced Server
Location: `examples/advanced-server/main.go`

Demonstrates:
- Middleware (logging, recovery)
- Lifecycle management
- Resource templates
- Builder patterns

### HTTP Server
Location: `examples/http-server/main.go`

Demonstrates:
- HTTP transport
- API key authentication
- Authenticated tool access

### WebSocket Server
Location: `examples/websocket-server/main.go`

Demonstrates:
- WebSocket transport
- Real-time bidirectional communication
- Math tools over WebSocket

## Testing

All implemented features have comprehensive tests with **95.8% overall coverage** (significantly exceeds 90% target):

```bash
go test ./...
```

Test coverage by package:
- ✅ auth: 100%
- ✅ auth/apikey: 100%
- ✅ mcp (core protocol types): 100%
- ✅ internal/jsonrpc: 100%
- ✅ transport/stdio: 100%
- ✅ builder: 97.6%
- ✅ server: 93.4%
- ✅ transport/http: 87%
- ✅ client: 84.4% (improved from 54.4% with async tests)

See [TEST_COVERAGE.md](./TEST_COVERAGE.md) for detailed coverage report.

## Newly Implemented Features (Latest Session)

### 8. SSE Transport
Location: `transport/sse/`

Server-Sent Events transport for one-way streaming:

```go
transport := sse.New("http://localhost:8080/events")
client := client.New(transport)
```

Server implementation with streaming:

```go
handler := sse.NewMCPSSEHandler(func(ctx context.Context, req []byte) ([]byte, error) {
    return processRequest(req)
})
server := sse.NewServer(":8080", handler)
server.ListenAndServe()
```

### 9. Server Composition
Location: `server/composition.go`

Mount multiple servers under different namespaces:

```go
cs := server.NewCompositeServer("main")

// Create sub-servers
apiServer := server.New("api-server")
adminServer := server.New("admin-server")

// Mount them
cs.Mount("api", apiServer)
cs.Mount("admin", adminServer)

// Tools, resources, and prompts are automatically namespaced
// e.g., "api/list-users", "admin/restart-service"
```

### 10. WebSocket Transport
Location: `transport/websocket/`

WebSocket transport for real-time bidirectional communication:

**Client:**
```go
transport := websocket.New("ws://localhost:8080")
client := client.New(transport)
```

**Server:**
```go
handler := websocket.NewServer(":8080", func(ctx context.Context, msg []byte) ([]byte, error) {
	// Parse message
	var mcpMsg mcp.Message
	json.Unmarshal(msg, &mcpMsg)

	// Handle with MCP server
	response := srv.HandleMessage(ctx, &mcpMsg)

	// Return serialized response
	return json.Marshal(response)
})

server.ListenAndServe()
```

**Features:**
- Full-duplex communication over WebSocket
- Custom origin checking
- Large message support with buffering
- Concurrent message handling

**Dependencies:** `github.com/gorilla/websocket v1.5.3`

### 11. Proxy Server
Location: `server/proxy/`

Proxy server that forwards MCP requests to a backend server:

```go
// Create backend client
backendClient := client.New(backendTransport)

// Create proxy that forwards to backend
proxy, err := proxy.New("proxy-server", backendClient)

// Proxy automatically syncs tools, resources, and prompts from backend
// and forwards all requests
```

**Use cases:**
- Load balancing across multiple MCP servers
- Adding authentication/authorization layer
- Request logging and monitoring
- Protocol translation

### 12. JWT Authentication
Location: `auth/jwt/`

JWT-based authentication with configurable signing methods and expiration:

```go
// Generate signing key
key, _ := jwt.GenerateRandomKey(32)

// Create JWT provider
provider := jwt.New(key,
	jwt.WithIssuer("my-mcp-server"),
	jwt.WithExpiration(24*time.Hour),
	jwt.WithSigningMethod(jwt.SigningMethodHS256),
)

// Create token
token, _ := provider.CreateToken(
	"user123",
	"user@example.com",
	[]string{"read", "write"},
	map[string]interface{}{"role": "admin"},
)

// Validate token
claims, _ := provider.ValidateToken(ctx, token)

// Use in HTTP middleware
handler := provider.Middleware()(myHandler)
```

**Features:**
- HS256, HS384, HS512 signing methods
- Configurable expiration and issuer
- Custom claims support
- HTTP middleware integration
- Automatic token validation

**Dependencies:** `github.com/golang-jwt/jwt/v5 v5.3.0`

### 13. OAuth 2.0 Providers
Location: `auth/oauth/`

OAuth 2.0 authentication for Google, GitHub, and Azure:

**Google OAuth:**
```go
provider := oauth.New(
	oauth.Google,
	"client-id",
	"client-secret",
	"http://localhost:8080/callback",
	[]string{"email", "profile"},
)

// Generate auth URL
authURL := provider.AuthCodeURL("random-state")

// Handle callback and exchange code
token, _ := provider.Exchange(ctx, code)

// Validate and get user claims
claims, _ := provider.ValidateToken(ctx, token.AccessToken)
```

**GitHub OAuth:**
```go
provider := oauth.New(
	oauth.GitHub,
	"client-id",
	"client-secret",
	"http://localhost:8080/callback",
	[]string{"user", "repo"},
)
```

**Advanced Options:**
```go
provider := oauth.New(oauth.Google, clientID, clientSecret, redirectURL, scopes,
	oauth.WithVerifyEmail(true),
	oauth.WithScopeMapping(map[string][]string{
		"role": {"admin", "user"},
	}),
	oauth.WithCustomEndpoint(authURL, tokenURL),
)

// HTTP middleware
handler := provider.Middleware()(myHandler)

// Callback handler
http.HandleFunc("/callback", provider.HandleCallback())
```

**Supported Providers:**
- Google (`oauth.Google`)
- GitHub (`oauth.GitHub`)
- Azure (`oauth.Azure`)
- Custom providers via `WithCustomEndpoint`

**Dependencies:** `golang.org/x/oauth2 v0.31.0`

### 14. CLI Tool (mcpcli)
Location: `cmd/mcpcli/`

Comprehensive command-line interface for MCP server management:

**Installation:**
```bash
go install github.com/jmcarbo/fullmcp/cmd/mcpcli@latest
```

**Commands:**
```bash
# Server operations
mcpcli ping                    # Test connection
mcpcli info                    # Display server capabilities

# Tools
mcpcli list-tools              # List available tools
mcpcli call-tool add --args '{"a":5,"b":3}'

# Resources
mcpcli list-resources          # List available resources
mcpcli read-resource config://app

# Prompts
mcpcli list-prompts            # List available prompts
mcpcli get-prompt greeting --args name=Alice
```

**Features:**
- Full MCP protocol support (tools, resources, prompts)
- JSON output for scripting (`--json` flag)
- Verbose mode for debugging (`--verbose`)
- Configurable timeouts (`--timeout`)
- Pipe-friendly output
- Auto-completion support

**Global Flags:**
- `-t, --timeout <seconds>` - Request timeout (default: 30)
- `-v, --verbose` - Enable verbose output
- `--json` - Output as JSON for scripting

**Dependencies:** `github.com/spf13/cobra v1.10.1`

### 15. Comprehensive Documentation
Location: `README.md`

Complete production-ready documentation including:

**Features:**
- Overview of all MCP capabilities (Tools, Resources, Prompts, Auth, Transports)
- Installation instructions
- Quick start examples for server and client
- Detailed feature documentation
- Performance benchmarks table
- Testing instructions
- CLI tool usage examples
- API reference for common operations

**Updated Documentation:**
- ✅ Full README.md with comprehensive coverage
- ✅ Installation and setup guide
- ✅ Quick start examples
- ✅ Feature-by-feature documentation
- ✅ Performance metrics
- ✅ Testing guidelines

### 16. Advanced Sampling Features
Location: `mcp/sampling.go`, `client/sampling.go`, `server/sampling.go`

LLM completion requests from servers to clients (MCP 2024-11-05 specification):

**Core Types:**
```go
type SamplingMessage struct {
    Role    string          // "user" or "assistant"
    Content SamplingContent // Message content
}

type CreateMessageRequest struct {
    Messages         []SamplingMessage
    ModelPreferences *ModelPreferences
    SystemPrompt     string
    MaxTokens        *int
    Temperature      *float64
    StopSequences    []string
    Metadata         map[string]string
}

type CreateMessageResult struct {
    Role       string          // "assistant"
    Content    SamplingContent // Generated content
    Model      string          // Actual model used
    StopReason string          // endTurn, stopSequence, maxTokens, error
}
```

**Client-Side Handler:**
```go
client := client.New(transport, client.WithSamplingHandler(
    func(ctx context.Context, req *mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error) {
        // Call your LLM API
        return callLLM(req)
    },
))
```

**Server-Side Usage:**
```go
srv := server.New("ai-server", server.EnableSampling())

// From tool handler
req := server.NewSamplingRequest().
    WithSystemPrompt("You are a helpful assistant").
    WithMaxTokens(100).
    WithTemperature(0.7).
    WithModelPreferences(
        server.NewModelPreferences("claude-3-sonnet", "gpt-4").
            WithIntelligencePriority(0.8).
            WithSpeedPriority(0.5),
    ).
    AddUserMessage("What is 2+2?")

result, _ := srv.CreateMessage(ctx, req)
```

**Features:**
- Builder pattern for request construction
- Model preferences with intelligence/speed priorities
- Support for text and image content
- Multi-turn conversations
- Stop sequences and temperature control
- Comprehensive tests and example

**Files:**
- `mcp/sampling.go` - Core types
- `mcp/sampling_builder.go` - Fluent builder methods
- `mcp/sampling_test.go` - Type serialization tests
- `client/sampling.go` - Client handler support
- `server/sampling.go` - Server capability
- `examples/sampling/main.go` - Full demonstration

### 17. Roots (Filesystem Boundaries)
Location: `mcp/roots.go`, `client/roots.go`, `server/roots.go`

Filesystem boundaries for security and access control (MCP 2024-11-05 specification):

**Core Types:**
```go
type Root struct {
    URI  string // URI of the root (e.g., "file:///path")
    Name string // Optional human-readable name
}

type RootsListResult struct {
    Roots []Root
}

type RootsCapability struct {
    ListChanged bool // Whether client emits change notifications
}
```

**Client-Side Provider:**
```go
rootsProvider := func(ctx context.Context) ([]mcp.Root, error) {
    return []mcp.Root{
        {
            URI:  "file:///home/user/projects/myapp",
            Name: "Main Project",
        },
        {
            URI:  "file:///home/user/Documents",
            Name: "Documents",
        },
    }, nil
}

client := client.New(transport, client.WithRoots(rootsProvider))

// Notify server of changes
client.NotifyRootsChanged()
```

**Server-Side Handler:**
```go
srv := server.New("file-server",
    server.WithRootsHandler(func(ctx context.Context) {
        // Handle roots change notification
        log.Println("Roots have changed, refreshing access controls")
    }),
)

// Request roots from client
roots, _ := srv.ListRoots(ctx)
```

**Features:**
- Security boundaries for file access
- Dynamic roots with change notifications
- Support for various URI schemes (file://, https://, git://)
- Client declares capability during initialization
- Server receives notifications via `notifications/roots/list_changed`
- Bidirectional communication support

**Use Cases:**
- IDE workspace folders
- File server access control
- Code analysis tool boundaries
- Multi-project environments

**Files:**
- `mcp/roots.go` - Core types
- `mcp/roots_test.go` - Serialization tests
- `client/roots.go` - Client provider and notifications
- `server/roots.go` - Server request handling
- `examples/roots/main.go` - Full demonstration with security examples

### 18. Logging Protocol Extensions
Location: `mcp/logging.go`, `server/logging.go`, `client/logging.go`

Structured logging with configurable severity levels (MCP 2024-11-05 specification):

**Core Types:**
```go
type LogLevel string

const (
    LogLevelDebug     LogLevel = "debug"
    LogLevelInfo      LogLevel = "info"
    LogLevelNotice    LogLevel = "notice"
    LogLevelWarning   LogLevel = "warning"
    LogLevelError     LogLevel = "error"
    LogLevelCritical  LogLevel = "critical"
    LogLevelAlert     LogLevel = "alert"
    LogLevelEmergency LogLevel = "emergency"
)

type LogMessage struct {
    Level  LogLevel               // Severity level
    Logger string                 // Optional logger name
    Data   map[string]interface{} // Structured log data
}
```

**Server-Side Logging:**
```go
srv := server.New("my-server", server.EnableLogging())

// Logging methods
srv.Log(mcp.LogLevelInfo, "mylogger", map[string]interface{}{
    "event": "startup",
    "port":  8080,
})

// Convenience methods
srv.LogDebug("debug-logger", data)
srv.LogInfo("info-logger", data)
srv.LogWarning("warn-logger", data)
srv.LogError("error-logger", data)
```

**Client-Side Handler:**
```go
client := client.New(transport,
    client.WithLogHandler(func(ctx context.Context, msg *mcp.LogMessage) {
        log.Printf("[%s] %s: %v", msg.Level, msg.Logger, msg.Data)
    }),
)

// Set minimum log level
client.SetLogLevel(ctx, mcp.LogLevelInfo)
```

**Features:**
- RFC 5424 syslog severity levels
- Structured logging with arbitrary JSON data
- Configurable minimum log level
- Log filtering by severity
- Optional logger names
- Client-side log aggregation

**Protocol:**
- Client sets level: `logging/setLevel`
- Server sends logs: `notifications/message`

**Files:**
- `mcp/logging.go` - Core types and level comparison
- `mcp/logging_test.go` - Type tests
- `server/logging.go` - Server logging manager
- `client/logging.go` - Client log handler
- `examples/logging/main.go` - Full demonstration

### 19. Progress Notifications
Location: `mcp/progress.go`, `server/progress.go`, `client/progress.go`

Progress tracking for long-running operations (MCP 2024-11-05 specification):

**Core Types:**
```go
type ProgressNotification struct {
    ProgressToken interface{} // Unique token (string or int)
    Progress      float64      // Current progress value
    Total         *float64     // Optional total value
}
```

**Server-Side Progress:**
```go
srv := server.New("my-server", server.WithProgress())

// Send progress updates
total := 100.0
srv.NotifyProgress("task-123", 50.0, &total) // 50%

// Without total (indefinite progress)
srv.NotifyProgress("task-456", 42.0, nil)

// Using progress context
pc := server.NewProgressContext(token, srv.Progress)
pc.Update(75.0, &total)
```

**Client-Side Handler:**
```go
client := client.New(transport,
    client.WithProgressHandler(
        func(ctx context.Context, notif *mcp.ProgressNotification) {
            if notif.Total != nil {
                percent := (notif.Progress / *notif.Total) * 100
                fmt.Printf("[%v] %.1f%% complete\n",
                    notif.ProgressToken, percent)
            } else {
                fmt.Printf("[%v] Processed: %.0f\n",
                    notif.ProgressToken, notif.Progress)
            }
        },
    ),
)
```

**Features:**
- Unique progress tokens per operation
- Determinate progress (with total)
- Indeterminate progress (without total)
- Floating-point progress values
- Multiple concurrent operations
- Token-based correlation

**Requirements:**
- Progress MUST increase with each notification
- Tokens MUST be unique across active requests
- Progress and total MAY be floating point

**Use Cases:**
- File uploads/downloads
- Batch processing
- Data import/export
- Report generation
- Search/indexing
- Long computations

**Files:**
- `mcp/progress.go` - Core types
- `mcp/progress_test.go` - Type tests
- `server/progress.go` - Server progress tracker
- `client/progress.go` - Client progress handler
- `examples/progress/main.go` - Full demonstration

### 20. Cancellation Support
Location: `mcp/cancellation.go`, `server/cancellation.go`, `client/cancellation.go`

Request cancellation for long-running operations (MCP 2024-11-05 specification):

**Core Types:**
```go
type CancelledNotification struct {
    RequestID interface{} // ID of request to cancel
    Reason    string      // Optional reason
}
```

**Server-Side Cancellation:**
```go
srv := server.New("my-server", server.WithCancellation())

// In handler with cancellable context
ctx, cancel := context.WithCancel(parentCtx)
srv.RegisterCancellable(requestID, cancel)
defer srv.UnregisterCancellable(requestID)

// Check for cancellation
select {
case <-ctx.Done():
    return nil, ctx.Err()
case result := <-workDone:
    return result, nil
}
```

**Client-Side Cancellation:**
```go
// Send cancellation notification
err := client.CancelRequest(requestID, "User cancelled")
```

**Features:**
- Context-based cancellation
- Cancellation registration/unregistration
- Optional cancellation reasons
- Race condition handling
- Best-effort cancellation (no guarantees)

**Requirements:**
- MUST only cancel requests in same direction
- SHOULD ignore responses after sending cancellation
- MAY include optional reason string

**Constraints:**
- Network latency may cause race conditions
- Cancellation may arrive after completion
- No guarantee cancellation will be processed

**Use Cases:**
- User-initiated cancellation
- Timeouts
- Resource limit exceeded
- Client shutdown
- Priority changes
- Duplicate requests

**Files:**
- `mcp/cancellation.go` - Core types
- `mcp/cancellation_test.go` - Type tests
- `server/cancellation.go` - Server cancellation manager
- `client/cancellation.go` - Client cancellation sender
- `examples/cancellation/main.go` - Full demonstration

### 21. Ping Utility
Location: `mcp/ping.go`, `server/server.go`, `client/ping.go`

Simple connection health check utility (MCP 2024-11-05 specification):

**Purpose:**
- Keep connections alive
- Detect dead connections
- Verify server responsiveness
- Monitor connection latency
- Implement health checks

**Server-Side:**
```go
// Ping is always available, no special setup needed
srv := server.New("my-server")

// Server automatically handles ping requests
// Returns empty object on success
```

**Client-Side:**
```go
// Send ping request
err := client.Ping(ctx)
if err != nil {
    log.Printf("Server unreachable: %v", err)
}

// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := client.Ping(ctx)
```

**Features:**
- Always available (no capabilities needed)
- Simple request/response pattern
- Empty result on success
- Either party can initiate
- Useful for connection monitoring

**Protocol:**
- Request: `{"jsonrpc": "2.0", "method": "ping", "id": 1}`
- Response: `{"jsonrpc": "2.0", "result": {}, "id": 1}`

**Use Cases:**
- Keep-alive for idle connections
- Health monitoring
- Failover detection
- Load balancing decisions
- Connection validation
- Latency measurement

**Files:**
- `mcp/ping.go` - Documentation
- `client/ping.go` - Client ping method
- `examples/ping/main.go` - Full demonstration

### 22. Completion (Argument Autocompletion)
Location: `mcp/completion.go`, `server/completion.go`, `client/completion.go`

IDE-like argument autocompletion for prompts and resources (MCP 2024-11-05 specification):

**Core Types:**
```go
type CompletionRef struct {
    Type string // "ref/prompt" or "ref/resource"
    Name string // Name of prompt or resource
}

type CompletionArgument struct {
    Name  string // Argument name
    Value string // Partial value typed so far
}

type CompleteRequest struct {
    Ref      CompletionRef
    Argument CompletionArgument
}
```

**Server-Side Registration:**
```go
srv := server.New("my-server", server.WithCompletion())

// Register prompt completion
srv.RegisterPromptCompletion("code_review",
    func(ctx context.Context, ref mcp.CompletionRef, arg mcp.CompletionArgument) ([]string, error) {
        if arg.Name == "language" {
            languages := []string{"Go", "Python", "JavaScript"}
            // Filter by partial value
            var filtered []string
            for _, lang := range languages {
                if strings.HasPrefix(strings.ToLower(lang), strings.ToLower(arg.Value)) {
                    filtered = append(filtered, lang)
                }
            }
            return filtered, nil
        }
        return []string{}, nil
    },
)

// Register resource completion
srv.RegisterResourceCompletion("file:///", filePathHandler)
```

**Client-Side Usage:**
```go
ref := mcp.CompletionRef{
    Type: "ref/prompt",
    Name: "code_review",
}

arg := mcp.CompletionArgument{
    Name:  "language",
    Value: "Ja", // User typed "Ja"
}

suggestions, _ := client.GetCompletion(ctx, ref, arg)
// suggestions: ["Java", "JavaScript"]
```

**Features:**
- IDE-like autocomplete experience
- Filter by partial input
- Support for prompts and resources
- Case-insensitive matching
- Rich completion metadata (optional)
- Dynamic suggestion lists

**Reference Types:**
- `ref/prompt` - Complete prompt arguments
- `ref/resource` - Complete resource URIs

**Use Cases:**
- Programming language selection
- File path completion
- Configuration key suggestions
- Resource URI completion
- Enum value suggestions
- User name autocomplete

**Files:**
- `mcp/completion.go` - Core types
- `mcp/completion_test.go` - Type tests
- `server/completion.go` - Server completion manager
- `client/completion.go` - Client completion requester
- `examples/completion/main.go` - Full demonstration

## MCP 2025-03-26 Specification Updates

### 23. Tool Annotations
Location: `mcp/types.go`, `server/tool.go`, `builder/tool.go`

Enhanced tool metadata for better describing tool behavior (MCP 2025-03-26 specification):

**Tool Annotations:**
```go
type Tool struct {
    Name        string
    Description string
    InputSchema map[string]interface{}
    // 2025-03-26 annotations
    Title           string // Human-readable title
    ReadOnlyHint    *bool  // Tool doesn't modify environment
    DestructiveHint *bool  // Tool may perform destructive updates
    IdempotentHint  *bool  // Repeated calls have no additional effect
    OpenWorldHint   *bool  // Tool may interact with external entities
}
```

**Builder API:**
```go
tool := builder.NewTool("delete_file").
    Description("Delete a file").
    Title("File Deletion Tool").
    Destructive().           // Mark as destructive
    Idempotent().           // Safe to retry
    Handler(deleteFileFunc).
    Build()

readTool := builder.NewTool("read_file").
    Description("Read file contents").
    ReadOnly().             // Mark as read-only
    Handler(readFileFunc).
    Build()

apiTool := builder.NewTool("fetch_data").
    Description("Fetch data from API").
    OpenWorld().            // Interacts with external systems
    Handler(fetchFunc).
    Build()
```

**Features:**
- Declarative behavior hints for LLMs
- Better tool selection by AI agents
- Safety checks and validation
- UI affordances (confirmation dialogs for destructive operations)
- Better error messages

**Use Cases:**
- Mark read-only tools for parallel execution
- Require confirmation for destructive operations
- Indicate idempotent operations for retry logic
- Warn about external dependencies

**Files:**
- `mcp/types.go` - Tool type with annotations
- `server/tool.go` - ToolHandler with annotations
- `builder/tool.go` - Builder methods for annotations

### 24. Progress Messages
Location: `mcp/progress.go`, `server/progress.go`

Added optional message field to progress notifications (MCP 2025-03-26 specification):

**Enhanced Progress Type:**
```go
type ProgressNotification struct {
    ProgressToken interface{} // Unique token
    Progress      float64      // Current progress value
    Total         *float64     // Optional total value
    Message       string       `json:"message,omitempty"` // Descriptive status (2025-03-26)
}
```

**Server-Side Usage:**
```go
srv := server.New("my-server", server.WithProgress())

// Send progress with descriptive message
srv.NotifyWithMessage(
    "task-123",
    50.0,
    &total,
    "Processing file 50 of 100...",
)
```

**Features:**
- Descriptive status messages for better UX
- Context about current operation
- User-friendly progress updates
- Backward compatible (message is optional)

**Files:**
- `mcp/progress.go` - Enhanced ProgressNotification
- `server/progress.go` - NotifyWithMessage method

### 25. Audio Content Support
Location: `mcp/types.go`

Added audio content type for multimedia applications (MCP 2025-03-26 specification):

**Audio Content Type:**
```go
type AudioContent struct {
    Type     string `json:"type"`     // "audio"
    Data     string `json:"data"`     // Base64-encoded audio data
    MimeType string `json:"mimeType"` // e.g., "audio/wav", "audio/mp3"
}

func (a AudioContent) ContentType() string {
    return a.Type
}
```

**Usage:**
```go
audioContent := mcp.AudioContent{
    Type:     "audio",
    Data:     base64AudioData,
    MimeType: "audio/wav",
}
```

**Supported MIME Types:**
- `audio/wav` - WAV audio
- `audio/mp3` - MP3 audio
- `audio/mpeg` - MPEG audio
- `audio/ogg` - Ogg Vorbis
- `audio/webm` - WebM audio

**Use Cases:**
- Voice assistant responses
- Audio transcription
- Speech synthesis
- Audio analysis tools
- Multimedia content delivery

**Files:**
- `mcp/types.go` - AudioContent type

### 26. Completions Capability Declaration
Location: `mcp/types.go`, `server/server.go`

Added capability declaration for argument autocompletion (MCP 2025-03-26 specification):

**Capability Type:**
```go
type CompletionsCapability struct {
    // Empty struct indicating completion support
}

type ServerCapabilities struct {
    Tools       *ToolsCapability
    Resources   *ResourcesCapability
    Prompts     *PromptsCapability
    Completions *CompletionsCapability `json:"completions,omitempty"` // 2025-03-26
}
```

**Server Declaration:**
```go
srv := server.New("my-server", server.WithCompletion())

// Completions capability automatically declared in initialize response
// when WithCompletion() is used
```

**Features:**
- Explicit capability declaration during initialization
- Clients can detect completion support
- Enables autocomplete features in clients
- Backward compatible (omitted if not supported)

**Files:**
- `mcp/types.go` - CompletionsCapability type
- `server/server.go` - Capability declaration in initialize

### 27. JSON-RPC Batching
Location: `internal/jsonrpc/jsonrpc.go`, `server/server.go`, `client/client.go`

Support for sending multiple requests in a single batch (JSON-RPC 2.0 / MCP 2025-03-26):

**Core Features:**
```go
// Server-side automatic batch detection
func (s *Server) HandleBatch(ctx context.Context, messages []*mcp.Message) []*mcp.Message {
    responses := make([]*mcp.Message, 0, len(messages))
    for _, msg := range messages {
        response := s.HandleMessage(ctx, msg)
        if response != nil {
            responses = append(responses, response)
        }
    }
    return responses
}
```

**Client-Side Batching:**
```go
// Batch multiple requests
errors := client.BatchCall(ctx, []*client.BatchRequest{
    {Method: "tools/list", Result: &toolsResult},
    {Method: "resources/list", Result: &resourcesResult},
    {Method: "prompts/list", Result: &promptsResult},
})

// Check individual errors
for i, err := range errors {
    if err != nil {
        log.Printf("Request %d failed: %v", i, err)
    }
}
```

**JSON-RPC Format:**
```json
// Request (array)
[
  {"jsonrpc": "2.0", "method": "tools/list", "id": 1},
  {"jsonrpc": "2.0", "method": "resources/list", "id": 2},
  {"jsonrpc": "2.0", "method": "prompts/list", "id": 3}
]

// Response (array)
[
  {"jsonrpc": "2.0", "result": {"tools": [...]}, "id": 1},
  {"jsonrpc": "2.0", "result": {"resources": [...]}, "id": 2},
  {"jsonrpc": "2.0", "result": {"prompts": [...]}, "id": 3}
]
```

**Performance Benefits:**
- Reduces latency by 3x for 3 requests (single round-trip vs 3)
- Ideal for initial loads, dashboard refreshes, parallel operations
- Server processes independently, one failure doesn't affect others

**Features:**
- Automatic batch detection (JSON array vs object)
- Order-preserving responses
- Independent error handling per request
- No transaction semantics (each request independent)
- Notifications in batch receive no response

**Use Cases:**
- Initial application load (tools, resources, prompts)
- Dashboard data refresh
- Parallel independent operations
- Bulk data fetching
- Multi-tool execution

**Files:**
- `internal/jsonrpc/jsonrpc.go` - ReadAny(), ReadBatch(), WriteBatch()
- `internal/jsonrpc/batch_test.go` - Comprehensive batch tests
- `server/server.go` - HandleBatch() method
- `client/client.go` - BatchCall() method
- `examples/batching/main.go` - Full demonstration with examples

### 28. Streamable HTTP Transport
Location: `transport/streamhttp/streamhttp.go`

New unified HTTP transport for bi-directional communication (MCP 2025-03-26 specification):

**Key Features:**
- Single endpoint supporting POST and GET methods
- Session management with Mcp-Session-Id headers
- Stream resumption with Last-Event-ID
- Automatic batch request support
- Origin validation for security

**Architecture:**
```go
// Client-to-Server (POST)
POST /mcp HTTP/1.1
Content-Type: application/json
Mcp-Session-Id: a1b2c3d4...

{"jsonrpc":"2.0","method":"tools/list","id":1}

// Server-to-Client (GET - SSE Stream)
GET /mcp HTTP/1.1
Accept: text/event-stream
Mcp-Session-Id: a1b2c3d4...

HTTP/1.1 200 OK
Content-Type: text/event-stream

id: 1
data: {"jsonrpc":"2.0","method":"sampling/createMessage",...}
```

**Client Usage:**
```go
transport := streamhttp.New("http://localhost:8080/mcp")
conn, err := transport.Connect(ctx)
defer conn.Close()

// Session ID automatically managed after initialization
// POST and GET connections handled transparently
```

**Server Usage:**
```go
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Process MCP messages
})

server := streamhttp.NewServer(":8080", handler,
    streamhttp.WithAllowedOrigin("http://localhost:3000"),
)

server.ListenAndServe()
```

**Session Management:**
- Server assigns unique session ID during initialization
- Client includes session ID in all subsequent requests
- Cryptographically secure session IDs (UUID/hash)
- Session store for state management
- Automatic session timeout support

**Stream Resumption:**
- Event IDs for tracking message delivery
- Last-Event-ID header for resuming after disconnect
- Server buffers recent events for replay
- Prevents message loss on reconnection

**Advantages over HTTP+SSE:**
- Single unified endpoint (simpler architecture)
- Built-in session management
- Standardized resumption support
- Better scalability with unified state
- Native batch request support
- Improved security with origin validation

**Security Features:**
- Origin header validation
- Session ID validation
- Localhost binding support
- Secure session ID generation
- Connection timeout handling

**Features:**
- Bi-directional communication over HTTP
- Server-Sent Events for server-to-client push
- Chunked transfer encoding support
- Keep-alive for long-lived connections
- Error handling with appropriate HTTP status codes

**Use Cases:**
- Web browser applications
- Mobile apps with flaky connections
- Cloud services with server-to-server communication
- Real-time notifications and progress updates
- Long-running operations with bidirectional streaming

**Files:**
- `transport/streamhttp/streamhttp.go` - Core implementation
- `transport/streamhttp/streamhttp_test.go` - 15 comprehensive tests
- `examples/streamhttp/main.go` - Full demonstration with examples

### 29. OAuth 2.1 Authentication
Location: `auth/oauth21/oauth21.go`

Complete OAuth 2.1 authentication provider with mandatory PKCE support (MCP 2025-03-26 specification):

**OAuth 2.1 Improvements over OAuth 2.0:**
- **PKCE mandatory** for all clients (public and confidential)
- **Implicit grant removed** (insecure, token in URL)
- **Password grant removed** (credentials in request body)
- **Strict redirect URI matching** (exact string, no wildcards)
- **Refresh token rotation recommended**
- **S256 code challenge method** (SHA-256)

**PKCE (Proof Key for Code Exchange):**
```go
// Generate PKCE challenge (mandatory in OAuth 2.1)
challenge, err := oauth21.GeneratePKCEChallenge()
// Returns:
// - CodeVerifier: 43-128 char random string (kept secret)
// - CodeChallenge: BASE64URL(SHA256(code_verifier))
// - Method: "S256"

// Authorization URL with PKCE
state := "random-state-string"
authURL := provider.AuthCodeURLWithPKCE(state, challenge)
// Includes: code_challenge, code_challenge_method parameters

// Token exchange with PKCE verification
token, err := provider.ExchangeWithPKCE(ctx, code, state)
// Server verifies: SHA256(code_verifier) == code_challenge
```

**PKCE Flow:**
1. Client generates random `code_verifier` (43-128 chars)
2. Client computes `code_challenge = SHA256(code_verifier)`
3. Client sends `code_challenge` in authorization request
4. Server stores `code_challenge` with authorization code
5. Client exchanges code + `code_verifier` for token
6. Server verifies `SHA256(code_verifier) == code_challenge`

**Strict Redirect URI Validation:**
```go
// OAuth 2.1 requires exact string matching
err := provider.ValidateRedirectURI("http://localhost:8080/callback")

// ✓ Allowed (exact match):
//   Registered: http://localhost:8080/callback
//   Provided:   http://localhost:8080/callback

// ✗ Rejected (no wildcards):
//   Registered: http://localhost:8080/callback
//   Provided:   http://localhost:8080/callback/extra

// ✗ Rejected (no port mismatch):
//   Registered: http://localhost:8080/callback
//   Provided:   http://localhost:8081/callback
```

**Provider Setup:**
```go
// Google OAuth 2.1
provider := oauth21.New(
    oauth21.Google,
    "client-id",
    "client-secret",
    "http://localhost:8080/callback",
    []string{"email", "profile"},
)

// GitHub OAuth 2.1
provider := oauth21.New(
    oauth21.GitHub,
    "client-id",
    "client-secret",
    "http://localhost:8080/callback",
    []string{"user", "repo"},
)

// Custom provider
provider := oauth21.New(
    oauth21.Azure,
    "client-id",
    "client-secret",
    "http://localhost:8080/callback",
    []string{"openid", "email"},
    oauth21.WithCustomEndpoint(
        "https://custom.com/authorize",
        "https://custom.com/token",
    ),
    oauth21.WithUserInfoURL("https://custom.com/userinfo"),
)
```

**Complete Flow:**
```go
// 1. Generate PKCE challenge
challenge, _ := oauth21.GeneratePKCEChallenge()

// 2. Generate authorization URL
state := generateRandomState()
authURL := provider.AuthCodeURLWithPKCE(state, challenge)

// 3. Redirect user to authURL
http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)

// 4. Handle callback
func handleCallback(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    state := r.URL.Query().Get("state")

    // 5. Exchange code for token (with PKCE)
    token, err := provider.ExchangeWithPKCE(ctx, code, state)

    // 6. Get user info
    claims, err := provider.ValidateToken(ctx, token.AccessToken)

    // User authenticated: claims.Subject, claims.Email
}
```

**Security Benefits:**
- **PKCE Protection:**
  - Prevents authorization code interception attacks
  - Protects against man-in-the-middle attacks
  - No client secret needed for public clients (SPAs, mobile)
- **Strict Redirect URI:**
  - Prevents redirect URI manipulation
  - Eliminates open redirect vulnerabilities
  - No partial matching exploits
- **Removed Insecure Flows:**
  - No Implicit grant (prevents token exposure in URL)
  - No Password grant (prevents credential exposure)
  - Forces secure authorization code flow

**Client Types:**
- **Public Clients** (PKCE mandatory):
  - Single Page Applications (SPAs)
  - Mobile applications
  - Desktop applications
  - Cannot securely store client secret
  - PKCE provides protection without secret
- **Confidential Clients** (PKCE recommended):
  - Web servers
  - Backend services
  - Can securely store client secret
  - PKCE provides additional security layer

**OAuth 2.1 Features:**
```go
features := oauth21.GetOAuth21Features()
// Returns OAuth21Features{
//     MandatoryPKCE:         true,
//     StrictRedirectURI:     true,
//     ImplicitGrantRemoved:  true,
//     PasswordGrantRemoved:  true,
//     CodeChallengeMethod:   "S256",
//     MinimumVerifierLength: 43,
//     MaximumVerifierLength: 128,
// }
```

**Provider Options:**
```go
// Email verification
oauth21.WithVerifyEmail(true)

// Scope mapping
oauth21.WithScopeMapping(map[string][]string{
    "role": {"admin", "user"},
})

// Custom endpoints
oauth21.WithCustomEndpoint(authURL, tokenURL)

// Custom user info URL
oauth21.WithUserInfoURL(url)
```

**HTTP Middleware:**
```go
// Protect routes with OAuth
protected := provider.Middleware()(handler)

// Extracts Bearer token, validates, adds claims to context
claims := auth.GetClaims(r.Context())
```

**Migration from OAuth 2.0:**
1. **Add PKCE to all flows:**
   - Before: `provider.AuthCodeURL(state)`
   - After: `provider.AuthCodeURLWithPKCE(state, challenge)`
2. **Update token exchange:**
   - Before: `provider.Exchange(ctx, code)`
   - After: `provider.ExchangeWithPKCE(ctx, code, state)`
3. **Remove insecure flows:**
   - Remove: Implicit grant implementations
   - Remove: Password grant implementations
4. **Enforce strict redirect URIs:**
   - Update: Use exact string matching
   - Remove: Wildcard or pattern matching

**Best Practices:**
- **State Parameter:**
  - Generate cryptographically secure random state
  - Store state in session
  - Validate state on callback
  - Use once and discard
- **PKCE:**
  - Use S256 method (SHA-256)
  - Generate new verifier for each flow
  - Store verifier securely
  - Clear verifier after exchange
- **Redirect URIs:**
  - Pre-register all redirect URIs
  - Use HTTPS in production
  - Validate exact match
  - No wildcards or patterns
- **Token Handling:**
  - Store tokens securely
  - Use short-lived access tokens
  - Implement token refresh
  - Rotate refresh tokens

**Common Errors:**
- `invalid_request` - Missing code_challenge parameter (add PKCE)
- `invalid_grant` - code_verifier doesn't match (use same verifier)
- `redirect_uri_mismatch` - URI doesn't exactly match (exact string)

**Files:**
- `auth/oauth21/oauth21.go` - Complete OAuth 2.1 implementation
- `auth/oauth21/oauth21_test.go` - Comprehensive tests (17 tests)
- `examples/oauth21/main.go` - Full demonstration with examples

## Pending 2025-03-26 Features

All features from MCP 2025-03-26 specification have been implemented (7/7 complete).

## MCP 2025-06-18 Specification Updates

The MCP 2025-06-18 specification brings significant updates focused on type safety, user interaction, and protocol simplification.

### 30. Tool Output Schemas
Location: `mcp/types.go`, `server/tool.go`, `builder/tool.go`

Tools can now specify expected output structure using JSON Schema (MCP 2025-06-18):

**Type Definition:**
```go
type Tool struct {
    Name         string                 `json:"name"`
    InputSchema  map[string]interface{} `json:"inputSchema"`
    OutputSchema map[string]interface{} `json:"outputSchema,omitempty"` // 2025-06-18
    Description  string                 `json:"description,omitempty"`
    // ...
}
```

**Manual Schema:**
```go
tool := &ToolHandler{
    Name: "analyze_code",
    OutputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "complexity": map[string]interface{}{"type": "number"},
            "issues": map[string]interface{}{"type": "array"},
            "score": map[string]interface{}{
                "type": "number",
                "minimum": 0,
                "maximum": 100,
            },
        },
        "required": []string{"complexity", "issues", "score"},
    },
}
```

**Auto-Generated from Go Types:**
```go
type AnalysisResult struct {
    Complexity int      `json:"complexity"`
    Issues     []string `json:"issues"`
    Score      float64  `json:"score"`
}

tool := builder.NewTool("analyze_code").
    Handler(func(ctx context.Context, input CodeInput) (AnalysisResult, error) {
        // Implementation
    }).
    OutputSchemaFromType(AnalysisResult{}).  // Auto-generates schema!
    Build()
```

**Benefits:**
- Clients know output structure before calling tool
- LLMs can better understand and use tool results
- Enables validation and type checking
- Improves developer experience with autocomplete
- Better error detection

**Use Cases:**
- Structured data analysis tools
- API integration tools with known responses
- Code generation tools with typed output
- Data transformation tools
- Validation and testing tools

**Files:**
- `mcp/types.go` - Tool type with OutputSchema field
- `server/tool.go` - ToolHandler with OutputSchema support
- `builder/tool.go` - OutputSchema() and OutputSchemaFromType() methods

### 31. Title Fields
Location: `mcp/types.go`

Human-friendly display names for resources, prompts, and templates (MCP 2025-06-18):

**Type Updates:**
```go
type Resource struct {
    URI   string `json:"uri"`
    Name  string `json:"name"`             // Technical ID
    Title string `json:"title,omitempty"`  // Display name (2025-06-18)
    // ...
}

type Prompt struct {
    Name  string `json:"name"`             // Technical ID
    Title string `json:"title,omitempty"`  // Display name (2025-06-18)
    // ...
}

type ResourceTemplate struct {
    Name  string `json:"name"`             // Technical ID
    Title string `json:"title,omitempty"`  // Display name (2025-06-18)
    // ...
}
```

**Usage:**
```go
resource := &Resource{
    URI:   "file:///docs/api.md",
    Name:  "api_docs",                     // For technical reference
    Title: "API Documentation",            // For display
}

prompt := &Prompt{
    Name:  "code_review",                  // For technical reference
    Title: "Code Review Assistant",        // For display
}
```

**Benefits:**
- Better UX in client applications
- Separate technical IDs from display text
- Easier internationalization
- Improved discoverability
- Clearer user interfaces

**Files:**
- `mcp/types.go` - Updated Resource, Prompt, ResourceTemplate types

### 32. Metadata (_meta) Fields
Location: `mcp/types.go`

Extensible metadata for resources, prompts, and templates (MCP 2025-06-18):

**Type Updates:**
```go
type Resource struct {
    URI  string                 `json:"uri"`
    Name string                 `json:"name"`
    Meta map[string]interface{} `json:"_meta,omitempty"` // 2025-06-18
    // ...
}
```

**Usage:**
```go
resource := &Resource{
    URI:  "file:///project/config.json",
    Name: "project_config",
    Meta: map[string]interface{}{
        "lastModified": "2025-06-18T10:30:00Z",
        "audience":     []string{"user", "assistant"},
        "priority":     0.8,
        "version":      "2.0",
        "author":       "dev-team",
        "tags":         []string{"config", "production"},
    },
}
```

**Common Metadata Fields:**
- `lastModified`: ISO 8601 timestamp for cache management
- `audience`: `["user", "assistant"]` for targeting
- `priority`: `0.0-1.0` for ordering/ranking
- `version`: Semantic version string
- `author`: Creator identification
- `tags`: Categorization and filtering
- Custom application-specific fields

**Benefits:**
- Extensible without breaking protocol
- Version tracking and cache management
- Audience targeting
- Priority-based ordering
- Rich application metadata

**Files:**
- `mcp/types.go` - Meta fields on Resource, Prompt, ResourceTemplate

### 33. Resource Links in Tool Results
Location: `mcp/types.go`

Tools can return resource links to connect results with data (MCP 2025-06-18):

**Type Definition:**
```go
type ResourceLinkContent struct {
    Type        string                 `json:"type"` // "resource"
    Resource    Resource               `json:"resource"`
    Annotations map[string]interface{} `json:"annotations,omitempty"`
}
```

**Usage:**
```go
// Tool returns text and resource link
result := []Content{
    TextContent{
        Type: "text",
        Text: "Analysis complete. Full report attached.",
    },
    ResourceLinkContent{
        Type: "resource",
        Resource: Resource{
            URI:      "file:///reports/analysis-2025-06-18.pdf",
            Name:     "analysis_report",
            Title:    "Code Analysis Report",
            MimeType: "application/pdf",
        },
        Annotations: map[string]interface{}{
            "generated": "2025-06-18T10:30:00Z",
            "size":      2048000,
            "pages":     42,
        },
    },
}
```

**Benefits:**
- Connect tool results to generated files
- Reference supplementary data
- Link to detailed reports
- Attach analysis artifacts
- Better result organization

**Use Cases:**
- Code analysis tools generating reports
- Data transformation tools creating files
- Testing tools with artifact output
- Documentation generators
- Export/conversion tools

**Files:**
- `mcp/types.go` - ResourceLinkContent type

### 34. Elicitation Capability
Location: `mcp/types.go`

Servers can request structured user input with JSON Schema validation (MCP 2025-06-18):

**Types:**
```go
type ElicitationCapability struct {
    // Empty struct indicating elicitation support
}

type ElicitationRequest struct {
    Schema      map[string]interface{} `json:"schema"`      // JSON Schema
    Description string                 `json:"description,omitempty"` // What for
}

type ElicitationResponse struct {
    Action string                 `json:"action"` // "accept", "decline", "cancel"
    Data   map[string]interface{} `json:"data,omitempty"`   // User-provided data
}
```

**Server Request:**
```go
request := ElicitationRequest{
    Description: "Please provide API configuration",
    Schema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "api_key": map[string]interface{}{
                "type":        "string",
                "description": "Your API key",
                "minLength":   20,
            },
            "region": map[string]interface{}{
                "type": "string",
                "enum": []string{"us-east", "us-west", "eu"},
            },
        },
        "required": []string{"api_key"},
    },
}
```

**User Response:**
```go
// Accept with data
response := ElicitationResponse{
    Action: "accept",
    Data: map[string]interface{}{
        "api_key": "sk-...",
        "region":  "us-east",
    },
}

// Decline
response := ElicitationResponse{
    Action: "decline",
}

// Cancel
response := ElicitationResponse{
    Action: "cancel",
}
```

**Schema Restrictions:**
- Primitive types only: string, number, boolean, enum
- Flat object structures (no deep nesting)
- Validation constraints: min/max, formats, patterns

**Security Guidelines:**
- Servers MUST NOT request sensitive info without justification
- Clients SHOULD show clear server identification
- Clients SHOULD allow response review
- Clients SHOULD offer decline/cancel options
- Clients SHOULD implement rate limiting

**Use Cases:**
- API key configuration
- Service credentials
- User preferences
- Configuration parameters
- Dynamic form input

**Files:**
- `mcp/types.go` - Elicitation types

### 35. Protocol Version Update
Location: `server/server.go`, `client/client.go`, tests

Updated protocol version to 2025-06-18:

**Server:**
```go
result := map[string]interface{}{
    "protocolVersion": "2025-06-18",  // Updated from 2024-11-05
    "capabilities":    caps,
    "serverInfo": map[string]string{
        "name":    s.name,
        "version": s.version,
    },
}
```

**Client:**
```go
err := c.call(ctx, "initialize", map[string]interface{}{
    "protocolVersion": "2025-06-18",  // Updated from 2024-11-05
    "capabilities":    capabilities,
    "clientInfo": map[string]string{
        "name":    "fullmcp-client",
        "version": "0.1.0",
    },
})
```

**Version Negotiation:**
1. Client sends latest supported version
2. Server responds with same or alternative version
3. Client disconnects if incompatible

**Files:**
- `server/server.go` - Initialize response with 2025-06-18
- `client/client.go` - Initialize request with 2025-06-18
- All test files updated to 2025-06-18

### 36. Breaking Change: JSON-RPC Batching Removed
Location: `internal/jsonrpc/jsonrpc.go`, `server/server.go`, `client/client.go`

JSON-RPC batching support was removed in MCP 2025-06-18:

**Removed Methods:**
```go
// REMOVED from MessageReader:
func (mr *MessageReader) ReadBatch() ([]*mcp.Message, error)
func (mr *MessageReader) ReadAny() ([]*mcp.Message, bool, error)

// REMOVED from MessageWriter:
func (mw *MessageWriter) WriteBatch(messages []*mcp.Message) error

// REMOVED from Server:
func (s *Server) HandleBatch(ctx context.Context, messages []*mcp.Message) []*mcp.Message

// REMOVED from Client:
type BatchRequest
func (c *Client) BatchCall(ctx context.Context, requests []*BatchRequest) []error
```

**Migration:**
Send individual requests instead of batches:
```go
// Before (2025-03-26):
errors := client.BatchCall(ctx, []*BatchRequest{
    {Method: "tools/list", Result: &toolsResult},
    {Method: "resources/list", Result: &resourcesResult},
})

// After (2025-06-18):
toolsResult, err1 := client.ListTools(ctx)
resourcesResult, err2 := client.ListResources(ctx)
```

**Rationale:**
- Simplifies protocol
- Reduces implementation complexity
- Better request tracking
- Clearer error semantics
- Easier debugging

**Files:**
- `internal/jsonrpc/jsonrpc.go` - Removed batch methods
- `server/server.go` - Removed HandleBatch()
- `client/client.go` - Removed BatchCall()
- `internal/jsonrpc/batch_test.go` - Deleted
- `examples/batching/` - Deleted

## MCP 2025-06-18 Summary

**Additions (6 new features):**
1. ✅ Tool output schemas
2. ✅ Title fields (Resource, Prompt, Template)
3. ✅ _meta fields (Resource, Prompt, Template)
4. ✅ Resource links in tool results
5. ✅ Elicitation capability
6. ✅ Protocol version updated to 2025-06-18

**Breaking Changes (1):**
1. ❌ JSON-RPC batching removed

**Files Created:**
- `examples/mcp-2025-06-18/main.go` - Comprehensive demonstration

**Files Modified:**
- `mcp/types.go` - Added output schemas, title, _meta, resource links, elicitation
- `server/tool.go` - Added OutputSchema support
- `builder/tool.go` - Added OutputSchema() and OutputSchemaFromType()
- `server/server.go` - Updated protocol version, removed batching
- `client/client.go` - Updated protocol version, removed batching
- `internal/jsonrpc/jsonrpc.go` - Removed batch methods
- `CLAUDE.md` - Updated protocol version and features
- `README.md` - Updated to reflect 2025-06-18
- All test files - Updated protocol version

## Testing & Quality Assurance

### Test Improvements (Latest Session)

#### Proxy Server Tests
Location: `server/proxy/proxy_test.go`

**Improvements:**
- Fixed async initialization race conditions
- Implemented channel-based mock transport for reliable bidirectional communication
- Added proper connection establishment with timeouts
- Implemented proper cleanup and graceful shutdown
- All tests now pass reliably without timing issues

**Test Coverage:**
- Basic proxy functionality
- Multiple capability proxying (tools, resources, prompts)
- Empty backend handling
- Error handling with failed backends

#### Performance Benchmarks
Locations: `server/benchmark_test.go`, `client/benchmark_test.go`, `builder/benchmark_test.go`

**Comprehensive benchmarks for:**

**Server Operations:**
- Tool registration: ~83 ns/op
- Tool call: ~33 ns/op
- Resource read: ~22 ns/op
- Message handling: ~1,283 ns/op
- Prompt get: ~86 ns/op
- List tools (100 tools): ~2,292 ns/op
- List resources (100 resources): ~2,410 ns/op
- Middleware chain: ~877 ns/op

**Client Operations:**
- Connection: ~25,403 ns/op
- List tools: ~5,113 ns/op
- Call tool: ~8,166 ns/op
- Read resource: ~7,826 ns/op
- Concurrent calls: ~3,943 ns/op

**Builder Operations:**
- Simple tool build: ~14,488 ns/op
- Complex tool build: ~26,279 ns/op
- Resource build: ~3.3 ns/op
- Resource template build: ~13 ns/op
- Prompt build: ~22 ns/op

**Run benchmarks:**
```bash
go test -bench=. -benchmem ./server ./client ./builder
```

#### Integration Tests
Location: `integration_test.go`

**End-to-end test scenarios:**
- Complete tool workflow (list, call multiple tools)
- Resource workflow (list, read resources)
- Concurrent clients (5 clients × 10 operations each)
- Error handling (tool failures, non-existent resources)
- Complex workflow (tools + resources combined)

**All integration tests pass reliably with proper synchronization and cleanup.**

**Run integration tests:**
```bash
go test -v -run=TestIntegration ./...
```

## Next Steps

### High Priority
1. ✅ Improve proxy server tests (async initialization) - **COMPLETED**
2. ✅ Performance benchmarks - **COMPLETED**
3. ✅ Integration tests - **COMPLETED**

### Medium Priority
1. ✅ Comprehensive documentation - **COMPLETED**
2. ✅ Advanced sampling features - **COMPLETED**
3. ✅ Roots (filesystem boundaries) - **COMPLETED**

### Low Priority
1. ✅ Logging protocol extensions - **COMPLETED**
2. ✅ Progress notifications - **COMPLETED**
3. ✅ Cancellation support - **COMPLETED**

### Additional Features
1. ✅ Ping utility - **COMPLETED**
2. ✅ Completion/complete (argument autocompletion) - **COMPLETED**

## Migration from Spec

The implementation follows the specification with some additions:

1. **Enhanced Builder Patterns**: More fluent API options than spec
2. **Server Context**: Additional helper methods for common operations
3. **Middleware**: Standard middleware patterns for Go HTTP servers
4. **Resource Templates**: Simplified API for common use cases

## Dependencies

Current dependencies (from `go.mod`):

```
github.com/invopop/jsonschema v0.12.0       // JSON Schema generation
github.com/gorilla/websocket v1.5.3         // WebSocket transport
github.com/golang-jwt/jwt/v5 v5.3.0         // JWT authentication
golang.org/x/oauth2 v0.31.0                 // OAuth 2.0 authentication
github.com/spf13/cobra v1.10.1              // CLI framework
```

All features are production-ready with comprehensive test coverage.
