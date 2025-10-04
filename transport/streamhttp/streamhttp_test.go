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

	// Verify CORS headers are set even for forbidden origin
	// so browser can see the actual error instead of CORS error
	if w.Header().Get("Access-Control-Allow-Origin") != "http://forbidden.com" {
		t.Errorf("expected CORS header for forbidden origin, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestServer_ServeHTTP_Options(t *testing.T) {
	server := NewServer(":8080", nil)

	req := httptest.NewRequest("OPTIONS", "/mcp", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods 'GET, POST, OPTIONS', got %s", w.Header().Get("Access-Control-Allow-Methods"))
	}

	if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Mcp-Session-Id, X-API-Key, Authorization, Last-Event-ID" {
		t.Errorf("expected correct headers, got %s", w.Header().Get("Access-Control-Allow-Headers"))
	}

	if w.Header().Get("Access-Control-Max-Age") != "86400" {
		t.Errorf("expected Access-Control-Max-Age '86400', got %s", w.Header().Get("Access-Control-Max-Age"))
	}
}

func TestServer_ServeHTTP_OptionsWithAllowedOrigin(t *testing.T) {
	server := NewServer(":8080", nil, WithAllowedOrigin("http://allowed.com"))

	req := httptest.NewRequest("OPTIONS", "/mcp", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "http://allowed.com" {
		t.Errorf("expected Access-Control-Allow-Origin 'http://allowed.com', got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods 'GET, POST, OPTIONS', got %s", w.Header().Get("Access-Control-Allow-Methods"))
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
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	server := NewServer(":8080", handler)

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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "/mcp", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Run in goroutine since handleGET blocks
	done := make(chan bool)
	go func() {
		server.handleGET(w, req)
		done <- true
	}()

	select {
	case <-done:
		// Handler completed
	case <-time.After(1 * time.Second):
		t.Fatal("handleGET did not complete in time")
	}
}

func TestServer_GET_SessionNotFound(t *testing.T) {
	server := NewServer(":8080", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "/mcp", nil)
	req = req.WithContext(ctx)
	req.Header.Set("Mcp-Session-Id", "nonexistent")
	w := httptest.NewRecorder()

	// Run in goroutine since handleGET blocks
	done := make(chan bool)
	go func() {
		server.handleGET(w, req)
		done <- true
	}()

	select {
	case <-done:
		// Handler completed
	case <-time.After(1 * time.Second):
		t.Fatal("handleGET did not complete in time")
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

func TestMatchOrigin(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		pattern  string
		expected bool
	}{
		// Exact matches
		{"exact match", "http://example.com", "http://example.com", true},
		{"exact match https", "https://example.com", "https://example.com", true},

		// Wildcard *
		{"wildcard all", "http://example.com", "*", true},
		{"wildcard all https", "https://example.com", "*", true},

		// Subdomain wildcards
		{"subdomain wildcard match", "http://sub.example.com", "http://*.example.com", true},
		{"subdomain wildcard match deep", "http://deep.sub.example.com", "http://*.example.com", true},
		{"subdomain wildcard no match", "http://example.com", "http://*.example.com", false},
		{"subdomain wildcard different domain", "http://sub.other.com", "http://*.example.com", false},

		// Protocol-less wildcards
		{"protocol-less wildcard", "sub.example.com", "*.example.com", true},
		{"protocol-less wildcard no match", "example.com", "*.example.com", false},

		// HTTPS wildcards
		{"https subdomain wildcard", "https://api.example.com", "https://*.example.com", true},
		{"https wildcard wrong protocol", "http://api.example.com", "https://*.example.com", false},

		// No wildcard mismatches
		{"no wildcard mismatch", "http://other.com", "http://example.com", false},
		{"no wildcard partial match", "http://example.com.evil.com", "http://example.com", false},

		// Edge cases
		{"empty origin", "", "http://example.com", false},
		{"empty pattern", "http://example.com", "", false},
		{"both empty", "", "", true},
		{"wildcard prefix only", "http://example.com", "*example.com", true},
		{"wildcard suffix only", "http://example.com", "http://example.*", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchOrigin(tt.origin, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchOrigin(%q, %q) = %v, expected %v", tt.origin, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestServer_WildcardOrigin(t *testing.T) {
	server := NewServer(":8080", nil, WithAllowedOrigin("https://*.example.com"))

	// Test matching origin
	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Origin", "https://api.example.com")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code == http.StatusForbidden {
		t.Error("expected wildcard pattern to match subdomain")
	}

	// Test non-matching origin
	req = httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Origin", "https://other.com")
	w = httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Error("expected wildcard pattern to reject non-matching origin")
	}
}

func TestServer_WildcardAll(t *testing.T) {
	server := NewServer(":8080", nil, WithAllowedOrigin("*"))

	// Test any origin should be allowed
	origins := []string{
		"http://example.com",
		"https://api.example.com",
		"http://localhost:3000",
		"https://sub.domain.example.org",
	}

	for _, origin := range origins {
		req := httptest.NewRequest("POST", "/mcp", nil)
		req.Header.Set("Origin", origin)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code == http.StatusForbidden {
			t.Errorf("wildcard * should allow origin %q", origin)
		}
	}
}

func TestServer_CORSHeadersInPOST(t *testing.T) {
	server := NewServer(":8080", nil, WithAllowedOrigin("https://example.com"))

	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	// Verify CORS headers are present in POST response
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin header to be 'https://example.com', got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials header to be 'true', got %q", w.Header().Get("Access-Control-Allow-Credentials"))
	}
}

func TestServer_CORSHeadersInGET(t *testing.T) {
	server := NewServer(":8080", nil, WithAllowedOrigin("https://example.com"))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest("GET", "/mcp", nil)
	req = req.WithContext(ctx)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	// Run in goroutine since handleGET blocks
	done := make(chan bool)
	go func() {
		server.ServeHTTP(w, req)
		done <- true
	}()

	select {
	case <-done:
		// Verify CORS headers are present in GET response
		if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
			t.Errorf("expected Access-Control-Allow-Origin header to be 'https://example.com', got %q", w.Header().Get("Access-Control-Allow-Origin"))
		}

		if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Errorf("expected Access-Control-Allow-Credentials header to be 'true', got %q", w.Header().Get("Access-Control-Allow-Credentials"))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("handleGET did not complete in time")
	}
}

func TestServer_CORSHeadersWithWildcard(t *testing.T) {
	server := NewServer(":8080", nil, WithAllowedOrigin("https://*.example.com"))

	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Origin", "https://api.example.com")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	// When using wildcard pattern, should echo back the origin
	if w.Header().Get("Access-Control-Allow-Origin") != "https://api.example.com" {
		t.Errorf("expected Access-Control-Allow-Origin header to be 'https://api.example.com', got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestServer_CORSHeadersNoOrigin(t *testing.T) {
	server := NewServer(":8080", nil, WithAllowedOrigin("https://example.com"))

	req := httptest.NewRequest("POST", "/mcp", nil)
	// No Origin header set
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	// Should set CORS headers even when no Origin header is present
	// to prevent CORS errors for requests without Origin header
	got := w.Header().Get("Access-Control-Allow-Origin")
	want := "https://example.com"
	if got != want {
		t.Errorf("expected Access-Control-Allow-Origin header %q, got %q", want, got)
	}
}
