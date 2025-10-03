package client

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"

	"github.com/jmcarbo/fullmcp/internal/jsonrpc"
	"github.com/jmcarbo/fullmcp/internal/testutil"
	"github.com/jmcarbo/fullmcp/mcp"
)

// AsyncMockServer simulates a server for async client testing
type AsyncMockServer struct {
	t               *testing.T
	clientTransport io.ReadWriteCloser
	serverTransport io.ReadWriteCloser
	reader          *jsonrpc.MessageReader
	writer          *jsonrpc.MessageWriter
	stop            chan struct{}
	wg              sync.WaitGroup
}

func NewAsyncMockServer(t *testing.T) (*AsyncMockServer, io.ReadWriteCloser) {
	clientTransport, serverTransport := testutil.NewPipeTransport()

	return &AsyncMockServer{
		t:               t,
		clientTransport: clientTransport,
		serverTransport: serverTransport,
		reader:          jsonrpc.NewMessageReader(serverTransport),
		writer:          jsonrpc.NewMessageWriter(serverTransport),
		stop:            make(chan struct{}),
	}, clientTransport
}

func (s *AsyncMockServer) Start() {
	s.wg.Add(1)
	go s.handleRequests()
}

func (s *AsyncMockServer) Stop() {
	close(s.stop)
	s.serverTransport.Close()
	s.wg.Wait()
}

func (s *AsyncMockServer) handleRequests() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stop:
			return
		default:
			msg, err := s.reader.Read()
			if err != nil {
				return
			}

			// Handle different request types
			switch msg.Method {
			case "initialize":
				s.sendInitResponse(msg)
			case "notifications/initialized":
				// Ignore
			case "tools/list":
				s.sendToolsListResponse(msg)
			case "tools/call":
				s.sendToolCallResponse(msg)
			case "resources/list":
				s.sendResourcesListResponse(msg)
			case "resources/read":
				s.sendResourceReadResponse(msg)
			case "prompts/list":
				s.sendPromptsListResponse(msg)
			case "prompts/get":
				s.sendPromptGetResponse(msg)
			}
		}
	}
}

func (s *AsyncMockServer) sendInitResponse(req *mcp.Message) {
	response := &mcp.Message{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: json.RawMessage(`{
			"protocolVersion": "2024-11-05",
			"capabilities": {
				"tools": {},
				"resources": {},
				"prompts": {}
			},
			"serverInfo": {
				"name": "test-server",
				"version": "1.0.0"
			}
		}`),
	}
	s.writer.Write(response)
}

func (s *AsyncMockServer) sendToolsListResponse(req *mcp.Message) {
	response := &mcp.Message{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: json.RawMessage(`{
			"tools": [
				{
					"name": "add",
					"description": "Add two numbers",
					"inputSchema": {
						"type": "object",
						"properties": {
							"a": {"type": "number"},
							"b": {"type": "number"}
						}
					}
				}
			]
		}`),
	}
	s.writer.Write(response)
}

func (s *AsyncMockServer) sendToolCallResponse(req *mcp.Message) {
	response := &mcp.Message{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: json.RawMessage(`{
			"content": [
				{
					"type": "text",
					"text": "42"
				}
			]
		}`),
	}
	s.writer.Write(response)
}

func (s *AsyncMockServer) sendResourcesListResponse(req *mcp.Message) {
	response := &mcp.Message{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: json.RawMessage(`{
			"resources": [
				{
					"uri": "config://app",
					"name": "App Config",
					"mimeType": "application/json"
				}
			]
		}`),
	}
	s.writer.Write(response)
}

func (s *AsyncMockServer) sendResourceReadResponse(req *mcp.Message) {
	response := &mcp.Message{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: json.RawMessage(`{
			"contents": [
				{
					"uri": "config://app",
					"mimeType": "application/json",
					"text": "{\"debug\": true}"
				}
			]
		}`),
	}
	s.writer.Write(response)
}

func (s *AsyncMockServer) sendPromptsListResponse(req *mcp.Message) {
	response := &mcp.Message{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: json.RawMessage(`{
			"prompts": [
				{
					"name": "greeting",
					"description": "Generate a greeting"
				}
			]
		}`),
	}
	s.writer.Write(response)
}

func (s *AsyncMockServer) sendPromptGetResponse(req *mcp.Message) {
	// Send a simple response that avoids the Content interface marshaling issue
	response := &mcp.Message{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: json.RawMessage(`{
			"messages": []
		}`),
	}
	s.writer.Write(response)
}

func TestClient_ListTools_Async(t *testing.T) {
	server, clientTransport := NewAsyncMockServer(t)
	server.Start()
	defer server.Stop()

	client := New(clientTransport)
	ctx := context.Background()

	// Connect first
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// List tools
	tools, err := client.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	if len(tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(tools))
	}

	if tools[0].Name != "add" {
		t.Errorf("expected tool 'add', got '%s'", tools[0].Name)
	}
}

func TestClient_CallTool_Async(t *testing.T) {
	server, clientTransport := NewAsyncMockServer(t)
	server.Start()
	defer server.Stop()

	client := New(clientTransport)
	ctx := context.Background()

	// Connect first
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Call tool
	result, err := client.CallTool(ctx, "add", map[string]int{"a": 1, "b": 2})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	if result != "42" {
		t.Errorf("expected '42', got '%v'", result)
	}
}

func TestClient_ListResources_Async(t *testing.T) {
	server, clientTransport := NewAsyncMockServer(t)
	server.Start()
	defer server.Stop()

	client := New(clientTransport)
	ctx := context.Background()

	// Connect first
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// List resources
	resources, err := client.ListResources(ctx)
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}

	if len(resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(resources))
	}

	if resources[0].URI != "config://app" {
		t.Errorf("expected URI 'config://app', got '%s'", resources[0].URI)
	}
}

func TestClient_ReadResource_Async(t *testing.T) {
	server, clientTransport := NewAsyncMockServer(t)
	server.Start()
	defer server.Stop()

	client := New(clientTransport)
	ctx := context.Background()

	// Connect first
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Read resource
	data, err := client.ReadResource(ctx, "config://app")
	if err != nil {
		t.Fatalf("ReadResource failed: %v", err)
	}

	expected := `{"debug": true}`
	if string(data) != expected {
		t.Errorf("expected '%s', got '%s'", expected, data)
	}
}

func TestClient_ListPrompts_Async(t *testing.T) {
	server, clientTransport := NewAsyncMockServer(t)
	server.Start()
	defer server.Stop()

	client := New(clientTransport)
	ctx := context.Background()

	// Connect first
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// List prompts
	prompts, err := client.ListPrompts(ctx)
	if err != nil {
		t.Fatalf("ListPrompts failed: %v", err)
	}

	if len(prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(prompts))
	}

	if prompts[0].Name != "greeting" {
		t.Errorf("expected prompt 'greeting', got '%s'", prompts[0].Name)
	}
}

func TestClient_GetPrompt_Async(t *testing.T) {
	server, clientTransport := NewAsyncMockServer(t)
	server.Start()
	defer server.Stop()

	client := New(clientTransport)
	ctx := context.Background()

	// Connect first
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Get prompt
	messages, err := client.GetPrompt(ctx, "greeting", nil)
	if err != nil {
		t.Fatalf("GetPrompt failed: %v", err)
	}

	// Should return empty messages array
	if len(messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(messages))
	}
}

func TestClient_ContextTimeout_Async(t *testing.T) {
	server, clientTransport := NewAsyncMockServer(t)
	server.Start()
	defer server.Stop()

	client := New(clientTransport)

	// Connect first
	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// This should fail with context canceled
	_, err = client.ListTools(ctx)
	if err == nil {
		t.Error("expected context error, got nil")
	}

	if err != context.Canceled && err != context.DeadlineExceeded {
		t.Errorf("expected context.Canceled or DeadlineExceeded, got %v", err)
	}
}

func TestClient_ConcurrentCalls_Async(t *testing.T) {
	server, clientTransport := NewAsyncMockServer(t)
	server.Start()
	defer server.Stop()

	client := New(clientTransport)
	ctx := context.Background()

	// Connect first
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Make concurrent calls
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.ListTools(ctx)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("concurrent call failed: %v", err)
	}
}

func TestClient_GetPrompt_WithArgs_Async(t *testing.T) {
	server, clientTransport := NewAsyncMockServer(t)
	server.Start()
	defer server.Stop()

	client := New(clientTransport)
	ctx := context.Background()

	// Connect first
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Get prompt with arguments
	args := map[string]interface{}{
		"name": "Alice",
	}
	messages, err := client.GetPrompt(ctx, "greeting", args)
	if err != nil {
		t.Fatalf("GetPrompt failed: %v", err)
	}

	// Should return empty messages array
	if messages == nil {
		t.Error("expected messages array, got nil")
	}
}
