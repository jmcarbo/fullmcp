// Package main demonstrates an MCP server with HTTP transport and authentication.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jmcarbo/fullmcp/auth"
	"github.com/jmcarbo/fullmcp/auth/apikey"
	"github.com/jmcarbo/fullmcp/builder"
	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
	"github.com/jmcarbo/fullmcp/transport/streamhttp"
)

type CalculationInput struct {
	A float64 `json:"a" jsonschema:"description=First number"`
	B float64 `json:"b" jsonschema:"description=Second number"`
}

func setupServer() *server.Server {
	srv := server.New("http-math-server",
		server.WithVersion("1.0.0"),
		server.WithInstructions("Math server with HTTP transport and authentication"),
	)

	addTool, _ := builder.NewTool("add").
		Description("Add two numbers").
		Handler(func(_ context.Context, input CalculationInput) (float64, error) {
			return input.A + input.B, nil
		}).
		Build()
	_ = srv.AddTool(addTool)

	multiplyTool, _ := builder.NewTool("multiply").
		Description("Multiply two numbers").
		Handler(func(_ context.Context, input CalculationInput) (float64, error) {
			return input.A * input.B, nil
		}).
		Build()
	_ = srv.AddTool(multiplyTool)

	calcPrompt := builder.NewPrompt("calculation").
		Description("Create a calculation prompt").
		Argument("operation", "The operation to perform", true).
		Argument("a", "First number", true).
		Argument("b", "Second number", true).
		Renderer(func(_ context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			operation := args["operation"].(string)
			a := args["a"]
			b := args["b"]

			return []*mcp.PromptMessage{
				{
					Role: "user",
					Content: []mcp.Content{
						mcp.TextContent{
							Type: "text",
							Text: fmt.Sprintf("Please %s %v and %v", operation, a, b),
						},
					},
				},
			}, nil
		}).
		Build()
	_ = srv.AddPrompt(calcPrompt)

	return srv
}

func createMCPHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, Authorization")

		// Handle CORS preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request", http.StatusBadRequest)
			return
		}
		defer func() { _ = r.Body.Close() }()

		var msg mcp.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			http.Error(w, "invalid JSON-RPC message", http.StatusBadRequest)
			return
		}

		response := srv.HandleMessage(r.Context(), &msg)
		if response == nil {
			// For notifications, return 202 Accepted with no body
			w.WriteHeader(http.StatusAccepted)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
	}
}

func runStreamingServer(addr, allowedOrigin string, srv *server.Server, authProvider *apikey.Provider) {
	log.Printf("Starting Streamable HTTP MCP server on %s", addr)
	log.Println("Transport: HTTP with Server-Sent Events (SSE)")
	log.Println("Use API key: secret-key-123")

	mcpHandler := createMCPHandler(srv)

	var serverOpts []streamhttp.ServerOption
	if allowedOrigin != "" {
		serverOpts = append(serverOpts, streamhttp.WithAllowedOrigin(allowedOrigin))
	}

	streamServer := streamhttp.NewServer(addr, mcpHandler, serverOpts...)
	authenticatedStreamServer := authProvider.Middleware()(streamServer)

	log.Printf("Starting authenticated streaming server on %s", addr)
	if err := http.ListenAndServe(addr, authenticatedStreamServer); err != nil {
		log.Fatal(err)
	}
}

func runRegularServer(addr string, srv *server.Server, authProvider *apikey.Provider) {
	log.Printf("Starting HTTP MCP server on %s", addr)
	log.Println("Transport: Regular HTTP (POST only)")
	log.Println("Use API key: secret-key-123")
	log.Printf("Example: curl -H 'X-API-Key: secret-key-123' -X POST http://%s/mcp -d '{\"jsonrpc\":\"2.0\",\"method\":\"tools/list\",\"id\":1}'", getHostURL(addr))

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", createMCPHandler(srv))

	handler := authProvider.Middleware()(mux)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func main() {
	addr := flag.String("addr", ":8080", "TCP address to listen on (e.g., :8080 or localhost:3000)")
	useStreaming := flag.Bool("stream", false, "Use HTTP streaming (SSE) transport instead of regular HTTP")
	allowedOrigin := flag.String("origin", "", "Allowed origin for CORS (only used with -stream)")
	flag.Parse()

	authProvider := apikey.New()
	authProvider.AddKey("secret-key-123", auth.Claims{
		Subject: "user-1",
		Email:   "user@example.com",
		Scopes:  []string{"read", "write"},
	})

	srv := setupServer()

	if *useStreaming {
		runStreamingServer(*addr, *allowedOrigin, srv, authProvider)
	} else {
		runRegularServer(*addr, srv, authProvider)
	}
}

// getHostURL returns a user-friendly URL based on the listen address
func getHostURL(addr string) string {
	if addr[0] == ':' {
		return "localhost" + addr
	}
	return addr
}
