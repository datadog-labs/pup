# Pup - Datadog API CLI

Go-based CLI wrapper for Datadog APIs. Provides OAuth2 + API key authentication for 28 command groups with 200+ subcommands across 33 API domains.

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
# Clone and build
git clone https://github.com/DataDog/pup.git && cd pup
go build -o pup .

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
├── cmd/                    # 28 command files (metrics, logs, monitors, etc.)
│   ├── root.go            # Root command + global flags
│   └── auth.go            # OAuth2 authentication
├── pkg/
│   ├── auth/              # OAuth2 + DCR + token storage
│   ├── client/            # Datadog API client wrapper
│   ├── config/            # Configuration management
│   ├── formatter/         # Output formatting (JSON, YAML, table)
│   └── util/              # Time parsing, validation
└── docs/                  # Extended documentation
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
- Follow Go conventions and idioms
- Use `gofmt` and `golangci-lint`
- Keep functions small and focused
- Write clear, self-documenting code

**Error Handling:**
- Use standard Go error patterns
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Never expose API keys in errors

**Testing:**
- Write unit tests for public functions
- Use table-driven tests
- Mock external dependencies
- Maintain >80% coverage (CI enforced)

### 3. Commit

Stage specific files and commit with conventional format:

```bash
git add pkg/specific/files.go
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
- Change 1 with file reference (file.go:123)
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
- Tests with race detection
- Code coverage ≥80% (enforced)
- `golangci-lint` checks
- Build verification

**Coverage badge auto-updates on main branch.**

See [TESTING.md](docs/TESTING.md) for details.

## Core Dependencies

- **Go 1.21+** with generics
- **datadog-api-client-go** - Official API client
- **cobra** - CLI framework
- **viper** - Configuration
- **keyring** - OS keychain integration

## Implementation Status

- **28 command files** implemented
- **200+ subcommands** across 33 domains
- **93.9% test coverage** in pkg/
- **23/33 commands** fully working
- **7/33 commands** blocked by API client library issues
- **3/33 commands** placeholder (API endpoints pending)

See [COMMANDS.md](docs/COMMANDS.md) for detailed status.

## License

Apache 2.0 - Copyright 2024-present Datadog, Inc.

## Support

- Issues: [GitHub Issues](https://github.com/DataDog/pup/issues)
- Community: [Datadog Community](https://community.datadoghq.com/)
