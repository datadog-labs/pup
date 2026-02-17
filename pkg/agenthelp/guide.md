# Pup Agent Guide

Pup is a CLI for the Datadog API. This guide helps AI coding agents use pup effectively.

## Quick Start

```bash
# Authenticate (one-time setup)
pup auth login

# Or use API keys
export DD_API_KEY="your-key" DD_APP_KEY="your-key" DD_SITE="datadoghq.com"

# Get the full command schema (recommended first step for agents)
pup --help

# Get schema for a specific domain
pup logs --help
```

## Authentication

- **OAuth2 (recommended):** `pup auth login` — opens browser for secure login
- **API keys:** Set `DD_API_KEY`, `DD_APP_KEY`, and `DD_SITE` environment variables
- OAuth2 tokens are stored in the OS keychain and refresh automatically
- Some endpoints (logs search) require API keys even with OAuth2

## Logs

### Query Syntax
```
status:error                    # Filter by status
service:web-app                 # Filter by service
@user.id:12345                  # Filter by custom attribute
host:i-*                        # Wildcard matching
"exact error message"           # Exact phrase matching
status:error AND service:web    # Boolean AND
status:error OR status:warn     # Boolean OR
NOT status:info                 # Negation
-status:info                    # Shorthand negation
```

### Commands
```bash
# Search logs (v1 API, follows pagination)
pup logs search --query="status:error" --from=1h --limit=100

# Query logs (v2 API)
pup logs query --query="service:api AND status:error" --from=4h

# Aggregate logs (counts, distributions)
pup logs aggregate --query="*" --from=1h --compute="count" --group-by="service"
pup logs aggregate --query="service:api" --from=1h --compute="avg(@duration)" --group-by="@http.status_code"

# List logs (v2 API, simple)
pup logs list --from=1h --limit=20

# Storage tiers: indexes (default), online-archives, flex
pup logs search --query="*" --from=30d --storage="flex"
```

### Tips
- Always specify `--from` for a time range
- Use `aggregate` for counting/statistics, not `search` + local processing
- Default limit is 50; increase with `--limit` up to 1000
- Supported compute functions: count, avg, sum, min, max, cardinality, percentile

## Metrics

### Query Syntax
```
<aggregation>:<metric_name>{<filter>} by {<group>}

# Examples:
avg:system.cpu.user{*}                          # All hosts, average CPU
avg:system.cpu.user{env:prod} by {host}         # By host, production only
sum:trace.servlet.request.hits{service:web}     # Request count
max:system.mem.used{*} by {host}                # Max memory by host
```

### Commands
```bash
# Query metrics (timeseries)
pup metrics query --query="avg:system.cpu.user{env:prod} by {host}" --from=1h

# List metric names
pup metrics list --query="system.cpu"

# Get metric metadata
pup metrics metadata get "system.cpu.user"

# Submit metrics
pup metrics submit --metric="custom.metric" --value=42 --tags="env:test"
```

### Tips
- Aggregations: avg, sum, min, max, count
- Always include `{...}` filter even if empty: `{*}` means all
- Time range defaults to 1h if not specified

## Monitors

### Commands
```bash
# List monitors with filtering
pup monitors list --tags="env:production" --limit=500
pup monitors list --name="CPU" --tags="team:backend"

# Search monitors (full-text)
pup monitors search --query="database"

# Get monitor details
pup monitors get 12345678

# Delete monitor (prompts for confirmation; use --yes to skip)
pup monitors delete 12345678 --yes
```

### Tips
- Use `--tags` for efficient filtering at the API level
- Default limit is 200; max is 1000
- Search supports full-text across monitor names and queries
- Monitor states: OK, Alert, Warn, No Data

## APM / Traces

### Query Syntax
```
service:<name>                  # Filter by service
resource_name:<path>            # Filter by resource/endpoint
@duration:>5000000000           # Duration > 5 seconds (NANOSECONDS!)
status:error                    # Error traces only
operation_name:rack.request     # Filter by operation
env:production                  # Filter by environment
```

**CRITICAL: APM durations are in NANOSECONDS**
- 1 millisecond = 1,000,000 ns
- 1 second = 1,000,000,000 ns
- 5 seconds = 5,000,000,000 ns

### Commands
```bash
# List APM services
pup apm services list

# Search traces
pup traces search --query="service:web-api AND @duration:>1000000000" --from=1h

# List spans
pup traces list --query="service:web-api" --from=1h --limit=50
```

## RUM (Real User Monitoring)

### Query Syntax
```
@type:error                     # Error events
@session.type:user              # User sessions (not synthetic)
@view.url_path:/checkout        # Specific page
@action.type:click              # Click actions
service:<app-name>              # Filter by application
```

### Commands
```bash
# List RUM applications
pup rum apps list

# Search RUM events
pup rum events search --query="@type:error" --from=1h

# Aggregate RUM data
pup rum aggregate --query="@type:view" --from=4h --compute="avg(@view.loading_time)"
```

## Incidents

```bash
# List active incidents
pup incidents list --query="status:active"

# Get incident details
pup incidents get <incident-id>

# List incident timeline
pup incidents timeline <incident-id>
```

## SLOs

```bash
# List all SLOs
pup slos list

# Get SLO details
pup slos get <slo-id>

# Get SLO history
pup slos history <slo-id> --from=7d
```

## Security

```bash
# List security rules
pup security rules list

# Search security signals
pup security signals list --query="status:critical" --from=1d

# List security filters
pup security filters list
```

## Dashboards

```bash
# List dashboards
pup dashboards list

# Get dashboard details (includes all widgets and queries)
pup dashboards get <dashboard-id>
```

## Events

```bash
# List events
pup events list --from=1h

# Search events by source
pup events search --query="sources:pagerduty" --from=1d
```

## Common Patterns

### Error Investigation
```bash
# 1. Check for errors in logs
pup logs aggregate --query="status:error" --from=1h --compute="count" --group-by="service"

# 2. Drill into the affected service
pup logs search --query="status:error AND service:<name>" --from=1h --limit=20

# 3. Check monitors for that service
pup monitors list --tags="service:<name>"

# 4. Check recent deployments/events
pup events list --from=4h
```

### Performance Investigation
```bash
# 1. Check service latency
pup metrics query --query="avg:trace.servlet.request.duration{service:<name>} by {resource_name}" --from=1h

# 2. Find slow traces (>5 seconds)
pup traces search --query="service:<name> AND @duration:>5000000000" --from=1h

# 3. Check resource utilization
pup metrics query --query="avg:system.cpu.user{service:<name>} by {host}" --from=1h
```

## Time Ranges

All time-related flags (`--from`, `--to`) accept:

| Format | Example | Description |
|--------|---------|-------------|
| Relative short | `1h`, `30m`, `7d`, `5s` | Ago from now |
| Relative long | `5min`, `2hours`, `3days` | Ago from now |
| With spaces | `"5 minutes"`, `"2 hours"` | Ago from now |
| RFC3339 | `2024-01-01T00:00:00Z` | Absolute time |
| Unix ms | `1704067200000` | Milliseconds since epoch |
| Keyword | `now` | Current time |

## Output Formats

```bash
# JSON (default, recommended for agents)
pup monitors list --output=json

# Table (human-readable)
pup monitors list --output=table

# YAML
pup monitors list --output=yaml
```

## Error Handling

| Status | Meaning | Action |
|--------|---------|--------|
| 401 | Authentication failed | Run `pup auth login` or check API keys |
| 403 | Insufficient permissions | Verify API/App key permissions |
| 404 | Resource not found | Check the ID or resource name |
| 429 | Rate limited | Wait and retry with backoff |
| 5xx | Server error | Retry after a short delay |

## Agent Mode

Agent mode is auto-detected when running inside AI coding assistants (Claude Code, Cursor, Codex, etc.) or can be enabled explicitly:

```bash
# Explicit flag
pup --agent monitors list

# Environment variable
FORCE_AGENT_MODE=1 pup monitors list

# Auto-detected from: CLAUDECODE, CLAUDE_CODE, CURSOR_AGENT, CODEX, AIDER, etc.
```

In agent mode:
- Confirmation prompts are auto-approved (no stdin hangs)
- Output is JSON by default
- Structured error responses with suggestions
