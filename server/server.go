package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/jmcarbo/fullmcp/internal/jsonrpc"
	"github.com/jmcarbo/fullmcp/mcp"
)

// Server is the main MCP server
type Server struct {
	name         string
	version      string
	instructions string

	tools     *ToolManager
	resources *ResourceManager
	prompts   *PromptManager

	middleware   []Middleware
	lifespan     LifespanFunc
	sampling     *SamplingCapability
	rootsHandler RootsHandler
	logging      *LoggingManager
	progress     *ProgressTracker
	cancellation *CancellationManager
	completion   *CompletionManager
}

// Option configures a Server
type Option func(*Server)

// New creates a new MCP server
func New(name string, opts ...Option) *Server {
	s := &Server{
		name:      name,
		tools:     NewToolManager(),
		resources: NewResourceManager(),
		prompts:   NewPromptManager(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// WithVersion sets the server version
func WithVersion(version string) Option {
	return func(s *Server) {
		s.version = version
	}
}

// WithInstructions sets server instructions
func WithInstructions(instructions string) Option {
	return func(s *Server) {
		s.instructions = instructions
	}
}

// WithMiddleware adds middleware to the server
func WithMiddleware(mw ...Middleware) Option {
	return func(s *Server) {
		s.middleware = append(s.middleware, mw...)
	}
}

// WithLifespan sets the server lifespan function
func WithLifespan(fn LifespanFunc) Option {
	return func(s *Server) {
		s.lifespan = fn
	}
}

// AddTool registers a tool
func (s *Server) AddTool(handler *ToolHandler) error {
	return s.tools.Register(handler)
}

// AddResource registers a resource
func (s *Server) AddResource(handler *ResourceHandler) error {
	return s.resources.Register(handler)
}

// AddResourceTemplate registers a resource template
func (s *Server) AddResourceTemplate(handler *ResourceTemplateHandler) error {
	return s.resources.RegisterTemplate(handler)
}

// AddPrompt registers a prompt
func (s *Server) AddPrompt(handler *PromptHandler) error {
	return s.prompts.Register(handler)
}

// Run starts the server with stdio transport
func (s *Server) Run(ctx context.Context) error {
	return s.Serve(ctx, NewStdioTransport())
}

// Serve starts the server with a custom transport
func (s *Server) Serve(ctx context.Context, conn io.ReadWriteCloser) error {
	reader := jsonrpc.NewMessageReader(conn)
	writer := jsonrpc.NewMessageWriter(conn)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		response := s.HandleMessage(ctx, msg)
		if response != nil {
			if err := writer.Write(response); err != nil {
				return err
			}
		}
	}
}

type messageHandler func(context.Context, *mcp.Message) *mcp.Message

// getMessageRouter returns the method routing map
func (s *Server) getMessageRouter() map[string]messageHandler {
	return map[string]messageHandler{
		"initialize":                       func(_ context.Context, msg *mcp.Message) *mcp.Message { return s.handleInitialize(msg) },
		"tools/list":                       s.handleToolsList,
		"tools/call":                       s.handleToolsCall,
		"resources/list":                   func(_ context.Context, msg *mcp.Message) *mcp.Message { return s.handleResourcesList(msg) },
		"resources/read":                   s.handleResourcesRead,
		"resources/templates/list":         func(_ context.Context, msg *mcp.Message) *mcp.Message { return s.handleResourceTemplatesList(msg) },
		"prompts/list":                     func(_ context.Context, msg *mcp.Message) *mcp.Message { return s.handlePromptsList(msg) },
		"prompts/get":                      s.handlePromptsGet,
		"notifications/roots/list_changed": s.handleRootsListChanged,
		"logging/setLevel":                 s.handleLoggingSetLevel,
		"notifications/cancelled":          s.handleCancelled,
		"ping":                             func(_ context.Context, msg *mcp.Message) *mcp.Message { return s.handlePing(msg) },
		"completion/complete":              s.handleCompletionComplete,
	}
}

// HandleMessage processes an MCP message and returns a response
func (s *Server) HandleMessage(ctx context.Context, msg *mcp.Message) *mcp.Message {
	if msg.Method == "" {
		return nil
	}

	router := s.getMessageRouter()
	if handler, ok := router[msg.Method]; ok {
		return handler(ctx, msg)
	}

	return s.errorResponse(msg.ID, mcp.MethodNotFound, "method not found")
}

func (s *Server) handleInitialize(msg *mcp.Message) *mcp.Message {
	caps := mcp.ServerCapabilities{
		Tools:     &mcp.ToolsCapability{},
		Resources: &mcp.ResourcesCapability{},
		Prompts:   &mcp.PromptsCapability{},
	}

	// Add completions capability if enabled (2025-03-26)
	if s.completion != nil {
		caps.Completions = &mcp.CompletionsCapability{}
	}

	result := map[string]interface{}{
		"protocolVersion": "2025-06-18",
		"capabilities":    caps,
		"serverInfo": map[string]string{
			"name":    s.name,
			"version": s.version,
		},
	}

	return s.successResponse(msg.ID, result)
}

func (s *Server) handleToolsList(ctx context.Context, msg *mcp.Message) *mcp.Message {
	tools, _ := s.tools.List(ctx)
	result := map[string]interface{}{
		"tools": tools,
	}
	return s.successResponse(msg.ID, result)
}

func (s *Server) handleToolsCall(ctx context.Context, msg *mcp.Message) *mcp.Message {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.errorResponse(msg.ID, mcp.InvalidParams, "invalid parameters")
	}

	result, err := s.tools.Call(ctx, params.Name, params.Arguments)
	if err != nil {
		return s.errorResponse(msg.ID, mcp.InternalError, err.Error())
	}

	content := []mcp.TextContent{
		{
			Type: "text",
			Text: fmt.Sprintf("%v", result),
		},
	}

	return s.successResponse(msg.ID, map[string]interface{}{
		"content": content,
	})
}

func (s *Server) handleResourcesList(msg *mcp.Message) *mcp.Message {
	resources := s.resources.List()
	result := map[string]interface{}{
		"resources": resources,
	}
	return s.successResponse(msg.ID, result)
}

func (s *Server) handleResourcesRead(ctx context.Context, msg *mcp.Message) *mcp.Message {
	var params struct {
		URI string `json:"uri"`
	}

	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.errorResponse(msg.ID, mcp.InvalidParams, "invalid parameters")
	}

	data, err := s.resources.Read(ctx, params.URI)
	if err != nil {
		return s.errorResponse(msg.ID, mcp.InternalError, err.Error())
	}

	contents := []map[string]interface{}{
		{
			"uri":      params.URI,
			"mimeType": "text/plain",
			"text":     string(data),
		},
	}

	return s.successResponse(msg.ID, map[string]interface{}{
		"contents": contents,
	})
}

func (s *Server) handleResourceTemplatesList(msg *mcp.Message) *mcp.Message {
	templates := s.resources.ListTemplates()
	result := map[string]interface{}{
		"resourceTemplates": templates,
	}
	return s.successResponse(msg.ID, result)
}

func (s *Server) handlePromptsList(msg *mcp.Message) *mcp.Message {
	prompts := s.prompts.List()
	result := map[string]interface{}{
		"prompts": prompts,
	}
	return s.successResponse(msg.ID, result)
}

func (s *Server) handlePromptsGet(ctx context.Context, msg *mcp.Message) *mcp.Message {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.errorResponse(msg.ID, mcp.InvalidParams, "invalid parameters")
	}

	messages, err := s.prompts.Get(ctx, params.Name, params.Arguments)
	if err != nil {
		return s.errorResponse(msg.ID, mcp.InternalError, err.Error())
	}

	return s.successResponse(msg.ID, map[string]interface{}{
		"messages": messages,
	})
}

func (s *Server) handleRootsListChanged(ctx context.Context, _ *mcp.Message) *mcp.Message {
	// This is a notification, so no response is expected
	if s.rootsHandler != nil {
		go s.rootsHandler(ctx)
	}
	return nil
}

func (s *Server) handleLoggingSetLevel(ctx context.Context, msg *mcp.Message) *mcp.Message {
	var params mcp.SetLevelRequest
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.errorResponse(msg.ID, mcp.InvalidParams, "invalid parameters")
	}

	if err := s.SetLogLevel(ctx, params.Level); err != nil {
		return s.errorResponse(msg.ID, mcp.InternalError, err.Error())
	}

	return s.successResponse(msg.ID, map[string]interface{}{})
}

func (s *Server) handleCancelled(_ context.Context, msg *mcp.Message) *mcp.Message {
	// This is a notification, so no response is expected
	if s.cancellation != nil {
		var notification mcp.CancelledNotification
		if err := json.Unmarshal(msg.Params, &notification); err == nil {
			s.cancellation.HandleCancellation(&notification)
		}
	}
	return nil
}

func (s *Server) handlePing(msg *mcp.Message) *mcp.Message {
	// Ping just returns an empty success response
	return s.successResponse(msg.ID, map[string]interface{}{})
}

func (s *Server) handleCompletionComplete(ctx context.Context, msg *mcp.Message) *mcp.Message {
	if s.completion == nil {
		// Return empty completions if not enabled
		return s.successResponse(msg.ID, map[string]interface{}{
			"completion": map[string]interface{}{
				"values": []string{},
			},
		})
	}

	var params mcp.CompleteRequest
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return s.errorResponse(msg.ID, mcp.InvalidParams, "invalid parameters")
	}

	values, err := s.completion.GetCompletion(ctx, params.Ref, params.Argument)
	if err != nil {
		return s.errorResponse(msg.ID, mcp.InternalError, err.Error())
	}

	return s.successResponse(msg.ID, map[string]interface{}{
		"completion": map[string]interface{}{
			"values": values,
		},
	})
}

func (s *Server) successResponse(id interface{}, result interface{}) *mcp.Message {
	resultJSON, _ := json.Marshal(result)
	return &mcp.Message{
		JSONRPC: "2.0",
		ID:      id,
		Result:  resultJSON,
	}
}

func (s *Server) errorResponse(id interface{}, code mcp.ErrorCode, message string) *mcp.Message {
	return &mcp.Message{
		JSONRPC: "2.0",
		ID:      id,
		Error: &mcp.RPCError{
			Code:    int(code),
			Message: message,
		},
	}
}
