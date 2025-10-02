package client

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jmcarbo/fullmcp/internal/testutil"
	"github.com/jmcarbo/fullmcp/mcp"
)

func TestClient_Connect(t *testing.T) {
	transport := testutil.NewMockTransport()

	// Prepare initialize response
	initResponse := &mcp.Message{
		JSONRPC: "2.0",
		ID:      float64(1),
		Result: json.RawMessage(`{
			"protocolVersion": "2024-11-05",
			"capabilities": {},
			"serverInfo": {"name": "test-server", "version": "1.0.0"}
		}`),
	}
	transport.WriteMessage(initResponse)

	c := New(transport)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.Connect(ctx)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	// Verify initialize request was sent
	initRequest, err := transport.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read init request: %v", err)
	}

	if initRequest.Method != "initialize" {
		t.Errorf("expected method 'initialize', got '%s'", initRequest.Method)
	}
}

func TestClient_ListTools(t *testing.T) {
	t.Skip("Mock transport needs improvement for async message handling")
}

func TestClient_CallTool(t *testing.T) {
	t.Skip("Mock transport needs improvement for async message handling")
}

func TestClient_ListResources(t *testing.T) {
	t.Skip("Mock transport needs improvement for async message handling")
}

func TestClient_ReadResource(t *testing.T) {
	t.Skip("Mock transport needs improvement for async message handling")
}

func TestClient_ListPrompts(t *testing.T) {
	t.Skip("Mock transport needs improvement for async message handling")
}

func TestClient_Close(t *testing.T) {
	transport := testutil.NewMockTransport()
	c := New(transport)

	err := c.Close()
	if err != nil {
		t.Fatalf("failed to close: %v", err)
	}

	if !transport.Closed {
		t.Error("expected transport to be closed")
	}
}

func TestClient_ContextTimeout(t *testing.T) {
	t.Skip("Mock transport needs improvement for async message handling")
}
