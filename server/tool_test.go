package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

func TestToolManager_Register(t *testing.T) {
	tm := NewToolManager()

	handler := &ToolHandler{
		Name:        "test-tool",
		Description: "A test tool",
		Schema:      map[string]interface{}{"type": "object"},
		Handler: func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			return "result", nil
		},
	}

	err := tm.Register(handler)
	if err != nil {
		t.Fatalf("failed to register tool: %v", err)
	}

	// Try to register same tool again
	err = tm.Register(handler)
	if err == nil {
		t.Error("expected error when registering duplicate tool")
	}
}

func TestToolManager_Call(t *testing.T) {
	tm := NewToolManager()

	handler := &ToolHandler{
		Name: "echo",
		Handler: func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			var input map[string]string
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, err
			}
			return input["message"], nil
		},
	}

	tm.Register(handler)

	ctx := context.Background()
	args := json.RawMessage(`{"message":"hello"}`)

	result, err := tm.Call(ctx, "echo", args)
	if err != nil {
		t.Fatalf("failed to call tool: %v", err)
	}

	if result != "hello" {
		t.Errorf("expected 'hello', got '%v'", result)
	}
}

func TestToolManager_Call_NotFound(t *testing.T) {
	tm := NewToolManager()
	ctx := context.Background()

	_, err := tm.Call(ctx, "nonexistent", json.RawMessage(`{}`))
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}

	notFoundErr, ok := err.(*mcp.NotFoundError)
	if !ok {
		t.Errorf("expected NotFoundError, got %T", err)
	}

	if notFoundErr.Type != "tool" {
		t.Errorf("expected type 'tool', got '%s'", notFoundErr.Type)
	}
}

func TestToolManager_List(t *testing.T) {
	tm := NewToolManager()

	handlers := []*ToolHandler{
		{
			Name:        "tool1",
			Description: "First tool",
			Schema:      map[string]interface{}{"type": "object"},
		},
		{
			Name:        "tool2",
			Description: "Second tool",
			Schema:      map[string]interface{}{"type": "object"},
		},
	}

	for _, handler := range handlers {
		handler.Handler = func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			return nil, nil
		}
		tm.Register(handler)
	}

	tools, _ := tm.List(context.Background())
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	if !toolNames["tool1"] || !toolNames["tool2"] {
		t.Error("expected both tool1 and tool2 in list")
	}
}

func TestToolManager_ConcurrentAccess(t *testing.T) {
	tm := NewToolManager()

	handler := &ToolHandler{
		Name: "concurrent",
		Handler: func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			return "ok", nil
		},
	}

	tm.Register(handler)

	ctx := context.Background()
	done := make(chan bool)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_, err := tm.Call(ctx, "concurrent", json.RawMessage(`{}`))
			if err != nil {
				t.Errorf("concurrent call failed: %v", err)
			}
			done <- true
		}()
	}

	// Concurrent list operations
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = tm.List(context.Background())
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestToolHandler_Tags(t *testing.T) {
	handler := &ToolHandler{
		Name: "tagged-tool",
		Tags: []string{"math", "utility"},
		Handler: func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			return nil, nil
		},
	}

	if len(handler.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(handler.Tags))
	}

	if handler.Tags[0] != "math" || handler.Tags[1] != "utility" {
		t.Errorf("unexpected tags: %v", handler.Tags)
	}
}

func TestToolManager_ValidationSuccess(t *testing.T) {
	tm := NewToolManager()

	handler := &ToolHandler{
		Name:        "add",
		Description: "Add two numbers",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{"type": "number"},
				"b": map[string]interface{}{"type": "number"},
			},
			"required":             []interface{}{"a", "b"},
			"additionalProperties": false,
		},
		Handler: func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			var input struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, err
			}
			return input.A + input.B, nil
		},
	}

	if err := tm.Register(handler); err != nil {
		t.Fatalf("failed to register tool: %v", err)
	}

	ctx := context.Background()
	args := json.RawMessage(`{"a": 5, "b": 3}`)

	result, err := tm.Call(ctx, "add", args)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if result != 8.0 {
		t.Errorf("expected 8, got %v", result)
	}
}

func TestToolManager_ValidationMissingRequiredField(t *testing.T) {
	tm := NewToolManager()

	handler := &ToolHandler{
		Name: "add",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{"type": "number"},
				"b": map[string]interface{}{"type": "number"},
			},
			"required": []interface{}{"a", "b"},
		},
		Handler: func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			return nil, nil
		},
	}

	tm.Register(handler)

	ctx := context.Background()
	args := json.RawMessage(`{"a": 5}`) // Missing 'b'

	_, err := tm.Call(ctx, "add", args)
	if err == nil {
		t.Fatal("expected validation error for missing required field")
	}

	validationErr, ok := err.(*mcp.ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	if validationErr.Message == "" {
		t.Error("validation error message is empty")
	}

	t.Logf("Validation error message: %s", validationErr.Message)
}

func TestToolManager_ValidationAdditionalProperties(t *testing.T) {
	tm := NewToolManager()

	handler := &ToolHandler{
		Name: "add",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{"type": "number"},
				"b": map[string]interface{}{"type": "number"},
			},
			"required":             []interface{}{"a", "b"},
			"additionalProperties": false,
		},
		Handler: func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			return nil, nil
		},
	}

	tm.Register(handler)

	ctx := context.Background()
	args := json.RawMessage(`{"a": 5, "b": 3, "invalid": true}`)

	_, err := tm.Call(ctx, "add", args)
	if err == nil {
		t.Fatal("expected validation error for additional property")
	}

	validationErr, ok := err.(*mcp.ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	t.Logf("Validation error message: %s", validationErr.Message)
}

func TestToolManager_ValidationWrongType(t *testing.T) {
	tm := NewToolManager()

	handler := &ToolHandler{
		Name: "add",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{"type": "number"},
				"b": map[string]interface{}{"type": "number"},
			},
			"required": []interface{}{"a", "b"},
		},
		Handler: func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			return nil, nil
		},
	}

	tm.Register(handler)

	ctx := context.Background()
	args := json.RawMessage(`{"a": "not a number", "b": 3}`)

	_, err := tm.Call(ctx, "add", args)
	if err == nil {
		t.Fatal("expected validation error for wrong type")
	}

	validationErr, ok := err.(*mcp.ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	t.Logf("Validation error message: %s", validationErr.Message)
}

func TestToolManager_NoSchemaNoValidation(t *testing.T) {
	tm := NewToolManager()

	handler := &ToolHandler{
		Name:   "flexible",
		Schema: nil, // No schema = no validation
		Handler: func(ctx context.Context, args json.RawMessage) (interface{}, error) {
			return "ok", nil
		},
	}

	tm.Register(handler)

	ctx := context.Background()
	args := json.RawMessage(`{"anything": "goes"}`)

	result, err := tm.Call(ctx, "flexible", args)
	if err != nil {
		t.Fatalf("unexpected error with no schema: %v", err)
	}

	if result != "ok" {
		t.Errorf("expected 'ok', got %v", result)
	}
}
