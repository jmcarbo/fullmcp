// Package main demonstrates a basic MCP server with math operations.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jmcarbo/fullmcp/builder"
	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

type MathInput struct {
	A float64 `json:"a" jsonschema:"description=First number"`
	B float64 `json:"b" jsonschema:"description=Second number"`
}

func main() {
	srv := server.New("math-server",
		server.WithVersion("1.0.0"),
		server.WithInstructions("A simple math server with basic operations"),
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

	// Add a resource
	_ = srv.AddResource(&server.ResourceHandler{
		URI:         "config://app",
		Name:        "Application Config",
		Description: "Application configuration",
		MimeType:    "application/json",
		Reader: func(_ context.Context) ([]byte, error) {
			return []byte(`{"debug": true, "version": "1.0.0"}`), nil
		},
	})

	// Add a prompt
	_ = srv.AddPrompt(&server.PromptHandler{
		Name:        "greeting",
		Description: "Generate a greeting message",
		Arguments: []mcp.PromptArgument{
			{Name: "name", Description: "Person's name", Required: true},
		},
		Renderer: func(_ context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			name := "there"
			if n, ok := args["name"].(string); ok {
				name = n
			}

			return []*mcp.PromptMessage{
				{
					Role: "user",
					Content: []mcp.Content{
						mcp.TextContent{
							Type: "text",
							Text: fmt.Sprintf("Hello, %s!", name),
						},
					},
				},
			}, nil
		},
	})

	log.Println("Starting math-server...")
	if err := srv.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
