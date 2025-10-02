package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

// BenchmarkToolRegistration measures the performance of registering tools
func BenchmarkToolRegistration(b *testing.B) {
	srv := New("benchmark-server")

	toolHandler := &ToolHandler{
		Name:        "test-tool",
		Description: "Test tool for benchmarking",
		Schema:      map[string]interface{}{"type": "object"},
		Handler: func(_ context.Context, _ json.RawMessage) (interface{}, error) {
			return "result", nil
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = srv.AddTool(toolHandler)
	}
}

// BenchmarkToolCall measures the performance of calling a tool
func BenchmarkToolCall(b *testing.B) {
	srv := New("benchmark-server")

	toolHandler := &ToolHandler{
		Name:        "echo",
		Description: "Echo tool",
		Schema:      map[string]interface{}{"type": "object"},
		Handler: func(_ context.Context, args json.RawMessage) (interface{}, error) {
			return string(args), nil
		},
	}
	_ = srv.AddTool(toolHandler)

	ctx := context.Background()
	args := json.RawMessage(`{"message":"hello"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = srv.tools.Call(ctx, "echo", args)
	}
}

// BenchmarkResourceRead measures the performance of reading a resource
func BenchmarkResourceRead(b *testing.B) {
	srv := New("benchmark-server")

	resourceHandler := &ResourceHandler{
		URI:         "config://test",
		Name:        "Test Config",
		Description: "Test configuration",
		MimeType:    "application/json",
		Reader: func(_ context.Context) ([]byte, error) {
			return []byte(`{"key":"value"}`), nil
		},
	}
	_ = srv.AddResource(resourceHandler)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = srv.resources.Read(ctx, "config://test")
	}
}

// BenchmarkMessageHandling measures the performance of handling messages
func BenchmarkMessageHandling(b *testing.B) {
	srv := New("benchmark-server")

	toolHandler := &ToolHandler{
		Name:        "add",
		Description: "Add two numbers",
		Schema:      map[string]interface{}{"type": "object"},
		Handler: func(_ context.Context, args json.RawMessage) (interface{}, error) {
			var data struct {
				A int `json:"a"`
				B int `json:"b"`
			}
			_ = json.Unmarshal(args, &data)
			return data.A + data.B, nil
		},
	}
	_ = srv.AddTool(toolHandler)

	ctx := context.Background()
	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"add","arguments":{"a":5,"b":3}}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = srv.HandleMessage(ctx, msg)
	}
}

// BenchmarkPromptGet measures the performance of getting a prompt
func BenchmarkPromptGet(b *testing.B) {
	srv := New("benchmark-server")

	promptHandler := &PromptHandler{
		Name:        "greeting",
		Description: "Greeting prompt",
		Arguments:   []mcp.PromptArgument{{Name: "name", Description: "User name", Required: true}},
		Renderer: func(_ context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			name := args["name"].(string)
			return []*mcp.PromptMessage{{
				Role: "user",
				Content: []mcp.Content{
					&mcp.TextContent{
						Type: "text",
						Text: "Hello, " + name,
					},
				},
			}}, nil
		},
	}
	_ = srv.AddPrompt(promptHandler)

	ctx := context.Background()
	args := map[string]interface{}{"name": "Alice"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = srv.prompts.Get(ctx, "greeting", args)
	}
}

// BenchmarkListTools measures the performance of listing tools
func BenchmarkListTools(b *testing.B) {
	srv := New("benchmark-server")

	// Add 100 tools
	for i := 0; i < 100; i++ {
		toolHandler := &ToolHandler{
			Name:        "tool-" + string(rune(i)),
			Description: "Test tool",
			Schema:      map[string]interface{}{"type": "object"},
			Handler: func(_ context.Context, _ json.RawMessage) (interface{}, error) {
				return "result", nil
			},
		}
		_ = srv.AddTool(toolHandler)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = srv.tools.List(ctx)
	}
}

// BenchmarkListResources measures the performance of listing resources
func BenchmarkListResources(b *testing.B) {
	srv := New("benchmark-server")

	// Add 100 resources
	for i := 0; i < 100; i++ {
		resourceHandler := &ResourceHandler{
			URI:         "config://test-" + string(rune(i)),
			Name:        "Test Config",
			Description: "Test configuration",
			MimeType:    "application/json",
			Reader: func(_ context.Context) ([]byte, error) {
				return []byte(`{"key":"value"}`), nil
			},
		}
		_ = srv.AddResource(resourceHandler)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = srv.resources.List()
	}
}

// BenchmarkMiddlewareChain measures middleware overhead
func BenchmarkMiddlewareChain(b *testing.B) {
	// Create middleware chain
	mw1 := func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (*Response, error) {
			return next(ctx, req)
		}
	}

	mw2 := func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (*Response, error) {
			return next(ctx, req)
		}
	}

	srv := New("benchmark-server", WithMiddleware(mw1, mw2))

	toolHandler := &ToolHandler{
		Name:        "noop",
		Description: "No-op tool",
		Schema:      map[string]interface{}{"type": "object"},
		Handler: func(_ context.Context, _ json.RawMessage) (interface{}, error) {
			return nil, nil
		},
	}
	_ = srv.AddTool(toolHandler)

	ctx := context.Background()
	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"noop","arguments":{}}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = srv.HandleMessage(ctx, msg)
	}
}
