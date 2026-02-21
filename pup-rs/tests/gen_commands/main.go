// gen_commands generates test command lists for both Go pup and Rust pup-rs
// by discovering every leaf subcommand and emitting test invocations.
//
// The Go CLI outputs JSON for --help (agent mode), so we parse the JSON tree.
// The Rust CLI outputs text for --help (clap), so we parse text recursively.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const testBodyFile = "/tmp/pup_test_body.json"

// testArgs maps leaf command names to stub arguments for mock server testing.
var testArgs = map[string]string{
	"list":              "",
	"get":               "test-id-123",
	"delete":            "test-id-123",
	"create":            "--file " + testBodyFile,
	"update":            "test-id-123 --file " + testBodyFile,
	"search":            "--query '*'",
	"query":             "--query test",
	"archive":           "test-id-123",
	"unarchive":         "test-id-123",
	"cancel":            "test-id-123",
	"trigger":           "test-id-123",
	"status":            "",
	"schema":            "",
	"guide":             "",
	"activate":          "test-id-123",
	"deactivate":        "test-id-123",
	"import":            testBodyFile,
	"unlink":            "test-id-123",
	"link":              "test-id-123 --file " + testBodyFile,
	"register":          "test-id-123",
	"unregister":        "test-id-123",
	"bulk-export":       "",
	"patch-deployment":  "test-id-123 --file " + testBodyFile,
	"send":              "--file " + testBodyFile,
	"events-send":       "--file " + testBodyFile,
	"set":               "myalias 'monitors list'",
	"add":               "test-host tag1 tag2",
	"summary":           "",
	"hourly":            "",
	"projected":         "",
	"by-org":            "--start-month 2024-01-01T00:00:00Z",
	"attribution":       "--start 2024-01",
	"ip-ranges":         "",
	"accounts":          "",
	"instances":         "",
	"users":             "test-instance",
	"assignment-groups": "test-instance",
	"business-services": "test-instance",
	"delete-account":    "test-id-123",
	"versions":          "",
	"configure":         "--file " + testBodyFile,
	"upgrade":           "--file " + testBodyFile,
	"branch-summary":    "--repo test --branch main",
	"commit-summary":    "--repo test --commit abc123",
	"scanner-rules":     "",
	"remove":            "test-team test-user",
	"create-issue":      "test-id-123 --file " + testBodyFile,
	"create-ticket":     "test-id-123 --file " + testBodyFile,
}

// intIDParents are domains where IDs are integers (i64) not strings.
var intIDParents = map[string]bool{
	"monitors": true, "events": true, "notebooks": true,
}

// fullPathOverrides map full command paths (space-separated) to specific args.
// These override the leaf-based testArgs when the command has non-standard args.
var fullPathOverrides = map[string]string{
	// incidents settings takes no ID for get/update
	"incidents settings get":    "",
	"incidents settings update": "--file " + testBodyFile,
	// incidents handles update takes only --file
	"incidents handles update":  "--file " + testBodyFile,
	// organizations get takes no args
	"organizations get":         "",
	// logs query needs --query not --view-name
	"logs query":                "--query test",
	// metrics query needs specific flags
	"metrics query":             "--query 'avg:system.cpu.user{*}'",
	// usage hourly needs --start
	"usage hourly":              "--start 1d",
	// cost by-org needs start-month
	"cost by-org":               "--start-month 2024-01",
	// slos status needs an id
	"slos status":               "test-id-123",
	// rum events has args not subcommands
	"rum events":                "",
	// rum retention-filters needs --app-id
	"rum retention-filters list":   "test-id-123",
	"rum retention-filters get":    "test-id-123 test-filter-123",
	"rum retention-filters create": "test-id-123 --file " + testBodyFile,
	"rum retention-filters update": "test-id-123 test-filter-123 --file " + testBodyFile,
	"rum retention-filters delete": "test-id-123 test-filter-123",
	// rum heatmaps query needs view-name
	"rum heatmaps query":          "test-view",
	// rum playlists get needs integer
	"rum playlists get":           "12345",
	// synthetics suites delete takes IDs
	"synthetics suites delete":    "test-id-123",
	// cicd pipelines list vs get
	"cicd pipelines list":         "",
	"cicd pipelines get":          "--pipeline-id test-id-123",
	// cicd dora patch-deployment
	"cicd dora patch-deployment":  "test-id-123 --file " + testBodyFile,
	// cicd flaky-tests search
	"cicd flaky-tests search":     "",
	"cicd flaky-tests update":     "--file " + testBodyFile,
	// apm entities/dependencies need time args
	"apm entities list":           "",
	"apm dependencies list":       "",
	// infrastructure hosts get
	"infrastructure hosts get":    "test-host-123",
	// cloud oci products list
	"cloud oci products list":     "",
	// integrations slack/pagerduty/webhooks
	"integrations slack list":     "",
	"integrations pagerduty list": "",
	"integrations webhooks list":  "",
}

// skipCommands are meta commands that should not be tested.
var skipCommands = map[string]bool{
	"help": true, "completions": true, "completion": true,
	"version": true, "auth": true, "test": true, "__complete": true,
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: gen_commands <go-binary> <rust-binary>\n")
		os.Exit(1)
	}
	// Create minimal test body file for --file arguments
	os.WriteFile(testBodyFile, []byte(`{}`), 0644)

	goBin := os.Args[1]
	rustBin := os.Args[2]

	goCommands := discoverGoCommands(goBin)
	rustCommands := discoverRustCommands(rustBin, 5)

	writeCommandFile("go_commands.txt", goCommands)
	writeCommandFile("rust_commands.txt", rustCommands)

	fmt.Fprintf(os.Stderr, "Generated %d Go commands, %d Rust commands\n",
		len(goCommands), len(rustCommands))
}

// --- Go CLI: JSON tree parsing ---

type jsonHelpSchema struct {
	Commands []jsonCommand `json:"commands"`
}

type jsonCommand struct {
	Name        string        `json:"name"`
	Subcommands []jsonCommand `json:"subcommands"`
}

// discoverGoCommands parses the Go CLI's JSON --help output and walks the
// tree to find all leaf commands.
func discoverGoCommands(binary string) []string {
	cmd := exec.Command(binary, "--help")
	cmd.Env = helpEnv()
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Go CLI --help failed: %v\n", err)
		return nil
	}

	var schema jsonHelpSchema
	if err := json.Unmarshal(output, &schema); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Go CLI --help is not JSON, falling back to text\n")
		return discoverRustCommands(binary, 3)
	}

	var commands []string
	for _, c := range schema.Commands {
		if skipCommands[c.Name] {
			continue
		}
		walkJSONTree(binary, []string{c.Name}, c.Subcommands, &commands)
	}
	return commands
}

// walkJSONTree recursively walks the JSON command tree to find leaf commands.
func walkJSONTree(binary string, path []string, subs []jsonCommand, commands *[]string) {
	if len(subs) == 0 {
		// Leaf command
		cmdStr := generateTestCommand(binary, path)
		if cmdStr != "" {
			*commands = append(*commands, cmdStr)
		}
		return
	}
	for _, sub := range subs {
		if skipCommands[sub.Name] {
			continue
		}
		walkJSONTree(binary, append(append([]string{}, path...), sub.Name), sub.Subcommands, commands)
	}
}

// --- Rust CLI: text help parsing ---

// discoverRustCommands walks --help text output recursively.
func discoverRustCommands(binary string, maxDepth int) []string {
	var commands []string
	discoverText(binary, nil, maxDepth, &commands)
	return commands
}

func discoverText(binary string, prefix []string, depth int, commands *[]string) {
	if depth <= 0 {
		return
	}

	args := append(append([]string{}, prefix...), "--help")
	cmd := exec.Command(binary, args...)
	cmd.Env = helpEnv()
	output, _ := cmd.CombinedOutput()

	subs := parseTextSubcommands(string(output))
	if len(subs) == 0 {
		cmdStr := generateTestCommand(binary, prefix)
		if cmdStr != "" {
			*commands = append(*commands, cmdStr)
		}
		return
	}

	for _, sub := range subs {
		if skipCommands[sub] {
			continue
		}
		discoverText(binary, append(append([]string{}, prefix...), sub), depth-1, commands)
	}
}

var subcommandRe = regexp.MustCompile(`^\s{2,4}(\S+)\s`)

func parseTextSubcommands(helpOutput string) []string {
	var subs []string
	inCommands := false
	scanner := bufio.NewScanner(strings.NewReader(helpOutput))
	for scanner.Scan() {
		line := scanner.Text()
		lower := strings.ToLower(strings.TrimSpace(line))

		if strings.Contains(lower, "commands:") || strings.Contains(lower, "subcommands:") {
			inCommands = true
			continue
		}

		if inCommands {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" ||
				strings.HasPrefix(trimmed, "Options:") ||
				strings.HasPrefix(trimmed, "FLAGS:") ||
				strings.HasPrefix(trimmed, "Flags:") ||
				strings.HasPrefix(trimmed, "Arguments:") ||
				strings.HasPrefix(trimmed, "ARGUMENTS:") {
				inCommands = false
				continue
			}
			m := subcommandRe.FindStringSubmatch(line)
			if m != nil {
				subs = append(subs, m[1])
			}
		}
	}
	return subs
}

// --- Common helpers ---

func helpEnv() []string {
	return append(os.Environ(),
		"DD_API_KEY=test-key",
		"DD_APP_KEY=test-app-key",
		"DD_SITE=datadoghq.com",
	)
}

func generateTestCommand(binary string, parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	// Check full path overrides first (most specific)
	fullPath := strings.Join(parts, " ")
	args := ""
	overridden := false
	if a, ok := fullPathOverrides[fullPath]; ok {
		args = a
		overridden = true
	}

	// Fall back to leaf-based lookup
	if !overridden {
		leaf := parts[len(parts)-1]
		if a, ok := testArgs[leaf]; ok {
			args = a
		}
	}

	// For domains that use integer IDs, replace test-id-123 with 12345
	needsIntID := false
	for _, p := range parts {
		if intIDParents[p] {
			needsIntID = true
			break
		}
	}
	if needsIntID {
		args = strings.ReplaceAll(args, "test-id-123", "12345")
	}

	cmd := binary + " " + strings.Join(parts, " ")
	if args != "" {
		cmd += " " + args
	}
	return cmd
}

func writeCommandFile(filename string, commands []string) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer f.Close()
	for _, cmd := range commands {
		fmt.Fprintln(f, cmd)
	}
}
