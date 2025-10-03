// Package main demonstrates MCP tools that return typed content (images, audio, resources).
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/jmcarbo/fullmcp/builder"
	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

type ImageInput struct {
	Width  int    `json:"width" jsonschema:"description=Image width in pixels"`
	Height int    `json:"height" jsonschema:"description=Image height in pixels"`
	Color  string `json:"color,omitempty" jsonschema:"description=Background color (hex)"`
}

type DataInput struct {
	Format string `json:"format" jsonschema:"description=Output format (json or text)"`
	Data   string `json:"data" jsonschema:"description=Data to process"`
}

func main() {
	srv := server.New("typed-content-server",
		server.WithVersion("1.0.0"),
		server.WithInstructions("Demonstrates tools returning typed content"),
	)

	// Tool that returns TextContent (explicit)
	textTool, err := builder.NewTool("echo").
		Description("Echo text back as TextContent").
		Handler(func(_ context.Context, input struct {
			Message string `json:"message"`
		}) (mcp.TextContent, error) {
			return mcp.TextContent{
				Type: "text",
				Text: input.Message,
			}, nil
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	_ = srv.AddTool(textTool)

	// Tool that returns ImageContent
	imageTool, err := builder.NewTool("generate_placeholder_image").
		Description("Generate a placeholder image (returns ImageContent with base64 data)").
		Handler(func(_ context.Context, input ImageInput) (mcp.ImageContent, error) {
			// In a real implementation, this would generate an actual image
			// For demo, we create a small 1x1 transparent PNG
			pngData := []byte{
				0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
				0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
				0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1
				0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
				0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
				0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
				0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
				0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
				0x42, 0x60, 0x82,
			}

			return mcp.ImageContent{
				Type:     "image",
				Data:     base64.StdEncoding.EncodeToString(pngData),
				MimeType: "image/png",
			}, nil
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	_ = srv.AddTool(imageTool)

	// Tool that returns AudioContent
	audioTool, err := builder.NewTool("generate_tone").
		Description("Generate an audio tone (returns AudioContent with base64 data)").
		Handler(func(_ context.Context, input struct {
			Duration int `json:"duration" jsonschema:"description=Duration in seconds"`
		}) (mcp.AudioContent, error) {
			// In a real implementation, this would generate actual audio
			// For demo, return placeholder base64 data
			audioData := []byte("mock audio data")

			return mcp.AudioContent{
				Type:     "audio",
				Data:     base64.StdEncoding.EncodeToString(audioData),
				MimeType: "audio/wav",
			}, nil
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	_ = srv.AddTool(audioTool)

	// Tool that returns multiple Content items
	multiContentTool, err := builder.NewTool("analyze_data").
		Description("Analyze data and return multiple content blocks").
		Handler(func(_ context.Context, input DataInput) ([]mcp.Content, error) {
			// Return multiple content blocks
			return []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Analysis of data in %s format:", input.Format),
				},
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Data length: %d characters", len(input.Data)),
				},
				mcp.ResourceContent{
					Type:     "resource",
					URI:      "data://processed",
					MimeType: "application/json",
					Text:     fmt.Sprintf(`{"format":"%s","length":%d}`, input.Format, len(input.Data)),
				},
			}, nil
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	_ = srv.AddTool(multiContentTool)

	// Tool that returns string (backward compatibility)
	simpleTool, err := builder.NewTool("simple_text").
		Description("Simple tool that returns a string (auto-converted to TextContent)").
		Handler(func(_ context.Context, input struct {
			Text string `json:"text"`
		}) (string, error) {
			return "You said: " + input.Text, nil
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	_ = srv.AddTool(simpleTool)

	// Tool that returns struct (auto-converted to JSON TextContent)
	structTool, err := builder.NewTool("get_info").
		Description("Returns structured data (auto-converted to JSON)").
		Handler(func(_ context.Context, input struct {
			Name string `json:"name"`
		}) (struct {
			Name      string `json:"name"`
			Timestamp string `json:"timestamp"`
			Status    string `json:"status"`
		}, error) {
			return struct {
				Name      string `json:"name"`
				Timestamp string `json:"timestamp"`
				Status    string `json:"status"`
			}{
				Name:      input.Name,
				Timestamp: "2025-01-01T00:00:00Z",
				Status:    "active",
			}, nil
		}).
		Build()
	if err != nil {
		log.Fatal(err)
	}
	_ = srv.AddTool(structTool)

	log.Println("Starting typed-content-server...")
	log.Println("This server demonstrates:")
	log.Println("  - Tools returning TextContent, ImageContent, AudioContent")
	log.Println("  - Tools returning multiple Content items")
	log.Println("  - Tools returning simple types (string, struct) - auto-converted")
	if err := srv.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
