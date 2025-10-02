package mcp

import (
	"encoding/json"
	"testing"
)

func TestSamplingTypes(t *testing.T) {
	// Test CreateMessageRequest serialization
	req := &CreateMessageRequest{
		Messages: []SamplingMessage{
			{
				Role: "user",
				Content: SamplingContent{
					Type: "text",
					Text: "Hello, world!",
				},
			},
		},
		SystemPrompt: "You are a helpful assistant",
	}

	maxTokens := 100
	req.MaxTokens = &maxTokens

	temp := 0.7
	req.Temperature = &temp

	req.ModelPreferences = &ModelPreferences{
		Hints: []ModelHint{
			{Name: "claude-3-sonnet"},
		},
	}

	intelligencePriority := 0.8
	speedPriority := 0.5
	req.ModelPreferences.IntelligencePriority = &intelligencePriority
	req.ModelPreferences.SpeedPriority = &speedPriority

	// Serialize
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	// Deserialize
	var decoded CreateMessageRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	// Verify
	if decoded.SystemPrompt != "You are a helpful assistant" {
		t.Errorf("system prompt mismatch")
	}

	if *decoded.MaxTokens != 100 {
		t.Errorf("maxTokens mismatch: got %d, want 100", *decoded.MaxTokens)
	}

	if *decoded.Temperature != 0.7 {
		t.Errorf("temperature mismatch: got %f, want 0.7", *decoded.Temperature)
	}

	if len(decoded.Messages) != 1 {
		t.Errorf("messages count mismatch: got %d, want 1", len(decoded.Messages))
	}

	if decoded.Messages[0].Content.Text != "Hello, world!" {
		t.Errorf("message text mismatch")
	}
}

func TestCreateMessageResult(t *testing.T) {
	result := &CreateMessageResult{
		Role: "assistant",
		Content: SamplingContent{
			Type: "text",
			Text: "The capital of France is Paris.",
		},
		Model:      "claude-3-sonnet-20240307",
		StopReason: StopReasonEndTurn,
	}

	// Serialize
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal result: %v", err)
	}

	// Deserialize
	var decoded CreateMessageResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Verify
	if decoded.Role != "assistant" {
		t.Errorf("role mismatch")
	}

	if decoded.Content.Text != "The capital of France is Paris." {
		t.Errorf("content text mismatch")
	}

	if decoded.Model != "claude-3-sonnet-20240307" {
		t.Errorf("model mismatch")
	}

	if decoded.StopReason != StopReasonEndTurn {
		t.Errorf("stop reason mismatch")
	}
}

func TestStopReasons(t *testing.T) {
	// Test all stop reason constants exist
	reasons := []string{
		StopReasonEndTurn,
		StopReasonStopSequence,
		StopReasonMaxTokens,
		StopReasonError,
	}

	for _, reason := range reasons {
		if reason == "" {
			t.Errorf("stop reason constant is empty")
		}
	}
}
