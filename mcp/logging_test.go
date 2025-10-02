package mcp

import (
	"encoding/json"
	"testing"
)

func TestLogLevel_Value(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected int
	}{
		{LogLevelDebug, 0},
		{LogLevelInfo, 1},
		{LogLevelNotice, 2},
		{LogLevelWarning, 3},
		{LogLevelError, 4},
		{LogLevelCritical, 5},
		{LogLevelAlert, 6},
		{LogLevelEmergency, 7},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			if got := tt.level.Value(); got != tt.expected {
				t.Errorf("Value() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestLogLevel_ShouldLog(t *testing.T) {
	tests := []struct {
		level    LogLevel
		minLevel LogLevel
		expected bool
	}{
		{LogLevelDebug, LogLevelInfo, false},
		{LogLevelInfo, LogLevelInfo, true},
		{LogLevelWarning, LogLevelInfo, true},
		{LogLevelError, LogLevelWarning, true},
		{LogLevelDebug, LogLevelDebug, true},
		{LogLevelEmergency, LogLevelError, true},
		{LogLevelNotice, LogLevelWarning, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.level)+"_vs_"+string(tt.minLevel), func(t *testing.T) {
			if got := tt.level.ShouldLog(tt.minLevel); got != tt.expected {
				t.Errorf("ShouldLog() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSetLevelRequest_Serialization(t *testing.T) {
	req := SetLevelRequest{
		Level: LogLevelInfo,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded SetLevelRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Level != LogLevelInfo {
		t.Errorf("level mismatch: got %s, want %s", decoded.Level, LogLevelInfo)
	}
}

func TestLogMessage_Serialization(t *testing.T) {
	msg := LogMessage{
		Level:  LogLevelError,
		Logger: "database",
		Data: map[string]interface{}{
			"error":   "Connection failed",
			"details": map[string]interface{}{"host": "localhost", "port": 5432},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded LogMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Level != LogLevelError {
		t.Errorf("level mismatch")
	}

	if decoded.Logger != "database" {
		t.Errorf("logger mismatch")
	}

	if decoded.Data["error"] != "Connection failed" {
		t.Errorf("data mismatch")
	}
}

func TestLogMessage_WithoutLogger(t *testing.T) {
	msg := LogMessage{
		Level: LogLevelInfo,
		Data:  map[string]interface{}{"message": "Hello"},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded LogMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Logger != "" {
		t.Errorf("expected empty logger, got %s", decoded.Logger)
	}
}
