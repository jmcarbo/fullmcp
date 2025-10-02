package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/xeipuuv/gojsonschema"
)

// ToolFunc is a function that can be registered as a tool
type ToolFunc func(context.Context, json.RawMessage) (interface{}, error)

// ToolHandler wraps a tool function with metadata
type ToolHandler struct {
	Name         string
	Description  string
	Schema       map[string]interface{}
	OutputSchema map[string]interface{} // 2025-06-18
	Handler      ToolFunc
	Tags         []string
	// 2025-03-26 annotations
	Title           string
	ReadOnlyHint    *bool
	DestructiveHint *bool
	IdempotentHint  *bool
	OpenWorldHint   *bool
}

// ToolManager manages tool registration and execution
type ToolManager struct {
	tools map[string]*ToolHandler
	mu    sync.RWMutex
}

// NewToolManager creates a new tool manager
func NewToolManager() *ToolManager {
	return &ToolManager{
		tools: make(map[string]*ToolHandler),
	}
}

// Register registers a tool
func (tm *ToolManager) Register(handler *ToolHandler) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tools[handler.Name]; exists {
		return fmt.Errorf("tool already registered: %s", handler.Name)
	}

	tm.tools[handler.Name] = handler
	return nil
}

// Call executes a tool
func (tm *ToolManager) Call(ctx context.Context, name string, args json.RawMessage) (interface{}, error) {
	tm.mu.RLock()
	handler, exists := tm.tools[name]
	tm.mu.RUnlock()

	if !exists {
		return nil, &mcp.NotFoundError{Type: "tool", Name: name}
	}

	// Validate arguments against JSON schema if schema is defined
	if handler.Schema != nil {
		if err := tm.validateArguments(args, handler.Schema); err != nil {
			return nil, err
		}
	}

	return handler.Handler(ctx, args)
}

// validateArguments validates JSON arguments against a JSON schema
func (tm *ToolManager) validateArguments(args json.RawMessage, schema map[string]interface{}) error {
	// Convert schema to JSON
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return &mcp.ValidationError{Message: fmt.Sprintf("invalid schema: %v", err)}
	}

	// Create schema loader
	schemaLoader := gojsonschema.NewBytesLoader(schemaJSON)

	// Create document loader from arguments
	documentLoader := gojsonschema.NewBytesLoader(args)

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return &mcp.ValidationError{Message: fmt.Sprintf("validation error: %v", err)}
	}

	if !result.Valid() {
		// Build error message from validation errors
		errMsg := "invalid arguments: "
		for i, desc := range result.Errors() {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += desc.String()
		}
		return &mcp.ValidationError{Message: errMsg}
	}

	return nil
}

// List returns all registered tools
func (tm *ToolManager) List(_ context.Context) ([]*mcp.Tool, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tools := make([]*mcp.Tool, 0, len(tm.tools))
	for _, handler := range tm.tools {
		tools = append(tools, &mcp.Tool{
			Name:            handler.Name,
			Description:     handler.Description,
			InputSchema:     handler.Schema,
			OutputSchema:    handler.OutputSchema, // 2025-06-18
			Title:           handler.Title,
			ReadOnlyHint:    handler.ReadOnlyHint,
			DestructiveHint: handler.DestructiveHint,
			IdempotentHint:  handler.IdempotentHint,
			OpenWorldHint:   handler.OpenWorldHint,
		})
	}

	return tools, nil
}
