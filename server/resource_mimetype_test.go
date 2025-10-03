package server

import (
	"context"
	"encoding/json"
	"testing"
)

func TestResourceManager_ReadWithMetadata(t *testing.T) {
	rm := NewResourceManager()

	tests := []struct {
		name         string
		handler      *ResourceHandler
		wantMimeType string
		wantData     string
	}{
		{
			name: "JSON resource",
			handler: &ResourceHandler{
				URI:      "config://app",
				Name:     "config",
				MimeType: "application/json",
				Reader: func(ctx context.Context) ([]byte, error) {
					return []byte(`{"key":"value"}`), nil
				},
			},
			wantMimeType: "application/json",
			wantData:     `{"key":"value"}`,
		},
		{
			name: "Plain text resource",
			handler: &ResourceHandler{
				URI:      "text://doc",
				Name:     "document",
				MimeType: "text/plain",
				Reader: func(ctx context.Context) ([]byte, error) {
					return []byte("Hello, World!"), nil
				},
			},
			wantMimeType: "text/plain",
			wantData:     "Hello, World!",
		},
		{
			name: "Resource without MIME type (defaults to text/plain)",
			handler: &ResourceHandler{
				URI:  "data://test",
				Name: "test",
				// MimeType not set
				Reader: func(ctx context.Context) ([]byte, error) {
					return []byte("test data"), nil
				},
			},
			wantMimeType: "text/plain",
			wantData:     "test data",
		},
		{
			name: "HTML resource",
			handler: &ResourceHandler{
				URI:      "html://page",
				Name:     "page",
				MimeType: "text/html",
				Reader: func(ctx context.Context) ([]byte, error) {
					return []byte("<html><body>Hello</body></html>"), nil
				},
			},
			wantMimeType: "text/html",
			wantData:     "<html><body>Hello</body></html>",
		},
		{
			name: "XML resource",
			handler: &ResourceHandler{
				URI:      "xml://data",
				Name:     "xmldata",
				MimeType: "application/xml",
				Reader: func(ctx context.Context) ([]byte, error) {
					return []byte("<root><item>value</item></root>"), nil
				},
			},
			wantMimeType: "application/xml",
			wantData:     "<root><item>value</item></root>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rm.Register(tt.handler)
			if err != nil {
				t.Fatalf("failed to register handler: %v", err)
			}

			content, err := rm.ReadWithMetadata(context.Background(), tt.handler.URI)
			if err != nil {
				t.Fatalf("ReadWithMetadata() error = %v", err)
			}

			if content.MimeType != tt.wantMimeType {
				t.Errorf("MimeType = %v, want %v", content.MimeType, tt.wantMimeType)
			}

			if string(content.Data) != tt.wantData {
				t.Errorf("Data = %v, want %v", string(content.Data), tt.wantData)
			}

			if content.URI != tt.handler.URI {
				t.Errorf("URI = %v, want %v", content.URI, tt.handler.URI)
			}
		})
	}
}

func TestResourceManager_ReadWithMetadata_Template(t *testing.T) {
	rm := NewResourceManager()

	template := &ResourceTemplateHandler{
		URITemplate: "user://{id}",
		Name:        "user",
		MimeType:    "application/json",
		Reader: func(ctx context.Context, params map[string]string) ([]byte, error) {
			data := map[string]string{
				"id":   params["id"],
				"name": "User " + params["id"],
			}
			return json.Marshal(data)
		},
	}

	err := rm.RegisterTemplate(template)
	if err != nil {
		t.Fatalf("failed to register template: %v", err)
	}

	content, err := rm.ReadWithMetadata(context.Background(), "user://123")
	if err != nil {
		t.Fatalf("ReadWithMetadata() error = %v", err)
	}

	if content.MimeType != "application/json" {
		t.Errorf("MimeType = %v, want application/json", content.MimeType)
	}

	if content.URI != "user://123" {
		t.Errorf("URI = %v, want user://123", content.URI)
	}

	// Verify JSON content
	var result map[string]string
	if err := json.Unmarshal(content.Data, &result); err != nil {
		t.Errorf("failed to unmarshal JSON: %v", err)
	}

	if result["id"] != "123" {
		t.Errorf("id = %v, want 123", result["id"])
	}
}

func TestResourceManager_ReadWithMetadata_NotFound(t *testing.T) {
	rm := NewResourceManager()

	_, err := rm.ReadWithMetadata(context.Background(), "nonexistent://resource")
	if err == nil {
		t.Error("expected error for non-existent resource")
	}
}

func TestResourceManager_Read_BackwardCompatibility(t *testing.T) {
	rm := NewResourceManager()

	handler := &ResourceHandler{
		URI:      "test://data",
		Name:     "test",
		MimeType: "application/json",
		Reader: func(ctx context.Context) ([]byte, error) {
			return []byte(`{"test":"data"}`), nil
		},
	}

	err := rm.Register(handler)
	if err != nil {
		t.Fatalf("failed to register handler: %v", err)
	}

	// Test old Read() method still works
	data, err := rm.Read(context.Background(), "test://data")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if string(data) != `{"test":"data"}` {
		t.Errorf("Data = %v, want {\"test\":\"data\"}", string(data))
	}
}
