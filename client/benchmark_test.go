package client

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

// mockTransport for benchmarking
type benchTransport struct {
	readCh  chan []byte
	writeCh chan []byte
	closed  chan struct{}
	mu      sync.Mutex
}

func newBenchTransportPair() (*benchTransport, *benchTransport) {
	ch1 := make(chan []byte, 1000)
	ch2 := make(chan []byte, 1000)

	client := &benchTransport{
		readCh:  ch2,
		writeCh: ch1,
		closed:  make(chan struct{}),
	}

	server := &benchTransport{
		readCh:  ch1,
		writeCh: ch2,
		closed:  make(chan struct{}),
	}

	return client, server
}

func (m *benchTransport) Read(p []byte) (int, error) {
	select {
	case <-m.closed:
		return 0, io.EOF
	case data := <-m.readCh:
		n := copy(p, data)
		if n < len(data) {
			go func() { m.readCh <- data[n:] }()
		}
		return n, nil
	}
}

func (m *benchTransport) Write(p []byte) (int, error) {
	select {
	case <-m.closed:
		return 0, io.EOF
	case m.writeCh <- append([]byte(nil), p...):
		return len(p), nil
	}
}

func (m *benchTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	select {
	case <-m.closed:
	default:
		close(m.closed)
	}
	return nil
}

// mockServer responds to client requests
func mockServerResponder(conn *benchTransport) {
	for {
		var msg mcp.Message
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		if err := json.Unmarshal(buf[:n], &msg); err != nil {
			continue
		}

		// Send back a simple response
		response := &mcp.Message{
			JSONRPC: "2.0",
			ID:      msg.ID,
			Result:  json.RawMessage(`{"tools":[],"resources":[],"prompts":[]}`),
		}

		respBytes, _ := json.Marshal(response)
		respBytes = append(respBytes, '\n')
		_, _ = conn.Write(respBytes)
	}
}

// BenchmarkClientConnect measures connection performance
func BenchmarkClientConnect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		clientConn, serverConn := newBenchTransportPair()
		go mockServerResponder(serverConn)
		c := New(clientConn)
		b.StartTimer()

		ctx := context.Background()
		_ = c.Connect(ctx)

		b.StopTimer()
		_ = c.Close()
		_ = clientConn.Close()
		_ = serverConn.Close()
	}
}

// BenchmarkClientListTools measures tool listing performance
func BenchmarkClientListTools(b *testing.B) {
	clientConn, serverConn := newBenchTransportPair()
	go mockServerResponder(serverConn)

	c := New(clientConn)
	ctx := context.Background()
	_ = c.Connect(ctx)
	defer func() { _ = c.Close() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.ListTools(ctx)
	}
}

// BenchmarkClientCallTool measures tool call performance
func BenchmarkClientCallTool(b *testing.B) {
	clientConn, serverConn := newBenchTransportPair()

	// Start mock server that responds to tool calls
	go func() {
		for {
			var msg mcp.Message
			buf := make([]byte, 4096)
			n, err := serverConn.Read(buf)
			if err != nil {
				return
			}

			if err := json.Unmarshal(buf[:n], &msg); err != nil {
				continue
			}

			var response *mcp.Message
			if msg.Method == "tools/call" {
				response = &mcp.Message{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Result:  json.RawMessage(`{"content":[{"type":"text","text":"result"}]}`),
				}
			} else {
				response = &mcp.Message{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Result:  json.RawMessage(`{"tools":[]}`),
				}
			}

			respBytes, _ := json.Marshal(response)
			respBytes = append(respBytes, '\n')
			_, _ = serverConn.Write(respBytes)
		}
	}()

	c := New(clientConn)
	ctx := context.Background()
	_ = c.Connect(ctx)
	defer func() { _ = c.Close() }()

	args := json.RawMessage(`{"a":5,"b":3}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.CallTool(ctx, "add", args)
	}
}

// BenchmarkClientReadResource measures resource read performance
func BenchmarkClientReadResource(b *testing.B) {
	clientConn, serverConn := newBenchTransportPair()

	// Start mock server
	go func() {
		for {
			var msg mcp.Message
			buf := make([]byte, 4096)
			n, err := serverConn.Read(buf)
			if err != nil {
				return
			}

			if err := json.Unmarshal(buf[:n], &msg); err != nil {
				continue
			}

			var response *mcp.Message
			if msg.Method == "resources/read" {
				response = &mcp.Message{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Result:  json.RawMessage(`{"contents":[{"uri":"config://test","mimeType":"text/plain","text":"data"}]}`),
				}
			} else {
				response = &mcp.Message{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Result:  json.RawMessage(`{"resources":[]}`),
				}
			}

			respBytes, _ := json.Marshal(response)
			respBytes = append(respBytes, '\n')
			_, _ = serverConn.Write(respBytes)
		}
	}()

	c := New(clientConn)
	ctx := context.Background()
	_ = c.Connect(ctx)
	defer func() { _ = c.Close() }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.ReadResource(ctx, "config://test")
	}
}

// BenchmarkClientConcurrentCalls measures concurrent request handling
func BenchmarkClientConcurrentCalls(b *testing.B) {
	clientConn, serverConn := newBenchTransportPair()
	go mockServerResponder(serverConn)

	c := New(clientConn)
	ctx := context.Background()
	_ = c.Connect(ctx)
	defer func() { _ = c.Close() }()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = c.ListTools(ctx)
		}
	})
}
