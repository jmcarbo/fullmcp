# CLAUDE.md

This file provides guidance to AI assistants when working with code in this repository.

## Project Overview

FullMCP is a production-ready Golang implementation of the Model Context Protocol (MCP), inspired by fastmcp. It provides both client and server implementations with support for Tools, Resources, and Prompts.

**Module:** `github.com/jmcarbo/fullmcp`
**Go Version:** 1.21+

## Development Commands

### Build
```bash
go build ./...
```

### Run Tests
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests in a specific package
go test ./server
go test ./client
go test ./builder

# Run a specific test
go test -run TestName ./package
```

### Run Example Server
```bash
cd examples/basic-server
go run main.go
```

### Dependency Management
```bash
# Add a new dependency
go get <package>

# Tidy dependencies
go mod tidy
```

## Architecture

### Core Components

**mcp/** - Protocol types and error definitions
- `types.go`: Core MCP types (Tool, Resource, Prompt, Message, Capabilities)
- `errors.go`: MCP error codes and error types

**server/** - Server implementation
- `server.go`: Main server with JSON-RPC message handling
- `tool.go`: ToolManager for registering and executing tools
- `resource.go`: ResourceManager for managing read-only data sources
- `prompt.go`: PromptManager for reusable message templates
- `transport.go`: Transport abstraction (stdio by default)

**client/** - Client implementation
- `client.go`: MCP client with async message handling using goroutines and channels
- Methods: `ListTools()`, `CallTool()`, `ListResources()`, `ReadResource()`, `ListPrompts()`, `GetPrompt()`

**builder/** - Fluent API for creating MCP entities
- `tool.go`: ToolBuilder with reflection-based schema generation using `jsonschema` package
- Automatically generates JSON schemas from Go struct types

**transport/stdio/** - Stdio transport implementation
- Used for communication over stdin/stdout

**internal/jsonrpc/** - JSON-RPC 2.0 implementation
- `MessageReader` and `MessageWriter` for protocol communication

### Key Design Patterns

1. **Manager Pattern**: ToolManager, ResourceManager, PromptManager handle registration and execution
2. **Builder Pattern**: Fluent APIs for constructing tools with automatic schema generation
3. **Reflection-based Schema Generation**: Tool input types automatically generate JSON schemas
4. **Concurrent Message Handling**: Client uses goroutines with channel-based pending request tracking

### Protocol Flow

**Server**: Listens for JSON-RPC messages → Routes to handler (initialize, tools/list, tools/call, resources/read, etc.) → Returns JSON-RPC response

**Client**: Sends JSON-RPC request with unique ID → Stores pending channel in map → Background goroutine routes response to channel → Returns result

## Important Implementation Notes

- All tool handlers must accept `context.Context` as first parameter
- Tool builder automatically generates JSON schema from second parameter type
- Server uses stdio transport by default; custom transports implement `io.ReadWriteCloser`
- Client manages pending requests using atomic counter and concurrent-safe map
- MCP protocol version: `2025-06-18`
- Tool output schemas: Tools can specify expected output structure (2025-06-18)
- Elicitation: Servers can request structured user input (2025-06-18)
- Resource metadata: _meta fields for version tracking and audience targeting (2025-06-18)
- Title fields: Human-friendly display names for resources and prompts (2025-06-18)
- Commits should follow conventional commit conventions (breaking:, feat:, fix:, etc.)
