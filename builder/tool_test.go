package builder

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type TestInput struct {
	A int    `json:"a" jsonschema:"description=First number"`
	B int    `json:"b" jsonschema:"description=Second number"`
	C string `json:"c,omitempty" jsonschema:"description=Optional string"`
}

func TestToolBuilder_Build(t *testing.T) {
	builder := NewTool("test-tool").
		Description("A test tool").
		Handler(func(ctx context.Context, input TestInput) (int, error) {
			return input.A + input.B, nil
		}).
		Tags("math", "test")

	handler, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build tool: %v", err)
	}

	if handler.Name != "test-tool" {
		t.Errorf("expected name 'test-tool', got '%s'", handler.Name)
	}

	if handler.Description != "A test tool" {
		t.Errorf("expected description 'A test tool', got '%s'", handler.Description)
	}

	if len(handler.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(handler.Tags))
	}

	if handler.Schema == nil {
		t.Fatal("expected schema to be generated")
	}
}

func TestToolBuilder_Handler_Execution(t *testing.T) {
	builder := NewTool("add").
		Handler(func(ctx context.Context, input TestInput) (int, error) {
			return input.A + input.B, nil
		})

	handler, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build tool: %v", err)
	}

	ctx := context.Background()
	args := json.RawMessage(`{"a": 5, "b": 3}`)

	result, err := handler.Handler(ctx, args)
	if err != nil {
		t.Fatalf("handler execution failed: %v", err)
	}

	sum, ok := result.(int)
	if !ok {
		t.Fatalf("expected int result, got %T", result)
	}

	if sum != 8 {
		t.Errorf("expected 8, got %d", sum)
	}
}

func TestToolBuilder_Handler_WithError(t *testing.T) {
	expectedErr := errors.New("calculation failed")

	builder := NewTool("divide").
		Handler(func(ctx context.Context, input TestInput) (int, error) {
			if input.B == 0 {
				return 0, expectedErr
			}
			return input.A / input.B, nil
		})

	handler, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build tool: %v", err)
	}

	ctx := context.Background()
	args := json.RawMessage(`{"a": 10, "b": 0}`)

	_, err = handler.Handler(ctx, args)
	if err == nil {
		t.Error("expected error from handler")
	}

	if err != expectedErr {
		t.Errorf("expected specific error, got %v", err)
	}
}

func TestToolBuilder_NoHandler(t *testing.T) {
	builder := NewTool("no-handler").
		Description("Tool without handler")

	_, err := builder.Build()
	if err == nil {
		t.Error("expected error for missing handler")
	}
}

func TestToolBuilder_InvalidHandler(t *testing.T) {
	builder := NewTool("invalid").
		Handler("not a function")

	_, err := builder.Build()
	if err == nil {
		t.Error("expected error for non-function handler")
	}
}

func TestToolBuilder_NoContextArg(t *testing.T) {
	builder := NewTool("no-context").
		Handler(func(input TestInput) (int, error) {
			return 0, nil
		})

	_, err := builder.Build()
	if err == nil {
		t.Error("expected error for handler without context")
	}
}

func TestToolBuilder_ContextOnly(t *testing.T) {
	builder := NewTool("context-only").
		Handler(func(ctx context.Context) (string, error) {
			return "ok", nil
		})

	handler, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build tool: %v", err)
	}

	ctx := context.Background()
	result, err := handler.Handler(ctx, json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("handler execution failed: %v", err)
	}

	if result != "ok" {
		t.Errorf("expected 'ok', got '%v'", result)
	}
}

func TestToolBuilder_InvalidJSON(t *testing.T) {
	builder := NewTool("parse-error").
		Handler(func(ctx context.Context, input TestInput) (int, error) {
			return input.A + input.B, nil
		})

	handler, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build tool: %v", err)
	}

	ctx := context.Background()
	args := json.RawMessage(`{invalid json}`)

	_, err = handler.Handler(ctx, args)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestToolBuilder_ComplexTypes(t *testing.T) {
	type ComplexInput struct {
		Items []string          `json:"items"`
		Meta  map[string]string `json:"meta"`
	}

	builder := NewTool("complex").
		Handler(func(ctx context.Context, input ComplexInput) (int, error) {
			return len(input.Items), nil
		})

	handler, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build tool: %v", err)
	}

	ctx := context.Background()
	args := json.RawMessage(`{"items":["a","b","c"],"meta":{"key":"value"}}`)

	result, err := handler.Handler(ctx, args)
	if err != nil {
		t.Fatalf("handler execution failed: %v", err)
	}

	count, ok := result.(int)
	if !ok {
		t.Fatalf("expected int result, got %T", result)
	}

	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

func TestToolBuilder_SchemaGeneration(t *testing.T) {
	builder := NewTool("schema-test").
		Handler(func(ctx context.Context, input TestInput) (int, error) {
			return 0, nil
		})

	handler, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build tool: %v", err)
	}

	if handler.Schema == nil {
		t.Fatal("expected schema to be generated")
	}

	// Schema should be a map
	if len(handler.Schema) == 0 {
		t.Error("expected non-empty schema")
	}
}

func TestToolBuilder_Chaining(t *testing.T) {
	handler, err := NewTool("chained").
		Description("Chained builder test").
		Tags("test", "example").
		Handler(func(ctx context.Context) (string, error) {
			return "result", nil
		}).
		Build()
	if err != nil {
		t.Fatalf("failed to build tool: %v", err)
	}

	if handler.Name != "chained" {
		t.Errorf("expected name 'chained', got '%s'", handler.Name)
	}

	if handler.Description != "Chained builder test" {
		t.Errorf("unexpected description: %s", handler.Description)
	}

	if len(handler.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(handler.Tags))
	}
}

func TestToolBuilder_EmptyArgs(t *testing.T) {
	builder := NewTool("no-args").
		Handler(func(ctx context.Context, input TestInput) (int, error) {
			return input.A + input.B, nil
		})

	handler, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build tool: %v", err)
	}

	ctx := context.Background()
	args := json.RawMessage(`{}`)

	result, err := handler.Handler(ctx, args)
	if err != nil {
		t.Fatalf("handler execution failed: %v", err)
	}

	sum, ok := result.(int)
	if !ok {
		t.Fatalf("expected int result, got %T", result)
	}

	if sum != 0 {
		t.Errorf("expected 0 for empty args, got %d", sum)
	}
}
