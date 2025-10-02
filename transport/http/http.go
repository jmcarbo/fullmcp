// Package http provides HTTP transport for MCP (Model Context Protocol).
package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// Transport implements HTTP transport for MCP
type Transport struct {
	url     string
	client  *http.Client
	headers map[string]string
}

// Option configures the HTTP transport
type Option func(*Transport)

// New creates a new HTTP transport
func New(url string, opts ...Option) *Transport {
	t := &Transport{
		url:     url,
		client:  &http.Client{},
		headers: make(map[string]string),
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

// WithHeaders sets custom HTTP headers
func WithHeaders(headers map[string]string) Option {
	return func(t *Transport) {
		for k, v := range headers {
			t.headers[k] = v
		}
	}
}

// WithAPIKey sets the X-API-Key header
func WithAPIKey(apiKey string) Option {
	return func(t *Transport) {
		t.headers["X-API-Key"] = apiKey
	}
}

// Connect establishes an HTTP connection
func (t *Transport) Connect(ctx context.Context) (io.ReadWriteCloser, error) {
	return &httpConn{
		url:     t.url,
		client:  t.client,
		ctx:     ctx,
		headers: t.headers,
	}, nil
}

// Close closes the transport (no-op for HTTP)
func (t *Transport) Close() error {
	return nil
}

// httpConn implements a pseudo-connection over HTTP
type httpConn struct {
	url       string
	client    *http.Client
	ctx       context.Context
	buf       bytes.Buffer
	mu        sync.Mutex
	writeMu   sync.Mutex  // Serializes concurrent Write operations
	dataCond  *sync.Cond
	hasData   bool
	closed    bool
	sessionID string
	headers   map[string]string
}

// Read reads from the response buffer, blocking until data is available
func (c *httpConn) Read(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Wait for data to be available
	for !c.hasData && !c.closed {
		if c.dataCond == nil {
			c.dataCond = sync.NewCond(&c.mu)
		}
		c.dataCond.Wait()
	}

	if c.closed && c.buf.Len() == 0 {
		return 0, io.EOF
	}

	n, err := c.buf.Read(p)
	if c.buf.Len() == 0 {
		c.hasData = false
	}
	return n, err
}

// Write sends an HTTP POST request and stores the response
func (c *httpConn) Write(p []byte) (int, error) {
	// Serialize Write operations to prevent concurrent buffer overwrites
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	req, err := http.NewRequestWithContext(c.ctx, "POST", c.url, bytes.NewReader(p))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// Add session ID if we have one
	c.mu.Lock()
	sessionID := c.sessionID
	c.mu.Unlock()
	if sessionID != "" {
		req.Header.Set("mcp-session-id", sessionID)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Capture session ID from response if present
	if respSessionID := resp.Header.Get("mcp-session-id"); respSessionID != "" {
		c.mu.Lock()
		c.sessionID = respSessionID
		c.mu.Unlock()
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Read response into buffer (only if status is 200 OK)
	if resp.StatusCode == http.StatusOK {
		c.mu.Lock()
		c.buf.Reset()
		_, err = io.Copy(&c.buf, resp.Body)
		if err != nil {
			c.mu.Unlock()
			return 0, err
		}

		// Signal that data is available
		c.hasData = true
		if c.dataCond != nil {
			c.dataCond.Signal()
		}
		c.mu.Unlock()
	}
	// For 202 Accepted (notifications), no response body expected

	return len(p), nil
}

// Close closes the connection
func (c *httpConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	if c.dataCond != nil {
		c.dataCond.Broadcast()
	}
	return nil
}

// Server provides HTTP server support for MCP
type Server struct {
	handler http.Handler
	addr    string
}

// NewServer creates a new HTTP server for MCP
func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
	}
}

// ListenAndServe starts the HTTP server
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.handler)
}

// MCPHandler implements http.Handler for MCP
type MCPHandler struct {
	handleFunc func(context.Context, []byte) ([]byte, error)
}

// NewMCPHandler creates an HTTP handler for MCP
func NewMCPHandler(handleFunc func(context.Context, []byte) ([]byte, error)) *MCPHandler {
	return &MCPHandler{
		handleFunc: handleFunc,
	}
}

// ServeHTTP implements http.Handler
func (h *MCPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	response, err := h.handleFunc(r.Context(), body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(response)
}
