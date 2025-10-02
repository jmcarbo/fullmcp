// Package mcp defines core types and interfaces for the Model Context Protocol.
package mcp

import "fmt"

// ErrorCode represents JSON-RPC error codes
type ErrorCode int

// Standard JSON-RPC error codes
const (
	ParseError     ErrorCode = -32700
	InvalidRequest ErrorCode = -32600
	MethodNotFound ErrorCode = -32601
	InvalidParams  ErrorCode = -32602
	InternalError  ErrorCode = -32603
)

// Error represents an MCP protocol error
type Error struct {
	Code    ErrorCode
	Message string
	Data    interface{}
}

func (e *Error) Error() string {
	return fmt.Sprintf("MCP error %d: %s", e.Code, e.Message)
}

// NotFoundError represents a not found error
type NotFoundError struct {
	Type string // "tool", "resource", "prompt"
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Type, e.Name)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}
