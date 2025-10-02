package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

func TestWithMiddleware(t *testing.T) {
	called := false
	mw := func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (*Response, error) {
			called = true
			return next(ctx, req)
		}
	}

	srv := New("test", WithMiddleware(mw))

	if len(srv.middleware) != 1 {
		t.Fatalf("expected 1 middleware, got %d", len(srv.middleware))
	}

	// Test that middleware is actually set
	handler := func(ctx context.Context, req *Request) (*Response, error) {
		return &Response{Result: "ok"}, nil
	}

	wrappedHandler := ApplyMiddleware(handler, srv.middleware)
	wrappedHandler(context.Background(), &Request{Method: "test"})

	if !called {
		t.Error("expected middleware to be called")
	}
}

func TestWithMultipleMiddleware(t *testing.T) {
	callCount := 0

	mw1 := func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (*Response, error) {
			callCount++
			return next(ctx, req)
		}
	}

	mw2 := func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (*Response, error) {
			callCount++
			return next(ctx, req)
		}
	}

	srv := New("test", WithMiddleware(mw1, mw2))

	if len(srv.middleware) != 2 {
		t.Fatalf("expected 2 middleware, got %d", len(srv.middleware))
	}

	handler := func(ctx context.Context, req *Request) (*Response, error) {
		return &Response{Result: "ok"}, nil
	}

	wrappedHandler := ApplyMiddleware(handler, srv.middleware)
	wrappedHandler(context.Background(), &Request{Method: "test"})

	if callCount != 2 {
		t.Errorf("expected 2 middleware calls, got %d", callCount)
	}
}

func TestWithLifespan_Option(t *testing.T) {
	lifespanCalled := false

	lifespanFunc := func(ctx context.Context, s *Server) (context.Context, func(), error) {
		lifespanCalled = true
		return ctx, func() {}, nil
	}

	srv := New("test", WithLifespan(lifespanFunc))

	if srv.lifespan == nil {
		t.Fatal("expected lifespan to be set")
	}

	// Verify it's the right function
	srv.lifespan(context.Background(), srv)

	if !lifespanCalled {
		t.Error("expected lifespan function to be called")
	}
}

func TestAddResourceTemplate_Integration(t *testing.T) {
	srv := New("test")

	handler := &ResourceTemplateHandler{
		URITemplate: "test:///{id}",
		Reader: func(ctx context.Context, params map[string]string) ([]byte, error) {
			return []byte("id: " + params["id"]), nil
		},
	}

	err := srv.AddResourceTemplate(handler)
	if err != nil {
		t.Fatalf("AddResourceTemplate failed: %v", err)
	}

	// Verify we can read using the template
	ctx := context.Background()
	data, err := srv.resources.Read(ctx, "test:///123")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	expected := "id: 123"
	if string(data) != expected {
		t.Errorf("expected '%s', got '%s'", expected, data)
	}
}

func TestHandleResourceTemplatesList(t *testing.T) {
	srv := New("test")

	// Add some templates
	srv.AddResourceTemplate(&ResourceTemplateHandler{
		URITemplate: "file:///{path}",
		Name:        "File Reader",
	})

	srv.AddResourceTemplate(&ResourceTemplateHandler{
		URITemplate: "data:///{id}",
		Name:        "Data Reader",
	})

	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "resources/templates/list",
	}

	response := srv.handleResourceTemplatesList(msg)

	if response.Error != nil {
		t.Fatalf("expected no error, got %v", response.Error)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	templates, ok := result["resourceTemplates"]
	if !ok {
		t.Fatal("expected resourceTemplates in result")
	}

	templatesList, ok := templates.([]interface{})
	if !ok {
		t.Fatalf("expected templates to be array, got %T", templates)
	}

	if len(templatesList) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templatesList))
	}
}
