package server

import (
	"context"

	"github.com/jmcarbo/fullmcp/mcp"
)

// Middleware wraps request handling
type Middleware func(Handler) Handler

// Handler processes MCP requests
type Handler func(context.Context, *Request) (*Response, error)

// Request represents an MCP request
type Request struct {
	Method string
	Params interface{}
	ID     interface{}
}

// Response represents an MCP response
type Response struct {
	Result interface{}
	Error  *mcp.RPCError
}

// ApplyMiddleware applies middleware chain to handler
func ApplyMiddleware(handler Handler, middleware []Middleware) Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

// LoggingMiddleware logs requests and responses
func LoggingMiddleware(logger Logger) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (*Response, error) {
			logger.Infof("Request: %s", req.Method)
			resp, err := next(ctx, req)
			if err != nil {
				logger.Errorf("Error: %v", err)
			}
			return resp, err
		}
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req *Request) (resp *Response, err error) {
			defer func() {
				if r := recover(); r != nil {
					resp = &Response{
						Error: &mcp.RPCError{
							Code:    int(mcp.InternalError),
							Message: "internal server error",
						},
					}
				}
			}()
			return next(ctx, req)
		}
	}
}

// Logger interface for middleware
type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}
