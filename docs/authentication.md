# Authentication

FullMCP provides pluggable authentication mechanisms for securing MCP servers. Multiple authentication providers are available with support for API keys, JWT tokens, and OAuth 2.0.

## Table of Contents

- [Overview](#overview)
- [API Key Authentication](#api-key-authentication)
- [JWT Authentication](#jwt-authentication)
- [OAuth 2.0](#oauth-20)
- [Custom Authentication](#custom-authentication)
- [Best Practices](#best-practices)

## Overview

Authentication in FullMCP follows these principles:

- **Pluggable**: Swap authentication providers easily
- **Middleware-based**: Integrate with HTTP handlers seamlessly
- **Claims-based**: Standardized claims structure
- **Scope support**: Fine-grained permission control

### Claims Structure

All authentication providers return standardized claims:

```go
type Claims struct {
    Subject   string            // User identifier
    Email     string            // User email
    Scopes    []string          // Permissions
    ExpiresAt time.Time         // Expiration time
    IssuedAt  time.Time         // Issue time
    Extra     map[string]string // Additional custom claims
}
```

## API Key Authentication

Simple key-based authentication for internal services or testing.

### Setup

```go
import (
    "github.com/jmcarbo/fullmcp/auth"
    "github.com/jmcarbo/fullmcp/auth/apikey"
    "github.com/jmcarbo/fullmcp/transport/http"
)

func main() {
    // Create API key provider
    authProvider := apikey.New()

    // Add keys with claims
    authProvider.AddKey("secret-key-123", auth.Claims{
        Subject: "user-1",
        Email:   "user@example.com",
        Scopes:  []string{"read", "write"},
    })

    authProvider.AddKey("readonly-key-456", auth.Claims{
        Subject: "service-account",
        Email:   "service@example.com",
        Scopes:  []string{"read"},
    })

    // Create MCP server
    srv := server.New("secure-server")

    // Create HTTP server with auth middleware
    httpServer := http.NewServer(":8080", srv,
        http.WithMiddleware(authProvider.Middleware()),
    )

    log.Fatal(httpServer.ListenAndServe())
}
```

### Client Usage

```go
import "github.com/jmcarbo/fullmcp/transport/http"

// Create authenticated client
transport := http.New("http://localhost:8080",
    http.WithAPIKey("secret-key-123"),
)

client := client.New(transport)
```

### Adding Keys Programmatically

```go
// Add key
authProvider.AddKey("new-key", auth.Claims{
    Subject: "new-user",
    Email:   "new@example.com",
    Scopes:  []string{"read"},
})

// Remove key
authProvider.RemoveKey("old-key")

// Update key (remove old, add new)
authProvider.RemoveKey("key-to-update")
authProvider.AddKey("key-to-update", auth.Claims{
    Subject: "updated-user",
    Scopes:  []string{"read", "write", "admin"},
})
```

### Key Rotation

```go
// Rotate keys periodically
ticker := time.NewTicker(24 * time.Hour)
go func() {
    for range ticker.C {
        // Generate new key
        newKey := generateSecureKey()

        // Add new key
        authProvider.AddKey(newKey, claims)

        // Remove old key after grace period
        time.AfterFunc(1*time.Hour, func() {
            authProvider.RemoveKey(oldKey)
        })

        // Notify users of new key
        notifyKeyRotation(newKey)
    }
}()
```

## JWT Authentication

Token-based authentication with signature verification and expiration.

### Setup

```go
import (
    "github.com/jmcarbo/fullmcp/auth/jwt"
    "time"
)

func main() {
    // Generate or load signing key
    key, err := jwt.GenerateRandomKey(32)
    if err != nil {
        log.Fatal(err)
    }

    // Create JWT provider with options
    jwtProvider := jwt.New(key,
        jwt.WithIssuer("mcp-server"),
        jwt.WithExpiration(24*time.Hour),
        jwt.WithRefreshEnabled(true),
    )

    // Use as middleware
    httpServer := http.NewServer(":8080", srv,
        http.WithMiddleware(jwtProvider.Middleware()),
    )

    log.Fatal(httpServer.ListenAndServe())
}
```

### Creating Tokens

```go
// Create token for user
token, err := jwtProvider.CreateToken(
    "user123",                    // subject
    "user@example.com",           // email
    []string{"read", "write"},    // scopes
    map[string]string{            // extra claims
        "role": "admin",
        "org":  "acme-corp",
    },
)

if err != nil {
    log.Fatal(err)
}

// Token is ready to use
fmt.Println("Token:", token)
```

### Validating Tokens

```go
// Validate token
claims, err := jwtProvider.ValidateToken(ctx, token)
if err != nil {
    log.Printf("Invalid token: %v", err)
    return
}

// Check scopes
if !claims.HasScope("write") {
    log.Printf("User lacks write permission")
    return
}

// Access claims
fmt.Printf("User: %s (%s)\n", claims.Subject, claims.Email)
fmt.Printf("Scopes: %v\n", claims.Scopes)
```

### Client Usage

```go
// Create client with JWT
transport := http.New("http://localhost:8080",
    http.WithJWT(token),
)

client := client.New(transport)
```

### Token Refresh

```go
// Enable refresh tokens
jwtProvider := jwt.New(key,
    jwt.WithRefreshEnabled(true),
    jwt.WithRefreshExpiration(7*24*time.Hour),
)

// Create access and refresh tokens
accessToken, refreshToken, err := jwtProvider.CreateTokenPair(
    "user123",
    "user@example.com",
    []string{"read", "write"},
    nil,
)

// Later, refresh the access token
newAccessToken, err := jwtProvider.RefreshToken(ctx, refreshToken)
```

### Persisting Keys

```go
// Save key to file
if err := jwt.SaveKeyToFile(key, "jwt-key.bin"); err != nil {
    log.Fatal(err)
}

// Load key from file
key, err := jwt.LoadKeyFromFile("jwt-key.bin")
if err != nil {
    log.Fatal(err)
}
```

## OAuth 2.0

OAuth 2.0 integration with support for Google, GitHub, and Azure.

### Google OAuth

```go
import "github.com/jmcarbo/fullmcp/auth/oauth"

func main() {
    // Create OAuth provider
    provider := oauth.New(
        oauth.Google,
        "client-id.apps.googleusercontent.com",
        "client-secret",
        "http://localhost:8080/callback",
        []string{"email", "profile"},
    )

    // Generate authorization URL
    http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
        state := generateRandomState()
        authURL := provider.AuthCodeURL(state)
        http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
    })

    // Handle OAuth callback
    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        state := r.URL.Query().Get("state")

        // Verify state to prevent CSRF
        if !verifyState(state) {
            http.Error(w, "Invalid state", http.StatusBadRequest)
            return
        }

        // Exchange code for token
        claims, err := provider.HandleCallback(r.Context(), code)
        if err != nil {
            http.Error(w, "Auth failed", http.StatusUnauthorized)
            return
        }

        // Create session
        createSession(w, claims)

        http.Redirect(w, r, "/dashboard", http.StatusFound)
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### GitHub OAuth

```go
provider := oauth.New(
    oauth.GitHub,
    "github-client-id",
    "github-client-secret",
    "http://localhost:8080/callback",
    []string{"user:email"},
)
```

### Azure OAuth

```go
provider := oauth.New(
    oauth.Azure,
    "azure-client-id",
    "azure-client-secret",
    "http://localhost:8080/callback",
    []string{"openid", "profile", "email"},
    oauth.WithAzureTenant("tenant-id"),
)
```

### Complete OAuth Flow

```go
func setupOAuth() {
    provider := oauth.New(
        oauth.Google,
        os.Getenv("GOOGLE_CLIENT_ID"),
        os.Getenv("GOOGLE_CLIENT_SECRET"),
        "http://localhost:8080/callback",
        []string{"email", "profile"},
    )

    // Store for state verification
    states := make(map[string]time.Time)
    var statesMux sync.Mutex

    // Login endpoint
    http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
        state := generateSecureState()

        statesMux.Lock()
        states[state] = time.Now()
        statesMux.Unlock()

        authURL := provider.AuthCodeURL(state)
        http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
    })

    // Callback endpoint
    http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        state := r.URL.Query().Get("state")

        // Verify state
        statesMux.Lock()
        createdAt, exists := states[state]
        delete(states, state)
        statesMux.Unlock()

        if !exists || time.Since(createdAt) > 10*time.Minute {
            http.Error(w, "Invalid or expired state", http.StatusBadRequest)
            return
        }

        // Exchange for token and get claims
        claims, err := provider.HandleCallback(r.Context(), code)
        if err != nil {
            log.Printf("OAuth error: %v", err)
            http.Error(w, "Authentication failed", http.StatusUnauthorized)
            return
        }

        // Create session
        sessionToken := createSession(claims)

        // Set cookie
        http.SetCookie(w, &http.Cookie{
            Name:     "session",
            Value:    sessionToken,
            Path:     "/",
            HttpOnly: true,
            Secure:   true,
            SameSite: http.SameSiteLaxMode,
            MaxAge:   86400, // 24 hours
        })

        http.Redirect(w, r, "/dashboard", http.StatusFound)
    })

    // Clean up expired states periodically
    go func() {
        ticker := time.NewTicker(1 * time.Hour)
        for range ticker.C {
            statesMux.Lock()
            for state, createdAt := range states {
                if time.Since(createdAt) > 1*time.Hour {
                    delete(states, state)
                }
            }
            statesMux.Unlock()
        }
    }()
}
```

## Custom Authentication

Implement custom authentication providers.

### Provider Interface

```go
type Provider interface {
    Authenticate(ctx context.Context, token string) (*Claims, error)
    Middleware() func(http.Handler) http.Handler
}
```

### Example: Database-Backed Authentication

```go
type DatabaseAuth struct {
    db *sql.DB
}

func NewDatabaseAuth(db *sql.DB) *DatabaseAuth {
    return &DatabaseAuth{db: db}
}

func (da *DatabaseAuth) Authenticate(ctx context.Context, token string) (*auth.Claims, error) {
    var claims auth.Claims
    var scopes string

    err := da.db.QueryRowContext(ctx, `
        SELECT user_id, email, scopes, expires_at
        FROM auth_tokens
        WHERE token = ? AND expires_at > NOW()
    `, token).Scan(&claims.Subject, &claims.Email, &scopes, &claims.ExpiresAt)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("invalid token")
        }
        return nil, err
    }

    claims.Scopes = strings.Split(scopes, ",")
    return &claims, nil
}

func (da *DatabaseAuth) Middleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if token == "" {
                http.Error(w, "Missing authorization", http.StatusUnauthorized)
                return
            }

            // Remove "Bearer " prefix if present
            token = strings.TrimPrefix(token, "Bearer ")

            claims, err := da.Authenticate(r.Context(), token)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            // Add claims to context
            ctx := auth.WithClaims(r.Context(), claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Example: Multi-Provider Authentication

```go
type MultiAuth struct {
    providers []auth.Provider
}

func NewMultiAuth(providers ...auth.Provider) *MultiAuth {
    return &MultiAuth{providers: providers}
}

func (ma *MultiAuth) Authenticate(ctx context.Context, token string) (*auth.Claims, error) {
    var lastErr error

    for _, provider := range ma.providers {
        claims, err := provider.Authenticate(ctx, token)
        if err == nil {
            return claims, nil
        }
        lastErr = err
    }

    return nil, fmt.Errorf("authentication failed: %w", lastErr)
}

func (ma *MultiAuth) Middleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            if token == "" {
                http.Error(w, "Missing authorization", http.StatusUnauthorized)
                return
            }

            claims, err := ma.Authenticate(r.Context(), token)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            ctx := auth.WithClaims(r.Context(), claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Usage
apiKeyAuth := apikey.New()
jwtAuth := jwt.New(key)

multiAuth := NewMultiAuth(apiKeyAuth, jwtAuth)
httpServer := http.NewServer(":8080", srv,
    http.WithMiddleware(multiAuth.Middleware()),
)
```

## Best Practices

### Secure Key Generation

```go
// ✅ Good: Cryptographically secure random key
key, err := jwt.GenerateRandomKey(32)

// ❌ Bad: Predictable key
key := []byte("my-secret-key")
```

### Token Storage

```go
// ✅ Good: Store in secure, httpOnly cookie
http.SetCookie(w, &http.Cookie{
    Name:     "token",
    Value:    token,
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteStrictMode,
})

// ❌ Bad: Store in localStorage (vulnerable to XSS)
```

### Scope Validation

```go
// ✅ Good: Check specific scopes
if !claims.HasScope("admin") {
    return fmt.Errorf("insufficient permissions")
}

// ❌ Bad: Assume authenticated = authorized
```

### Error Messages

```go
// ✅ Good: Generic error messages
return fmt.Errorf("authentication failed")

// ❌ Bad: Reveals information
return fmt.Errorf("user john@example.com not found")
```

### HTTPS Only

```go
// ✅ Good: Enforce HTTPS in production
if !r.TLS && os.Getenv("ENV") == "production" {
    http.Error(w, "HTTPS required", http.StatusForbidden)
    return
}
```

### Rate Limiting

```go
// Implement rate limiting for authentication endpoints
rateLimiter := NewRateLimiter(10, time.Minute) // 10 attempts per minute

func authHandler(w http.ResponseWriter, r *http.Request) {
    ip := getClientIP(r)

    if !rateLimiter.Allow(ip) {
        http.Error(w, "Too many requests", http.StatusTooManyRequests)
        return
    }

    // Proceed with authentication
}
```

### Token Expiration

```go
// ✅ Good: Short-lived access tokens with refresh tokens
jwt.WithExpiration(15 * time.Minute)
jwt.WithRefreshExpiration(7 * 24 * time.Hour)

// ❌ Bad: Long-lived tokens
jwt.WithExpiration(365 * 24 * time.Hour)
```

## Testing

### Mock Authentication

```go
type MockAuth struct {
    claims *auth.Claims
    err    error
}

func (ma *MockAuth) Authenticate(ctx context.Context, token string) (*auth.Claims, error) {
    return ma.claims, ma.err
}

func (ma *MockAuth) Middleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := auth.WithClaims(r.Context(), ma.claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Usage in tests
mockAuth := &MockAuth{
    claims: &auth.Claims{
        Subject: "test-user",
        Scopes:  []string{"read", "write"},
    },
}
```

## Related Documentation

- [Transports](./transports.md)
- [Middleware](./middleware.md)
- [Architecture Overview](./architecture.md)
