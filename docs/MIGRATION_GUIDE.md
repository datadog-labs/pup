# Migration Guide: Go pup to Rust pup-rs

## Overview

The Rust version (`pup-rs`) is a drop-in replacement for the Go version (`pup`).
All commands, flags, and output formats are compatible. This guide covers the
differences you should be aware of when switching.

## What's the Same

- All 47 command groups with identical subcommands
- JSON, YAML, and table output formats (`--format json|yaml|table`)
- OAuth2 + API key authentication
- OS keychain token storage (same service/account names)
- Agent mode (`FORCE_AGENT_MODE=1`) with JSON schema
- Config file format (`~/.config/pup/config.yaml`)
- Environment variables (`DD_API_KEY`, `DD_APP_KEY`, `DD_SITE`, etc.)
- Exit codes (0 = success, non-zero = error)

## What's Different

### New Commands (Rust-only)

| Command | Description |
|---------|-------------|
| `dashboards create` | Create dashboards from JSON files |
| `dashboards update` | Update dashboards from JSON files |
| `monitors create` | Create monitors from JSON files |
| `monitors update` | Update monitors from JSON files |
| `slos create` | Create SLOs from JSON files |
| `slos update` | Update SLOs from JSON files |
| `downtime create` | Create downtimes from JSON files |
| `completions` | Generate shell completions (bash, zsh, fish, powershell) |

### Argument Style Changes

Some commands use positional arguments instead of named flags for primary identifiers:

| Command | Go | Rust |
|---------|-----|------|
| `rum apps get` | `--app-id ID` | `ID` (positional) |
| `rum metrics get` | `--metric-id ID` | `ID` (positional) |
| `cicd pipelines get` | `--pipeline-id ID` | `ID` (positional) |

### Help Text

- Human-mode `--help` uses clap formatting (slightly different layout from cobra)
- Agent-mode `--help` returns identical JSON schema structure (390/390 descriptions match)
- All commands include an EXAMPLES section in help output

### Error Messages

- Error messages use clap's format instead of cobra's
- Unknown flag errors have a different format but include the same information
- Exit codes are consistent (0 = success, non-zero = error)

### Performance

See [BENCHMARKS.md](BENCHMARKS.md) for detailed numbers. Summary:

| Metric | Go | Rust | Improvement |
|--------|-----|------|-------------|
| Binary size (stripped) | 37.3 MB | 25.7 MB | 31% smaller |
| Startup time | 9.3ms | 7.8ms | 16% faster |
| Peak memory | 19.3 MB | 14.4 MB | 25% less |

## Installation

### Build from source

```bash
cd pup-rs
cargo build --release
cp target/release/pup-rs /usr/local/bin/pup
```

### Via Homebrew (when available)

```bash
brew tap datadog-labs/pack
brew install pup
```

## Verifying the Migration

After installing the Rust version, verify everything works:

```bash
# Check version
pup version

# Verify auth (if using OAuth2)
pup auth status

# Test a read command
pup monitors list

# Test output formats
pup monitors list --format json
pup monitors list --format yaml
pup monitors list --format table

# Test agent mode (if applicable)
FORCE_AGENT_MODE=1 pup monitors --help
```

## Shell Completions

The Rust version includes built-in shell completion generation:

```bash
# Bash
pup completions bash > /usr/local/etc/bash_completion.d/pup

# Zsh
pup completions zsh > "${fpath[1]}/_pup"

# Fish
pup completions fish > ~/.config/fish/completions/pup.fish
```

## Configuration Compatibility

The Rust version reads the same config file (`~/.config/pup/config.yaml`) with the
same format. No changes are needed to your existing configuration.

Configuration precedence is identical:

1. Command-line flags (highest priority)
2. Environment variables
3. Config file (`~/.config/pup/config.yaml`)
4. Default values (lowest priority)

## Authentication Compatibility

### OAuth2

The Rust version uses the same OS keychain service and account names as the Go
version. Existing OAuth2 tokens stored by the Go version will be read by the Rust
version without re-authentication.

### API Keys

Environment variables work identically:

```bash
export DD_API_KEY="your-api-key"
export DD_APP_KEY="your-app-key"
export DD_SITE="datadoghq.com"
```

## Troubleshooting

### Command not found after installation

Ensure the binary is in your `PATH`:

```bash
which pup
# Should return /usr/local/bin/pup or similar
```

### OAuth2 token not found

If the Rust version cannot find your existing OAuth2 token, re-authenticate:

```bash
pup auth login
```

### Different help output format

The help text layout is different (clap vs cobra) but contains the same information.
This is expected and does not affect functionality.

### Reporting issues

If you encounter differences in behavior between the Go and Rust versions, please
file an issue at [GitHub Issues](https://github.com/datadog-labs/pup/issues) with:

- The command you ran
- Expected output (from Go version)
- Actual output (from Rust version)
- Your OS and architecture
