package auth

import (
	"context"
	"testing"
)

func TestWithClaims(t *testing.T) {
	ctx := context.Background()
	claims := Claims{
		Subject: "user-123",
		Email:   "user@example.com",
		Scopes:  []string{"read", "write"},
		Extra: map[string]interface{}{
			"custom": "value",
		},
	}

	ctx = WithClaims(ctx, claims)

	retrieved, ok := GetClaims(ctx)
	if !ok {
		t.Fatal("expected claims to be present in context")
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

	for i, scope := range claims.Scopes {
		if retrieved.Scopes[i] != scope {
			t.Errorf("scope %d: expected '%s', got '%s'", i, scope, retrieved.Scopes[i])
		}
	}

	if retrieved.Extra["custom"] != claims.Extra["custom"] {
		t.Errorf("expected extra field 'custom' to match")
	}
}

func TestGetClaims_NotPresent(t *testing.T) {
	ctx := context.Background()

	_, ok := GetClaims(ctx)
	if ok {
		t.Error("expected claims to not be present in empty context")
	}
}

func TestGetClaims_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), claimsContextKey, "not-claims")

	_, ok := GetClaims(ctx)
	if ok {
		t.Error("expected GetClaims to fail with wrong type in context")
	}
}

func TestClaims_EmptyScopes(t *testing.T) {
	ctx := context.Background()
	claims := Claims{
		Subject: "user-123",
		Email:   "user@example.com",
		Scopes:  []string{},
	}

	ctx = WithClaims(ctx, claims)

	retrieved, ok := GetClaims(ctx)
	if !ok {
		t.Fatal("expected claims to be present")
	}

	if len(retrieved.Scopes) != 0 {
		t.Errorf("expected 0 scopes, got %d", len(retrieved.Scopes))
	}
}

func TestClaims_NilExtra(t *testing.T) {
	ctx := context.Background()
	claims := Claims{
		Subject: "user-123",
		Email:   "user@example.com",
		Extra:   nil,
	}

	ctx = WithClaims(ctx, claims)

	retrieved, ok := GetClaims(ctx)
	if !ok {
		t.Fatal("expected claims to be present")
	}

	if retrieved.Extra != nil {
		t.Error("expected Extra to be nil")
	}
}
