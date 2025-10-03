package mcp

import "encoding/json"

// Content represents MCP content blocks
type Content interface {
	ContentType() string
}

// TextContent represents text content
type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ContentType returns the content type
func (t TextContent) ContentType() string {
	return t.Type
}

// ImageContent represents image content
type ImageContent struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	MimeType string `json:"mimeType"`
}

// ContentType returns the content type
func (i ImageContent) ContentType() string {
	return i.Type
}

// AudioContent represents audio content (2025-03-26)
type AudioContent struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	MimeType string `json:"mimeType"`
}

// ContentType returns the content type
func (a AudioContent) ContentType() string {
	return a.Type
}

// ResourceContent represents resource content
type ResourceContent struct {
	Type     string `json:"type"`
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
}

// ContentType returns the content type
func (r ResourceContent) ContentType() string {
	return r.Type
}

// ResourceLinkContent represents a resource link in tool results (2025-06-18)
type ResourceLinkContent struct {
	Type        string                 `json:"type"` // "resource"
	Resource    Resource               `json:"resource"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

// ContentType returns the content type
func (rl ResourceLinkContent) ContentType() string {
	return rl.Type
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	// 2025-06-18: Tool output schemas
	OutputSchema map[string]interface{} `json:"outputSchema,omitempty"` // JSON Schema for tool output
	// 2025-03-26 annotations
	Title           string `json:"title,omitempty"`           // Human-readable title
	ReadOnlyHint    *bool  `json:"readOnlyHint,omitempty"`    // Tool doesn't modify environment
	DestructiveHint *bool  `json:"destructiveHint,omitempty"` // Tool may perform destructive updates
	IdempotentHint  *bool  `json:"idempotentHint,omitempty"`  // Repeated calls have no additional effect
	OpenWorldHint   *bool  `json:"openWorldHint,omitempty"`   // Tool may interact with external entities
}

// Resource represents an MCP resource
type Resource struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Title       string                 `json:"title,omitempty"` // Human-readable title (2025-06-18)
	Description string                 `json:"description,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Meta        map[string]interface{} `json:"_meta,omitempty"` // Metadata (2025-06-18)
}

// ResourceTemplate for parameterized resources
type ResourceTemplate struct {
	URITemplate string                 `json:"uriTemplate"`
	Name        string                 `json:"name"`
	Title       string                 `json:"title,omitempty"` // Human-readable title (2025-06-18)
	Description string                 `json:"description,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Meta        map[string]interface{} `json:"_meta,omitempty"` // Metadata (2025-06-18)
}

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string                 `json:"name"`
	Title       string                 `json:"title,omitempty"` // Human-readable title (2025-06-18)
	Description string                 `json:"description,omitempty"`
	Arguments   []PromptArgument       `json:"arguments,omitempty"`
	Meta        map[string]interface{} `json:"_meta,omitempty"` // Metadata (2025-06-18)
}

// PromptArgument represents a prompt argument
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// PromptMessage for prompt responses
type PromptMessage struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

// UnmarshalJSON implements custom JSON unmarshaling for PromptMessage
func (pm *PromptMessage) UnmarshalJSON(data []byte) error {
	// Create temporary struct with raw content
	var temp struct {
		Role    string            `json:"role"`
		Content []json.RawMessage `json:"content"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	pm.Role = temp.Role
	pm.Content = make([]Content, 0, len(temp.Content))

	// Unmarshal each content item based on its type field
	for _, rawContent := range temp.Content {
		// First, peek at the type field
		var typeCheck struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(rawContent, &typeCheck); err != nil {
			return err
		}

		// Create appropriate concrete type based on type field
		var content Content
		switch typeCheck.Type {
		case "text":
			var tc TextContent
			if err := json.Unmarshal(rawContent, &tc); err != nil {
				return err
			}
			content = tc
		case "image":
			var ic ImageContent
			if err := json.Unmarshal(rawContent, &ic); err != nil {
				return err
			}
			content = ic
		case "audio":
			var ac AudioContent
			if err := json.Unmarshal(rawContent, &ac); err != nil {
				return err
			}
			content = ac
		case "resource":
			// Could be ResourceContent or ResourceLinkContent
			// Check if it has "resource" field to distinguish
			var check map[string]interface{}
			if err := json.Unmarshal(rawContent, &check); err != nil {
				return err
			}
			if _, hasResource := check["resource"]; hasResource {
				var rlc ResourceLinkContent
				if err := json.Unmarshal(rawContent, &rlc); err != nil {
					return err
				}
				content = rlc
			} else {
				var rc ResourceContent
				if err := json.Unmarshal(rawContent, &rc); err != nil {
					return err
				}
				content = rc
			}
		default:
			// Unknown type - use TextContent as fallback
			var tc TextContent
			if err := json.Unmarshal(rawContent, &tc); err != nil {
				return err
			}
			content = tc
		}

		pm.Content = append(pm.Content, content)
	}

	return nil
}

// ServerCapabilities represents server capabilities
type ServerCapabilities struct {
	Tools       *ToolsCapability       `json:"tools,omitempty"`
	Resources   *ResourcesCapability   `json:"resources,omitempty"`
	Prompts     *PromptsCapability     `json:"prompts,omitempty"`
	Completions *CompletionsCapability `json:"completions,omitempty"` // 2025-03-26
}

// ToolsCapability represents tools capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resources capability
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability represents prompts capability
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// CompletionsCapability represents completions capability (2025-03-26)
type CompletionsCapability struct {
	// Empty struct indicating completion support
}

// ClientCapabilities represents client capabilities
type ClientCapabilities struct {
	Roots       *RootsCapability       `json:"roots,omitempty"` // 2025-06-18
	Sampling    *SamplingCapability    `json:"sampling,omitempty"`
	Elicitation *ElicitationCapability `json:"elicitation,omitempty"` // 2025-06-18
}

// SamplingCapability represents sampling capability
type SamplingCapability struct {
	// Empty struct indicating sampling support
}

// ElicitationCapability represents elicitation capability (2025-06-18)
type ElicitationCapability struct {
	// Empty struct indicating elicitation support
}

// ElicitationRequest represents an elicitation request from server (2025-06-18)
type ElicitationRequest struct {
	Schema      map[string]interface{} `json:"schema"`                // JSON Schema for requested data
	Description string                 `json:"description,omitempty"` // What the data is for
}

// ElicitationResponse represents user's response to elicitation (2025-06-18)
type ElicitationResponse struct {
	Action string                 `json:"action"`         // "accept", "decline", or "cancel"
	Data   map[string]interface{} `json:"data,omitempty"` // User-provided data (if accepted)
}

// Message represents a JSON-RPC 2.0 message envelope
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
