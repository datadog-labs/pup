# Command Reference

Complete reference for all 38 command groups in Pup.

## Command Pattern

```bash
pup <domain> <action> [options]           # Simple commands
pup <domain> <subgroup> <action> [options] # Nested commands
```

## Status Legend

- ✅ **WORKING** - Command compiles and runs (requires valid auth)
- ⚠️ **API BLOCKED** - Implementation correct, waiting for API client library updates
- ⏳ **PLACEHOLDER** - Skeleton implementation, API endpoints pending

## Command Index

| Domain | Subcommands | File | Status |
|--------|-------------|------|--------|
| auth | login, logout, status, refresh | cmd/auth.go | ✅ |
| metrics | query, list, get, search | cmd/metrics.go | ✅ |
| logs | search, list, aggregate | cmd/logs.go | ✅ |
| traces | search, list, aggregate | cmd/traces.go | ✅ |
| monitors | list, get, delete, search | cmd/monitors.go | ✅ |
| dashboards | list, get, delete, url | cmd/dashboards.go | ✅ |
| slos | list, get, create, update, delete, corrections | cmd/slos.go | ✅ |
| incidents | list, get, attachments | cmd/incidents.go | ✅ |
| rum | apps, sessions | cmd/rum.go | ✅ |
| cicd | pipelines, events | cmd/cicd.go | ✅ |
| vulnerabilities | search, list | cmd/vulnerabilities.go | ✅ |
| static-analysis | custom-rulesets | cmd/vulnerabilities.go | ✅ |
| downtime | list, get, cancel | cmd/downtime.go | ✅ |
| tags | list, get, add, update, delete | cmd/tags.go | ✅ |
| events | list, search, get | cmd/events.go | ✅ |
| on-call | teams (CRUD, memberships) | cmd/on_call.go | ✅ |
| audit-logs | list, search | cmd/audit_logs.go | ✅ |
| api-keys | list, get, create, delete | cmd/api_keys.go | ✅ |
| app-keys | list, get, register, unregister | cmd/app_keys.go | ✅ |
| infrastructure | hosts (list, get) | cmd/infrastructure.go | ✅ |
| synthetics | tests, locations | cmd/synthetics.go | ✅ |
| users | list, get, roles | cmd/users.go | ✅ |
| notebooks | list, get, delete | cmd/notebooks.go | ✅ |
| security | rules, signals, findings (list, get, search) | cmd/security.go | ✅ |
| organizations | get, list | cmd/organizations.go | ✅ |
| service-catalog | list, get | cmd/service_catalog.go | ✅ |
| error-tracking | issues (list, get) | cmd/error_tracking.go | ✅ |
| scorecards | list, get | cmd/scorecards.go | ✅ |
| usage | summary, hourly | cmd/usage.go | ✅ |
| apm | services (list, stats, operations, resources), entities (list), dependencies (list), flow-map | cmd/apm.go | ✅ |
| cost | projected, attribution, by-org | cmd/cost.go | ✅ |
| product-analytics | events send | cmd/product_analytics.go | ✅ |
| data-governance | scanner-rules (list) | cmd/data_governance.go | ✅ |
| obs-pipelines | list, get | cmd/obs_pipelines.go | ⏳ |
| network | flows, devices | cmd/network.go | ⏳ |
| cloud | aws, gcp, azure (list) | cmd/cloud.go | ✅ |
| integrations | slack, pagerduty, webhooks | cmd/integrations.go | ✅ |
| misc | ip-ranges, status | cmd/miscellaneous.go | ✅ |
| cases | create, get, search, assign, archive, projects | cmd/cases.go | ✅ |

**Summary:** 34 working, 0 API-blocked, 3 placeholders

**Note:** RUM command (cmd/rum.go) is fully operational. Apps and sessions work completely. Metrics and retention-filters support list/get operations (create/update/delete operations pending due to complex API type structures).

## Common Patterns

### List Operations
```bash
pup <domain> list [--flags]
pup monitors list --tags="env:production"
pup dashboards list
```

### Get Operations
```bash
pup <domain> get <id>
pup monitors get 12345678
pup slos get abc-123-def
```

### Search/Query
```bash
pup logs search --query="status:error" --from="1h"
pup metrics query --query="avg:system.cpu.user{*}"
pup events search --query="@user.id:12345"
```

### Create/Update/Delete
```bash
pup <domain> create [--flags]
pup <domain> update <id> [--flags]
pup <domain> delete <id> [--yes]
```

### Nested Commands
```bash
pup rum apps list
pup rum metrics get <id>
pup cicd pipelines list
pup security rules list
pup infrastructure hosts list
```

## Domain Categories

### Data & Observability
- **metrics** - Time-series metrics (query, list, get, search)
- **logs** - Log search and analysis (search, list, aggregate)
- **traces** - APM traces (search, list, aggregate)
- **rum** - Real User Monitoring (apps, metrics, retention-filters, sessions)
- **events** - Infrastructure events (list, search, get)

### Monitoring & Alerting
- **monitors** - Monitor management (list, get, delete)
- **dashboards** - Dashboard management (list, get, delete, url)
- **slos** - Service Level Objectives (list, get, create, update, delete)
- **synthetics** - Synthetic monitoring (tests, locations)
- **notebooks** - Investigation notebooks (list, get, delete)
- **downtime** - Monitor downtime (list, get, cancel)

### Infrastructure & Performance
- **infrastructure** - Host inventory (hosts list, hosts get)
- **network** - Network monitoring (flows list, devices list)
- **tags** - Host tag management (list, get, add, update, delete)

### Security & Compliance
- **security** - Security monitoring (rules, signals, findings)
- **vulnerabilities** - Vulnerability management (search, list)
- **static-analysis** - Code security (ast, custom-rulesets, sca, coverage)
- **audit-logs** - Audit trail (list, search)
- **data-governance** - Sensitive data scanning (scanner-rules list)

### Cloud & Integrations
- **cloud** - Cloud providers (aws, gcp, azure)
- **integrations** - Third-party integrations (slack, pagerduty, webhooks)

### Development & Quality
- **cicd** - CI/CD visibility (pipelines, events)
- **error-tracking** - Error management (issues list, issues get)
- **scorecards** - Service quality (list, get)
- **service-catalog** - Service registry (list, get)

### Operations & Incident Response
- **incidents** - Incident management (list, get, attachments)
- **on-call** - Team management (create, update, delete teams; manage memberships with roles)
- **cases** - Case management (create, search, assign, archive, projects)

### Organization & Access
- **users** - User management (list, get, roles)
- **organizations** - Org settings (get, list)
- **api-keys** - API key management (list, get, create, delete)
- **app-keys** - App key registration for Action Connections (list, get, register, unregister)

### Cost & Usage
- **usage** - Usage and billing (summary, hourly)
- **cost** - Cost management (projected, attribution, by-org)

### Configuration & Data Management
- **obs-pipelines** - Observability pipelines (list, get)
- **misc** - Miscellaneous (ip-ranges, status)
- **product-analytics** - Product analytics events (send)

## Global Flags

Available on all commands:

```bash
--config string      Config file path (default: ~/.config/pup/config.yaml)
--site string        Datadog site (default: datadoghq.com)
--output string      Output format: json, yaml, table (default: json)
--verbose            Enable verbose logging
--yes                Skip confirmation prompts
```

## Recent Enhancements (v2.54.0 API Client Update)

The upgrade to datadog-api-client-go v2.54.0 has resolved all previous API blocking issues and added new capabilities:

### Newly Unblocked Commands
All 7 previously blocked commands now work:
- ✅ **audit-logs** - Full audit log search and listing
- ✅ **cicd** - CI/CD pipeline visibility and events
- ✅ **events** - Infrastructure event management
- ✅ **tags** - Host tag operations
- ✅ **usage** - Usage and billing metrics
- ✅ **vulnerabilities** - Security vulnerability tracking
- ✅ **static-analysis** - Code security analysis

### New Command Groups
- ✅ **app-keys** - App key registration for Action Connections and Workflow Automation
- ✅ **cost** - Cost management with projected costs, attribution by tags, and organizational breakdowns
- ✅ **product-analytics** - Send server-side product analytics events with custom properties

### Enhanced Existing Commands
- **security findings** - Now includes get and search capabilities with advanced filtering (severity, status, resource type, evaluation results)
