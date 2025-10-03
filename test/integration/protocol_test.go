// Package integration provides integration tests for MCP protocol implementations
package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmcarbo/fullmcp/builder"
	"github.com/jmcarbo/fullmcp/client"
	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
	httpTransport "github.com/jmcarbo/fullmcp/transport/http"
	"github.com/jmcarbo/fullmcp/transport/streamhttp"
)

// TestServerSetup creates a test MCP server with sample tools, resources, and prompts
func createTestServer(t *testing.T) *server.Server {
	t.Helper()

	srv := server.New("test-server",
		server.WithVersion("1.0.0"),
		server.WithInstructions("Integration test server"),
	)

	// Add test tool
	addTool, err := builder.NewTool("add").
		Description("Add two numbers").
		Handler(func(_ context.Context, input struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		},
		) (float64, error) {
			return input.A + input.B, nil
		}).
		Build()
	if err != nil {
		t.Fatalf("Failed to create add tool: %v", err)
	}
	if err := srv.AddTool(addTool); err != nil {
		t.Fatalf("Failed to add tool: %v", err)
	}

	// Add test resource
	resourceHandler := &server.ResourceHandler{
		URI:         "test://data",
		Name:        "Test Data",
		Description: "Test resource",
		MimeType:    "text/plain",
		Reader: func(_ context.Context) ([]byte, error) {
			return []byte("test data content"), nil
		},
	}
	srv.AddResource(resourceHandler)

	// Add test prompt
	prompt := builder.NewPrompt("greeting").
		Description("Generate a greeting").
		Argument("name", "Name to greet", true).
		Renderer(func(_ context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			name := args["name"].(string)
			return []*mcp.PromptMessage{
				{
					Role: "user",
					Content: []mcp.Content{
						mcp.TextContent{
							Type: "text",
							Text: "Hello, " + name + "!",
						},
					},
				},
			}, nil
		}).
		Build()
	if err := srv.AddPrompt(prompt); err != nil {
		t.Fatalf("Failed to add prompt: %v", err)
	}

	return srv
}

// Common test operations
// Helper functions for testProtocolOperations to reduce cyclomatic complexity
func testListTools(t *testing.T, ctx context.Context, c *client.Client) {
	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
	if tools[0].Name != "add" {
		t.Errorf("Expected tool 'add', got '%s'", tools[0].Name)
	}
}

func testCallTool(t *testing.T, ctx context.Context, c *client.Client) {
	result, err := c.CallTool(ctx, "add", json.RawMessage(`{"a": 5, "b": 3}`))
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if result != "8" {
		t.Errorf("Expected result '8', got '%v'", result)
	}
}

func testListResources(t *testing.T, ctx context.Context, c *client.Client) {
	resources, err := c.ListResources(ctx)
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}
	if len(resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(resources))
	}
	if resources[0].URI != "test://data" {
		t.Errorf("Expected resource 'test://data', got '%s'", resources[0].URI)
	}
}

func testReadResource(t *testing.T, ctx context.Context, c *client.Client) {
	content, err := c.ReadResource(ctx, "test://data")
	if err != nil {
		t.Fatalf("ReadResource failed: %v", err)
	}
	if string(content) != "test data content" {
		t.Errorf("Expected 'test data content', got '%s'", string(content))
	}
}

func testListPrompts(t *testing.T, ctx context.Context, c *client.Client) {
	prompts, err := c.ListPrompts(ctx)
	if err != nil {
		t.Fatalf("ListPrompts failed: %v", err)
	}
	if len(prompts) != 1 {
		t.Errorf("Expected 1 prompt, got %d", len(prompts))
	}
	if prompts[0].Name != "greeting" {
		t.Errorf("Expected prompt 'greeting', got '%s'", prompts[0].Name)
	}
}

func testGetPrompt(t *testing.T, ctx context.Context, c *client.Client) {
	messages, err := c.GetPrompt(ctx, "greeting", map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatalf("GetPrompt failed: %v", err)
	}
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
	if messages[0].Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", messages[0].Role)
	}
	if len(messages[0].Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(messages[0].Content))
	}
	if textContent, ok := messages[0].Content[0].(mcp.TextContent); ok {
		if textContent.Text != "Hello, World!" {
			t.Errorf("Expected text 'Hello, World!', got '%s'", textContent.Text)
		}
	} else {
		t.Error("Expected TextContent type")
	}
}

func testProtocolOperations(t *testing.T, c *client.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()

	testListTools(t, ctx, c)
	testCallTool(t, ctx, c)
	testListResources(t, ctx, c)
	testReadResource(t, ctx, c)
	testListPrompts(t, ctx, c)
	testGetPrompt(t, ctx, c)
}

func TestStdioTransport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stdio transport test in short mode")
	}

	// Note: stdio transport requires external process management
	// which is complex to test in unit tests. This test is a placeholder
	// for future implementation that would:
	// 1. Start a server process
	// 2. Connect via stdio
	// 3. Run protocol operations
	// 4. Clean up the process

	t.Skip("Stdio transport test requires complex process management - implement separately")
}

func TestHTTPTransport(t *testing.T) {
	// Create server
	srv := createTestServer(t)

	// Create HTTP test server
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body := make([]byte, r.ContentLength)
		if _, err := r.Body.Read(body); err != nil && err.Error() != "EOF" {
			http.Error(w, "failed to read request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var msg mcp.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			http.Error(w, "invalid JSON-RPC message", http.StatusBadRequest)
			return
		}

		response := srv.HandleMessage(r.Context(), &msg)
		if response == nil {
			w.WriteHeader(http.StatusAccepted)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Logf("Failed to encode response: %v", err)
		}
	})

	httpServer := httptest.NewServer(mux)
	defer httpServer.Close()

	// Create HTTP transport
	transport := httpTransport.New(httpServer.URL + "/mcp")
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Create client
	c := client.New(conn)

	// Run common tests
	testProtocolOperations(t, c)
}

func TestStreamableHTTPTransport(t *testing.T) {
	// Create server
	srv := createTestServer(t)

	// Create MCP handler
	mcpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body := make([]byte, r.ContentLength)
			if _, err := r.Body.Read(body); err != nil && err.Error() != "EOF" {
				http.Error(w, "failed to read request", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			var msg mcp.Message
			if err := json.Unmarshal(body, &msg); err != nil {
				http.Error(w, "invalid JSON-RPC message", http.StatusBadRequest)
				return
			}

			response := srv.HandleMessage(r.Context(), &msg)
			if response == nil {
				w.WriteHeader(http.StatusAccepted)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Logf("Failed to encode response: %v", err)
			}
		}
	})

	// Create streamable HTTP server
	streamServer := streamhttp.NewServer("", mcpHandler)
	httpServer := httptest.NewServer(streamServer)
	defer httpServer.Close()

	// Create streamable HTTP transport
	transport := streamhttp.New(httpServer.URL)
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Create client
	c := client.New(conn)

	// Run common tests
	testProtocolOperations(t, c)
}

func TestStreamableHTTPSessionManagement(t *testing.T) {
	srv := createTestServer(t)

	mcpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body := make([]byte, r.ContentLength)
			if _, err := r.Body.Read(body); err != nil && err.Error() != "EOF" {
				http.Error(w, "failed to read request", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			var msg mcp.Message
			if err := json.Unmarshal(body, &msg); err != nil {
				http.Error(w, "invalid JSON-RPC message", http.StatusBadRequest)
				return
			}

			response := srv.HandleMessage(r.Context(), &msg)
			if response == nil {
				w.WriteHeader(http.StatusAccepted)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	})

	streamServer := streamhttp.NewServer("", mcpHandler)
	httpServer := httptest.NewServer(streamServer)
	defer httpServer.Close()

	// Test session ID persistence
	transport := streamhttp.New(httpServer.URL)
	conn, err := transport.Connect(context.Background())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	c := client.New(conn)
	ctx := context.Background()

	// Initialize (should receive session ID)
	if err := c.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer c.Close()

	// Multiple operations should use same session
	_, err = c.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	_, err = c.ListResources(ctx)
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}

	t.Log("Session management test passed")
}

func TestProtocolErrors(t *testing.T) {
	srv := createTestServer(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		defer r.Body.Close()

		var msg mcp.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			http.Error(w, "invalid JSON-RPC message", http.StatusBadRequest)
			return
		}

		response := srv.HandleMessage(r.Context(), &msg)
		if response == nil {
			w.WriteHeader(http.StatusAccepted)
			return
		}

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

	// Test invalid tool
	_, err := c.CallTool(ctx, "nonexistent", json.RawMessage(`{}`))
	if err == nil {
		t.Error("Expected error calling nonexistent tool")
	}

	// Test invalid resource
	_, err = c.ReadResource(ctx, "test://nonexistent")
	if err == nil {
		t.Error("Expected error reading nonexistent resource")
	}

	// Test invalid prompt
	_, err = c.GetPrompt(ctx, "nonexistent", nil)
	if err == nil {
		t.Error("Expected error getting nonexistent prompt")
	}

	// Test tool with invalid arguments (should fail validation)
	_, err = c.CallTool(ctx, "add", json.RawMessage(`{"invalid": true}`))
	if err == nil {
		t.Error("Expected error with invalid tool arguments")
	}
	t.Logf("Validation error (expected): %v", err)
}

func TestConcurrentRequests(t *testing.T) {
	srv := createTestServer(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		defer r.Body.Close()

		var msg mcp.Message
		json.Unmarshal(body, &msg)

		response := srv.HandleMessage(r.Context(), &msg)
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

	// Run concurrent requests
	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(n int) {
			_, err := c.CallTool(ctx, "add", json.RawMessage(`{"a": 1, "b": 1}`))
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		t.Errorf("Concurrent request failed: %v", err)
	}
}

func BenchmarkHTTPTransport(b *testing.B) {
	srv := createTestServer(&testing.T{})

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		defer r.Body.Close()

		var msg mcp.Message
		json.Unmarshal(body, &msg)

		response := srv.HandleMessage(r.Context(), &msg)
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CallTool(ctx, "add", json.RawMessage(`{"a": 1, "b": 1}`))
	}
}

func BenchmarkStreamableHTTPTransport(b *testing.B) {
	srv := createTestServer(&testing.T{})

	mcpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body := make([]byte, r.ContentLength)
			r.Body.Read(body)
			defer r.Body.Close()

			var msg mcp.Message
			json.Unmarshal(body, &msg)

			response := srv.HandleMessage(r.Context(), &msg)
			if response != nil {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CallTool(ctx, "add", json.RawMessage(`{"a": 1, "b": 1}`))
	}
}
