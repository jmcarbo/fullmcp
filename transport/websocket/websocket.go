// Package websocket provides WebSocket transport for MCP (Model Context Protocol).
package websocket

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Transport implements WebSocket transport for MCP
type Transport struct {
	url      string
	dialer   *websocket.Dialer
	headers  http.Header
	conn     *websocket.Conn
	connMu   sync.RWMutex
	readBuf  []byte
	readMu   sync.Mutex
	writeMu  sync.Mutex
}

// Option configures the WebSocket transport
type Option func(*Transport)

// New creates a new WebSocket transport
func New(url string, opts ...Option) *Transport {
	t := &Transport{
		url:     url,
		dialer:  websocket.DefaultDialer,
		headers: http.Header{},
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// WithDialer sets a custom WebSocket dialer
func WithDialer(dialer *websocket.Dialer) Option {
	return func(t *Transport) {
		t.dialer = dialer
	}
}

// WithHeaders sets custom headers for the WebSocket handshake
func WithHeaders(headers http.Header) Option {
	return func(t *Transport) {
		t.headers = headers
	}
}

// Connect establishes a WebSocket connection
func (t *Transport) Connect(ctx context.Context) (io.ReadWriteCloser, error) {
	conn, _, err := t.dialer.DialContext(ctx, t.url, t.headers)
	if err != nil {
		return nil, fmt.Errorf("websocket dial failed: %w", err)
	}

	t.connMu.Lock()
	t.conn = conn
	t.connMu.Unlock()

	return &wsConn{
		conn:    conn,
		readBuf: &t.readBuf,
		readMu:  &t.readMu,
		writeMu: &t.writeMu,
	}, nil
}

// Close closes the transport
func (t *Transport) Close() error {
	t.connMu.RLock()
	conn := t.conn
	t.connMu.RUnlock()

	if conn != nil {
		return conn.Close()
	}
	return nil
}

// wsConn wraps a WebSocket connection to implement io.ReadWriteCloser
type wsConn struct {
	conn    *websocket.Conn
	readBuf *[]byte
	readMu  *sync.Mutex
	writeMu *sync.Mutex
}

// Read reads from the WebSocket connection
func (c *wsConn) Read(p []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	// If we have buffered data, read from it first
	if len(*c.readBuf) > 0 {
		n := copy(p, *c.readBuf)
		*c.readBuf = (*c.readBuf)[n:]
		return n, nil
	}

	// Read next message
	messageType, data, err := c.conn.ReadMessage()
	if err != nil {
		return 0, err
	}

	if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
		return 0, fmt.Errorf("unexpected message type: %d", messageType)
	}

	// Copy what fits into p, buffer the rest
	n := copy(p, data)
	if n < len(data) {
		*c.readBuf = data[n:]
	}

	return n, nil
}

// Write writes to the WebSocket connection
func (c *wsConn) Write(p []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	err := c.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// Close closes the WebSocket connection
func (c *wsConn) Close() error {
	return c.conn.Close()
}

// Server provides WebSocket server support for MCP
type Server struct {
	upgrader websocket.Upgrader
	handler  MessageHandler
	addr     string
}

// MessageHandler processes WebSocket messages
type MessageHandler func(ctx context.Context, msg []byte) ([]byte, error)

// NewServer creates a new WebSocket server for MCP
func NewServer(addr string, handler MessageHandler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true // Allow all origins by default
			},
		},
	}
}

// WithCheckOrigin sets a custom origin checker
func (s *Server) WithCheckOrigin(checkOrigin func(r *http.Request) bool) *Server {
	s.upgrader.CheckOrigin = checkOrigin
	return s
}

// ListenAndServe starts the WebSocket server
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleWebSocket)
	return http.ListenAndServe(s.addr, mux)
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade connection", http.StatusBadRequest)
		return
	}
	defer func() { _ = conn.Close() }()

	ctx := r.Context()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			// Check if it's an unexpected close error (could be logged)
			_ = websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)
			break
		}

		if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
			continue
		}

		response, err := s.handler(ctx, message)
		if err != nil {
			// Send error response
			errMsg := []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error()))
			_ = conn.WriteMessage(websocket.TextMessage, errMsg)
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
			break
		}
	}
}
