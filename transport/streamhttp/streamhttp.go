// Package streamhttp provides Streamable HTTP transport for MCP (2025-03-26 specification).
// This transport replaces the HTTP+SSE transport from 2024-11-05 with improved bi-directional
// streaming support using POST for client-to-server and GET for server-to-client via SSE.
package streamhttp

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Transport implements Streamable HTTP transport for MCP
type Transport struct {
	url         string
	client      *http.Client
	sessionID   string
	sseReader   *sseReader
	sseReady    chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	eventIDLock sync.Mutex
	lastEventID string
	headers     map[string]string
}

// Option configures the Streamable HTTP transport
type Option func(*Transport)

// New creates a new Streamable HTTP transport
func New(url string, opts ...Option) *Transport {
	ctx, cancel := context.WithCancel(context.Background())
	t := &Transport{
		url:      url,
		client:   &http.Client{},
		ctx:      ctx,
		cancel:   cancel,
		headers:  make(map[string]string),
		sseReady: make(chan struct{}),
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

// WithSessionID sets the session ID
func WithSessionID(sessionID string) Option {
	return func(t *Transport) {
		t.sessionID = sessionID
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

// Connect establishes a Streamable HTTP connection
func (t *Transport) Connect(_ context.Context) (io.ReadWriteCloser, error) {
	conn := &streamConn{
		transport: t,
		readBuf:   &bytes.Buffer{},
	}
	conn.bufferCond = sync.NewCond(&conn.mu)

	// Open SSE stream in background to avoid blocking
	// This allows the client to send POST requests before the SSE stream is ready
	go func() {
		reader, err := t.openSSEStream()
		if err != nil {
			// SSE connection failed, but we still allow POST requests
			close(t.sseReady)
			return
		}

		t.mu.Lock()
		t.sseReader = reader
		t.mu.Unlock()
		close(t.sseReady)
	}()

	return conn, nil
}

// Close closes the transport
func (t *Transport) Close() error {
	t.cancel()
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sseReader != nil {
		return t.sseReader.Close()
	}
	return nil
}

// openSSEStream opens the SSE stream for server-to-client messages
func (t *Transport) openSSEStream() (*sseReader, error) {
	req, err := http.NewRequestWithContext(t.ctx, "GET", t.url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	// Add custom headers
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	// Include session ID if present
	if t.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", t.sessionID)
	}

	// Include Last-Event-ID for resumption
	t.eventIDLock.Lock()
	if t.lastEventID != "" {
		req.Header.Set("Last-Event-ID", t.lastEventID)
	}
	t.eventIDLock.Unlock()

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("SSE connection failed: %d", resp.StatusCode)
	}

	return &sseReader{
		resp:      resp,
		scanner:   bufio.NewScanner(resp.Body),
		transport: t,
	}, nil
}

// post sends a POST request to the server
func (t *Transport) post(data []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(t.ctx, "POST", t.url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	// Add custom headers
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	// Include session ID if present
	if t.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", t.sessionID)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Check for session ID in response (during initialization)
	if sessionID := resp.Header.Get("Mcp-Session-Id"); sessionID != "" && t.sessionID == "" {
		t.mu.Lock()
		t.sessionID = sessionID
		t.mu.Unlock()
	}

	// 202 Accepted means notification/response (no body expected)
	if resp.StatusCode == http.StatusAccepted {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// streamConn implements a connection over Streamable HTTP
type streamConn struct {
	transport  *Transport
	readBuf    *bytes.Buffer
	mu         sync.Mutex
	bufferCond *sync.Cond
}

// tryReadBuffer attempts to read from the buffer if it has data
func (c *streamConn) tryReadBuffer(p []byte) (n int, hasData bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.readBuf.Len() > 0 {
		n, err = c.readBuf.Read(p)
		return n, true, err
	}
	return 0, false, nil
}

// waitForSSEReady waits for the SSE stream to be ready
func (c *streamConn) waitForSSEReady(p []byte) (n int, done bool, err error) {
	select {
	case <-c.transport.sseReady:
		return 0, false, nil
	default:
		if n, hasData, err := c.tryReadBuffer(p); hasData {
			return n, true, err
		}
		<-c.transport.sseReady
		return 0, false, nil
	}
}

// readFromSSEStream reads data from the SSE stream
func (c *streamConn) readFromSSEStream(p []byte, reader *sseReader) (int, error) {
	sseDataChan := make(chan []byte, 1)
	sseErrChan := make(chan error, 1)

	go func() {
		data, err := reader.ReadEvent()
		if err != nil {
			sseErrChan <- err
			return
		}
		sseDataChan <- data
	}()

	c.mu.Lock()
	for c.readBuf.Len() == 0 {
		select {
		case data := <-sseDataChan:
			if data != nil {
				c.readBuf.Write(data)
				c.readBuf.WriteByte('\n')
			}
		case err := <-sseErrChan:
			c.mu.Unlock()
			return 0, err
		default:
			c.bufferCond.Wait()
		}
	}

	n, err := c.readBuf.Read(p)
	c.mu.Unlock()
	return n, err
}

// Read reads from the SSE stream
func (c *streamConn) Read(p []byte) (int, error) {
	if n, hasData, err := c.tryReadBuffer(p); hasData {
		return n, err
	}

	if n, done, err := c.waitForSSEReady(p); done {
		return n, err
	}

	if n, hasData, err := c.tryReadBuffer(p); hasData {
		return n, err
	}

	c.transport.mu.Lock()
	reader := c.transport.sseReader
	c.transport.mu.Unlock()

	if reader == nil {
		if n, hasData, err := c.tryReadBuffer(p); hasData {
			return n, err
		}
		return 0, nil
	}

	return c.readFromSSEStream(p, reader)
}

// Write sends a POST request with the data
func (c *streamConn) Write(p []byte) (int, error) {
	response, err := c.transport.post(p)
	if err != nil {
		return 0, err
	}

	// If there's a response, buffer it for Read to consume
	if response != nil {
		c.mu.Lock()
		c.readBuf.Write(response)
		// json.Decoder expects newline-delimited JSON
		c.readBuf.WriteByte('\n')
		c.bufferCond.Signal()
		c.mu.Unlock()
	}

	return len(p), nil
}

// Close closes the connection
func (c *streamConn) Close() error {
	return c.transport.Close()
}

// sseReader reads Server-Sent Events
type sseReader struct {
	resp      *http.Response
	scanner   *bufio.Scanner
	transport *Transport
	mu        sync.Mutex
}

// parseSSEField parses a SSE field:value line
func parseSSEField(line string) (field, value string, ok bool) {
	if len(line) == 0 || line[0] == ':' {
		return "", "", false
	}

	colonIdx := 0
	for i, ch := range line {
		if ch == ':' {
			colonIdx = i
			break
		}
	}

	if colonIdx == 0 {
		return "", "", false
	}

	field = line[:colonIdx]
	if colonIdx+1 < len(line) {
		value = line[colonIdx+1:]
		if len(value) > 0 && value[0] == ' ' {
			value = value[1:]
		}
	}
	return field, value, true
}

// updateEventID updates the transport's last event ID if needed
func (r *sseReader) updateEventID(eventID string) {
	if eventID != "" {
		r.transport.eventIDLock.Lock()
		r.transport.lastEventID = eventID
		r.transport.eventIDLock.Unlock()
	}
}

// processSSELine processes a single SSE line and updates data/eventID
func processSSELine(line string, data *[]byte, eventID *string) {
	field, value, ok := parseSSEField(line)
	if !ok {
		return
	}

	switch field {
	case "data":
		*data = append(*data, []byte(value)...)
		*data = append(*data, '\n')
	case "id":
		*eventID = value
	}
}

// ReadEvent reads the next SSE event
func (r *sseReader) ReadEvent() ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var data []byte
	var eventID string

	for r.scanner.Scan() {
		line := r.scanner.Text()

		if line == "" {
			if len(data) > 0 {
				r.updateEventID(eventID)
				return data, nil
			}
			continue
		}

		processSSELine(line, &data, &eventID)
	}

	if err := r.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, io.EOF
}

// Close closes the SSE reader
func (r *sseReader) Close() error {
	if r.resp != nil {
		return r.resp.Body.Close()
	}
	return nil
}

// Server provides Streamable HTTP server support for MCP
type Server struct {
	handler       http.Handler
	addr          string
	sessionStore  *SessionStore
	allowedOrigin string
}

// ServerOption configures the Streamable HTTP server
type ServerOption func(*Server)

// NewServer creates a new Streamable HTTP server for MCP
func NewServer(addr string, handler http.Handler, opts ...ServerOption) *Server {
	s := &Server{
		addr:         addr,
		handler:      handler,
		sessionStore: NewSessionStore(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithAllowedOrigin sets the allowed origin for CORS
func WithAllowedOrigin(origin string) ServerOption {
	return func(s *Server) {
		s.allowedOrigin = origin
	}
}

// matchOrigin checks if an origin matches the allowed pattern (supports wildcards)
func matchOrigin(origin, pattern string) bool {
	if pattern == "*" {
		return true
	}

	if pattern == origin {
		return true
	}

	// Handle wildcard patterns like "*.example.com" or "https://*.example.com"
	if !hasWildcard(pattern) {
		return false
	}

	// Split pattern into prefix and suffix around the wildcard
	parts := splitWildcard(pattern)
	if len(parts) != 2 {
		return false
	}

	prefix, suffix := parts[0], parts[1]

	// Check if origin starts with prefix and ends with suffix
	if len(origin) < len(prefix)+len(suffix) {
		return false
	}

	return hasPrefix(origin, prefix) && hasSuffix(origin, suffix)
}

// hasWildcard checks if a pattern contains a wildcard
func hasWildcard(pattern string) bool {
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '*' {
			return true
		}
	}
	return false
}

// splitWildcard splits a pattern on the first wildcard
func splitWildcard(pattern string) []string {
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '*' {
			return []string{pattern[:i], pattern[i+1:]}
		}
	}
	return []string{pattern}
}

// hasPrefix checks if s starts with prefix
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// hasSuffix checks if s ends with suffix
func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// setCORSHeaders sets CORS headers on the response
func (s *Server) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}

	// Determine allowed origin to return
	allowedOrigin := "*"
	if s.allowedOrigin != "" {
		// If we have a specific pattern and origin matches, echo it back
		if matchOrigin(origin, s.allowedOrigin) {
			allowedOrigin = origin
		} else {
			// Don't set CORS headers for non-matching origins
			return
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// ListenAndServe starts the Streamable HTTP server
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s)
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Validate origin for security
	if s.allowedOrigin != "" {
		origin := r.Header.Get("Origin")
		if origin != "" && !matchOrigin(origin, s.allowedOrigin) {
			http.Error(w, "forbidden origin", http.StatusForbidden)
			return
		}
	}

	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		allowedOrigin := "*"
		if s.allowedOrigin != "" {
			allowedOrigin = s.allowedOrigin
		}
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id, X-API-Key, Authorization, Last-Event-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Set CORS headers for actual requests
	s.setCORSHeaders(w, r)

	switch r.Method {
	case http.MethodPost:
		s.handlePOST(w, r)
	case http.MethodGet:
		s.handleGET(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePOST handles POST requests (client-to-server messages)
func (s *Server) handlePOST(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("Mcp-Session-Id")

	// Create or get session
	session := s.sessionStore.GetOrCreate(sessionID)

	// If this is initialization and no session ID, generate one
	if sessionID == "" && session.ID == "" {
		session.ID = generateSessionID()
		w.Header().Set("Mcp-Session-Id", session.ID)
	}

	// Delegate to the wrapped handler (which includes auth and MCP processing)
	if s.handler != nil {
		s.handler.ServeHTTP(w, r)
	}
}

// handleGET handles GET requests (server-to-client SSE stream)
func (s *Server) handleGET(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("Mcp-Session-Id")

	// Get or create session (allow initial connection without session ID)
	session := s.sessionStore.GetOrCreate(sessionID)

	// Generate session ID if not present
	if session.ID == "" {
		session.ID = generateSessionID()
		sessionID = session.ID
		s.sessionStore.Store(sessionID, session)
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial newline to establish connection
	_, _ = fmt.Fprintf(w, ":\n\n")
	flusher.Flush()

	// Check for Last-Event-ID for resumption
	lastEventID := r.Header.Get("Last-Event-ID")
	if lastEventID != "" {
		// TODO: Replay missed events from lastEventID
		_ = lastEventID
	}

	// Stream events from session
	session.mu.Lock()
	session.sseWriter = w
	session.sseFlusher = flusher
	session.mu.Unlock()

	// Keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			// Send keep-alive comment
			_, _ = fmt.Fprintf(w, ": keep-alive\n\n")
			flusher.Flush()
		}
	}
}

// SessionStore manages sessions
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionStore creates a new session store
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

// Get retrieves a session
func (ss *SessionStore) Get(id string) *Session {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.sessions[id]
}

// GetOrCreate retrieves or creates a session
func (ss *SessionStore) GetOrCreate(id string) *Session {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if id != "" {
		if session, exists := ss.sessions[id]; exists {
			return session
		}
	}

	session := &Session{
		ID:        id,
		CreatedAt: time.Now(),
	}

	if id != "" {
		ss.sessions[id] = session
	}

	return session
}

// Store saves a session
func (ss *SessionStore) Store(id string, session *Session) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.sessions[id] = session
}

// Delete removes a session
func (ss *SessionStore) Delete(id string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.sessions, id)
}

// Session represents a client session
type Session struct {
	ID         string
	CreatedAt  time.Time
	mu         sync.Mutex
	sseWriter  http.ResponseWriter
	sseFlusher http.Flusher
}

// SendEvent sends an SSE event to the client
func (s *Session) SendEvent(data []byte, eventID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sseWriter == nil {
		return fmt.Errorf("no SSE connection")
	}

	if eventID != "" {
		_, _ = fmt.Fprintf(s.sseWriter, "id: %s\n", eventID)
	}
	_, _ = fmt.Fprintf(s.sseWriter, "data: %s\n\n", data)
	s.sseFlusher.Flush()

	return nil
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
