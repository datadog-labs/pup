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

	"github.com/DataDog/fetch/pkg/auth/callback"
	"github.com/DataDog/fetch/pkg/auth/dcr"
	"github.com/DataDog/fetch/pkg/auth/oauth"
	"github.com/DataDog/fetch/pkg/auth/storage"
	"github.com/DataDog/fetch/pkg/auth/types"
	"github.com/DataDog/fetch/pkg/formatter"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "OAuth2 authentication commands",
	Long:  `Manage OAuth2 authentication with Datadog using PKCE flow and Dynamic Client Registration.`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login via OAuth2",
	Long:  `Authenticate with Datadog using OAuth2 browser-based login flow with PKCE protection.`,
	RunE:  runAuthLogin,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  `Display current authentication status and token information.`,
	RunE:  runAuthStatus,
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

	// Initialize storage
	store, err := storage.NewFileStorage()
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
	defer callbackServer.Stop()

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
	fmt.Printf("   Access token expires: %s\n", tokens.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("   Token stored in: ~/.config/fetch/\n")

	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	site := cfg.Site

	// Initialize storage
	store, err := storage.NewFileStorage()
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
		fmt.Println("   Run 'fetch auth login' to authenticate")
		return nil
	}

	// Check if expired
	expired := tokens.IsExpired()

	status := map[string]interface{}{
		"authenticated": !expired,
		"site":          site,
		"expires_at":    tokens.ExpiresAt.Format(time.RFC3339),
		"token_type":    tokens.TokenType,
		"has_refresh":   tokens.RefreshToken != "",
	}

	if expired {
		status["status"] = "expired"
		fmt.Println("⚠️  Token expired")
		fmt.Println("   Run 'fetch auth refresh' to refresh or 'fetch auth login' to re-authenticate")
	} else {
		status["status"] = "valid"
		timeLeft := time.Until(tokens.ExpiresAt)
		fmt.Printf("✅ Authenticated for site: %s\n", site)
		fmt.Printf("   Token expires in: %s\n", timeLeft.Round(time.Second))
	}

	output, err := formatter.ToJSON(status)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s\n", output)
	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	site := cfg.Site

	// Initialize storage
	store, err := storage.NewFileStorage()
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
	store, err := storage.NewFileStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Load tokens
	tokens, err := store.LoadTokens(site)
	if err != nil {
		return fmt.Errorf("failed to load tokens: %w", err)
	}

	if tokens == nil || tokens.RefreshToken == "" {
		return fmt.Errorf("no refresh token available - please run 'fetch auth login'")
	}

	// Load client credentials
	creds, err := store.LoadClientCredentials(site)
	if err != nil {
		return fmt.Errorf("failed to load client credentials: %w", err)
	}

	if creds == nil {
		return fmt.Errorf("no client credentials found - please run 'fetch auth login'")
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
	fmt.Printf("   New token expires: %s\n", newTokens.ExpiresAt.Format(time.RFC3339))

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
