// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/DataDog/pup/pkg/auth/callback"
	"github.com/DataDog/pup/pkg/auth/dcr"
	"github.com/DataDog/pup/pkg/auth/oauth"
	"github.com/DataDog/pup/pkg/auth/storage"
	"github.com/DataDog/pup/pkg/auth/types"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "OAuth2 authentication commands",
	Long: `Manage OAuth2 authentication with Datadog.

OAuth2 provides secure, browser-based authentication with better security than
API keys. It uses PKCE (Proof Key for Code Exchange) and Dynamic Client
Registration for maximum security.

AUTHENTICATION METHODS:
  Pup supports two authentication methods:

  1. OAuth2 (RECOMMENDED):
     - Browser-based login flow
     - Short-lived access tokens (1 hour)
     - Automatic token refresh
     - Per-installation credentials
     - Granular OAuth scopes
     - Better audit trail

  2. API Keys (LEGACY):
     - Environment variables (DD_API_KEY, DD_APP_KEY)
     - Long-lived credentials
     - Organization-wide access
     - Manual rotation required

OAUTH2 FEATURES:
  • PKCE Protection (S256): Prevents authorization code interception
  • Dynamic Client Registration: Unique credentials per installation
  • CSRF Protection: State parameter validation
  • Secure Storage: Tokens stored in ~/.config/pup/ with 0600 permissions
  • Auto Refresh: Tokens refresh automatically before expiration
  • Multi-Site: Separate credentials for each Datadog site

COMMANDS:
  login       Authenticate via browser with OAuth2
  status      Check current authentication status
  refresh     Manually refresh access token
  logout      Clear all stored credentials

OAUTH2 SCOPES:
  The following scopes are requested during login:
  • Dashboards: dashboards_read, dashboards_write
  • Monitors: monitors_read, monitors_write, monitors_downtime
  • APM: apm_read
  • SLOs: slos_read, slos_write, slos_corrections
  • Incidents: incident_read, incident_write
  • Synthetics: synthetics_read, synthetics_write
  • Security: security_monitoring_*
  • RUM: rum_apps_read, rum_apps_write
  • Infrastructure: hosts_read
  • Users: user_access_read, user_self_profile_read
  • Cases: cases_read, cases_write
  • Events: events_read
  • Logs: logs_read_data, logs_read_index_data
  • Metrics: metrics_read, timeseries_query
  • Usage: usage_read

EXAMPLES:
  # Login with OAuth2
  pup auth login

  # Check authentication status
  pup auth status

  # Refresh access token
  pup auth refresh

  # Logout and clear credentials
  pup auth logout

  # Login to different Datadog site
  DD_SITE=datadoghq.eu pup auth login

MULTI-SITE SUPPORT:
  Each Datadog site maintains separate credentials:

  DD_SITE=datadoghq.com pup auth login     # US1 (default)
  DD_SITE=datadoghq.eu pup auth login      # EU1
  DD_SITE=us3.datadoghq.com pup auth login # US3
  DD_SITE=us5.datadoghq.com pup auth login # US5
  DD_SITE=ap1.datadoghq.com pup auth login # AP1

TOKEN STORAGE:
  Credentials are stored in:
  • ~/.config/pup/tokens_<site>.json - OAuth2 tokens
  • ~/.config/pup/client_<site>.json - DCR client credentials

  File permissions are set to 0600 (read/write owner only).

SECURITY:
  • Tokens never logged or printed
  • PKCE prevents code interception
  • State parameter prevents CSRF
  • Unique client per installation
  • Tokens auto-refresh before expiration

For detailed OAuth2 documentation, see: docs/OAUTH2.md`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login via OAuth2",
	Long: `Authenticate with Datadog using OAuth2 browser-based login.

This command initiates a secure OAuth2 authentication flow:
  1. Registers OAuth client with Datadog (first time only)
  2. Generates PKCE challenge for security
  3. Starts local callback server
  4. Opens browser to Datadog authorization page
  5. Waits for you to approve requested scopes
  6. Exchanges authorization code for tokens
  7. Stores tokens securely in ~/.config/pup/

OAUTH2 FLOW:
  ┌─────────────────────────────────────────────────┐
  │ 1. Check for existing client registration       │
  │ 2. Register new client if needed (DCR)          │
  │ 3. Generate PKCE challenge (S256)               │
  │ 4. Generate state for CSRF protection           │
  │ 5. Start local callback server                  │
  │ 6. Open browser to Datadog auth page            │
  │ 7. User approves 36 OAuth scopes                │
  │ 8. Datadog redirects to callback with code      │
  │ 9. Exchange code for tokens (with PKCE)         │
  │ 10. Store tokens securely                       │
  └─────────────────────────────────────────────────┘

EXAMPLES:
  # Login to default site (datadoghq.com)
  pup auth login

  # Login to EU site
  DD_SITE=datadoghq.eu pup auth login

  # Login to US3 site
  DD_SITE=us3.datadoghq.com pup auth login

WHAT HAPPENS:
  1. A local HTTP server starts on http://127.0.0.1:<random-port>/callback
  2. Your browser opens to Datadog's authorization page
  3. You review and approve the requested OAuth scopes
  4. Datadog redirects back to the local callback server
  5. Access and refresh tokens are securely stored
  6. You're ready to use pup commands!

IF BROWSER DOESN'T OPEN:
  The command will print the authorization URL. Copy and paste it into
  your browser manually.

TIMEOUT:
  The callback server waits 5 minutes for authorization. If you don't
  complete the flow within 5 minutes, run the command again.

CREDENTIALS STORAGE:
  After successful login, credentials are stored in:
  • ~/.config/pup/tokens_<site>.json     (access & refresh tokens)
  • ~/.config/pup/client_<site>.json     (OAuth client credentials)

  Files have 0600 permissions (read/write owner only).

TOKEN LIFETIME:
  • Access Token: 1 hour (automatically refreshed)
  • Refresh Token: 30 days (used to get new access tokens)

SCOPES REQUESTED:
  The login flow requests 36 OAuth scopes covering:
  • Dashboards, Monitors, SLOs, Incidents
  • APM traces, RUM, Synthetics
  • Logs, Metrics, Events
  • Security monitoring
  • Infrastructure and hosts
  • User and team management
  • Cases and usage data

  See: pup auth --help for complete scope list`,
	RunE: runAuthLogin,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long: `Display current authentication status and token information.

This command checks your current OAuth2 authentication status including:
• Whether you're authenticated
• Which Datadog site you're authenticated with
• When your access token expires
• Whether you have a valid refresh token

EXAMPLES:
  # Check authentication status
  pup auth status

  # Check status for specific site
  DD_SITE=datadoghq.eu pup auth status

  # Check status and parse with jq
  pup auth status | jq '.authenticated'

OUTPUT:
  {
    "authenticated": true,
    "site": "datadoghq.com",
    "expires_at": "2024-02-03T14:30:00Z",
    "token_type": "Bearer",
    "has_refresh": true,
    "status": "valid"
  }

OUTPUT FIELDS:
  • authenticated: Boolean - whether you have valid credentials
  • site: The Datadog site you're authenticated with
  • expires_at: When the access token expires (ISO 8601)
  • token_type: Token type (always "Bearer")
  • has_refresh: Whether a refresh token is available
  • status: "valid", "expired", or "missing"

STATUS VALUES:
  • valid: Access token is valid and not expired
  • expired: Access token has expired (run 'pup auth refresh')
  • missing: No credentials found (run 'pup auth login')

WHEN TOKEN IS EXPIRED:
  If your token is expired, you'll see:

    ⚠️  Token expired
    Run 'pup auth refresh' to refresh or 'pup auth login' to re-authenticate

  The access token automatically refreshes when making API calls, but you can
  manually refresh with 'pup auth refresh'.

WHEN NOT AUTHENTICATED:
  If you're not authenticated, you'll see:

    ❌ Not authenticated
    Run 'pup auth login' to authenticate

  This means no OAuth2 credentials were found. Run 'pup auth login' to start
  the authentication flow.

REFRESH TOKEN:
  The refresh token (valid for 30 days) is used to obtain new access tokens
  without requiring a new browser login. If the refresh token expires, you'll
  need to run 'pup auth login' again.`,
	RunE: runAuthStatus,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear tokens",
	Long:  `Clear stored OAuth2 tokens and client credentials.`,
	RunE:  runAuthLogout,
}

var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh access token",
	Long:  `Manually refresh the OAuth2 access token using the refresh token.`,
	RunE:  runAuthRefresh,
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authRefreshCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	site := cfg.Site
	fmt.Printf("🔐 Starting OAuth2 login for site: %s\n\n", site)

	// Initialize storage (auto-detects keychain vs file)
	store, err := storage.GetStorage(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Check for existing client credentials
	creds, err := store.LoadClientCredentials(site)
	if err != nil {
		return fmt.Errorf("failed to load client credentials: %w", err)
	}

	// Start callback server
	callbackServer, err := callback.NewServer()
	if err != nil {
		return fmt.Errorf("failed to create callback server: %w", err)
	}

	if err := callbackServer.Start(); err != nil {
		return fmt.Errorf("failed to start callback server: %w", err)
	}
	defer func() {
		if err := callbackServer.Stop(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to stop callback server: %v\n", err)
		}
	}()

	redirectURI := callbackServer.RedirectURI()
	fmt.Printf("📡 Callback server started on: %s\n", redirectURI)

	// Register client if needed
	if creds == nil {
		fmt.Println("📝 Registering new OAuth2 client...")
		dcrClient := dcr.NewClient(site)
		creds, err = dcrClient.Register(redirectURI, types.DefaultScopes())
		if err != nil {
			return fmt.Errorf("failed to register client: %w", err)
		}

		// Save client credentials
		if err := store.SaveClientCredentials(site, creds); err != nil {
			return fmt.Errorf("failed to save client credentials: %w", err)
		}
		fmt.Println("✓ Client registered successfully")
	} else {
		fmt.Println("✓ Using existing client registration")
	}

	// Generate PKCE challenge
	pkce, err := oauth.GeneratePKCEChallenge()
	if err != nil {
		return fmt.Errorf("failed to generate PKCE challenge: %w", err)
	}

	// Generate state for CSRF protection
	state, err := oauth.GenerateState()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	// Build authorization URL
	oauthClient := oauth.NewClient(site)
	authURL := oauthClient.BuildAuthorizationURL(
		creds.ClientID,
		redirectURI,
		state,
		pkce,
		types.DefaultScopes(),
	)

	// Open browser
	fmt.Println("\n🌐 Opening browser for authentication...")
	fmt.Printf("If the browser doesn't open, visit: %s\n\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("⚠️  Could not open browser automatically: %v\n", err)
		fmt.Printf("Please open this URL manually: %s\n\n", authURL)
	}

	// Wait for callback
	fmt.Println("⏳ Waiting for authorization...")
	result, err := callbackServer.WaitForCallback(5 * time.Minute)
	if err != nil {
		return fmt.Errorf("failed to receive callback: %w", err)
	}

	// Check for OAuth error
	if result.Error != "" {
		return fmt.Errorf("OAuth error: %s - %s", result.Error, result.ErrorDescription)
	}

	// Validate callback
	if err := oauthClient.ValidateCallback(result.Code, result.State, state); err != nil {
		return fmt.Errorf("invalid callback: %w", err)
	}

	// Exchange code for tokens
	fmt.Println("🔄 Exchanging authorization code for tokens...")
	dcrClient := dcr.NewClient(site)
	tokens, err := dcrClient.ExchangeCode(result.Code, redirectURI, pkce.Verifier, creds)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	// Save tokens
	if err := store.SaveTokens(site, tokens); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	fmt.Println("\n✅ Login successful!")
	expiresAt := time.Unix(tokens.IssuedAt+tokens.ExpiresIn, 0)
	fmt.Printf("   Access token expires: %s\n", expiresAt.Format(time.RFC3339))
	fmt.Printf("   Token stored in: %s\n", storage.GetStorageDescription())

	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	site := cfg.Site

	// Initialize storage
	store, err := storage.GetStorage(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Load tokens
	tokens, err := store.LoadTokens(site)
	if err != nil {
		return fmt.Errorf("failed to load tokens: %w", err)
	}

	if tokens == nil {
		fmt.Println("❌ Not authenticated")
		fmt.Println("   Run 'pup auth login' to authenticate")
		return nil
	}

	// Check if expired
	expired := tokens.IsExpired()

	status := map[string]interface{}{
		"authenticated": !expired,
		"site":          site,
		"expires_at":    time.Unix(tokens.IssuedAt+tokens.ExpiresIn, 0).Format(time.RFC3339),
		"token_type":    tokens.TokenType,
		"has_refresh":   tokens.RefreshToken != "",
	}

	if expired {
		status["status"] = "expired"
		fmt.Println("⚠️  Token expired")
		fmt.Println("   Run 'pup auth refresh' to refresh or 'pup auth login' to re-authenticate")
	} else {
		status["status"] = "valid"
		expiresAt := time.Unix(tokens.IssuedAt+tokens.ExpiresIn, 0)
		timeLeft := time.Until(expiresAt)
		fmt.Printf("✅ Authenticated for site: %s\n", site)
		fmt.Printf("   Token expires in: %s\n", timeLeft.Round(time.Second))
	}

	output, err := formatter.FormatOutput(status, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}

	fmt.Printf("\n%s\n", output)
	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	site := cfg.Site

	// Initialize storage
	store, err := storage.GetStorage(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Delete tokens
	if err := store.DeleteTokens(site); err != nil {
		return fmt.Errorf("failed to delete tokens: %w", err)
	}

	// Delete client credentials
	if err := store.DeleteClientCredentials(site); err != nil {
		return fmt.Errorf("failed to delete client credentials: %w", err)
	}

	fmt.Printf("✅ Logged out from site: %s\n", site)
	fmt.Println("   All tokens and credentials have been removed")

	return nil
}

func runAuthRefresh(cmd *cobra.Command, args []string) error {
	site := cfg.Site

	// Initialize storage
	store, err := storage.GetStorage(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Load tokens
	tokens, err := store.LoadTokens(site)
	if err != nil {
		return fmt.Errorf("failed to load tokens: %w", err)
	}

	if tokens == nil || tokens.RefreshToken == "" {
		return fmt.Errorf("no refresh token available - please run 'pup auth login'")
	}

	// Load client credentials
	creds, err := store.LoadClientCredentials(site)
	if err != nil {
		return fmt.Errorf("failed to load client credentials: %w", err)
	}

	if creds == nil {
		return fmt.Errorf("no client credentials found - please run 'pup auth login'")
	}

	// Refresh tokens
	fmt.Println("🔄 Refreshing access token...")
	dcrClient := dcr.NewClient(site)
	newTokens, err := dcrClient.RefreshToken(tokens.RefreshToken, creds)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Save new tokens
	if err := store.SaveTokens(site, newTokens); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	fmt.Println("✅ Token refreshed successfully!")
	expiresAt := time.Unix(newTokens.IssuedAt+newTokens.ExpiresIn, 0)
	fmt.Printf("   New token expires: %s\n", expiresAt.Format(time.RFC3339))

	return nil
}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
