package jsonrpc

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

func TestMessageReader_Read(t *testing.T) {
	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      123,
		Method:  "test/method",
		Params:  json.RawMessage(`{"key":"value"}`),
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(msg); err != nil {
		t.Fatalf("failed to encode message: %v", err)
	}

	reader := NewMessageReader(&buf)
	readMsg, err := reader.Read()
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	if readMsg.Method != msg.Method {
		t.Errorf("expected method '%s', got '%s'", msg.Method, readMsg.Method)
	}

	if string(readMsg.Params) != string(msg.Params) {
		t.Errorf("expected params '%s', got '%s'", string(msg.Params), string(readMsg.Params))
	}
}

func TestMessageReader_Read_EOF(t *testing.T) {
	var buf bytes.Buffer
	reader := NewMessageReader(&buf)

	_, err := reader.Read()
	if err != io.EOF {
		t.Errorf("expected EOF error, got %v", err)
	}
}

func TestMessageReader_Read_InvalidJSON(t *testing.T) {
	buf := bytes.NewBufferString("{invalid json}")
	reader := NewMessageReader(buf)

	_, err := reader.Read()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMessageWriter_Write(t *testing.T) {
	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      456,
		Method:  "test/method",
		Params:  json.RawMessage(`{"test":true}`),
	}

	var buf bytes.Buffer
	writer := NewMessageWriter(&buf)

	if err := writer.Write(msg); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	// Read back and verify
	var readMsg mcp.Message
	decoder := json.NewDecoder(&buf)
	if err := decoder.Decode(&readMsg); err != nil {
		t.Fatalf("failed to decode message: %v", err)
	}

	if readMsg.Method != msg.Method {
		t.Errorf("expected method '%s', got '%s'", msg.Method, readMsg.Method)
	}

	idFloat, ok := readMsg.ID.(float64)
	if !ok {
		t.Fatalf("expected ID to be float64, got %T", readMsg.ID)
	}

	if int(idFloat) != 456 {
		t.Errorf("expected ID 456, got %d", int(idFloat))
	}
}

func TestMessageWriter_Write_WithError(t *testing.T) {
	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      789,
		Error: &mcp.RPCError{
			Code:    -32600,
			Message: "Invalid Request",
		},
	}

	var buf bytes.Buffer
	writer := NewMessageWriter(&buf)

	if err := writer.Write(msg); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	// Read back and verify
	var readMsg mcp.Message
	decoder := json.NewDecoder(&buf)
	if err := decoder.Decode(&readMsg); err != nil {
		t.Fatalf("failed to decode message: %v", err)
	}

	if readMsg.Error == nil {
		t.Fatal("expected error to be set")
	}

	if readMsg.Error.Code != msg.Error.Code {
		t.Errorf("expected error code %d, got %d", msg.Error.Code, readMsg.Error.Code)
	}
}

func TestMessageReader_Write_RoundTrip(t *testing.T) {
	messages := []*mcp.Message{
		{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "initialize",
			Params:  json.RawMessage(`{"version":"1.0"}`),
		},
		{
			JSONRPC: "2.0",
			ID:      2,
			Result:  json.RawMessage(`{"status":"ok"}`),
		},
		{
			JSONRPC: "2.0",
			ID:      3,
			Error: &mcp.RPCError{
				Code:    -32601,
				Message: "Method not found",
			},
		},
	}

	var buf bytes.Buffer
	writer := NewMessageWriter(&buf)

	// Write all messages
	for _, msg := range messages {
		if err := writer.Write(msg); err != nil {
			t.Fatalf("failed to write message: %v", err)
		}
	}

	// Read all messages back
	reader := NewMessageReader(&buf)
	for i, originalMsg := range messages {
		readMsg, err := reader.Read()
		if err != nil {
			t.Fatalf("failed to read message %d: %v", i, err)
		}

		if readMsg.Method != originalMsg.Method {
			t.Errorf("message %d: expected method '%s', got '%s'", i, originalMsg.Method, readMsg.Method)
		}
	}
}
