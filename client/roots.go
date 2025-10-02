package client

import (
	"context"

	"github.com/jmcarbo/fullmcp/mcp"
)

// RootsProvider is a function that returns the list of roots for the client
type RootsProvider func(ctx context.Context) ([]mcp.Root, error)

// WithRoots configures the client to support roots capability
func WithRoots(provider RootsProvider) Option {
	return func(c *Client) {
		c.rootsProvider = provider
	}
}

// ListRoots is called by servers to request the list of roots
// This is typically handled automatically by the message handler
func (c *Client) handleRootsList(ctx context.Context) (*mcp.RootsListResult, error) {
	if c.rootsProvider == nil {
		return nil, &mcp.Error{
			Code:    mcp.MethodNotFound,
			Message: "roots not supported by this client",
		}
	}

	roots, err := c.rootsProvider(ctx)
	if err != nil {
		return nil, err
	}

	return &mcp.RootsListResult{
		Roots: roots,
	}, nil
}

// NotifyRootsChanged sends a notification to the server that the roots list has changed
func (c *Client) NotifyRootsChanged() error {
	if c.rootsProvider == nil {
		return &mcp.Error{
			Code:    mcp.InternalError,
			Message: "roots not configured for this client",
		}
	}
	return c.notify("notifications/roots/list_changed", nil)
}
