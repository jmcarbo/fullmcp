package mcp

import (
	"encoding/json"
	"testing"
)

// Test Tool with OutputSchema (2025-06-18)
func TestTool_OutputSchema(t *testing.T) {
	tool := Tool{
		Name:        "analyze",
		Description: "Analyzes code",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
		OutputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"score": map[string]interface{}{"type": "number"},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var tool2 Tool
	if err := json.Unmarshal(data, &tool2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if tool2.OutputSchema == nil {
		t.Error("expected OutputSchema to be preserved")
	}

	props, ok := tool2.OutputSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("expected properties in OutputSchema")
	}

	if _, exists := props["score"]; !exists {
		t.Error("expected score property in OutputSchema")
	}
}

// Test Resource with Title and Meta (2025-06-18)
func TestResource_TitleAndMeta(t *testing.T) {
	resource := Resource{
		URI:   "file:///test.txt",
		Name:  "test_file",
		Title: "Test File",
		Meta: map[string]interface{}{
			"version":  "1.0",
			"priority": 0.8,
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var resource2 Resource
	if err := json.Unmarshal(data, &resource2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resource2.Title != "Test File" {
		t.Errorf("expected title 'Test File', got '%s'", resource2.Title)
	}

	if resource2.Meta == nil {
		t.Fatal("expected Meta to be preserved")
	}

	if v, ok := resource2.Meta["version"]; !ok || v != "1.0" {
		t.Errorf("expected version '1.0', got %v", v)
	}
}

// Test ResourceTemplate with Title and Meta (2025-06-18)
func TestResourceTemplate_TitleAndMeta(t *testing.T) {
	template := ResourceTemplate{
		URITemplate: "file:///{path}",
		Name:        "file_template",
		Title:       "File Template",
		Meta: map[string]interface{}{
			"category": "files",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(template)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var template2 ResourceTemplate
	if err := json.Unmarshal(data, &template2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if template2.Title != "File Template" {
		t.Errorf("expected title 'File Template', got '%s'", template2.Title)
	}

	if template2.Meta == nil {
		t.Fatal("expected Meta to be preserved")
	}
}

// Test Prompt with Title and Meta (2025-06-18)
func TestPrompt_TitleAndMeta(t *testing.T) {
	prompt := Prompt{
		Name:  "review",
		Title: "Code Review",
		Meta: map[string]interface{}{
			"category": "development",
			"tags":     []string{"code", "review"},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(prompt)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var prompt2 Prompt
	if err := json.Unmarshal(data, &prompt2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if prompt2.Title != "Code Review" {
		t.Errorf("expected title 'Code Review', got '%s'", prompt2.Title)
	}

	if prompt2.Meta == nil {
		t.Fatal("expected Meta to be preserved")
	}
}

// Test ResourceLinkContent (2025-06-18)
func TestResourceLinkContent(t *testing.T) {
	link := ResourceLinkContent{
		Type: "resource",
		Resource: Resource{
			URI:  "file:///report.pdf",
			Name: "report",
		},
		Annotations: map[string]interface{}{
			"size": 1024,
		},
	}

	if link.ContentType() != "resource" {
		t.Errorf("expected type 'resource', got '%s'", link.ContentType())
	}

	// Marshal to JSON
	data, err := json.Marshal(link)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var link2 ResourceLinkContent
	if err := json.Unmarshal(data, &link2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if link2.Resource.URI != "file:///report.pdf" {
		t.Errorf("expected URI 'file:///report.pdf', got '%s'", link2.Resource.URI)
	}

	if link2.Annotations == nil {
		t.Fatal("expected Annotations to be preserved")
	}
}

// Test ClientCapabilities (2025-06-18)
func TestClientCapabilities(t *testing.T) {
	caps := ClientCapabilities{
		Roots: &RootsCapability{
			ListChanged: true,
		},
		Sampling: &SamplingCapability{},
		Elicitation: &ElicitationCapability{},
	}

	// Marshal to JSON
	data, err := json.Marshal(caps)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var caps2 ClientCapabilities
	if err := json.Unmarshal(data, &caps2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if caps2.Roots == nil {
		t.Error("expected Roots to be preserved")
	}

	if caps2.Sampling == nil {
		t.Error("expected Sampling to be preserved")
	}

	if caps2.Elicitation == nil {
		t.Error("expected Elicitation to be preserved")
	}
}

// Test ElicitationRequest (2025-06-18)
func TestElicitationRequest(t *testing.T) {
	req := ElicitationRequest{
		Description: "Enter API key",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"api_key": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"api_key"},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var req2 ElicitationRequest
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req2.Description != "Enter API key" {
		t.Errorf("expected description 'Enter API key', got '%s'", req2.Description)
	}

	if req2.Schema == nil {
		t.Fatal("expected Schema to be preserved")
	}
}

// Test ElicitationResponse (2025-06-18)
func TestElicitationResponse(t *testing.T) {
	tests := []struct {
		name     string
		response ElicitationResponse
	}{
		{
			name: "accept",
			response: ElicitationResponse{
				Action: "accept",
				Data: map[string]interface{}{
					"api_key": "sk-123",
				},
			},
		},
		{
			name: "decline",
			response: ElicitationResponse{
				Action: "decline",
			},
		},
		{
			name: "cancel",
			response: ElicitationResponse{
				Action: "cancel",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			// Unmarshal back
			var resp2 ElicitationResponse
			if err := json.Unmarshal(data, &resp2); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if resp2.Action != tt.response.Action {
				t.Errorf("expected action '%s', got '%s'", tt.response.Action, resp2.Action)
			}

			if tt.response.Data != nil && resp2.Data == nil {
				t.Error("expected Data to be preserved")
			}
		})
	}
}

// Test JSON marshaling of all new fields together
func TestNewFields_JSONRoundTrip(t *testing.T) {
	tool := Tool{
		Name:         "test",
		InputSchema:  map[string]interface{}{"type": "object"},
		OutputSchema: map[string]interface{}{"type": "string"},
		Title:        "Test Tool",
	}

	resource := Resource{
		URI:   "file:///test",
		Name:  "test",
		Title: "Test Resource",
		Meta:  map[string]interface{}{"version": "1.0"},
	}

	prompt := Prompt{
		Name:  "test",
		Title: "Test Prompt",
		Meta:  map[string]interface{}{"category": "test"},
	}

	// Test that all can be marshaled
	if _, err := json.Marshal(tool); err != nil {
		t.Errorf("tool marshal failed: %v", err)
	}

	if _, err := json.Marshal(resource); err != nil {
		t.Errorf("resource marshal failed: %v", err)
	}

	if _, err := json.Marshal(prompt); err != nil {
		t.Errorf("prompt marshal failed: %v", err)
	}
}
