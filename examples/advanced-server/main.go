// Package main demonstrates an advanced MCP server with middleware and lifecycle management.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmcarbo/fullmcp/builder"
	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

type MathInput struct {
	A float64 `json:"a" jsonschema:"description=First number"`
	B float64 `json:"b" jsonschema:"description=Second number"`
}

// Simple logger implementation
type SimpleLogger struct{}

func (l *SimpleLogger) Infof(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

func (l *SimpleLogger) Errorf(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

func main() {
	// Create server with middleware and lifecycle
	srv := server.New("advanced-math-server",
		server.WithVersion("2.0.0"),
		server.WithInstructions("An advanced math server with middleware and lifecycle management"),
		server.WithMiddleware(
			server.RecoveryMiddleware(),
			server.LoggingMiddleware(&SimpleLogger{}),
		),
		server.WithLifespan(func(ctx context.Context, _ *server.Server) (context.Context, func(), error) {
			log.Println("Server starting up...")

			// Initialize resources (e.g., database connections)
			cleanup := func() {
				log.Println("Server shutting down...")
			}

			return ctx, cleanup, nil
		}),
	)

	// Add tools using builder
	addTool, err := builder.NewTool("add").
		Description("Add two numbers").
		Handler(func(_ context.Context, input MathInput) (float64, error) {
			return input.A + input.B, nil
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	_ = srv.AddTool(addTool)

	multiplyTool, err := builder.NewTool("multiply").
		Description("Multiply two numbers").
		Handler(func(_ context.Context, input MathInput) (float64, error) {
			return input.A * input.B, nil
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	_ = srv.AddTool(multiplyTool)

	// Add a static resource using builder
	configResource := builder.NewResource("config://app").
		Name("Application Config").
		Description("Application configuration").
		MimeType("application/json").
		Reader(func(_ context.Context) ([]byte, error) {
			return []byte(`{"debug": true, "version": "2.0.0"}`), nil
		}).
		Build()
	_ = srv.AddResource(configResource)

	// Add a resource template for file reading
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

	// Add a prompt using builder
	greetingPrompt := builder.NewPrompt("greeting").
		Description("Generate a greeting message").
		Argument("name", "Person's name", true).
		Argument("language", "Language for greeting", false).
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

	log.Println("Starting advanced-math-server...")
	if err := srv.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
