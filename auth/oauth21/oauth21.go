// Package oauth21 provides OAuth 2.1 authentication for MCP servers with mandatory PKCE support.
// OAuth 2.1 is a consolidation of best practices from OAuth 2.0, making PKCE mandatory and
// removing insecure flows like Implicit and Resource Owner Password Credentials grants.
package oauth21

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jmcarbo/fullmcp/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Provider implements OAuth 2.1 authentication with mandatory PKCE
type Provider struct {
	config           *oauth2.Config
	userInfoURL      string
	emailKey         string
	subjectKey       string
	verifyEmail      bool
	scopeMapping     map[string][]string
	pkceVerifiers    map[string]string // state -> code_verifier
	strictRedirectURI bool
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

// Option configures the OAuth 2.1 provider
type Option func(*Provider)

// New creates a new OAuth 2.1 provider with mandatory PKCE
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
		config:            config,
		userInfoURL:       userInfoURL,
		emailKey:          emailKey,
		subjectKey:        subjectKey,
		verifyEmail:       false,
		scopeMapping:      make(map[string][]string),
		pkceVerifiers:     make(map[string]string),
		strictRedirectURI: true, // OAuth 2.1 requires exact string matching
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

// PKCEChallenge represents PKCE challenge parameters
type PKCEChallenge struct {
	CodeVerifier  string
	CodeChallenge string
	Method        string // "S256" or "plain"
}

// GeneratePKCEChallenge generates a PKCE code challenge and verifier
// OAuth 2.1 requires PKCE for all authorization code flows
func GeneratePKCEChallenge() (*PKCEChallenge, error) {
	// Generate code_verifier (43-128 characters)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, fmt.Errorf("failed to generate verifier: %w", err)
	}

	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Generate code_challenge = BASE64URL(SHA256(code_verifier))
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return &PKCEChallenge{
		CodeVerifier:  verifier,
		CodeChallenge: challenge,
		Method:        "S256", // OAuth 2.1 recommends S256
	}, nil
}

// AuthCodeURLWithPKCE returns the URL for OAuth authorization with PKCE
// PKCE is mandatory in OAuth 2.1
func (p *Provider) AuthCodeURLWithPKCE(state string, challenge *PKCEChallenge) string {
	// Store verifier for later exchange
	p.pkceVerifiers[state] = challenge.CodeVerifier

	// OAuth 2.1 requires PKCE parameters
	return p.config.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", challenge.CodeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", challenge.Method),
	)
}

// ExchangeWithPKCE exchanges an authorization code for a token using PKCE
// OAuth 2.1 requires the code_verifier parameter
func (p *Provider) ExchangeWithPKCE(ctx context.Context, code, state string) (*oauth2.Token, error) {
	verifier, ok := p.pkceVerifiers[state]
	if !ok {
		return nil, fmt.Errorf("code verifier not found for state")
	}

	// Clean up verifier after use
	defer delete(p.pkceVerifiers, state)

	// Exchange with code_verifier (OAuth 2.1 requirement)
	return p.config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
}

// ValidateRedirectURI validates redirect URI using exact string matching
// OAuth 2.1 requires strict redirect URI matching (no wildcards or partial matches)
func (p *Provider) ValidateRedirectURI(providedURI string) error {
	if !p.strictRedirectURI {
		return nil
	}

	// Exact string comparison (OAuth 2.1 requirement)
	if providedURI != p.config.RedirectURL {
		return fmt.Errorf("redirect URI mismatch: expected exact match of %s", p.config.RedirectURL)
	}

	return nil
}

// Authenticate exchanges a code for a token and returns it as a string
func (p *Provider) Authenticate(ctx context.Context, credentials interface{}) (string, error) {
	// OAuth 2.1 requires structured credentials with state for PKCE
	creds, ok := credentials.(map[string]string)
	if !ok {
		return "", fmt.Errorf("invalid credentials type, expected map with code and state")
	}

	code := creds["code"]
	state := creds["state"]

	if code == "" || state == "" {
		return "", fmt.Errorf("code and state are required")
	}

	token, err := p.ExchangeWithPKCE(ctx, code, state)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}

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

// HandleCallback is a helper to handle OAuth callbacks with PKCE validation
func (p *Provider) HandleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if code == "" {
			http.Error(w, "missing authorization code", http.StatusBadRequest)
			return
		}

		if state == "" {
			http.Error(w, "missing state parameter", http.StatusBadRequest)
			return
		}

		// OAuth 2.1: Validate redirect URI (exact match)
		// In a real implementation, you'd validate against the original request
		// For now, we'll skip this as it requires session management

		// Exchange with PKCE
		token, err := p.ExchangeWithPKCE(r.Context(), code, state)
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

// Features documents the OAuth 2.1 compliance features
type Features struct {
	MandatoryPKCE          bool   // PKCE required for all clients
	StrictRedirectURI      bool   // Exact string matching for redirect URIs
	ImplicitGrantRemoved   bool   // Implicit flow not supported
	PasswordGrantRemoved   bool   // Resource Owner Password Credentials not supported
	CodeChallengeMethod    string // "S256" recommended
	MinimumVerifierLength  int    // 43 characters
	MaximumVerifierLength  int    // 128 characters
}

// GetOAuth21Features returns the OAuth 2.1 compliance features
func GetOAuth21Features() Features {
	return Features{
		MandatoryPKCE:         true,
		StrictRedirectURI:     true,
		ImplicitGrantRemoved:  true,
		PasswordGrantRemoved:  true,
		CodeChallengeMethod:   "S256",
		MinimumVerifierLength: 43,
		MaximumVerifierLength: 128,
	}
}
