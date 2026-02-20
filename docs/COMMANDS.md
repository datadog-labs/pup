# Command Reference

Complete reference for all 41 command groups in Pup.

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
| traces | - | cmd/traces_simple.go | ❌ |
| monitors | list, get, delete, search | cmd/monitors.go | ✅ |
| dashboards | list, get, delete, url | cmd/dashboards.go | ✅ |
| slos | list, get, delete, status | cmd/slos.go | ✅ |
| incidents | list, get, attachments, settings, handles, postmortem-templates | cmd/incidents.go | ✅ |
| rum | apps, metrics, retention-filters, sessions, playlists, heatmaps | cmd/rum.go | ✅ |
| cicd | pipelines, events, dora, flaky-tests | cmd/cicd.go | ✅ |
| static-analysis | custom-rulesets | cmd/vulnerabilities.go | ✅ |
| downtime | list, get, cancel | cmd/downtime.go | ✅ |
| tags | list, get, add, update, delete | cmd/tags.go | ✅ |
| events | list, search, get | cmd/events.go | ✅ |
| on-call | teams (CRUD, memberships) | cmd/on_call.go | ✅ |
| audit-logs | list, search | cmd/audit_logs.go | ✅ |
| api-keys | list, get, create, delete | cmd/api_keys.go | ✅ |
| app-keys | list, get, create, update, delete | cmd/app_keys.go | ✅ |
| infrastructure | hosts (list, get) | cmd/infrastructure.go | ✅ |
| synthetics | tests, locations, suites | cmd/synthetics.go | ✅ |
| users | list, get, roles | cmd/users.go | ✅ |
| notebooks | list, get, delete | cmd/notebooks.go | ✅ |
| security | rules, signals, findings, content-packs, risk-scores | cmd/security.go | ✅ |
| organizations | get, list | cmd/organizations.go | ✅ |
| service-catalog | list, get | cmd/service_catalog.go | ✅ |
| error-tracking | issues (search, get) | cmd/error_tracking.go | ✅ |
| scorecards | list, get | cmd/scorecards.go | ✅ |
| usage | summary, hourly | cmd/usage.go | ✅ |
| apm | services (list, stats, operations, resources), entities (list), dependencies (list), flow-map | cmd/apm.go | ✅ |
| cost | projected, attribution, by-org | cmd/cost.go | ✅ |
| product-analytics | events send | cmd/product_analytics.go | ✅ |
| data-governance | scanner-rules (list) | cmd/data_governance.go | ✅ |
| obs-pipelines | list, get | cmd/obs_pipelines.go | ⏳ |
| network | flows, devices | cmd/network.go | ⏳ |
| cloud | aws, gcp, azure, oci | cmd/cloud.go | ✅ |
| integrations | slack, pagerduty, webhooks, jira, servicenow | cmd/integrations.go | ✅ |
| misc | ip-ranges, status | cmd/miscellaneous.go | ✅ |
| cases | create, get, search, assign, archive, projects, jira, servicenow, move | cmd/cases.go | ✅ |
| status-pages | pages, components, degradations | cmd/status_pages.go | ✅ |
| code-coverage | branch-summary, commit-summary | cmd/code_coverage.go | ✅ |
| hamr | connections (get, create) | cmd/hamr.go | ✅ |

**Summary:** 37 working, 0 API-blocked, 2 placeholders

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
pup logs search --query="service:api" --from="7d" --storage="flex"
pup metrics search --query="avg:system.cpu.user{*}" --from="1h"
pup metrics query --query="avg:system.cpu.user{*}" --from="1h"
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
- **traces** - APM traces (not yet implemented - use `apm` commands instead)
- **rum** - Real User Monitoring (apps, metrics, retention-filters, sessions)
- **events** - Infrastructure events (list, search, get)

### Monitoring & Alerting
- **monitors** - Monitor management (list, get, delete)
- **dashboards** - Dashboard management (list, get, delete, url)
- **slos** - Service Level Objectives (list, get, delete, status)
- **synthetics** - Synthetic monitoring (tests, locations, suites)
- **notebooks** - Investigation notebooks (list, get, delete)
- **downtime** - Monitor downtime (list, get, cancel)
- **status-pages** - Status pages with components and degradations

### Infrastructure & Performance
- **infrastructure** - Host inventory (hosts list, hosts get)
- **network** - Network monitoring (flows list, devices list)
- **tags** - Host tag management (list, get, add, update, delete)

### Security & Compliance
- **security** - Security monitoring (rules, signals, findings, content-packs, risk-scores)
- **static-analysis** - Code security (ast, custom-rulesets, sca, coverage)
- **audit-logs** - Audit trail (list, search)
- **data-governance** - Sensitive data scanning (scanner-rules list)

### Cloud & Integrations
- **cloud** - Cloud providers (aws, gcp, azure, oci)
- **integrations** - Third-party integrations (slack, pagerduty, webhooks, jira, servicenow)

### Development & Quality
- **cicd** - CI/CD visibility (pipelines, events, dora, flaky-tests)
- **code-coverage** - Code coverage summaries (branch, commit)
- **error-tracking** - Error management (issues search, issues get)
- **scorecards** - Service quality (list, get)
- **service-catalog** - Service registry (list, get)

### Operations & Incident Response
- **incidents** - Incident management (list, get, attachments, settings, handles, postmortem-templates)
- **on-call** - Team management (create, update, delete teams; manage memberships with roles)
- **cases** - Case management (create, search, assign, archive, projects, jira, servicenow, move)
- **hamr** - High Availability Multi-Region connections

### Organization & Access
- **users** - User management (list, get, roles)
- **organizations** - Org settings (get, list)
- **api-keys** - API key management (list, get, create, delete)
- **app-keys** - Application key management (list, get, create, update, delete)

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

## Recent Enhancements (v2.55.0 API Client Update)

The upgrade to datadog-api-client-go v2.55.0 adds 3 new command groups and ~60 new subcommands across 9 existing domains.

### New Command Groups
- ✅ **status-pages** - Status page management (pages, components, degradations CRUD)
- ✅ **code-coverage** - Code coverage summaries (branch-level and commit-level)
- ✅ **hamr** - High Availability Multi-Region connections

### Enhanced Existing Commands
- **integrations** - Added Jira integration (accounts, templates CRUD) and ServiceNow integration (instances, templates, users, assignment groups, business services)
- **cloud** - Added OCI integration (tenancy configs CRUD, products)
- **synthetics** - Added suites management (V2 API: search, get, create, update, delete)
- **security** - Added content packs (list, activate, deactivate), bulk rule export, and entity risk scores
- **incidents** - Added global settings, handles, and postmortem template management
- **cases** - Added Jira/ServiceNow issue linking, case project moves, and notification rules
- **cicd** - Added DORA deployment patching and flaky tests management
- **slos** - Added SLO status query (V2 API)
- **rum** - Replaced playlist/heatmap placeholders with working RUM Replay API implementations
