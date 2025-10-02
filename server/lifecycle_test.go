package server

import (
	"context"
	"testing"
)

func TestLifespanFunc_Integration(t *testing.T) {
	cleanupCalled := false
	contextValue := "test-value"

	lifespanFunc := func(ctx context.Context, s *Server) (context.Context, func(), error) {
		// Add value to context
		ctx = context.WithValue(ctx, "test-key", contextValue)

		cleanup := func() {
			cleanupCalled = true
		}

		return ctx, cleanup, nil
	}

	srv := New("test-server", WithLifespan(lifespanFunc))

	if srv.lifespan == nil {
		t.Fatal("expected lifespan function to be set")
	}

	// Simulate calling the lifespan function
	ctx := context.Background()
	newCtx, cleanup, err := srv.lifespan(ctx, srv)
	if err != nil {
		t.Fatalf("lifespan function failed: %v", err)
	}

	// Verify context was modified
	value := newCtx.Value("test-key")
	if value != contextValue {
		t.Errorf("expected context value '%s', got '%v'", contextValue, value)
	}

	// Verify cleanup wasn't called yet
	if cleanupCalled {
		t.Error("cleanup should not be called yet")
	}

	// Call cleanup
	if cleanup != nil {
		cleanup()
	}

	// Verify cleanup was called
	if !cleanupCalled {
		t.Error("expected cleanup to be called")
	}
}

func TestLifespanFunc_Error(t *testing.T) {
	expectedErr := &ErrorContext{Message: "initialization failed"}

	lifespanFunc := func(ctx context.Context, s *Server) (context.Context, func(), error) {
		return nil, nil, expectedErr
	}

	srv := New("test-server", WithLifespan(lifespanFunc))

	ctx := context.Background()
	_, _, err := srv.lifespan(ctx, srv)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestLifespanFunc_NilCleanup(t *testing.T) {
	lifespanFunc := func(ctx context.Context, s *Server) (context.Context, func(), error) {
		return ctx, nil, nil
	}

	srv := New("test-server", WithLifespan(lifespanFunc))

	ctx := context.Background()
	_, cleanup, err := srv.lifespan(ctx, srv)
	if err != nil {
		t.Fatalf("lifespan function failed: %v", err)
	}

	// Verify nil cleanup doesn't cause issues
	if cleanup != nil {
		t.Error("expected nil cleanup")
	}
}
