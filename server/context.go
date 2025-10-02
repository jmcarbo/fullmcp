// Package server provides MCP server implementation and context management.
package server

import "context"

type contextKey string

const (
	serverContextKey  contextKey = "mcp.server"
	sessionContextKey contextKey = "mcp.session"
)

// Context provides access to server capabilities from within handlers
type Context struct {
	server  *Server
	session interface{} // Session information
}

// WithContext adds server context to the given context
func (s *Server) WithContext(ctx context.Context, session interface{}) context.Context {
	sc := &Context{
		server:  s,
		session: session,
	}
	return context.WithValue(ctx, serverContextKey, sc)
}

// FromContext retrieves server context from the given context
func FromContext(ctx context.Context) *Context {
	sc, _ := ctx.Value(serverContextKey).(*Context)
	return sc
}

// ReadResource reads a resource from the server
func (sc *Context) ReadResource(_ context.Context, uri string) ([]byte, error) {
	if sc == nil || sc.server == nil {
		return nil, &ErrorContext{Message: "server context not available"}
	}
	return sc.server.resources.Read(context.Background(), uri)
}

// CallTool calls a tool from the server
func (sc *Context) CallTool(_ context.Context, _ string, _ interface{}) (interface{}, error) {
	if sc == nil || sc.server == nil {
		return nil, &ErrorContext{Message: "server context not available"}
	}
	// This would need proper marshaling in real implementation
	return nil, &ErrorContext{Message: "not implemented"}
}

// ErrorContext represents a context error
type ErrorContext struct {
	Message string
}

func (e *ErrorContext) Error() string {
	return e.Message
}
