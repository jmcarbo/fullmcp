package mcp

// ProgressToken is a unique identifier for tracking progress
// Can be string or integer
type ProgressToken interface{}

// ProgressNotification represents a progress update notification
type ProgressNotification struct {
	ProgressToken ProgressToken `json:"progressToken"`   // Unique token for this operation
	Progress      float64       `json:"progress"`        // Current progress value
	Total         *float64      `json:"total,omitempty"` // Optional total value
	Message       string        `json:"message,omitempty"` // Optional descriptive status (2025-03-26)
}

// RequestMeta contains metadata for requests, including progress tracking
type RequestMeta struct {
	ProgressToken ProgressToken `json:"progressToken,omitempty"`
}
