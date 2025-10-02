// Package jwt provides JWT authentication for MCP servers.
package jwt

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmcarbo/fullmcp/auth"
)

// Provider implements JWT authentication
type Provider struct {
	signingKey    []byte
	signingMethod jwt.SigningMethod
	issuer        string
	expiration    time.Duration
}

// Option configures the JWT provider
type Option func(*Provider)

// New creates a new JWT provider
func New(signingKey []byte, opts ...Option) *Provider {
	p := &Provider{
		signingKey:    signingKey,
		signingMethod: jwt.SigningMethodHS256,
		issuer:        "mcp-server",
		expiration:    24 * time.Hour,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// WithSigningMethod sets the JWT signing method
func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(p *Provider) {
		p.signingMethod = method
	}
}

// WithIssuer sets the JWT issuer
func WithIssuer(issuer string) Option {
	return func(p *Provider) {
		p.issuer = issuer
	}
}

// WithExpiration sets the JWT expiration duration
func WithExpiration(expiration time.Duration) Option {
	return func(p *Provider) {
		p.expiration = expiration
	}
}

// GenerateRandomKey generates a random signing key
func GenerateRandomKey(size int) ([]byte, error) {
	key := make([]byte, size)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// CustomClaims extends standard JWT claims with auth.Claims
type CustomClaims struct {
	Subject string                 `json:"sub"`
	Email   string                 `json:"email,omitempty"`
	Scopes  []string               `json:"scopes,omitempty"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
	jwt.RegisteredClaims
}

// Authenticate creates a JWT token from credentials
func (p *Provider) Authenticate(_ context.Context, credentials interface{}) (string, error) {
	claims, ok := credentials.(auth.Claims)
	if !ok {
		return "", fmt.Errorf("invalid credentials type, expected auth.Claims")
	}

	now := time.Now()
	jwtClaims := CustomClaims{
		Subject: claims.Subject,
		Email:   claims.Email,
		Scopes:  claims.Scopes,
		Extra:   claims.Extra,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p.issuer,
			Subject:   claims.Subject,
			ExpiresAt: jwt.NewNumericDate(now.Add(p.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(p.signingMethod, jwtClaims)
	return token.SignedString(p.signingKey)
}

// ValidateToken validates a JWT token and returns claims
func (p *Provider) ValidateToken(_ context.Context, tokenString string) (auth.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if token.Method != p.signingMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return p.signingKey, nil
	})

	if err != nil {
		return auth.Claims{}, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return auth.Claims{}, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return auth.Claims{}, fmt.Errorf("invalid claims type")
	}

	return auth.Claims{
		Subject: claims.Subject,
		Email:   claims.Email,
		Scopes:  claims.Scopes,
		Extra:   claims.Extra,
	}, nil
}

// Middleware returns HTTP middleware for JWT authentication
func (p *Provider) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, "unauthorized: missing token", http.StatusUnauthorized)
				return
			}

			claims, err := p.ValidateToken(r.Context(), token)
			if err != nil {
				http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			ctx := auth.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken extracts the JWT token from the request
// Checks Authorization header (Bearer <token>)
func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

// CreateToken is a helper to create a token with specific claims
func (p *Provider) CreateToken(subject, email string, scopes []string, extra map[string]interface{}) (string, error) {
	claims := auth.Claims{
		Subject: subject,
		Email:   email,
		Scopes:  scopes,
		Extra:   extra,
	}
	return p.Authenticate(context.Background(), claims)
}
