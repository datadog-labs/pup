// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/pup/internal/version"
	"github.com/DataDog/pup/pkg/agenthelp"
	"github.com/DataDog/pup/pkg/client"
	"github.com/DataDog/pup/pkg/config"
	"github.com/DataDog/pup/pkg/formatter"
	"github.com/DataDog/pup/pkg/useragent"
	"github.com/spf13/cobra"
)

// defaultClientFactory is the production client factory
func defaultClientFactory(cfg *config.Config) (*client.Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return client.New(cfg)
}

var (
	cfg          *config.Config
	ddClient     *client.Client
	outputFormat string
	autoApprove  bool
	agentFlag    bool

	// Dependency injection points for testing
	clientFactory           = defaultClientFactory
	outputWriter  io.Writer = os.Stdout
	inputReader   io.Reader = os.Stdin
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "pup",
	Short: "Pup - Datadog API CLI wrapper",
	Long: `Pup is a Go-based command-line wrapper that provides easy interaction
with Datadog APIs. It supports both API key and OAuth2 authentication.`,
	Version:      version.Version,
	SilenceUsage: true, // Don't show usage on errors, only on --help or invalid args
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return ExecuteWithArgs(os.Args[1:])
}

// hasFlag checks if a flag is present in the args.
func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

// firstNonFlagArg returns the first argument that isn't a flag.
func firstNonFlagArg(args []string) string {
	for _, a := range args {
		if !isFlag(a) {
			return a
		}
	}
	return ""
}

// printAgentSchema generates and prints the JSON schema for the given subtree.
func printAgentSchema(subtree string) error {
	var data interface{}
	if subtree != "" {
		schema := agenthelp.GenerateSubtreeSchema(rootCmd, subtree)
		if schema != nil {
			data = schema
		}
	}
	if data == nil {
		s := agenthelp.GenerateSchema(rootCmd)
		data = &s
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}
	printOutput("%s\n", string(out))
	return nil
}

// ExecuteWithArgs executes the root command with the given arguments
func ExecuteWithArgs(args []string) error {
	// In agent mode, intercept --help/-h and return structured JSON schema instead
	// of cobra's human-oriented help text. This lets agents run the natural
	// "pup --help" or "pup logs --help" and get the machine-readable schema
	// without needing to know about a special flag.
	if (hasFlag(args, "--help") || hasFlag(args, "-h")) && useragent.IsAgentMode() {
		return printAgentSchema(firstNonFlagArg(args))
	}

	// IMPORTANT: Aliases are checked LAST to prevent overriding built-in commands.
	// This ensures that no alias can shadow an existing pup command, even if validation
	// is bypassed or a new command is added that conflicts with an existing alias.
	//
	// Priority order:
	// 1. Built-in commands (version, auth, metrics, etc.)
	// 2. Aliases (only if no built-in command matches)

	// Check if the first argument might be an alias
	// Only resolve as alias if it's NOT a built-in command
	if len(args) > 0 && !isFlag(args[0]) && !isBuiltinCommand(args[0]) {
		if aliasCommand, err := config.GetAlias(args[0]); err == nil {
			// Expand the alias by replacing args[0] with the alias command
			expandedArgs := expandAlias(aliasCommand, args[1:])
			rootCmd.SetArgs(expandedArgs)
			return rootCmd.Execute()
		}
	}

	// Not an alias or is a built-in command, execute normally
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

// expandAlias expands an alias command and appends additional arguments
func expandAlias(aliasCommand string, additionalArgs []string) []string {
	// Split the alias command into parts
	// Simple split by spaces (could be enhanced to handle quoted strings)
	parts := splitCommand(aliasCommand)

	// Append any additional arguments passed after the alias
	result := make([]string, 0, len(parts)+len(additionalArgs))
	result = append(result, parts...)
	result = append(result, additionalArgs...)

	return result
}

// splitCommand splits a command string by spaces, respecting quotes
func splitCommand(command string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, r := range command {
		switch {
		case r == '"' || r == '\'':
			if !inQuote {
				inQuote = true
				quoteChar = r
			} else if r == quoteChar {
				inQuote = false
				quoteChar = 0
			} else {
				current.WriteRune(r)
			}
		case r == ' ' && !inQuote:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// isFlag checks if a string is a flag (starts with -)
func isFlag(s string) bool {
	return len(s) > 0 && s[0] == '-'
}

// isBuiltinCommand checks if a command name matches a registered cobra command
// This ensures aliases cannot override built-in commands at runtime.
//
// CRITICAL SECURITY CHECK: This function is used in ExecuteWithArgs to ensure
// that built-in commands ALWAYS take precedence over aliases, even if:
// - Alias validation is bypassed (e.g., manual config.yml editing)
// - New commands are added after aliases are created
// - The reserved command list in alias.go becomes out of sync
//
// DO NOT REMOVE THIS CHECK - it prevents aliases from shadowing built-in commands.
func isBuiltinCommand(name string) bool {
	// Check if the command exists in rootCmd's registered commands
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == name {
			return true
		}
		// Also check aliases defined by cobra commands themselves
		for _, alias := range cmd.Aliases {
			if alias == name {
				return true
			}
		}
	}
	return false
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format (json, table, yaml)")
	rootCmd.PersistentFlags().BoolVarP(&autoApprove, "yes", "y", false, "Skip confirmation prompts (auto-approve all operations)")
	rootCmd.PersistentFlags().BoolVar(&agentFlag, "agent", false, "Enable agent mode (auto-detected for AI coding assistants)")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(aliasCmd)
	rootCmd.AddCommand(metricsCmd)
	rootCmd.AddCommand(monitorsCmd)
	rootCmd.AddCommand(dashboardsCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(tracesCmd)
	rootCmd.AddCommand(slosCmd)
	rootCmd.AddCommand(incidentsCmd)
	rootCmd.AddCommand(rumCmd)
	rootCmd.AddCommand(cicdCmd)
	rootCmd.AddCommand(staticAnalysisCmd)
	rootCmd.AddCommand(downtimeCmd)
	rootCmd.AddCommand(tagsCmd)
	rootCmd.AddCommand(eventsCmd)
	rootCmd.AddCommand(onCallCmd)
	rootCmd.AddCommand(auditLogsCmd)
	rootCmd.AddCommand(apiKeysCmd)
	rootCmd.AddCommand(appKeysCmd)
	rootCmd.AddCommand(infrastructureCmd)
	rootCmd.AddCommand(syntheticsCmd)
	rootCmd.AddCommand(usersCmd)
	rootCmd.AddCommand(notebooksCmd)
	rootCmd.AddCommand(securityCmd)
	rootCmd.AddCommand(organizationsCmd)
	rootCmd.AddCommand(serviceCatalogCmd)
	rootCmd.AddCommand(errorTrackingCmd)
	rootCmd.AddCommand(scorecardsCmd)
	rootCmd.AddCommand(usageCmd)
	rootCmd.AddCommand(costCmd)
	rootCmd.AddCommand(dataGovernanceCmd)
	rootCmd.AddCommand(obsPipelinesCmd)
	rootCmd.AddCommand(networkCmd)
	rootCmd.AddCommand(cloudCmd)
	rootCmd.AddCommand(integrationsCmd)
	rootCmd.AddCommand(miscCmd)
	rootCmd.AddCommand(investigationsCmd)
	rootCmd.AddCommand(productAnalyticsCmd)
	rootCmd.AddCommand(casesCmd)
	rootCmd.AddCommand(apmCmd)
	rootCmd.AddCommand(agentCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Set auto-approve from flag if specified
	if autoApprove {
		cfg.AutoApprove = true
		if err := os.Setenv("DD_CLI_AUTO_APPROVE", "true"); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to set DD_CLI_AUTO_APPROVE: %v\n", err)
		}
	}

	// Determine agent mode: explicit flag takes precedence, then auto-detect
	if agentFlag || useragent.IsAgentMode() {
		cfg.AgentMode = true
		cfg.AutoApprove = true
	}
}

// isAgentMode returns true if the current session is in agent mode.
func isAgentMode() bool {
	return cfg != nil && cfg.AgentMode
}

// getClient returns a configured Datadog client
func getClient() (*client.Client, error) {
	if ddClient != nil {
		return ddClient, nil
	}

	var err error
	ddClient, err = clientFactory(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return ddClient, nil
}

// getClientForEndpoint creates a client appropriate for the endpoint
// If the endpoint doesn't support OAuth, it forces API key authentication
func getClientForEndpoint(method, path string) (*client.Client, error) {
	// Check if this endpoint requires API keys
	if client.RequiresAPIKeyFallback(method, path) {
		// This endpoint doesn't support OAuth, use API keys
		if cfg.APIKey == "" || cfg.AppKey == "" {
			return nil, fmt.Errorf(
				"endpoint %s %s does not support OAuth authentication. "+
					"Please set DD_API_KEY and DD_APP_KEY environment variables",
				method, path,
			)
		}

		// Try to use the mocked factory if in test mode (allows test to fail intentionally)
		// This respects the clientFactory mock in tests
		c, err := clientFactory(cfg)
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	// Endpoint supports OAuth, use standard client
	return getClient()
}

// printOutput writes formatted output (for testing)
func printOutput(format string, a ...any) {
	_, _ = fmt.Fprintf(outputWriter, format, a...)
}

// formatAndPrint formats data with optional agent envelope and prints it.
// In agent mode, data is wrapped in an AgentEnvelope with metadata.
// In human mode, data is formatted using the standard formatter.
func formatAndPrint(data interface{}, meta *formatter.Metadata) error {
	if isAgentMode() {
		output, err := formatter.WrapForAgent(data, meta)
		if err != nil {
			return err
		}
		printOutput("%s\n", output)
		return nil
	}

	output, err := formatter.FormatOutput(data, formatter.OutputFormat(outputFormat))
	if err != nil {
		return err
	}
	printOutput("%s\n", output)
	return nil
}

// readConfirmation reads user confirmation from input
func readConfirmation() (string, error) {
	scanner := bufio.NewScanner(inputReader)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", scanner.Err()
}

// extractAPIErrorBody extracts the API response body from a
// datadog.GenericOpenAPIError. The datadog-api-client-go library consumes
// http.Response.Body during deserialization and stores the bytes in the error.
// Callers that try to re-read http.Response.Body will always get empty data.
func extractAPIErrorBody(err error) string {
	if err == nil {
		return ""
	}
	var apiErr datadog.GenericOpenAPIError
	if errors.As(err, &apiErr) {
		if body := apiErr.Body(); len(body) > 0 {
			return string(body)
		}
	}
	return ""
}

// formatAPIError creates user-friendly error messages for API errors.
// It extracts the API response body from GenericOpenAPIError when available
// and appends contextual guidance based on the HTTP status code.
// In agent mode, returns a structured JSON error.
func formatAPIError(operation string, err error, response any) error {
	// Extract status code from the response.
	// The datadog API client returns *http.Response (field) for some APIs
	// and types with StatusCode() method for others.
	statusCode := 0
	switch r := response.(type) {
	case *http.Response:
		if r != nil {
			statusCode = r.StatusCode
		}
	case interface{ StatusCode() int }:
		if r != nil {
			statusCode = r.StatusCode()
		}
	}

	if statusCode > 0 {
		apiBody := extractAPIErrorBody(err)

		// In agent mode, return structured JSON error
		if isAgentMode() {
			jsonErr, fmtErr := formatter.FormatAgentError(operation, statusCode, err.Error(), apiBody)
			if fmtErr == nil {
				return fmt.Errorf("%s", jsonErr)
			}
		}

		baseMsg := fmt.Sprintf("failed to %s: %v (status: %d)", operation, err, statusCode)

		// Include API response body if available
		if apiBody != "" {
			baseMsg = fmt.Sprintf("failed to %s: %v (status: %d)\nAPI Response: %s", operation, err, statusCode, apiBody)
		}

		switch {
		case statusCode >= 500:
			return fmt.Errorf("%s\n\nThe Datadog API is experiencing issues. Please try again later or check https://status.datadoghq.com/", baseMsg)
		case statusCode == 429:
			return fmt.Errorf("%s\n\nYou are being rate limited. Please wait a moment and try again.", baseMsg)
		case statusCode == 403:
			return fmt.Errorf("%s\n\nAccess denied. Verify your API/App keys have the required permissions.", baseMsg)
		case statusCode == 401:
			return fmt.Errorf("%s\n\nAuthentication failed. Run 'pup auth login' or verify your DD_API_KEY and DD_APP_KEY.", baseMsg)
		case statusCode == 404:
			return fmt.Errorf("%s\n\nResource not found. Verify the ID or check if the resource was deleted.", baseMsg)
		case statusCode >= 400:
			return fmt.Errorf("%s\n\nInvalid request. Check your parameters and try again.", baseMsg)
		default:
			return fmt.Errorf("%s", baseMsg)
		}
	}

	return fmt.Errorf("failed to %s: %v", operation, err)
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.BuildInfo())
	},
}

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test connection and credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cfg.Validate(); err != nil {
			return err
		}

		fmt.Println("Configuration is valid:")
		fmt.Printf("  Site: %s\n", cfg.Site)

		// Display API key info if present
		if len(cfg.APIKey) >= 12 {
			fmt.Printf("  API Key: %s...%s\n", cfg.APIKey[:8], cfg.APIKey[len(cfg.APIKey)-4:])
		} else if len(cfg.APIKey) > 0 {
			fmt.Printf("  API Key: %s (too short - may be invalid)\n", cfg.APIKey)
		} else {
			fmt.Println("  API Key: (not set - using OAuth2 or will prompt)")
		}

		// Display App key info if present
		if len(cfg.AppKey) >= 12 {
			fmt.Printf("  App Key: %s...%s\n", cfg.AppKey[:8], cfg.AppKey[len(cfg.AppKey)-4:])
		} else if len(cfg.AppKey) > 0 {
			fmt.Printf("  App Key: %s (too short - may be invalid)\n", cfg.AppKey)
		} else {
			fmt.Println("  App Key: (not set - using OAuth2 or will prompt)")
		}

		fmt.Println("\nConnection test successful!")

		return nil
	},
}
