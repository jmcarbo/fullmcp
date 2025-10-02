package oauth21

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPKCEChallenge_Verification(t *testing.T) {
	challenge, err := GeneratePKCEChallenge()
	if err != nil {
		t.Fatalf("failed to generate challenge: %v", err)
	}

	// Manually verify the challenge
	hash := sha256.Sum256([]byte(challenge.CodeVerifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	if challenge.CodeChallenge != expectedChallenge {
		t.Errorf("challenge mismatch: expected %s, got %s", expectedChallenge, challenge.CodeChallenge)
	}
}

func TestProvider_HandleCallback_Success(t *testing.T) {
	// Create a mock token server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"test-token","token_type":"Bearer"}`))
	}))
	defer tokenServer.Close()

	// Create a mock user info server
	userInfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"123","email":"test@example.com"}`))
	}))
	defer userInfoServer.Close()

	provider := New(
		Google,
		"test-client",
		"test-secret",
		"http://localhost/callback",
		[]string{"email"},
		WithCustomEndpoint(tokenServer.URL+"/auth", tokenServer.URL+"/token"),
		WithUserInfoURL(userInfoServer.URL),
	)

	// Generate PKCE challenge
	challenge, err := GeneratePKCEChallenge()
	if err != nil {
		t.Fatalf("failed to generate challenge: %v", err)
	}

	// Generate auth URL to store verifier
	state := "test-state"
	_ = provider.AuthCodeURLWithPKCE(state, challenge)

	// Verify verifier was stored
	if provider.pkceVerifiers[state] != challenge.CodeVerifier {
		t.Error("verifier was not stored correctly")
	}
}

func TestProvider_ValidateRedirectURI_Strict(t *testing.T) {
	redirectURI := "http://localhost:8080/callback"
	provider := New(Google, "client", "secret", redirectURI, []string{"email"})

	tests := []struct {
		name      string
		uri       string
		shouldErr bool
	}{
		{"exact match", "http://localhost:8080/callback", false},
		{"extra path", "http://localhost:8080/callback/extra", true},
		{"different port", "http://localhost:8081/callback", true},
		{"different protocol", "https://localhost:8080/callback", true},
		{"different host", "http://example.com:8080/callback", true},
		{"missing path", "http://localhost:8080", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidateRedirectURI(tt.uri)
			if tt.shouldErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestProvider_ProviderTypes(t *testing.T) {
	tests := []struct {
		name             string
		providerType     ProviderType
		expectedUserInfo string
		expectedSubKey   string
		expectedEmailKey string
	}{
		{
			name:             "Google",
			providerType:     Google,
			expectedUserInfo: "https://www.googleapis.com/oauth2/v2/userinfo",
			expectedSubKey:   "id",
			expectedEmailKey: "email",
		},
		{
			name:             "GitHub",
			providerType:     GitHub,
			expectedUserInfo: "https://api.github.com/user",
			expectedSubKey:   "id",
			expectedEmailKey: "email",
		},
		{
			name:             "Azure",
			providerType:     Azure,
			expectedSubKey:   "sub",
			expectedEmailKey: "email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := New(
				tt.providerType,
				"client",
				"secret",
				"http://localhost/callback",
				[]string{"email"},
			)

			if tt.expectedUserInfo != "" && provider.userInfoURL != tt.expectedUserInfo {
				t.Errorf("expected userInfoURL %s, got %s", tt.expectedUserInfo, provider.userInfoURL)
			}

			if provider.subjectKey != tt.expectedSubKey {
				t.Errorf("expected subjectKey %s, got %s", tt.expectedSubKey, provider.subjectKey)
			}

			if provider.emailKey != tt.expectedEmailKey {
				t.Errorf("expected emailKey %s, got %s", tt.expectedEmailKey, provider.emailKey)
			}
		})
	}
}

func TestProvider_WithOptions(t *testing.T) {
	t.Run("WithVerifyEmail", func(t *testing.T) {
		provider := New(
			Google,
			"client",
			"secret",
			"http://localhost/callback",
			[]string{"email"},
			WithVerifyEmail(true),
		)

		if !provider.verifyEmail {
			t.Error("expected verifyEmail to be true")
		}
	})

	t.Run("WithScopeMapping", func(t *testing.T) {
		mapping := map[string][]string{
			"role": {"admin", "user"},
		}

		provider := New(
			Google,
			"client",
			"secret",
			"http://localhost/callback",
			[]string{"email"},
			WithScopeMapping(mapping),
		)

		if len(provider.scopeMapping) != 1 {
			t.Errorf("expected 1 scope mapping, got %d", len(provider.scopeMapping))
		}
	})

	t.Run("WithCustomEndpoint", func(t *testing.T) {
		authURL := "https://custom.com/auth"
		tokenURL := "https://custom.com/token"

		provider := New(
			Google,
			"client",
			"secret",
			"http://localhost/callback",
			[]string{"email"},
			WithCustomEndpoint(authURL, tokenURL),
		)

		if provider.config.Endpoint.AuthURL != authURL {
			t.Errorf("expected auth URL %s, got %s", authURL, provider.config.Endpoint.AuthURL)
		}

		if provider.config.Endpoint.TokenURL != tokenURL {
			t.Errorf("expected token URL %s, got %s", tokenURL, provider.config.Endpoint.TokenURL)
		}
	})

	t.Run("WithUserInfoURL", func(t *testing.T) {
		customURL := "https://custom.com/userinfo"

		provider := New(
			Google,
			"client",
			"secret",
			"http://localhost/callback",
			[]string{"email"},
			WithUserInfoURL(customURL),
		)

		if provider.userInfoURL != customURL {
			t.Errorf("expected userInfoURL %s, got %s", customURL, provider.userInfoURL)
		}
	})
}

func TestProvider_Authenticate_ValidationErrors(t *testing.T) {
	provider := New(Google, "client", "secret", "http://localhost/callback", []string{"email"})

	tests := []struct {
		name        string
		credentials interface{}
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "string credentials",
			credentials: "simple-string",
			shouldError: true,
			errorMsg:    "invalid credentials type",
		},
		{
			name:        "missing state",
			credentials: map[string]string{"code": "abc"},
			shouldError: true,
			errorMsg:    "code and state are required",
		},
		{
			name:        "missing code",
			credentials: map[string]string{"state": "xyz"},
			shouldError: true,
			errorMsg:    "code and state are required",
		},
		{
			name:        "empty values",
			credentials: map[string]string{"code": "", "state": ""},
			shouldError: true,
			errorMsg:    "code and state are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider.Authenticate(context.Background(), tt.credentials)
			if !tt.shouldError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Error("expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			}
		})
	}
}

func TestProvider_Middleware_EmailVerification(t *testing.T) {
	provider := New(
		Google,
		"client",
		"secret",
		"http://localhost/callback",
		[]string{"email"},
		WithVerifyEmail(true),
	)

	// Create a mock user info server that returns user without email
	userInfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"123"}`)) // No email field
	}))
	defer userInfoServer.Close()

	provider.userInfoURL = userInfoServer.URL

	middleware := provider.Middleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should fail because email verification is required but email is missing
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestProvider_StrictRedirectURIEnabled(t *testing.T) {
	provider := New(Google, "client", "secret", "http://localhost/callback", []string{"email"})

	if !provider.strictRedirectURI {
		t.Error("expected strictRedirectURI to be enabled by default (OAuth 2.1 requirement)")
	}
}

func TestPKCEChallenge_S256Method(t *testing.T) {
	challenge, err := GeneratePKCEChallenge()
	if err != nil {
		t.Fatalf("failed to generate challenge: %v", err)
	}

	if challenge.Method != "S256" {
		t.Errorf("expected method S256, got %s", challenge.Method)
	}
}

func TestProvider_AuthCodeURLWithPKCE_Parameters(t *testing.T) {
	provider := New(Google, "client", "secret", "http://localhost/callback", []string{"email"})
	challenge, _ := GeneratePKCEChallenge()
	state := "test-state"

	authURL := provider.AuthCodeURLWithPKCE(state, challenge)

	// Check that URL contains PKCE parameters
	if !strings.Contains(authURL, "code_challenge=") {
		t.Error("auth URL missing code_challenge parameter")
	}

	if !strings.Contains(authURL, "code_challenge_method=S256") {
		t.Error("auth URL missing code_challenge_method parameter")
	}

	if !strings.Contains(authURL, "state="+state) {
		t.Error("auth URL missing state parameter")
	}
}
