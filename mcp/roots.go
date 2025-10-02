package mcp

// Root represents a filesystem boundary or URI scope that a server should operate within
type Root struct {
	URI  string `json:"uri"`            // URI of the root (e.g., "file:///path" or "https://example.com")
	Name string `json:"name,omitempty"` // Optional human-readable name for the root
}

// RootsListResult represents the response to a roots/list request
type RootsListResult struct {
	Roots []Root `json:"roots"`
}

// RootsCapability indicates whether the client supports roots and notifications
type RootsCapability struct {
	ListChanged bool `json:"listChanged"` // Whether client will emit notifications when roots change
}
