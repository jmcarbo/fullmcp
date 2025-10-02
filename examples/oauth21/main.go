// Package main demonstrates OAuth 2.1 authentication in MCP (2025-03-26 specification)
package main

import (
	"fmt"

	"github.com/jmcarbo/fullmcp/auth/oauth21"
)

func main() {
	fmt.Println("MCP OAuth 2.1 Authentication Example")
	fmt.Println("=====================================")
	fmt.Println()

	// Example 1: What is OAuth 2.1?
	fmt.Println("💡 OAuth 2.1 Overview")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("OAuth 2.1 is a consolidation of OAuth 2.0 security best practices,")
	fmt.Println("removing insecure flows and making PKCE mandatory for all clients.")
	fmt.Println()
	fmt.Println("Key Changes from OAuth 2.0:")
	fmt.Println("  ✓ PKCE now MANDATORY (not optional)")
	fmt.Println("  ✗ Implicit grant REMOVED")
	fmt.Println("  ✗ Resource Owner Password Credentials grant REMOVED")
	fmt.Println("  ✓ Strict redirect URI matching REQUIRED")
	fmt.Println("  ✓ Refresh token rotation RECOMMENDED")
	fmt.Println()

	// Example 2: OAuth 2.1 Features
	fmt.Println("🔐 OAuth 2.1 Compliance Features")
	fmt.Println("================================")
	fmt.Println()

	features := oauth21.GetOAuth21Features()
	fmt.Printf("  Mandatory PKCE:              %v\n", features.MandatoryPKCE)
	fmt.Printf("  Strict Redirect URI:         %v\n", features.StrictRedirectURI)
	fmt.Printf("  Implicit Grant Removed:      %v\n", features.ImplicitGrantRemoved)
	fmt.Printf("  Password Grant Removed:      %v\n", features.PasswordGrantRemoved)
	fmt.Printf("  Code Challenge Method:       %s\n", features.CodeChallengeMethod)
	fmt.Printf("  Min Verifier Length:         %d chars\n", features.MinimumVerifierLength)
	fmt.Printf("  Max Verifier Length:         %d chars\n", features.MaximumVerifierLength)
	fmt.Println()

	// Example 3: PKCE Flow
	fmt.Println("🔄 PKCE (Proof Key for Code Exchange)")
	fmt.Println("======================================")
	fmt.Println()
	fmt.Println("PKCE protects against authorization code interception attacks.")
	fmt.Println()
	fmt.Println("Flow:")
	fmt.Println("  1. Client generates code_verifier (random string)")
	fmt.Println("  2. Client computes code_challenge = SHA256(code_verifier)")
	fmt.Println("  3. Client sends code_challenge in authorization request")
	fmt.Println("  4. Server stores code_challenge")
	fmt.Println("  5. Client exchanges code + code_verifier for token")
	fmt.Println("  6. Server verifies SHA256(code_verifier) == code_challenge")
	fmt.Println()

	// Example 4: Generate PKCE Challenge
	fmt.Println("🔑 Generate PKCE Challenge")
	fmt.Println("==========================")
	fmt.Println()

	challenge, err := oauth21.GeneratePKCEChallenge()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("  Code Verifier:  %s\n", challenge.CodeVerifier[:20]+"...")
	fmt.Printf("  Code Challenge: %s\n", challenge.CodeChallenge[:20]+"...")
	fmt.Printf("  Method:         %s (SHA-256)\n", challenge.Method)
	fmt.Println()
	fmt.Println("Code verifier is kept secret by the client")
	fmt.Println("Code challenge is sent to authorization server")
	fmt.Println()

	// Example 5: Authorization Flow
	fmt.Println("📋 Authorization Flow with PKCE")
	fmt.Println("===============================")
	fmt.Println()

	provider := oauth21.New(
		oauth21.Google,
		"your-client-id",
		"your-client-secret",
		"http://localhost:8080/callback",
		[]string{"email", "profile"},
	)

	state := "random-state-string"
	authURL := provider.AuthCodeURLWithPKCE(state, challenge)

	fmt.Println("Step 1: Generate Authorization URL")
	fmt.Printf("  URL: %s\n", authURL[:60]+"...")
	fmt.Println()
	fmt.Println("Parameters included:")
	fmt.Println("  • client_id")
	fmt.Println("  • redirect_uri")
	fmt.Println("  • response_type=code")
	fmt.Println("  • scope")
	fmt.Println("  • state")
	fmt.Println("  • code_challenge       (PKCE)")
	fmt.Println("  • code_challenge_method (PKCE)")
	fmt.Println()

	fmt.Println("Step 2: User Authorization")
	fmt.Println("  → User redirected to authorization URL")
	fmt.Println("  → User logs in and grants permission")
	fmt.Println("  → Server redirects back with code")
	fmt.Println()

	fmt.Println("Step 3: Token Exchange with PKCE")
	fmt.Println("  POST /token")
	fmt.Println("  {")
	fmt.Println("    grant_type: authorization_code,")
	fmt.Println("    code: <authorization_code>,")
	fmt.Println("    redirect_uri: <same_as_before>,")
	fmt.Println("    client_id: <client_id>,")
	fmt.Println("    code_verifier: <code_verifier>  ← PKCE verification")
	fmt.Println("  }")
	fmt.Println()

	// Example 6: Provider Setup
	fmt.Println("🔧 Provider Setup Examples")
	fmt.Println("==========================")
	fmt.Println()

	fmt.Println("Google:")
	fmt.Print(`
  provider := oauth21.New(
      oauth21.Google,
      "client-id",
      "client-secret",
      "http://localhost:8080/callback",
      []string{"email", "profile"},
  )
`)

	fmt.Println("GitHub:")
	fmt.Print(`
  provider := oauth21.New(
      oauth21.GitHub,
      "client-id",
      "client-secret",
      "http://localhost:8080/callback",
      []string{"user", "repo"},
  )
`)

	fmt.Println("Custom Provider:")
	fmt.Print(`
  provider := oauth21.New(
      oauth21.Azure,
      "client-id",
      "client-secret",
      "http://localhost:8080/callback",
      []string{"openid", "email"},
      oauth21.WithCustomEndpoint(
          "https://custom.com/authorize",
          "https://custom.com/token",
      ),
      oauth21.WithUserInfoURL("https://custom.com/userinfo"),
  )
`)

	// Example 7: Redirect URI Validation
	fmt.Println("🎯 Strict Redirect URI Matching")
	fmt.Println("================================")
	fmt.Println()
	fmt.Println("OAuth 2.1 requires EXACT string matching (no wildcards):")
	fmt.Println()
	fmt.Println("✓ Allowed:")
	fmt.Println("  Registered:   http://localhost:8080/callback")
	fmt.Println("  Provided:     http://localhost:8080/callback")
	fmt.Println("  Result:       MATCH")
	fmt.Println()
	fmt.Println("✗ Rejected:")
	fmt.Println("  Registered:   http://localhost:8080/callback")
	fmt.Println("  Provided:     http://localhost:8080/callback/extra")
	fmt.Println("  Result:       MISMATCH")
	fmt.Println()
	fmt.Println("✗ Rejected:")
	fmt.Println("  Registered:   http://localhost:8080/callback")
	fmt.Println("  Provided:     http://localhost:8081/callback")
	fmt.Println("  Result:       MISMATCH (different port)")
	fmt.Println()

	// Example 8: Security Benefits
	fmt.Println("🛡️  Security Benefits")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("1. PKCE Protection:")
	fmt.Println("   • Prevents authorization code interception")
	fmt.Println("   • Protects against man-in-the-middle attacks")
	fmt.Println("   • No client secret needed for public clients")
	fmt.Println()
	fmt.Println("2. Strict Redirect URI:")
	fmt.Println("   • Prevents redirect URI manipulation")
	fmt.Println("   • Eliminates open redirect vulnerabilities")
	fmt.Println("   • No partial matching exploits")
	fmt.Println()
	fmt.Println("3. Removed Insecure Flows:")
	fmt.Println("   • No Implicit grant (token in URL)")
	fmt.Println("   • No Password grant (credentials in request)")
	fmt.Println("   • Forces secure authorization code flow")
	fmt.Println()

	// Example 9: Client Types
	fmt.Println("📱 Client Types")
	fmt.Println("===============")
	fmt.Println()
	fmt.Println("Public Clients (PKCE mandatory):")
	fmt.Println("  • Single Page Applications (SPAs)")
	fmt.Println("  • Mobile applications")
	fmt.Println("  • Desktop applications")
	fmt.Println("  → Cannot securely store client secret")
	fmt.Println("  → PKCE provides protection")
	fmt.Println()
	fmt.Println("Confidential Clients (PKCE recommended):")
	fmt.Println("  • Web servers")
	fmt.Println("  • Backend services")
	fmt.Println("  → Can securely store client secret")
	fmt.Println("  → PKCE provides additional security")
	fmt.Println()

	// Example 10: Implementation Code
	fmt.Println("💻 Implementation Example")
	fmt.Println("=========================")
	fmt.Println()
	fmt.Println("Complete OAuth 2.1 flow:")
	//nolint:govet // Example code contains format directives
	fmt.Print(`
  // 1. Generate PKCE challenge
  challenge, _ := oauth21.GeneratePKCEChallenge()

  // 2. Generate authorization URL
  state := generateRandomState()
  authURL := provider.AuthCodeURLWithPKCE(state, challenge)

  // 3. Redirect user to authURL
  http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)

  // 4. Handle callback
  func handleCallback(w http.ResponseWriter, r *http.Request) {
      code := r.URL.Query().Get("code")
      state := r.URL.Query().Get("state")

      // 5. Exchange code for token (with PKCE)
      token, err := provider.ExchangeWithPKCE(ctx, code, state)
      if err != nil {
          // Handle error
      }

      // 6. Get user info
      claims, err := provider.ValidateToken(ctx, token.AccessToken)
      if err != nil {
          // Handle error
      }

      // User is authenticated
      log.Printf("User: ` + "%" + `s (` + "%" + `s)", claims.Subject, claims.Email)
  }
`)

	// Example 11: Migration from OAuth 2.0
	fmt.Println("🔄 Migration from OAuth 2.0")
	fmt.Println("============================")
	fmt.Println()
	fmt.Println("Changes needed:")
	fmt.Println()
	fmt.Println("1. Add PKCE to all flows:")
	fmt.Println("   Before: provider.AuthCodeURL(state)")
	fmt.Println("   After:  provider.AuthCodeURLWithPKCE(state, challenge)")
	fmt.Println()
	fmt.Println("2. Update token exchange:")
	fmt.Println("   Before: provider.Exchange(ctx, code)")
	fmt.Println("   After:  provider.ExchangeWithPKCE(ctx, code, state)")
	fmt.Println()
	fmt.Println("3. Remove insecure flows:")
	fmt.Println("   Remove: Implicit grant")
	fmt.Println("   Remove: Password grant")
	fmt.Println()
	fmt.Println("4. Enforce strict redirect URIs:")
	fmt.Println("   Update: Exact string matching required")
	fmt.Println()

	// Example 12: Best Practices
	fmt.Println("📋 Best Practices")
	fmt.Println("=================")
	fmt.Println()
	fmt.Println("State Parameter:")
	fmt.Println("  ✓ Generate cryptographically secure random state")
	fmt.Println("  ✓ Store state in session")
	fmt.Println("  ✓ Validate state on callback")
	fmt.Println("  ✓ Use once and discard")
	fmt.Println()
	fmt.Println("PKCE:")
	fmt.Println("  ✓ Use S256 method (SHA-256)")
	fmt.Println("  ✓ Generate new verifier for each flow")
	fmt.Println("  ✓ Store verifier securely")
	fmt.Println("  ✓ Clear verifier after exchange")
	fmt.Println()
	fmt.Println("Redirect URIs:")
	fmt.Println("  ✓ Pre-register all redirect URIs")
	fmt.Println("  ✓ Use HTTPS in production")
	fmt.Println("  ✓ Validate exact match")
	fmt.Println("  ✗ No wildcards or patterns")
	fmt.Println()
	fmt.Println("Token Handling:")
	fmt.Println("  ✓ Store tokens securely")
	fmt.Println("  ✓ Use short-lived access tokens")
	fmt.Println("  ✓ Implement token refresh")
	fmt.Println("  ✓ Rotate refresh tokens")
	fmt.Println()

	// Example 13: Common Errors
	fmt.Println("⚠️  Common Errors")
	fmt.Println("=================")
	fmt.Println()
	fmt.Println("Error: invalid_request")
	fmt.Println("  Cause: Missing code_challenge parameter")
	fmt.Println("  Fix:   Add PKCE parameters to authorization URL")
	fmt.Println()
	fmt.Println("Error: invalid_grant")
	fmt.Println("  Cause: code_verifier doesn't match code_challenge")
	fmt.Println("  Fix:   Ensure same verifier used for exchange")
	fmt.Println()
	fmt.Println("Error: redirect_uri_mismatch")
	fmt.Println("  Cause: Redirect URI doesn't exactly match")
	fmt.Println("  Fix:   Use exact string from registration")
	fmt.Println()

	fmt.Println("✨ OAuth 2.1 demonstration complete!")
	fmt.Println()
	fmt.Println("Key Takeaways:")
	fmt.Println("  1. PKCE is mandatory for all clients")
	fmt.Println("  2. Use exact redirect URI matching")
	fmt.Println("  3. Implicit and Password grants removed")
	fmt.Println("  4. S256 method recommended for PKCE")
	fmt.Println("  5. More secure by default")
	fmt.Println("  6. Compatible with OAuth 2.0 with best practices")
}
