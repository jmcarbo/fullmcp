package client

import (
	"context"

	"github.com/jmcarbo/fullmcp/mcp"
)

// LogHandler is called when the client receives a log message notification
type LogHandler func(ctx context.Context, msg *mcp.LogMessage)

// WithLogHandler configures a handler for log message notifications
func WithLogHandler(handler LogHandler) Option {
	return func(c *Client) {
		c.logHandler = handler
	}
}

// SetLogLevel sends a logging/setLevel request to the server
func (c *Client) SetLogLevel(ctx context.Context, level mcp.LogLevel) error {
	params := mcp.SetLevelRequest{
		Level: level,
	}

	return c.call(ctx, "logging/setLevel", params, nil)
}
