package server

import (
	"context"
	"errors"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

type testLogger struct {
	infos  []string
	errors []string
}

func (l *testLogger) Infof(format string, args ...interface{}) {
	l.infos = append(l.infos, format)
}

func (l *testLogger) Errorf(format string, args ...interface{}) {
	l.errors = append(l.errors, format)
}

func TestLoggingMiddleware(t *testing.T) {
	logger := &testLogger{}
	middleware := LoggingMiddleware(logger)

	called := false
	handler := func(ctx context.Context, req *Request) (*Response, error) {
		called = true
		return &Response{Result: "ok"}, nil
	}

	wrappedHandler := middleware(handler)

	ctx := context.Background()
	req := &Request{Method: "test/method"}

	resp, err := wrappedHandler(ctx, req)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if !called {
		t.Error("expected wrapped handler to be called")
	}

	if resp.Result != "ok" {
		t.Errorf("expected result 'ok', got '%v'", resp.Result)
	}

	if len(logger.infos) != 1 {
		t.Fatalf("expected 1 info log, got %d", len(logger.infos))
	}

	if logger.infos[0] != "Request: %s" {
		t.Errorf("unexpected log message: %s", logger.infos[0])
	}
}

func TestLoggingMiddleware_WithError(t *testing.T) {
	logger := &testLogger{}
	middleware := LoggingMiddleware(logger)

	expectedErr := errors.New("handler error")
	handler := func(ctx context.Context, req *Request) (*Response, error) {
		return nil, expectedErr
	}

	wrappedHandler := middleware(handler)

	ctx := context.Background()
	req := &Request{Method: "test/method"}

	_, err := wrappedHandler(ctx, req)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if len(logger.infos) != 1 {
		t.Fatalf("expected 1 info log, got %d", len(logger.infos))
	}

	if len(logger.errors) != 1 {
		t.Fatalf("expected 1 error log, got %d", len(logger.errors))
	}

	if logger.errors[0] != "Error: %v" {
		t.Errorf("unexpected error log message: %s", logger.errors[0])
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	middleware := RecoveryMiddleware()

	handler := func(ctx context.Context, req *Request) (*Response, error) {
		panic("test panic")
	}

	wrappedHandler := middleware(handler)

	ctx := context.Background()
	req := &Request{Method: "test/method"}

	resp, err := wrappedHandler(ctx, req)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("expected response")
	}

	if resp.Error == nil {
		t.Fatal("expected error in response")
	}

	if resp.Error.Code != int(mcp.InternalError) {
		t.Errorf("expected InternalError code, got %d", resp.Error.Code)
	}

	if resp.Error.Message != "internal server error" {
		t.Errorf("expected 'internal server error', got '%s'", resp.Error.Message)
	}
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	middleware := RecoveryMiddleware()

	called := false
	handler := func(ctx context.Context, req *Request) (*Response, error) {
		called = true
		return &Response{Result: "ok"}, nil
	}

	wrappedHandler := middleware(handler)

	ctx := context.Background()
	req := &Request{Method: "test/method"}

	resp, err := wrappedHandler(ctx, req)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if !called {
		t.Error("expected handler to be called")
	}

	if resp.Result != "ok" {
		t.Errorf("expected result 'ok', got '%v'", resp.Result)
	}

	if resp.Error != nil {
		t.Error("expected no error in response")
	}
}

func TestApplyMiddleware(t *testing.T) {
	var callOrder []string

	mw1 := func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (*Response, error) {
			callOrder = append(callOrder, "mw1-before")
			resp, err := next(ctx, req)
			callOrder = append(callOrder, "mw1-after")
			return resp, err
		}
	}

	mw2 := func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (*Response, error) {
			callOrder = append(callOrder, "mw2-before")
			resp, err := next(ctx, req)
			callOrder = append(callOrder, "mw2-after")
			return resp, err
		}
	}

	handler := func(ctx context.Context, req *Request) (*Response, error) {
		callOrder = append(callOrder, "handler")
		return &Response{Result: "ok"}, nil
	}

	middleware := []Middleware{mw1, mw2}
	wrappedHandler := ApplyMiddleware(handler, middleware)

	ctx := context.Background()
	req := &Request{Method: "test"}

	_, err := wrappedHandler(ctx, req)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(callOrder) != len(expected) {
		t.Fatalf("expected %d calls, got %d", len(expected), len(callOrder))
	}

	for i, call := range expected {
		if callOrder[i] != call {
			t.Errorf("call %d: expected '%s', got '%s'", i, call, callOrder[i])
		}
	}
}

func TestApplyMiddleware_Empty(t *testing.T) {
	called := false
	handler := func(ctx context.Context, req *Request) (*Response, error) {
		called = true
		return &Response{Result: "ok"}, nil
	}

	middleware := []Middleware{}
	wrappedHandler := ApplyMiddleware(handler, middleware)

	ctx := context.Background()
	req := &Request{Method: "test"}

	_, err := wrappedHandler(ctx, req)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if !called {
		t.Error("expected handler to be called")
	}
}
