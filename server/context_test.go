package server

import (
	"context"
	"testing"
)

func TestServerContext_WithContext(t *testing.T) {
	srv := New("test-server")
	ctx := context.Background()

	session := "test-session"
	ctx = srv.WithContext(ctx, session)

	sc := FromContext(ctx)
	if sc == nil {
		t.Fatal("expected server context to be present")
	}

	if sc.server != srv {
		t.Error("expected server to match")
	}

	if sc.session != session {
		t.Errorf("expected session '%v', got '%v'", session, sc.session)
	}
}

func TestServerContext_FromContext_NotPresent(t *testing.T) {
	ctx := context.Background()

	sc := FromContext(ctx)
	if sc != nil {
		t.Error("expected nil server context for empty context")
	}
}

func TestServerContext_ReadResource(t *testing.T) {
	srv := New("test-server")

	// Add a resource
	srv.AddResource(&ResourceHandler{
		URI:  "test://resource",
		Name: "Test Resource",
		Reader: func(ctx context.Context) ([]byte, error) {
			return []byte("test data"), nil
		},
	})

	ctx := context.Background()
	ctx = srv.WithContext(ctx, nil)

	sc := FromContext(ctx)
	if sc == nil {
		t.Fatal("expected server context")
	}

	data, err := sc.ReadResource(ctx, "test://resource")
	if err != nil {
		t.Fatalf("ReadResource failed: %v", err)
	}

	expected := "test data"
	if string(data) != expected {
		t.Errorf("expected '%s', got '%s'", expected, data)
	}
}

func TestServerContext_ReadResource_NotFound(t *testing.T) {
	srv := New("test-server")

	ctx := context.Background()
	ctx = srv.WithContext(ctx, nil)

	sc := FromContext(ctx)
	if sc == nil {
		t.Fatal("expected server context")
	}

	_, err := sc.ReadResource(ctx, "nonexistent://resource")
	if err == nil {
		t.Error("expected error for nonexistent resource")
	}
}

func TestServerContext_ReadResource_NilContext(t *testing.T) {
	var sc *Context

	ctx := context.Background()
	_, err := sc.ReadResource(ctx, "test://resource")
	if err == nil {
		t.Error("expected error for nil server context")
	}

	errCtx, ok := err.(*ErrorContext)
	if !ok {
		t.Fatalf("expected ErrorContext, got %T", err)
	}

	if errCtx.Message != "server context not available" {
		t.Errorf("unexpected error message: %s", errCtx.Message)
	}
}

func TestServerContext_CallTool(t *testing.T) {
	srv := New("test-server")

	ctx := context.Background()
	ctx = srv.WithContext(ctx, nil)

	sc := FromContext(ctx)
	if sc == nil {
		t.Fatal("expected server context")
	}

	// CallTool is not fully implemented yet
	_, err := sc.CallTool(ctx, "test-tool", nil)
	if err == nil {
		t.Error("expected error for unimplemented CallTool")
	}
}

func TestServerContext_CallTool_NilContext(t *testing.T) {
	var sc *Context

	ctx := context.Background()
	_, err := sc.CallTool(ctx, "test-tool", nil)
	if err == nil {
		t.Error("expected error for nil server context")
	}

	errCtx, ok := err.(*ErrorContext)
	if !ok {
		t.Fatalf("expected ErrorContext, got %T", err)
	}

	if errCtx.Message != "server context not available" {
		t.Errorf("unexpected error message: %s", errCtx.Message)
	}
}

func TestErrorContext_Error(t *testing.T) {
	err := &ErrorContext{Message: "test error"}

	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got '%s'", err.Error())
	}
}
