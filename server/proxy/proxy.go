// Package proxy provides a proxy server that forwards MCP requests to a backend server.
package proxy

import (
	"context"
	"encoding/json"

	"github.com/jmcarbo/fullmcp/client"
	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

// Server is a proxy that forwards requests to a backend MCP server
type Server struct {
	*server.Server
	backend *client.Client
}

// Option configures the proxy server
type Option func(*Server)

// New creates a new proxy server that forwards all requests to the backend
func New(name string, backend *client.Client, opts ...Option) (*Server, error) {
	srv := server.New(name)

	ps := &Server{
		Server:  srv,
		backend: backend,
	}

	// Apply options
	for _, opt := range opts {
		opt(ps)
	}

	// Register proxy handlers by fetching from backend and creating local handlers
	if err := ps.syncFromBackend(context.Background()); err != nil {
		return nil, err
	}

	return ps, nil
}

// WithServerOptions sets options for the underlying server
func WithServerOptions(serverOpts ...server.Option) Option {
	return func(ps *Server) {
		for _, opt := range serverOpts {
			opt(ps.Server)
		}
	}
}

// syncFromBackend fetches all tools, resources, and prompts from the backend
// and creates proxy handlers for them
func (ps *Server) syncFromBackend(ctx context.Context) error {
	// Sync tools
	tools, err := ps.backend.ListTools(ctx)
	if err != nil {
		return err
	}

	for _, tool := range tools {
		toolName := tool.Name
		toolHandler := &server.ToolHandler{
			Name:        tool.Name,
			Description: tool.Description,
			Schema:      tool.InputSchema,
			Handler: func(ctx context.Context, args json.RawMessage) (interface{}, error) {
				return ps.backend.CallTool(ctx, toolName, args)
			},
		}
		if err := ps.Server.AddTool(toolHandler); err != nil {
			return err
		}
	}

	// Sync resources
	resources, err := ps.backend.ListResources(ctx)
	if err != nil {
		return err
	}

	for _, resource := range resources {
		resourceURI := resource.URI
		resourceHandler := &server.ResourceHandler{
			URI:         resource.URI,
			Name:        resource.Name,
			Description: resource.Description,
			MimeType:    resource.MimeType,
			Reader: func(ctx context.Context) ([]byte, error) {
				contents, err := ps.backend.ReadResource(ctx, resourceURI)
				if err != nil {
					return nil, err
				}

				// Return first content item as bytes
				if len(contents) > 0 {
					return contentToBytes(contents[0]), nil
				}

				return nil, nil
			},
		}
		if err := ps.Server.AddResource(resourceHandler); err != nil {
			return err
		}
	}

	// Sync prompts
	prompts, err := ps.backend.ListPrompts(ctx)
	if err != nil {
		return err
	}

	for _, prompt := range prompts {
		promptName := prompt.Name
		promptHandler := &server.PromptHandler{
			Name:        prompt.Name,
			Description: prompt.Description,
			Arguments:   prompt.Arguments,
			Renderer: func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
				return ps.backend.GetPrompt(ctx, promptName, args)
			},
		}
		if err := ps.Server.AddPrompt(promptHandler); err != nil {
			return err
		}
	}

	return nil
}

// contentToBytes converts a content item to bytes
func contentToBytes(content interface{}) []byte {
	// Attempt to marshal content as JSON if it's a struct
	data, _ := json.Marshal(content)
	return data
}
