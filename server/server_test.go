package server

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

type mockTransport struct {
	reader *bytes.Buffer
	writer *bytes.Buffer
	mu     sync.Mutex
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		reader: &bytes.Buffer{},
		writer: &bytes.Buffer{},
	}
}

func (m *mockTransport) Read(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.reader.Read(p)
}

func (m *mockTransport) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.writer.Write(p)
}

func (m *mockTransport) Close() error {
	return nil
}

func (m *mockTransport) writeMessage(msg *mcp.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return json.NewEncoder(m.reader).Encode(msg)
}

func (m *mockTransport) readResponse() (*mcp.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var msg mcp.Message
	if err := json.NewDecoder(m.writer).Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func TestServer_New(t *testing.T) {
	srv := New("test-server",
		WithVersion("1.0.0"),
		WithInstructions("Test instructions"),
	)

	if srv.name != "test-server" {
		t.Errorf("expected name 'test-server', got '%s'", srv.name)
	}

	if srv.version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", srv.version)
	}

	if srv.instructions != "Test instructions" {
		t.Errorf("expected instructions 'Test instructions', got '%s'", srv.instructions)
	}
}

func TestServer_Initialize(t *testing.T) {
	srv := New("test-server", WithVersion("1.0.0"))
	transport := newMockTransport()

	// Send initialize request
	initMsg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2025-06-18"}`),
	}
	transport.writeMessage(initMsg)

	// Send EOF to stop server
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		transport.mu.Lock()
		transport.reader.Write([]byte{}) // Trigger EOF
		transport.mu.Unlock()
		cancel()
	}()

	srv.Serve(ctx, transport)

	// Read response
	response, err := transport.readResponse()
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if result["protocolVersion"] != "2025-06-18" {
		t.Errorf("unexpected protocol version: %v", result["protocolVersion"])
	}
}

func TestServer_ToolsList(t *testing.T) {
	srv := New("test-server")
	srv.AddTool(&ToolHandler{
		Name:        "test-tool",
		Description: "A test tool",
		Schema:      map[string]interface{}{"type": "object"},
		Handler: func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			return "result", nil
		},
	})

	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	response := srv.HandleMessage(context.Background(), msg)
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	var result struct {
		Tools []*mcp.Tool `json:"tools"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(result.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(result.Tools))
	}

	if result.Tools[0].Name != "test-tool" {
		t.Errorf("expected tool name 'test-tool', got '%s'", result.Tools[0].Name)
	}
}

func TestServer_ToolsCall(t *testing.T) {
	srv := New("test-server")
	srv.AddTool(&ToolHandler{
		Name: "add",
		Handler: func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			var input struct {
				A int `json:"a"`
				B int `json:"b"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, err
			}
			return input.A + input.B, nil
		},
	})

	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"add","arguments":{"a":5,"b":3}}`),
	}

	response := srv.HandleMessage(context.Background(), msg)
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		t.Fatal("expected content in result")
	}
}

func TestServer_ResourcesList(t *testing.T) {
	srv := New("test-server")
	srv.AddResource(&ResourceHandler{
		URI:  "test://resource",
		Name: "Test Resource",
		Reader: func(ctx context.Context) ([]byte, error) {
			return []byte("content"), nil
		},
	})

	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "resources/list",
	}

	response := srv.HandleMessage(context.Background(), msg)
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	var result struct {
		Resources []*mcp.Resource `json:"resources"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(result.Resources))
	}
}

func TestServer_ResourcesRead(t *testing.T) {
	srv := New("test-server")
	srv.AddResource(&ResourceHandler{
		URI: "test://file",
		Reader: func(ctx context.Context) ([]byte, error) {
			return []byte("file content"), nil
		},
	})

	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "resources/read",
		Params:  json.RawMessage(`{"uri":"test://file"}`),
	}

	response := srv.HandleMessage(context.Background(), msg)
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	var result struct {
		Contents []struct {
			URI  string `json:"uri"`
			Text string `json:"text"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(result.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(result.Contents))
	}

	if result.Contents[0].Text != "file content" {
		t.Errorf("expected 'file content', got '%s'", result.Contents[0].Text)
	}
}

func TestServer_PromptsList(t *testing.T) {
	srv := New("test-server")
	srv.AddPrompt(&PromptHandler{
		Name: "test-prompt",
		Renderer: func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{}, nil
		},
	})

	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "prompts/list",
	}

	response := srv.HandleMessage(context.Background(), msg)
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	var result struct {
		Prompts []*mcp.Prompt `json:"prompts"`
	}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(result.Prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(result.Prompts))
	}
}

func TestServer_PromptsGet(t *testing.T) {
	srv := New("test-server")
	srv.AddPrompt(&PromptHandler{
		Name: "greeting",
		Renderer: func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{
				{
					Role: "user",
					Content: []mcp.Content{
						mcp.TextContent{Type: "text", Text: "Hello"},
					},
				},
			}, nil
		},
	})

	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "prompts/get",
		Params:  json.RawMessage(`{"name":"greeting","arguments":{}}`),
	}

	response := srv.HandleMessage(context.Background(), msg)
	if response.Error != nil {
		t.Fatalf("unexpected error: %v", response.Error)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	messages, ok := result["messages"].([]interface{})
	if !ok || len(messages) != 1 {
		t.Fatalf("expected 1 message, got %v", result["messages"])
	}
}

func TestServer_MethodNotFound(t *testing.T) {
	srv := New("test-server")

	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown/method",
	}

	response := srv.HandleMessage(context.Background(), msg)
	if response.Error == nil {
		t.Fatal("expected error for unknown method")
	}

	if response.Error.Code != int(mcp.MethodNotFound) {
		t.Errorf("expected error code %d, got %d", mcp.MethodNotFound, response.Error.Code)
	}
}

func TestServer_InvalidParams(t *testing.T) {
	srv := New("test-server")

	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  json.RawMessage(`{invalid json}`),
	}

	response := srv.HandleMessage(context.Background(), msg)
	if response.Error == nil {
		t.Fatal("expected error for invalid params")
	}

	if response.Error.Code != int(mcp.InvalidParams) {
		t.Errorf("expected error code %d, got %d", mcp.InvalidParams, response.Error.Code)
	}
}
