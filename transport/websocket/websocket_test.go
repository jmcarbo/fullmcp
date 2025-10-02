package websocket

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNew(t *testing.T) {
	transport := New("ws://localhost:8080")
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
	if transport.url != "ws://localhost:8080" {
		t.Errorf("expected url ws://localhost:8080, got %s", transport.url)
	}
}

func TestWithDialer(t *testing.T) {
	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	transport := New("ws://localhost:8080", WithDialer(dialer))
	if transport.dialer != dialer {
		t.Error("expected custom dialer to be set")
	}
}

func TestWithHeaders(t *testing.T) {
	headers := http.Header{
		"Authorization": []string{"Bearer token123"},
	}
	transport := New("ws://localhost:8080", WithHeaders(headers))
	if transport.headers.Get("Authorization") != "Bearer token123" {
		t.Error("expected custom headers to be set")
	}
}

func TestServerEcho(t *testing.T) {
	// Create echo handler
	handler := func(ctx context.Context, msg []byte) ([]byte, error) {
		return msg, nil
	}

	// Create server
	server := NewServer(":0", handler)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer httpServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	// Create client transport
	transport := New(wsURL)
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Test echo
	testMsg := []byte(`{"jsonrpc":"2.0","method":"test","id":1}`)
	n, err := conn.Write(testMsg)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	if n != len(testMsg) {
		t.Errorf("expected to write %d bytes, wrote %d", len(testMsg), n)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	response := string(buf[:n])
	expected := string(testMsg)
	if response != expected {
		t.Errorf("expected response %s, got %s", expected, response)
	}
}

func TestServerError(t *testing.T) {
	// Create handler that returns error
	handler := func(ctx context.Context, msg []byte) ([]byte, error) {
		return nil, fmt.Errorf("test error")
	}

	// Create server
	server := NewServer(":0", handler)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer httpServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	// Create client transport
	transport := New(wsURL)
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Send message
	testMsg := []byte(`{"jsonrpc":"2.0","method":"test","id":1}`)
	_, err = conn.Write(testMsg)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Read error response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	response := string(buf[:n])
	if !strings.Contains(response, "test error") {
		t.Errorf("expected error message to contain 'test error', got %s", response)
	}
}

func TestMultipleMessages(t *testing.T) {
	// Create echo handler
	handler := func(ctx context.Context, msg []byte) ([]byte, error) {
		return msg, nil
	}

	// Create server
	server := NewServer(":0", handler)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer httpServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	// Create client transport
	transport := New(wsURL)
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Send multiple messages
	for i := 1; i <= 3; i++ {
		msg := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"test","id":%d}`, i))

		_, err := conn.Write(msg)
		if err != nil {
			t.Fatalf("failed to write message %d: %v", i, err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("failed to read message %d: %v", i, err)
		}

		response := string(buf[:n])
		expected := string(msg)
		if response != expected {
			t.Errorf("message %d: expected %s, got %s", i, expected, response)
		}
	}
}

func TestLargeMessage(t *testing.T) {
	// Create echo handler
	handler := func(ctx context.Context, msg []byte) ([]byte, error) {
		return msg, nil
	}

	// Create server
	server := NewServer(":0", handler)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer httpServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	// Create client transport
	transport := New(wsURL)
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Create large message (10KB)
	largeData := strings.Repeat("a", 10000)
	testMsg := []byte(fmt.Sprintf(`{"jsonrpc":"2.0","method":"test","data":"%s","id":1}`, largeData))

	_, err = conn.Write(testMsg)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Read response in chunks
	var response []byte
	buf := make([]byte, 1024)

	for len(response) < len(testMsg) {
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("failed to read: %v", err)
		}
		response = append(response, buf[:n]...)
	}

	if string(response) != string(testMsg) {
		t.Errorf("large message mismatch: expected %d bytes, got %d bytes", len(testMsg), len(response))
	}
}

func TestServerWithCheckOrigin(t *testing.T) {
	// Create handler
	handler := func(ctx context.Context, msg []byte) ([]byte, error) {
		return msg, nil
	}

	// Create server with custom origin checker
	server := NewServer(":0", handler).WithCheckOrigin(func(r *http.Request) bool {
		return r.Header.Get("Origin") == "http://allowed.com"
	})

	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer httpServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	// Test with allowed origin
	headers := http.Header{
		"Origin": []string{"http://allowed.com"},
	}
	transport := New(wsURL, WithHeaders(headers))
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("failed to connect with allowed origin: %v", err)
	}
	_ = conn.Close()

	// Test with disallowed origin
	badHeaders := http.Header{
		"Origin": []string{"http://evil.com"},
	}
	transport2 := New(wsURL, WithHeaders(badHeaders))
	conn2, err := transport2.Connect(context.Background())
	if err == nil {
		_ = conn2.Close()
		t.Error("expected connection to fail with disallowed origin")
	}
}

func TestConnectFailure(t *testing.T) {
	// Try to connect to non-existent server
	transport := New("ws://localhost:99999")
	ctx := context.Background()
	_, err := transport.Connect(ctx)
	if err == nil {
		t.Error("expected connection to fail")
	}
}

func TestCloseTransport(t *testing.T) {
	// Create echo handler
	handler := func(ctx context.Context, msg []byte) ([]byte, error) {
		return msg, nil
	}

	server := NewServer(":0", handler)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	transport := New(wsURL)
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	// Close via connection
	err = conn.Close()
	if err != nil {
		t.Errorf("failed to close connection: %v", err)
	}

	// Close via transport (may error if already closed, which is fine)
	_ = transport.Close()
}

func TestReadAfterPartialBuffer(t *testing.T) {
	// Create echo handler
	handler := func(ctx context.Context, msg []byte) ([]byte, error) {
		return msg, nil
	}

	server := NewServer(":0", handler)
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	transport := New(wsURL)
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Send message
	testMsg := []byte("Hello, World!")
	_, err = conn.Write(testMsg)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Read in small chunks
	buf1 := make([]byte, 5)
	n1, err := conn.Read(buf1)
	if err != nil {
		t.Fatalf("failed to read first chunk: %v", err)
	}

	buf2 := make([]byte, 100)
	n2, err := conn.Read(buf2)
	if err != nil {
		t.Fatalf("failed to read second chunk: %v", err)
	}

	result := string(buf1[:n1]) + string(buf2[:n2])
	if result != string(testMsg) {
		t.Errorf("expected %s, got %s", testMsg, result)
	}
}
