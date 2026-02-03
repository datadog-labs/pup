# LLM Agent Guide for Pup CLI

This guide helps AI agents (LLMs) understand and effectively use the Pup CLI tool.

## Quick Reference

### Discovery Commands

```bash
# See all available commands
pup --help

# Get detailed help for any command
pup <command> --help
pup <command> <subcommand> --help

# Examples
pup monitors --help
pup monitors list --help
pup auth login --help
```

### Authentication

```bash
# OAuth2 (Recommended)
pup auth login     # Browser-based login
pup auth status    # Check auth status
pup auth refresh   # Refresh token
pup auth logout    # Logout

# API Keys (Legacy)
export DD_API_KEY="..."
export DD_APP_KEY="..."
```

## Command Patterns

### List Resources

```bash
# General pattern
pup <resource> list [--filters]

# Examples
pup monitors list
pup monitors list --name="CPU"
pup monitors list --tags="env:production"

pup dashboards list
pup slos list
pup incidents list
```

### Get Resource Details

```bash
# General pattern
pup <resource> get <id>

# Examples
pup monitors get 12345678
pup dashboards get abc-def-123
pup slos get abc-123-def
pup incidents get abc-123-def
```

### Delete Resources

```bash
# General pattern (requires confirmation)
pup <resource> delete <id>

# With auto-approve (no confirmation)
pup <resource> delete <id> --yes

# Examples
pup monitors delete 12345678 --yes
pup dashboards delete abc-def-123 --yes
pup slos delete abc-123-def --yes
```

## Parsing Output

All commands output JSON by default. Use `jq` for parsing:

```bash
# Get monitor names
pup monitors list | jq '.[] | .name'

# Filter by field
pup monitors list | jq '.[] | select(.overall_state == "Alert")'

# Extract specific fields
pup monitors get 12345678 | jq '{name: .name, state: .overall_state}'

# Count resources
pup dashboards list | jq '.dashboards | length'
```

## Common Tasks

### Monitor Management

```bash
# Find all critical monitors
pup monitors list | jq '.[] | select(.overall_state == "Alert")'

# Get monitor by name pattern
pup monitors list --name="CPU" | jq '.[0] | .id'

# Check monitor status
pup monitors get 12345678 | jq '.overall_state'

# List production monitors
pup monitors list --tags="env:production"
```

### Dashboard Management

```bash
# Find dashboard by name
pup dashboards list | jq '.dashboards[] | select(.title | contains("API"))'

# Get dashboard ID
pup dashboards list | jq '.dashboards[0] | .id'

# Backup dashboard
pup dashboards get abc-def-123 > dashboard-backup.json

# Export all dashboards
pup dashboards list | jq -r '.dashboards[].id' | \
  xargs -I {} pup dashboards get {} > dashboards-{}.json
```

### SLO Monitoring

```bash
# Find breaching SLOs
pup slos list | jq '.data[] | select(.status.state == "breaching")'

# Check error budget
pup slos get abc-123 | jq '.data.error_budget_remaining'

# List all SLO statuses
pup slos list | jq '.data[] | {name: .name, state: .status.state, budget: .status.error_budget_remaining}'
```

### Incident Response

```bash
# Find active incidents
pup incidents list | jq '.data[] | select(.state == "active")'

# Find SEV-1 incidents
pup incidents list | jq '.data[] | select(.severity == "SEV-1")'

# Get incident timeline
pup incidents get abc-123 | jq '.data.timeline'

# Check customer impact
pup incidents list | jq '.data[] | select(.customer_impacted == true)'
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
$ pup monitors list --help

FILTERS:
  --name      Filter by monitor name (substring match)
  --tags      Filter by tags (comma-separated)

EXAMPLES:
  pup monitors list
  pup monitors list --name="CPU"
  pup monitors list --tags="env:production"

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
pup monitors get 99999999
echo $?  # Non-zero

# Capture errors
if ! pup monitors get 99999999 2>&1; then
  echo "Monitor not found"
fi

# Parse error messages
pup monitors get 99999999 2>&1 | grep "Error"
```

## Automation Patterns

### Confirmation Bypass

```bash
# Method 1: --yes flag
pup monitors delete 12345678 --yes

# Method 2: Environment variable
DD_AUTO_APPROVE=true pup monitors delete 12345678

# Method 3: Global flag
pup monitors delete 12345678 -y
```

### Multi-Site Operations

```bash
# Operate on different sites
DD_SITE=datadoghq.com pup monitors list
DD_SITE=datadoghq.eu pup monitors list
DD_SITE=us3.datadoghq.com pup monitors list
```

### Batch Operations

```bash
# Delete multiple monitors
for id in 111 222 333; do
  pup monitors delete $id --yes
done

# Backup all dashboards
pup dashboards list | jq -r '.dashboards[].id' | while read id; do
  pup dashboards get "$id" > "dashboard-$id.json"
done
```

## LLM-Specific Tips

### 1. Always Check Help First

Before using any command, check its help text:

```bash
pup <command> --help
pup <command> <subcommand> --help
```

### 2. Use Structured Output

All commands output JSON. Parse with jq:

```bash
pup <command> <subcommand> | jq '.'
```

### 3. Understand Authentication

Check if authenticated before running commands:

```bash
pup auth status
```

### 4. Filter with jq, Not Grep

Use jq for structured filtering:

```bash
# Good
pup monitors list | jq '.[] | select(.name | contains("CPU"))'

# Less good
pup monitors list | grep "CPU"
```

### 5. Save Output for Analysis

Save command output to files for later analysis:

```bash
pup monitors list > monitors.json
jq '.[] | select(.overall_state == "Alert")' monitors.json
```

## Resource Types

### Monitors
- **ID Format**: Numeric (e.g., 12345678)
- **List Command**: `pup monitors list`
- **Get Command**: `pup monitors get <id>`
- **Filter By**: name, tags

### Dashboards
- **ID Format**: UUID-like (e.g., abc-def-123)
- **List Command**: `pup dashboards list`
- **Get Command**: `pup dashboards get <id>`
- **Filter By**: None (use jq)

### SLOs
- **ID Format**: UUID-like (e.g., abc-123-def)
- **List Command**: `pup slos list`
- **Get Command**: `pup slos get <id>`
- **Filter By**: None (use jq)

### Incidents
- **ID Format**: UUID-like (e.g., abc-123-def)
- **List Command**: `pup incidents list`
- **Get Command**: `pup incidents get <id>`
- **Filter By**: None (use jq)

## Output Format

### JSON (Default)

```bash
pup monitors list
# Returns: JSON array or object
```

### Table (Future)

```bash
pup monitors list --output=table
# Returns: Formatted table
```

### YAML (Future)

```bash
pup monitors list --output=yaml
# Returns: YAML format
```

## Troubleshooting

### Authentication Issues

```bash
# Check auth status
pup auth status

# Re-authenticate
pup auth login

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
pup monitors list | jq '.[] | .id'

# Then get specific resource
pup monitors get <valid-id>
```

## Best Practices

1. **Always authenticate first**
   ```bash
   pup auth login
   ```

2. **Use filters to reduce data**
   ```bash
   pup monitors list --tags="env:production"
   ```

3. **Save responses for reuse**
   ```bash
   pup monitors list > monitors.json
   ```

4. **Check help for each command**
   ```bash
   pup <command> --help
   ```

5. **Use jq for parsing**
   ```bash
   pup monitors list | jq '.[] | .name'
   ```

6. **Auto-approve for automation**
   ```bash
   pup monitors delete <id> --yes
   ```

## Example Workflows

### Morning Health Check

```bash
# Check authentication
pup auth status

# Check for alerts
pup monitors list | jq '.[] | select(.overall_state == "Alert")'

# Check active incidents
pup incidents list | jq '.data[] | select(.state == "active")'

# Check SLO breaches
pup slos list | jq '.data[] | select(.status.state == "breaching")'
```

### Dashboard Backup

```bash
# List all dashboards
pup dashboards list > dashboard-list.json

# Backup each dashboard
cat dashboard-list.json | jq -r '.dashboards[].id' | while read id; do
  pup dashboards get "$id" > "backups/dashboard-$id.json"
  echo "Backed up dashboard $id"
done
```

### Monitor Audit

```bash
# Find untagged monitors
pup monitors list | jq '.[] | select(.tags | length == 0)'

# Find monitors without notifications
pup monitors list | jq '.[] | select(.message | contains("@") | not)'

# Find monitors in Alert state
pup monitors list | jq '.[] | select(.overall_state == "Alert")'
```

## Additional Resources

- **Main Docs**: README.md
- **OAuth2 Guide**: docs/OAUTH2.md
- **Developer Guide**: CLAUDE.md
- **Implementation**: SUMMARY.md
