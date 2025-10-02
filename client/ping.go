package client

import (
	"context"
)

// Ping sends a ping request to the server to verify the connection is alive
func (c *Client) Ping(ctx context.Context) error {
	return c.call(ctx, "ping", nil, nil)
}
