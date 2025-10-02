// Package apikey provides API key authentication for MCP servers.
package apikey

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jmcarbo/fullmcp/auth"
)

// Provider implements API key authentication
type Provider struct {
	keys map[string]auth.Claims
}

// New creates a new API key provider
func New() *Provider {
	return &Provider{
		keys: make(map[string]auth.Claims),
	}
}

// AddKey adds an API key with associated claims
func (p *Provider) AddKey(apiKey string, claims auth.Claims) {
	p.keys[apiKey] = claims
}

// Authenticate validates an API key
func (p *Provider) Authenticate(_ context.Context, credentials interface{}) (string, error) {
	apiKey, ok := credentials.(string)
	if !ok {
		return "", fmt.Errorf("invalid credentials type")
	}

	if _, exists := p.keys[apiKey]; !exists {
		return "", fmt.Errorf("invalid API key")
	}

	return apiKey, nil
}

// ValidateToken validates an API key token
func (p *Provider) ValidateToken(_ context.Context, token string) (auth.Claims, error) {
	claims, exists := p.keys[token]
	if !exists {
		return auth.Claims{}, fmt.Errorf("invalid API key")
	}

	return claims, nil
}

// Middleware returns HTTP middleware for API key authentication
func (p *Provider) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractAPIKey(r)
			if token == "" {
				http.Error(w, "unauthorized: missing API key", http.StatusUnauthorized)
				return
			}

			claims, err := p.ValidateToken(r.Context(), token)
			if err != nil {
				http.Error(w, "unauthorized: invalid API key", http.StatusUnauthorized)
				return
			}

			ctx := auth.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractAPIKey extracts the API key from the request
// Checks Authorization header (Bearer <key>) or X-API-Key header
func extractAPIKey(r *http.Request) string {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Check X-API-Key header
	return r.Header.Get("X-API-Key")
}
