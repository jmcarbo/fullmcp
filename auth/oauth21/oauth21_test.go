package oauth21

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	provider := New(Google, "client-id", "client-secret", "http://localhost/callback", []string{"email"})

	if provider == nil {
		t.Fatal("expected non-nil provider")
	}

	if provider.config.ClientID != "client-id" {
		t.Errorf("expected client ID 'client-id', got '%s'", provider.config.ClientID)
	}

	if provider.config.RedirectURL != "http://localhost/callback" {
		t.Errorf("expected redirect URL 'http://localhost/callback', got '%s'", provider.config.RedirectURL)
	}

	// OAuth 2.1 defaults
	if !provider.strictRedirectURI {
		t.Error("expected strict redirect URI matching to be enabled")
	}
}

func TestGeneratePKCEChallenge(t *testing.T) {
	challenge, err := GeneratePKCEChallenge()
	if err != nil {
		t.Fatalf("failed to generate PKCE challenge: %v", err)
	}

	if challenge == nil {
		t.Fatal("expected non-nil challenge")
	}

	// Verify code_verifier length (43-128 characters)
	if len(challenge.CodeVerifier) < 43 || len(challenge.CodeVerifier) > 128 {
		t.Errorf("code_verifier length %d out of range [43, 128]", len(challenge.CodeVerifier))
	}

	// Verify code_challenge is base64url encoded
	_, err = base64.RawURLEncoding.DecodeString(challenge.CodeChallenge)
	if err != nil {
		t.Errorf("code_challenge is not valid base64url: %v", err)
	}

	// Verify method is S256
	if challenge.Method != "S256" {
		t.Errorf("expected method 'S256', got '%s'", challenge.Method)
	}

	// Verify challenges are unique
	challenge2, _ := GeneratePKCEChallenge()
	if challenge.CodeVerifier == challenge2.CodeVerifier {
		t.Error("expected unique code verifiers")
	}
}

func TestAuthCodeURLWithPKCE(t *testing.T) {
	provider := New(Google, "client-id", "client-secret", "http://localhost/callback", []string{"email"})

	challenge, err := GeneratePKCEChallenge()
	if err != nil {
		t.Fatalf("failed to generate challenge: %v", err)
	}

	state := "test-state-123"
	authURL := provider.AuthCodeURLWithPKCE(state, challenge)

	if authURL == "" {
		t.Fatal("expected non-empty auth URL")
	}

	// Verify URL contains PKCE parameters
	if !strings.Contains(authURL, "code_challenge=") {
		t.Error("auth URL missing code_challenge parameter")
	}

	if !strings.Contains(authURL, "code_challenge_method=S256") {
		t.Error("auth URL missing code_challenge_method parameter")
	}

	// Verify verifier is stored
	if provider.pkceVerifiers[state] != challenge.CodeVerifier {
		t.Error("code verifier not stored correctly")
	}
}

func TestValidateRedirectURI_Exact(t *testing.T) {
	redirectURI := "http://localhost:8080/callback"
	provider := New(Google, "client-id", "client-secret", redirectURI, []string{"email"})

	// Exact match should succeed
	err := provider.ValidateRedirectURI(redirectURI)
	if err != nil {
		t.Errorf("exact match should succeed: %v", err)
	}

	// Different URI should fail (OAuth 2.1 requirement)
	err = provider.ValidateRedirectURI("http://localhost:8080/callback2")
	if err == nil {
		t.Error("expected error for mismatched redirect URI")
	}

	// Partial match should fail (OAuth 2.1 requirement)
	err = provider.ValidateRedirectURI("http://localhost:8080/callback/extra")
	if err == nil {
		t.Error("expected error for partial redirect URI match")
	}
}

func TestWithVerifyEmail(t *testing.T) {
	provider := New(Google, "client-id", "client-secret", "http://localhost/callback", []string{"email"},
		WithVerifyEmail(true),
	)

	if !provider.verifyEmail {
		t.Error("expected verifyEmail to be true")
	}
}

func TestWithCustomEndpoint(t *testing.T) {
	authURL := "https://custom.com/auth"
	tokenURL := "https://custom.com/token"

	provider := New(Google, "client-id", "client-secret", "http://localhost/callback", []string{"email"},
		WithCustomEndpoint(authURL, tokenURL),
	)

	if provider.config.Endpoint.AuthURL != authURL {
		t.Errorf("expected auth URL '%s', got '%s'", authURL, provider.config.Endpoint.AuthURL)
	}

	if provider.config.Endpoint.TokenURL != tokenURL {
		t.Errorf("expected token URL '%s', got '%s'", tokenURL, provider.config.Endpoint.TokenURL)
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "valid bearer token",
			header:   "Bearer abc123",
			expected: "abc123",
		},
		{
			name:     "no bearer prefix",
			header:   "abc123",
			expected: "",
		},
		{
			name:     "empty header",
			header:   "",
			expected: "",
		},
		{
			name:     "lowercase bearer",
			header:   "bearer abc123",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", tt.header)

			result := extractBearerToken(req)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestMiddleware_MissingToken(t *testing.T) {
	provider := New(Google, "client-id", "client-secret", "http://localhost/callback", []string{"email"})

	handler := provider.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetOAuth21Features(t *testing.T) {
	features := GetOAuth21Features()

	if !features.MandatoryPKCE {
		t.Error("OAuth 2.1 should require mandatory PKCE")
	}

	if !features.StrictRedirectURI {
		t.Error("OAuth 2.1 should require strict redirect URI matching")
	}

	if !features.ImplicitGrantRemoved {
		t.Error("OAuth 2.1 should have Implicit grant removed")
	}

	if !features.PasswordGrantRemoved {
		t.Error("OAuth 2.1 should have Password grant removed")
	}

	if features.CodeChallengeMethod != "S256" {
		t.Errorf("expected S256 method, got %s", features.CodeChallengeMethod)
	}

	if features.MinimumVerifierLength != 43 {
		t.Errorf("expected min length 43, got %d", features.MinimumVerifierLength)
	}

	if features.MaximumVerifierLength != 128 {
		t.Errorf("expected max length 128, got %d", features.MaximumVerifierLength)
	}
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	provider := New(Google, "client-id", "client-secret", "http://localhost/callback", []string{"email"})

	// OAuth 2.1 requires structured credentials
	_, err := provider.Authenticate(context.Background(), "simple-string")
	if err == nil {
		t.Error("expected error for invalid credentials type")
	}

	// Missing state
	_, err = provider.Authenticate(context.Background(), map[string]string{
		"code": "abc123",
	})
	if err == nil {
		t.Error("expected error for missing state")
	}

	// Missing code
	_, err = provider.Authenticate(context.Background(), map[string]string{
		"state": "state123",
	})
	if err == nil {
		t.Error("expected error for missing code")
	}
}

func TestPKCEChallenge_Uniqueness(t *testing.T) {
	challenges := make(map[string]bool)

	// Generate 100 challenges and verify uniqueness
	for i := 0; i < 100; i++ {
		challenge, err := GeneratePKCEChallenge()
		if err != nil {
			t.Fatalf("failed to generate challenge: %v", err)
		}

		if challenges[challenge.CodeVerifier] {
			t.Error("duplicate code verifier generated")
		}

		challenges[challenge.CodeVerifier] = true
	}
}

func TestProviderType_GitHub(t *testing.T) {
	provider := New(GitHub, "client-id", "client-secret", "http://localhost/callback", []string{"user"})

	if provider.userInfoURL != "https://api.github.com/user" {
		t.Errorf("expected GitHub user info URL, got %s", provider.userInfoURL)
	}
}

func TestProviderType_Azure(t *testing.T) {
	provider := New(Azure, "client-id", "client-secret", "http://localhost/callback", []string{"openid"})

	if provider.emailKey != "email" {
		t.Errorf("expected email key 'email', got %s", provider.emailKey)
	}

	if provider.subjectKey != "sub" {
		t.Errorf("expected subject key 'sub', got %s", provider.subjectKey)
	}
}

func TestExchangeWithPKCE_MissingVerifier(t *testing.T) {
	provider := New(Google, "client-id", "client-secret", "http://localhost/callback", []string{"email"})

	// Try to exchange without storing verifier first
	_, err := provider.ExchangeWithPKCE(context.Background(), "code123", "unknown-state")
	if err == nil {
		t.Error("expected error for missing verifier")
	}

	if !strings.Contains(err.Error(), "code verifier not found") {
		t.Errorf("expected 'code verifier not found' error, got: %v", err)
	}
}

func TestWithScopeMapping(t *testing.T) {
	mapping := map[string][]string{
		"role": {"admin", "user"},
	}

	provider := New(Google, "client-id", "client-secret", "http://localhost/callback", []string{"email"},
		WithScopeMapping(mapping),
	)

	if len(provider.scopeMapping) != 1 {
		t.Errorf("expected 1 scope mapping, got %d", len(provider.scopeMapping))
	}

	if len(provider.scopeMapping["role"]) != 2 {
		t.Errorf("expected 2 mapped scopes for 'role', got %d", len(provider.scopeMapping["role"]))
	}
}

func TestWithUserInfoURL(t *testing.T) {
	customURL := "https://custom.com/userinfo"

	provider := New(Google, "client-id", "client-secret", "http://localhost/callback", []string{"email"},
		WithUserInfoURL(customURL),
	)

	if provider.userInfoURL != customURL {
		t.Errorf("expected user info URL '%s', got '%s'", customURL, provider.userInfoURL)
	}
}
