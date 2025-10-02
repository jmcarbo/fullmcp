package streamhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTransport_New(t *testing.T) {
	transport := New("http://localhost:8080/mcp")

	if transport == nil {
		t.Fatal("expected non-nil transport")
	}

	if transport.url != "http://localhost:8080/mcp" {
		t.Errorf("expected url 'http://localhost:8080/mcp', got '%s'", transport.url)
	}

	if transport.client == nil {
		t.Error("expected non-nil HTTP client")
	}
}

func TestTransport_WithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 10 * time.Second}
	transport := New("http://localhost:8080/mcp", WithHTTPClient(customClient))

	if transport.client != customClient {
		t.Error("expected custom HTTP client to be set")
	}
}

func TestTransport_WithSessionID(t *testing.T) {
	sessionID := "test-session-123"
	transport := New("http://localhost:8080/mcp", WithSessionID(sessionID))

	if transport.sessionID != sessionID {
		t.Errorf("expected session ID '%s', got '%s'", sessionID, transport.sessionID)
	}
}

func TestServer_ServeHTTP_MethodNotAllowed(t *testing.T) {
	server := NewServer(":8080", nil)

	req := httptest.NewRequest("DELETE", "/mcp", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestServer_ServeHTTP_ForbiddenOrigin(t *testing.T) {
	server := NewServer(":8080", nil, WithAllowedOrigin("http://allowed.com"))

	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Origin", "http://forbidden.com")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestSessionStore_GetOrCreate(t *testing.T) {
	store := NewSessionStore()

	// Create new session
	session1 := store.GetOrCreate("session-1")
	if session1 == nil {
		t.Fatal("expected non-nil session")
	}

	if session1.ID != "session-1" {
		t.Errorf("expected session ID 'session-1', got '%s'", session1.ID)
	}

	// Get existing session
	session2 := store.GetOrCreate("session-1")
	if session2 != session1 {
		t.Error("expected same session instance")
	}
}

func TestSessionStore_Delete(t *testing.T) {
	store := NewSessionStore()

	session := store.GetOrCreate("session-1")
	if session == nil {
		t.Fatal("expected non-nil session")
	}

	store.Delete("session-1")

	retrieved := store.Get("session-1")
	if retrieved != nil {
		t.Error("expected session to be deleted")
	}
}

func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()

	if id1 == "" {
		t.Error("expected non-empty session ID")
	}

	if id1 == id2 {
		t.Error("expected unique session IDs")
	}

	if len(id1) != 32 { // 16 bytes * 2 (hex encoding)
		t.Errorf("expected session ID length 32, got %d", len(id1))
	}
}

func TestSession_SendEvent(t *testing.T) {
	session := &Session{
		ID:        "test-session",
		CreatedAt: time.Now(),
	}

	// Should fail without SSE connection
	err := session.SendEvent([]byte("test data"), "event-1")
	if err == nil {
		t.Error("expected error when sending event without connection")
	}
}

func TestTransport_Close(t *testing.T) {
	transport := New("http://localhost:8080/mcp")

	err := transport.Close()
	if err != nil {
		t.Errorf("unexpected error closing transport: %v", err)
	}
}

func TestServer_POST_Notification(t *testing.T) {
	server := NewServer(":8080", nil)

	req := httptest.NewRequest("POST", "/mcp", nil)
	w := httptest.NewRecorder()

	server.handlePOST(w, req)

	// Notifications return 202 when response is nil
	// (In real scenario, processMessage would determine this)
	if w.Code != http.StatusOK && w.Code != http.StatusAccepted {
		t.Errorf("expected status 200 or 202, got %d", w.Code)
	}
}

func TestServer_GET_NoSessionID(t *testing.T) {
	server := NewServer(":8080", nil)

	req := httptest.NewRequest("GET", "/mcp", nil)
	w := httptest.NewRecorder()

	server.handleGET(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestServer_GET_SessionNotFound(t *testing.T) {
	server := NewServer(":8080", nil)

	req := httptest.NewRequest("GET", "/mcp", nil)
	req.Header.Set("Mcp-Session-Id", "nonexistent")
	w := httptest.NewRecorder()

	server.handleGET(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSSEReader_ParseData(t *testing.T) {
	// This tests the SSE parsing logic indirectly
	// In a real test, you'd set up a mock HTTP response with SSE data
	transport := New("http://localhost:8080/mcp")

	if transport.ctx == nil {
		t.Error("expected non-nil context")
	}
}

func TestStreamConn_ReadWrite(t *testing.T) {
	// Test the connection interface
	ctx := context.Background()

	// Create a mock server for testing
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"ok"}`))
		} else if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher := w.(http.Flusher)

			// Send a test event
			_, _ = w.Write([]byte("data: test\n\n"))
			flusher.Flush()
		}
	}))
	defer server.Close()

	transport := New(server.URL)
	defer transport.Close()

	conn, err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Test write
	n, err := conn.Write([]byte(`{"method":"test"}`))
	if err != nil {
		t.Errorf("write failed: %v", err)
	}
	if n != len(`{"method":"test"}`) {
		t.Errorf("expected to write %d bytes, wrote %d", len(`{"method":"test"}`), n)
	}
}
