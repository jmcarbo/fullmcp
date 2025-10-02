package mcp

import (
	"encoding/json"
	"testing"
)

func TestCancelledNotification_Serialization(t *testing.T) {
	notif := CancelledNotification{
		RequestID: "request-123",
		Reason:    "User requested cancellation",
	}

	data, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CancelledNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.RequestID != "request-123" {
		t.Errorf("requestId mismatch: got %v, want request-123", decoded.RequestID)
	}

	if decoded.Reason != "User requested cancellation" {
		t.Errorf("reason mismatch: got %s", decoded.Reason)
	}
}

func TestCancelledNotification_IntegerRequestID(t *testing.T) {
	notif := CancelledNotification{
		RequestID: 42,
		Reason:    "Timeout",
	}

	data, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CancelledNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// JSON unmarshaling converts numbers to float64
	requestIDFloat, ok := decoded.RequestID.(float64)
	if !ok {
		t.Errorf("expected float64 requestID, got %T", decoded.RequestID)
	}

	if requestIDFloat != 42 {
		t.Errorf("requestID mismatch: got %f, want 42", requestIDFloat)
	}
}

func TestCancelledNotification_WithoutReason(t *testing.T) {
	notif := CancelledNotification{
		RequestID: "request-456",
	}

	data, err := json.Marshal(notif)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CancelledNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Reason != "" {
		t.Errorf("expected empty reason, got %s", decoded.Reason)
	}
}
