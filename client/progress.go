package client

import (
	"context"

	"github.com/jmcarbo/fullmcp/mcp"
)

// ProgressHandler is called when the client receives a progress notification
type ProgressHandler func(ctx context.Context, notification *mcp.ProgressNotification)

// WithProgressHandler configures a handler for progress notifications
func WithProgressHandler(handler ProgressHandler) Option {
	return func(c *Client) {
		c.progressHandler = handler
	}
}
