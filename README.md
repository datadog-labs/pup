# Pup - Datadog API CLI Wrapper

[![CI](https://github.com/DataDog/pup/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/DataDog/pup/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/DataDog/pup/main/.github/badges/coverage.json)](https://github.com/DataDog/pup/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

A Go-based command-line wrapper for easy interaction with Datadog APIs.

## Features

- **Native Go Implementation**: Fast, cross-platform binary
- **OAuth2 Authentication**: Secure browser-based login with PKCE protection
- **API Key Support**: Traditional API key authentication still available
- **Simple Commands**: Intuitive CLI for common Datadog operations
- **JSON Output**: Structured output for easy parsing and automation
- **Dynamic Client Registration**: Each installation gets unique OAuth credentials

## Installation

```bash
# Clone the repository
git clone https://github.com/DataDog/pup.git
cd pup

# Build
go build -o pup .

# Install (optional)
go install
```

## Authentication

Pup supports two authentication methods. **OAuth2 is preferred** and will be used automatically if you've logged in.

### OAuth2 Authentication (Preferred)

OAuth2 provides secure, browser-based authentication with automatic token refresh.

```bash
# Set your Datadog site (optional)
export DD_SITE="datadoghq.com"  # Defaults to datadoghq.com

# Login via browser
pup auth login

# Use any command - OAuth tokens are used automatically
pup monitors list

# Check status
pup auth status

# Logout
pup auth logout
```

**Token Storage**: Tokens are stored securely in your system's keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service). Set `DD_TOKEN_STORAGE=file` to use file-based storage instead.

**Note**: OAuth2 requires Dynamic Client Registration (DCR) to be enabled on your Datadog site. If DCR is not available yet, use API key authentication.

See [docs/OAUTH2.md](docs/OAUTH2.md) for detailed OAuth2 documentation.

### API Key Authentication (Fallback)

If OAuth2 tokens are not available, Pup automatically falls back to API key authentication.

```bash
export DD_API_KEY="your-datadog-api-key"
export DD_APP_KEY="your-datadog-application-key"
export DD_SITE="datadoghq.com"  # Optional, defaults to datadoghq.com

# Use any command - API keys are used automatically
pup monitors list
```

### Authentication Priority

Pup checks for authentication in this order:
1. **OAuth2 tokens** (from `pup auth login`) - Used if valid tokens exist
2. **API keys** (from `DD_API_KEY` and `DD_APP_KEY`) - Used if OAuth tokens not available

## Usage

### Authentication

```bash
# OAuth2 login (recommended)
pup auth login

# Check authentication status
pup auth status

# Refresh access token
pup auth refresh

# Logout
pup auth logout
```

### Test Connection

```bash
pup test
```

### Monitors

```bash
# List all monitors
pup monitors list

# Get specific monitor
pup monitors get 12345678

# Delete monitor
pup monitors delete 12345678 --yes
```

### Dashboards

```bash
# List all dashboards
pup dashboards list

# Get dashboard details
pup dashboards get abc-123-def

# Delete dashboard
pup dashboards delete abc-123-def --yes
```

### SLOs

```bash
# List all SLOs
pup slos list

# Get SLO details
pup slos get abc-123

# Delete SLO
pup slos delete abc-123 --yes
```

### Incidents

```bash
# List all incidents
pup incidents list

# Get incident details
pup incidents get abc-123-def
```

## Global Flags

- `-o, --output`: Output format (json, table, yaml) - default: json
- `-y, --yes`: Skip confirmation prompts for destructive operations

## Environment Variables

- `DD_API_KEY`: Datadog API key (optional if using OAuth2)
- `DD_APP_KEY`: Datadog Application key (optional if using OAuth2)
- `DD_SITE`: Datadog site (default: datadoghq.com)
- `DD_AUTO_APPROVE`: Auto-approve destructive operations (true/false)
- `DD_TOKEN_STORAGE`: Token storage backend (keychain or file, default: auto-detect)

## Development

```bash
# Run tests
go test ./...

# Build
go build -o pup .

# Run without building
go run main.go monitors list
```

## License

Apache License 2.0 - see LICENSE for details.

## Documentation

For detailed documentation, see [CLAUDE.md](CLAUDE.md).
