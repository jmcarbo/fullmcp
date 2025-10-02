package client

import (
	"github.com/jmcarbo/fullmcp/mcp"
)

// CancelRequest sends a cancellation notification for a request
func (c *Client) CancelRequest(requestID interface{}, reason string) error {
	notification := mcp.CancelledNotification{
		RequestID: requestID,
		Reason:    reason,
	}

	return c.notify("notifications/cancelled", notification)
}
