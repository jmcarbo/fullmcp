// Package oauth provides OAuth 2.0 authentication for MCP servers.
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jmcarbo/fullmcp/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Provider implements OAuth 2.0 authentication
type Provider struct {
	config       *oauth2.Config
	userInfoURL  string
	emailKey     string
	subjectKey   string
	verifyEmail  bool
	scopeMapping map[string][]string
}

// ProviderType represents the OAuth provider type
type ProviderType string

const (
	// Google OAuth provider
	Google ProviderType = "google"
	// GitHub OAuth provider
	GitHub ProviderType = "github"
	// Azure OAuth provider
	Azure ProviderType = "azure"
)

// Option configures the OAuth provider
type Option func(*Provider)

// New creates a new OAuth provider
func New(providerType ProviderType, clientID, clientSecret string, redirectURL string, scopes []string, opts ...Option) *Provider {
	var config *oauth2.Config
	var userInfoURL string
	var emailKey, subjectKey string

	switch providerType {
	case Google:
		config = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		}
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
		emailKey = "email"
		subjectKey = "id"

	case GitHub:
		config = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     github.Endpoint,
		}
		userInfoURL = "https://api.github.com/user"
		emailKey = "email"
		subjectKey = "id"

	default:
		// Generic OAuth2
		config = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
		}
		emailKey = "email"
		subjectKey = "sub"
	}

	p := &Provider{
		config:       config,
		userInfoURL:  userInfoURL,
		emailKey:     emailKey,
		subjectKey:   subjectKey,
		verifyEmail:  false,
		scopeMapping: make(map[string][]string),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// WithVerifyEmail enables email verification
func WithVerifyEmail(verify bool) Option {
	return func(p *Provider) {
		p.verifyEmail = verify
	}
}

// WithScopeMapping sets custom scope mapping
func WithScopeMapping(mapping map[string][]string) Option {
	return func(p *Provider) {
		p.scopeMapping = mapping
	}
}

// WithCustomEndpoint sets a custom OAuth endpoint
func WithCustomEndpoint(authURL, tokenURL string) Option {
	return func(p *Provider) {
		p.config.Endpoint = oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		}
	}
}

// WithUserInfoURL sets a custom user info URL
func WithUserInfoURL(url string) Option {
	return func(p *Provider) {
		p.userInfoURL = url
	}
}

// AuthCodeURL returns the URL for OAuth authorization
func (p *Provider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Exchange exchanges an authorization code for a token
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

// Authenticate exchanges a code for a token and returns it as a string
func (p *Provider) Authenticate(ctx context.Context, credentials interface{}) (string, error) {
	code, ok := credentials.(string)
	if !ok {
		return "", fmt.Errorf("invalid credentials type, expected authorization code")
	}

	token, err := p.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}

	// Return the access token
	return token.AccessToken, nil
}

// ValidateToken validates an OAuth token and retrieves user info
func (p *Provider) ValidateToken(ctx context.Context, accessToken string) (auth.Claims, error) {
	// Create HTTP client with token
	token := &oauth2.Token{AccessToken: accessToken}
	client := p.config.Client(ctx, token)

	// Fetch user info
	resp, err := client.Get(p.userInfoURL)
	if err != nil {
		return auth.Claims{}, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return auth.Claims{}, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return auth.Claims{}, fmt.Errorf("failed to decode user info: %w", err)
	}

	// Extract claims
	claims := auth.Claims{
		Extra: userInfo,
	}

	// Extract subject
	if sub, ok := userInfo[p.subjectKey]; ok {
		claims.Subject = fmt.Sprintf("%v", sub)
	}

	// Extract email
	if email, ok := userInfo[p.emailKey]; ok {
		claims.Email = fmt.Sprintf("%v", email)
	}

	// Map scopes from user info if available
	if len(p.scopeMapping) > 0 {
		claims.Scopes = p.mapScopes(userInfo)
	}

	return claims, nil
}

// mapScopes maps user info to scopes based on configuration
func (p *Provider) mapScopes(userInfo map[string]interface{}) []string {
	var scopes []string
	for key, mappedScopes := range p.scopeMapping {
		if val, ok := userInfo[key]; ok {
			valStr := fmt.Sprintf("%v", val)
			for _, scope := range mappedScopes {
				if strings.Contains(valStr, scope) {
					scopes = append(scopes, scope)
				}
			}
		}
	}
	return scopes
}

// Middleware returns HTTP middleware for OAuth authentication
func (p *Provider) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				http.Error(w, "unauthorized: missing token", http.StatusUnauthorized)
				return
			}

			claims, err := p.ValidateToken(r.Context(), token)
			if err != nil {
				http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			// Verify email if required
			if p.verifyEmail && claims.Email == "" {
				http.Error(w, "unauthorized: email verification required", http.StatusUnauthorized)
				return
			}

			ctx := auth.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractBearerToken extracts the bearer token from the Authorization header
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

// HandleCallback is a helper to handle OAuth callbacks
func (p *Provider) HandleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing authorization code", http.StatusBadRequest)
			return
		}

		token, err := p.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, "failed to exchange code", http.StatusInternalServerError)
			return
		}

		// Get user claims
		claims, err := p.ValidateToken(r.Context(), token.AccessToken)
		if err != nil {
			http.Error(w, "failed to get user info", http.StatusInternalServerError)
			return
		}

		// Return token and claims as JSON
		response := map[string]interface{}{
			"access_token": token.AccessToken,
			"claims":       claims,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}
