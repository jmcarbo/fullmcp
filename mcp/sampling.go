// Package mcp provides sampling types for LLM completions
package mcp

// SamplingMessage represents a message in a sampling request
type SamplingMessage struct {
	Role    string          `json:"role"`    // "user" or "assistant"
	Content SamplingContent `json:"content"` // Message content
}

// SamplingContent represents the content of a sampling message
type SamplingContent struct {
	Type string `json:"type"`           // "text" or "image"
	Text string `json:"text,omitempty"` // For text content
	Data string `json:"data,omitempty"` // For image content (base64)
}

// ModelPreferences specifies preferences for model selection
type ModelPreferences struct {
	Hints                []ModelHint `json:"hints,omitempty"`                // Suggested models
	IntelligencePriority *float64    `json:"intelligencePriority,omitempty"` // 0-1, higher = prefer more capable models
	SpeedPriority        *float64    `json:"speedPriority,omitempty"`        // 0-1, higher = prefer faster models
}

// ModelHint suggests a preferred model
type ModelHint struct {
	Name string `json:"name"` // Model name hint (e.g., "claude-3-sonnet")
}

// CreateMessageRequest represents a request to create a message via sampling
type CreateMessageRequest struct {
	Messages         []SamplingMessage `json:"messages"`                   // Conversation messages
	ModelPreferences *ModelPreferences `json:"modelPreferences,omitempty"` // Model selection preferences
	SystemPrompt     string            `json:"systemPrompt,omitempty"`     // System prompt to guide model behavior
	MaxTokens        *int              `json:"maxTokens,omitempty"`        // Maximum tokens in response
	Temperature      *float64          `json:"temperature,omitempty"`      // Sampling temperature
	StopSequences    []string          `json:"stopSequences,omitempty"`    // Stop generation at these sequences
	Metadata         map[string]string `json:"metadata,omitempty"`         // Additional metadata
}

// CreateMessageResult represents the result of a sampling request
type CreateMessageResult struct {
	Role       string          `json:"role"`                 // "assistant"
	Content    SamplingContent `json:"content"`              // Generated content
	Model      string          `json:"model"`                // Actual model used
	StopReason string          `json:"stopReason,omitempty"` // Reason for stopping (e.g., "endTurn", "stopSequence", "maxTokens")
}

// Common stop reasons
const (
	StopReasonEndTurn      = "endTurn"      // Natural end of model's turn
	StopReasonStopSequence = "stopSequence" // Hit a stop sequence
	StopReasonMaxTokens    = "maxTokens"    // Reached max tokens limit
	StopReasonError        = "error"        // Error occurred
)
