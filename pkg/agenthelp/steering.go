// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package agenthelp

// GetQuerySyntax returns query syntax documentation for each domain.
func GetQuerySyntax() map[string]string {
	return map[string]string{
		"logs": `status:error, service:web-app, @attr:val, host:i-*, "exact phrase", AND/OR/NOT operators, -status:info (negation), wildcards with *`,
		"metrics": `<aggregation>:<metric_name>{<filter>} by {<group>}. Example: avg:system.cpu.user{env:prod} by {host}. Aggregations: avg, sum, min, max, count`,
		"monitors": `Use --name for substring search, --tags for tag filtering (comma-separated). Search via --query for full-text search`,
		"apm": `service:<name> resource_name:<path> @duration:>5000000000 (nanoseconds!) status:error operation_name:<op>. Duration is always in nanoseconds`,
		"rum": `@type:error @session.type:user @view.url_path:/checkout @action.type:click service:<app-name>`,
		"security": `@workflow.rule.type:log_detection source:cloudtrail @network.client.ip:10.0.0.0/8 status:critical`,
		"events": `sources:nagios,pagerduty status:error priority:normal tags:env:prod`,
		"traces": `service:<name> resource_name:<path> @duration:>5s (shorthand) env:production`,
	}
}

// GetTimeFormats returns documentation for supported time formats.
func GetTimeFormats() TimeFormats {
	return TimeFormats{
		Relative: []string{"5s", "30m", "1h", "4h", "1d", "7d", "1w", "30d", "5min", "2hours", "3days"},
		Absolute: []string{"Unix timestamp in milliseconds", "RFC3339 (2024-01-01T00:00:00Z)"},
		Examples: []string{
			`--from=1h (1 hour ago)`,
			`--from=30m --to=now`,
			`--from=7d --to=1d (7 days ago to 1 day ago)`,
			`--from=2024-01-01T00:00:00Z --to=2024-01-02T00:00:00Z`,
			`--from="5 minutes"`,
		},
	}
}

// GetWorkflows returns common multi-step workflows for agents.
func GetWorkflows() []Workflow {
	return []Workflow{
		{
			Name: "Investigate errors",
			Steps: []string{
				`pup logs search --query="status:error" --from=1h --limit=20`,
				`pup logs aggregate --query="status:error" --from=1h --compute="count" --group-by="service"`,
				`pup monitors list --tags="env:production" --limit=50`,
			},
		},
		{
			Name: "Performance investigation",
			Steps: []string{
				`pup metrics query --query="avg:trace.servlet.request.duration{env:prod} by {service}" --from=1h`,
				`pup logs search --query="@duration:>5000000000" --from=1h --limit=20`,
				`pup apm services list`,
			},
		},
		{
			Name: "Monitor status check",
			Steps: []string{
				`pup monitors list --tags="env:production" --limit=500`,
				`pup monitors search --query="status:Alert"`,
				`pup monitors get <monitor_id>`,
			},
		},
		{
			Name: "Security audit",
			Steps: []string{
				`pup audit-logs search --query="*" --from=1d --limit=100`,
				`pup security rules list`,
				`pup security signals list --query="status:critical" --from=1d`,
			},
		},
		{
			Name: "Service health overview",
			Steps: []string{
				`pup slos list`,
				`pup monitors list --tags="team:<team_name>"`,
				`pup incidents list --query="status:active"`,
			},
		},
	}
}

// GetBestPractices returns agent-specific best practices.
func GetBestPractices() []string {
	return []string{
		"Always specify --from to set a time range; most commands default to 1h but be explicit",
		"Start with narrow time ranges (1h) then widen if needed; large ranges are slow and expensive",
		"Filter by service first when investigating issues: --query='service:<name>'",
		"Use --limit to control result size; default varies by command (50-200)",
		"For monitors, use --tags to filter rather than listing all and parsing locally",
		"APM durations are in NANOSECONDS: 1 second = 1000000000, 5ms = 5000000",
		"Use 'pup logs aggregate' for counts and distributions instead of fetching all logs and counting locally",
		"Prefer JSON output (default) for structured parsing; use --output=table only for human display",
		"Chain narrow queries: first aggregate to find patterns, then search for specific examples",
		"Use 'pup monitors search' for full-text search, 'pup monitors list' for tag/name filtering",
	}
}

// GetAntiPatterns returns common mistakes agents should avoid.
func GetAntiPatterns() []string {
	return []string{
		"Don't omit --from on time-series queries; you'll get unexpected time ranges or errors",
		"Don't use --limit=1000 as a first step; start with small limits and refine queries",
		"Don't list all monitors/logs without filters in large organizations (>10k monitors)",
		"Don't assume APM durations are in seconds or milliseconds; they are in NANOSECONDS",
		"Don't fetch raw logs to count them; use 'pup logs aggregate --compute=count' instead",
		"Don't use --from=30d unless you specifically need a month of data; it's slow",
		"Don't retry failed requests without checking the error; 401 means re-authenticate, 403 means missing permissions",
		"Don't use 'pup metrics query' without specifying an aggregation (avg, sum, max, min, count)",
		"Don't pipe large JSON responses through multiple jq transforms; use query filters at the API level",
	}
}
