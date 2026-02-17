// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/DataDog/pup/pkg/agenthelp"
	"github.com/spf13/cobra"
)

var agentSchemaCompact bool

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent tooling: schema, guide, and diagnostics for AI coding assistants",
	Long: `Commands for AI coding assistants to interact with pup efficiently.

In agent mode (auto-detected or via --agent / FORCE_AGENT_MODE=1),
--help returns structured JSON schema instead of human-readable text.

COMMANDS:
  schema    Output the complete command schema as JSON
  guide     Output the comprehensive steering guide

EXAMPLES:
  # Get full JSON schema (all commands, flags, query syntax)
  pup agent schema

  # Get compact schema (command names and flags only, fewer tokens)
  pup agent schema --compact

  # Get the steering guide
  pup agent guide

  # Get guide for a specific domain
  pup agent guide logs`,
}

var agentSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Output command schema as JSON",
	Long: `Output the complete pup command schema as structured JSON.

Includes all commands, flags, query syntax, time formats, workflows,
best practices, and anti-patterns. This is the same output returned
by --help when running in agent mode.

FLAGS:
  --compact    Output minimal schema (command names and flags only)

EXAMPLES:
  pup agent schema
  pup agent schema --compact`,
	RunE: runAgentSchema,
}

var agentGuideCmd = &cobra.Command{
	Use:   "guide [domain]",
	Short: "Output the comprehensive steering guide",
	Long: `Output the pup steering guide for AI coding assistants.

Without arguments, outputs the full guide. With a domain argument,
outputs only the section relevant to that domain.

EXAMPLES:
  pup agent guide
  pup agent guide logs
  pup agent guide metrics
  pup agent guide monitors
  pup agent guide apm`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAgentGuide,
}

func init() {
	agentSchemaCmd.Flags().BoolVar(&agentSchemaCompact, "compact", false, "Output minimal schema (names + flags only)")

	agentCmd.AddCommand(agentSchemaCmd)
	agentCmd.AddCommand(agentGuideCmd)
}

func runAgentSchema(cmd *cobra.Command, args []string) error {
	root := cmd.Root()

	var data interface{}
	if agentSchemaCompact {
		data = agenthelp.GenerateCompactSchema(root)
	} else {
		data = agenthelp.GenerateSchema(root)
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	printOutput("%s\n", string(out))
	return nil
}

func runAgentGuide(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		printOutput("%s\n", agenthelp.GetGuideSection(args[0]))
		return nil
	}
	printOutput("%s\n", agenthelp.GetGuide())
	return nil
}

