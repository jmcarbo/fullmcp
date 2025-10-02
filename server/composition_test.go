package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

func TestCompositeServer_Mount(t *testing.T) {
	cs := NewCompositeServer("main")
	sub := New("sub")

	err := cs.Mount("api", sub)
	if err != nil {
		t.Fatalf("Mount failed: %v", err)
	}

	mounted := cs.GetMountedServers()
	if len(mounted) != 1 {
		t.Errorf("Expected 1 mounted server, got %d", len(mounted))
	}

	if mounted["api"] != sub {
		t.Error("Mounted server not found at expected prefix")
	}
}

func TestCompositeServer_Mount_EmptyPrefix(t *testing.T) {
	cs := NewCompositeServer("main")
	sub := New("sub")

	err := cs.Mount("", sub)
	if err == nil {
		t.Error("Expected error for empty prefix, got nil")
	}
}

func TestCompositeServer_Mount_Duplicate(t *testing.T) {
	cs := NewCompositeServer("main")
	sub1 := New("sub1")
	sub2 := New("sub2")

	_ = cs.Mount("api", sub1)
	err := cs.Mount("api", sub2)
	if err == nil {
		t.Error("Expected error for duplicate mount, got nil")
	}
}

func TestCompositeServer_Unmount(t *testing.T) {
	cs := NewCompositeServer("main")
	sub := New("sub")

	_ = cs.Mount("api", sub)
	err := cs.Unmount("api")
	if err != nil {
		t.Fatalf("Unmount failed: %v", err)
	}

	mounted := cs.GetMountedServers()
	if len(mounted) != 0 {
		t.Errorf("Expected 0 mounted servers, got %d", len(mounted))
	}
}

func TestCompositeServer_Unmount_NotFound(t *testing.T) {
	cs := NewCompositeServer("main")

	err := cs.Unmount("notfound")
	if err == nil {
		t.Error("Expected error for unmounting non-existent server, got nil")
	}
}

func TestCompositeServer_ListTools(t *testing.T) {
	cs := NewCompositeServer("main")

	// Add tool to main server
	mainTool := &ToolHandler{
		Name:        "main-tool",
		Description: "Main tool",
		Handler: func(ctx context.Context, arguments json.RawMessage) (interface{}, error) {
			return "main", nil
		},
	}
	_ = cs.AddTool(mainTool)

	// Create and mount sub server with tool
	sub := New("sub")
	subTool := &ToolHandler{
		Name:        "sub-tool",
		Description: "Sub tool",
		Handler: func(ctx context.Context, arguments json.RawMessage) (interface{}, error) {
			return "sub", nil
		},
	}
	_ = sub.AddTool(subTool)
	_ = cs.Mount("api", sub)

	// List all tools
	tools := cs.ListTools(context.Background())

	if len(tools) != 2 {
		t.Fatalf("Expected 2 tools, got %d", len(tools))
	}

	// Verify tool names
	foundMain := false
	foundSub := false
	for _, tool := range tools {
		if tool.Name == "main-tool" {
			foundMain = true
		}
		if tool.Name == "api/sub-tool" {
			foundSub = true
		}
	}

	if !foundMain {
		t.Error("Main tool not found")
	}
	if !foundSub {
		t.Error("Prefixed sub tool not found")
	}
}

func TestCompositeServer_CallTool(t *testing.T) {
	cs := NewCompositeServer("main")

	// Add tool to main server
	mainTool := &ToolHandler{
		Name:        "main-tool",
		Description: "Main tool",
		Handler: func(ctx context.Context, arguments json.RawMessage) (interface{}, error) {
			return "main-result", nil
		},
	}
	_ = cs.AddTool(mainTool)

	// Create and mount sub server with tool
	sub := New("sub")
	subTool := &ToolHandler{
		Name:        "sub-tool",
		Description: "Sub tool",
		Handler: func(ctx context.Context, arguments json.RawMessage) (interface{}, error) {
			return "sub-result", nil
		},
	}
	_ = sub.AddTool(subTool)
	_ = cs.Mount("api", sub)

	ctx := context.Background()

	// Call main tool
	result, err := cs.CallTool(ctx, "main-tool", nil)
	if err != nil {
		t.Fatalf("CallTool (main) failed: %v", err)
	}
	if result != "main-result" {
		t.Errorf("Expected 'main-result', got %v", result)
	}

	// Call prefixed sub tool
	result, err = cs.CallTool(ctx, "api/sub-tool", nil)
	if err != nil {
		t.Fatalf("CallTool (sub) failed: %v", err)
	}
	if result != "sub-result" {
		t.Errorf("Expected 'sub-result', got %v", result)
	}
}

func TestCompositeServer_ListResources(t *testing.T) {
	cs := NewCompositeServer("main")

	// Add resource to main server
	mainRes := &ResourceHandler{
		URI:      "config://main",
		Name:     "Main Config",
		MimeType: "text/plain",
		Reader: func(ctx context.Context) ([]byte, error) {
			return []byte("main"), nil
		},
	}
	_ = cs.AddResource(mainRes)

	// Create and mount sub server with resource
	sub := New("sub")
	subRes := &ResourceHandler{
		URI:      "config://sub",
		Name:     "Sub Config",
		MimeType: "text/plain",
		Reader: func(ctx context.Context) ([]byte, error) {
			return []byte("sub"), nil
		},
	}
	_ = sub.AddResource(subRes)
	_ = cs.Mount("api", sub)

	// List all resources
	resources := cs.ListResources()

	if len(resources) != 2 {
		t.Fatalf("Expected 2 resources, got %d", len(resources))
	}

	// Verify resource URIs
	foundMain := false
	foundSub := false
	for _, res := range resources {
		if res.URI == "config://main" {
			foundMain = true
		}
		if res.URI == "api/config://sub" {
			foundSub = true
		}
	}

	if !foundMain {
		t.Error("Main resource not found")
	}
	if !foundSub {
		t.Error("Prefixed sub resource not found")
	}
}

func TestCompositeServer_ReadResource(t *testing.T) {
	cs := NewCompositeServer("main")

	// Add resource to main server
	mainRes := &ResourceHandler{
		URI:      "config://main",
		Name:     "Main Config",
		MimeType: "text/plain",
		Reader: func(ctx context.Context) ([]byte, error) {
			return []byte("main-data"), nil
		},
	}
	_ = cs.AddResource(mainRes)

	// Create and mount sub server with resource
	sub := New("sub")
	subRes := &ResourceHandler{
		URI:      "config://sub",
		Name:     "Sub Config",
		MimeType: "text/plain",
		Reader: func(ctx context.Context) ([]byte, error) {
			return []byte("sub-data"), nil
		},
	}
	_ = sub.AddResource(subRes)
	_ = cs.Mount("api", sub)

	ctx := context.Background()

	// Read main resource
	result, err := cs.ReadResource(ctx, "config://main")
	if err != nil {
		t.Fatalf("ReadResource (main) failed: %v", err)
	}
	if string(result) != "main-data" {
		t.Errorf("Expected 'main-data', got %s", result)
	}

	// Read prefixed sub resource
	result, err = cs.ReadResource(ctx, "api/config://sub")
	if err != nil {
		t.Fatalf("ReadResource (sub) failed: %v", err)
	}
	if string(result) != "sub-data" {
		t.Errorf("Expected 'sub-data', got %s", result)
	}
}

func TestCompositeServer_ListPrompts(t *testing.T) {
	cs := NewCompositeServer("main")

	// Add prompt to main server
	mainPrompt := &PromptHandler{
		Name:        "main-prompt",
		Description: "Main prompt",
		Renderer: func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{{Role: "user", Content: []mcp.Content{&mcp.TextContent{Type: "text", Text: "main"}}}}, nil
		},
	}
	_ = cs.AddPrompt(mainPrompt)

	// Create and mount sub server with prompt
	sub := New("sub")
	subPrompt := &PromptHandler{
		Name:        "sub-prompt",
		Description: "Sub prompt",
		Renderer: func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{{Role: "user", Content: []mcp.Content{&mcp.TextContent{Type: "text", Text: "sub"}}}}, nil
		},
	}
	_ = sub.AddPrompt(subPrompt)
	_ = cs.Mount("api", sub)

	// List all prompts
	prompts := cs.ListPrompts()

	if len(prompts) != 2 {
		t.Fatalf("Expected 2 prompts, got %d", len(prompts))
	}

	// Verify prompt names
	foundMain := false
	foundSub := false
	for _, prompt := range prompts {
		if prompt.Name == "main-prompt" {
			foundMain = true
		}
		if prompt.Name == "api/sub-prompt" {
			foundSub = true
		}
	}

	if !foundMain {
		t.Error("Main prompt not found")
	}
	if !foundSub {
		t.Error("Prefixed sub prompt not found")
	}
}

func TestCompositeServer_GetPrompt(t *testing.T) {
	cs := NewCompositeServer("main")

	// Add prompt to main server
	mainPrompt := &PromptHandler{
		Name:        "main-prompt",
		Description: "Main prompt",
		Renderer: func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{{Role: "user", Content: []mcp.Content{&mcp.TextContent{Type: "text", Text: "main-message"}}}}, nil
		},
	}
	_ = cs.AddPrompt(mainPrompt)

	// Create and mount sub server with prompt
	sub := New("sub")
	subPrompt := &PromptHandler{
		Name:        "sub-prompt",
		Description: "Sub prompt",
		Renderer: func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
			return []*mcp.PromptMessage{{Role: "user", Content: []mcp.Content{&mcp.TextContent{Type: "text", Text: "sub-message"}}}}, nil
		},
	}
	_ = sub.AddPrompt(subPrompt)
	_ = cs.Mount("api", sub)

	ctx := context.Background()

	// Get main prompt
	result, err := cs.GetPrompt(ctx, "main-prompt", nil)
	if err != nil {
		t.Fatalf("GetPrompt (main) failed: %v", err)
	}
	if len(result) == 0 || len(result[0].Content) == 0 {
		t.Fatalf("Expected messages with content, got %+v", result)
	}
	textContent, ok := result[0].Content[0].(*mcp.TextContent)
	if !ok || textContent.Text != "main-message" {
		t.Errorf("Expected 'main-message', got %+v", result[0].Content[0])
	}

	// Get prefixed sub prompt
	result, err = cs.GetPrompt(ctx, "api/sub-prompt", nil)
	if err != nil {
		t.Fatalf("GetPrompt (sub) failed: %v", err)
	}
	if len(result) == 0 || len(result[0].Content) == 0 {
		t.Fatalf("Expected messages with content, got %+v", result)
	}
	textContent, ok = result[0].Content[0].(*mcp.TextContent)
	if !ok || textContent.Text != "sub-message" {
		t.Errorf("Expected 'sub-message', got %+v", result[0].Content[0])
	}
}
