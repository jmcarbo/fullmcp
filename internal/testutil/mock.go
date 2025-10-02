// Package testutil provides testing utilities for MCP implementation.
package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"

	"github.com/jmcarbo/fullmcp/mcp"
)

// MockTransport provides a mock transport for testing
type MockTransport struct {
	ReadBuffer  *bytes.Buffer
	WriteBuffer *bytes.Buffer
	Closed      bool
	mu          sync.Mutex
}

// NewMockTransport creates a new mock transport
func NewMockTransport() *MockTransport {
	return &MockTransport{
		ReadBuffer:  &bytes.Buffer{},
		WriteBuffer: &bytes.Buffer{},
	}
}

// Read implements io.Reader
func (m *MockTransport) Read(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ReadBuffer.Read(p)
}

// Write implements io.Writer
func (m *MockTransport) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.WriteBuffer.Write(p)
}

// Close implements io.Closer
func (m *MockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Closed = true
	return nil
}

// WriteMessage writes a message to the read buffer (for client to read)
func (m *MockTransport) WriteMessage(msg *mcp.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return json.NewEncoder(m.ReadBuffer).Encode(msg)
}

// ReadMessage reads a message from the write buffer (what client wrote)
func (m *MockTransport) ReadMessage() (*mcp.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var msg mcp.Message
	if err := json.NewDecoder(m.WriteBuffer).Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// PipeTransport creates a pair of transports connected via pipes
type PipeTransport struct {
	reader *io.PipeReader
	writer *io.PipeWriter
}

// NewPipeTransport creates a new pipe transport pair
func NewPipeTransport() (*PipeTransport, *PipeTransport) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	t1 := &PipeTransport{reader: r1, writer: w2}
	t2 := &PipeTransport{reader: r2, writer: w1}

	return t1, t2
}

func (p *PipeTransport) Read(b []byte) (int, error) {
	return p.reader.Read(b)
}

func (p *PipeTransport) Write(b []byte) (int, error) {
	return p.writer.Write(b)
}

// Close closes the pipe transport
func (p *PipeTransport) Close() error {
	_ = p.reader.Close()
	_ = p.writer.Close()
	return nil
}
