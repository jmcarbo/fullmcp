package server

import (
	"context"
	"errors"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

func TestResourceManager_Register(t *testing.T) {
	rm := NewResourceManager()

	handler := &ResourceHandler{
		URI:         "test://resource",
		Name:        "Test Resource",
		Description: "A test resource",
		Reader: func(ctx context.Context) ([]byte, error) {
			return []byte("content"), nil
		},
	}

	err := rm.Register(handler)
	if err != nil {
		t.Fatalf("failed to register resource: %v", err)
	}

	// Registering same URI again should overwrite
	err = rm.Register(handler)
	if err != nil {
		t.Fatalf("failed to re-register resource: %v", err)
	}
}

func TestResourceManager_Read(t *testing.T) {
	rm := NewResourceManager()

	expectedContent := []byte("test content")
	handler := &ResourceHandler{
		URI: "file:///test.txt",
		Reader: func(ctx context.Context) ([]byte, error) {
			return expectedContent, nil
		},
	}

	rm.Register(handler)

	ctx := context.Background()
	content, err := rm.Read(ctx, "file:///test.txt")
	if err != nil {
		t.Fatalf("failed to read resource: %v", err)
	}

	if string(content) != string(expectedContent) {
		t.Errorf("expected '%s', got '%s'", string(expectedContent), string(content))
	}
}

func TestResourceManager_Read_NotFound(t *testing.T) {
	rm := NewResourceManager()
	ctx := context.Background()

	_, err := rm.Read(ctx, "nonexistent://resource")
	if err == nil {
		t.Error("expected error for nonexistent resource")
	}

	notFoundErr, ok := err.(*mcp.NotFoundError)
	if !ok {
		t.Errorf("expected NotFoundError, got %T", err)
	}

	if notFoundErr.Type != "resource" {
		t.Errorf("expected type 'resource', got '%s'", notFoundErr.Type)
	}
}

func TestResourceManager_Read_ReaderError(t *testing.T) {
	rm := NewResourceManager()

	expectedErr := errors.New("read failed")
	handler := &ResourceHandler{
		URI: "error://resource",
		Reader: func(ctx context.Context) ([]byte, error) {
			return nil, expectedErr
		},
	}

	rm.Register(handler)

	ctx := context.Background()
	_, err := rm.Read(ctx, "error://resource")
	if err == nil {
		t.Error("expected error from reader")
	}

	if err != expectedErr {
		t.Errorf("expected specific error, got %v", err)
	}
}

func TestResourceManager_List(t *testing.T) {
	rm := NewResourceManager()

	handlers := []*ResourceHandler{
		{
			URI:         "file:///resource1.txt",
			Name:        "Resource 1",
			Description: "First resource",
			MimeType:    "text/plain",
		},
		{
			URI:         "file:///resource2.json",
			Name:        "Resource 2",
			Description: "Second resource",
			MimeType:    "application/json",
		},
	}

	for _, handler := range handlers {
		handler.Reader = func(ctx context.Context) ([]byte, error) {
			return []byte("content"), nil
		}
		rm.Register(handler)
	}

	resources := rm.List()
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}

	uris := make(map[string]bool)
	for _, resource := range resources {
		uris[resource.URI] = true
	}

	if !uris["file:///resource1.txt"] || !uris["file:///resource2.json"] {
		t.Error("expected both resources in list")
	}
}

func TestResourceManager_ConcurrentAccess(t *testing.T) {
	rm := NewResourceManager()

	handler := &ResourceHandler{
		URI: "concurrent://resource",
		Reader: func(ctx context.Context) ([]byte, error) {
			return []byte("ok"), nil
		},
	}

	rm.Register(handler)

	ctx := context.Background()
	done := make(chan bool)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_, err := rm.Read(ctx, "concurrent://resource")
			if err != nil {
				t.Errorf("concurrent read failed: %v", err)
			}
			done <- true
		}()
	}

	// Concurrent list operations
	for i := 0; i < 10; i++ {
		go func() {
			_ = rm.List()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestResourceHandler_WithMimeType(t *testing.T) {
	handler := &ResourceHandler{
		URI:      "test://resource",
		MimeType: "application/json",
		Reader: func(ctx context.Context) ([]byte, error) {
			return []byte(`{"key":"value"}`), nil
		},
	}

	if handler.MimeType != "application/json" {
		t.Errorf("expected mime type 'application/json', got '%s'", handler.MimeType)
	}
}
