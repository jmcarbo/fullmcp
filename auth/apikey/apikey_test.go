package apikey

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmcarbo/fullmcp/auth"
)

func TestProvider_AddKey(t *testing.T) {
	provider := New()

	claims := auth.Claims{
		Subject: "user-1",
		Email:   "user@example.com",
		Scopes:  []string{"read"},
	}

	provider.AddKey("test-key", claims)

	// Verify the key was added by trying to validate it
	ctx := context.Background()
	retrieved, err := provider.ValidateToken(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to validate added key: %v", err)
	}

	if retrieved.Subject != claims.Subject {
		t.Errorf("expected subject '%s', got '%s'", claims.Subject, retrieved.Subject)
	}
}

func TestProvider_Authenticate_Valid(t *testing.T) {
	provider := New()

	claims := auth.Claims{
		Subject: "user-1",
		Email:   "user@example.com",
	}

	provider.AddKey("valid-key", claims)

	ctx := context.Background()
	token, err := provider.Authenticate(ctx, "valid-key")
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}

	if token != "valid-key" {
		t.Errorf("expected token 'valid-key', got '%s'", token)
	}
}

func TestProvider_Authenticate_Invalid(t *testing.T) {
	provider := New()

	ctx := context.Background()
	_, err := provider.Authenticate(ctx, "invalid-key")
	if err == nil {
		t.Error("expected error for invalid API key")
	}
}

func TestProvider_Authenticate_WrongType(t *testing.T) {
	provider := New()

	ctx := context.Background()
	_, err := provider.Authenticate(ctx, 123) // Not a string
	if err == nil {
		t.Error("expected error for non-string credentials")
	}
}

func TestProvider_ValidateToken_Valid(t *testing.T) {
	provider := New()

	claims := auth.Claims{
		Subject: "user-1",
		Email:   "user@example.com",
		Scopes:  []string{"read", "write"},
	}

	provider.AddKey("test-key", claims)

	ctx := context.Background()
	retrieved, err := provider.ValidateToken(ctx, "test-key")
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}

	if retrieved.Subject != claims.Subject {
		t.Errorf("expected subject '%s', got '%s'", claims.Subject, retrieved.Subject)
	}

	if retrieved.Email != claims.Email {
		t.Errorf("expected email '%s', got '%s'", claims.Email, retrieved.Email)
	}

	if len(retrieved.Scopes) != len(claims.Scopes) {
		t.Fatalf("expected %d scopes, got %d", len(claims.Scopes), len(retrieved.Scopes))
	}
}

func TestProvider_ValidateToken_Invalid(t *testing.T) {
	provider := New()

	ctx := context.Background()
	_, err := provider.ValidateToken(ctx, "nonexistent-key")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestProvider_Middleware_ValidKey(t *testing.T) {
	provider := New()

	claims := auth.Claims{
		Subject: "user-1",
		Email:   "user@example.com",
	}

	provider.AddKey("secret-key", claims)

	handler := provider.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that claims are in context
		retrievedClaims, ok := auth.GetClaims(r.Context())
		if !ok {
			t.Error("expected claims in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if retrievedClaims.Subject != claims.Subject {
			t.Errorf("expected subject '%s', got '%s'", claims.Subject, retrievedClaims.Subject)
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "secret-key")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestProvider_Middleware_BearerToken(t *testing.T) {
	provider := New()

	claims := auth.Claims{
		Subject: "user-1",
	}

	provider.AddKey("bearer-token", claims)

	handler := provider.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer bearer-token")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestProvider_Middleware_MissingKey(t *testing.T) {
	provider := New()

	handler := provider.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without API key")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	// No API key header

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestProvider_Middleware_InvalidKey(t *testing.T) {
	provider := New()

	handler := provider.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid API key")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestExtractAPIKey_XAPIKey(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "test-key")

	key := extractAPIKey(req)
	if key != "test-key" {
		t.Errorf("expected 'test-key', got '%s'", key)
	}
}

func TestExtractAPIKey_Bearer(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	key := extractAPIKey(req)
	if key != "test-token" {
		t.Errorf("expected 'test-token', got '%s'", key)
	}
}

func TestExtractAPIKey_BearerPriority(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer bearer-key")
	req.Header.Set("X-API-Key", "header-key")

	// Bearer should take priority
	key := extractAPIKey(req)
	if key != "bearer-key" {
		t.Errorf("expected 'bearer-key', got '%s'", key)
	}
}

func TestExtractAPIKey_NoKey(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	key := extractAPIKey(req)
	if key != "" {
		t.Errorf("expected empty string, got '%s'", key)
	}
}

func TestExtractAPIKey_InvalidBearer(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidScheme test-token")

	key := extractAPIKey(req)
	if key != "" {
		t.Errorf("expected empty string for invalid scheme, got '%s'", key)
	}
}
