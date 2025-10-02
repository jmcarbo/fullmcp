package server

import (
	"context"

	"github.com/jmcarbo/fullmcp/mcp"
)

// CompletionHandler provides completion suggestions for prompt or resource arguments
type CompletionHandler func(ctx context.Context, ref mcp.CompletionRef, arg mcp.CompletionArgument) ([]string, error)

// CompletionManager manages completion handlers
type CompletionManager struct {
	handlers map[string]CompletionHandler // key: "prompt:name" or "resource:uri"
}

// NewCompletionManager creates a new completion manager
func NewCompletionManager() *CompletionManager {
	return &CompletionManager{
		handlers: make(map[string]CompletionHandler),
	}
}

// RegisterPromptCompletion registers a completion handler for a prompt
func (cm *CompletionManager) RegisterPromptCompletion(name string, handler CompletionHandler) {
	key := "prompt:" + name
	cm.handlers[key] = handler
}

// RegisterResourceCompletion registers a completion handler for a resource
func (cm *CompletionManager) RegisterResourceCompletion(uri string, handler CompletionHandler) {
	key := "resource:" + uri
	cm.handlers[key] = handler
}

// GetCompletion returns completion suggestions
func (cm *CompletionManager) GetCompletion(ctx context.Context, ref mcp.CompletionRef, arg mcp.CompletionArgument) ([]string, error) {
	var key string
	if ref.Type == "ref/prompt" {
		key = "prompt:" + ref.Name
	} else if ref.Type == "ref/resource" {
		key = "resource:" + ref.Name
	} else {
		return nil, &mcp.Error{
			Code:    mcp.InvalidParams,
			Message: "invalid reference type",
		}
	}

	handler, exists := cm.handlers[key]
	if !exists {
		// No handler registered, return empty completions
		return []string{}, nil
	}

	return handler(ctx, ref, arg)
}

// WithCompletion enables completion support
func WithCompletion() Option {
	return func(s *Server) {
		s.completion = NewCompletionManager()
	}
}

// Server completion methods

// RegisterPromptCompletion registers a completion handler for a prompt
func (s *Server) RegisterPromptCompletion(name string, handler CompletionHandler) {
	if s.completion != nil {
		s.completion.RegisterPromptCompletion(name, handler)
	}
}

// RegisterResourceCompletion registers a completion handler for a resource
func (s *Server) RegisterResourceCompletion(uri string, handler CompletionHandler) {
	if s.completion != nil {
		s.completion.RegisterResourceCompletion(uri, handler)
	}
}
