# Command Reference

Complete reference for all 42 command groups in Pup.

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
| auth | login, logout, status, refresh | src/commands/auth.rs | ✅ |
| metrics | query, list, get, search | src/commands/metrics.rs | ✅ |
| logs | search, list, aggregate | src/commands/logs.rs | ✅ |
| traces | - | - | ❌ |
| monitors | list, get, delete, search | src/commands/monitors.rs | ✅ |
| dashboards | list, get, delete, url | src/commands/dashboards.rs | ✅ |
| slos | list, get, delete, status | src/commands/slos.rs | ✅ |
| incidents | list, get, attachments, settings, handles, postmortem-templates | src/commands/incidents.rs | ✅ |
| rum | apps, metrics, retention-filters, sessions, playlists, heatmaps | src/commands/rum.rs | ✅ |
| cicd | pipelines, events, tests, dora, flaky-tests | src/commands/cicd.rs | ✅ |
| static-analysis | custom-rulesets | src/commands/static_analysis.rs | ✅ |
| downtime | list, get, cancel | src/commands/downtime.rs | ✅ |
| tags | list, get, add, update, delete | src/commands/tags.rs | ✅ |
| events | list, search, get | src/commands/events.rs | ✅ |
| on-call | teams (CRUD, memberships) | src/commands/on_call.rs | ✅ |
| audit-logs | list, search | src/commands/audit_logs.rs | ✅ |
| api-keys | list, get, create, delete | src/commands/api_keys.rs | ✅ |
| app-keys | list, get, create, update, delete | src/commands/app_keys.rs | ✅ |
| infrastructure | hosts (list, get) | src/commands/infrastructure.rs | ✅ |
| synthetics | tests, locations, suites | src/commands/synthetics.rs | ✅ |
| users | list, get, roles | src/commands/users.rs | ✅ |
| notebooks | list, get, delete | src/commands/notebooks.rs | ✅ |
| security | rules, signals, findings, content-packs, risk-scores | src/commands/security.rs | ✅ |
| organizations | get, list | src/commands/organizations.rs | ✅ |
| service-catalog | list, get | src/commands/service_catalog.rs | ✅ |
| error-tracking | issues (search, get) | src/commands/error_tracking.rs | ✅ |
| scorecards | list, get | src/commands/scorecards.rs | ✅ |
| usage | summary, hourly | src/commands/usage.rs | ✅ |
| apm | services (list, stats, operations, resources), entities (list), dependencies (list), flow-map | src/commands/apm.rs | ✅ |
| cost | projected, attribution, by-org | src/commands/cost.rs | ✅ |
| product-analytics | events send, analytics (scalar, timeseries), journey (funnel, timeseries, scalar, list, drop-off-analysis), retention (grid, timeseries, scalar, list, meta), sankey, segment (list, create, create-static, get, update, delete, templates) | src/commands/product_analytics.rs | ✅ |
| data-governance | scanner-rules (list) | src/commands/data_governance.rs | ✅ |
| obs-pipelines | list, get | src/commands/obs_pipelines.rs | ⏳ |
| network | flows, devices | src/commands/network.rs | ⏳ |
| cloud | aws, gcp, azure, oci | src/commands/cloud.rs | ✅ |
| integrations | slack, pagerduty, webhooks, jira, servicenow | src/commands/integrations.rs | ✅ |
| misc | ip-ranges, status | src/commands/misc.rs | ✅ |
| cases | create, get, search, assign, archive, projects, jira, servicenow, move | src/commands/cases.rs | ✅ |
| status-pages | pages, components, degradations | src/commands/status_pages.rs | ✅ |
| code-coverage | branch-summary, commit-summary | src/commands/code_coverage.rs | ✅ |
| hamr | connections (get, create) | src/commands/hamr.rs | ✅ |
| fleet | agents (list, get, versions), deployments (list, get, configure, upgrade, cancel), schedules (list, get, create, update, delete, trigger) | src/commands/fleet.rs | ✅ |

**Summary:** 38 working, 0 API-blocked, 2 placeholders

**Note:** RUM command is fully operational. Apps and sessions work completely. Metrics and retention-filters support list/get operations (create/update/delete operations pending due to complex API type structures).

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
- **cicd** - CI/CD visibility (pipelines, events, tests, dora, flaky-tests)
- **code-coverage** - Code coverage summaries (branch, commit)
- **error-tracking** - Error management (issues search, issues get)
- **scorecards** - Service quality (list, get)
- **service-catalog** - Service registry (list, get)

### Operations & Incident Response
- **incidents** - Incident management (list, get, attachments, settings, handles, postmortem-templates)
- **on-call** - Team management (create, update, delete teams; manage memberships with roles)
- **cases** - Case management (create, search, assign, archive, projects, jira, servicenow, move)
- **hamr** - High Availability Multi-Region connections
- **fleet** - Fleet Automation (agents, deployments, schedules)

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
- **product-analytics** - Product analytics (events, analytics queries, journey analysis, retention analysis, sankey flows, segment management)

## Global Flags

Available on all commands:

```bash
--config string      Config file path (default: ~/.config/pup/config.yaml)
--site string        Datadog site (default: datadoghq.com)
--output string      Output format: json, yaml, table (default: json)
--verbose            Enable verbose logging
--yes                Skip confirmation prompts
```

## Recent Enhancements

Recent API client updates added 3 new command groups and ~60 new subcommands across 9 existing domains.

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
