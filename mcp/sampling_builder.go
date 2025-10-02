package mcp

// Helper functions for building sampling requests

// AddMessage adds a message to the request
func (r *CreateMessageRequest) AddMessage(role, text string) *CreateMessageRequest {
	r.Messages = append(r.Messages, SamplingMessage{
		Role: role,
		Content: SamplingContent{
			Type: "text",
			Text: text,
		},
	})
	return r
}

// AddUserMessage adds a user message
func (r *CreateMessageRequest) AddUserMessage(text string) *CreateMessageRequest {
	return r.AddMessage("user", text)
}

// AddAssistantMessage adds an assistant message
func (r *CreateMessageRequest) AddAssistantMessage(text string) *CreateMessageRequest {
	return r.AddMessage("assistant", text)
}

// WithSystemPrompt adds a system prompt to the request
func (r *CreateMessageRequest) WithSystemPrompt(prompt string) *CreateMessageRequest {
	r.SystemPrompt = prompt
	return r
}

// WithMaxTokens sets the maximum tokens for the response
func (r *CreateMessageRequest) WithMaxTokens(tokens int) *CreateMessageRequest {
	r.MaxTokens = &tokens
	return r
}

// WithTemperature sets the sampling temperature
func (r *CreateMessageRequest) WithTemperature(temp float64) *CreateMessageRequest {
	r.Temperature = &temp
	return r
}

// WithModelPreferences sets model selection preferences
func (r *CreateMessageRequest) WithModelPreferences(prefs *ModelPreferences) *CreateMessageRequest {
	r.ModelPreferences = prefs
	return r
}

// WithIntelligencePriority sets intelligence priority (0-1)
func (p *ModelPreferences) WithIntelligencePriority(priority float64) *ModelPreferences {
	p.IntelligencePriority = &priority
	return p
}

// WithSpeedPriority sets speed priority (0-1)
func (p *ModelPreferences) WithSpeedPriority(priority float64) *ModelPreferences {
	p.SpeedPriority = &priority
	return p
}
