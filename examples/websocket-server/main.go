// Package main provides a WebSocket MCP server example
package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jmcarbo/fullmcp/builder"
	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
	"github.com/jmcarbo/fullmcp/transport/websocket"
)

type CalculateArgs struct {
	A int `json:"a" jsonschema:"required,description=First number"`
	B int `json:"b" jsonschema:"required,description=Second number"`
}

func main() {
	// Create MCP server
	srv := server.New("websocket-math-server", server.WithVersion("1.0.0"))

	// Add math tools
	addTool, _ := builder.NewTool("add").
		Description("Add two numbers").
		Handler(func(_ context.Context, args CalculateArgs) (int, error) {
			result := args.A + args.B
			log.Printf("add(%d, %d) = %d", args.A, args.B, result)
			return result, nil
		}).
		Build()
	_ = srv.AddTool(addTool)

	multiplyTool, _ := builder.NewTool("multiply").
		Description("Multiply two numbers").
		Handler(func(_ context.Context, args CalculateArgs) (int, error) {
			result := args.A * args.B
			log.Printf("multiply(%d, %d) = %d", args.A, args.B, result)
			return result, nil
		}).
		Build()
	_ = srv.AddTool(multiplyTool)

	// Add a simple resource
	configResource := builder.NewResource("config://server").
		Name("Server Config").
		Description("WebSocket server configuration").
		MimeType("application/json").
		Reader(func(_ context.Context) ([]byte, error) {
			return []byte(`{"host":"localhost","port":8080,"protocol":"ws"}`), nil
		}).
		Build()
	_ = srv.AddResource(configResource)

	// Create WebSocket server
	addr := ":8080"
	log.Printf("Starting WebSocket MCP server on %s", addr)
	log.Println("Connect using: ws://localhost:8080")

	wsServer := websocket.NewServer(addr, func(ctx context.Context, msgBytes []byte) ([]byte, error) {
		// Parse incoming message
		var msg mcp.Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			return nil, err
		}

		// Handle message
		response := srv.HandleMessage(ctx, &msg)
		if response == nil {
			return nil, nil
		}

		// Serialize response
		return json.Marshal(response)
	})

	// Start server
	if err := wsServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
