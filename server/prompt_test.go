package server

import (
	"context"
	"errors"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

func TestPromptManager_Register(t *testing.T) {
	pm := NewPromptManager()

	handler := &PromptHandler{
		Name:        "test-prompt",
		Description: "A test prompt",
		Renderer: func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{}, nil
		},
	}

	err := pm.Register(handler)
	if err != nil {
		t.Fatalf("failed to register prompt: %v", err)
	}

	// Registering same name again should overwrite
	err = pm.Register(handler)
	if err != nil {
		t.Fatalf("failed to re-register prompt: %v", err)
	}
}

func TestPromptManager_Get(t *testing.T) {
	pm := NewPromptManager()

	handler := &PromptHandler{
		Name: "greeting",
		Renderer: func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			name := args["name"].(string)
			return []*mcp.PromptMessage{
				{
					Role: "user",
					Content: []mcp.Content{
						mcp.TextContent{
							Type: "text",
							Text: "Hello, " + name,
						},
					},
				},
			}, nil
		},
	}

	pm.Register(handler)

	ctx := context.Background()
	args := map[string]interface{}{
		"name": "Alice",
	}

	messages, err := pm.Get(ctx, "greeting", args)
	if err != nil {
		t.Fatalf("failed to get prompt: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	if messages[0].Role != "user" {
		t.Errorf("expected role 'user', got '%s'", messages[0].Role)
	}
}

func TestPromptManager_Get_NotFound(t *testing.T) {
	pm := NewPromptManager()
	ctx := context.Background()

	_, err := pm.Get(ctx, "nonexistent", nil)
	if err == nil {
		t.Error("expected error for nonexistent prompt")
	}

	notFoundErr, ok := err.(*mcp.NotFoundError)
	if !ok {
		t.Errorf("expected NotFoundError, got %T", err)
	}

	if notFoundErr.Type != "prompt" {
		t.Errorf("expected type 'prompt', got '%s'", notFoundErr.Type)
	}
}

func TestPromptManager_Get_RendererError(t *testing.T) {
	pm := NewPromptManager()

	expectedErr := errors.New("render failed")
	handler := &PromptHandler{
		Name: "error-prompt",
		Renderer: func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return nil, expectedErr
		},
	}

	pm.Register(handler)

	ctx := context.Background()
	_, err := pm.Get(ctx, "error-prompt", nil)
	if err == nil {
		t.Error("expected error from renderer")
	}

	if err != expectedErr {
		t.Errorf("expected specific error, got %v", err)
	}
}

func TestPromptManager_List(t *testing.T) {
	pm := NewPromptManager()

	handlers := []*PromptHandler{
		{
			Name:        "prompt1",
			Description: "First prompt",
			Arguments: []mcp.PromptArgument{
				{Name: "arg1", Required: true},
			},
		},
		{
			Name:        "prompt2",
			Description: "Second prompt",
			Arguments: []mcp.PromptArgument{
				{Name: "arg2", Required: false},
			},
		},
	}

	for _, handler := range handlers {
		handler.Renderer = func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{}, nil
		}
		pm.Register(handler)
	}

	prompts := pm.List()
	if len(prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(prompts))
	}

	names := make(map[string]bool)
	for _, prompt := range prompts {
		names[prompt.Name] = true
	}

	if !names["prompt1"] || !names["prompt2"] {
		t.Error("expected both prompts in list")
	}
}

func TestPromptManager_WithArguments(t *testing.T) {
	pm := NewPromptManager()

	handler := &PromptHandler{
		Name: "parameterized",
		Arguments: []mcp.PromptArgument{
			{Name: "name", Description: "User name", Required: true},
			{Name: "age", Description: "User age", Required: false},
		},
		Renderer: func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{}, nil
		},
	}

	pm.Register(handler)

	prompts := pm.List()
	if len(prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(prompts))
	}

	if len(prompts[0].Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(prompts[0].Arguments))
	}

	if !prompts[0].Arguments[0].Required {
		t.Error("expected first argument to be required")
	}

	if prompts[0].Arguments[1].Required {
		t.Error("expected second argument to be optional")
	}
}

func TestPromptManager_ConcurrentAccess(t *testing.T) {
	pm := NewPromptManager()

	handler := &PromptHandler{
		Name: "concurrent",
		Renderer: func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{}, nil
		},
	}

	pm.Register(handler)

	ctx := context.Background()
	done := make(chan bool)

	// Concurrent gets
	for i := 0; i < 10; i++ {
		go func() {
			_, err := pm.Get(ctx, "concurrent", nil)
			if err != nil {
				t.Errorf("concurrent get failed: %v", err)
			}
			done <- true
		}()
	}

	// Concurrent list operations
	for i := 0; i < 10; i++ {
		go func() {
			_ = pm.List()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}
