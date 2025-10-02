// Package sse provides Server-Sent Events transport for MCP.
package sse

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Transport implements SSE transport for MCP client
type Transport struct {
	url    string
	client *http.Client
}

// Option configures the SSE transport
type Option func(*Transport)

// New creates a new SSE transport
func New(url string, opts ...Option) *Transport {
	t := &Transport{
		url:    url,
		client: &http.Client{},
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(t *Transport) {
		t.client = client
	}
}

// Connect establishes an SSE connection
func (t *Transport) Connect(ctx context.Context) (io.ReadWriteCloser, error) {
	return &sseConn{
		url:    t.url,
		client: t.client,
		ctx:    ctx,
		buf:    &bytes.Buffer{},
		closed: make(chan struct{}),
	}, nil
}

// Close closes the transport
func (t *Transport) Close() error {
	return nil
}

// sseConn implements a connection over SSE
type sseConn struct {
	url    string
	client *http.Client
	ctx    context.Context
	buf    *bytes.Buffer
	mu     sync.Mutex
	resp   *http.Response
	reader *bufio.Reader
	closed chan struct{}
	once   sync.Once
}

// Read reads from the SSE stream
func (c *sseConn) Read(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If we have data in buffer, read from it
	if c.buf.Len() > 0 {
		return c.buf.Read(p)
	}

	// If no reader yet, establish SSE connection
	if c.reader == nil {
		if err := c.connect(); err != nil {
			return 0, err
		}
	}

	// Read next SSE event
	for {
		select {
		case <-c.closed:
			return 0, io.EOF
		default:
		}

		line, err := c.reader.ReadBytes('\n')
		if err != nil {
			return 0, err
		}

		// Parse SSE format: "data: <json>\n\n"
		if bytes.HasPrefix(line, []byte("data: ")) {
			data := bytes.TrimPrefix(line, []byte("data: "))
			data = bytes.TrimSuffix(data, []byte("\n"))

			// Check for end marker
			if bytes.HasSuffix(data, []byte("\n")) {
				c.buf.Write(data)
				return c.buf.Read(p)
			}

			c.buf.Write(data)
		} else if len(bytes.TrimSpace(line)) == 0 {
			// Empty line marks end of event
			if c.buf.Len() > 0 {
				return c.buf.Read(p)
			}
		}
	}
}

// Write sends data via POST and reads response via SSE
func (c *sseConn) Write(p []byte) (int, error) {
	req, err := http.NewRequestWithContext(c.ctx, "POST", c.url, bytes.NewReader(p))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return len(p), nil
}

// Close closes the connection
func (c *sseConn) Close() error {
	c.once.Do(func() {
		close(c.closed)
		if c.resp != nil {
			_ = c.resp.Body.Close()
		}
	})
	return nil
}

// connect establishes the SSE stream
func (c *sseConn) connect() error {
	req, err := http.NewRequestWithContext(c.ctx, "GET", c.url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	c.resp = resp
	c.reader = bufio.NewReader(resp.Body)
	return nil
}

// Server provides SSE server support for MCP
type Server struct {
	handler Handler
	addr    string
}

// Handler processes MCP requests and streams responses
type Handler interface {
	HandleSSE(ctx context.Context, req []byte) (<-chan []byte, error)
}

// NewServer creates a new SSE server for MCP
func NewServer(addr string, handler Handler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
	}
}

// ListenAndServe starts the SSE server
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleSSE)
	return http.ListenAndServe(s.addr, mux)
}

// handleSSE handles SSE requests
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// For POST requests, read body and process
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request", http.StatusBadRequest)
			return
		}
		defer func() { _ = r.Body.Close() }()

		respChan, err := s.handler.HandleSSE(r.Context(), body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Stream responses
		for data := range respChan {
			_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
		return
	}

	// For GET requests, keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// MCPSSEHandler adapts a simple handler function to Handler
type MCPSSEHandler struct {
	handleFunc func(context.Context, []byte) ([]byte, error)
}

// NewMCPSSEHandler creates an SSE handler for MCP
func NewMCPSSEHandler(handleFunc func(context.Context, []byte) ([]byte, error)) *MCPSSEHandler {
	return &MCPSSEHandler{
		handleFunc: handleFunc,
	}
}

// HandleSSE implements Handler
func (h *MCPSSEHandler) HandleSSE(ctx context.Context, req []byte) (<-chan []byte, error) {
	respChan := make(chan []byte, 1)

	go func() {
		defer close(respChan)

		resp, err := h.handleFunc(ctx, req)
		if err != nil {
			// Send error as SSE event
			respChan <- []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error()))
			return
		}

		respChan <- resp
	}()

	return respChan, nil
}
