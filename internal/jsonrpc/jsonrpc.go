// Package jsonrpc provides JSON-RPC 2.0 message reading and writing utilities.
package jsonrpc

import (
	"encoding/json"
	"io"

	"github.com/jmcarbo/fullmcp/mcp"
)

// MessageReader reads JSON-RPC messages
type MessageReader struct {
	decoder *json.Decoder
}

// NewMessageReader creates a new message reader
func NewMessageReader(r io.Reader) *MessageReader {
	return &MessageReader{
		decoder: json.NewDecoder(r),
	}
}

// Read reads a message
func (mr *MessageReader) Read() (*mcp.Message, error) {
	var msg mcp.Message
	if err := mr.decoder.Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// MessageWriter writes JSON-RPC messages
type MessageWriter struct {
	encoder *json.Encoder
}

// NewMessageWriter creates a new message writer
func NewMessageWriter(w io.Writer) *MessageWriter {
	return &MessageWriter{
		encoder: json.NewEncoder(w),
	}
}

// Write writes a message
func (mw *MessageWriter) Write(msg *mcp.Message) error {
	return mw.encoder.Encode(msg)
}
