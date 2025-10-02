package server

import (
	"context"

	"github.com/jmcarbo/fullmcp/mcp"
)

// SamplingCapability represents the server's ability to request sampling from clients
type SamplingCapability struct {
	enabled bool
}

// EnableSampling returns an option that enables sampling capability
func EnableSampling() Option {
	return func(s *Server) {
		s.sampling = &SamplingCapability{enabled: true}
	}
}

// CreateMessage requests the client to create a message via LLM sampling
// This allows servers to leverage client-side LLM capabilities
func (s *Server) CreateMessage(_ context.Context, _ *mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error) {
	if s.sampling == nil || !s.sampling.enabled {
		return nil, &mcp.Error{
			Code:    mcp.MethodNotFound,
			Message: "sampling not enabled on this server",
		}
	}

	// In a real implementation, this would send a request to the connected client
	// For now, return an error indicating this needs to be implemented in the transport layer
	return nil, &mcp.Error{
		Code:    mcp.InternalError,
		Message: "sampling requests require bidirectional communication with client",
	}
}

// Helper functions for building sampling requests

// NewSamplingRequest creates a new sampling request
func NewSamplingRequest() *mcp.CreateMessageRequest {
	return &mcp.CreateMessageRequest{
		Messages: []mcp.SamplingMessage{},
	}
}

// NewModelPreferences creates model preferences with hints
func NewModelPreferences(modelHints ...string) *mcp.ModelPreferences {
	hints := make([]mcp.ModelHint, len(modelHints))
	for i, hint := range modelHints {
		hints[i] = mcp.ModelHint{Name: hint}
	}
	return &mcp.ModelPreferences{
		Hints: hints,
	}
}
