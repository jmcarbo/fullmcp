// Package client provides an MCP (Model Context Protocol) client implementation.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/jmcarbo/fullmcp/internal/jsonrpc"
	"github.com/jmcarbo/fullmcp/mcp"
)

// Client is an MCP client
type Client struct {
	transport io.ReadWriteCloser
	reader    *jsonrpc.MessageReader
	writer    *jsonrpc.MessageWriter

	mu      sync.Mutex
	nextID  atomic.Int64
	pending map[int64]chan *mcp.Message

	capabilities     *mcp.ServerCapabilities
	samplingHandler  SamplingHandler  // Handler for server-initiated sampling requests
	rootsProvider    RootsProvider    // Provider for client roots
	logHandler       LogHandler       // Handler for log message notifications
	progressHandler  ProgressHandler  // Handler for progress notifications
}

// Option configures a Client
type Option func(*Client)

// New creates a new MCP client
func New(transport io.ReadWriteCloser, opts ...Option) *Client {
	c := &Client{
		transport: transport,
		reader:    jsonrpc.NewMessageReader(transport),
		writer:    jsonrpc.NewMessageWriter(transport),
		pending:   make(map[int64]chan *mcp.Message),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Connect establishes a connection and initializes
func (c *Client) Connect(ctx context.Context) error {
	// Start message handler
	go c.handleMessages()

	// Initialize
	var initResult struct {
		ProtocolVersion string                 `json:"protocolVersion"`
		Capabilities    mcp.ServerCapabilities `json:"capabilities"`
		ServerInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
	}

	capabilities := map[string]interface{}{}
	if c.rootsProvider != nil {
		capabilities["roots"] = map[string]bool{
			"listChanged": true,
		}
	}

	if err := c.call(ctx, "initialize", map[string]interface{}{
		"protocolVersion": "2025-06-18",
		"capabilities":    capabilities,
		"clientInfo": map[string]string{
			"name":    "fullmcp-client",
			"version": "0.1.0",
		},
	}, &initResult); err != nil {
		return err
	}

	c.mu.Lock()
	c.capabilities = &initResult.Capabilities
	c.mu.Unlock()

	// Send initialized notification
	return c.notify("notifications/initialized", nil)
}

// Close closes the connection
func (c *Client) Close() error {
	if c.transport != nil {
		return c.transport.Close()
	}
	return nil
}

// ListTools lists available tools
func (c *Client) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
	var result struct {
		Tools []*mcp.Tool `json:"tools"`
	}

	if err := c.call(ctx, "tools/list", nil, &result); err != nil {
		return nil, err
	}

	return result.Tools, nil
}

// CallTool calls a tool
func (c *Client) CallTool(ctx context.Context, name string, args interface{}) (interface{}, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": args,
	}

	var result struct {
		Content []json.RawMessage `json:"content"`
	}

	if err := c.call(ctx, "tools/call", params, &result); err != nil {
		return nil, err
	}

	if len(result.Content) > 0 {
		var textContent mcp.TextContent
		if err := json.Unmarshal(result.Content[0], &textContent); err == nil {
			return textContent.Text, nil
		}
	}

	return result.Content, nil
}

// ListResources lists available resources
func (c *Client) ListResources(ctx context.Context) ([]*mcp.Resource, error) {
	var result struct {
		Resources []*mcp.Resource `json:"resources"`
	}

	if err := c.call(ctx, "resources/list", nil, &result); err != nil {
		return nil, err
	}

	return result.Resources, nil
}

// ReadResource reads a resource
func (c *Client) ReadResource(ctx context.Context, uri string) ([]byte, error) {
	params := map[string]interface{}{
		"uri": uri,
	}

	var result struct {
		Contents []struct {
			URI      string `json:"uri"`
			MimeType string `json:"mimeType"`
			Text     string `json:"text,omitempty"`
			Blob     string `json:"blob,omitempty"`
		} `json:"contents"`
	}

	if err := c.call(ctx, "resources/read", params, &result); err != nil {
		return nil, err
	}

	if len(result.Contents) == 0 {
		return nil, &mcp.NotFoundError{Type: "resource", Name: uri}
	}

	return []byte(result.Contents[0].Text), nil
}

// ListPrompts lists available prompts
func (c *Client) ListPrompts(ctx context.Context) ([]*mcp.Prompt, error) {
	var result struct {
		Prompts []*mcp.Prompt `json:"prompts"`
	}

	if err := c.call(ctx, "prompts/list", nil, &result); err != nil {
		return nil, err
	}

	return result.Prompts, nil
}

// GetPrompt gets a prompt
func (c *Client) GetPrompt(ctx context.Context, name string, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": args,
	}

	var result struct {
		Messages []*mcp.PromptMessage `json:"messages"`
	}

	if err := c.call(ctx, "prompts/get", params, &result); err != nil {
		return nil, err
	}

	return result.Messages, nil
}

func (c *Client) call(ctx context.Context, method string, params, result interface{}) error {
	// Check if context is already canceled before starting
	if ctx.Err() != nil {
		return ctx.Err()
	}

	id := c.nextID.Add(1)

	msg := &mcp.Message{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
	}

	if params != nil {
		paramsJSON, err := json.Marshal(params)
		if err != nil {
			return err
		}
		msg.Params = paramsJSON
	}

	respChan := make(chan *mcp.Message, 1)

	c.mu.Lock()
	c.pending[id] = respChan
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	if err := c.writer.Write(msg); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case resp := <-respChan:
		if resp.Error != nil {
			return fmt.Errorf("RPC error %d: %s", resp.Error.Code, resp.Error.Message)
		}

		if result != nil && resp.Result != nil {
			return json.Unmarshal(resp.Result, result)
		}

		return nil
	}
}

func (c *Client) notify(method string, params interface{}) error {
	msg := &mcp.Message{
		JSONRPC: "2.0",
		Method:  method,
	}

	if params != nil {
		paramsJSON, err := json.Marshal(params)
		if err != nil {
			return err
		}
		msg.Params = paramsJSON
	}

	return c.writer.Write(msg)
}

func (c *Client) handleMessages() {
	for {
		msg, err := c.reader.Read()
		if err != nil {
			return
		}

		// Handle notifications from server (no ID)
		if msg.Method != "" && msg.ID == nil {
			c.handleServerNotification(msg)
			continue
		}

		// Handle requests from server (like roots/list)
		if msg.Method != "" && msg.ID != nil {
			c.handleServerRequest(msg)
			continue
		}

		// Handle responses to client requests
		if msg.ID != nil {
			id, ok := msg.ID.(float64)
			if !ok {
				continue
			}

			c.mu.Lock()
			ch, exists := c.pending[int64(id)]
			c.mu.Unlock()

			if exists {
				ch <- msg
			}
		}
	}
}

func (c *Client) handleServerNotification(msg *mcp.Message) {
	switch msg.Method {
	case "notifications/message":
		// Handle log message notification
		if c.logHandler != nil {
			var logMsg mcp.LogMessage
			if err := json.Unmarshal(msg.Params, &logMsg); err == nil {
				go c.logHandler(context.Background(), &logMsg)
			}
		}
	case "notifications/progress":
		// Handle progress notification
		if c.progressHandler != nil {
			var progressNotif mcp.ProgressNotification
			if err := json.Unmarshal(msg.Params, &progressNotif); err == nil {
				go c.progressHandler(context.Background(), &progressNotif)
			}
		}
	}
}

func (c *Client) handleServerRequest(msg *mcp.Message) {
	var response *mcp.Message

	switch msg.Method {
	case "roots/list":
		result, err := c.handleRootsList(context.Background())
		if err != nil {
			response = c.errorResponse(msg.ID, mcp.InternalError, err.Error())
		} else {
			response = c.successResponse(msg.ID, result)
		}
	default:
		response = c.errorResponse(msg.ID, mcp.MethodNotFound, "method not found")
	}

	if response != nil {
		_ = c.writer.Write(response)
	}
}

func (c *Client) successResponse(id interface{}, result interface{}) *mcp.Message {
	resultJSON, _ := json.Marshal(result)
	return &mcp.Message{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultJSON,
	}
}

func (c *Client) errorResponse(id interface{}, code mcp.ErrorCode, message string) *mcp.Message {
	return &mcp.Message{
		JSONRPC: "2.0",
		ID:      id,
		Error: &mcp.RPCError{
			Code:    int(code),
			Message: message,
		},
	}
}
