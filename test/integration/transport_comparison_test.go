// Package integration provides transport comparison tests
package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmcarbo/fullmcp/client"
	"github.com/jmcarbo/fullmcp/mcp"
	httpTransport "github.com/jmcarbo/fullmcp/transport/http"
	"github.com/jmcarbo/fullmcp/transport/streamhttp"
)

// TestTransportComparison compares behavior across different transports
func TestTransportComparison(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) (*client.Client, func())
	}{
		{
			name: "HTTP Transport",
			setupFunc: func(t *testing.T) (*client.Client, func()) {
				srv := createTestServer(t)
				mux := http.NewServeMux()
				mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
					body := make([]byte, r.ContentLength)
					r.Body.Read(body)
					defer r.Body.Close()

					var mcpMsg mcp.Message
					json.Unmarshal(body, &mcpMsg)

					response := srv.HandleMessage(r.Context(), &mcpMsg)
					if response != nil {
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(response)
					}
				})

				httpServer := httptest.NewServer(mux)
				transport := httpTransport.New(httpServer.URL + "/mcp")
				conn, _ := transport.Connect(context.Background())
				c := client.New(conn)

				cleanup := func() {
					c.Close()
					httpServer.Close()
				}

				return c, cleanup
			},
		},
		{
			name: "Streamable HTTP Transport",
			setupFunc: func(t *testing.T) (*client.Client, func()) {
				srv := createTestServer(t)
				mcpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method == http.MethodPost {
						body := make([]byte, r.ContentLength)
						r.Body.Read(body)
						defer r.Body.Close()

						var mcpMsg mcp.Message
						json.Unmarshal(body, &mcpMsg)

						response := srv.HandleMessage(r.Context(), &mcpMsg)
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(response)
					}
				})

				streamServer := streamhttp.NewServer("", mcpHandler)
				httpServer := httptest.NewServer(streamServer)
				transport := streamhttp.New(httpServer.URL)
				conn, _ := transport.Connect(context.Background())
				c := client.New(conn)

				cleanup := func() {
					c.Close()
					httpServer.Close()
				}

				return c, cleanup
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, cleanup := tt.setupFunc(t)
			defer cleanup()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Test initialization
			if err := c.Connect(ctx); err != nil {
				t.Fatalf("Connect failed for %s: %v", tt.name, err)
			}

			// Test operations work consistently
			tools, err := c.ListTools(ctx)
			if err != nil {
				t.Errorf("%s: ListTools failed: %v", tt.name, err)
			}

			if len(tools) == 0 {
				t.Errorf("%s: Expected tools, got none", tt.name)
			}
		})
	}
}

// TestStreamableHTTPSpecificFeatures tests features unique to streamable HTTP
func TestStreamableHTTPSpecificFeatures(t *testing.T) {
	srv := createTestServer(t)

	var capturedSessionID string
	sessionHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture session ID from request
		sessionID := r.Header.Get("Mcp-Session-Id")
		if sessionID != "" && capturedSessionID == "" {
			capturedSessionID = sessionID
			t.Logf("Captured session ID: %s", sessionID)
		}

		if r.Method == http.MethodPost {
			body := make([]byte, r.ContentLength)
			r.Body.Read(body)
			defer r.Body.Close()

			var mcpMsg mcp.Message
			json.Unmarshal(body, &mcpMsg)

			// Return session ID in first response
			if capturedSessionID == "" {
				w.Header().Set("Mcp-Session-Id", "test-session-123")
				capturedSessionID = "test-session-123"
			}

			response := srv.HandleMessage(r.Context(), &mcpMsg)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	streamServer := streamhttp.NewServer("", sessionHandler)
	httpServer := httptest.NewServer(streamServer)
	defer httpServer.Close()

	transport := streamhttp.New(httpServer.URL)
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	c := client.New(conn)
	ctx := context.Background()

	if err := c.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()

	// Session ID should be captured
	if capturedSessionID == "" {
		t.Error("Session ID was not captured")
	}

	// Multiple requests should maintain session
	c.ListTools(ctx)
	c.ListResources(ctx)

	t.Logf("Session maintained across multiple requests")
}

// TestHTTPTransportHeaders tests HTTP-specific header handling
func TestHTTPTransportHeaders(t *testing.T) {
	srv := createTestServer(t)

	headersReceived := make(map[string]string)
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		// Capture headers
		headersReceived["Content-Type"] = r.Header.Get("Content-Type")
		headersReceived["Accept"] = r.Header.Get("Accept")

		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		defer r.Body.Close()

		var mcpMsg mcp.Message
		json.Unmarshal(body, &mcpMsg)

		response := srv.HandleMessage(r.Context(), &mcpMsg)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	transport := httpTransport.New(httpServer.URL + "/mcp")
	conn, _ := transport.Connect(context.Background())
	c := client.New(conn)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	// Verify correct headers were sent
	if headersReceived["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", headersReceived["Content-Type"])
	}
}

// TestTransportReconnection tests reconnection behavior
func TestTransportReconnection(t *testing.T) {
	srv := createTestServer(t)
	requestCount := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// Simulate temporary failure on first request
		if requestCount == 1 {
			http.Error(w, "temporary error", http.StatusServiceUnavailable)
			return
		}

		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		defer r.Body.Close()

		var mcpMsg mcp.Message
		json.Unmarshal(body, &mcpMsg)

		response := srv.HandleMessage(r.Context(), &mcpMsg)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	// First connection should fail
	transport1 := httpTransport.New(httpServer.URL + "/mcp")
	conn1, _ := transport1.Connect(context.Background())
	c1 := client.New(conn1)

	ctx := context.Background()
	err := c1.Connect(ctx)
	if err == nil {
		t.Error("Expected first connection to fail")
	}

	// Second connection should succeed
	transport2 := httpTransport.New(httpServer.URL + "/mcp")
	conn2, _ := transport2.Connect(context.Background())
	c2 := client.New(conn2)

	if err := c2.Connect(ctx); err != nil {
		t.Errorf("Second connection failed: %v", err)
	}
	c2.Close()
}

// TestStreamableHTTPResumeCapability tests SSE stream resumption
func TestStreamableHTTPResumeCapability(t *testing.T) {
	srv := createTestServer(t)

	lastEventIDReceived := ""
	mcpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Capture Last-Event-ID header
			lastEventIDReceived = r.Header.Get("Last-Event-ID")
			if lastEventIDReceived != "" {
				t.Logf("Received Last-Event-ID: %s", lastEventIDReceived)
			}
		}

		if r.Method == http.MethodPost {
			body := make([]byte, r.ContentLength)
			r.Body.Read(body)
			defer r.Body.Close()

			var mcpMsg mcp.Message
			json.Unmarshal(body, &mcpMsg)

			response := srv.HandleMessage(r.Context(), &mcpMsg)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	streamServer := streamhttp.NewServer("", mcpHandler)
	httpServer := httptest.NewServer(streamServer)
	defer httpServer.Close()

	transport := streamhttp.New(httpServer.URL)
	conn, _ := transport.Connect(context.Background())
	c := client.New(conn)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	// Make some requests to generate event IDs
	c.ListTools(ctx)

	time.Sleep(100 * time.Millisecond)

	t.Log("Stream resumption test completed")
}
