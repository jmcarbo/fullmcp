// Package auth provides authentication interfaces and utilities for MCP servers.
package auth

import (
	"context"
	"net/http"
)

// Provider handles authentication
type Provider interface {
	// Authenticate validates credentials and returns a token
	Authenticate(ctx context.Context, credentials interface{}) (string, error)

	// Middleware returns HTTP middleware for auth
	Middleware() func(http.Handler) http.Handler

	// ValidateToken validates a token
	ValidateToken(ctx context.Context, token string) (Claims, error)
}

// Claims represents authenticated user claims
type Claims struct {
	Subject string
	Email   string
	Scopes  []string
	Extra   map[string]interface{}
}

// contextKey is the type for context keys
type contextKey string

const claimsContextKey contextKey = "auth.claims"

// WithClaims adds claims to the context
func WithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// GetClaims retrieves claims from the context
func GetClaims(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(Claims)
	return claims, ok
}
