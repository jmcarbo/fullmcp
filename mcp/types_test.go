package mcp

import (
	"encoding/json"
	"testing"
)

func TestTextContent_ContentType(t *testing.T) {
	tc := TextContent{Type: "text", Text: "hello"}
	if tc.ContentType() != "text" {
		t.Errorf("expected 'text', got '%s'", tc.ContentType())
	}
}

func TestImageContent_ContentType(t *testing.T) {
	ic := ImageContent{Type: "image", Data: "base64data", MimeType: "image/png"}
	if ic.ContentType() != "image" {
		t.Errorf("expected 'image', got '%s'", ic.ContentType())
	}
}

func TestResourceContent_ContentType(t *testing.T) {
	rc := ResourceContent{Type: "resource", URI: "file:///test"}
	if rc.ContentType() != "resource" {
		t.Errorf("expected 'resource', got '%s'", rc.ContentType())
	}
}

func TestTool_Marshal(t *testing.T) {
	tool := Tool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param": map[string]interface{}{"type": "string"},
			},
		},
	}

	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("failed to marshal tool: %v", err)
	}

	var unmarshaled Tool
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal tool: %v", err)
	}

	if unmarshaled.Name != tool.Name {
		t.Errorf("expected name '%s', got '%s'", tool.Name, unmarshaled.Name)
	}
}

func TestResource_Marshal(t *testing.T) {
	resource := Resource{
		URI:         "file:///test.txt",
		Name:        "Test Resource",
		Description: "A test resource",
		MimeType:    "text/plain",
	}

	data, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf("failed to marshal resource: %v", err)
	}

	var unmarshaled Resource
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal resource: %v", err)
	}

	if unmarshaled.URI != resource.URI {
		t.Errorf("expected URI '%s', got '%s'", resource.URI, unmarshaled.URI)
	}
}

func TestPrompt_Marshal(t *testing.T) {
	prompt := Prompt{
		Name:        "test-prompt",
		Description: "A test prompt",
		Arguments: []PromptArgument{
			{Name: "name", Description: "User name", Required: true},
		},
	}

	data, err := json.Marshal(prompt)
	if err != nil {
		t.Fatalf("failed to marshal prompt: %v", err)
	}

	var unmarshaled Prompt
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal prompt: %v", err)
	}

	if unmarshaled.Name != prompt.Name {
		t.Errorf("expected name '%s', got '%s'", prompt.Name, unmarshaled.Name)
	}

	if len(unmarshaled.Arguments) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(unmarshaled.Arguments))
	}
}

func TestMessage_Marshal(t *testing.T) {
	msg := Message{
		JSONRPC: "2.0",
		ID:      123,
		Method:  "test/method",
		Params:  json.RawMessage(`{"key":"value"}`),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	var unmarshaled Message
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if unmarshaled.Method != msg.Method {
		t.Errorf("expected method '%s', got '%s'", msg.Method, unmarshaled.Method)
	}
}

func TestMessage_WithError(t *testing.T) {
	msg := Message{
		JSONRPC: "2.0",
		ID:      123,
		Error: &RPCError{
			Code:    -32600,
			Message: "Invalid Request",
			Data:    "additional info",
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	var unmarshaled Message
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if unmarshaled.Error == nil {
		t.Fatal("expected error to be set")
	}

	if unmarshaled.Error.Code != msg.Error.Code {
		t.Errorf("expected error code %d, got %d", msg.Error.Code, unmarshaled.Error.Code)
	}
}

func TestServerCapabilities(t *testing.T) {
	caps := ServerCapabilities{
		Tools: &ToolsCapability{
			ListChanged: true,
		},
		Resources: &ResourcesCapability{
			Subscribe:   true,
			ListChanged: true,
		},
		Prompts: &PromptsCapability{
			ListChanged: false,
		},
	}

	data, err := json.Marshal(caps)
	if err != nil {
		t.Fatalf("failed to marshal capabilities: %v", err)
	}

	var unmarshaled ServerCapabilities
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal capabilities: %v", err)
	}

	if unmarshaled.Tools == nil || !unmarshaled.Tools.ListChanged {
		t.Error("expected tools.listChanged to be true")
	}

	if unmarshaled.Resources == nil || !unmarshaled.Resources.Subscribe {
		t.Error("expected resources.subscribe to be true")
	}
}

// Helper functions for TestPromptMessageUnmarshalJSON to reduce cyclomatic complexity
func verifySingleTextContent(t *testing.T, pm *PromptMessage) {
	if pm.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", pm.Role)
	}
	if len(pm.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(pm.Content))
	}
	tc, ok := pm.Content[0].(TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", pm.Content[0])
	}
	if tc.Type != "text" {
		t.Errorf("Expected type 'text', got '%s'", tc.Type)
	}
	if tc.Text != "Hello, World!" {
		t.Errorf("Expected text 'Hello, World!', got '%s'", tc.Text)
	}
}

func verifyImageContent(t *testing.T, pm *PromptMessage) {
	if pm.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", pm.Role)
	}
	if len(pm.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(pm.Content))
	}
	ic, ok := pm.Content[0].(ImageContent)
	if !ok {
		t.Fatalf("Expected ImageContent, got %T", pm.Content[0])
	}
	if ic.Type != "image" {
		t.Errorf("Expected type 'image', got '%s'", ic.Type)
	}
	if ic.Data != "base64encodeddata" {
		t.Errorf("Expected data 'base64encodeddata', got '%s'", ic.Data)
	}
	if ic.MimeType != "image/png" {
		t.Errorf("Expected mimeType 'image/png', got '%s'", ic.MimeType)
	}
}

func verifyAudioContent(t *testing.T, pm *PromptMessage) {
	if len(pm.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(pm.Content))
	}
	ac, ok := pm.Content[0].(AudioContent)
	if !ok {
		t.Fatalf("Expected AudioContent, got %T", pm.Content[0])
	}
	if ac.Type != "audio" {
		t.Errorf("Expected type 'audio', got '%s'", ac.Type)
	}
	if ac.Data != "base64audiodata" {
		t.Errorf("Expected data 'base64audiodata', got '%s'", ac.Data)
	}
	if ac.MimeType != "audio/mp3" {
		t.Errorf("Expected mimeType 'audio/mp3', got '%s'", ac.MimeType)
	}
}

func verifyResourceContent(t *testing.T, pm *PromptMessage) {
	if len(pm.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(pm.Content))
	}
	rc, ok := pm.Content[0].(ResourceContent)
	if !ok {
		t.Fatalf("Expected ResourceContent, got %T", pm.Content[0])
	}
	if rc.Type != "resource" {
		t.Errorf("Expected type 'resource', got '%s'", rc.Type)
	}
	if rc.URI != "file:///test.txt" {
		t.Errorf("Expected uri 'file:///test.txt', got '%s'", rc.URI)
	}
	if rc.MimeType != "text/plain" {
		t.Errorf("Expected mimeType 'text/plain', got '%s'", rc.MimeType)
	}
	if rc.Text != "content" {
		t.Errorf("Expected text 'content', got '%s'", rc.Text)
	}
}

func verifyResourceLinkContent(t *testing.T, pm *PromptMessage) {
	if len(pm.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(pm.Content))
	}
	rlc, ok := pm.Content[0].(ResourceLinkContent)
	if !ok {
		t.Fatalf("Expected ResourceLinkContent, got %T", pm.Content[0])
	}
	if rlc.Type != "resource" {
		t.Errorf("Expected type 'resource', got '%s'", rlc.Type)
	}
	if rlc.Resource.URI != "test://resource" {
		t.Errorf("Expected uri 'test://resource', got '%s'", rlc.Resource.URI)
	}
	if rlc.Resource.Name != "Test Resource" {
		t.Errorf("Expected name 'Test Resource', got '%s'", rlc.Resource.Name)
	}
}

func verifyMixedContent(t *testing.T, pm *PromptMessage) {
	if len(pm.Content) != 4 {
		t.Fatalf("Expected 4 content items, got %d", len(pm.Content))
	}

	tc1, ok := pm.Content[0].(TextContent)
	if !ok {
		t.Errorf("Content[0]: Expected TextContent, got %T", pm.Content[0])
	} else if tc1.Text != "Here's an image:" {
		t.Errorf("Content[0]: Expected 'Here's an image:', got '%s'", tc1.Text)
	}

	ic, ok := pm.Content[1].(ImageContent)
	if !ok {
		t.Errorf("Content[1]: Expected ImageContent, got %T", pm.Content[1])
	} else if ic.Data != "img123" {
		t.Errorf("Content[1]: Expected 'img123', got '%s'", ic.Data)
	}

	tc2, ok := pm.Content[2].(TextContent)
	if !ok {
		t.Errorf("Content[2]: Expected TextContent, got %T", pm.Content[2])
	} else if tc2.Text != "And some audio:" {
		t.Errorf("Content[2]: Expected 'And some audio:', got '%s'", tc2.Text)
	}

	ac, ok := pm.Content[3].(AudioContent)
	if !ok {
		t.Errorf("Content[3]: Expected AudioContent, got %T", pm.Content[3])
	} else if ac.Data != "audio456" {
		t.Errorf("Content[3]: Expected 'audio456', got '%s'", ac.Data)
	}
}

func verifyEmptyContent(t *testing.T, pm *PromptMessage) {
	if pm.Role != "user" {
		t.Errorf("Expected role 'user', got '%s'", pm.Role)
	}
	if len(pm.Content) != 0 {
		t.Errorf("Expected 0 content items, got %d", len(pm.Content))
	}
}

func verifyUnknownTypeFallback(t *testing.T, pm *PromptMessage) {
	if len(pm.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(pm.Content))
	}
	tc, ok := pm.Content[0].(TextContent)
	if !ok {
		t.Fatalf("Expected TextContent fallback, got %T", pm.Content[0])
	}
	if tc.Type != "unknown_type" {
		t.Errorf("Expected type 'unknown_type', got '%s'", tc.Type)
	}
}

func verifyMissingRole(t *testing.T, pm *PromptMessage) {
	if pm.Role != "" {
		t.Errorf("Expected empty role, got '%s'", pm.Role)
	}
	if len(pm.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(pm.Content))
	}
}

func TestPromptMessageUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantErr  bool
		verify   func(*testing.T, *PromptMessage)
	}{
		{
			name: "single text content",
			jsonData: `{
				"role": "user",
				"content": [
					{"type": "text", "text": "Hello, World!"}
				]
			}`,
			wantErr: false,
			verify:  verifySingleTextContent,
		},
		{
			name: "image content",
			jsonData: `{
				"role": "assistant",
				"content": [
					{
						"type": "image",
						"data": "base64encodeddata",
						"mimeType": "image/png"
					}
				]
			}`,
			wantErr: false,
			verify:  verifyImageContent,
		},
		{
			name: "audio content",
			jsonData: `{
				"role": "user",
				"content": [
					{
						"type": "audio",
						"data": "base64audiodata",
						"mimeType": "audio/mp3"
					}
				]
			}`,
			wantErr: false,
			verify:  verifyAudioContent,
		},
		{
			name: "resource content",
			jsonData: `{
				"role": "user",
				"content": [
					{
						"type": "resource",
						"uri": "file:///test.txt",
						"mimeType": "text/plain",
						"text": "content"
					}
				]
			}`,
			wantErr: false,
			verify:  verifyResourceContent,
		},
		{
			name: "resource link content",
			jsonData: `{
				"role": "user",
				"content": [
					{
						"type": "resource",
						"resource": {
							"uri": "test://resource",
							"name": "Test Resource"
						}
					}
				]
			}`,
			wantErr: false,
			verify:  verifyResourceLinkContent,
		},
		{
			name: "mixed content types",
			jsonData: `{
				"role": "assistant",
				"content": [
					{"type": "text", "text": "Here's an image:"},
					{"type": "image", "data": "img123", "mimeType": "image/jpeg"},
					{"type": "text", "text": "And some audio:"},
					{"type": "audio", "data": "audio456", "mimeType": "audio/wav"}
				]
			}`,
			wantErr: false,
			verify:  verifyMixedContent,
		},
		{
			name: "empty content array",
			jsonData: `{
				"role": "user",
				"content": []
			}`,
			wantErr: false,
			verify:  verifyEmptyContent,
		},
		{
			name: "unknown content type - fallback to text",
			jsonData: `{
				"role": "user",
				"content": [
					{"type": "unknown_type", "text": "fallback text"}
				]
			}`,
			wantErr: false,
			verify:  verifyUnknownTypeFallback,
		},
		{
			name:     "invalid json",
			jsonData: `{"role": "user", "content": [invalid]}`,
			wantErr:  true,
		},
		{
			name:     "missing role field",
			jsonData: `{"content": [{"type": "text", "text": "hello"}]}`,
			wantErr:  false,
			verify:   verifyMissingRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pm PromptMessage
			err := json.Unmarshal([]byte(tt.jsonData), &pm)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.verify != nil {
				tt.verify(t, &pm)
			}
		})
	}
}

func TestContentTypeMethod(t *testing.T) {
	tests := []struct {
		name     string
		content  Content
		wantType string
	}{
		{
			name:     "TextContent",
			content:  TextContent{Type: "text", Text: "hello"},
			wantType: "text",
		},
		{
			name:     "ImageContent",
			content:  ImageContent{Type: "image", Data: "data", MimeType: "image/png"},
			wantType: "image",
		},
		{
			name:     "AudioContent",
			content:  AudioContent{Type: "audio", Data: "data", MimeType: "audio/mp3"},
			wantType: "audio",
		},
		{
			name:     "ResourceContent",
			content:  ResourceContent{Type: "resource", URI: "file:///test"},
			wantType: "resource",
		},
		{
			name: "ResourceLinkContent",
			content: ResourceLinkContent{
				Type:     "resource",
				Resource: Resource{URI: "test://res", Name: "Test"},
			},
			wantType: "resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.content.ContentType()
			if got != tt.wantType {
				t.Errorf("ContentType() = %v, want %v", got, tt.wantType)
			}
		})
	}
}
