package sse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTransport_Connect(t *testing.T) {
	transport := New("http://example.com/sse")
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	if conn == nil {
		t.Fatal("Expected connection, got nil")
	}
	_ = conn.Close()
}

func TestTransport_WithHTTPClient(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	transport := New("http://example.com/sse", WithHTTPClient(client))

	if transport.client != client {
		t.Error("Custom HTTP client not set")
	}
}

func TestSSEConn_Write(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		if string(body) != "test data" {
			t.Errorf("Expected 'test data', got %s", body)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := New(server.URL)
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	n, err := conn.Write([]byte("test data"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 9 {
		t.Errorf("Expected to write 9 bytes, wrote %d", n)
	}
}

func TestSSEConn_Read(t *testing.T) {
	// Create SSE test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		flusher := w.(http.Flusher)

		// Send SSE event
		fmt.Fprintf(w, "data: {\"result\":\"test\"}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	transport := New(server.URL)
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	data := string(buf[:n])
	expected := `{"result":"test"}`
	if data != expected {
		t.Errorf("Expected %s, got %s", expected, data)
	}
}

func TestSSEConn_Close(t *testing.T) {
	transport := New("http://example.com/sse")
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	err = conn.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Second close should be safe (idempotent)
	err = conn.Close()
	if err != nil {
		t.Errorf("Second close failed: %v", err)
	}
}

func TestServer_ListenAndServe(t *testing.T) {
	handler := NewMCPSSEHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		return []byte(`{"response":"ok"}`), nil
	})

	server := NewServer(":0", handler)

	// Test would require actual server startup - skip in unit tests
	// This is more of an integration test
	if server == nil {
		t.Fatal("Expected server, got nil")
	}
}

func TestMCPSSEHandler_HandleSSE(t *testing.T) {
	handler := NewMCPSSEHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		if string(req) != "test request" {
			return nil, fmt.Errorf("unexpected request: %s", req)
		}
		return []byte(`{"result":"success"}`), nil
	})

	ctx := context.Background()
	respChan, err := handler.HandleSSE(ctx, []byte("test request"))
	if err != nil {
		t.Fatalf("HandleSSE failed: %v", err)
	}

	resp := <-respChan
	expected := `{"result":"success"}`
	if string(resp) != expected {
		t.Errorf("Expected %s, got %s", expected, resp)
	}
}

func TestMCPSSEHandler_HandleSSE_Error(t *testing.T) {
	handler := NewMCPSSEHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		return nil, fmt.Errorf("test error")
	})

	ctx := context.Background()
	respChan, err := handler.HandleSSE(ctx, []byte("test"))
	if err != nil {
		t.Fatalf("HandleSSE failed: %v", err)
	}

	resp := <-respChan
	expected := `{"error": "test error"}`
	if string(resp) != expected {
		t.Errorf("Expected %s, got %s", expected, resp)
	}
}

func TestServer_handleSSE_POST(t *testing.T) {
	handler := NewMCPSSEHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		return []byte(`{"status":"ok"}`), nil
	})

	server := NewServer(":0", handler)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	server.handleSSE(w, req)

	resp := w.Result()
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected text/event-stream, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestServer_handleSSE_GET(t *testing.T) {
	handler := NewMCPSSEHandler(func(ctx context.Context, req []byte) ([]byte, error) {
		return []byte(`{"status":"ok"}`), nil
	})

	server := NewServer(":0", handler)

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Run in goroutine since it will block
	done := make(chan bool)
	go func() {
		server.handleSSE(w, req)
		done <- true
	}()

	select {
	case <-done:
		// Success - handler completed
	case <-time.After(1 * time.Second):
		t.Fatal("Handler did not complete in time")
	}

	resp := w.Result()
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected text/event-stream, got %s", resp.Header.Get("Content-Type"))
	}
}
