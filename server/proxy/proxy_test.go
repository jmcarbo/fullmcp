package proxy

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/jmcarbo/fullmcp/builder"
	"github.com/jmcarbo/fullmcp/client"
	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

type AddArgs struct {
	A int `json:"a"`
	B int `json:"b"`
}

// mockTransport simulates a bidirectional pipe using channels
type mockTransport struct {
	readCh  chan []byte
	writeCh chan []byte
	closed  chan struct{}
	mu      sync.Mutex
}

func newMockTransportPair() (*mockTransport, *mockTransport) {
	ch1 := make(chan []byte, 100)
	ch2 := make(chan []byte, 100)

	client := &mockTransport{
		readCh:  ch2,
		writeCh: ch1,
		closed:  make(chan struct{}),
	}

	server := &mockTransport{
		readCh:  ch1,
		writeCh: ch2,
		closed:  make(chan struct{}),
	}

	return client, server
}

func (m *mockTransport) Read(p []byte) (int, error) {
	select {
	case <-m.closed:
		return 0, io.EOF
	case data := <-m.readCh:
		n := copy(p, data)
		if n < len(data) {
			// Put remaining data back (simplified - real impl would buffer)
			go func() { m.readCh <- data[n:] }()
		}
		return n, nil
	}
}

func (m *mockTransport) Write(p []byte) (int, error) {
	select {
	case <-m.closed:
		return 0, io.EOF
	case m.writeCh <- append([]byte(nil), p...):
		return len(p), nil
	}
}

func (m *mockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	select {
	case <-m.closed:
		// Already closed
	default:
		close(m.closed)
	}
	return nil
}

func TestProxyServer(t *testing.T) {
	// Create backend server
	backend := server.New("backend-server")

	// Add a tool to backend
	addTool, _ := builder.NewTool("add").
		Description("Add two numbers").
		Handler(func(ctx context.Context, args AddArgs) (int, error) {
			return args.A + args.B, nil
		}).
		Build()

	if err := backend.AddTool(addTool); err != nil {
		t.Fatalf("failed to add tool: %v", err)
	}

	// Create transport pair
	clientConn, serverConn := newMockTransportPair()

	// Start backend server in goroutine
	backendCtx, backendCancel := context.WithCancel(context.Background())
	defer backendCancel()
	go func() {
		_ = backend.Serve(backendCtx, serverConn)
	}()

	// Give server goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Create client connected to backend and connect it
	backendClient := client.New(clientConn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := backendClient.Connect(ctx); err != nil {
		t.Fatalf("failed to connect to backend: %v", err)
	}
	defer func() { _ = backendClient.Close() }()

	// Create proxy server (this will sync from backend)
	proxy, err := New("proxy-server", backendClient)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	// Create mock transport for proxy
	proxyClientConn, proxyServerConn := newMockTransportPair()

	// Start proxy server in goroutine
	proxyCtx, proxyCancel := context.WithCancel(context.Background())
	defer proxyCancel()
	go func() {
		_ = proxy.Serve(proxyCtx, proxyServerConn)
	}()

	// Write initialize message to proxy
	initMsg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}`),
	}

	msgBytes, _ := json.Marshal(initMsg)
	msgBytes = append(msgBytes, '\n')
	_, _ = proxyClientConn.Write(msgBytes)

	// Read initialize response
	var initResp mcp.Message
	_ = json.NewDecoder(proxyClientConn).Decode(&initResp)

	// Test tools/list through proxy
	listMsg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	msgBytes, _ = json.Marshal(listMsg)
	msgBytes = append(msgBytes, '\n')
	_, _ = proxyClientConn.Write(msgBytes)

	var listResp mcp.Message
	if err := json.NewDecoder(proxyClientConn).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify we got tools from backend
	var result struct {
		Tools []*mcp.Tool `json:"tools"`
	}
	if err := json.Unmarshal(listResp.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(result.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(result.Tools))
	}

	if result.Tools[0].Name != "add" {
		t.Errorf("expected tool 'add', got '%s'", result.Tools[0].Name)
	}

	// Test tools/call through proxy
	callMsg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"add","arguments":{"a":5,"b":3}}`),
	}

	msgBytes, _ = json.Marshal(callMsg)
	msgBytes = append(msgBytes, '\n')
	_, _ = proxyClientConn.Write(msgBytes)

	var callResp mcp.Message
	if err := json.NewDecoder(proxyClientConn).Decode(&callResp); err != nil {
		t.Fatalf("failed to decode call response: %v", err)
	}

	// Verify result
	var callResult struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(callResp.Result, &callResult); err != nil {
		t.Fatalf("failed to unmarshal call result: %v", err)
	}

	if len(callResult.Content) == 0 {
		t.Fatal("expected content in response")
	}

	if callResult.Content[0].Text != "8" {
		t.Errorf("expected result '8', got '%s'", callResult.Content[0].Text)
	}

	// Cleanup
	_ = proxyClientConn.Close()
	_ = proxyServerConn.Close()
	_ = clientConn.Close()
	_ = serverConn.Close()
}

func TestProxyWithMultipleCapabilities(t *testing.T) {
	// Create backend server
	backend := server.New("backend-server")

	// Add tool
	addTool, _ := builder.NewTool("add").
		Description("Add two numbers").
		Handler(func(ctx context.Context, args AddArgs) (int, error) {
			return args.A + args.B, nil
		}).
		Build()
	_ = backend.AddTool(addTool)

	// Add resource
	configResource := builder.NewResource("config://app").
		Name("App Config").
		Description("Application configuration").
		MimeType("application/json").
		Reader(func(ctx context.Context) ([]byte, error) {
			return []byte(`{"debug":true}`), nil
		}).
		Build()
	_ = backend.AddResource(configResource)

	// Create transport pair
	clientConn, serverConn := newMockTransportPair()

	// Start backend server in goroutine
	backendCtx, backendCancel := context.WithCancel(context.Background())
	defer backendCancel()
	go func() {
		_ = backend.Serve(backendCtx, serverConn)
	}()

	// Give server goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Create client connected to backend and connect it
	backendClient := client.New(clientConn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := backendClient.Connect(ctx); err != nil {
		t.Fatalf("failed to connect to backend: %v", err)
	}
	defer func() { _ = backendClient.Close() }()

	// Create proxy server
	proxy, err := New("proxy-server", backendClient)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	// Create mock transport for proxy
	proxyClientConn, proxyServerConn := newMockTransportPair()

	// Start proxy server in goroutine
	proxyCtx, proxyCancel := context.WithCancel(context.Background())
	defer proxyCancel()
	doneCh := make(chan struct{})
	go func() {
		_ = proxy.Serve(proxyCtx, proxyServerConn)
		close(doneCh)
	}()

	// Helper to write message
	writeMsg := func(msg *mcp.Message) {
		msgBytes, _ := json.Marshal(msg)
		msgBytes = append(msgBytes, '\n')
		_, _ = proxyClientConn.Write(msgBytes)
	}

	// Helper to read response
	readMsg := func() *mcp.Message {
		var resp mcp.Message
		_ = json.NewDecoder(proxyClientConn).Decode(&resp)
		return &resp
	}

	// Initialize
	writeMsg(&mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}`),
	})
	_ = readMsg()

	// Test resources/list
	writeMsg(&mcp.Message{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "resources/list",
	})

	listResp := readMsg()
	var resourcesResult struct {
		Resources []*mcp.Resource `json:"resources"`
	}
	if err := json.Unmarshal(listResp.Result, &resourcesResult); err != nil {
		t.Fatalf("failed to unmarshal resources: %v", err)
	}

	if len(resourcesResult.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(resourcesResult.Resources))
	}

	// Cleanup
	_ = proxyClientConn.Close()
	_ = proxyServerConn.Close()
	_ = clientConn.Close()
	_ = serverConn.Close()

	// Wait a bit for server to notice close
	select {
	case <-doneCh:
	case <-time.After(time.Second):
		t.Log("proxy server didn't shutdown in time")
	}
}

func TestProxyWithEmptyBackend(t *testing.T) {
	// Create backend server with no tools/resources/prompts
	backend := server.New("empty-backend")

	// Create transport pair
	clientConn, serverConn := newMockTransportPair()

	// Start backend server in goroutine
	backendCtx, backendCancel := context.WithCancel(context.Background())
	defer backendCancel()
	go func() {
		_ = backend.Serve(backendCtx, serverConn)
	}()

	// Give server goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Create client connected to backend and connect it
	backendClient := client.New(clientConn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := backendClient.Connect(ctx); err != nil {
		t.Fatalf("failed to connect to backend: %v", err)
	}
	defer func() { _ = backendClient.Close() }()

	// Create proxy server
	proxy, err := New("proxy-server", backendClient)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	// Proxy should have been created successfully with no capabilities
	if proxy == nil {
		t.Fatal("expected non-nil proxy")
	}

	// Cleanup
	_ = clientConn.Close()
	_ = serverConn.Close()
}

func TestProxyErrorHandling(t *testing.T) {
	// Create a closed connection to simulate backend failure
	closedConn := &closedConn{}

	// Try to create proxy with closed backend
	backendClient := client.New(closedConn)

	// This should fail during sync
	_, err := New("proxy-server", backendClient)
	if err == nil {
		t.Error("expected error when creating proxy with failed backend")
	}
}

type closedConn struct{}

func (c *closedConn) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (c *closedConn) Write(p []byte) (int, error) {
	return 0, io.EOF
}

func (c *closedConn) Close() error {
	return nil
}
