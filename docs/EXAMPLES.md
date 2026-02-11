# Usage Examples

Common workflows and usage patterns for Pup CLI.

## Authentication

### OAuth2 Login (Recommended)
```bash
# Login with default site (datadoghq.com)
pup auth login

# Login with specific site
pup --site=datadoghq.eu auth login

# Check authentication status
pup auth status

# Logout
pup auth logout
```

### API Key Authentication (Legacy)
```bash
export DD_API_KEY="your-api-key"
export DD_APP_KEY="your-app-key"
export DD_SITE="datadoghq.com"
```

## Metrics

### List Metrics
```bash
# List all metrics
pup metrics list

# Filter by pattern
pup metrics list --filter="system.*"
pup metrics list --filter="custom.app.*"
```

### Search Metrics (v1 API)
```bash
# Classic query syntax
pup metrics search --query="avg:system.cpu.user{*}" --from="1h"

# Search with aggregation and grouping
pup metrics search --query="sum:app.requests{env:prod} by {service}" --from="4h"
```

### Query Metrics (v2 API)
```bash
# Timeseries formula query
pup metrics query --query="avg:system.cpu.user{*}" --from="1h" --to="now"

# Query with aggregation
pup metrics query --query="sum:app.requests{env:prod} by {service}" --from="4h"
```

## Monitors

### List Monitors
```bash
# List all monitors
pup monitors list

# Filter by tag
pup monitors list --tag="env:production"
pup monitors list --tag="team:backend"

# Multiple tags
pup monitors list --tag="env:prod" --tag="service:api"
```

### Get Monitor Details
```bash
# Get specific monitor by ID
pup monitors get 12345678
```

### Delete Monitor
```bash
# Delete monitor (prompts for confirmation)
pup monitors delete 12345678

# Skip confirmation
pup monitors delete 12345678 --yes
```

## Logs

### Search Logs
```bash
# Search for errors in last hour
pup logs search --query="status:error" --from="1h" --to="now"

# Search by service
pup logs search --query="service:web-app status:warn" --from="30m"

# Complex query with attributes
pup logs search --query="@user.id:12345 status:error" --limit=100

# Search with time range
pup logs search \
  --query="service:api" \
  --from="2024-02-04T10:00:00Z" \
  --to="2024-02-04T11:00:00Z"
```

### Aggregate Logs
```bash
# Count logs by status
pup logs aggregate \
  --query="service:web-app" \
  --from="1h" \
  --compute="count:*" \
  --group-by="status"
```

## Dashboards

### List Dashboards
```bash
# List all dashboards
pup dashboards list

# Output as table
pup dashboards list --output=table
```

### Get Dashboard
```bash
# Get dashboard details
pup dashboards get "abc-123-def"

# Get public URL for sharing
pup dashboards url "abc-123-def"
```

### Delete Dashboard
```bash
pup dashboards delete "abc-123-def" --yes
```

## SLOs

### List SLOs
```bash
# List all SLOs
pup slos list

# Filter by tag
pup slos list --tag="service:api"
```

### Get SLO
```bash
pup slos get "abc-123-def"
```

### Create SLO
```bash
pup slos create \
  --name="API Availability" \
  --type="metric" \
  --target=99.9 \
  --timeframe="30d"
```

### Manage SLO Corrections
```bash
# List corrections for SLO
pup slos corrections list "abc-123-def"

# Create correction
pup slos corrections create "abc-123-def" \
  --start="2024-02-04T10:00:00Z" \
  --end="2024-02-04T11:00:00Z" \
  --category="deployment"
```

## Incidents

### List Incidents
```bash
# List all incidents
pup incidents list

# Filter by status
pup incidents list --status="active"
```

### Get Incident
```bash
pup incidents get "abc-123"
```

### Create Incident
```bash
pup incidents create \
  --title="High Error Rate in API" \
  --severity="SEV-2" \
  --customer-impacted=true
```

### Update Incident
```bash
pup incidents update "abc-123" --status="resolved"
```

## RUM (Real User Monitoring)

### List RUM Applications
```bash
pup rum apps list
```

### Get RUM Application
```bash
pup rum apps get "abc-123"
```

### Search RUM Sessions
```bash
pup rum sessions search \
  --query="@application.id:abc-123" \
  --from="1h"
```

## Security

### List Security Rules
```bash
pup security rules list
```

### Get Security Rule
```bash
pup security rules get "abc-123"
```

### List Security Signals
```bash
pup security signals list --from="1h"
```

### Search Security Findings
```bash
pup security findings search \
  --query="@severity:high"
```

## Infrastructure

### List Hosts
```bash
# List all hosts
pup infrastructure hosts list

# Filter by tag
pup infrastructure hosts list --filter="env:production"
```

### Get Host
```bash
pup infrastructure hosts get "host-name"
```

## Tags

### List Host Tags
```bash
# List all host tags
pup tags list
```

### Get Tags for Host
```bash
pup tags get "host-name"
```

### Add Tags to Host
```bash
pup tags add "host-name" \
  --tag="env:production" \
  --tag="team:backend"
```

### Update Host Tags
```bash
pup tags update "host-name" \
  --tag="env:prod" \
  --tag="service:api"
```

## Users & Organizations

### List Users
```bash
pup users list
```

### Get User
```bash
pup users get "user-id"
```

### List Roles
```bash
pup users roles list
```

### Get Organization
```bash
pup organizations get
```

## API Keys

### List API Keys
```bash
pup api-keys list
```

### Get API Key
```bash
pup api-keys get "key-id"
```

### Create API Key
```bash
pup api-keys create --name="CI/CD Key"
```

### Delete API Key
```bash
pup api-keys delete "key-id" --yes
```

## Synthetics

### List Synthetic Tests
```bash
pup synthetics tests list
```

### Get Synthetic Test
```bash
pup synthetics tests get "test-id"
```

### List Synthetic Locations
```bash
pup synthetics locations list
```

## Output Formatting

### JSON Output (Default)
```bash
pup monitors list --output=json
```

### YAML Output
```bash
pup monitors list --output=yaml
```

### Table Output
```bash
pup monitors list --output=table
```

### Custom Fields
```bash
pup monitors list --fields="id,name,type,status"
```

## Advanced Usage

### Custom Config File
```bash
pup --config=/path/to/config.yaml monitors list
```

### Specify Datadog Site
```bash
pup --site=datadoghq.eu monitors list
```

### Verbose Output (Debug)
```bash
pup --verbose monitors list
```

### Skip Confirmation Prompts
```bash
pup --yes monitors delete 12345678
```

## Common Workflows

### Monitoring Dashboard
```bash
# List monitors for a service
pup monitors list --tag="service:api" --output=table

# Check recent logs
pup logs search --query="service:api" --from="1h" --output=table

# Query metrics
pup metrics query --query="avg:api.latency{*}" --from="1h"
```

### Incident Response
```bash
# Create incident
pup incidents create --title="API Down" --severity="SEV-1"

# Search related logs
pup logs search --query="status:error service:api" --from="1h"

# Check monitors
pup monitors list --tag="service:api"

# Update incident status
pup incidents update "incident-id" --status="investigating"
```

### Security Audit
```bash
# List recent security signals
pup security signals list --from="24h"

# Check security rules
pup security rules list

# Search security findings
pup security findings search --query="@severity:critical"

# Review audit logs
pup audit-logs list --from="7d"
```

### Infrastructure Review
```bash
# List all hosts
pup infrastructure hosts list --output=table

# Get host details
pup infrastructure hosts get "host-name"

# Review host tags
pup tags list --output=table
```

## Time Range Formats

### Relative Times
```bash
--from="1h"    # 1 hour ago
--from="30m"   # 30 minutes ago
--from="7d"    # 7 days ago
--from="now"   # Current time
```

### Absolute Times
```bash
--from="2024-02-04T10:00:00Z"
--to="2024-02-04T11:00:00Z"
```

### Unix Timestamps
```bash
--from="1707048000"  # Unix timestamp in seconds
```

## Environment Variables

```bash
# Authentication
export DD_API_KEY="your-api-key"
export DD_APP_KEY="your-app-key"
export DD_SITE="datadoghq.com"

# Configuration
export PUP_CONFIG="/path/to/config.yaml"
export PUP_OUTPUT="json"
export PUP_LOG_LEVEL="debug"
```

## Configuration File

Create `~/.config/pup/config.yaml`:

```yaml
site: datadoghq.com
output: json
verbose: false

# Default time ranges
default_from: 1h
default_to: now

# Output preferences
output_format: json
table_max_width: 120
```
