// Package stdio provides standard input/output transport for MCP.
package stdio

import (
	"io"
	"os"
)

// Transport implements stdio transport
type Transport struct {
	stdin  io.Reader
	stdout io.Writer
}

// New creates a stdio transport
func New() *Transport {
	return &Transport{
		stdin:  os.Stdin,
		stdout: os.Stdout,
	}
}

// Read implements io.Reader
func (t *Transport) Read(p []byte) (int, error) {
	return t.stdin.Read(p)
}

// Write implements io.Writer
func (t *Transport) Write(p []byte) (int, error) {
	return t.stdout.Write(p)
}

// Close implements io.Closer
func (t *Transport) Close() error {
	return nil
}
