# Pup - Datadog API CLI

Rust-based CLI wrapper for Datadog APIs. Provides OAuth2 + API key authentication for 48 command groups with 283+ subcommands across 37 API domains.

## Documentation Index

- **[COMMANDS.md](docs/COMMANDS.md)** - Complete command reference with all 33 domains
- **[CONTRIBUTING.md](docs/CONTRIBUTING.md)** - Git workflow, PR process, commit format
- **[TESTING.md](docs/TESTING.md)** - Test strategy, coverage requirements, CI/CD
- **[OAUTH2.md](docs/OAUTH2.md)** - OAuth2 implementation details (DCR, PKCE, token storage)
- **[EXAMPLES.md](docs/EXAMPLES.md)** - Usage examples and common workflows
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Design decisions and technical details
- **[TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)** - Common issues and debugging

## Quick Start

```bash
# Install via Homebrew (recommended)
brew tap datadog-labs/pack
brew install pup

# Or clone and build from source
git clone https://github.com/datadog-labs/pup.git && cd pup
cargo build --release
cp target/release/pup /usr/local/bin/pup

# Authenticate (OAuth2 recommended)
pup auth login

# Or use API keys (legacy)
export DD_API_KEY="key" DD_APP_KEY="key" DD_SITE="datadoghq.com"

# Run a command
pup monitors list --tag="env:production"
pup metrics query --query="avg:system.cpu.user{*}" --from="1h"
```

## Project Structure

```
pup/
├── src/
│   ├── main.rs            # CLI entry point, clap enums, command routing
│   ├── commands/           # 47 command modules (monitors.rs, logs.rs, etc.)
│   ├── auth/              # OAuth2 + DCR + PKCE + token storage
│   ├── client.rs          # Datadog API client wrapper
│   ├── config.rs          # Configuration management
│   ├── formatter.rs       # Output formatting (JSON, YAML, table, agent envelope)
│   ├── useragent.rs       # AI agent detection (FORCE_AGENT_MODE, Claude Code, etc.)
│   ├── util.rs            # Time parsing, validation
│   └── version.rs         # Version and build info
├── tests/                 # E2E parity report
├── docs/                  # Extended documentation
├── Cargo.toml             # Rust dependencies
└── Cargo.lock
```

## Command Structure

All commands follow consistent patterns:

```bash
pup <domain> <action> [options]
pup <domain> <subgroup> <action> [options]

# Examples
pup monitors list --tag="env:prod"
pup logs search --query="status:error" --from="1h"
pup rum apps list
pup security rules list
```

See [COMMANDS.md](docs/COMMANDS.md) for complete reference.

## Development Workflow (For Agents)

### 1. Branch Creation

```bash
git checkout -b <type>/<description>

# Types: feat, fix, refactor, docs, test, chore, perf
# Example: feat/add-metrics-filtering
```

### 2. Code Changes

**Code Style:**
- Follow Rust conventions and idioms
- Use `cargo fmt` and `cargo clippy`
- Keep functions small and focused
- Write clear, self-documenting code

**Error Handling:**
- Use `anyhow` for application errors: `anyhow::bail!("context")`
- Wrap errors with context: `.map_err(|e| anyhow::anyhow!("context: {e:?}"))`
- Never expose API keys in errors

**Testing:**
- Write unit tests for public functions
- Use `#[test]` with `assert_eq!`
- Mock external dependencies
- Run with `cargo test`

### 3. Commit

Stage specific files and commit with conventional format:

```bash
git add src/commands/specific_file.rs
git commit -m "$(cat <<'EOF'
<type>(<scope>): <subject>

<body describing what and why>

- Key change 1
- Key change 2

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"
```

**Commit types:** feat, fix, docs, style, refactor, test, chore

### 4. Create PR

```bash
gh pr create \
  --title "<type>(<scope>): <clear title>" \
  --body "$(cat <<'EOF'
## Summary
1-2 sentences describing what and why.

## Changes
- Change 1 with file reference (src/commands/foo.rs:123)
- Change 2 with file reference

## Testing
- How changes were tested
- Test coverage details

## Related Issues
Closes #N

---
🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for detailed workflow and examples.

## Security Requirements

**Authentication:**
- OAuth2 with PKCE (recommended): `pup auth login`
- API keys (legacy): `DD_API_KEY` + `DD_APP_KEY` + `DD_SITE`
- Token storage: OS keychain (primary), encrypted file (fallback)

**Security Rules:**
- Never commit credentials (API keys, tokens, secrets)
- Never log or print access/refresh tokens
- Validate all user inputs to prevent injection
- Use PKCE S256 for OAuth2 flows
- Encrypt fallback token storage with AES-256-GCM

## Configuration Precedence

1. Command-line flags (highest priority)
2. Environment variables
3. Config file (`~/.config/pup/config.yaml`)
4. Default values (lowest priority)

## CI/CD Requirements

**All PRs must pass:**
- `cargo test` (60 unit tests)
- `cargo clippy -- -D warnings`
- `cargo fmt --check`
- Cross-compilation for 4 targets

See [TESTING.md](docs/TESTING.md) for details.

## Core Dependencies

- **Rust (stable)**
- **datadog-api-client-rust** v0.27.0 - Official API client
- **clap** v4 - CLI framework (derive macros)
- **tokio** - Async runtime
- **serde** + **serde_json** + **serde_yaml** - Serialization
- **keyring** v3 - OS keychain integration
- **reqwest** - HTTP client

## Implementation Status

- **47 command modules** implemented
- **283 leaf subcommands** across 48 command groups
- **60 unit tests** passing
- **155/155 API output** matches Go version exactly
- **390/390 command descriptions** match Go version
- **31% smaller binary**, 16% faster startup, 25% less memory vs Go

See [BENCHMARKS.md](docs/BENCHMARKS.md) for performance details.

## License

Apache 2.0 - Copyright 2024-present Datadog, Inc.

## Support

- Issues: [GitHub Issues](https://github.com/datadog-labs/pup/issues)
- Community: [Datadog Community](https://community.datadoghq.com/)
