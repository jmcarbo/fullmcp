package client

import (
	"context"
	"encoding/json"

	"github.com/jmcarbo/fullmcp/mcp"
)

// SamplingHandler is a function that handles sampling requests from servers
type SamplingHandler func(ctx context.Context, req *mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error)

// WithSamplingHandler configures a sampling handler for the client
func WithSamplingHandler(handler SamplingHandler) Option {
	return func(c *Client) {
		c.samplingHandler = handler
	}
}

// TODO: Wire up handleSamplingRequest in the client message routing
// handleSamplingRequest processes a sampling/createMessage request from the server
//
//nolint:unused // Reserved for future server-initiated sampling requests
func (c *Client) _handleSamplingRequest(ctx context.Context, params json.RawMessage) (*mcp.CreateMessageResult, error) {
	if c.samplingHandler == nil {
		return nil, &mcp.Error{
			Code:    mcp.MethodNotFound,
			Message: "sampling not supported by this client",
		}
	}

	var req mcp.CreateMessageRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, &mcp.Error{
			Code:    mcp.InvalidParams,
			Message: "invalid sampling request parameters",
		}
	}

	return c.samplingHandler(ctx, &req)
}
