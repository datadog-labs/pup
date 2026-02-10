# Architecture & Design Decisions

Technical architecture and design rationale for Pup CLI.

## Technology Choices

### Why Go?

**Benefits:**
1. **Performance** - Compiled binary with fast startup (<100ms)
2. **Cross-platform** - Single binary for macOS, Linux, Windows
3. **Standard library** - Excellent networking, crypto, and HTTP support
4. **Concurrency** - Built-in goroutines for parallel operations
5. **Static typing** - Catch errors at compile time
6. **Small binaries** - ~15-20MB compiled size

**Tradeoffs:**
- More verbose than dynamic languages
- Requires compilation step (not interpreted)
- Less flexible than scripting languages

### Why Cobra?

**Benefits:**
1. **Industry standard** - Used by kubectl, hugo, gh, docker
2. **Rich features** - Subcommands, flags, aliases, help generation
3. **Shell completion** - Built-in bash/zsh/fish completion
4. **Community** - Large ecosystem and community support
5. **Documentation** - Excellent docs and examples

**Alternative considered:**
- `urfave/cli` - Simpler but less feature-rich

### Why Viper?

**Benefits:**
1. **Configuration management** - Handles config files, env vars, flags
2. **Precedence handling** - Clear priority: flags > env > config > defaults
3. **Multiple formats** - YAML, JSON, TOML support
4. **Watch functionality** - Can watch config file for changes

**Alternative considered:**
- Manual config parsing - Too much boilerplate

### Why OAuth2 over API Keys?

**OAuth2 advantages:**
1. **Security** - Short-lived tokens (1 hour) vs. long-lived keys
2. **Granular control** - Per-installation revocation
3. **Scope-based** - Request only necessary permissions
4. **User context** - Actions as authenticated user
5. **Auditable** - Better audit trail with user identity

**API Keys still supported:**
- Legacy compatibility
- Simpler for programmatic access
- No browser requirement

## Project Structure

```
pup/
├── cmd/                    # Command implementations
│   ├── root.go            # Root command + global flags
│   ├── auth.go            # OAuth2 authentication
│   ├── metrics.go         # Metrics domain
│   ├── monitors.go        # Monitors domain
│   └── ...                # 28 total command files
├── pkg/                   # Reusable packages
│   ├── auth/              # Authentication logic
│   │   ├── dcr/          # Dynamic Client Registration
│   │   ├── oauth/        # OAuth2 flow + PKCE
│   │   ├── storage/      # Token storage
│   │   └── callback/     # OAuth callback server
│   ├── client/           # Datadog API client wrapper
│   ├── config/           # Configuration management
│   ├── formatter/        # Output formatting
│   └── util/             # Utilities (time, validation)
└── docs/                 # Documentation
```

**Design principles:**
- `cmd/` contains Cobra commands (thin layer)
- `pkg/` contains business logic (testable)
- `internal/` for truly private code (currently just version)

## Authentication Architecture

### Token Storage Strategy

**Priority order:**
1. **OS Keychain** (primary) - Most secure
   - macOS: Keychain
   - Windows: Credential Manager
   - Linux: Secret Service / Keyring
2. **Encrypted file** (fallback) - When keychain unavailable
   - Location: `~/.config/pup/tokens.enc`
   - Encryption: AES-256-GCM
   - Key derivation: Machine-specific data

**Why not plaintext?**
- Security risk if file system compromised
- Better to require re-auth than expose tokens

### OAuth2 Flow

Based on [RFC 6749](https://tools.ietf.org/html/rfc6749) and [RFC 7636](https://tools.ietf.org/html/rfc7636) (PKCE):

```
1. DCR Registration → client_id, client_secret
2. PKCE Generation → verifier, challenge
3. Authorization URL → user approval
4. Callback Server → receive code
5. Token Exchange → access + refresh tokens
6. Secure Storage → keychain or encrypted file
7. Auto Refresh → before expiration
```

**Why PKCE?**
- Prevents authorization code interception
- Required for public clients (CLI)
- S256 method more secure than "plain"

### Token Refresh Strategy

**Automatic refresh triggers:**
- Token expires within 5 minutes
- API call with expired token
- Manual `pup auth refresh`

**Refresh flow:**
1. Check token expiration
2. Use refresh_token to get new access_token
3. Update stored tokens
4. Retry original API call

**Fallback:**
- If refresh fails, prompt user to re-authenticate

## User Agent

Custom user agent identifies pup CLI and detects AI coding assistants:

**Format:**
```
pup/v0.1.0 (go go1.25.0; os darwin; arch arm64)              # Without agent
pup/v0.1.0 (go go1.25.0; os darwin; arch arm64) claude-code  # With agent
```

**AI Agent Detection:**
- `CLAUDECODE=1` or `CLAUDE_CODE=1` → appends `claude-code`
- `CURSOR_AGENT=true` or `CURSOR_AGENT=1` → appends `cursor`
- Precedence: CLAUDECODE > CURSOR_AGENT

**Implementation:** `pkg/useragent` - Separate package for reusability and testing.

## API Client Wrapper

### Design Pattern

Thin wrapper around `datadog-api-client-go`:

```go
type Client struct {
    apiClient  *datadog.APIClient
    authConfig *config.AuthConfig
}

func (c *Client) QueryMetrics(ctx context.Context, query string) (*Response, error) {
    // Handle authentication
    ctx = c.withAuth(ctx)

    // Call Datadog API
    resp, _, err := c.apiClient.MetricsAPI.QueryMetrics(ctx).
        Query(query).
        Execute()

    // Handle errors
    return resp, handleAPIError(err)
}
```

**Benefits:**
- Centralized error handling
- Authentication injection
- Consistent retry logic
- Easy mocking for tests

### Error Handling

**Error wrapping pattern:**
```go
if err != nil {
    return fmt.Errorf("failed to query metrics: %w", err)
}
```

**Never expose:**
- API keys in error messages
- OAuth tokens in logs
- User credentials in output

## Command Structure

### Patterns

**Simple commands:**
```go
metricsCmd
├── queryCmd      # pup metrics query
├── listCmd       # pup metrics list
└── getCmd        # pup metrics get
```

**Nested commands:**
```go
rumCmd
├── appsCmd
│   ├── listCmd   # pup rum apps list
│   └── getCmd    # pup rum apps get
├── metricsCmd
│   └── getCmd    # pup rum metrics get
└── sessionsCmd
    └── searchCmd # pup rum sessions search
```

### Flag Consistency

Global flags available on all commands:
- `--config` - Config file path
- `--site` - Datadog site
- `--output` - Output format (json, yaml, table)
- `--verbose` - Enable debug logging
- `--yes` - Skip confirmations

Domain-specific flags:
- `--from`, `--to` - Time ranges (metrics, logs, traces)
- `--query` - Search query (logs, metrics, events)
- `--tag` - Tag filter (monitors, hosts)
- `--limit` - Result limit

## Output Formatting

### Three Formats

**JSON (default):**
- Machine-readable
- Preserves all fields
- No truncation

**YAML:**
- Human-readable
- Preserves structure
- Good for config files

**Table:**
- Compact display
- Truncates long values
- Good for terminals

### Formatter Design

```go
type Formatter interface {
    Format(data interface{}) ([]byte, error)
}

type JSONFormatter struct{}
type YAMLFormatter struct{}
type TableFormatter struct{}
```

## Configuration Management

### Precedence Order

1. **Command-line flags** (highest priority)
2. **Environment variables** (`DD_*`, `PUP_*`)
3. **Config file** (`~/.config/pup/config.yaml`)
4. **Default values** (lowest priority)

### Config File

Location: `~/.config/pup/config.yaml`

```yaml
# Authentication
site: datadoghq.com

# Output
output: json
verbose: false

# Defaults
default_from: 1h
default_to: now
```

## Performance Considerations

### Concurrency

**Parallel API requests:**
```go
var wg sync.WaitGroup
results := make(chan Result, len(ids))

for _, id := range ids {
    wg.Add(1)
    go func(id string) {
        defer wg.Done()
        result, err := client.Get(ctx, id)
        results <- Result{data: result, err: err}
    }(id)
}

wg.Wait()
close(results)
```

**Rate limiting:**
- Respect Datadog API limits (depends on plan)
- Implement exponential backoff for retries
- Use connection pooling

### Caching

**Considered but not implemented:**
- Metric metadata cache (reduce API calls)
- Monitor definition cache (improve list perf)
- TTL-based invalidation

**Reason:**
- Adds complexity
- Stale data risk
- CLI typically one-off operations

### Memory Management

**Streaming large results:**
```go
// Bad: load all into memory
results, _ := client.ListAll(ctx)
for _, r := range results {
    process(r)
}

// Good: stream with pagination
for page := 0; ; page++ {
    results, _ := client.List(ctx, page, 100)
    for _, r := range results {
        process(r)
    }
    if len(results) < 100 {
        break
    }
}
```

## Testing Strategy

### Package Tests (pkg/)

**High coverage (93.9% average):**
- Unit tests for all public functions
- Table-driven tests
- Mock external dependencies
- Test error paths

### Command Tests (cmd/)

**Structural tests:**
- Verify command hierarchy
- Check flags are registered
- Validate help text
- Test parent-child relationships

### Integration Tests (planned)

**Mock Datadog API:**
- `httptest.NewServer` for mock responses
- Test end-to-end command execution
- Validate request/response handling

## Security Architecture

### Threat Model

**Protected against:**
1. Token theft (encrypted storage)
2. Code interception (PKCE)
3. CSRF attacks (state parameter)
4. Man-in-the-middle (HTTPS only)
5. Credential exposure (never log tokens)

**Not protected against:**
- Compromised OS (keychain access)
- Malicious CLI binary (user must trust source)
- Datadog account compromise (outside scope)

### Security Best Practices

1. **Never commit secrets** - Gitignore tokens, keys
2. **Encrypt at rest** - Token storage encrypted
3. **Validate inputs** - Prevent injection attacks
4. **Use HTTPS** - All API calls over TLS
5. **Minimal scopes** - Request only needed permissions

## Future Enhancements

### Planned Features

1. **Shell completion** - Bash, zsh, fish
2. **Interactive mode** - TUI for complex operations
3. **Batch operations** - Process multiple resources
4. **Query caching** - Local cache for repeated queries
5. **Plugin system** - Extensible command system

### Considered but Deferred

1. **WebSocket support** - Real-time event streaming
2. **Local database** - SQLite for caching
3. **Desktop GUI** - Electron wrapper
4. **Web dashboard** - Browser-based UI

## Design Tradeoffs

### CLI vs. GUI

**Chose CLI because:**
- Easier to automate
- Scriptable workflows
- Lower resource usage
- Faster development

**GUI would provide:**
- Better discoverability
- Visual data representation
- Lower learning curve

### Monolith vs. Plugins

**Chose monolith because:**
- Simpler deployment (single binary)
- Faster startup (no plugin loading)
- Easier testing
- Consistent UX

**Plugins would provide:**
- Extensibility
- Community contributions
- Modular updates

### JSON vs. Protobuf

**Chose JSON because:**
- Human-readable
- Universal support
- Easier debugging
- Datadog API uses JSON

**Protobuf would provide:**
- Smaller payload size
- Faster parsing
- Schema validation

## References

- [Cobra Documentation](https://github.com/spf13/cobra)
- [Viper Documentation](https://github.com/spf13/viper)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Datadog API Client Go](https://github.com/DataDog/datadog-api-client-go)
