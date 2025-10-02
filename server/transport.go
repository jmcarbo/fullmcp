package server

import (
	"io"
	"os"
)

// StdioTransport implements stdio transport
type StdioTransport struct {
	stdin  io.Reader
	stdout io.Writer
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		stdin:  os.Stdin,
		stdout: os.Stdout,
	}
}

func (t *StdioTransport) Read(p []byte) (int, error) {
	return t.stdin.Read(p)
}

func (t *StdioTransport) Write(p []byte) (int, error) {
	return t.stdout.Write(p)
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	return nil
}
