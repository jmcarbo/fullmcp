package builder

import (
	"context"
	"testing"
)

// Test OutputSchema method (2025-06-18)
func TestToolBuilder_OutputSchema(t *testing.T) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"result": map[string]interface{}{"type": "string"},
		},
	}

	tool, err := NewTool("test").
		Handler(func(ctx context.Context) (string, error) {
			return "ok", nil
		}).
		OutputSchema(schema).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if tool.OutputSchema == nil {
		t.Fatal("expected OutputSchema to be set")
	}

	if tool.OutputSchema["type"] != "object" {
		t.Errorf("expected type 'object', got %v", tool.OutputSchema["type"])
	}
}

// Test OutputSchemaFromType method (2025-06-18)
func TestToolBuilder_OutputSchemaFromType(t *testing.T) {
	type Result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Count   int    `json:"count"`
	}

	tool, err := NewTool("test").
		Handler(func(ctx context.Context) (Result, error) {
			return Result{Success: true, Message: "ok", Count: 5}, nil
		}).
		OutputSchemaFromType(Result{}).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if tool.OutputSchema == nil {
		t.Fatal("expected OutputSchema to be generated")
	}

	// Debug: print the schema
	t.Logf("OutputSchema: %+v", tool.OutputSchema)

	// Check that properties were generated
	props, ok := tool.OutputSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected properties in OutputSchema, got type %T with keys %v", tool.OutputSchema["properties"], tool.OutputSchema)
	}

	if _, exists := props["success"]; !exists {
		t.Error("expected 'success' property in schema")
	}

	if _, exists := props["message"]; !exists {
		t.Error("expected 'message' property in schema")
	}

	if _, exists := props["count"]; !exists {
		t.Error("expected 'count' property in schema")
	}
}

// Test OutputSchemaFromType with nested struct
func TestToolBuilder_OutputSchemaFromType_Nested(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	tool, err := NewTool("test").
		Handler(func(ctx context.Context) (Person, error) {
			return Person{}, nil
		}).
		OutputSchemaFromType(Person{}).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if tool.OutputSchema == nil {
		t.Fatal("expected OutputSchema to be generated")
	}

	props, ok := tool.OutputSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected properties in OutputSchema")
	}

	// Check nested structure
	if _, exists := props["address"]; !exists {
		t.Error("expected 'address' property in schema")
	}
}

// Test that OutputSchema is included in tool listing
func TestToolBuilder_OutputSchemaInListing(t *testing.T) {
	schema := map[string]interface{}{
		"type":    "number",
		"minimum": 0,
		"maximum": 100,
	}

	tool, err := NewTool("test").
		Description("Test tool").
		Handler(func(ctx context.Context) (int, error) {
			return 42, nil
		}).
		OutputSchema(schema).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify OutputSchema is accessible
	if tool.OutputSchema == nil {
		t.Fatal("expected OutputSchema")
	}

	if tool.OutputSchema["type"] != "number" {
		t.Errorf("expected type 'number', got %v", tool.OutputSchema["type"])
	}

	if tool.OutputSchema["minimum"] != 0 {
		t.Errorf("expected minimum 0, got %v", tool.OutputSchema["minimum"])
	}

	if tool.OutputSchema["maximum"] != 100 {
		t.Errorf("expected maximum 100, got %v", tool.OutputSchema["maximum"])
	}
}

// Test OutputSchema with array type
func TestToolBuilder_OutputSchemaFromType_Array(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	tool, err := NewTool("test").
		Handler(func(ctx context.Context) ([]Item, error) {
			return []Item{}, nil
		}).
		OutputSchemaFromType([]Item{}).
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if tool.OutputSchema == nil {
		t.Fatal("expected OutputSchema to be generated")
	}

	// Schema for array should have type "array"
	schemaType, ok := tool.OutputSchema["type"].(string)
	if !ok || schemaType != "array" {
		t.Errorf("expected type 'array', got %v", tool.OutputSchema["type"])
	}
}

// Test OutputSchema with primitive type
func TestToolBuilder_OutputSchemaFromType_Primitive(t *testing.T) {
	tests := []struct {
		name         string
		value        interface{}
		expectedType string
	}{
		{"string", "", "string"},
		{"int", 0, "integer"},
		{"bool", false, "boolean"},
		{"float64", 0.0, "number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := NewTool("test").
				Handler(func(ctx context.Context) (interface{}, error) {
					return tt.value, nil
				}).
				OutputSchemaFromType(tt.value).
				Build()
			if err != nil {
				t.Fatalf("Build failed: %v", err)
			}

			if tool.OutputSchema == nil {
				t.Fatal("expected OutputSchema to be generated")
			}

			schemaType, ok := tool.OutputSchema["type"].(string)
			if !ok || schemaType != tt.expectedType {
				t.Errorf("expected type '%s', got %v", tt.expectedType, tool.OutputSchema["type"])
			}
		})
	}
}

// Test combining Title and OutputSchema (2025-06-18 features)
func TestToolBuilder_TitleAndOutputSchema(t *testing.T) {
	type Result struct {
		Score int `json:"score"`
	}

	tool, err := NewTool("analyze").
		Title("Code Analyzer").
		Description("Analyzes code quality").
		Handler(func(ctx context.Context, code string) (Result, error) {
			return Result{Score: 95}, nil
		}).
		OutputSchemaFromType(Result{}).
		ReadOnly().
		Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if tool.Title != "Code Analyzer" {
		t.Errorf("expected title 'Code Analyzer', got '%s'", tool.Title)
	}

	if tool.OutputSchema == nil {
		t.Fatal("expected OutputSchema")
	}

	if tool.ReadOnlyHint == nil || !*tool.ReadOnlyHint {
		t.Error("expected ReadOnlyHint to be true")
	}
}
