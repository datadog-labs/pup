# LLM Agent Guide for Fetch CLI

This guide helps AI agents (LLMs) understand and effectively use the Fetch CLI tool.

## Quick Reference

### Discovery Commands

```bash
# See all available commands
fetch --help

# Get detailed help for any command
fetch <command> --help
fetch <command> <subcommand> --help

# Examples
fetch monitors --help
fetch monitors list --help
fetch auth login --help
```

### Authentication

```bash
# OAuth2 (Recommended)
fetch auth login     # Browser-based login
fetch auth status    # Check auth status
fetch auth refresh   # Refresh token
fetch auth logout    # Logout

# API Keys (Legacy)
export DD_API_KEY="..."
export DD_APP_KEY="..."
```

## Command Patterns

### List Resources

```bash
# General pattern
fetch <resource> list [--filters]

# Examples
fetch monitors list
fetch monitors list --name="CPU"
fetch monitors list --tags="env:production"

fetch dashboards list
fetch slos list
fetch incidents list
```

### Get Resource Details

```bash
# General pattern
fetch <resource> get <id>

# Examples
fetch monitors get 12345678
fetch dashboards get abc-def-123
fetch slos get abc-123-def
fetch incidents get abc-123-def
```

### Delete Resources

```bash
# General pattern (requires confirmation)
fetch <resource> delete <id>

# With auto-approve (no confirmation)
fetch <resource> delete <id> --yes

# Examples
fetch monitors delete 12345678 --yes
fetch dashboards delete abc-def-123 --yes
fetch slos delete abc-123-def --yes
```

## Parsing Output

All commands output JSON by default. Use `jq` for parsing:

```bash
# Get monitor names
fetch monitors list | jq '.[] | .name'

# Filter by field
fetch monitors list | jq '.[] | select(.overall_state == "Alert")'

# Extract specific fields
fetch monitors get 12345678 | jq '{name: .name, state: .overall_state}'

# Count resources
fetch dashboards list | jq '.dashboards | length'
```

## Common Tasks

### Monitor Management

```bash
# Find all critical monitors
fetch monitors list | jq '.[] | select(.overall_state == "Alert")'

# Get monitor by name pattern
fetch monitors list --name="CPU" | jq '.[0] | .id'

# Check monitor status
fetch monitors get 12345678 | jq '.overall_state'

# List production monitors
fetch monitors list --tags="env:production"
```

### Dashboard Management

```bash
# Find dashboard by name
fetch dashboards list | jq '.dashboards[] | select(.title | contains("API"))'

# Get dashboard ID
fetch dashboards list | jq '.dashboards[0] | .id'

# Backup dashboard
fetch dashboards get abc-def-123 > dashboard-backup.json

# Export all dashboards
fetch dashboards list | jq -r '.dashboards[].id' | \
  xargs -I {} fetch dashboards get {} > dashboards-{}.json
```

### SLO Monitoring

```bash
# Find breaching SLOs
fetch slos list | jq '.data[] | select(.status.state == "breaching")'

# Check error budget
fetch slos get abc-123 | jq '.data.error_budget_remaining'

# List all SLO statuses
fetch slos list | jq '.data[] | {name: .name, state: .status.state, budget: .status.error_budget_remaining}'
```

### Incident Response

```bash
# Find active incidents
fetch incidents list | jq '.data[] | select(.state == "active")'

# Find SEV-1 incidents
fetch incidents list | jq '.data[] | select(.severity == "SEV-1")'

# Get incident timeline
fetch incidents get abc-123 | jq '.data.timeline'

# Check customer impact
fetch incidents list | jq '.data[] | select(.customer_impacted == true)'
```

## Help Text Structure

Each command provides structured help with these sections:

1. **CAPABILITIES**: What the command can do
2. **EXAMPLES**: Real-world usage examples
3. **OUTPUT FIELDS**: Description of output structure
4. **FILTERS**: Available filtering options
5. **AUTHENTICATION**: Auth requirements

### Example Help Output

```bash
$ fetch monitors list --help

FILTERS:
  --name      Filter by monitor name (substring match)
  --tags      Filter by tags (comma-separated)

EXAMPLES:
  fetch monitors list
  fetch monitors list --name="CPU"
  fetch monitors list --tags="env:production"

OUTPUT FIELDS:
  • id: Monitor ID
  • name: Monitor name
  • type: Monitor type
  • query: Monitor query
  • overall_state: Current state
```

## Error Handling

```bash
# Commands return non-zero exit codes on error
fetch monitors get 99999999
echo $?  # Non-zero

# Capture errors
if ! fetch monitors get 99999999 2>&1; then
  echo "Monitor not found"
fi

# Parse error messages
fetch monitors get 99999999 2>&1 | grep "Error"
```

## Automation Patterns

### Confirmation Bypass

```bash
# Method 1: --yes flag
fetch monitors delete 12345678 --yes

# Method 2: Environment variable
DD_AUTO_APPROVE=true fetch monitors delete 12345678

# Method 3: Global flag
fetch monitors delete 12345678 -y
```

### Multi-Site Operations

```bash
# Operate on different sites
DD_SITE=datadoghq.com fetch monitors list
DD_SITE=datadoghq.eu fetch monitors list
DD_SITE=us3.datadoghq.com fetch monitors list
```

### Batch Operations

```bash
# Delete multiple monitors
for id in 111 222 333; do
  fetch monitors delete $id --yes
done

# Backup all dashboards
fetch dashboards list | jq -r '.dashboards[].id' | while read id; do
  fetch dashboards get "$id" > "dashboard-$id.json"
done
```

## LLM-Specific Tips

### 1. Always Check Help First

Before using any command, check its help text:

```bash
fetch <command> --help
fetch <command> <subcommand> --help
```

### 2. Use Structured Output

All commands output JSON. Parse with jq:

```bash
fetch <command> <subcommand> | jq '.'
```

### 3. Understand Authentication

Check if authenticated before running commands:

```bash
fetch auth status
```

### 4. Filter with jq, Not Grep

Use jq for structured filtering:

```bash
# Good
fetch monitors list | jq '.[] | select(.name | contains("CPU"))'

# Less good
fetch monitors list | grep "CPU"
```

### 5. Save Output for Analysis

Save command output to files for later analysis:

```bash
fetch monitors list > monitors.json
jq '.[] | select(.overall_state == "Alert")' monitors.json
```

## Resource Types

### Monitors
- **ID Format**: Numeric (e.g., 12345678)
- **List Command**: `fetch monitors list`
- **Get Command**: `fetch monitors get <id>`
- **Filter By**: name, tags

### Dashboards
- **ID Format**: UUID-like (e.g., abc-def-123)
- **List Command**: `fetch dashboards list`
- **Get Command**: `fetch dashboards get <id>`
- **Filter By**: None (use jq)

### SLOs
- **ID Format**: UUID-like (e.g., abc-123-def)
- **List Command**: `fetch slos list`
- **Get Command**: `fetch slos get <id>`
- **Filter By**: None (use jq)

### Incidents
- **ID Format**: UUID-like (e.g., abc-123-def)
- **List Command**: `fetch incidents list`
- **Get Command**: `fetch incidents get <id>`
- **Filter By**: None (use jq)

## Output Format

### JSON (Default)

```bash
fetch monitors list
# Returns: JSON array or object
```

### Table (Future)

```bash
fetch monitors list --output=table
# Returns: Formatted table
```

### YAML (Future)

```bash
fetch monitors list --output=yaml
# Returns: YAML format
```

## Troubleshooting

### Authentication Issues

```bash
# Check auth status
fetch auth status

# Re-authenticate
fetch auth login

# Check API keys (legacy)
echo $DD_API_KEY
echo $DD_APP_KEY
```

### Rate Limiting

If you encounter rate limits:
- Reduce request frequency
- Use filters to limit data retrieved
- Cache responses when possible

### Invalid IDs

```bash
# Verify resource exists
fetch monitors list | jq '.[] | .id'

# Then get specific resource
fetch monitors get <valid-id>
```

## Best Practices

1. **Always authenticate first**
   ```bash
   fetch auth login
   ```

2. **Use filters to reduce data**
   ```bash
   fetch monitors list --tags="env:production"
   ```

3. **Save responses for reuse**
   ```bash
   fetch monitors list > monitors.json
   ```

4. **Check help for each command**
   ```bash
   fetch <command> --help
   ```

5. **Use jq for parsing**
   ```bash
   fetch monitors list | jq '.[] | .name'
   ```

6. **Auto-approve for automation**
   ```bash
   fetch monitors delete <id> --yes
   ```

## Example Workflows

### Morning Health Check

```bash
# Check authentication
fetch auth status

# Check for alerts
fetch monitors list | jq '.[] | select(.overall_state == "Alert")'

# Check active incidents
fetch incidents list | jq '.data[] | select(.state == "active")'

# Check SLO breaches
fetch slos list | jq '.data[] | select(.status.state == "breaching")'
```

### Dashboard Backup

```bash
# List all dashboards
fetch dashboards list > dashboard-list.json

# Backup each dashboard
cat dashboard-list.json | jq -r '.dashboards[].id' | while read id; do
  fetch dashboards get "$id" > "backups/dashboard-$id.json"
  echo "Backed up dashboard $id"
done
```

### Monitor Audit

```bash
# Find untagged monitors
fetch monitors list | jq '.[] | select(.tags | length == 0)'

# Find monitors without notifications
fetch monitors list | jq '.[] | select(.message | contains("@") | not)'

# Find monitors in Alert state
fetch monitors list | jq '.[] | select(.overall_state == "Alert")'
```

## Additional Resources

- **Main Docs**: README.md
- **OAuth2 Guide**: docs/OAUTH2.md
- **Developer Guide**: CLAUDE.md
- **Implementation**: SUMMARY.md
