// Package main demonstrates a Streamable HTTP MCP server with tools, resources, and prompts.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jmcarbo/fullmcp/builder"
	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
	"github.com/jmcarbo/fullmcp/transport/streamhttp"
)

type MathInput struct {
	A float64 `json:"a" jsonschema:"description=First number"`
	B float64 `json:"b" jsonschema:"description=Second number"`
}

type GreetInput struct {
	Name     string `json:"name" jsonschema:"description=Person's name"`
	Language string `json:"language,omitempty" jsonschema:"description=Language code (en, es, fr)"`
}

// setupServer creates and configures the MCP server with tools, resources, and prompts
func setupServer() *server.Server {
	srv := server.New("streamhttp-demo-server",
		server.WithVersion("1.0.0"),
		server.WithInstructions("Demo server with Streamable HTTP transport showcasing tools, resources, and prompts"),
	)

	// Add Tools
	addTool, _ := builder.NewTool("add").
		Description("Add two numbers").
		Handler(func(_ context.Context, input MathInput) (float64, error) {
			return input.A + input.B, nil
		}).
		Build()
	_ = srv.AddTool(addTool)

	multiplyTool, _ := builder.NewTool("multiply").
		Description("Multiply two numbers").
		Handler(func(_ context.Context, input MathInput) (float64, error) {
			return input.A * input.B, nil
		}).
		Build()
	_ = srv.AddTool(multiplyTool)

	greetTool, _ := builder.NewTool("greet").
		Description("Generate a greeting in different languages").
		Handler(func(_ context.Context, input GreetInput) (string, error) {
			greeting := "Hello"
			switch input.Language {
			case "es":
				greeting = "Hola"
			case "fr":
				greeting = "Bonjour"
			case "de":
				greeting = "Guten Tag"
			}
			return fmt.Sprintf("%s, %s!", greeting, input.Name), nil
		}).
		Build()
	_ = srv.AddTool(greetTool)

	// Add Resources
	configResource := builder.NewResource("config://server").
		Name("Server Configuration").
		Description("Current server configuration").
		MimeType("application/json").
		Reader(func(_ context.Context) ([]byte, error) {
			config := map[string]interface{}{
				"version":   "1.0.0",
				"debug":     true,
				"timestamp": time.Now().Format(time.RFC3339),
			}
			return json.Marshal(config)
		}).
		Build()
	_ = srv.AddResource(configResource)

	statusResource := builder.NewResource("status://health").
		Name("Server Health Status").
		Description("Current server health information").
		MimeType("application/json").
		Reader(func(_ context.Context) ([]byte, error) {
			status := map[string]interface{}{
				"status":  "healthy",
				"uptime":  time.Since(startTime).String(),
				"checked": time.Now().Format(time.RFC3339),
			}
			return json.Marshal(status)
		}).
		Build()
	_ = srv.AddResource(statusResource)

	// Add Resource Template for file reading
	fileTemplate := builder.NewResourceTemplate("file:///{path}").
		Name("File Reader").
		Description("Read files from the filesystem").
		MimeType("text/plain").
		ReaderSimple(func(_ context.Context, path string) ([]byte, error) {
			// In production, validate path to prevent directory traversal
			return os.ReadFile(path)
		}).
		Build()
	_ = srv.AddResourceTemplate(fileTemplate)

	// Add Prompts
	greetingPrompt := builder.NewPrompt("greeting").
		Description("Generate a greeting message").
		Argument("name", "Person's name", true).
		Argument("language", "Language for greeting (en, es, fr, de)", false).
		Renderer(func(_ context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			name := "there"
			if n, ok := args["name"].(string); ok {
				name = n
			}

			language := "en"
			if l, ok := args["language"].(string); ok {
				language = l
			}

			greeting := "Hello"
			switch language {
			case "es":
				greeting = "Hola"
			case "fr":
				greeting = "Bonjour"
			case "de":
				greeting = "Guten Tag"
			}

			return []*mcp.PromptMessage{
				{
					Role: "user",
					Content: []mcp.Content{
						mcp.TextContent{
							Type: "text",
							Text: fmt.Sprintf("%s, %s!", greeting, name),
						},
					},
				},
			}, nil
		}).
		Build()
	_ = srv.AddPrompt(greetingPrompt)

	calcPrompt := builder.NewPrompt("calculation").
		Description("Create a calculation request").
		Argument("operation", "The operation to perform (add, multiply)", true).
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

// createMCPHandler creates an HTTP handler for processing MCP messages
// Note: CORS is handled by streamhttp.Server, not here
func createMCPHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[MCP] %s %s from %s", r.Method, r.URL.Path, r.Header.Get("Origin"))

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
			log.Printf("[MCP] Invalid JSON: %v", err)
			http.Error(w, "invalid JSON-RPC message", http.StatusBadRequest)
			return
		}

		log.Printf("[MCP] Processing: method=%s id=%v", msg.Method, msg.ID)

		response := srv.HandleMessage(r.Context(), &msg)
		if response == nil {
			// For notifications, return 202 Accepted with no body
			log.Printf("[MCP] Notification processed: %s → 202 Accepted", msg.Method)
			w.WriteHeader(http.StatusAccepted)
			return
		}

		log.Printf("[MCP] Request processed: %s → 200 OK", msg.Method)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
	}
}

var startTime = time.Now()

func main() {
	addr := flag.String("addr", ":8080", "TCP address to listen on (e.g., :8080 or localhost:3000)")
	allowedOrigin := flag.String("origin", "*", "Allowed origin for CORS (* for all origins)")
	flag.Parse()

	srv := setupServer()
	mcpHandler := createMCPHandler(srv)

	// Create Streamable HTTP server with CORS support
	var serverOpts []streamhttp.ServerOption
	if *allowedOrigin != "" && *allowedOrigin != "*" {
		serverOpts = append(serverOpts, streamhttp.WithAllowedOrigin(*allowedOrigin))
	}

	streamServer := streamhttp.NewServer(*addr, mcpHandler, serverOpts...)

	log.Println("═══════════════════════════════════════════════════════════")
	log.Printf("Starting Streamable HTTP MCP Server")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Printf("Address:         %s", *addr)
	log.Printf("Transport:       Streamable HTTP (MCP 2025-03-26)")
	log.Printf("CORS Origin:     %s", *allowedOrigin)
	log.Printf("")
	log.Printf("For MCP Inspector, connect to: http://localhost%s", *addr)
	log.Println("───────────────────────────────────────────────────────────")
	log.Println("Available Tools:")
	log.Println("  • add       - Add two numbers")
	log.Println("  • multiply  - Multiply two numbers")
	log.Println("  • greet     - Generate multilingual greetings")
	log.Println("───────────────────────────────────────────────────────────")
	log.Println("Available Resources:")
	log.Println("  • config://server  - Server configuration")
	log.Println("  • status://health  - Server health status")
	log.Println("  • file:///{path}   - File reader template")
	log.Println("───────────────────────────────────────────────────────────")
	log.Println("Available Prompts:")
	log.Println("  • greeting     - Generate greeting messages")
	log.Println("  • calculation  - Create calculation requests")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Printf("Server ready at http://localhost%s", *addr)
	log.Println()

	if err := http.ListenAndServe(*addr, streamServer); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
