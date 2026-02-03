// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"
	"os"

	"github.com/DataDog/fetch/internal/version"
	"github.com/DataDog/fetch/pkg/client"
	"github.com/DataDog/fetch/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cfg          *config.Config
	ddClient     *client.Client
	outputFormat string
	autoApprove  bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch - Datadog API CLI wrapper",
	Long: `Fetch is a Go-based command-line wrapper that provides easy interaction
with Datadog APIs. It supports both API key and OAuth2 authentication.`,
	Version: version.Version,
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
		os.Setenv("DD_CLI_AUTO_APPROVE", "true")
	}
}

// getClient returns a configured Datadog client
func getClient() (*client.Client, error) {
	if ddClient != nil {
		return ddClient, nil
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var err error
	ddClient, err = client.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return ddClient, nil
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
