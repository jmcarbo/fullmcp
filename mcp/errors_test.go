package mcp

import (
	"errors"
	"testing"
)

func TestMCPError_Error(t *testing.T) {
	err := &Error{
		Code:    ParseError,
		Message: "parse failed",
		Data:    "extra info",
	}

	expected := "MCP error -32700: parse failed"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestNotFoundError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *NotFoundError
		expected string
	}{
		{
			name:     "tool not found",
			err:      &NotFoundError{Type: "tool", Name: "add"},
			expected: "tool not found: add",
		},
		{
			name:     "resource not found",
			err:      &NotFoundError{Type: "resource", Name: "config://app"},
			expected: "resource not found: config://app",
		},
		{
			name:     "prompt not found",
			err:      &NotFoundError{Type: "prompt", Name: "greeting"},
			expected: "prompt not found: greeting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "name",
		Message: "is required",
	}

	expected := "validation error on name: is required"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestErrorCode_Constants(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		expected int
	}{
		{"ParseError", ParseError, -32700},
		{"InvalidRequest", InvalidRequest, -32600},
		{"MethodNotFound", MethodNotFound, -32601},
		{"InvalidParams", InvalidParams, -32602},
		{"InternalError", InternalError, -32603},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.code) != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, int(tt.code))
			}
		})
	}
}

func TestNotFoundError_AsError(t *testing.T) {
	var err error = &NotFoundError{Type: "tool", Name: "test"}

	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Error("expected errors.As to succeed")
	}

	if notFound.Type != "tool" {
		t.Errorf("expected type 'tool', got '%s'", notFound.Type)
	}
}

func TestValidationError_AsError(t *testing.T) {
	var err error = &ValidationError{Field: "age", Message: "must be positive"}

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Error("expected errors.As to succeed")
	}

	if validationErr.Field != "age" {
		t.Errorf("expected field 'age', got '%s'", validationErr.Field)
	}
}
