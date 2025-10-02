package server

import (
	"context"
	"regexp"
	"testing"
)

func TestResourceManager_RegisterTemplate(t *testing.T) {
	rm := NewResourceManager()

	handler := &ResourceTemplateHandler{
		URITemplate: "file:///{path}",
		Name:        "File Reader",
		Description: "Read files",
		MimeType:    "text/plain",
		Reader: func(ctx context.Context, params map[string]string) ([]byte, error) {
			return []byte("content of " + params["path"]), nil
		},
	}

	err := rm.RegisterTemplate(handler)
	if err != nil {
		t.Fatalf("RegisterTemplate failed: %v", err)
	}

	// Verify pattern was compiled
	if handler.pattern == nil {
		t.Error("expected pattern to be compiled")
	}
}

func TestResourceManager_RegisterTemplate_Success(t *testing.T) {
	rm := NewResourceManager()

	handler := &ResourceTemplateHandler{
		URITemplate: "file:///{path}",
		Reader: func(ctx context.Context, params map[string]string) ([]byte, error) {
			return []byte("test"), nil
		},
	}

	err := rm.RegisterTemplate(handler)
	if err != nil {
		t.Fatalf("RegisterTemplate failed: %v", err)
	}

	templates := rm.ListTemplates()
	if len(templates) != 1 {
		t.Errorf("expected 1 template, got %d", len(templates))
	}
}

func TestResourceManager_ListTemplates(t *testing.T) {
	rm := NewResourceManager()

	handler1 := &ResourceTemplateHandler{
		URITemplate: "file:///{path}",
		Name:        "File Reader",
		Description: "Read files",
		MimeType:    "text/plain",
		Reader: func(ctx context.Context, params map[string]string) ([]byte, error) {
			return nil, nil
		},
	}

	handler2 := &ResourceTemplateHandler{
		URITemplate: "data:///{id}",
		Name:        "Data Reader",
		Description: "Read data",
		MimeType:    "application/json",
		Reader: func(ctx context.Context, params map[string]string) ([]byte, error) {
			return nil, nil
		},
	}

	rm.RegisterTemplate(handler1)
	rm.RegisterTemplate(handler2)

	templates := rm.ListTemplates()

	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}

	// Check that templates have correct fields
	foundFile := false
	foundData := false

	for _, tmpl := range templates {
		if tmpl.URITemplate == "file:///{path}" {
			foundFile = true
			if tmpl.Name != "File Reader" {
				t.Errorf("expected name 'File Reader', got '%s'", tmpl.Name)
			}
			if tmpl.MimeType != "text/plain" {
				t.Errorf("expected MIME type 'text/plain', got '%s'", tmpl.MimeType)
			}
		}

		if tmpl.URITemplate == "data:///{id}" {
			foundData = true
			if tmpl.Name != "Data Reader" {
				t.Errorf("expected name 'Data Reader', got '%s'", tmpl.Name)
			}
		}
	}

	if !foundFile {
		t.Error("file template not found in list")
	}

	if !foundData {
		t.Error("data template not found in list")
	}
}

func TestResourceManager_Read_WithTemplate(t *testing.T) {
	rm := NewResourceManager()

	handler := &ResourceTemplateHandler{
		URITemplate: "file:///{path}",
		Reader: func(ctx context.Context, params map[string]string) ([]byte, error) {
			path := params["path"]
			return []byte("content of " + path), nil
		},
	}

	rm.RegisterTemplate(handler)

	ctx := context.Background()
	data, err := rm.Read(ctx, "file:///test.txt")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	expected := "content of test.txt"
	if string(data) != expected {
		t.Errorf("expected '%s', got '%s'", expected, data)
	}
}

func TestResourceManager_Read_TemplateAfterStatic(t *testing.T) {
	rm := NewResourceManager()

	// Register static resource first
	staticHandler := &ResourceHandler{
		URI:  "file:///static.txt",
		Name: "Static File",
		Reader: func(ctx context.Context) ([]byte, error) {
			return []byte("static content"), nil
		},
	}
	rm.Register(staticHandler)

	// Register template
	templateHandler := &ResourceTemplateHandler{
		URITemplate: "file:///{path}",
		Reader: func(ctx context.Context, params map[string]string) ([]byte, error) {
			return []byte("template content"), nil
		},
	}
	rm.RegisterTemplate(templateHandler)

	ctx := context.Background()

	// Static should take priority
	data, err := rm.Read(ctx, "file:///static.txt")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	expected := "static content"
	if string(data) != expected {
		t.Errorf("expected '%s', got '%s'", expected, data)
	}

	// Template should match other URIs
	data, err = rm.Read(ctx, "file:///other.txt")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	expected = "template content"
	if string(data) != expected {
		t.Errorf("expected '%s', got '%s'", expected, data)
	}
}

func TestResourceTemplateHandler_Match(t *testing.T) {
	handler := &ResourceTemplateHandler{
		URITemplate: "file:///{path}",
	}

	pattern := templateToRegex(handler.URITemplate)
	handler.pattern = mustCompile(pattern)

	params, ok := handler.Match("file:///test.txt")
	if !ok {
		t.Fatal("expected match")
	}

	if params["path"] != "test.txt" {
		t.Errorf("expected path 'test.txt', got '%s'", params["path"])
	}
}

func TestResourceTemplateHandler_Match_NoMatch(t *testing.T) {
	handler := &ResourceTemplateHandler{
		URITemplate: "file:///{path}",
	}

	pattern := templateToRegex(handler.URITemplate)
	handler.pattern = mustCompile(pattern)

	_, ok := handler.Match("http:///test.txt")
	if ok {
		t.Error("expected no match for different scheme")
	}
}

func TestResourceTemplateHandler_Match_MultipleParams(t *testing.T) {
	handler := &ResourceTemplateHandler{
		URITemplate: "api:///{version}/{resource}",
	}

	pattern := templateToRegex(handler.URITemplate)
	handler.pattern = mustCompile(pattern)

	params, ok := handler.Match("api:///v1/users")
	if !ok {
		t.Fatal("expected match")
	}

	if params["version"] != "v1" {
		t.Errorf("expected version 'v1', got '%s'", params["version"])
	}

	if params["resource"] != "users" {
		t.Errorf("expected resource 'users', got '%s'", params["resource"])
	}
}

func TestTemplateToRegex(t *testing.T) {
	tests := []struct {
		template string
		uri      string
		expected map[string]string
		match    bool
	}{
		{
			template: "file:///{path}",
			uri:      "file:///test.txt",
			expected: map[string]string{"path": "test.txt"},
			match:    true,
		},
		{
			template: "api:///{version}/{resource}",
			uri:      "api:///v1/users",
			expected: map[string]string{"version": "v1", "resource": "users"},
			match:    true,
		},
		{
			template: "file:///{path}",
			uri:      "http:///test.txt",
			match:    false,
		},
		{
			template: "data:///{id}",
			uri:      "data:///123",
			expected: map[string]string{"id": "123"},
			match:    true,
		},
	}

	for _, tt := range tests {
		pattern := templateToRegex(tt.template)
		re := mustCompile(pattern)

		matches := re.FindStringSubmatch(tt.uri)
		if tt.match {
			if matches == nil {
				t.Errorf("template %s should match URI %s", tt.template, tt.uri)
				continue
			}

			params := make(map[string]string)
			for i, name := range re.SubexpNames() {
				if i > 0 && name != "" {
					params[name] = matches[i]
				}
			}

			for key, expected := range tt.expected {
				if params[key] != expected {
					t.Errorf("template %s, param %s: expected '%s', got '%s'", tt.template, key, expected, params[key])
				}
			}
		} else {
			if matches != nil {
				t.Errorf("template %s should not match URI %s", tt.template, tt.uri)
			}
		}
	}
}

// Helper function
func mustCompile(pattern string) *regexp.Regexp {
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}
	return re
}
