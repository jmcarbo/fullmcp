package builder

import (
	"context"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

// Type aliases for cleaner code
type (
	PromptMessage = mcp.PromptMessage
	Content       = mcp.Content
	TextContent   = mcp.TextContent
)

type SimpleTool struct {
	A int `json:"a" jsonschema:"required,description=First number"`
	B int `json:"b" jsonschema:"required,description=Second number"`
}

type ComplexTool struct {
	Name     string                 `json:"name" jsonschema:"required,description=Name field"`
	Age      int                    `json:"age" jsonschema:"required,description=Age field"`
	Email    string                 `json:"email" jsonschema:"description=Email field"`
	Tags     []string               `json:"tags" jsonschema:"description=Tags"`
	Metadata map[string]interface{} `json:"metadata" jsonschema:"description=Metadata"`
}

// BenchmarkToolBuilder measures tool building performance
func BenchmarkToolBuilder(b *testing.B) {
	handler := func(_ context.Context, args SimpleTool) (int, error) {
		return args.A + args.B, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewTool("add").
			Description("Add two numbers").
			Handler(handler).
			Build()
	}
}

// BenchmarkToolBuilderComplex measures complex tool building
func BenchmarkToolBuilderComplex(b *testing.B) {
	handler := func(_ context.Context, args ComplexTool) (string, error) {
		return args.Name, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewTool("complex").
			Description("Complex tool").
			Handler(handler).
			Build()
	}
}

// BenchmarkResourceBuilder measures resource building performance
func BenchmarkResourceBuilder(b *testing.B) {
	reader := func(_ context.Context) ([]byte, error) {
		return []byte(`{"key":"value"}`), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewResource("config://test").
			Name("Test Config").
			Description("Test configuration").
			MimeType("application/json").
			Reader(reader).
			Build()
	}
}

// BenchmarkResourceTemplateBuilder measures resource template building
func BenchmarkResourceTemplateBuilder(b *testing.B) {
	reader := func(_ context.Context, path string) ([]byte, error) {
		return []byte(path), nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewResourceTemplate("file:///{path}").
			Name("File Reader").
			Description("Read files").
			MimeType("text/plain").
			ReaderSimple(reader).
			Build()
	}
}

// BenchmarkPromptBuilder measures prompt building performance
func BenchmarkPromptBuilder(b *testing.B) {
	renderer := func(_ context.Context, args map[string]interface{}) ([]*PromptMessage, error) {
		name := args["name"].(string)
		return []*PromptMessage{{
			Role: "user",
			Content: []Content{
				&TextContent{Type: "text", Text: "Hello, " + name},
			},
		}}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewPrompt("greeting").
			Description("Greeting prompt").
			Argument("name", "User name", true).
			Renderer(renderer).
			Build()
	}
}

// BenchmarkToolBuildComplete measures complete tool building with schema generation
func BenchmarkToolBuildComplete(b *testing.B) {
	handler := func(_ context.Context, args SimpleTool) (int, error) {
		return args.A + args.B, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewTool("add").
			Description("Add two numbers").
			Handler(handler).
			Build()
	}
}

// BenchmarkComplexToolBuildComplete measures complex tool building
func BenchmarkComplexToolBuildComplete(b *testing.B) {
	handler := func(_ context.Context, args ComplexTool) (string, error) {
		return args.Name, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewTool("complex").
			Description("Complex tool").
			Handler(handler).
			Build()
	}
}
