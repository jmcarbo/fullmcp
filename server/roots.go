package server

import (
	"context"

	"github.com/jmcarbo/fullmcp/mcp"
)

// RootsHandler is called when the server receives a roots/list_changed notification
type RootsHandler func(ctx context.Context)

// WithRootsHandler configures a handler for roots change notifications
func WithRootsHandler(handler RootsHandler) Option {
	return func(s *Server) {
		s.rootsHandler = handler
	}
}

// ListRoots requests the list of roots from the client
// Note: This requires bidirectional communication with the client
func (s *Server) ListRoots(_ context.Context) ([]mcp.Root, error) {
	// In a real implementation, this would send a request to the connected client
	// For now, return an error indicating this needs to be implemented in the transport layer
	return nil, &mcp.Error{
		Code:    mcp.InternalError,
		Message: "roots/list requests require bidirectional communication with client",
	}
}
