# Traces Command Design

Resolves [#49](https://github.com/DataDog/pup/issues/49) — traces command is a placeholder returning "under development".

## Context

The `traces` command (`cmd/traces_simple.go`) is a 22-line placeholder. The Datadog API client (v2.54.0) provides a typed `SpansApi` with full search, list, and aggregate capabilities. The existing `apm` command covers service-level aggregated data (services, operations, dependencies, flow-map). The `traces` command complements `apm` by providing individual span-level querying.

**Relationship to `apm`:**
- `apm` = "Which services have problems?" (bird's eye, service-level aggregates)
- `traces` = "Show me the actual spans" (ground-level, individual span data)

Typical agent investigation flow: `apm services stats` (triage) -> `apm dependencies list` (topology) -> `traces search` (drill into specific spans) -> `traces aggregate` (compute stats over spans).

## Scope

- **In scope:** `traces search` and `traces aggregate` using the typed `datadogV2.SpansApi`
- **Out of scope:** SpansMetrics CRUD (separate future work), unifying time range flags across commands (separate issue)

## Command Interface

### traces search

Search for individual spans matching a query. Uses `SpansApi.ListSpans` (POST) with auto-pagination.

```bash
pup traces search --query "service:web-server @http.status_code:500" --from 1h --to now --limit 50 --sort -timestamp
```

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--query` | string | `"*"` | no | Span search syntax query |
| `--from` | string | `"1h"` | no | Start time (relative, RFC3339, unix ms) |
| `--to` | string | `"now"` | no | End time |
| `--limit` | int | `50` | no | Max spans to return (auto-paginates) |
| `--sort` | string | `"-timestamp"` | no | Sort: `timestamp` or `-timestamp` |

### traces aggregate

Compute aggregated stats over spans. Uses `SpansApi.AggregateSpans` (POST).

```bash
pup traces aggregate --query "service:web-server" --compute count --group-by @service --from 1h --to now
```

| Flag | Type | Default | Required | Description |
|------|------|---------|----------|-------------|
| `--query` | string | `"*"` | no | Span search syntax query |
| `--from` | string | `"1h"` | no | Start time |
| `--to` | string | `"now"` | no | End time |
| `--compute` | string | -- | yes | Aggregation: `count`, `avg:@duration`, `sum:@duration`, `min:@duration`, `max:@duration`, `pc90:@duration`, etc. |
| `--group-by` | string | -- | no | Facet to group by (e.g., `@service`, `@resource.name`, `@http.status_code`) |

## Implementation

### Approach: Typed API Client

Use `datadogV2.SpansApi` directly (same pattern as logs using `datadogV2.LogsApi`). Type-safe, compiler-checked, with built-in pagination helpers.

### Search flow

1. Parse `--from`/`--to` using `util.ParseTimeToUnixMilli`
2. Build `SpansListRequest` with `SpansQueryFilter` (query, from, to)
3. Set page limit per request (min of `--limit` and API max per page)
4. Use `ListSpansWithPagination` channel helper to auto-paginate
5. Collect spans up to `--limit`, format output via `formatter.FormatOutput`, print

### Aggregate flow

1. Parse `--from`/`--to`
2. Parse `--compute` into `SpansCompute` (aggregation type + metric field)
3. Optionally parse `--group-by` into `SpansGroupBy`
4. Build `SpansAggregateRequest`
5. Single API call (no pagination -- returns buckets)
6. Format output, print

### --compute parsing

Follows `type:field` convention (same as logs aggregate):
- `count` -> aggregation=count, no metric needed
- `avg:@duration` -> aggregation=avg, metric=@duration
- `pc90:@duration` -> aggregation=percentile (p90), metric=@duration

### Authentication

Uses `getClient()` (standard OAuth path). Spans API requires `apm_read` OAuth scope. If API key fallback is needed, endpoints will be added to `auth_validator.go`.

### Error handling

- Extract API error body via `extractAPIErrorBody`
- Include request context in errors (query, time range, status code)
- Domain-specific troubleshooting hints

### Time range flags

Uses `--from`/`--to` with flexible string parsing (consistent with logs, metrics, rum, cicd, audit-logs -- the dominant pattern at ~17 subcommands vs ~8 using `--start`/`--end`). Defaults: `--from "1h" --to "now"`.

## Files

| File | Action |
|------|--------|
| `cmd/traces_simple.go` | Delete |
| `cmd/traces_simple_test.go` | Delete |
| `cmd/traces.go` | Create |
| `cmd/traces_test.go` | Create |
| `docs/COMMANDS.md` | Update traces row and notes |

No changes to `cmd/root.go` (already registers `tracesCmd`).

## Testing

Unit tests in `cmd/traces_test.go` following established patterns:
- Command structure (Use, Short, Long, subcommand registration)
- Flag existence, types, defaults
- Flag validation (required flags, sort values, limit bounds, compute parsing)
- Search execution with mocked client (request building, pagination, error handling)
- Aggregate execution with mocked client (request building, compute/group-by, error handling)
- Coverage target: >80%

## Follow-up issues

- Unify time range flags: migrate `apm`, `usage`, `events list` from `--start`/`--end` (raw unix timestamps) to `--from`/`--to` (flexible parsing). Breaking change, separate PR.
- SpansMetrics CRUD: list, get, create, update, delete span-based metrics.
