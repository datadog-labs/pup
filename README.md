# Pup - Datadog API CLI Wrapper

[![CI](https://github.com/DataDog/pup/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/DataDog/pup/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

A Go-based command-line wrapper for easy interaction with Datadog APIs.

## Features

- **Native Go Implementation**: Fast, cross-platform binary
- **OAuth2 Authentication**: Secure browser-based login with PKCE protection
- **API Key Support**: Traditional API key authentication still available
- **Simple Commands**: Intuitive CLI for common Datadog operations
- **JSON Output**: Structured output for easy parsing and automation
- **Dynamic Client Registration**: Each installation gets unique OAuth credentials

## API Coverage

<!-- Last updated: 2026-02-10 | API Client: v2.54.0 -->

Pup implements **37 of 85+ available Datadog APIs** (43.5% coverage).

**Summary:**
- ✅ **34 Working** - Fully implemented and functional
- ⏳ **3 Planned** - Skeleton implementation, endpoints pending
- ❌ **48+ Not Implemented** - Available in Datadog but not yet in pup

See [docs/COMMANDS.md](docs/COMMANDS.md) for detailed command reference.

💡 **Tip:** Use Ctrl/Cmd+F to search for specific APIs. [Request features via GitHub Issues](https://github.com/DataDog/pup/issues).

---

<details>
<summary><b>📊 Core Observability (6/9 implemented)</b></summary>

| API Domain | Status | Pup Commands | Notes |
|------------|--------|--------------|-------|
| Metrics | ✅ | `metrics query`, `metrics list`, `metrics get`, `metrics search` | Full query and metadata support |
| Logs | ✅ | `logs search`, `logs list`, `logs aggregate` | V1 and V2 APIs supported |
| Traces | ✅ | `traces search`, `traces list`, `traces aggregate` | APM traces support |
| Events | ✅ | `events list`, `events search`, `events get` | Infrastructure event management |
| RUM | ✅ | `rum apps`, `rum sessions`, `rum metrics list/get`, `rum retention-filters list/get` | Apps, sessions, metrics, retention filters (create/update pending) |
| APM Services | ❌ | - | Not yet implemented |
| Profiling | ❌ | - | Not yet implemented |
| Session Replay | ❌ | - | Not yet implemented |
| Spans Metrics | ❌ | - | Not yet implemented |

</details>

<details>
<summary><b>🔔 Monitoring & Alerting (6/9 implemented)</b></summary>

| API Domain | Status | Pup Commands | Notes |
|------------|--------|--------------|-------|
| Monitors | ✅ | `monitors list`, `monitors get`, `monitors delete`, `monitors search` | Full CRUD support with advanced search |
| Dashboards | ✅ | `dashboards list`, `dashboards get`, `dashboards delete`, `dashboards url` | Full management capabilities |
| SLOs | ✅ | `slos list`, `slos get`, `slos create`, `slos update`, `slos delete`, `slos corrections` | Full CRUD plus corrections |
| Synthetics | ✅ | `synthetics tests list`, `synthetics locations list` | Test management support |
| Downtimes | ✅ | `downtime list`, `downtime get`, `downtime cancel` | Full downtime management |
| Notebooks | ✅ | `notebooks list`, `notebooks get`, `notebooks delete` | Investigation notebooks supported |
| Dashboard Lists | ❌ | - | Not yet implemented |
| Powerpacks | ❌ | - | Not yet implemented |
| Workflow Automation | ❌ | - | Not yet implemented |

</details>

<details>
<summary><b>🔒 Security & Compliance (6/9 implemented)</b></summary>

| API Domain | Status | Pup Commands | Notes |
|------------|--------|--------------|-------|
| Security Monitoring | ✅ | `security rules list`, `security signals list`, `security findings list/get/search` | Rules, signals, findings with advanced search |
| Vulnerabilities | ✅ | `vulnerabilities search`, `vulnerabilities list` | Vulnerability tracking and management |
| Static Analysis | ✅ | `static-analysis ast`, `static-analysis custom-rulesets`, `static-analysis sca`, `static-analysis coverage` | Code security analysis |
| Audit Logs | ✅ | `audit-logs list`, `audit-logs search` | Full audit log search and listing |
| Data Governance | ✅ | `data-governance scanner-rules list` | Sensitive data scanner rules |
| Application Security | ❌ | - | Not yet implemented |
| CSM Threats | ❌ | - | Not yet implemented |
| Cloud Security (CSPM) | ❌ | - | Not yet implemented |
| Sensitive Data Scanner | ❌ | - | Not yet implemented |

</details>

<details>
<summary><b>☁️ Infrastructure & Cloud (6/8 implemented)</b></summary>

| API Domain | Status | Pup Commands | Notes |
|------------|--------|--------------|-------|
| Infrastructure | ✅ | `infrastructure hosts list`, `infrastructure hosts get` | Host inventory management |
| Tags | ✅ | `tags list`, `tags get`, `tags add`, `tags update`, `tags delete` | Host tag operations |
| Network | ⏳ | `network flows list`, `network devices list` | Placeholder - API endpoints pending |
| Cloud (AWS) | ✅ | `cloud aws list` | AWS integration management |
| Cloud (GCP) | ✅ | `cloud gcp list` | GCP integration management |
| Cloud (Azure) | ✅ | `cloud azure list` | Azure integration management |
| Containers | ❌ | - | Not yet implemented |
| Processes | ❌ | - | Not yet implemented |

</details>

<details>
<summary><b>🚨 Incident & Operations (6/7 implemented)</b></summary>

| API Domain | Status | Pup Commands | Notes |
|------------|--------|--------------|-------|
| Incidents | ✅ | `incidents list`, `incidents get`, `incidents attachments` | Incident management with attachment support |
| On-Call (Teams) | ✅ | `on-call teams` (CRUD, memberships with roles) | Full team management system with admin/member roles |
| Case Management | ✅ | `cases` (create, search, assign, archive, projects) | Complete case management with priorities P1-P5 |
| Error Tracking | ✅ | `error-tracking issues list`, `error-tracking issues get` | Error issue management |
| Service Catalog | ✅ | `service-catalog list`, `service-catalog get` | Service registry management |
| Scorecards | ✅ | `scorecards list`, `scorecards get` | Service quality scores |
| Incident Services/Teams | ❌ | - | Not yet implemented |

</details>

<details>
<summary><b>🔧 CI/CD & Development (1/3 implemented)</b></summary>

| API Domain | Status | Pup Commands | Notes |
|------------|--------|--------------|-------|
| CI Visibility | ✅ | `cicd pipelines list`, `cicd events list` | CI/CD pipeline visibility and events |
| Test Optimization | ❌ | - | Not yet implemented |
| DORA Metrics | ❌ | - | Not yet implemented |

</details>

<details>
<summary><b>👥 Organization & Access (5/6 implemented)</b></summary>

| API Domain | Status | Pup Commands | Notes |
|------------|--------|--------------|-------|
| Users | ✅ | `users list`, `users get`, `users roles` | User and role management |
| Organizations | ✅ | `organizations get`, `organizations list` | Organization settings management |
| API Keys | ✅ | `api-keys list`, `api-keys get`, `api-keys create`, `api-keys delete` | Full API key CRUD |
| App Keys | ✅ | `app-keys list`, `app-keys get`, `app-keys register`, `app-keys unregister` | App key registration for Action Connections |
| Service Accounts | ✅ | - | Managed via users commands |
| Roles | ❌ | - | Only list via users |

</details>

<details>
<summary><b>⚙️ Platform & Configuration (7/9 implemented)</b></summary>

| API Domain | Status | Pup Commands | Notes |
|------------|--------|--------------|-------|
| Usage Metering | ✅ | `usage summary`, `usage hourly` | Usage and billing metrics |
| Cost Management | ✅ | `cost projected`, `cost attribution`, `cost by-org` | Cost attribution by tags and organizations |
| Product Analytics | ✅ | `product-analytics events send` | Server-side product analytics events |
| Integrations | ✅ | `integrations slack`, `integrations pagerduty`, `integrations webhooks` | Third-party integrations support |
| Observability Pipelines | ⏳ | `obs-pipelines list`, `obs-pipelines get` | Placeholder - API endpoints pending |
| Miscellaneous | ✅ | `misc ip-ranges`, `misc status` | IP ranges and status |
| Key Management | ❌ | - | Not yet implemented |
| IP Allowlist | ❌ | - | Not yet implemented |

</details>

## Installation

```bash
go install github.com/DataDog/pup@latest
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
