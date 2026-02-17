// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package agenthelp

import (
	"github.com/DataDog/pup/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Schema is the top-level structure returned by --help in agent mode or 'pup agent schema'.
type Schema struct {
	Version       string            `json:"version"`
	Description   string            `json:"description"`
	Auth          AuthInfo          `json:"auth"`
	GlobalFlags   []FlagInfo        `json:"global_flags"`
	Commands      []CommandInfo     `json:"commands"`
	QuerySyntax   map[string]string `json:"query_syntax"`
	TimeFormats   TimeFormats       `json:"time_formats"`
	Workflows     []Workflow        `json:"workflows"`
	BestPractices []string          `json:"best_practices"`
	AntiPatterns  []string          `json:"anti_patterns"`
}

// AuthInfo describes authentication options.
type AuthInfo struct {
	OAuth   string `json:"oauth"`
	APIKeys string `json:"api_keys"`
}

// FlagInfo describes a command flag.
type FlagInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
}

// CommandInfo describes a command or subcommand.
type CommandInfo struct {
	Name        string        `json:"name"`
	FullPath    string        `json:"full_path"`
	Description string        `json:"description"`
	Flags       []FlagInfo    `json:"flags,omitempty"`
	Examples    []string      `json:"examples,omitempty"`
	ReadOnly    bool          `json:"read_only"`
	Subcommands []CommandInfo `json:"subcommands,omitempty"`
}

// TimeFormats describes supported time format options.
type TimeFormats struct {
	Relative []string `json:"relative"`
	Absolute []string `json:"absolute"`
	Examples []string `json:"examples"`
}

// Workflow describes a multi-step agent workflow.
type Workflow struct {
	Name  string   `json:"name"`
	Steps []string `json:"steps"`
}

// GenerateSchema builds the complete schema from a cobra command tree.
func GenerateSchema(root *cobra.Command) Schema {
	return Schema{
		Version:     version.Version,
		Description: "Pup - Datadog API CLI. Provides OAuth2 + API key authentication for querying metrics, logs, monitors, traces, and 30+ other Datadog API domains.",
		Auth: AuthInfo{
			OAuth:   "pup auth login",
			APIKeys: "Set DD_API_KEY + DD_APP_KEY + DD_SITE environment variables",
		},
		GlobalFlags:   extractGlobalFlags(root),
		Commands:      extractCommands(root, ""),
		QuerySyntax:   GetQuerySyntax(),
		TimeFormats:   GetTimeFormats(),
		Workflows:     GetWorkflows(),
		BestPractices: GetBestPractices(),
		AntiPatterns:  GetAntiPatterns(),
	}
}

// GenerateSubtreeSchema builds a schema for a specific subtree of commands.
func GenerateSubtreeSchema(root *cobra.Command, subtreeName string) *Schema {
	for _, cmd := range root.Commands() {
		if cmd.Name() == subtreeName {
			schema := Schema{
				Version:     version.Version,
				Description: cmd.Short,
				Auth: AuthInfo{
					OAuth:   "pup auth login",
					APIKeys: "Set DD_API_KEY + DD_APP_KEY + DD_SITE environment variables",
				},
				GlobalFlags:   extractGlobalFlags(root),
				Commands:      []CommandInfo{buildCommandInfo(cmd, subtreeName)},
				QuerySyntax:   filterQuerySyntax(subtreeName),
				TimeFormats:   GetTimeFormats(),
				BestPractices: GetBestPractices(),
				AntiPatterns:  GetAntiPatterns(),
			}
			return &schema
		}
	}
	return nil
}

// CompactSchema is a minimal schema with just command names and flags.
type CompactSchema struct {
	Version  string           `json:"version"`
	Commands []CompactCommand `json:"commands"`
}

// CompactCommand is a minimal command representation.
type CompactCommand struct {
	Name        string           `json:"name"`
	Flags       []string         `json:"flags,omitempty"`
	Subcommands []CompactCommand `json:"subcommands,omitempty"`
}

// GenerateCompactSchema builds a token-efficient schema.
func GenerateCompactSchema(root *cobra.Command) CompactSchema {
	return CompactSchema{
		Version:  version.Version,
		Commands: extractCompactCommands(root),
	}
}

func extractGlobalFlags(cmd *cobra.Command) []FlagInfo {
	var flags []FlagInfo
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		flags = append(flags, FlagInfo{
			Name:        "--" + f.Name,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
		})
	})
	return flags
}

func extractCommands(parent *cobra.Command, prefix string) []CommandInfo {
	var commands []CommandInfo
	for _, cmd := range parent.Commands() {
		if cmd.Hidden || cmd.Name() == "help" || cmd.Name() == "completion" {
			continue
		}
		fullPath := cmd.Name()
		if prefix != "" {
			fullPath = prefix + " " + cmd.Name()
		}
		commands = append(commands, buildCommandInfo(cmd, fullPath))
	}
	return commands
}

func buildCommandInfo(cmd *cobra.Command, fullPath string) CommandInfo {
	info := CommandInfo{
		Name:        cmd.Name(),
		FullPath:    fullPath,
		Description: cmd.Short,
		ReadOnly:    isReadOnlyCommand(cmd.Name()),
		Flags:       extractLocalFlags(cmd),
		Examples:    extractExamples(cmd),
	}

	for _, sub := range cmd.Commands() {
		if sub.Hidden || sub.Name() == "help" || sub.Name() == "completion" {
			continue
		}
		subPath := fullPath + " " + sub.Name()
		info.Subcommands = append(info.Subcommands, buildCommandInfo(sub, subPath))
	}

	return info
}

func extractLocalFlags(cmd *cobra.Command) []FlagInfo {
	var flags []FlagInfo
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		flags = append(flags, FlagInfo{
			Name:        "--" + f.Name,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
		})
	})
	return flags
}

func extractExamples(cmd *cobra.Command) []string {
	if cmd.Example == "" {
		return nil
	}
	return []string{cmd.Example}
}

func extractCompactCommands(parent *cobra.Command) []CompactCommand {
	var commands []CompactCommand
	for _, cmd := range parent.Commands() {
		if cmd.Hidden || cmd.Name() == "help" || cmd.Name() == "completion" {
			continue
		}
		cc := CompactCommand{Name: cmd.Name()}
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			cc.Flags = append(cc.Flags, "--"+f.Name)
		})
		cc.Subcommands = extractCompactCommands(cmd)
		commands = append(commands, cc)
	}
	return commands
}

// isReadOnlyCommand returns true for commands that only read data.
func isReadOnlyCommand(name string) bool {
	writeCommands := map[string]bool{
		"delete": true, "create": true, "update": true, "set": true,
		"import": true, "login": true, "logout": true,
	}
	return !writeCommands[name]
}

// filterQuerySyntax returns query syntax relevant to a specific domain.
func filterQuerySyntax(domain string) map[string]string {
	all := GetQuerySyntax()
	if syntax, ok := all[domain]; ok {
		return map[string]string{domain: syntax}
	}
	return all
}
