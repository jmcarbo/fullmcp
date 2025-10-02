package stdio

import (
	"bytes"
	"io"
	"testing"
)

func TestTransport_New(t *testing.T) {
	transport := New()

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

func TestTransport_Read(t *testing.T) {
	// Create a custom transport with controlled input
	var input bytes.Buffer
	input.WriteString("test data")

	transport := &Transport{
		stdin:  &input,
		stdout: &bytes.Buffer{},
	}

	buf := make([]byte, 9)
	n, err := transport.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if n != 9 {
		t.Errorf("expected to read 9 bytes, got %d", n)
	}

	if string(buf) != "test data" {
		t.Errorf("expected 'test data', got '%s'", string(buf))
	}
}

func TestTransport_Write(t *testing.T) {
	var output bytes.Buffer

	transport := &Transport{
		stdin:  &bytes.Buffer{},
		stdout: &output,
	}

	data := []byte("output data")
	n, err := transport.Write(data)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if n != len(data) {
		t.Errorf("expected to write %d bytes, got %d", len(data), n)
	}

	if output.String() != "output data" {
		t.Errorf("expected 'output data', got '%s'", output.String())
	}
}

func TestTransport_Close(t *testing.T) {
	transport := New()

	err := transport.Close()
	if err != nil {
		t.Fatalf("close failed: %v", err)
	}
}

func TestTransport_ReadEOF(t *testing.T) {
	var input bytes.Buffer // Empty buffer

	transport := &Transport{
		stdin:  &input,
		stdout: &bytes.Buffer{},
	}

	buf := make([]byte, 10)
	_, err := transport.Read(buf)
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
}

func TestTransport_MultipleReads(t *testing.T) {
	var input bytes.Buffer
	input.WriteString("first\nsecond\nthird\n")

	transport := &Transport{
		stdin:  &input,
		stdout: &bytes.Buffer{},
	}

	buf1 := make([]byte, 6)
	n1, err := transport.Read(buf1)
	if err != nil {
		t.Fatalf("first read failed: %v", err)
	}
	if string(buf1[:n1]) != "first\n" {
		t.Errorf("expected 'first\\n', got '%s'", string(buf1[:n1]))
	}

	buf2 := make([]byte, 7)
	n2, err := transport.Read(buf2)
	if err != nil {
		t.Fatalf("second read failed: %v", err)
	}
	if string(buf2[:n2]) != "second\n" {
		t.Errorf("expected 'second\\n', got '%s'", string(buf2[:n2]))
	}
}

func TestTransport_MultipleWrites(t *testing.T) {
	var output bytes.Buffer

	transport := &Transport{
		stdin:  &bytes.Buffer{},
		stdout: &output,
	}

	writes := []string{"first\n", "second\n", "third\n"}

	for _, data := range writes {
		_, err := transport.Write([]byte(data))
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}

	expected := "first\nsecond\nthird\n"
	if output.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, output.String())
	}
}
