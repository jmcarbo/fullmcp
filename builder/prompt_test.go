package builder

import (
	"context"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

func TestPromptBuilder_Build(t *testing.T) {
	prompt := NewPrompt("greeting").
		Description("Generate a greeting").
		Argument("name", "Person's name", true).
		Argument("language", "Language code", false).
		Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{
				{
					Role: "user",
					Content: []mcp.Content{
						mcp.TextContent{Type: "text", Text: "Hello"},
					},
				},
			}, nil
		}).
		Tags("greeting", "prompt").
		Build()

	if prompt.Name != "greeting" {
		t.Errorf("expected name 'greeting', got '%s'", prompt.Name)
	}

	if prompt.Description != "Generate a greeting" {
		t.Errorf("expected specific description, got '%s'", prompt.Description)
	}

	if len(prompt.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got %d", len(prompt.Arguments))
	}

	if prompt.Arguments[0].Name != "name" {
		t.Errorf("expected first argument 'name', got '%s'", prompt.Arguments[0].Name)
	}

	if !prompt.Arguments[0].Required {
		t.Error("expected first argument to be required")
	}

	if prompt.Arguments[1].Name != "language" {
		t.Errorf("expected second argument 'language', got '%s'", prompt.Arguments[1].Name)
	}

	if prompt.Arguments[1].Required {
		t.Error("expected second argument to be optional")
	}

	if len(prompt.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(prompt.Tags))
	}

	if prompt.Renderer == nil {
		t.Error("expected renderer to be set")
	}
}

func TestPromptBuilder_Renderer_Execution(t *testing.T) {
	prompt := NewPrompt("test").
		Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
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
		}).
		Build()

	ctx := context.Background()
	args := map[string]interface{}{
		"name": "Alice",
	}

	messages, err := prompt.Renderer(ctx, args)
	if err != nil {
		t.Fatalf("renderer execution failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	if messages[0].Role != "user" {
		t.Errorf("expected role 'user', got '%s'", messages[0].Role)
	}

	if len(messages[0].Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(messages[0].Content))
	}

	textContent, ok := messages[0].Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}

	expected := "Hello, Alice"
	if textContent.Text != expected {
		t.Errorf("expected '%s', got '%s'", expected, textContent.Text)
	}
}

func TestPromptBuilder_Arguments_Multiple(t *testing.T) {
	args := []mcp.PromptArgument{
		{Name: "arg1", Description: "First arg", Required: true},
		{Name: "arg2", Description: "Second arg", Required: false},
		{Name: "arg3", Description: "Third arg", Required: true},
	}

	prompt := NewPrompt("multi-arg").
		Arguments(args...).
		Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{}, nil
		}).
		Build()

	if len(prompt.Arguments) != 3 {
		t.Fatalf("expected 3 arguments, got %d", len(prompt.Arguments))
	}

	for i, arg := range prompt.Arguments {
		if arg.Name != args[i].Name {
			t.Errorf("argument %d: expected name '%s', got '%s'", i, args[i].Name, arg.Name)
		}

		if arg.Description != args[i].Description {
			t.Errorf("argument %d: expected description '%s', got '%s'", i, args[i].Description, arg.Description)
		}

		if arg.Required != args[i].Required {
			t.Errorf("argument %d: expected required=%v, got %v", i, args[i].Required, arg.Required)
		}
	}
}

func TestPromptBuilder_Chaining(t *testing.T) {
	prompt := NewPrompt("chained").
		Description("Chained prompt").
		Argument("param", "A parameter", true).
		Tags("test", "example").
		Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{
				{
					Role: "assistant",
					Content: []mcp.Content{
						mcp.TextContent{Type: "text", Text: "Response"},
					},
				},
			}, nil
		}).
		Build()

	if prompt.Name != "chained" {
		t.Errorf("expected name 'chained', got '%s'", prompt.Name)
	}

	if prompt.Description != "Chained prompt" {
		t.Errorf("expected description 'Chained prompt', got '%s'", prompt.Description)
	}

	if len(prompt.Arguments) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(prompt.Arguments))
	}

	if len(prompt.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(prompt.Tags))
	}
}

func TestPromptBuilder_NoArguments(t *testing.T) {
	prompt := NewPrompt("no-args").
		Description("Prompt without arguments").
		Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{
				{
					Role: "system",
					Content: []mcp.Content{
						mcp.TextContent{Type: "text", Text: "System message"},
					},
				},
			}, nil
		}).
		Build()

	if len(prompt.Arguments) != 0 {
		t.Fatalf("expected 0 arguments, got %d", len(prompt.Arguments))
	}

	ctx := context.Background()
	messages, err := prompt.Renderer(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("renderer execution failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("expected role 'system', got '%s'", messages[0].Role)
	}
}

func TestPromptBuilder_MultipleContentBlocks(t *testing.T) {
	prompt := NewPrompt("multi-content").
		Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{
				{
					Role: "user",
					Content: []mcp.Content{
						mcp.TextContent{Type: "text", Text: "First block"},
						mcp.TextContent{Type: "text", Text: "Second block"},
					},
				},
			}, nil
		}).
		Build()

	ctx := context.Background()
	messages, err := prompt.Renderer(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("renderer execution failed: %v", err)
	}

	if len(messages[0].Content) != 2 {
		t.Fatalf("expected 2 content blocks, got %d", len(messages[0].Content))
	}
}
