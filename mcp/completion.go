package mcp

// CompletionRef represents a reference to what is being completed
type CompletionRef struct {
	Type string `json:"type"` // "ref/prompt" or "ref/resource"
	Name string `json:"name"` // Name of the prompt or resource
}

// CompletionArgument represents the argument being completed
type CompletionArgument struct {
	Name  string `json:"name"`            // Argument name
	Value string `json:"value,omitempty"` // Partial value typed so far
}

// CompleteRequest represents a request for completion suggestions
type CompleteRequest struct {
	Ref      CompletionRef      `json:"ref"`      // What is being completed
	Argument CompletionArgument `json:"argument"` // Argument being completed
}

// CompletionValue represents a single completion suggestion
type CompletionValue struct {
	Value  string                 `json:"value"`            // Suggested value
	Label  string                 `json:"label,omitempty"`  // Optional display label
	Detail string                 `json:"detail,omitempty"` // Optional additional detail
	Data   map[string]interface{} `json:"data,omitempty"`   // Optional metadata
}

// CompleteResult represents the response with completion suggestions
type CompleteResult struct {
	Completion struct {
		Values      []string          `json:"values"`                // List of suggested values
		Total       *int              `json:"total,omitempty"`       // Total available (if paginated)
		HasMore     *bool             `json:"hasMore,omitempty"`     // More results available
		Completions []CompletionValue `json:"completions,omitempty"` // Rich completions
	} `json:"completion"`
}
