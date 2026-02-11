// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DataDog/pup/pkg/config"
	"github.com/spf13/cobra"
)

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Create shortcuts for pup commands",
	Long: `Aliases can be used to make shortcuts for pup commands or to compose multiple commands.

Aliases are stored in ~/.config/pup/config.yml and can be used like any other pup command.

EXAMPLES:
  # Create an alias for a complex logs query
  pup alias set prod-errors "logs search --query='status:error' --tag='env:prod'"

  # Use the alias
  pup prod-errors

  # List all aliases
  pup alias list

  # Delete an alias
  pup alias delete prod-errors

  # Import aliases from a file
  pup alias import aliases.yml

AVAILABLE COMMANDS:
  set         Create a shortcut for a pup command
  list        List your aliases
  delete      Delete set aliases
  import      Import aliases from a YAML file

Run 'pup alias <command> --help' for more information about a command.`,
}

var aliasSetCmd = &cobra.Command{
	Use:   "set <name> <command>",
	Short: "Create a shortcut for a pup command",
	Long: `Create an alias for a pup command.

The alias name should be a single word (no spaces), and the command should be
the full pup command you want to run (without the 'pup' prefix).

EXAMPLES:
  # Create an alias for listing production monitors
  pup alias set prod-monitors "monitors list --tag='env:production'"

  # Create an alias for a common metrics query
  pup alias set cpu-avg "metrics query --query='avg:system.cpu.user{*}' --from='1h'"

  # Create an alias for RUM applications
  pup alias set rum-apps "rum apps list"

  # Create an alias with multiple parameters
  pup alias set search-errors "logs search --query='status:error'"

USAGE:
  Once an alias is set, you can use it like any other pup command:

  $ pup prod-monitors
  # Executes: pup monitors list --tag='env:production'

NOTES:
  - Alias names cannot contain spaces or special characters
  - Aliases are stored in ~/.config/pup/config.yml
  - You can override an existing alias by setting it again
  - Use quotes around commands with spaces or special characters`,
	Args: cobra.ExactArgs(2),
	RunE: runAliasSet,
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your aliases",
	Long: `List all configured aliases.

This command displays all aliases stored in ~/.config/pup/config.yml,
showing the alias name and the command it expands to.

EXAMPLES:
  # List all aliases
  pup alias list

OUTPUT FORMAT:
  Aliases are displayed in alphabetical order:

  prod-errors => logs search --query='status:error' --tag='env:prod'
  prod-monitors => monitors list --tag='env:production'
  rum-apps => rum apps list`,
	Args: cobra.NoArgs,
	RunE: runAliasList,
}

var aliasDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete set aliases",
	Long: `Delete one or more aliases.

This command removes aliases from your configuration file.
You can delete multiple aliases at once by providing multiple names.

EXAMPLES:
  # Delete a single alias
  pup alias delete prod-errors

  # Delete multiple aliases
  pup alias delete prod-errors prod-monitors rum-apps

NOTES:
  - The command will fail if any of the specified aliases don't exist
  - Changes are saved to ~/.config/pup/config.yml`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAliasDelete,
}

var aliasImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import aliases from a YAML file",
	Long: `Import aliases from a YAML file.

The file should be in the same format as ~/.config/pup/config.yml:

  aliases:
    prod-errors: logs search --query='status:error' --tag='env:prod'
    prod-monitors: monitors list --tag='env:production'
    rum-apps: rum apps list

EXAMPLES:
  # Import aliases from a file
  pup alias import team-aliases.yml

  # Import from a different location
  pup alias import /path/to/aliases.yml

NOTES:
  - Imported aliases will overwrite existing aliases with the same name
  - The import file must be valid YAML
  - Changes are saved to ~/.config/pup/config.yml`,
	Args: cobra.ExactArgs(1),
	RunE: runAliasImport,
}

func init() {
	aliasCmd.AddCommand(aliasSetCmd)
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasDeleteCmd)
	aliasCmd.AddCommand(aliasImportCmd)
}

func runAliasSet(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("requires exactly 2 arguments: <name> <command>")
	}
	name := args[0]
	command := args[1]

	// Validate alias name (no spaces, no special chars except hyphens and underscores)
	if !isValidAliasName(name) {
		return fmt.Errorf("invalid alias name: '%s' (only letters, numbers, hyphens, and underscores are allowed)", name)
	}

	// Check if alias name conflicts with existing commands
	if isReservedCommand(name) {
		return fmt.Errorf("alias name '%s' conflicts with an existing pup command", name)
	}

	// Set the alias
	if err := config.SetAlias(name, command); err != nil {
		return fmt.Errorf("failed to set alias: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("✓ Added alias '%s'\n", name)
	fmt.Printf("  %s => %s\n", name, command)
	fmt.Printf("\nStored in: %s\n", configPath)

	return nil
}

func runAliasList(cmd *cobra.Command, args []string) error {
	aliases, err := config.LoadAliases()
	if err != nil {
		return fmt.Errorf("failed to load aliases: %w", err)
	}

	if len(aliases) == 0 {
		fmt.Println("No aliases configured.")
		fmt.Println("\nUse 'pup alias set <name> <command>' to create an alias.")
		return nil
	}

	// Sort aliases alphabetically
	names := make([]string, 0, len(aliases))
	for name := range aliases {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Printf("Aliases (%d):\n\n", len(aliases))
	for _, name := range names {
		fmt.Printf("  %s => %s\n", name, aliases[name])
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("\nStored in: %s\n", configPath)

	return nil
}

func runAliasDelete(cmd *cobra.Command, args []string) error {
	// Delete each alias
	for _, name := range args {
		if err := config.DeleteAlias(name); err != nil {
			return fmt.Errorf("failed to delete alias '%s': %w", name, err)
		}
		fmt.Printf("✓ Deleted alias '%s'\n", name)
	}

	return nil
}

func runAliasImport(cmd *cobra.Command, args []string) error {
	filepath := args[0]

	if err := config.ImportAliases(filepath); err != nil {
		return fmt.Errorf("failed to import aliases: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("✓ Imported aliases from %s\n", filepath)
	fmt.Printf("\nAliases stored in: %s\n", configPath)

	return nil
}

// isValidAliasName checks if an alias name is valid
func isValidAliasName(name string) bool {
	if name == "" {
		return false
	}

	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}

	return true
}

// isReservedCommand checks if a name conflicts with existing commands
// IMPORTANT: This validation prevents aliases from shadowing built-in commands.
// The list must be kept in sync with commands registered in root.go.
//
// Additionally, ExecuteWithArgs in root.go uses isBuiltinCommand() to check
// registered cobra commands at runtime, ensuring aliases ALWAYS execute after
// built-in commands, even if this validation is bypassed or new commands are added.
//
// DO NOT REMOVE OR WEAKEN THIS CHECK - it's a security/safety feature.
func isReservedCommand(name string) bool {
	// List of reserved command names
	reserved := []string{
		"alias", "auth", "version", "test", "help",
		"metrics", "monitors", "dashboards", "logs", "traces",
		"slos", "incidents", "rum", "cicd", "static-analysis",
		"downtime", "tags", "events", "on-call", "audit-logs",
		"api-keys", "app-keys", "infrastructure", "synthetics",
		"users", "notebooks", "security", "organizations",
		"service-catalog", "error-tracking", "scorecards",
		"usage", "cost", "data-governance", "obs-pipelines",
		"network", "cloud", "integrations", "misc",
		"investigations", "product-analytics", "cases", "apm",
	}

	// Convert to lowercase for case-insensitive comparison
	nameLower := strings.ToLower(name)
	for _, cmd := range reserved {
		if nameLower == cmd {
			return true
		}
	}

	return false
}
