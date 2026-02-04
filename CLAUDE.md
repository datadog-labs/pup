# Pup - Datadog API CLI Wrapper

## Overview

Pup is a Go-based command-line wrapper that provides easy interaction with Datadog APIs. It builds upon the foundation of the [datadog-api-claude-plugin](https://github.com/DataDog/datadog-api-claude-plugin) to provide a native Go experience for developers who need to interact with Datadog's comprehensive monitoring and observability platform.

## Project Goals

1. **Go-Native Implementation**: Provide a performant, cross-platform CLI tool written in Go
2. **OAuth2 Authentication**: Support secure OAuth2 authentication with PKCE flow (in addition to traditional API keys)
3. **Simplified API Access**: Abstract complex Datadog API interactions into simple, intuitive CLI commands
4. **Developer Experience**: Focus on ergonomics and usability for daily operations
5. **Claude Code Integration**: Enable seamless integration with Claude Code for AI-assisted workflows

## Architecture

### Authentication

Pup supports two authentication methods:

#### 1. API Key Authentication (Traditional)
```bash
export DD_API_KEY="your-api-key"
export DD_APP_KEY="your-app-key"
export DD_SITE="datadoghq.com"
```

#### 2. OAuth2 Authentication (Recommended)
OAuth2 authentication with PKCE flow provides enhanced security features:

- **Dynamic Client Registration (DCR)**: Each CLI installation registers as a unique OAuth client
- **PKCE Protection**: Prevents authorization code interception attacks
- **Secure Token Storage**: Uses OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- **Automatic Token Refresh**: Background refresh before token expiration
- **Multi-site Support**: Works with all Datadog sites (US1, EU1, US3, US5, AP1, staging)
- **Granular Revocation**: Revoke access for one installation without affecting others

**OAuth2 Commands:**
```bash
# Login via browser-based OAuth flow
pup auth login

# Check authentication status
pup auth status

# Manually refresh access token
pup auth refresh

# Logout and clear stored tokens
pup auth logout
```

**OAuth2 Flow:**
1. User runs `pup auth login`
2. CLI performs Dynamic Client Registration with Datadog
3. CLI generates PKCE code verifier and challenge
4. Browser opens to Datadog authorization page
5. User approves requested OAuth scopes
6. Datadog redirects to local callback server with authorization code
7. CLI exchanges code for access/refresh tokens using PKCE verifier
8. Tokens stored securely in OS keychain
9. CLI automatically refreshes tokens before expiration

**OAuth2 Scopes Requested:**
- **Dashboards**: `dashboards_read`, `dashboards_write`
- **Monitors**: `monitors_read`, `monitors_write`, `monitors_downtime`
- **APM/Traces**: `apm_read`
- **SLOs**: `slos_read`, `slos_write`, `slos_corrections`
- **Incidents**: `incident_read`, `incident_write`
- **Synthetics**: `synthetics_read`, `synthetics_write`, `synthetics_global_variable_*`, `synthetics_private_location_*`
- **Security**: `security_monitoring_signals_read`, `security_monitoring_rules_read`, `security_monitoring_findings_read`, `security_monitoring_suppressions_read`, `security_monitoring_filters_read`
- **RUM**: `rum_apps_read`, `rum_apps_write`, `rum_retention_filters_read`, `rum_retention_filters_write`
- **Infrastructure**: `hosts_read`
- **Users**: `user_access_read`, `user_self_profile_read`
- **Cases**: `cases_read`, `cases_write`
- **Events**: `events_read`
- **Logs**: `logs_read_data`, `logs_read_index_data`
- **Metrics**: `metrics_read`, `timeseries_query`
- **Usage**: `usage_read`

### Command Structure

```bash
pup <domain> <action> [options]
```

Example:
```bash
pup metrics query --query="avg:system.cpu.user{*}" --from="1h" --to="now"
pup monitors list --tag="env:production"
pup logs search --query="status:error" --from="1h"
```

### Core Domains

Based on the datadog-api-claude-plugin architecture, Pup organizes commands into functional domains:

#### Data & Observability
Query and analyze telemetry data:
- **logs**: Search and analyze log data
- **traces**: Query APM traces and spans
- **metrics**: Query time-series metrics
- **rum**: Real User Monitoring data
- **events**: Infrastructure events
- **security**: Security signals and findings
- **audience-management**: RUM user/account segmentation

#### Monitoring & Alerting
Set up monitoring and visualization:
- **monitors**: Monitors, templates, notifications, downtimes
- **dashboards**: Visualization dashboards
- **slos**: Service Level Objectives
- **synthetics**: Synthetic monitoring tests
- **notebooks**: Investigation notebooks
- **powerpacks**: Reusable dashboard templates

#### Configuration & Data Management
Configure data collection and processing:
- **observability-pipelines**: Infrastructure-level data collection and routing
- **log-configuration**: Log archives, pipelines, indexes, destinations
- **apm-configuration**: APM retention and span-based metrics
- **rum-metrics-retention**: RUM retention and metrics

#### Infrastructure & Performance
Monitor infrastructure and performance:
- **infrastructure**: Host inventory and monitoring
- **container-monitoring**: Kubernetes and container metrics
- **database-monitoring**: Database performance
- **network-performance**: Network flow analysis
- **fleet-automation**: Agent deployment at scale

#### Security & Compliance
Security operations and posture management:
- **security-posture-management**: CSPM, vulnerabilities, SBOM
- **application-security**: ASM runtime threat detection
- **cloud-workload-security**: CWS runtime security
- **agentless-scanning**: Cloud resource scanning
- **static-analysis**: Code security scanning

#### Cloud & Integrations
Cloud provider and third-party integrations:
- **aws-integration**: AWS monitoring and security
- **gcp-integration**: GCP monitoring and security
- **azure-integration**: Azure monitoring and security
- **third-party-integrations**: PagerDuty, Slack, OpsGenie, etc.

#### Development & Quality
CI/CD and code quality:
- **cicd**: Pipeline visibility and testing
- **error-tracking**: Application error management
- **scorecards**: Service quality tracking
- **service-catalog**: Service registry

#### Operations & Automation
Incident response and automation:
- **incident-response**: On-call, incidents, case management
- **workflows**: Automated workflows
- **app-builder**: Custom internal applications

#### Organization & Access
User management and governance:
- **user-access-management**: Users, teams, SCIM, auth mappings
- **saml-configuration**: SAML SSO setup
- **organization-management**: Multi-org settings
- **data-governance**: Access controls, sensitive data
- **audit-logs**: Audit trail
- **api-management**: API and application keys

#### Cost & Usage
Cost monitoring and optimization:
- **cloud-cost**: Cloud cost tracking
- **usage-metering**: Datadog usage attribution
- **data-deletion**: Data retention policies

## Technical Stack

### Core Dependencies

- **Go 1.21+**: Modern Go with generics support
- **datadog-api-client-go**: Official Datadog API client ([github.com/DataDog/datadog-api-client-go](https://github.com/DataDog/datadog-api-client-go))
- **cobra**: CLI framework for command structure
- **viper**: Configuration management
- **keyring**: OS keychain integration for secure token storage (based on patterns from the TypeScript implementation)

### Token Storage Strategy

Following the pattern from the TypeScript plugin PR #84, implement secure token storage:

1. **Primary Storage**: OS Keychain
   - macOS: Keychain
   - Windows: Credential Manager
   - Linux: Secret Service / Keyring

2. **Fallback Storage**: Encrypted file
   - Location: `~/.config/pup/tokens.enc`
   - Encryption: AES-256-GCM with machine-specific key derivation

3. **Token Migration**: Support migrating from API key to OAuth2 authentication

### OAuth2 Implementation

Based on PR #84 from the TypeScript plugin, implement:

1. **DCR Client** (`pkg/auth/dcr/`)
   - Dynamic Client Registration with Datadog OAuth API
   - Client credential storage and retrieval
   - Per-site client management

2. **OAuth Client** (`pkg/auth/oauth/`)
   - PKCE code challenge generation (S256 method)
   - Authorization URL generation
   - Token exchange
   - CSRF state parameter validation

3. **Token Storage** (`pkg/auth/storage/`)
   - Keychain integration (primary)
   - Encrypted file storage (fallback)
   - Token migration utilities

4. **Token Refresher** (`pkg/auth/refresh/`)
   - Background token refresh scheduler
   - Automatic refresh before expiration
   - Refresh token rotation handling

5. **Callback Server** (`pkg/auth/callback/`)
   - Local HTTP server for OAuth callback
   - PKCS authorization code reception
   - State parameter validation
   - Error handling and user feedback

## Project Structure

```
fetch/
├── cmd/
│   ├── root.go                 # Root command and global flags
│   ├── auth.go                 # OAuth2 authentication commands
│   ├── metrics.go              # Metrics domain commands
│   ├── monitors.go             # Monitors domain commands
│   ├── dashboards.go           # Dashboards domain commands
│   ├── logs.go                 # Logs domain commands
│   ├── traces.go               # Traces domain commands
│   └── ...                     # Other domain commands
├── pkg/
│   ├── auth/
│   │   ├── dcr/                # Dynamic Client Registration
│   │   │   ├── client.go       # DCR API client
│   │   │   ├── types.go        # DCR type definitions
│   │   │   └── storage.go      # Client credentials storage
│   │   ├── oauth/              # OAuth2 client
│   │   │   ├── client.go       # OAuth2 flow implementation
│   │   │   ├── pkce.go         # PKCE utilities
│   │   │   └── scopes.go       # OAuth scope definitions
│   │   ├── storage/            # Token storage
│   │   │   ├── keychain.go     # OS keychain integration
│   │   │   ├── encrypted.go    # Encrypted file storage
│   │   │   ├── factory.go      # Storage selection logic
│   │   │   └── migration.go    # Token migration utilities
│   │   ├── refresh/            # Token refresh
│   │   │   └── refresher.go    # Automatic token refresh
│   │   └── callback/           # OAuth callback server
│   │       └── server.go       # Local HTTP server
│   ├── client/
│   │   ├── client.go           # Datadog API client wrapper
│   │   └── config.go           # Client configuration
│   ├── config/
│   │   └── config.go           # Application configuration
│   ├── formatter/
│   │   ├── json.go             # JSON output formatting
│   │   ├── table.go            # Table output formatting
│   │   └── yaml.go             # YAML output formatting
│   └── util/
│       ├── time.go             # Time parsing utilities
│       └── validation.go       # Input validation
├── internal/
│   └── version/
│       └── version.go          # Version information
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
├── README.md
├── CLAUDE.md                   # This file
└── main.go                     # Application entry point
```

## Development Guidelines

### Workflow for Claude Code

**When completing any task that involves code changes:**
1. Create a feature branch with appropriate prefix (feat/, fix/, etc.)
2. Make changes following the code style guidelines below
3. Stage specific files (avoid `git add .`)
4. Commit with conventional commit format
5. Create PR using `gh pr create` with detailed description

See the [Automated Development Workflow](#automated-development-workflow-for-claude-code) section for complete details and examples.

### Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` and `golangci-lint` for code formatting and linting
- Write idiomatic Go with clear, self-documenting code
- Use meaningful variable and function names
- Keep functions small and focused on a single responsibility

### Error Handling

- Use Go's standard error handling patterns
- Wrap errors with context using `fmt.Errorf` with `%w`
- Provide clear, actionable error messages
- Never expose API keys or sensitive data in error messages

### Testing

- Write unit tests for all public functions
- Use table-driven tests for multiple test cases
- Mock external dependencies (Datadog API calls)
- Aim for >80% code coverage
- Include integration tests for critical paths

### Configuration Precedence

Configuration values are resolved in the following order (highest to lowest priority):

1. Command-line flags
2. Environment variables
3. Configuration file (`~/.config/pup/config.yaml`)
4. Default values

### Security Considerations

- **Never commit credentials**: API keys, OAuth tokens, or client secrets
- **Use OS keychain**: Primary storage for sensitive credentials
- **Encrypt fallback storage**: When keychain is unavailable
- **Validate all inputs**: Prevent injection attacks
- **Secure callback server**: Use random port, validate state parameter
- **Token rotation**: Support refresh token rotation
- **Audit logging**: Log authentication events for security monitoring

### OAuth2 Security Best Practices

Based on the TypeScript plugin implementation:

1. **Dynamic Client Registration**: Each installation gets unique credentials
2. **PKCE S256**: Use SHA256 code challenge method
3. **State Parameter**: CSRF protection with random state value
4. **Secure Redirect URI**: `http://127.0.0.1:<random-port>/callback` only
5. **Token Storage**: Never log or print access/refresh tokens
6. **Scope Minimization**: Request only necessary OAuth scopes
7. **Token Expiration**: Refresh tokens before expiration
8. **Graceful Degradation**: Fall back to encrypted file if keychain unavailable

## Contributing

### Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/DataDog/pup.git
   cd pup
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the project:
   ```bash
   go build -o pup .
   ```

4. Run tests:
   ```bash
   go test ./...
   ```

5. Run with local changes:
   ```bash
   go run main.go <command>
   ```

### Automated Development Workflow (For Claude Code)

When working on tasks, follow this automated workflow:

#### 1. Create Feature Branch

Create a descriptive branch name based on the work being done:

```bash
# Branch naming convention: <type>/<short-description>
# Examples:
git checkout -b feat/oauth2-token-refresh
git checkout -b fix/metrics-query-timeout
git checkout -b refactor/simplify-auth-client
git checkout -b docs/update-readme-oauth
```

**Branch type prefixes:**
- `feat/` - New features
- `fix/` - Bug fixes
- `refactor/` - Code refactoring
- `docs/` - Documentation updates
- `test/` - Test additions/updates
- `chore/` - Maintenance tasks
- `perf/` - Performance improvements

#### 2. Make Changes and Commit

After completing the work:

1. **Stage relevant files** (prefer specific files over `git add .`):
   ```bash
   git add pkg/auth/oauth/client.go pkg/auth/oauth/client_test.go
   ```

2. **Commit with conventional commit message**:
   ```bash
   git commit -m "$(cat <<'EOF'
   <type>(<scope>): <subject>

   <body describing what changed and why>

   - Key change 1
   - Key change 2
   - Key change 3

   Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
   EOF
   )"
   ```

**Important**: Always include the Co-Authored-By line to credit Claude's contribution.

#### 3. Create Pull Request with gh CLI

Use `gh` CLI to push and create PR in one step:

```bash
gh pr create \
  --title "<type>(<scope>): <clear, concise title>" \
  --body "$(cat <<'EOF'
## Summary
Brief overview of what this PR does (1-2 sentences).

## Changes
- Specific change 1 with file reference
- Specific change 2 with file reference
- Specific change 3 with file reference

## Testing
- Test scenarios covered
- How to verify the changes

## Related Issues
Closes #<issue-number> (if applicable)
Fixes #<issue-number> (if applicable)

---
🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)" \
  --label "<appropriate-labels>" \
  --draft  # Optional: use --draft for work-in-progress

# Example:
gh pr create \
  --title "feat(auth): implement OAuth2 token refresh with PKCE" \
  --body "$(cat <<'EOF'
## Summary
Implements automatic OAuth2 token refresh using PKCE flow to maintain authentication without user intervention.

## Changes
- Added token refresher in pkg/auth/refresh/refresher.go:45
- Implemented background refresh scheduler
- Added unit tests for refresh logic in pkg/auth/refresh/refresher_test.go
- Updated OAuth client to use refresh tokens

## Testing
- Unit tests verify refresh token exchange
- Integration tests validate automatic refresh before expiration
- Manual test: wait 50 minutes and verify token auto-refreshes

## Related Issues
Closes #42

---
🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)" \
  --label "enhancement,auth"
```

#### 4. PR Description Best Practices

**Good PR descriptions include:**
- **Summary**: What and why in 1-2 sentences
- **Changes**: Bulleted list of specific changes with file references
- **Testing**: How the changes were tested
- **Related Issues**: Link to GitHub issues using `Closes #N` or `Fixes #N`
- **Screenshots/Examples**: For UI changes or CLI output changes
- **Breaking Changes**: Clearly marked if any
- **Migration Guide**: If breaking changes require user action

**Example of excellent PR body:**
```markdown
## Summary
Adds OAuth2 authentication with PKCE to replace API key authentication, providing better security and per-installation access control.

## Changes
- Implemented DCR client in pkg/auth/dcr/client.go
- Added PKCE challenge generation in pkg/auth/oauth/pkce.go:23
- Integrated OS keychain storage in pkg/auth/storage/keychain.go
- Added `pup auth login` command in cmd/auth.go:156
- Updated CLAUDE.md with OAuth2 documentation

## Testing
- Unit tests cover all OAuth2 flow steps
- Integration tests validate end-to-end authentication
- Manual testing on macOS, Linux, and Windows
- Verified keychain storage and fallback to encrypted file

## Breaking Changes
None. OAuth2 is opt-in; API key authentication still works.

## Related Issues
Closes #42
Implements RFC: #38

---
🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

#### 5. Complete Automated Workflow Example

Here's the complete workflow in one script:

```bash
# 1. Create feature branch
git checkout -b feat/add-oauth2-auth

# 2. [Make code changes...]

# 3. Stage specific files
git add pkg/auth/oauth/ cmd/auth.go

# 4. Commit with proper message
git commit -m "$(cat <<'EOF'
feat(auth): add OAuth2 authentication with PKCE

Implement OAuth2 authentication flow including:
- Dynamic Client Registration (DCR)
- PKCE code challenge generation
- Secure token storage via OS keychain
- Automatic token refresh

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"

# 5. Create PR with gh CLI
gh pr create \
  --title "feat(auth): add OAuth2 authentication with PKCE" \
  --body "$(cat <<'EOF'
## Summary
Adds OAuth2 authentication with PKCE flow to provide secure, per-installation authentication as an alternative to API keys.

## Changes
- Implemented DCR client in pkg/auth/dcr/client.go
- Added PKCE utilities in pkg/auth/oauth/pkce.go
- Created token storage with keychain integration
- Added `pup auth login/logout/status` commands
- Updated documentation in CLAUDE.md

## Testing
- Unit tests for OAuth flow components
- Integration tests for end-to-end flow
- Manual testing on macOS/Linux/Windows
- Verified token refresh automation

## Related Issues
Closes #42

---
🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)" \
  --label "enhancement,auth"
```

### Pull Request Process (For Human Contributors)

1. **Create a feature branch**: `git checkout -b feature/your-feature-name`
2. **Make your changes**: Write clear, well-documented code
3. **Write tests**: Ensure new code is covered by tests
4. **Run tests**: `go test ./...`
5. **Run linters**: `golangci-lint run`
6. **Commit changes**: Use clear, descriptive commit messages
7. **Push branch**: `git push origin feature/your-feature-name`
8. **Create PR**: Open a pull request with detailed description (use `gh pr create` or web UI)
9. **Address feedback**: Respond to review comments promptly
10. **Merge**: Once approved, squash and merge

### Commit Message Format

Follow conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test additions or changes
- `chore`: Build process or tooling changes

**Example:**
```
feat(auth): add OAuth2 authentication with PKCE

Implement OAuth2 authentication flow with PKCE protection including:
- Dynamic Client Registration (DCR)
- Secure token storage via OS keychain
- Automatic token refresh
- Local callback server for authorization code

Closes #42
```

## Related Resources

### Datadog API Documentation
- [Datadog API Reference](https://docs.datadoghq.com/api/latest/)
- [Datadog API Client Go](https://github.com/DataDog/datadog-api-client-go)
- [Datadog OpenAPI Spec](https://github.com/DataDog/datadog-api-spec) (Private)

### Related Projects
- [datadog-api-claude-plugin](https://github.com/DataDog/datadog-api-claude-plugin) - TypeScript-based plugin (reference implementation)
- [PR #84: OAuth2 Authentication](https://github.com/DataDog/datadog-api-claude-plugin/pull/84) - OAuth2 implementation reference

### OAuth2 & Security
- [RFC 7636: PKCE](https://tools.ietf.org/html/rfc7636) - Proof Key for Code Exchange
- [RFC 7591: OAuth 2.0 Dynamic Client Registration](https://tools.ietf.org/html/rfc7591)
- [OAuth 2.0 Best Practices](https://tools.ietf.org/html/draft-ietf-oauth-security-topics)

### Go Libraries
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [keyring](https://github.com/99designs/keyring) - OS keychain integration

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

Unless explicitly stated otherwise all files in this repository are licensed under the Apache License Version 2.0.

This product includes software developed at Datadog (https://www.datadoghq.com/).
Copyright 2024-present Datadog, Inc.

## Support

- **Issues**: [GitHub Issues](https://github.com/DataDog/pup/issues)
- **Documentation**: This file and inline code documentation
- **Community**: [Datadog Community](https://community.datadoghq.com/)

## Roadmap

### Phase 1: Foundation (MVP)
- [x] Project structure setup
- [ ] Basic CLI framework with Cobra
- [ ] Configuration management with Viper
- [ ] API key authentication
- [ ] Core Datadog client wrapper
- [ ] Basic metrics commands
- [ ] Basic monitors commands
- [ ] Error handling and formatting

### Phase 2: OAuth2 Authentication
- [ ] Dynamic Client Registration (DCR)
- [ ] OAuth2 PKCE flow implementation
- [ ] Local callback server
- [ ] OS keychain integration
- [ ] Encrypted file fallback storage
- [ ] Automatic token refresh
- [ ] Token migration utilities
- [ ] `pup auth` command suite

### Phase 3: Core Domains
- [ ] Logs domain commands
- [ ] Traces domain commands
- [ ] Dashboards domain commands
- [ ] SLOs domain commands
- [ ] Incidents domain commands
- [ ] Synthetics domain commands

### Phase 4: Advanced Features
- [ ] Enhanced output formatting (tables, JSON, YAML)
- [ ] Configuration file support
- [ ] Shell completion (bash, zsh, fish)
- [ ] Interactive mode for complex operations
- [ ] Batch operations support
- [ ] Query result caching

### Phase 5: Extended Domains
- [ ] Security domain commands
- [ ] Infrastructure domain commands
- [ ] RUM domain commands
- [ ] Events domain commands
- [ ] All remaining domains from plugin

### Phase 6: Polish & Distribution
- [ ] Comprehensive documentation
- [ ] Integration tests
- [ ] Performance optimization
- [ ] Release automation
- [ ] Binary distribution (homebrew, apt, etc.)
- [ ] Docker image

## Design Decisions

### Why Go?

1. **Performance**: Compiled binary with fast startup time
2. **Cross-platform**: Single binary for macOS, Linux, Windows
3. **Standard library**: Excellent networking and crypto support
4. **Concurrency**: Built-in support for concurrent operations
5. **Static typing**: Catch errors at compile time
6. **Native OAuth2**: `golang.org/x/oauth2` standard library

### Why Cobra?

1. **Industry standard**: Used by kubectl, hugo, GitHub CLI
2. **Rich features**: Subcommands, flags, aliases, help generation
3. **Completion**: Built-in shell completion support
4. **Community**: Large ecosystem and community support

### Why OAuth2 over API Keys?

1. **Security**: No long-lived credentials in files
2. **Granular control**: Per-installation revocation
3. **Scope-based**: Request only necessary permissions
4. **User context**: Actions performed as authenticated user
5. **Auditable**: Better audit trail with OAuth tokens

### Token Storage Strategy

1. **OS Keychain First**: Most secure, native integration
2. **Encrypted File Fallback**: For systems without keychain
3. **No plaintext**: Never store tokens in plaintext
4. **Machine-specific**: Encryption keys tied to machine

## Usage Examples

### Authentication

```bash
# OAuth2 login (recommended)
pup auth login

# Check authentication status
pup auth status

# Logout
pup auth logout

# API key authentication (legacy)
export DD_API_KEY="your-api-key"
export DD_APP_KEY="your-app-key"
```

### Metrics

```bash
# List all metrics
pup metrics list

# Filter metrics by pattern
pup metrics list --filter="system.*"

# Query metric data
pup metrics query --query="avg:system.cpu.user{*}" --from="1h" --to="now"

# Query with custom aggregation
pup metrics query --query="sum:app.requests{env:prod} by {service}" --from="4h"
```

### Monitors

```bash
# List all monitors
pup monitors list

# Filter monitors by tag
pup monitors list --tag="env:production"

# Get specific monitor
pup monitors get 12345678

# Search monitors by name
pup monitors search "CPU"

# Delete monitor (requires confirmation)
pup monitors delete 12345678
```

### Logs

```bash
# Search for errors
pup logs search --query="status:error" --from="1h" --to="now"

# Search by service
pup logs search --query="service:web-app status:warn"

# Complex query with attributes
pup logs search --query="@user.id:12345 status:error" --limit=100
```

### Dashboards

```bash
# List all dashboards
pup dashboards list

# Get dashboard details
pup dashboards get "abc-123-def"

# Get dashboard public URL
pup dashboards url "abc-123-def"
```

### Output Formatting

```bash
# JSON output (default)
pup monitors list --output=json

# Table output
pup monitors list --output=table

# YAML output
pup monitors list --output=yaml

# Custom fields
pup monitors list --fields="id,name,type,status"
```

### Advanced Usage

```bash
# Use custom config file
pup --config=/path/to/config.yaml monitors list

# Specify Datadog site
pup --site=datadoghq.eu monitors list

# Verbose output for debugging
pup --verbose monitors list

# Silent mode (no prompts)
pup --yes monitors delete 12345678
```

## Testing Strategy

### Unit Tests

Test individual functions and methods in isolation:

```go
func TestParseTimeParam(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        want     time.Time
        wantErr  bool
    }{
        {"relative hour", "1h", time.Now().Add(-1 * time.Hour), false},
        {"relative minutes", "30m", time.Now().Add(-30 * time.Minute), false},
        {"relative days", "7d", time.Now().Add(-7 * 24 * time.Hour), false},
        {"now", "now", time.Now(), false},
        {"invalid", "invalid", time.Time{}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseTimeParam(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("parseTimeParam() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            // Assert got matches want (with tolerance for time drift)
        })
    }
}
```

### Integration Tests

Test end-to-end workflows with mocked Datadog API:

```go
func TestMetricsQuery(t *testing.T) {
    // Setup mock Datadog API server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validate request
        assert.Equal(t, "GET", r.Method)
        assert.Contains(t, r.URL.Path, "/api/v2/query/timeseries")

        // Return mock response
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(mockMetricsResponse)
    }))
    defer server.Close()

    // Test metrics query command
    cmd := exec.Command("pup", "metrics", "query",
        "--query=avg:system.cpu.user{*}",
        "--from=1h",
        "--to=now",
        "--api-url="+server.URL)

    output, err := cmd.CombinedOutput()
    assert.NoError(t, err)
    assert.Contains(t, string(output), "system.cpu.user")
}
```

### OAuth2 Testing

Test OAuth2 flow with mock servers:

```go
func TestOAuthLogin(t *testing.T) {
    // Setup mock OAuth server
    authServer := httptest.NewServer(/* OAuth handlers */)
    defer authServer.Close()

    // Test OAuth login flow
    // 1. DCR registration
    // 2. Authorization URL generation
    // 3. Token exchange
    // 4. Token storage
    // 5. Token refresh
}
```

## Performance Considerations

### Concurrency

- Use goroutines for parallel API requests when fetching multiple resources
- Implement rate limiting to respect Datadog API limits
- Use connection pooling for HTTP client
- Consider worker pools for bulk operations

### Caching

- Cache metric metadata to reduce API calls
- Cache monitor definitions for improved list performance
- Implement TTL-based cache invalidation
- Use local cache for frequently accessed data

### Memory Management

- Stream large result sets instead of loading into memory
- Use pagination for large lists
- Implement result limiting with sensible defaults
- Clean up resources properly (defer close)

## Troubleshooting

### Common Issues

**OAuth2 Login Fails:**
- Check network connectivity to Datadog
- Verify site parameter is correct
- Ensure port is available for callback server
- Check firewall rules allow localhost connections

**Token Refresh Fails:**
- Verify refresh token hasn't expired
- Check network connectivity
- Try `pup auth logout` and `pup auth login` to re-authenticate

**Keychain Access Denied:**
- Grant keychain access permissions
- Falls back to encrypted file storage automatically
- Check `~/.config/pup/tokens.enc` exists

**API Rate Limiting:**
- Implement exponential backoff
- Reduce concurrent requests
- Use pagination with smaller page sizes

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
# Enable debug logging
pup --verbose monitors list

# Or set environment variable
export PUP_LOG_LEVEL=debug
pup monitors list
```

### Getting Help

1. Check documentation in this file
2. Search existing GitHub issues
3. Review Datadog API documentation
4. Open new GitHub issue with:
   - Pup version (`pup version`)
   - Command that failed
   - Error message and stack trace (if available)
   - Steps to reproduce
