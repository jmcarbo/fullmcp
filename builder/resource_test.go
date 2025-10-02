package builder

import (
	"context"
	"testing"
)

func TestResourceBuilder_Build(t *testing.T) {
	resource := NewResource("config://app").
		Name("App Config").
		Description("Application configuration").
		MimeType("application/json").
		Reader(func(ctx context.Context) ([]byte, error) {
			return []byte(`{"debug": true}`), nil
		}).
		Tags("config", "app").
		Build()

	if resource.URI != "config://app" {
		t.Errorf("expected URI 'config://app', got '%s'", resource.URI)
	}

	if resource.Name != "App Config" {
		t.Errorf("expected name 'App Config', got '%s'", resource.Name)
	}

	if resource.Description != "Application configuration" {
		t.Errorf("expected specific description, got '%s'", resource.Description)
	}

	if resource.MimeType != "application/json" {
		t.Errorf("expected MIME type 'application/json', got '%s'", resource.MimeType)
	}

	if len(resource.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(resource.Tags))
	}

	if resource.Reader == nil {
		t.Error("expected reader to be set")
	}
}

func TestResourceBuilder_Reader_Execution(t *testing.T) {
	expectedData := []byte(`{"status": "ok"}`)

	resource := NewResource("test://data").
		Reader(func(ctx context.Context) ([]byte, error) {
			return expectedData, nil
		}).
		Build()

	ctx := context.Background()
	data, err := resource.Reader(ctx)
	if err != nil {
		t.Fatalf("reader execution failed: %v", err)
	}

	if string(data) != string(expectedData) {
		t.Errorf("expected %s, got %s", expectedData, data)
	}
}

func TestResourceBuilder_Chaining(t *testing.T) {
	resource := NewResource("file://test.txt").
		Name("Test File").
		Description("A test file").
		MimeType("text/plain").
		Tags("test").
		Reader(func(ctx context.Context) ([]byte, error) {
			return []byte("test content"), nil
		}).
		Build()

	if resource.URI != "file://test.txt" {
		t.Errorf("expected URI 'file://test.txt', got '%s'", resource.URI)
	}

	if resource.Name != "Test File" {
		t.Errorf("expected name 'Test File', got '%s'", resource.Name)
	}
}

func TestResourceTemplateBuilder_Build(t *testing.T) {
	template := NewResourceTemplate("file:///{path}").
		Name("File Reader").
		Description("Read files").
		MimeType("text/plain").
		Reader(func(ctx context.Context, params map[string]string) ([]byte, error) {
			path := params["path"]
			return []byte("content of " + path), nil
		}).
		Tags("file", "template").
		Build()

	if template.URITemplate != "file:///{path}" {
		t.Errorf("expected URI template 'file:///{path}', got '%s'", template.URITemplate)
	}

	if template.Name != "File Reader" {
		t.Errorf("expected name 'File Reader', got '%s'", template.Name)
	}

	if template.Description != "Read files" {
		t.Errorf("expected specific description, got '%s'", template.Description)
	}

	if template.MimeType != "text/plain" {
		t.Errorf("expected MIME type 'text/plain', got '%s'", template.MimeType)
	}

	if len(template.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(template.Tags))
	}

	if template.Reader == nil {
		t.Error("expected reader to be set")
	}
}

func TestResourceTemplateBuilder_Reader_Execution(t *testing.T) {
	template := NewResourceTemplate("data:///{id}").
		Reader(func(ctx context.Context, params map[string]string) ([]byte, error) {
			id := params["id"]
			return []byte("data-" + id), nil
		}).
		Build()

	ctx := context.Background()
	params := map[string]string{"id": "123"}

	data, err := template.Reader(ctx, params)
	if err != nil {
		t.Fatalf("reader execution failed: %v", err)
	}

	expected := "data-123"
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, data)
	}
}

func TestResourceTemplateBuilder_ReaderSimple(t *testing.T) {
	template := NewResourceTemplate("file:///{path}").
		ReaderSimple(func(ctx context.Context, path string) ([]byte, error) {
			return []byte("file: " + path), nil
		}).
		Build()

	ctx := context.Background()
	params := map[string]string{"path": "test.txt"}

	data, err := template.Reader(ctx, params)
	if err != nil {
		t.Fatalf("reader execution failed: %v", err)
	}

	expected := "file: test.txt"
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, data)
	}
}

func TestResourceTemplateBuilder_ReaderSimple_NoParams(t *testing.T) {
	template := NewResourceTemplate("file:///{path}").
		ReaderSimple(func(ctx context.Context, path string) ([]byte, error) {
			return []byte("file: " + path), nil
		}).
		Build()

	ctx := context.Background()
	params := map[string]string{} // Empty params

	_, err := template.Reader(ctx, params)
	if err == nil {
		t.Error("expected error for empty parameters")
	}
}

func TestResourceTemplateBuilder_Chaining(t *testing.T) {
	template := NewResourceTemplate("user:///{id}").
		Name("User Data").
		Description("User information").
		MimeType("application/json").
		Tags("user", "data").
		ReaderSimple(func(ctx context.Context, id string) ([]byte, error) {
			return []byte(`{"id": "` + id + `"}`), nil
		}).
		Build()

	if template.URITemplate != "user:///{id}" {
		t.Errorf("expected URI template 'user:///{id}', got '%s'", template.URITemplate)
	}

	if template.Name != "User Data" {
		t.Errorf("expected name 'User Data', got '%s'", template.Name)
	}
}
