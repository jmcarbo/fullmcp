package server

import (
	"context"
	"sync"

	"github.com/jmcarbo/fullmcp/mcp"
)

// PromptFunc renders a prompt
type PromptFunc func(context.Context, map[string]interface{}) ([]*mcp.PromptMessage, error)

// PromptHandler wraps a prompt function
type PromptHandler struct {
	Name        string
	Description string
	Arguments   []mcp.PromptArgument
	Renderer    PromptFunc
	Tags        []string
}

// PromptManager manages prompts
type PromptManager struct {
	prompts map[string]*PromptHandler
	mu      sync.RWMutex
}

// NewPromptManager creates a new prompt manager
func NewPromptManager() *PromptManager {
	return &PromptManager{
		prompts: make(map[string]*PromptHandler),
	}
}

// Register registers a prompt
func (pm *PromptManager) Register(handler *PromptHandler) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.prompts[handler.Name] = handler
	return nil
}

// Get renders a prompt
func (pm *PromptManager) Get(ctx context.Context, name string, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
	pm.mu.RLock()
	handler, exists := pm.prompts[name]
	pm.mu.RUnlock()

	if !exists {
		return nil, &mcp.NotFoundError{Type: "prompt", Name: name}
	}

	return handler.Renderer(ctx, args)
}

// List returns all prompts
func (pm *PromptManager) List() []*mcp.Prompt {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	prompts := make([]*mcp.Prompt, 0, len(pm.prompts))
	for _, handler := range pm.prompts {
		prompts = append(prompts, &mcp.Prompt{
			Name:        handler.Name,
			Description: handler.Description,
			Arguments:   handler.Arguments,
		})
	}

	return prompts
}
