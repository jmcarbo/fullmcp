package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	transport := New("http://localhost:8080")

	if transport.url != "http://localhost:8080" {
		t.Errorf("expected URL 'http://localhost:8080', got '%s'", transport.url)
	}

	if transport.client == nil {
		t.Error("expected client to be initialized")
	}
}

func TestNew_WithHTTPClient(t *testing.T) {
	customClient := &http.Client{}
	transport := New("http://localhost:8080", WithHTTPClient(customClient))

	if transport.client != customClient {
		t.Error("expected custom client to be used")
	}
}

func TestTransport_Connect(t *testing.T) {
	transport := New("http://localhost:8080")

	ctx := context.Background()
	conn, err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if conn == nil {
		t.Fatal("expected connection to be created")
	}

	defer conn.Close()

	// Verify it implements ReadWriteCloser
	var _ io.ReadWriteCloser = conn
}

func TestTransport_Close(t *testing.T) {
	transport := New("http://localhost:8080")

	err := transport.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestHTTPConn_Write_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Read request body
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"test": "data"}` {
			t.Errorf("unexpected request body: %s", body)
		}

		// Write response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"response": "ok"}`))
	}))
	defer server.Close()

	transport := New(server.URL)
	ctx := context.Background()
	conn, _ := transport.Connect(ctx)
	defer conn.Close()

	// Write request
	n, err := conn.Write([]byte(`{"test": "data"}`))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if n != len(`{"test": "data"}`) {
		t.Errorf("expected %d bytes written, got %d", len(`{"test": "data"}`), n)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	response := string(buf[:n])
	expected := `{"response": "ok"}`
	if response != expected {
		t.Errorf("expected response '%s', got '%s'", expected, response)
	}
}

func TestHTTPConn_Write_Error(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	transport := New(server.URL)
	ctx := context.Background()
	conn, _ := transport.Connect(ctx)
	defer conn.Close()

	_, err := conn.Write([]byte(`{"test": "data"}`))
	if err == nil {
		t.Error("expected error for server error response")
	}
}

func TestHTTPConn_Read(t *testing.T) {
	transport := New("http://localhost:8080")
	ctx := context.Background()
	conn, _ := transport.Connect(ctx)
	defer conn.Close()

	httpConn := conn.(*httpConn)

	// Write some data to the buffer
	testData := []byte("test data")
	httpConn.buf.Write(testData)

	// Read the data
	buf := make([]byte, len(testData))
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if n != len(testData) {
		t.Errorf("expected %d bytes read, got %d", len(testData), n)
	}

	if !bytes.Equal(buf, testData) {
		t.Errorf("expected '%s', got '%s'", testData, buf)
	}
}

func TestHTTPConn_Close(t *testing.T) {
	transport := New("http://localhost:8080")
	ctx := context.Background()
	conn, _ := transport.Connect(ctx)

	err := conn.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestNewMCPHandler(t *testing.T) {
	handleFunc := func(ctx context.Context, data []byte) ([]byte, error) {
		return []byte(`{"result": "ok"}`), nil
	}

	handler := NewMCPHandler(handleFunc)
	if handler == nil {
		t.Fatal("expected handler to be created")
	}

	if handler.handleFunc == nil {
		t.Error("expected handleFunc to be set")
	}
}

func TestMCPHandler_ServeHTTP_Success(t *testing.T) {
	handleFunc := func(ctx context.Context, data []byte) ([]byte, error) {
		// Echo the data back
		return data, nil
	}

	handler := NewMCPHandler(handleFunc)

	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader([]byte(`{"test": "data"}`)))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	body := w.Body.String()
	expected := `{"test": "data"}`
	if body != expected {
		t.Errorf("expected body '%s', got '%s'", expected, body)
	}
}

func TestMCPHandler_ServeHTTP_MethodNotAllowed(t *testing.T) {
	handleFunc := func(ctx context.Context, data []byte) ([]byte, error) {
		return []byte(`{"result": "ok"}`), nil
	}

	handler := NewMCPHandler(handleFunc)

	req := httptest.NewRequest("GET", "/mcp", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestMCPHandler_ServeHTTP_HandleFuncError(t *testing.T) {
	handleFunc := func(ctx context.Context, data []byte) ([]byte, error) {
		return nil, io.ErrUnexpectedEOF
	}

	handler := NewMCPHandler(handleFunc)

	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader([]byte(`{"test": "data"}`)))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestNewServer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := NewServer(":8080", handler)
	if server == nil {
		t.Fatal("expected server to be created")
	}

	if server.addr != ":8080" {
		t.Errorf("expected addr ':8080', got '%s'", server.addr)
	}

	if server.handler == nil {
		t.Error("expected handler to be set")
	}
}
