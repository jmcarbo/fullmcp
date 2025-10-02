package server

import "context"

// LifespanFunc is called during server lifecycle
// It receives the context and server, and returns:
// - A new context (potentially with values added)
// - A cleanup function to call on shutdown
// - An error if initialization failed
type LifespanFunc func(context.Context, *Server) (context.Context, func(), error)
