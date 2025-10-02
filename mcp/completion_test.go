package mcp

import (
	"encoding/json"
	"testing"
)

func TestCompleteRequest_Serialization(t *testing.T) {
	req := CompleteRequest{
		Ref: CompletionRef{
			Type: "ref/prompt",
			Name: "code_review",
		},
		Argument: CompletionArgument{
			Name:  "language",
			Value: "Go",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CompleteRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Ref.Type != "ref/prompt" {
		t.Errorf("ref type mismatch")
	}

	if decoded.Ref.Name != "code_review" {
		t.Errorf("ref name mismatch")
	}

	if decoded.Argument.Name != "language" {
		t.Errorf("argument name mismatch")
	}
}

func TestCompleteResult_Serialization(t *testing.T) {
	result := CompleteResult{}
	result.Completion.Values = []string{"Go", "Python", "JavaScript"}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CompleteResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(decoded.Completion.Values) != 3 {
		t.Errorf("values count mismatch: got %d, want 3", len(decoded.Completion.Values))
	}

	if decoded.Completion.Values[0] != "Go" {
		t.Errorf("first value mismatch")
	}
}

func TestCompletionRef_ResourceType(t *testing.T) {
	ref := CompletionRef{
		Type: "ref/resource",
		Name: "file:///path/to/resource",
	}

	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CompletionRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Type != "ref/resource" {
		t.Errorf("type mismatch")
	}
}

func TestCompletionValue_RichCompletion(t *testing.T) {
	value := CompletionValue{
		Value:  "Python",
		Label:  "Python 3.11",
		Detail: "Latest Python version",
		Data: map[string]interface{}{
			"version": "3.11.0",
		},
	}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CompletionValue
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Value != "Python" {
		t.Errorf("value mismatch")
	}

	if decoded.Label != "Python 3.11" {
		t.Errorf("label mismatch")
	}
}
