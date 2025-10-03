package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/jmcarbo/fullmcp/server"
)

// ToolBuilder creates tools from functions
type ToolBuilder struct {
	name         string
	description  string
	fn           interface{}
	tags         []string
	outputSchema map[string]interface{} // 2025-06-18
	// 2025-03-26 annotations
	title           string
	readOnlyHint    *bool
	destructiveHint *bool
	idempotentHint  *bool
	openWorldHint   *bool
}

// NewTool creates a new tool builder
func NewTool(name string) *ToolBuilder {
	return &ToolBuilder{name: name}
}

// Description sets the tool description
func (tb *ToolBuilder) Description(desc string) *ToolBuilder {
	tb.description = desc
	return tb
}

// Handler sets the tool handler function
func (tb *ToolBuilder) Handler(fn interface{}) *ToolBuilder {
	tb.fn = fn
	return tb
}

// OutputSchema sets the tool output schema (2025-06-18)
func (tb *ToolBuilder) OutputSchema(schema map[string]interface{}) *ToolBuilder {
	tb.outputSchema = schema
	return tb
}

// OutputSchemaFromType generates output schema from a Go type (2025-06-18)
func (tb *ToolBuilder) OutputSchemaFromType(outputType interface{}) *ToolBuilder {
	reflector := jsonschema.Reflector{
		DoNotReference: true, // Inline all schemas instead of using $ref
	}
	jsonSchema := reflector.Reflect(outputType)
	schemaBytes, _ := json.Marshal(jsonSchema)
	var schema map[string]interface{}
	_ = json.Unmarshal(schemaBytes, &schema)
	tb.outputSchema = schema
	return tb
}

// Tags sets the tool tags
func (tb *ToolBuilder) Tags(tags ...string) *ToolBuilder {
	tb.tags = tags
	return tb
}

// Title sets a human-readable title
func (tb *ToolBuilder) Title(title string) *ToolBuilder {
	tb.title = title
	return tb
}

// ReadOnly marks this tool as read-only (doesn't modify environment)
func (tb *ToolBuilder) ReadOnly() *ToolBuilder {
	val := true
	tb.readOnlyHint = &val
	return tb
}

// Destructive marks this tool as potentially destructive
func (tb *ToolBuilder) Destructive() *ToolBuilder {
	val := true
	tb.destructiveHint = &val
	return tb
}

// Idempotent marks this tool as idempotent (repeated calls have no additional effect)
func (tb *ToolBuilder) Idempotent() *ToolBuilder {
	val := true
	tb.idempotentHint = &val
	return tb
}

// OpenWorld marks this tool as interacting with external entities
func (tb *ToolBuilder) OpenWorld() *ToolBuilder {
	val := true
	tb.openWorldHint = &val
	return tb
}

// validateFunctionSignature validates the handler function signature
func validateFunctionSignature(fnType reflect.Type) error {
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("handler must be a function")
	}

	if fnType.NumIn() < 1 {
		return fmt.Errorf("handler must accept at least context.Context")
	}

	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !fnType.In(0).Implements(ctxType) {
		return fmt.Errorf("first argument must be context.Context")
	}

	return nil
}

// generateJSONSchema generates JSON schema from input type
func generateJSONSchema(fnType reflect.Type) map[string]interface{} {
	if fnType.NumIn() > 1 {
		inputType := fnType.In(1)
		reflector := jsonschema.Reflector{}
		jsonSchema := reflector.Reflect(reflect.New(inputType).Interface())
		schemaBytes, _ := json.Marshal(jsonSchema)
		var schema map[string]interface{}
		_ = json.Unmarshal(schemaBytes, &schema)
		return schema
	}

	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

// createHandlerWrapper creates a wrapper function for the tool handler
func (tb *ToolBuilder) createHandlerWrapper(fnType reflect.Type) func(context.Context, json.RawMessage) (interface{}, error) {
	return func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		fnValue := reflect.ValueOf(tb.fn)
		callArgs := []reflect.Value{reflect.ValueOf(ctx)}

		if fnType.NumIn() > 1 {
			inputType := fnType.In(1)
			input := reflect.New(inputType).Interface()

			if len(args) > 0 {
				if err := json.Unmarshal(args, input); err != nil {
					return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
				}
			}

			callArgs = append(callArgs, reflect.ValueOf(input).Elem())
		}

		results := fnValue.Call(callArgs)

		if len(results) == 2 {
			if !results[1].IsNil() {
				return nil, results[1].Interface().(error)
			}
			return results[0].Interface(), nil
		}

		return nil, fmt.Errorf("invalid handler signature")
	}
}

// Build creates the ToolHandler
func (tb *ToolBuilder) Build() (*server.ToolHandler, error) {
	if tb.fn == nil {
		return nil, fmt.Errorf("handler function is required")
	}

	fnType := reflect.TypeOf(tb.fn)
	if err := validateFunctionSignature(fnType); err != nil {
		return nil, err
	}

	schema := generateJSONSchema(fnType)
	handler := tb.createHandlerWrapper(fnType)

	return &server.ToolHandler{
		Name:            tb.name,
		Description:     tb.description,
		Schema:          schema,
		OutputSchema:    tb.outputSchema, // 2025-06-18
		Handler:         handler,
		Tags:            tb.tags,
		Title:           tb.title,
		ReadOnlyHint:    tb.readOnlyHint,
		DestructiveHint: tb.destructiveHint,
		IdempotentHint:  tb.idempotentHint,
		OpenWorldHint:   tb.openWorldHint,
	}, nil
}
