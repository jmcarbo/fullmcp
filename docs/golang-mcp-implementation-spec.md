Git repository: https://github.com/jmcarbo/fastmcp

# Golang MCP Client/Server Library Implementation Specification

## Executive Summary

This document provides a comprehensive specification for implementing a Golang MCP (Model Context Protocol) client/server library that mirrors the functionality of [fastmcp](https://github.com/jlowin/fastmcp). The library will provide idiomatic Go interfaces for building MCP servers and clients with support for tools, resources, prompts, authentication, multiple transports, and advanced features.

## Table of Contents

1. [Overview](#overview)
2. [Core MCP Protocol Requirements](#core-mcp-protocol-requirements)
3. [Architecture Design](#architecture-design)
4. [Package Structure](#package-structure)
5. [Core Components](#core-components)
6. [Server Implementation](#server-implementation)
7. [Client Implementation](#client-implementation)
8. [Transport Layer](#transport-layer)
9. [Authentication](#authentication)
10. [Advanced Features](#advanced-features)
11. [Testing Strategy](#testing-strategy)
12. [Migration Path](#migration-path)

---

## 1. Overview

### 1.1 Project Goals

- **Idiomatic Go**: Follow Go conventions and best practices
- **Feature Parity**: Match fastmcp's capabilities
- **Type Safety**: Leverage Go's static typing
- **Performance**: Optimize for concurrent operations
- **Simplicity**: Minimal boilerplate for common use cases
- **Extensibility**: Support advanced patterns and customization

### 1.2 Target Use Cases

1. Building MCP servers that expose tools, resources, and prompts
2. Creating MCP clients to interact with servers
3. Proxy servers and server composition
4. Enterprise authentication and authorization
5. Multi-transport deployment (stdio, HTTP, SSE)

---

## 2. Core MCP Protocol Requirements

### 2.1 Protocol Specification

**Version**: 2025-06-18 (latest)

**Base Protocol**: JSON-RPC 2.0

**Architecture**: Client-Server with stateful connections

### 2.2 Message Types

```go
// JSON-RPC 2.0 message envelope
type Message struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      interface{}     `json:"id,omitempty"`
    Method  string          `json:"method,omitempty"`
    Params  json.RawMessage `json:"params,omitempty"`
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

### 2.3 Core Primitives

1. **Tools**: Executable functions that AI models can invoke
2. **Resources**: Read-only data sources (files, databases, APIs)
3. **Prompts**: Reusable message templates and workflows

### 2.4 Required Methods

**Server Methods**:
- `initialize` - Capability negotiation
- `tools/list` - List available tools
- `tools/call` - Execute a tool
- `resources/list` - List available resources
- `resources/read` - Read resource content
- `resources/templates/list` - List resource templates
- `prompts/list` - List available prompts
- `prompts/get` - Get prompt with arguments

**Client Methods**:
- `ping` - Keepalive
- `notifications/initialized` - Post-init notification

**Optional Features**:
- Sampling (LLM text generation)
- Roots (filesystem boundaries)
- Logging
- Progress notifications
- Cancellation

---

## 3. Architecture Design

### 3.1 Layered Architecture

```
┌─────────────────────────────────────┐
│   Application Layer                 │
│   (User Code)                       │
└─────────────────────────────────────┘
           ↓
┌─────────────────────────────────────┐
│   High-Level API                    │
│   (Server/Client Builders)          │
└─────────────────────────────────────┘
           ↓
┌─────────────────────────────────────┐
│   Core Protocol Layer               │
│   (Tools, Resources, Prompts)       │
└─────────────────────────────────────┘
           ↓
┌─────────────────────────────────────┐
│   JSON-RPC Layer                    │
│   (Message Handling)                │
└─────────────────────────────────────┘
           ↓
┌─────────────────────────────────────┐
│   Transport Layer                   │
│   (stdio, HTTP, SSE, WebSocket)     │
└─────────────────────────────────────┘
```

### 3.2 Design Principles

1. **Interface-First**: Define interfaces before implementations
2. **Context Propagation**: Use `context.Context` throughout
3. **Error Handling**: Explicit error returns, typed errors
4. **Concurrency**: Safe for concurrent use
5. **Testability**: Dependency injection, mock-friendly
6. **Extensibility**: Plugin architecture for custom components

---

## 4. Package Structure

```
fullmcp/
├── README.md
├── go.mod
├── go.sum
│
├── mcp/                    # Core protocol types
│   ├── types.go           # Protocol message types
│   ├── errors.go          # Error types
│   └── schema.go          # JSON Schema helpers
│
├── server/                 # Server implementation
│   ├── server.go          # Main server type
│   ├── handler.go         # Protocol handlers
│   ├── tool.go            # Tool management
│   ├── resource.go        # Resource management
│   ├── prompt.go          # Prompt management
│   ├── middleware.go      # Middleware support
│   ├── context.go         # Server context
│   └── lifecycle.go       # Lifecycle hooks
│
├── client/                 # Client implementation
│   ├── client.go          # Main client type
│   ├── session.go         # Session management
│   ├── transport.go       # Transport interface
│   └── options.go         # Client options
│
├── transport/              # Transport implementations
│   ├── stdio/
│   │   └── stdio.go       # Stdio transport
│   ├── http/
│   │   ├── streamable.go  # Streamable HTTP
│   │   └── sse.go         # Server-Sent Events
│   └── websocket/
│       └── websocket.go   # WebSocket transport
│
├── auth/                   # Authentication
│   ├── auth.go            # Auth interface
│   ├── oauth/
│   │   ├── google.go      # Google OAuth
│   │   ├── github.go      # GitHub OAuth
│   │   └── azure.go       # Azure OAuth
│   ├── jwt/
│   │   └── jwt.go         # JWT authentication
│   └── apikey/
│       └── apikey.go      # API key auth
│
├── registry/               # Component registry
│   ├── tool.go            # Tool registry
│   ├── resource.go        # Resource registry
│   └── prompt.go          # Prompt registry
│
├── builder/                # Fluent builders
│   ├── server.go          # Server builder
│   ├── tool.go            # Tool builder
│   ├── resource.go        # Resource builder
│   └── prompt.go          # Prompt builder
│
├── codec/                  # Encoding/decoding
│   ├── json.go            # JSON codec
│   └── schema.go          # Schema generation
│
├── util/                   # Utilities
│   ├── logging.go         # Logging interface
│   ├── reflection.go      # Reflection helpers
│   └── validation.go      # Validation helpers
│
├── examples/               # Example implementations
│   ├── basic-server/
│   ├── basic-client/
│   ├── oauth-server/
│   └── proxy-server/
│
└── internal/               # Internal packages
    ├── jsonrpc/           # JSON-RPC implementation
    └── testutil/          # Testing utilities
```

---

## 5. Core Components

### 5.1 Protocol Types

```go
package mcp

import (
    "encoding/json"
    "time"
)

// Content represents MCP content blocks
type Content interface {
    ContentType() string
}

type TextContent struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

type ImageContent struct {
    Type     string `json:"type"`
    Data     string `json:"data"`
    MimeType string `json:"mimeType"`
}

type ResourceContent struct {
    Type     string `json:"type"`
    URI      string `json:"uri"`
    MimeType string `json:"mimeType,omitempty"`
    Text     string `json:"text,omitempty"`
}

// Tool represents an MCP tool
type Tool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description,omitempty"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

// Resource represents an MCP resource
type Resource struct {
    URI         string `json:"uri"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    MimeType    string `json:"mimeType,omitempty"`
}

// ResourceTemplate for parameterized resources
type ResourceTemplate struct {
    URITemplate string `json:"uriTemplate"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    MimeType    string `json:"mimeType,omitempty"`
}

// Prompt represents an MCP prompt
type Prompt struct {
    Name        string           `json:"name"`
    Description string           `json:"description,omitempty"`
    Arguments   []PromptArgument `json:"arguments,omitempty"`
}

type PromptArgument struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Required    bool   `json:"required,omitempty"`
}

// PromptMessage for prompt responses
type PromptMessage struct {
    Role    string   `json:"role"`
    Content []Content `json:"content"`
}
```

### 5.2 Error Types

```go
package mcp

import "fmt"

type ErrorCode int

const (
    ParseError     ErrorCode = -32700
    InvalidRequest ErrorCode = -32600
    MethodNotFound ErrorCode = -32601
    InvalidParams  ErrorCode = -32602
    InternalError  ErrorCode = -32603
)

type MCPError struct {
    Code    ErrorCode
    Message string
    Data    interface{}
}

func (e *MCPError) Error() string {
    return fmt.Sprintf("MCP error %d: %s", e.Code, e.Message)
}

// Typed errors
type NotFoundError struct {
    Type string // "tool", "resource", "prompt"
    Name string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found: %s", e.Type, e.Name)
}

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}
```

---

## 6. Server Implementation

### 6.1 Core Server Type

```go
package server

import (
    "context"
    "sync"

    "github.com/yourusername/fullmcp/mcp"
)

// Server is the main MCP server
type Server struct {
    name         string
    version      string
    instructions string

    tools     *ToolManager
    resources *ResourceManager
    prompts   *PromptManager

    middleware []Middleware
    auth       AuthProvider

    lifespan   LifespanFunc
    mu         sync.RWMutex
}

// ServerOption configures a Server
type ServerOption func(*Server)

// New creates a new MCP server
func New(name string, opts ...ServerOption) *Server {
    s := &Server{
        name:      name,
        tools:     NewToolManager(),
        resources: NewResourceManager(),
        prompts:   NewPromptManager(),
    }

    for _, opt := range opts {
        opt(s)
    }

    return s
}

// Common options
func WithVersion(version string) ServerOption {
    return func(s *Server) {
        s.version = version
    }
}

func WithInstructions(instructions string) ServerOption {
    return func(s *Server) {
        s.instructions = instructions
    }
}

func WithAuth(auth AuthProvider) ServerOption {
    return func(s *Server) {
        s.auth = auth
    }
}

func WithMiddleware(mw ...Middleware) ServerOption {
    return func(s *Server) {
        s.middleware = append(s.middleware, mw...)
    }
}
```

### 6.2 Tool Management

```go
package server

import (
    "context"
    "encoding/json"
    "fmt"
    "reflect"

    "github.com/yourusername/fullmcp/mcp"
)

// ToolFunc is a function that can be registered as a tool
type ToolFunc func(context.Context, json.RawMessage) (interface{}, error)

// ToolHandler wraps a tool function with metadata
type ToolHandler struct {
    Name        string
    Description string
    Schema      map[string]interface{}
    Handler     ToolFunc
    Tags        []string
}

// ToolManager manages tool registration and execution
type ToolManager struct {
    tools map[string]*ToolHandler
    mu    sync.RWMutex
}

func NewToolManager() *ToolManager {
    return &ToolManager{
        tools: make(map[string]*ToolHandler),
    }
}

// Register registers a tool
func (tm *ToolManager) Register(handler *ToolHandler) error {
    tm.mu.Lock()
    defer tm.mu.Unlock()

    if _, exists := tm.tools[handler.Name]; exists {
        return fmt.Errorf("tool already registered: %s", handler.Name)
    }

    tm.tools[handler.Name] = handler
    return nil
}

// Call executes a tool
func (tm *ToolManager) Call(ctx context.Context, name string, args json.RawMessage) (interface{}, error) {
    tm.mu.RLock()
    handler, exists := tm.tools[name]
    tm.mu.RUnlock()

    if !exists {
        return nil, &mcp.NotFoundError{Type: "tool", Name: name}
    }

    return handler.Handler(ctx, args)
}

// List returns all registered tools
func (tm *ToolManager) List() []*mcp.Tool {
    tm.mu.RLock()
    defer tm.mu.RUnlock()

    tools := make([]*mcp.Tool, 0, len(tm.tools))
    for _, handler := range tm.tools {
        tools = append(tools, &mcp.Tool{
            Name:        handler.Name,
            Description: handler.Description,
            InputSchema: handler.Schema,
        })
    }

    return tools
}
```

### 6.3 Tool Builder with Reflection

```go
package builder

import (
    "context"
    "encoding/json"
    "fmt"
    "reflect"

    "github.com/yourusername/fullmcp/server"
    "github.com/invopop/jsonschema"
)

// ToolBuilder creates tools from functions
type ToolBuilder struct {
    name        string
    description string
    fn          interface{}
    tags        []string
}

func NewTool(name string) *ToolBuilder {
    return &ToolBuilder{name: name}
}

func (tb *ToolBuilder) Description(desc string) *ToolBuilder {
    tb.description = desc
    return tb
}

func (tb *ToolBuilder) Handler(fn interface{}) *ToolBuilder {
    tb.fn = fn
    return tb
}

func (tb *ToolBuilder) Tags(tags ...string) *ToolBuilder {
    tb.tags = tags
    return tb
}

// Build creates the ToolHandler
func (tb *ToolBuilder) Build() (*server.ToolHandler, error) {
    if tb.fn == nil {
        return nil, fmt.Errorf("handler function is required")
    }

    fnType := reflect.TypeOf(tb.fn)
    if fnType.Kind() != reflect.Func {
        return nil, fmt.Errorf("handler must be a function")
    }

    // Validate signature
    if fnType.NumIn() < 1 {
        return nil, fmt.Errorf("handler must accept at least context.Context")
    }

    // First arg must be context.Context
    if !fnType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
        return nil, fmt.Errorf("first argument must be context.Context")
    }

    // Generate JSON schema from input type
    var schema map[string]interface{}
    if fnType.NumIn() > 1 {
        inputType := fnType.In(1)
        reflector := jsonschema.Reflector{}
        jsonSchema := reflector.Reflect(reflect.New(inputType).Interface())
        schemaBytes, _ := json.Marshal(jsonSchema)
        json.Unmarshal(schemaBytes, &schema)
    }

    // Create wrapper function
    handler := func(ctx context.Context, args json.RawMessage) (interface{}, error) {
        fnValue := reflect.ValueOf(tb.fn)

        callArgs := []reflect.Value{reflect.ValueOf(ctx)}

        if fnType.NumIn() > 1 {
            inputType := fnType.In(1)
            input := reflect.New(inputType).Interface()
            if err := json.Unmarshal(args, input); err != nil {
                return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
            }
            callArgs = append(callArgs, reflect.ValueOf(input).Elem())
        }

        results := fnValue.Call(callArgs)

        // Handle return values (result, error)
        if len(results) == 2 {
            if !results[1].IsNil() {
                return nil, results[1].Interface().(error)
            }
            return results[0].Interface(), nil
        }

        return nil, fmt.Errorf("invalid handler signature")
    }

    return &server.ToolHandler{
        Name:        tb.name,
        Description: tb.description,
        Schema:      schema,
        Handler:     handler,
        Tags:        tb.tags,
    }, nil
}
```

### 6.4 Fluent Server API

```go
package server

// AddTool registers a tool using the builder pattern
func (s *Server) AddTool(name string, fn interface{}, opts ...ToolOption) error {
    builder := NewToolBuilder(name, fn)

    for _, opt := range opts {
        opt(builder)
    }

    handler, err := builder.Build()
    if err != nil {
        return err
    }

    return s.tools.Register(handler)
}

// Tool creates a tool builder for more control
func (s *Server) Tool(name string) *ToolBuilder {
    return &ToolBuilder{
        server: s,
        name:   name,
    }
}

// Example usage:
// server.Tool("multiply").
//     Description("Multiply two numbers").
//     Handler(func(ctx context.Context, input MultiplyInput) (float64, error) {
//         return input.A * input.B, nil
//     }).
//     Register()
```

### 6.5 Resource Management

```go
package server

import (
    "context"
    "sync"

    "github.com/yourusername/fullmcp/mcp"
)

// ResourceFunc reads resource content
type ResourceFunc func(context.Context) ([]byte, error)

// ResourceHandler wraps a resource function
type ResourceHandler struct {
    URI         string
    Name        string
    Description string
    MimeType    string
    Reader      ResourceFunc
    Tags        []string
}

// ResourceManager manages resources
type ResourceManager struct {
    resources map[string]*ResourceHandler
    templates map[string]*ResourceTemplateHandler
    mu        sync.RWMutex
}

func NewResourceManager() *ResourceManager {
    return &ResourceManager{
        resources: make(map[string]*ResourceHandler),
        templates: make(map[string]*ResourceTemplateHandler),
    }
}

// Register registers a resource
func (rm *ResourceManager) Register(handler *ResourceHandler) error {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    rm.resources[handler.URI] = handler
    return nil
}

// Read reads a resource
func (rm *ResourceManager) Read(ctx context.Context, uri string) ([]byte, error) {
    rm.mu.RLock()
    handler, exists := rm.resources[uri]
    rm.mu.RUnlock()

    if !exists {
        return nil, &mcp.NotFoundError{Type: "resource", Name: uri}
    }

    return handler.Reader(ctx)
}

// List returns all resources
func (rm *ResourceManager) List() []*mcp.Resource {
    rm.mu.RLock()
    defer rm.mu.RUnlock()

    resources := make([]*mcp.Resource, 0, len(rm.resources))
    for _, handler := range rm.resources {
        resources = append(resources, &mcp.Resource{
            URI:         handler.URI,
            Name:        handler.Name,
            Description: handler.Description,
            MimeType:    handler.MimeType,
        })
    }

    return resources
}
```

### 6.6 Prompt Management

```go
package server

import (
    "context"
    "sync"

    "github.com/yourusername/fullmcp/mcp"
)

// PromptFunc renders a prompt
type PromptFunc func(context.Context, map[string]interface{}) ([]*mcp.PromptMessage, error)

// PromptHandler wraps a prompt function
type PromptHandler struct {
    Name        string
    Description string
    Arguments   []mcp.PromptArgument
    Renderer    PromptFunc
    Tags        []string
}

// PromptManager manages prompts
type PromptManager struct {
    prompts map[string]*PromptHandler
    mu      sync.RWMutex
}

func NewPromptManager() *PromptManager {
    return &PromptManager{
        prompts: make(map[string]*PromptHandler),
    }
}

// Register registers a prompt
func (pm *PromptManager) Register(handler *PromptHandler) error {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    pm.prompts[handler.Name] = handler
    return nil
}

// Get renders a prompt
func (pm *PromptManager) Get(ctx context.Context, name string, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
    pm.mu.RLock()
    handler, exists := pm.prompts[name]
    pm.mu.RUnlock()

    if !exists {
        return nil, &mcp.NotFoundError{Type: "prompt", Name: name}
    }

    return handler.Renderer(ctx, args)
}

// List returns all prompts
func (pm *PromptManager) List() []*mcp.Prompt {
    pm.mu.RLock()
    defer pm.mu.RUnlock()

    prompts := make([]*mcp.Prompt, 0, len(pm.prompts))
    for _, handler := range pm.prompts {
        prompts = append(prompts, &mcp.Prompt{
            Name:        handler.Name,
            Description: handler.Description,
            Arguments:   handler.Arguments,
        })
    }

    return prompts
}
```

### 6.7 Middleware Support

```go
package server

import "context"

// Middleware wraps request handling
type Middleware func(next Handler) Handler

// Handler processes MCP requests
type Handler func(context.Context, *Request) (*Response, error)

// Request represents an MCP request
type Request struct {
    Method string
    Params interface{}
}

// Response represents an MCP response
type Response struct {
    Result interface{}
}

// Apply middleware chain
func (s *Server) applyMiddleware(handler Handler) Handler {
    for i := len(s.middleware) - 1; i >= 0; i-- {
        handler = s.middleware[i](handler)
    }
    return handler
}

// Example middleware: logging
func LoggingMiddleware(logger Logger) Middleware {
    return func(next Handler) Handler {
        return func(ctx context.Context, req *Request) (*Response, error) {
            logger.Infof("Request: %s", req.Method)
            resp, err := next(ctx, req)
            if err != nil {
                logger.Errorf("Error: %v", err)
            }
            return resp, err
        }
    }
}
```

---

## 7. Client Implementation

### 7.1 Core Client Type

```go
package client

import (
    "context"
    "encoding/json"
    "sync"
    "sync/atomic"

    "github.com/yourusername/fullmcp/mcp"
)

// Client is an MCP client
type Client struct {
    transport Transport
    session   *Session

    mu          sync.Mutex
    nextID      atomic.Int64
    pending     map[int64]chan *Response

    capabilities *mcp.ServerCapabilities
}

// ClientOption configures a Client
type ClientOption func(*Client)

// New creates a new MCP client
func New(transport Transport, opts ...ClientOption) *Client {
    c := &Client{
        transport: transport,
        pending:   make(map[int64]chan *Response),
    }

    for _, opt := range opts {
        opt(c)
    }

    return c
}

// Connect establishes a connection
func (c *Client) Connect(ctx context.Context) error {
    session, err := c.transport.Connect(ctx)
    if err != nil {
        return err
    }

    c.session = session

    // Start message handler
    go c.handleMessages()

    // Initialize
    result, err := c.initialize(ctx)
    if err != nil {
        return err
    }

    c.capabilities = result.Capabilities

    return nil
}

// Close closes the connection
func (c *Client) Close() error {
    if c.session != nil {
        return c.session.Close()
    }
    return nil
}
```

### 7.2 Client Methods

```go
package client

import (
    "context"
    "encoding/json"

    "github.com/yourusername/fullmcp/mcp"
)

// ListTools lists available tools
func (c *Client) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
    var result struct {
        Tools []*mcp.Tool `json:"tools"`
    }

    if err := c.call(ctx, "tools/list", nil, &result); err != nil {
        return nil, err
    }

    return result.Tools, nil
}

// CallTool calls a tool
func (c *Client) CallTool(ctx context.Context, name string, args interface{}) ([]mcp.Content, error) {
    params := map[string]interface{}{
        "name":      name,
        "arguments": args,
    }

    var result struct {
        Content []mcp.Content `json:"content"`
    }

    if err := c.call(ctx, "tools/call", params, &result); err != nil {
        return nil, err
    }

    return result.Content, nil
}

// ListResources lists available resources
func (c *Client) ListResources(ctx context.Context) ([]*mcp.Resource, error) {
    var result struct {
        Resources []*mcp.Resource `json:"resources"`
    }

    if err := c.call(ctx, "resources/list", nil, &result); err != nil {
        return nil, err
    }

    return result.Resources, nil
}

// ReadResource reads a resource
func (c *Client) ReadResource(ctx context.Context, uri string) ([]byte, error) {
    params := map[string]interface{}{
        "uri": uri,
    }

    var result struct {
        Contents []struct {
            URI      string `json:"uri"`
            MimeType string `json:"mimeType"`
            Text     string `json:"text,omitempty"`
            Blob     string `json:"blob,omitempty"`
        } `json:"contents"`
    }

    if err := c.call(ctx, "resources/read", params, &result); err != nil {
        return nil, err
    }

    if len(result.Contents) == 0 {
        return nil, &mcp.NotFoundError{Type: "resource", Name: uri}
    }

    // Handle text or blob
    if result.Contents[0].Text != "" {
        return []byte(result.Contents[0].Text), nil
    }

    // Decode base64 blob if present
    // ... implementation

    return nil, nil
}

// ListPrompts lists available prompts
func (c *Client) ListPrompts(ctx context.Context) ([]*mcp.Prompt, error) {
    var result struct {
        Prompts []*mcp.Prompt `json:"prompts"`
    }

    if err := c.call(ctx, "prompts/list", nil, &result); err != nil {
        return nil, err
    }

    return result.Prompts, nil
}

// GetPrompt gets a prompt
func (c *Client) GetPrompt(ctx context.Context, name string, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
    params := map[string]interface{}{
        "name":      name,
        "arguments": args,
    }

    var result struct {
        Messages []*mcp.PromptMessage `json:"messages"`
    }

    if err := c.call(ctx, "prompts/get", params, &result); err != nil {
        return nil, err
    }

    return result.Messages, nil
}

// call makes a JSON-RPC call
func (c *Client) call(ctx context.Context, method string, params, result interface{}) error {
    id := c.nextID.Add(1)

    msg := &Message{
        JSONRPC: "2.0",
        ID:      id,
        Method:  method,
    }

    if params != nil {
        paramsJSON, err := json.Marshal(params)
        if err != nil {
            return err
        }
        msg.Params = paramsJSON
    }

    respChan := make(chan *Response, 1)

    c.mu.Lock()
    c.pending[id] = respChan
    c.mu.Unlock()

    defer func() {
        c.mu.Lock()
        delete(c.pending, id)
        c.mu.Unlock()
    }()

    if err := c.session.Send(msg); err != nil {
        return err
    }

    select {
    case <-ctx.Done():
        return ctx.Err()
    case resp := <-respChan:
        if resp.Error != nil {
            return resp.Error
        }

        if result != nil && resp.Result != nil {
            return json.Unmarshal(resp.Result, result)
        }

        return nil
    }
}
```

### 7.3 Session Management

```go
package client

import (
    "context"
    "io"
)

// Session represents an active connection
type Session struct {
    conn   io.ReadWriteCloser
    reader *MessageReader
    writer *MessageWriter
}

// Send sends a message
func (s *Session) Send(msg *Message) error {
    return s.writer.Write(msg)
}

// Receive receives a message
func (s *Session) Receive() (*Message, error) {
    return s.reader.Read()
}

// Close closes the session
func (s *Session) Close() error {
    return s.conn.Close()
}

// MessageReader reads JSON-RPC messages
type MessageReader struct {
    decoder *json.Decoder
}

func (mr *MessageReader) Read() (*Message, error) {
    var msg Message
    if err := mr.decoder.Decode(&msg); err != nil {
        return nil, err
    }
    return &msg, nil
}

// MessageWriter writes JSON-RPC messages
type MessageWriter struct {
    encoder *json.Encoder
}

func (mw *MessageWriter) Write(msg *Message) error {
    return mw.encoder.Encode(msg)
}
```

---

## 8. Transport Layer

### 8.1 Transport Interface

```go
package transport

import (
    "context"
    "io"
)

// Transport handles connection establishment
type Transport interface {
    Connect(ctx context.Context) (io.ReadWriteCloser, error)
    Close() error
}
```

### 8.2 Stdio Transport

```go
package stdio

import (
    "context"
    "io"
    "os"
)

// Transport implements stdio transport
type Transport struct {
    stdin  io.Reader
    stdout io.Writer
}

// New creates a stdio transport
func New() *Transport {
    return &Transport{
        stdin:  os.Stdin,
        stdout: os.Stdout,
    }
}

// Connect returns a ReadWriteCloser for stdio
func (t *Transport) Connect(ctx context.Context) (io.ReadWriteCloser, error) {
    return &stdioConn{
        reader: t.stdin,
        writer: t.stdout,
    }, nil
}

type stdioConn struct {
    reader io.Reader
    writer io.Writer
}

func (c *stdioConn) Read(p []byte) (int, error) {
    return c.reader.Read(p)
}

func (c *stdioConn) Write(p []byte) (int, error) {
    return c.writer.Write(p)
}

func (c *stdioConn) Close() error {
    return nil
}

func (t *Transport) Close() error {
    return nil
}
```

### 8.3 HTTP Transport

```go
package http

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "sync"
)

// Transport implements HTTP transport
type Transport struct {
    url    string
    client *http.Client
}

// New creates an HTTP transport
func New(url string, opts ...Option) *Transport {
    t := &Transport{
        url:    url,
        client: &http.Client{},
    }

    for _, opt := range opts {
        opt(t)
    }

    return t
}

type Option func(*Transport)

func WithHTTPClient(client *http.Client) Option {
    return func(t *Transport) {
        t.client = client
    }
}

// Connect establishes HTTP connection
func (t *Transport) Connect(ctx context.Context) (io.ReadWriteCloser, error) {
    return &httpConn{
        url:    t.url,
        client: t.client,
    }, nil
}

func (t *Transport) Close() error {
    return nil
}

type httpConn struct {
    url    string
    client *http.Client
    buf    bytes.Buffer
    mu     sync.Mutex
}

func (c *httpConn) Read(p []byte) (int, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.buf.Read(p)
}

func (c *httpConn) Write(p []byte) (int, error) {
    // Send HTTP POST request
    resp, err := c.client.Post(c.url, "application/json", bytes.NewReader(p))
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return 0, fmt.Errorf("HTTP error: %d", resp.StatusCode)
    }

    // Read response into buffer
    c.mu.Lock()
    defer c.mu.Unlock()

    c.buf.Reset()
    _, err = io.Copy(&c.buf, resp.Body)
    return len(p), err
}

func (c *httpConn) Close() error {
    return nil
}
```

### 8.4 SSE Transport

```go
package sse

import (
    "bufio"
    "bytes"
    "context"
    "fmt"
    "io"
    "net/http"
    "strings"
)

// Transport implements SSE transport
type Transport struct {
    url       string
    client    *http.Client
    messageCh chan []byte
}

// New creates an SSE transport
func New(url string) *Transport {
    return &Transport{
        url:       url,
        client:    &http.Client{},
        messageCh: make(chan []byte, 10),
    }
}

// Connect establishes SSE connection
func (t *Transport) Connect(ctx context.Context) (io.ReadWriteCloser, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", t.url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Accept", "text/event-stream")

    resp, err := t.client.Do(req)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("SSE error: %d", resp.StatusCode)
    }

    conn := &sseConn{
        response:  resp,
        reader:    bufio.NewReader(resp.Body),
        writeCh:   make(chan []byte),
        readCh:    make(chan []byte, 10),
        closeOnce: &sync.Once{},
    }

    go conn.readLoop()

    return conn, nil
}

func (t *Transport) Close() error {
    return nil
}

type sseConn struct {
    response  *http.Response
    reader    *bufio.Reader
    writeCh   chan []byte
    readCh    chan []byte
    closeOnce *sync.Once
}

func (c *sseConn) readLoop() {
    for {
        line, err := c.reader.ReadString('\n')
        if err != nil {
            close(c.readCh)
            return
        }

        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "data: ") {
            data := strings.TrimPrefix(line, "data: ")
            c.readCh <- []byte(data)
        }
    }
}

func (c *sseConn) Read(p []byte) (int, error) {
    data, ok := <-c.readCh
    if !ok {
        return 0, io.EOF
    }

    n := copy(p, data)
    return n, nil
}

func (c *sseConn) Write(p []byte) (int, error) {
    // SSE is read-only; writing requires separate HTTP POST
    // This would need to be implemented based on the specific SSE endpoint
    return 0, fmt.Errorf("SSE transport does not support writing")
}

func (c *sseConn) Close() error {
    c.closeOnce.Do(func() {
        c.response.Body.Close()
    })
    return nil
}
```

---

## 9. Authentication

### 9.1 Auth Interface

```go
package auth

import (
    "context"
    "net/http"
)

// Provider handles authentication
type Provider interface {
    // Authenticate validates credentials and returns a token
    Authenticate(ctx context.Context, credentials interface{}) (string, error)

    // Middleware returns HTTP middleware for auth
    Middleware() func(http.Handler) http.Handler

    // ValidateToken validates a token
    ValidateToken(ctx context.Context, token string) (Claims, error)
}

// Claims represents authenticated user claims
type Claims struct {
    Subject string
    Email   string
    Scopes  []string
    Extra   map[string]interface{}
}
```

### 9.2 OAuth Provider

```go
package oauth

import (
    "context"
    "fmt"
    "net/http"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

// Provider implements OAuth authentication
type Provider struct {
    config *oauth2.Config
}

// NewGoogleProvider creates a Google OAuth provider
func NewGoogleProvider(clientID, clientSecret, redirectURL string, scopes []string) *Provider {
    return &Provider{
        config: &oauth2.Config{
            ClientID:     clientID,
            ClientSecret: clientSecret,
            RedirectURL:  redirectURL,
            Scopes:       scopes,
            Endpoint:     google.Endpoint,
        },
    }
}

// AuthURL returns the OAuth authorization URL
func (p *Provider) AuthURL(state string) string {
    return p.config.AuthCodeURL(state)
}

// Exchange exchanges an auth code for a token
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
    return p.config.Exchange(ctx, code)
}

// Authenticate implements auth.Provider
func (p *Provider) Authenticate(ctx context.Context, credentials interface{}) (string, error) {
    code, ok := credentials.(string)
    if !ok {
        return "", fmt.Errorf("invalid credentials type")
    }

    token, err := p.Exchange(ctx, code)
    if err != nil {
        return "", err
    }

    return token.AccessToken, nil
}

// ValidateToken validates an OAuth token
func (p *Provider) ValidateToken(ctx context.Context, token string) (auth.Claims, error) {
    // Implementation depends on provider
    // For Google, use TokenInfo endpoint or verify JWT
    return auth.Claims{}, nil
}

// Middleware returns OAuth middleware
func (p *Provider) Middleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            claims, err := p.ValidateToken(r.Context(), token)
            if err != nil {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            ctx := contextWithClaims(r.Context(), claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func extractToken(r *http.Request) string {
    auth := r.Header.Get("Authorization")
    if strings.HasPrefix(auth, "Bearer ") {
        return strings.TrimPrefix(auth, "Bearer ")
    }
    return ""
}
```

### 9.3 JWT Provider

```go
package jwt

import (
    "context"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// Provider implements JWT authentication
type Provider struct {
    secret []byte
    issuer string
}

// New creates a JWT provider
func New(secret []byte, issuer string) *Provider {
    return &Provider{
        secret: secret,
        issuer: issuer,
    }
}

// GenerateToken generates a JWT token
func (p *Provider) GenerateToken(claims auth.Claims, expiration time.Duration) (string, error) {
    now := time.Now()

    jwtClaims := jwt.MapClaims{
        "sub":    claims.Subject,
        "email":  claims.Email,
        "scopes": claims.Scopes,
        "iss":    p.issuer,
        "iat":    now.Unix(),
        "exp":    now.Add(expiration).Unix(),
    }

    for k, v := range claims.Extra {
        jwtClaims[k] = v
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
    return token.SignedString(p.secret)
}

// ValidateToken validates a JWT token
func (p *Provider) ValidateToken(ctx context.Context, tokenString string) (auth.Claims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return p.secret, nil
    })

    if err != nil {
        return auth.Claims{}, err
    }

    if !token.Valid {
        return auth.Claims{}, fmt.Errorf("invalid token")
    }

    claims := token.Claims.(jwt.MapClaims)

    return auth.Claims{
        Subject: claims["sub"].(string),
        Email:   claims["email"].(string),
        Scopes:  convertScopes(claims["scopes"]),
    }, nil
}

func convertScopes(v interface{}) []string {
    if v == nil {
        return nil
    }

    switch s := v.(type) {
    case []interface{}:
        scopes := make([]string, len(s))
        for i, scope := range s {
            scopes[i] = scope.(string)
        }
        return scopes
    default:
        return nil
    }
}
```

---

## 10. Advanced Features

### 10.1 Server Composition (Mounting)

```go
package server

// Mount mounts another server with an optional prefix
func (s *Server) Mount(prefix string, other *Server) error {
    // Mount tools
    for name, tool := range other.tools.tools {
        mountedName := name
        if prefix != "" {
            mountedName = prefix + "_" + name
        }

        s.tools.tools[mountedName] = tool
    }

    // Mount resources with URI prefix
    for uri, resource := range other.resources.resources {
        mountedURI := uri
        if prefix != "" {
            mountedURI = addURIPrefix(uri, prefix)
        }

        s.resources.resources[mountedURI] = resource
    }

    // Mount prompts
    for name, prompt := range other.prompts.prompts {
        mountedName := name
        if prefix != "" {
            mountedName = prefix + "_" + name
        }

        s.prompts.prompts[mountedName] = prompt
    }

    return nil
}

func addURIPrefix(uri, prefix string) string {
    // Parse URI and add prefix to path component
    // Example: "file:///path" -> "file:///prefix/path"
    return uri // Simplified
}
```

### 10.2 Proxy Server

```go
package proxy

import (
    "context"

    "github.com/yourusername/fullmcp/client"
    "github.com/yourusername/fullmcp/server"
)

// Server is a proxy that forwards requests to a backend
type Server struct {
    *server.Server
    client *client.Client
}

// New creates a proxy server
func New(backend *client.Client) *Server {
    srv := server.New("proxy")

    ps := &Server{
        Server: srv,
        client: backend,
    }

    // Forward all tool calls
    ps.tools = &proxyToolManager{
        client: backend,
    }

    return ps
}

type proxyToolManager struct {
    client *client.Client
}

func (ptm *proxyToolManager) Call(ctx context.Context, name string, args json.RawMessage) (interface{}, error) {
    return ptm.client.CallTool(ctx, name, args)
}

func (ptm *proxyToolManager) List() []*mcp.Tool {
    tools, _ := ptm.client.ListTools(context.Background())
    return tools
}
```

### 10.3 Context and Dependencies

```go
package server

import (
    "context"
)

type contextKey string

const (
    serverContextKey contextKey = "mcp.server"
    sessionContextKey contextKey = "mcp.session"
)

// ServerContext provides access to server capabilities
type ServerContext struct {
    server  *Server
    session *Session
}

// WithContext adds server context
func (s *Server) WithContext(ctx context.Context, session *Session) context.Context {
    sc := &ServerContext{
        server:  s,
        session: session,
    }

    return context.WithValue(ctx, serverContextKey, sc)
}

// FromContext retrieves server context
func FromContext(ctx context.Context) *ServerContext {
    sc, _ := ctx.Value(serverContextKey).(*ServerContext)
    return sc
}

// Log sends a log message
func (sc *ServerContext) Log(level string, message string) error {
    // Send logging notification
    return nil
}

// Progress sends a progress notification
func (sc *ServerContext) Progress(token string, progress, total int) error {
    // Send progress notification
    return nil
}

// ReadResource reads a resource from the server
func (sc *ServerContext) ReadResource(ctx context.Context, uri string) ([]byte, error) {
    return sc.server.resources.Read(ctx, uri)
}
```

### 10.4 Lifecycle Hooks

```go
package server

import "context"

// LifespanFunc is called during server lifecycle
type LifespanFunc func(context.Context, *Server) (context.Context, func(), error)

// WithLifespan sets a lifespan function
func WithLifespan(fn LifespanFunc) ServerOption {
    return func(s *Server) {
        s.lifespan = fn
    }
}

// Example usage:
// server.New("my-server", server.WithLifespan(func(ctx context.Context, s *Server) (context.Context, func(), error) {
//     // Setup
//     db, err := sql.Open("postgres", connStr)
//     if err != nil {
//         return nil, nil, err
//     }
//
//     ctx = context.WithValue(ctx, "db", db)
//
//     cleanup := func() {
//         db.Close()
//     }
//
//     return ctx, cleanup, nil
// }))
```

### 10.5 Resource Templates

```go
package server

import (
    "context"
    "regexp"
    "strings"
)

// ResourceTemplateHandler handles parameterized resources
type ResourceTemplateHandler struct {
    URITemplate string
    Name        string
    Description string
    MimeType    string
    Reader      func(context.Context, map[string]string) ([]byte, error)
    pattern     *regexp.Regexp
}

// NewResourceTemplate creates a resource template
func NewResourceTemplate(uriTemplate string, reader func(context.Context, map[string]string) ([]byte, error)) (*ResourceTemplateHandler, error) {
    // Convert URI template to regex
    // Example: "file:///{path}" -> "file:///(?P<path>.*)"
    pattern := templateToRegex(uriTemplate)

    return &ResourceTemplateHandler{
        URITemplate: uriTemplate,
        Reader:      reader,
        pattern:     regexp.MustCompile(pattern),
    }, nil
}

// Match checks if a URI matches the template
func (rth *ResourceTemplateHandler) Match(uri string) (map[string]string, bool) {
    matches := rth.pattern.FindStringSubmatch(uri)
    if matches == nil {
        return nil, false
    }

    params := make(map[string]string)
    for i, name := range rth.pattern.SubexpNames() {
        if i > 0 && name != "" {
            params[name] = matches[i]
        }
    }

    return params, true
}

func templateToRegex(template string) string {
    // Convert {param} to (?P<param>...)
    re := regexp.MustCompile(`\{(\w+)\}`)
    return re.ReplaceAllString(regexp.QuoteMeta(template), `(?P<$1>[^/]+)`)
}

// ResourceManager with template support
func (rm *ResourceManager) RegisterTemplate(handler *ResourceTemplateHandler) error {
    rm.mu.Lock()
    defer rm.mu.Unlock()

    rm.templates[handler.URITemplate] = handler
    return nil
}

func (rm *ResourceManager) Read(ctx context.Context, uri string) ([]byte, error) {
    rm.mu.RLock()
    defer rm.mu.RUnlock()

    // Try exact match first
    if handler, exists := rm.resources[uri]; exists {
        return handler.Reader(ctx)
    }

    // Try templates
    for _, template := range rm.templates {
        if params, ok := template.Match(uri); ok {
            return template.Reader(ctx, params)
        }
    }

    return nil, &mcp.NotFoundError{Type: "resource", Name: uri}
}
```

---

## 11. Testing Strategy

### 11.1 Unit Testing

```go
package server_test

import (
    "context"
    "testing"

    "github.com/yourusername/fullmcp/server"
)

func TestToolRegistration(t *testing.T) {
    s := server.New("test")

    err := s.AddTool("add", func(ctx context.Context, input struct {
        A int `json:"a"`
        B int `json:"b"`
    }) (int, error) {
        return input.A + input.B, nil
    })

    if err != nil {
        t.Fatalf("failed to register tool: %v", err)
    }

    tools := s.tools.List()
    if len(tools) != 1 {
        t.Fatalf("expected 1 tool, got %d", len(tools))
    }

    if tools[0].Name != "add" {
        t.Errorf("expected tool name 'add', got %q", tools[0].Name)
    }
}

func TestToolExecution(t *testing.T) {
    s := server.New("test")

    s.AddTool("multiply", func(ctx context.Context, input struct {
        A float64 `json:"a"`
        B float64 `json:"b"`
    }) (float64, error) {
        return input.A * input.B, nil
    })

    ctx := context.Background()

    args := json.RawMessage(`{"a": 3.5, "b": 2.0}`)
    result, err := s.tools.Call(ctx, "multiply", args)

    if err != nil {
        t.Fatalf("tool call failed: %v", err)
    }

    if result != 7.0 {
        t.Errorf("expected 7.0, got %v", result)
    }
}
```

### 11.2 Integration Testing

```go
package integration_test

import (
    "context"
    "testing"

    "github.com/yourusername/fullmcp/client"
    "github.com/yourusername/fullmcp/server"
    "github.com/yourusername/fullmcp/transport/inmemory"
)

func TestClientServer(t *testing.T) {
    // Create server
    srv := server.New("test")

    srv.AddTool("echo", func(ctx context.Context, input struct {
        Message string `json:"message"`
    }) (string, error) {
        return input.Message, nil
    })

    // Create in-memory transport
    transport := inmemory.New(srv)

    // Create client
    c := client.New(transport)

    ctx := context.Background()

    if err := c.Connect(ctx); err != nil {
        t.Fatalf("failed to connect: %v", err)
    }
    defer c.Close()

    // List tools
    tools, err := c.ListTools(ctx)
    if err != nil {
        t.Fatalf("failed to list tools: %v", err)
    }

    if len(tools) != 1 {
        t.Fatalf("expected 1 tool, got %d", len(tools))
    }

    // Call tool
    result, err := c.CallTool(ctx, "echo", map[string]interface{}{
        "message": "hello",
    })

    if err != nil {
        t.Fatalf("tool call failed: %v", err)
    }

    // Verify result
    if len(result) == 0 {
        t.Fatal("expected non-empty result")
    }
}
```

### 11.3 Mock Transport

```go
package inmemory

import (
    "context"
    "io"

    "github.com/yourusername/fullmcp/server"
)

// Transport provides in-memory transport for testing
type Transport struct {
    server *server.Server
}

// New creates an in-memory transport
func New(s *server.Server) *Transport {
    return &Transport{server: s}
}

// Connect creates an in-memory connection
func (t *Transport) Connect(ctx context.Context) (io.ReadWriteCloser, error) {
    r, w := io.Pipe()

    conn := &inMemoryConn{
        reader: r,
        writer: w,
    }

    // Start server handler in background
    go t.server.Handle(ctx, conn)

    return conn, nil
}

type inMemoryConn struct {
    reader *io.PipeReader
    writer *io.PipeWriter
}

func (c *inMemoryConn) Read(p []byte) (int, error) {
    return c.reader.Read(p)
}

func (c *inMemoryConn) Write(p []byte) (int, error) {
    return c.writer.Write(p)
}

func (c *inMemoryConn) Close() error {
    c.reader.Close()
    c.writer.Close()
    return nil
}
```

---

## 12. Migration Path

### 12.1 Example: Basic Server

**Python (fastmcp)**:
```python
from fastmcp import FastMCP

mcp = FastMCP("math-server")

@mcp.tool
def add(a: float, b: float) -> float:
    """Add two numbers."""
    return a + b

@mcp.tool
def multiply(a: float, b: float) -> float:
    """Multiply two numbers."""
    return a * b

mcp.run()
```

**Go (fullmcp)**:
```go
package main

import (
    "context"

    "github.com/yourusername/fullmcp/server"
)

type MathInput struct {
    A float64 `json:"a" description:"First number"`
    B float64 `json:"b" description:"Second number"`
}

func main() {
    srv := server.New("math-server")

    srv.AddTool("add", func(ctx context.Context, input MathInput) (float64, error) {
        return input.A + input.B, nil
    }, server.WithDescription("Add two numbers"))

    srv.AddTool("multiply", func(ctx context.Context, input MathInput) (float64, error) {
        return input.A * input.B, nil
    }, server.WithDescription("Multiply two numbers"))

    srv.Run(context.Background())
}
```

### 12.2 Example: Resource Server

**Python (fastmcp)**:
```python
@mcp.resource("config://app")
def get_config() -> str:
    return json.dumps({"debug": True})

@mcp.resource("file:///{path}")
def read_file(path: str) -> str:
    with open(path) as f:
        return f.read()
```

**Go (fullmcp)**:
```go
srv.AddResource("config://app", func(ctx context.Context) ([]byte, error) {
    config := map[string]interface{}{"debug": true}
    return json.Marshal(config)
})

srv.AddResourceTemplate("file:///{path}", func(ctx context.Context, params map[string]string) ([]byte, error) {
    return os.ReadFile(params["path"])
})
```

### 12.3 Example: Client

**Python (fastmcp)**:
```python
from fastmcp.client import Client

async with Client("http://localhost:8080") as client:
    tools = await client.list_tools()
    result = await client.call_tool("add", {"a": 5, "b": 3})
```

**Go (fullmcp)**:
```go
import (
    "github.com/yourusername/fullmcp/client"
    "github.com/yourusername/fullmcp/transport/http"
)

func main() {
    transport := http.New("http://localhost:8080")
    c := client.New(transport)

    ctx := context.Background()

    if err := c.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer c.Close()

    tools, err := c.ListTools(ctx)
    if err != nil {
        log.Fatal(err)
    }

    result, err := c.CallTool(ctx, "add", map[string]interface{}{
        "a": 5,
        "b": 3,
    })
}
```

---

## 13. Implementation Roadmap

### Phase 1: Core Protocol (4-6 weeks)
- [ ] Define core types and interfaces
- [ ] Implement JSON-RPC layer
- [ ] Build stdio transport
- [ ] Create basic server and client
- [ ] Tool registration and execution
- [ ] Resource management
- [ ] Prompt management

### Phase 2: Transports (2-3 weeks)
- [ ] HTTP transport
- [ ] SSE transport
- [ ] WebSocket transport (optional)
- [ ] Transport testing framework

### Phase 3: Advanced Features (3-4 weeks)
- [ ] Resource templates
- [ ] Server composition/mounting
- [ ] Proxy server
- [ ] Middleware system
- [ ] Context and dependencies
- [ ] Lifecycle hooks

### Phase 4: Authentication (2-3 weeks)
- [ ] Auth interface
- [ ] OAuth providers (Google, GitHub, Azure)
- [ ] JWT authentication
- [ ] API key authentication
- [ ] Auth middleware

### Phase 5: Developer Experience (2-3 weeks)
- [ ] Builder APIs
- [ ] Reflection-based schema generation
- [ ] Comprehensive examples
- [ ] Documentation
- [ ] CLI tools

### Phase 6: Testing & Polish (2-3 weeks)
- [ ] Unit tests (>80% coverage)
- [ ] Integration tests
- [ ] Performance benchmarks
- [ ] Documentation review
- [ ] API stability review

**Total Estimated Time**: 15-22 weeks (4-5.5 months)

---

## 14. Dependencies

### Required Go Packages

```go
module github.com/yourusername/fullmcp

go 1.21

require (
    github.com/invopop/jsonschema v0.12.0        // JSON Schema generation
    github.com/golang-jwt/jwt/v5 v5.2.0          // JWT authentication
    golang.org/x/oauth2 v0.15.0                  // OAuth2 support
    github.com/gorilla/websocket v1.5.1          // WebSocket transport
    google.golang.org/api v0.154.0               // Google APIs
    github.com/Azure/azure-sdk-for-go v68.0.0    // Azure integration
)
```

---

## 15. Success Criteria

1. **Feature Parity**: Implements all core fastmcp features
2. **Idiomatic Go**: Follows Go conventions and best practices
3. **Performance**: Handles 1000+ concurrent connections
4. **Type Safety**: Compile-time type checking for tools/resources
5. **Documentation**: Comprehensive docs and examples
6. **Testing**: >80% test coverage
7. **Stability**: Stable API with semantic versioning
8. **Community**: Active examples and integration guides

---

## 16. Conclusion

This specification provides a complete blueprint for implementing a production-ready Golang MCP library. The design prioritizes:

- **Simplicity**: Easy to use for common cases
- **Power**: Supports advanced patterns when needed
- **Safety**: Type-safe, concurrent-safe implementations
- **Extensibility**: Plugin architecture for customization

The resulting library will enable Go developers to build robust MCP servers and clients with minimal boilerplate while maintaining full control when needed.
