package jwt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmcarbo/fullmcp/auth"
)

func TestNew(t *testing.T) {
	key := []byte("test-secret-key")
	provider := New(key)

	if provider == nil {
		t.Fatal("expected non-nil provider")
	}

	if string(provider.signingKey) != string(key) {
		t.Error("signing key not set correctly")
	}

	if provider.signingMethod != jwt.SigningMethodHS256 {
		t.Error("expected default signing method HS256")
	}

	if provider.issuer != "mcp-server" {
		t.Error("expected default issuer 'mcp-server'")
	}

	if provider.expiration != 24*time.Hour {
		t.Error("expected default expiration 24 hours")
	}
}

func TestWithOptions(t *testing.T) {
	key := []byte("test-key")
	provider := New(key,
		WithIssuer("custom-issuer"),
		WithExpiration(1*time.Hour),
		WithSigningMethod(jwt.SigningMethodHS512),
	)

	if provider.issuer != "custom-issuer" {
		t.Errorf("expected issuer 'custom-issuer', got '%s'", provider.issuer)
	}

	if provider.expiration != 1*time.Hour {
		t.Errorf("expected expiration 1 hour, got %v", provider.expiration)
	}

	if provider.signingMethod != jwt.SigningMethodHS512 {
		t.Error("expected signing method HS512")
	}
}

func TestGenerateRandomKey(t *testing.T) {
	key, err := GenerateRandomKey(32)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("expected key length 32, got %d", len(key))
	}

	// Generate another key and ensure they're different
	key2, err := GenerateRandomKey(32)
	if err != nil {
		t.Fatalf("failed to generate second key: %v", err)
	}

	if string(key) == string(key2) {
		t.Error("expected different random keys")
	}
}

func TestAuthenticate(t *testing.T) {
	key := []byte("test-secret-key")
	provider := New(key)

	claims := auth.Claims{
		Subject: "user123",
		Email:   "user@example.com",
		Scopes:  []string{"read", "write"},
		Extra: map[string]interface{}{
			"role": "admin",
		},
	}

	token, err := provider.Authenticate(context.Background(), claims)
	if err != nil {
		t.Fatalf("failed to authenticate: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	// Validate the token
	validatedClaims, err := provider.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if validatedClaims.Subject != claims.Subject {
		t.Errorf("expected subject '%s', got '%s'", claims.Subject, validatedClaims.Subject)
	}

	if validatedClaims.Email != claims.Email {
		t.Errorf("expected email '%s', got '%s'", claims.Email, validatedClaims.Email)
	}

	if len(validatedClaims.Scopes) != len(claims.Scopes) {
		t.Errorf("expected %d scopes, got %d", len(claims.Scopes), len(validatedClaims.Scopes))
	}
}

func TestAuthenticateInvalidCredentials(t *testing.T) {
	key := []byte("test-secret-key")
	provider := New(key)

	_, err := provider.Authenticate(context.Background(), "invalid")
	if err == nil {
		t.Error("expected error for invalid credentials type")
	}
}

func TestValidateTokenExpired(t *testing.T) {
	key := []byte("test-secret-key")
	provider := New(key, WithExpiration(-1*time.Hour)) // Already expired

	claims := auth.Claims{
		Subject: "user123",
	}

	token, err := provider.Authenticate(context.Background(), claims)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Try to validate expired token
	_, err = provider.ValidateToken(context.Background(), token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidateTokenInvalidSignature(t *testing.T) {
	key1 := []byte("key1")
	key2 := []byte("key2")

	provider1 := New(key1)
	provider2 := New(key2)

	claims := auth.Claims{
		Subject: "user123",
	}

	// Create token with provider1
	token, err := provider1.Authenticate(context.Background(), claims)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Try to validate with provider2 (different key)
	_, err = provider2.ValidateToken(context.Background(), token)
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestValidateTokenInvalidFormat(t *testing.T) {
	key := []byte("test-secret-key")
	provider := New(key)

	_, err := provider.ValidateToken(context.Background(), "invalid.token.format")
	if err == nil {
		t.Error("expected error for invalid token format")
	}
}

func TestMiddleware(t *testing.T) {
	key := []byte("test-secret-key")
	provider := New(key)

	claims := auth.Claims{
		Subject: "user123",
		Email:   "user@example.com",
		Scopes:  []string{"read"},
	}

	token, err := provider.Authenticate(context.Background(), claims)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that claims are in context
		gotClaims, ok := auth.GetClaims(r.Context())
		if !ok {
			t.Error("expected claims in context")
			return
		}

		if gotClaims.Subject != claims.Subject {
			t.Errorf("expected subject '%s', got '%s'", claims.Subject, gotClaims.Subject)
		}

		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	middleware := provider.Middleware()
	wrappedHandler := middleware(handler)

	// Test with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestMiddlewareNoToken(t *testing.T) {
	key := []byte("test-secret-key")
	provider := New(key)

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

func TestMiddlewareInvalidToken(t *testing.T) {
	key := []byte("test-secret-key")
	provider := New(key)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	middleware := provider.Middleware()
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestExtractToken(t *testing.T) {
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
			name:     "bearer with space only",
			header:   "Bearer ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			got := extractToken(req)
			if got != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

func TestCreateToken(t *testing.T) {
	key := []byte("test-secret-key")
	provider := New(key)

	token, err := provider.CreateToken(
		"user456",
		"user456@example.com",
		[]string{"admin", "write"},
		map[string]interface{}{"org": "acme"},
	)

	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	// Validate the token
	claims, err := provider.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.Subject != "user456" {
		t.Errorf("expected subject 'user456', got '%s'", claims.Subject)
	}

	if claims.Email != "user456@example.com" {
		t.Errorf("expected email 'user456@example.com', got '%s'", claims.Email)
	}

	if len(claims.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(claims.Scopes))
	}

	if claims.Extra["org"] != "acme" {
		t.Errorf("expected org 'acme', got '%v'", claims.Extra["org"])
	}
}

func TestValidateTokenWrongSigningMethod(t *testing.T) {
	// Create token with HS256
	key := []byte("test-secret-key")
	provider1 := New(key, WithSigningMethod(jwt.SigningMethodHS256))

	claims := auth.Claims{Subject: "user123"}
	token, _ := provider1.Authenticate(context.Background(), claims)

	// Try to validate with HS512
	provider2 := New(key, WithSigningMethod(jwt.SigningMethodHS512))
	_, err := provider2.ValidateToken(context.Background(), token)

	if err == nil {
		t.Error("expected error for wrong signing method")
	}
}
