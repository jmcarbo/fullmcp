package server

import (
	"bytes"
	"testing"
)

func TestStdioTransport_New(t *testing.T) {
	transport := NewStdioTransport()

	if transport == nil {
		t.Fatal("expected non-nil transport")
	}

	if transport.stdin == nil {
		t.Error("expected stdin to be set")
	}

	if transport.stdout == nil {
		t.Error("expected stdout to be set")
	}
}

func TestStdioTransport_ReadWrite(t *testing.T) {
	var input bytes.Buffer
	var output bytes.Buffer

	input.WriteString("test input")

	transport := &StdioTransport{
		stdin:  &input,
		stdout: &output,
	}

	// Test Read
	buf := make([]byte, 10)
	n, err := transport.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if n != 10 {
		t.Errorf("expected to read 10 bytes, got %d", n)
	}
	if string(buf) != "test input" {
		t.Errorf("expected 'test input', got '%s'", string(buf))
	}

	// Test Write
	data := []byte("test output")
	n, err = transport.Write(data)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected to write %d bytes, got %d", len(data), n)
	}
	if output.String() != "test output" {
		t.Errorf("expected 'test output', got '%s'", output.String())
	}
}

func TestStdioTransport_Close(t *testing.T) {
	transport := NewStdioTransport()

	err := transport.Close()
	if err != nil {
		t.Fatalf("close failed: %v", err)
	}
}
