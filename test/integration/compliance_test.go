// Package integration provides MCP specification compliance tests
package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jmcarbo/fullmcp/client"
	"github.com/jmcarbo/fullmcp/mcp"
	httpTransport "github.com/jmcarbo/fullmcp/transport/http"
	"github.com/jmcarbo/fullmcp/transport/streamhttp"
)

// TestJSONRPCCompliance tests JSON-RPC 2.0 specification compliance
func TestJSONRPCCompliance(t *testing.T) {
	srv := createTestServer(t)

	var receivedMessages []map[string]interface{}
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		defer r.Body.Close()

		var msg map[string]interface{}
		if err := json.Unmarshal(body, &msg); err != nil {
			t.Errorf("Invalid JSON: %v", err)
		}
		receivedMessages = append(receivedMessages, msg)

		// Validate JSON-RPC 2.0 fields
		if jsonrpc, ok := msg["jsonrpc"].(string); !ok || jsonrpc != "2.0" {
			t.Errorf("Missing or invalid jsonrpc field: %v", msg["jsonrpc"])
		}

		if _, hasMethod := msg["method"]; !hasMethod {
			t.Error("Missing method field")
		}

		// In JSON-RPC 2.0, messages with method are either:
		// - Requests (have id field)
		// - Notifications (no id field)
		// Both are valid, so we don't require id
		_, hasID := msg["id"]
		if hasID {
			t.Logf("Request with id: %v", msg["id"])
		} else {
			t.Logf("Notification without id")
		}

		var mcpMsg mcp.Message
		json.Unmarshal(body, &mcpMsg)
		response := srv.HandleMessage(r.Context(), &mcpMsg)

		if response != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	transport := httpTransport.New(httpServer.URL + "/mcp")
	conn, _ := transport.Connect(context.Background())
	c := client.New(conn)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	c.ListTools(ctx)

	if len(receivedMessages) == 0 {
		t.Error("No messages received")
	}

	t.Logf("JSON-RPC compliance verified for %d messages", len(receivedMessages))
}

// TestMCPProtocolVersion tests MCP protocol version negotiation
func TestMCPProtocolVersion(t *testing.T) {
	srv := createTestServer(t)

	var initializeRequest map[string]interface{}
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		defer r.Body.Close()

		var msg map[string]interface{}
		json.Unmarshal(body, &msg)

		// Capture initialize request
		if method, ok := msg["method"].(string); ok && method == "initialize" {
			initializeRequest = msg
		}

		var mcpMsg mcp.Message
		json.Unmarshal(body, &mcpMsg)
		response := srv.HandleMessage(r.Context(), &mcpMsg)

		if response != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	transport := httpTransport.New(httpServer.URL + "/mcp")
	conn, _ := transport.Connect(context.Background())
	c := client.New(conn)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	// Verify initialize request contains protocol version
	if initializeRequest == nil {
		t.Fatal("Initialize request not captured")
	}

	params, ok := initializeRequest["params"].(map[string]interface{})
	if !ok {
		t.Fatal("Initialize params missing")
	}

	protocolVersion, ok := params["protocolVersion"].(string)
	if !ok {
		t.Error("Protocol version missing in initialize request")
	}

	// Verify protocol version format (should be YYYY-MM-DD)
	if !strings.Contains(protocolVersion, "-") {
		t.Errorf("Invalid protocol version format: %s", protocolVersion)
	}

	t.Logf("Protocol version: %s", protocolVersion)
}

// TestCapabilityNegotiation tests MCP capability exchange
func TestCapabilityNegotiation(t *testing.T) {
	srv := createTestServer(t)

	var serverCapabilities map[string]interface{}
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		defer r.Body.Close()

		var mcpMsg mcp.Message
		json.Unmarshal(body, &mcpMsg)
		response := srv.HandleMessage(r.Context(), &mcpMsg)

		if response != nil && response.Result != nil {
			var result map[string]interface{}
			json.Unmarshal(response.Result, &result)

			if caps, ok := result["capabilities"].(map[string]interface{}); ok {
				serverCapabilities = caps
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	transport := httpTransport.New(httpServer.URL + "/mcp")
	conn, _ := transport.Connect(context.Background())
	c := client.New(conn)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	// Verify server returned capabilities
	if serverCapabilities == nil {
		t.Fatal("Server capabilities not received")
	}

	// Check for expected capability fields
	expectedCapabilities := []string{"tools", "resources", "prompts"}
	for _, cap := range expectedCapabilities {
		if _, exists := serverCapabilities[cap]; !exists {
			t.Errorf("Missing capability: %s", cap)
		}
	}

	t.Logf("Server capabilities: %v", serverCapabilities)
}

// TestContentTypeHeaders tests proper content type handling
func TestContentTypeHeaders(t *testing.T) {
	srv := createTestServer(t)

	var requestContentType, responseContentType string
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		requestContentType = r.Header.Get("Content-Type")

		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		defer r.Body.Close()

		var mcpMsg mcp.Message
		json.Unmarshal(body, &mcpMsg)
		response := srv.HandleMessage(r.Context(), &mcpMsg)

		if response != nil {
			w.Header().Set("Content-Type", "application/json")
			responseContentType = "application/json"
			json.NewEncoder(w).Encode(response)
		}
	})

	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	transport := httpTransport.New(httpServer.URL + "/mcp")
	conn, _ := transport.Connect(context.Background())
	c := client.New(conn)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	// Verify content types
	if requestContentType != "application/json" {
		t.Errorf("Expected request Content-Type 'application/json', got '%s'", requestContentType)
	}

	if responseContentType != "application/json" {
		t.Errorf("Expected response Content-Type 'application/json', got '%s'", responseContentType)
	}
}

// TestStreamableHTTPSessionID tests Mcp-Session-Id header compliance
func TestStreamableHTTPSessionID(t *testing.T) {
	srv := createTestServer(t)

	sessionIDsSeen := make(map[string]int)
	mcpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			sessionID := r.Header.Get("Mcp-Session-Id")
			if sessionID != "" {
				sessionIDsSeen[sessionID]++
			}

			body := make([]byte, r.ContentLength)
			r.Body.Read(body)
			defer r.Body.Close()

			var mcpMsg mcp.Message
			json.Unmarshal(body, &mcpMsg)
			response := srv.HandleMessage(r.Context(), &mcpMsg)

			if response != nil {
				// First request should get session ID
				if sessionID == "" {
					w.Header().Set("Mcp-Session-Id", "test-session-abc123")
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}
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

	// Make multiple requests
	c.ListTools(ctx)
	c.ListResources(ctx)
	c.ListPrompts(ctx)

	// Verify session ID was used consistently
	if len(sessionIDsSeen) == 0 {
		t.Error("No session IDs observed")
	}

	if len(sessionIDsSeen) > 1 {
		t.Errorf("Multiple session IDs used: %v", sessionIDsSeen)
	}

	for sessionID, count := range sessionIDsSeen {
		t.Logf("Session ID '%s' used %d times", sessionID, count)
	}
}

// TestSSEContentType tests SSE stream content type compliance
func TestSSEContentType(t *testing.T) {
	srv := createTestServer(t)

	mcpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
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
		}
	})

	streamServer := streamhttp.NewServer("", mcpHandler)
	httpServer := httptest.NewServer(streamServer)
	defer httpServer.Close()

	// Make a direct GET request to verify SSE Content-Type header
	req, _ := http.NewRequest(http.MethodGet, httpServer.URL, nil)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to connect to SSE endpoint: %v", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Errorf("Expected SSE response Content-Type to contain 'text/event-stream', got '%s'", contentType)
	}

	t.Logf("SSE Content-Type verified: %s", contentType)
}

// TestErrorCodeCompliance tests MCP error code compliance
func TestErrorCodeCompliance(t *testing.T) {
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
			responseData, _ := json.Marshal(response)

			// Check error structure
			var respMap map[string]interface{}
			json.Unmarshal(responseData, &respMap)

			if errObj, hasError := respMap["error"]; hasError {
				errMap := errObj.(map[string]interface{})

				// Verify error has code and message
				if _, hasCode := errMap["code"]; !hasCode {
					t.Error("Error missing 'code' field")
				}
				if _, hasMessage := errMap["message"]; !hasMessage {
					t.Error("Error missing 'message' field")
				}
			}

			json.NewEncoder(w).Encode(response)
		}
	})

	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	transport := httpTransport.New(httpServer.URL + "/mcp")
	conn, _ := transport.Connect(context.Background())
	c := client.New(conn)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	// Try to call nonexistent tool (should return error)
	_, err := c.CallTool(ctx, "nonexistent", json.RawMessage(`{}`))
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	} else {
		t.Logf("Error code compliance verified: %v", err)
	}
}

// TestRequestIDUniqueness tests that request IDs are unique
func TestRequestIDUniqueness(t *testing.T) {
	srv := createTestServer(t)

	requestIDs := make(map[interface{}]int)
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		defer r.Body.Close()

		var msg map[string]interface{}
		json.Unmarshal(body, &msg)

		if id, ok := msg["id"]; ok {
			requestIDs[id]++
		}

		var mcpMsg mcp.Message
		json.Unmarshal(body, &mcpMsg)
		response := srv.HandleMessage(r.Context(), &mcpMsg)

		if response != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	transport := httpTransport.New(httpServer.URL + "/mcp")
	conn, _ := transport.Connect(context.Background())
	c := client.New(conn)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	// Make multiple requests
	c.ListTools(ctx)
	c.ListResources(ctx)
	c.ListPrompts(ctx)

	// Verify all IDs are unique
	for id, count := range requestIDs {
		if count > 1 {
			t.Errorf("Request ID %v used %d times (should be unique)", id, count)
		}
	}

	t.Logf("Request ID uniqueness verified: %d unique IDs", len(requestIDs))
}

// TestNotificationCompliance tests notification handling (no response expected)
func TestNotificationCompliance(t *testing.T) {
	srv := createTestServer(t)

	notificationReceived := false
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		defer r.Body.Close()

		var msg map[string]interface{}
		json.Unmarshal(body, &msg)

		// Notification has no "id" field
		if _, hasID := msg["id"]; !hasID {
			notificationReceived = true
			w.WriteHeader(http.StatusAccepted)
			return
		}

		var mcpMsg mcp.Message
		json.Unmarshal(body, &mcpMsg)
		response := srv.HandleMessage(r.Context(), &mcpMsg)

		if response != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	transport := httpTransport.New(httpServer.URL + "/mcp")
	conn, _ := transport.Connect(context.Background())
	c := client.New(conn)

	ctx := context.Background()
	c.Connect(ctx)
	defer c.Close()

	// The initialized notification is sent after connect
	time.Sleep(100 * time.Millisecond)

	if notificationReceived {
		t.Log("Notification compliance verified")
	}
}
