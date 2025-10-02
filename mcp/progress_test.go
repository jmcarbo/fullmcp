package mcp

import (
	"encoding/json"
	"testing"
)

func TestProgressNotification_Serialization(t *testing.T) {
	total := 100.0
	notif := ProgressNotification{
		ProgressToken: "task-123",
		Progress:      50.0,
		Total:         &total,
	}

	data, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ProgressNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ProgressToken != "task-123" {
		t.Errorf("token mismatch: got %v, want task-123", decoded.ProgressToken)
	}

	if decoded.Progress != 50.0 {
		t.Errorf("progress mismatch: got %f, want 50.0", decoded.Progress)
	}

	if decoded.Total == nil || *decoded.Total != 100.0 {
		t.Errorf("total mismatch")
	}
}

func TestProgressNotification_WithoutTotal(t *testing.T) {
	notif := ProgressNotification{
		ProgressToken: "task-456",
		Progress:      25.0,
	}

	data, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ProgressNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Total != nil {
		t.Errorf("expected nil total, got %v", *decoded.Total)
	}
}

func TestProgressNotification_IntegerToken(t *testing.T) {
	notif := ProgressNotification{
		ProgressToken: 12345,
		Progress:      75.5,
	}

	data, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ProgressNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// JSON unmarshaling converts numbers to float64
	tokenFloat, ok := decoded.ProgressToken.(float64)
	if !ok {
		t.Errorf("expected float64 token, got %T", decoded.ProgressToken)
	}

	if tokenFloat != 12345 {
		t.Errorf("token mismatch: got %f, want 12345", tokenFloat)
	}
}

func TestProgressNotification_FloatValues(t *testing.T) {
	total := 100.5
	notif := ProgressNotification{
		ProgressToken: "float-test",
		Progress:      33.33,
		Total:         &total,
	}

	data, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ProgressNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Progress != 33.33 {
		t.Errorf("progress mismatch: got %f, want 33.33", decoded.Progress)
	}

	if *decoded.Total != 100.5 {
		t.Errorf("total mismatch: got %f, want 100.5", *decoded.Total)
	}
}

func TestRequestMeta_Serialization(t *testing.T) {
	meta := RequestMeta{
		ProgressToken: "meta-token",
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RequestMeta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ProgressToken != "meta-token" {
		t.Errorf("token mismatch")
	}
}
