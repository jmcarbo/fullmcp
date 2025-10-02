package mcp

// CancelledNotification represents a request cancellation notification
type CancelledNotification struct {
	RequestID interface{} `json:"requestId"` // ID of the request to cancel (string or number)
	Reason    string      `json:"reason,omitempty"` // Optional reason for cancellation
}
