// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/DataDog/pup/internal/version"
	"github.com/DataDog/pup/pkg/client"
	"github.com/DataDog/pup/pkg/config"
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

	// Dependency injection points for testing
	clientFactory = defaultClientFactory
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
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format (json, table, yaml)")
	rootCmd.PersistentFlags().BoolVarP(&autoApprove, "yes", "y", false, "Skip confirmation prompts (auto-approve all operations)")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(metricsCmd)
	rootCmd.AddCommand(monitorsCmd)
	rootCmd.AddCommand(dashboardsCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(tracesCmd)
	rootCmd.AddCommand(slosCmd)
	rootCmd.AddCommand(incidentsCmd)
	rootCmd.AddCommand(rumCmd)
	rootCmd.AddCommand(cicdCmd)
	rootCmd.AddCommand(vulnerabilitiesCmd)
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

// printOutput writes formatted output (for testing)
func printOutput(format string, a ...any) {
	_, _ = fmt.Fprintf(outputWriter, format, a...)
}

// readConfirmation reads user confirmation from input
func readConfirmation() (string, error) {
	scanner := bufio.NewScanner(inputReader)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", scanner.Err()
}

// formatAPIError creates user-friendly error messages for API errors
func formatAPIError(operation string, err error, response any) error {
	type httpResponse interface {
		StatusCode() int
	}

	if r, ok := response.(httpResponse); ok && r != nil {
		statusCode := r.StatusCode()
		baseMsg := fmt.Sprintf("failed to %s: %v (status: %d)", operation, err, statusCode)

		switch {
		case statusCode >= 500:
			// 5xx Server errors
			return fmt.Errorf("%s\n\nThe Datadog API is experiencing issues. Please try again later or check https://status.datadoghq.com/", baseMsg)
		case statusCode == 429:
			// Rate limiting
			return fmt.Errorf("%s\n\nYou are being rate limited. Please wait a moment and try again.", baseMsg)
		case statusCode == 403:
			// Forbidden
			return fmt.Errorf("%s\n\nAccess denied. Verify your API/App keys have the required permissions.", baseMsg)
		case statusCode == 401:
			// Unauthorized
			return fmt.Errorf("%s\n\nAuthentication failed. Run 'pup auth login' or verify your DD_API_KEY and DD_APP_KEY.", baseMsg)
		case statusCode == 404:
			// Not found
			return fmt.Errorf("%s\n\nResource not found. Verify the ID or check if the resource was deleted.", baseMsg)
		case statusCode >= 400:
			// Other 4xx client errors
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
		fmt.Printf("  API Key: %s...%s\n", cfg.APIKey[:8], cfg.APIKey[len(cfg.APIKey)-4:])
		fmt.Printf("  App Key: %s...%s\n", cfg.AppKey[:8], cfg.AppKey[len(cfg.AppKey)-4:])
		fmt.Println("\nConnection test successful!")

		return nil
	},
}
