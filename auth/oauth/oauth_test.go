package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmcarbo/fullmcp/auth"
)

func TestNew(t *testing.T) {
	provider := New(Google, "client-id", "client-secret", "http://localhost/callback", []string{"email", "profile"})

	if provider == nil {
		t.Fatal("expected non-nil provider")
	}

	if provider.config.ClientID != "client-id" {
		t.Error("client ID not set correctly")
	}

	if provider.config.ClientSecret != "client-secret" {
		t.Error("client secret not set correctly")
	}

	if provider.config.RedirectURL != "http://localhost/callback" {
		t.Error("redirect URL not set correctly")
	}
}

func TestNewGoogle(t *testing.T) {
	provider := New(Google, "id", "secret", "http://localhost/callback", []string{"email"})

	if provider.userInfoURL != "https://www.googleapis.com/oauth2/v2/userinfo" {
		t.Error("Google user info URL not set correctly")
	}

	if provider.emailKey != "email" {
		t.Error("email key not set correctly")
	}

	if provider.subjectKey != "id" {
		t.Error("subject key not set correctly")
	}
}

func TestNewGitHub(t *testing.T) {
	provider := New(GitHub, "id", "secret", "http://localhost/callback", []string{"user"})

	if provider.userInfoURL != "https://api.github.com/user" {
		t.Error("GitHub user info URL not set correctly")
	}
}

func TestWithOptions(t *testing.T) {
	scopeMapping := map[string][]string{
		"role": {"admin", "user"},
	}

	provider := New(Google, "id", "secret", "http://localhost/callback", []string{"email"},
		WithVerifyEmail(true),
		WithScopeMapping(scopeMapping),
		WithUserInfoURL("https://custom.example.com/userinfo"),
		WithCustomEndpoint("https://auth.example.com/authorize", "https://auth.example.com/token"),
	)

	if !provider.verifyEmail {
		t.Error("verify email not set")
	}

	if len(provider.scopeMapping) != 1 {
		t.Error("scope mapping not set")
	}

	if provider.userInfoURL != "https://custom.example.com/userinfo" {
		t.Error("custom user info URL not set")
	}

	if provider.config.Endpoint.AuthURL != "https://auth.example.com/authorize" {
		t.Error("custom auth URL not set")
	}
}

func TestAuthCodeURL(t *testing.T) {
	provider := New(Google, "id", "secret", "http://localhost/callback", []string{"email"})

	url := provider.AuthCodeURL("random-state")

	if url == "" {
		t.Error("expected non-empty auth URL")
	}

	// URL should contain client_id
	if !contains(url, "client_id=id") {
		t.Error("auth URL should contain client_id")
	}

	// URL should contain state
	if !contains(url, "state=random-state") {
		t.Error("auth URL should contain state")
	}
}

func TestValidateToken(t *testing.T) {
	// Create mock OAuth server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return mock user info
		userInfo := map[string]interface{}{
			"id":    "12345",
			"email": "user@example.com",
			"name":  "Test User",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userInfo)
	}))
	defer mockServer.Close()

	provider := New(Google, "id", "secret", "http://localhost/callback", []string{"email"},
		WithUserInfoURL(mockServer.URL),
	)

	claims, err := provider.ValidateToken(context.Background(), "mock-access-token")
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.Subject != "12345" {
		t.Errorf("expected subject '12345', got '%s'", claims.Subject)
	}

	if claims.Email != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got '%s'", claims.Email)
	}

	if claims.Extra["name"] != "Test User" {
		t.Error("expected name in extra claims")
	}
}

func TestValidateTokenServerError(t *testing.T) {
	// Create mock server that returns error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	provider := New(Google, "id", "secret", "http://localhost/callback", []string{"email"},
		WithUserInfoURL(mockServer.URL),
	)

	_, err := provider.ValidateToken(context.Background(), "mock-access-token")
	if err == nil {
		t.Error("expected error for server error")
	}
}

func TestMapScopes(t *testing.T) {
	scopeMapping := map[string][]string{
		"role": {"admin", "user"},
	}

	provider := New(Google, "id", "secret", "http://localhost/callback", []string{"email"},
		WithScopeMapping(scopeMapping),
	)

	userInfo := map[string]interface{}{
		"role": "admin",
	}

	scopes := provider.mapScopes(userInfo)

	if len(scopes) != 1 {
		t.Errorf("expected 1 scope, got %d", len(scopes))
	}

	if scopes[0] != "admin" {
		t.Errorf("expected scope 'admin', got '%s'", scopes[0])
	}
}

func TestMiddleware(t *testing.T) {
	// Create mock OAuth server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInfo := map[string]interface{}{
			"id":    "12345",
			"email": "user@example.com",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userInfo)
	}))
	defer mockServer.Close()

	provider := New(Google, "id", "secret", "http://localhost/callback", []string{"email"},
		WithUserInfoURL(mockServer.URL),
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.GetClaims(r.Context())
		if !ok {
			t.Error("expected claims in context")
			return
		}

		if claims.Email != "user@example.com" {
			t.Errorf("expected email 'user@example.com', got '%s'", claims.Email)
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := provider.Middleware()
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer mock-token")
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestMiddlewareNoToken(t *testing.T) {
	provider := New(Google, "id", "secret", "http://localhost/callback", []string{"email"})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	middleware := provider.Middleware()
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestMiddlewareVerifyEmail(t *testing.T) {
	// Create mock server that returns user without email
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInfo := map[string]interface{}{
			"id": "12345",
			// No email
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userInfo)
	}))
	defer mockServer.Close()

	provider := New(Google, "id", "secret", "http://localhost/callback", []string{"email"},
		WithUserInfoURL(mockServer.URL),
		WithVerifyEmail(true),
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	middleware := provider.Middleware()
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer mock-token")
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for missing email, got %d", rr.Code)
	}
}

func TestHandleCallback(t *testing.T) {
	// This test is more complex as it requires mocking the OAuth2 token exchange
	// For now, we'll test the basic structure

	provider := New(Google, "id", "secret", "http://localhost/callback", []string{"email"})

	handler := provider.HandleCallback()

	// Test missing code
	req := httptest.NewRequest("GET", "/callback", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing code, got %d", rr.Code)
	}
}

func TestAuthenticateInvalidCredentials(t *testing.T) {
	provider := New(Google, "id", "secret", "http://localhost/callback", []string{"email"})

	_, err := provider.Authenticate(context.Background(), 12345)
	if err == nil {
		t.Error("expected error for invalid credentials type")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			got := extractBearerToken(req)
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
