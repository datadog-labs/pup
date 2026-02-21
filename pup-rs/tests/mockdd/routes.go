// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024-present Datadog, Inc.

package main

import (
	"embed"
	"regexp"
	"strings"
)

//go:embed fixtures/*.json
var fixtureFS embed.FS

// Route defines a mock API route with method, URL pattern, and fixture response.
type Route struct {
	Method  string
	Pattern string
	regex   *regexp.Regexp
	Fixture []byte
	Status  int // HTTP status code (0 or unset means 200)
}

// Match returns true if the given path matches this route's pattern.
func (r *Route) Match(path string) bool {
	return r.regex.MatchString(path)
}

type routeDef struct {
	method  string
	pattern string
	fixture string
	status  int // 0 means 200
}

func buildRoutes() []Route {
	routes := []routeDef{
		// ---- V1 APIs ----

		// Monitors
		{"GET", "/api/v1/monitor", "v1_monitors.json", 0},
		{"GET", "/api/v1/monitor/{id}", "v1_monitor.json", 0},
		{"POST", "/api/v1/monitor", "v1_monitor.json", 0},
		{"PUT", "/api/v1/monitor/{id}", "v1_monitor.json", 0},
		{"DELETE", "/api/v1/monitor/{id}", "v1_deleted.json", 0},
		{"GET", "/api/v1/monitor/search", "v1_monitors.json", 0},

		// Dashboards
		{"GET", "/api/v1/dashboard", "v1_dashboards.json", 0},
		{"GET", "/api/v1/dashboard/{id}", "v1_dashboard.json", 0},
		{"POST", "/api/v1/dashboard", "v1_dashboard.json", 0},
		{"PUT", "/api/v1/dashboard/{id}", "v1_dashboard.json", 0},
		{"DELETE", "/api/v1/dashboard/{id}", "v1_deleted.json", 0},

		// Metrics
		{"GET", "/api/v1/metrics", "v1_metrics.json", 0},
		{"GET", "/api/v1/query", "v1_metrics.json", 0},
		{"POST", "/api/v2/series", "v2_ok.json", 0},
		{"GET", "/api/v1/metrics/{name}", "v1_metrics.json", 0},
		{"PUT", "/api/v1/metrics/{name}", "v1_metrics.json", 0},

		// SLOs
		{"GET", "/api/v1/slo", "v1_slos.json", 0},
		{"GET", "/api/v1/slo/{id}", "v1_slo.json", 0},
		{"POST", "/api/v1/slo", "v1_slo.json", 0},
		{"PUT", "/api/v1/slo/{id}", "v1_slo.json", 0},
		{"DELETE", "/api/v1/slo/{id}", "v1_deleted.json", 0},

		// Synthetics
		{"GET", "/api/v1/synthetics/tests", "v1_synthetics_tests.json", 0},
		{"GET", "/api/v1/synthetics/tests/{id}", "v1_synthetics_tests.json", 0},
		{"GET", "/api/v1/synthetics/tests/search", "v1_synthetics_tests.json", 0},
		{"GET", "/api/v1/synthetics/locations", "v2_generic_list.json", 0},

		// Events
		{"GET", "/api/v1/events", "v1_events.json", 0},
		{"GET", "/api/v1/events/{id}", "v1_events.json", 0},

		// Downtimes
		{"GET", "/api/v2/downtime", "v2_generic_list.json", 0},
		{"GET", "/api/v2/downtime/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/downtime", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/downtime/{id}", "v2_ok.json", 0},

		// Tags
		{"GET", "/api/v1/tags/hosts", "v1_tags.json", 0},
		{"GET", "/api/v1/tags/hosts/{host}", "v1_host_tags.json", 0},
		{"POST", "/api/v1/tags/hosts/{host}", "v1_tags.json", 0},
		{"PUT", "/api/v1/tags/hosts/{host}", "v1_tags.json", 0},
		{"DELETE", "/api/v1/tags/hosts/{host}", "v1_deleted.json", 0},

		// Hosts
		{"GET", "/api/v1/hosts", "v1_hosts.json", 0},
		{"GET", "/api/v1/hosts/totals", "v1_hosts.json", 0},

		// Notebooks
		{"GET", "/api/v1/notebooks", "v1_notebooks.json", 0},
		{"GET", "/api/v1/notebooks/{id}", "v1_notebook.json", 0},
		{"POST", "/api/v1/notebooks", "v1_notebooks.json", 0},
		{"PUT", "/api/v1/notebooks/{id}", "v1_notebooks.json", 0},
		{"DELETE", "/api/v1/notebooks/{id}", "v1_deleted.json", 0},

		// Organizations
		{"GET", "/api/v1/org", "v2_generic_list.json", 0},
		{"GET", "/api/v1/org/{id}", "v2_generic_data.json", 0},

		// IP Ranges
		{"GET", "/", "v1_ip_ranges.json", 0},

		// Validate
		{"GET", "/api/v1/validate", "v1_validate.json", 0},

		// AWS/GCP/Azure
		{"GET", "/api/v1/integration/aws", "v2_generic_list.json", 0},
		{"GET", "/api/v1/integration/gcp", "v1_gcp_accounts.json", 0},
		{"GET", "/api/v1/integration/azure", "v1_azure_accounts.json", 0},

		// ---- V2 APIs ----

		// Logs (underscore + hyphen variants)
		{"POST", "/api/v2/logs/events/search", "v2_logs_list.json", 0},
		{"POST", "/api/v2/logs/analytics/aggregate", "v2_logs_aggregate.json", 0},
		{"GET", "/api/v2/logs/config/archives", "v2_generic_list.json", 0},
		{"GET", "/api/v2/logs/config/archives/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/logs/config/archives/{id}", "v2_ok.json", 0},
		{"GET", "/api/v2/logs/config/custom_destinations", "v2_generic_list.json", 0},
		{"GET", "/api/v2/logs/config/custom_destinations/{id}", "v2_generic_data.json", 0},
		{"GET", "/api/v2/logs/config/custom-destinations", "v2_generic_list.json", 0},
		{"GET", "/api/v2/logs/config/custom-destinations/{id}", "v2_generic_data.json", 0},
		{"GET", "/api/v2/logs/config/metrics", "v2_generic_list.json", 0},
		{"GET", "/api/v2/logs/config/metrics/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/logs/config/metrics/{id}", "v2_ok.json", 0},
		{"GET", "/api/v2/logs/config/restriction_queries", "v2_generic_list.json", 0},
		{"GET", "/api/v2/logs/config/restriction_queries/{id}", "v2_generic_data.json", 0},
		{"GET", "/api/v2/logs/config/restriction-queries", "v2_generic_list.json", 0},
		{"GET", "/api/v2/logs/config/restriction-queries/{id}", "v2_generic_data.json", 0},

		// Incidents (both path variants: Go uses global/settings, global/incident-handles; Rust uses config/settings, config/handles)
		{"GET", "/api/v2/incidents", "v2_incidents.json", 0},
		{"GET", "/api/v2/incidents/{id}", "v2_incident.json", 0},
		{"GET", "/api/v2/incidents/{id}/attachments", "v2_generic_list.json", 0},
		{"DELETE", "/api/v2/incidents/{id}/attachments/{aid}", "v2_ok.json", 0},
		{"GET", "/api/v2/incidents/config/settings", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/incidents/config/settings", "v2_generic_data.json", 0},
		{"GET", "/api/v2/incidents/config/global/settings", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/incidents/config/global/settings", "v2_generic_data.json", 0},
		{"GET", "/api/v2/incidents/config/handles", "v2_generic_list.json", 0},
		{"POST", "/api/v2/incidents/config/handles", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/incidents/config/handles", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/incidents/config/handles", "v2_ok.json", 0},
		{"GET", "/api/v2/incidents/config/global/incident-handles", "v2_generic_list.json", 0},
		{"POST", "/api/v2/incidents/config/global/incident-handles", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/incidents/config/global/incident-handles", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/incidents/config/global/incident-handles", "v2_ok.json", 0},
		{"GET", "/api/v2/incidents/config/postmortem_templates", "v2_generic_list.json", 0},
		{"GET", "/api/v2/incidents/config/postmortem_templates/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/incidents/config/postmortem_templates", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/incidents/config/postmortem_templates/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/incidents/config/postmortem_templates/{id}", "v2_ok.json", 0},
		{"GET", "/api/v2/incidents/config/postmortem-templates", "v2_generic_list.json", 0},
		{"GET", "/api/v2/incidents/config/postmortem-templates/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/incidents/config/postmortem-templates", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/incidents/config/postmortem-templates/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/incidents/config/postmortem-templates/{id}", "v2_ok.json", 0},

		// Users
		{"GET", "/api/v2/users", "v2_users.json", 0},
		{"GET", "/api/v2/users/{id}", "v2_generic_data.json", 0},
		{"GET", "/api/v2/roles", "v2_generic_list.json", 0},

		// Security (underscore + hyphen variants)
		{"GET", "/api/v2/security_monitoring/rules", "v2_security_rules.json", 0},
		{"GET", "/api/v2/security_monitoring/rules/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/security_monitoring/signals/search", "v2_generic_list.json", 0},
		{"POST", "/api/v2/security_monitoring/rules/bulk_export", "v2_ok.json", 0},
		{"POST", "/api/v2/security-monitoring/rules/bulk-export", "v2_ok.json", 0},
		{"GET", "/api/v2/posture_management/findings", "v2_generic_list.json", 0},
		{"POST", "/api/v2/security/findings/search", "v2_generic_list.json", 0},
		{"GET", "/api/v2/security_monitoring/content_packs", "v2_generic_list.json", 0},
		{"POST", "/api/v2/security_monitoring/content_packs/{id}/activate", "v2_ok.json", 0},
		{"POST", "/api/v2/security_monitoring/content_packs/{id}/deactivate", "v2_ok.json", 0},
		{"GET", "/api/v2/security-monitoring/content-packs", "v2_generic_list.json", 0},
		{"POST", "/api/v2/security-monitoring/content-packs/{id}/activate", "v2_ok.json", 0},
		{"POST", "/api/v2/security-monitoring/content-packs/{id}/deactivate", "v2_ok.json", 0},
		{"GET", "/api/v2/risk_scores", "v2_generic_list.json", 0},
		{"GET", "/api/v2/risk-scores/entities", "v2_generic_list.json", 0},

		// Cases
		{"POST", "/api/v2/cases/search", "v2_cases.json", 0},
		{"GET", "/api/v2/cases", "v2_cases.json", 0},
		{"POST", "/api/v2/cases", "v2_generic_data.json", 0},
		// Projects must be BEFORE /cases/{id} to prevent "projects" matching as {id}
		{"GET", "/api/v2/cases/projects", "v2_generic_list.json", 0},
		{"GET", "/api/v2/cases/projects/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/cases/projects", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/cases/projects/{id}", "v2_ok.json", 0},
		{"GET", "/api/v2/cases/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/cases/{id}/archive", "v2_generic_data.json", 0},
		{"POST", "/api/v2/cases/{id}/unarchive", "v2_generic_data.json", 0},
		{"POST", "/api/v2/cases/{id}/assign", "v2_generic_data.json", 0},
		{"POST", "/api/v2/cases/{id}/priority", "v2_generic_data.json", 0},
		{"POST", "/api/v2/cases/{id}/status", "v2_generic_data.json", 0},
		{"POST", "/api/v2/cases/{id}/jira/issue", "v2_generic_data.json", 0},
		{"POST", "/api/v2/cases/{id}/jira/link", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/cases/{id}/jira/unlink", "v2_ok.json", 0},
		{"POST", "/api/v2/cases/{id}/servicenow/ticket", "v2_generic_data.json", 0},
		{"GET", "/api/v2/cases/projects/{id}/notification-rules", "v2_generic_list.json", 0},
		{"POST", "/api/v2/cases/projects/{id}/notification-rules", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/cases/projects/{id}/notification-rules/{aid}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/cases/projects/{id}/notification-rules/{aid}", "v2_ok.json", 0},

		// RUM (underscore + hyphen + Rust config paths)
		{"GET", "/api/v2/rum/applications", "v2_rum_apps.json", 0},
		{"GET", "/api/v2/rum/applications/{id}", "v2_rum_app.json", 0},
		{"POST", "/api/v2/rum/applications", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/rum/applications/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/rum/applications/{id}", "v2_ok.json", 0},
		{"POST", "/api/v2/rum/events/search", "v2_generic_list.json", 0},
		{"GET", "/api/v2/rum/events", "v2_generic_list.json", 0},
		// RUM metrics (Go path: /rum/metrics, Rust path: /rum/config/metrics)
		{"GET", "/api/v2/rum/metrics", "v2_generic_list.json", 0},
		{"GET", "/api/v2/rum/metrics/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/rum/metrics", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/rum/metrics/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/rum/metrics/{id}", "v2_ok.json", 0},
		{"GET", "/api/v2/rum/config/metrics", "v2_generic_list.json", 0},
		{"GET", "/api/v2/rum/config/metrics/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/rum/config/metrics", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/rum/config/metrics/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/rum/config/metrics/{id}", "v2_ok.json", 0},
		// RUM retention filters (Go: /rum/retention_filters, Rust: /rum/applications/{id}/retention_filters)
		{"GET", "/api/v2/rum/retention_filters", "v2_generic_list.json", 0},
		{"GET", "/api/v2/rum/retention_filters/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/rum/retention_filters", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/rum/retention_filters/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/rum/retention_filters/{id}", "v2_ok.json", 0},
		{"GET", "/api/v2/rum/applications/{id}/retention_filters", "v2_generic_list.json", 0},
		{"GET", "/api/v2/rum/applications/{id}/retention_filters/{aid}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/rum/applications/{id}/retention_filters", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/rum/applications/{id}/retention_filters/{aid}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/rum/applications/{id}/retention_filters/{aid}", "v2_ok.json", 0},
		// RUM playlists + heatmaps
		{"GET", "/api/v2/rum/replay/playlists", "v2_generic_list.json", 0},
		{"GET", "/api/v2/rum/replay/playlists/{id}", "v2_generic_data.json", 0},
		{"GET", "/api/v2/replay/heatmap/snapshots", "v2_generic_list.json", 0},

		// CI/CD
		{"POST", "/api/v2/ci/pipelines/events/search", "v2_pipelines.json", 0},
		{"POST", "/api/v2/ci/tests/events/search", "v2_generic_list.json", 0},
		{"GET", "/api/v2/ci/tests/events", "v2_generic_list.json", 0},
		{"POST", "/api/v2/ci/tests/aggregate", "v2_generic_data.json", 0},
		{"POST", "/api/v2/ci/pipelines/aggregate", "v2_generic_data.json", 0},

		// API Keys
		{"GET", "/api/v2/api_keys", "v2_generic_list.json", 0},
		{"GET", "/api/v2/api_keys/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/api_keys", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/api_keys/{id}", "v2_ok.json", 0},

		// App Keys
		{"GET", "/api/v2/application_keys", "v2_generic_list.json", 0},
		{"GET", "/api/v2/application_keys/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/application_keys", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/application_keys/{id}", "v2_ok.json", 0},

		// Teams / On-Call (Go uses /teams, Rust uses /team for single-get)
		{"GET", "/api/v2/teams", "v2_teams.json", 0},
		{"GET", "/api/v2/teams/{id}", "v2_team.json", 0},
		{"GET", "/api/v2/team/{id}", "v2_team.json", 0},
		{"POST", "/api/v2/teams", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/teams/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/teams/{id}", "v2_ok.json", 0},

		// Fleet (Rust uses /api/unstable/, Go uses /api/v2/)
		{"GET", "/api/v2/fleet/agents", "v2_fleet_agents.json", 0},
		{"GET", "/api/unstable/fleet/agents", "v2_fleet_agents.json", 0},
		{"GET", "/api/v2/fleet/agents/{id}", "v2_generic_data.json", 0},
		{"GET", "/api/unstable/fleet/agents/{id}", "v2_generic_data.json", 0},
		{"GET", "/api/v2/fleet/agents/versions", "v2_generic_list.json", 0},
		{"GET", "/api/unstable/fleet/agents/versions", "v2_generic_list.json", 0},
		{"GET", "/api/v2/fleet/deployments", "v2_generic_list.json", 0},
		{"GET", "/api/unstable/fleet/deployments", "v2_generic_list.json", 0},
		{"GET", "/api/v2/fleet/deployments/{id}", "v2_generic_data.json", 0},
		{"GET", "/api/unstable/fleet/deployments/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/fleet/deployments/configure", "v2_generic_data.json", 0},
		{"POST", "/api/v2/fleet/deployments/upgrade", "v2_generic_data.json", 0},
		{"GET", "/api/v2/fleet/schedules", "v2_generic_list.json", 0},
		{"GET", "/api/unstable/fleet/schedules", "v2_generic_list.json", 0},
		{"GET", "/api/v2/fleet/schedules/{id}", "v2_generic_data.json", 0},
		{"GET", "/api/unstable/fleet/schedules/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/fleet/schedules", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/fleet/schedules/{id}", "v2_generic_data.json", 0},
		{"PATCH", "/api/unstable/fleet/schedules/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/fleet/schedules/{id}/trigger", "v2_ok.json", 0},
		{"DELETE", "/api/v2/fleet/schedules/{id}", "v2_ok.json", 0},
		{"DELETE", "/api/unstable/fleet/schedules/{id}", "v2_ok.json", 0},

		// Audit Logs
		{"POST", "/api/v2/audit/events/search", "v2_generic_list.json", 0},

		// Events
		{"POST", "/api/v2/events/search", "v2_generic_list.json", 0},

		// Error Tracking
		{"GET", "/api/v2/error-tracking/issues", "v2_generic_list.json", 0},
		{"GET", "/api/v2/error-tracking/issues/{id}", "v2_generic_data.json", 0},

		// Usage
		{"GET", "/api/v2/usage/hourly_usage", "v2_generic_list.json", 0},
		{"GET", "/api/v1/usage/summary", "v2_generic_list.json", 0},

		// Cost
		{"GET", "/api/v2/usage/projected_cost", "v2_generic_list.json", 0},
		{"GET", "/api/v2/usage/cost_by_org", "v2_generic_list.json", 0},
		{"GET", "/api/v2/cost_by_org", "v2_generic_list.json", 0},

		// Service Catalog
		{"GET", "/api/v2/services/definitions", "v2_generic_list.json", 0},
		{"GET", "/api/v2/services/definitions/{name}", "v2_generic_data.json", 0},

		// Integrations
		{"GET", "/api/v2/integration/jira/accounts", "v2_generic_list.json", 0},
		{"DELETE", "/api/v2/integration/jira/accounts/{id}", "v2_ok.json", 0},
		{"GET", "/api/v2/integration/jira/templates", "v2_generic_list.json", 0},
		{"GET", "/api/v2/integration/jira/templates/{id}", "v2_jira_template.json", 0},
		{"GET", "/api/v2/integration/jira/issue-templates/{id}", "v2_jira_template.json", 0},
		{"POST", "/api/v2/integration/jira/templates", "v2_generic_data.json", 0},
		{"POST", "/api/v2/integration/jira/issue-templates", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/integration/jira/templates/{id}", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/integration/jira/issue-templates/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/integration/jira/templates/{id}", "v2_ok.json", 0},
		{"DELETE", "/api/v2/integration/jira/issue-templates/{id}", "v2_ok.json", 0},
		{"GET", "/api/v2/integration/servicenow/instances", "v2_generic_list.json", 0},
		{"GET", "/api/v2/integration/servicenow/templates", "v2_generic_list.json", 0},
		{"GET", "/api/v2/integration/servicenow/handles", "v2_generic_list.json", 0},
		{"GET", "/api/v2/integration/servicenow/templates/{id}", "v2_servicenow_template.json", 0},
		{"GET", "/api/v2/integration/servicenow/handles/{id}", "v2_servicenow_template.json", 0},
		{"POST", "/api/v2/integration/servicenow/templates", "v2_generic_data.json", 0},
		{"POST", "/api/v2/integration/servicenow/handles", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/integration/servicenow/templates/{id}", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/integration/servicenow/handles/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/integration/servicenow/templates/{id}", "v2_ok.json", 0},
		{"DELETE", "/api/v2/integration/servicenow/handles/{id}", "v2_ok.json", 0},

		// HAMR
		{"GET", "/api/v2/hamr/connections/org", "v2_hamr.json", 0},
		{"POST", "/api/v2/hamr/connections/org", "v2_hamr.json", 0},

		// Data Governance (Go uses /config/rules, Rust uses /config)
		{"GET", "/api/v2/sensitive-data-scanner/config/rules", "v2_generic_list.json", 0},
		{"GET", "/api/v2/sensitive-data-scanner/config", "v2_scanner_config.json", 0},

		// Investigations
		{"GET", "/api/v2/investigations", "v2_generic_list.json", 0},
		{"GET", "/api/v2/investigations/{id}", "v2_generic_data.json", 0},
		{"GET", "/api/v2/bits-ai/investigations", "v2_generic_list.json", 0},
		{"GET", "/api/v2/bits-ai/investigations/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/bits-ai/investigations", "v2_generic_data.json", 0},

		// HAMR (alt path)
		{"GET", "/api/v2/hamr", "v2_hamr.json", 0},

		// App Key Registrations (Go-specific ActionConnectionAPI)
		{"GET", "/api/v2/actions/app_key_registrations", "v2_generic_list.json", 0},
		{"GET", "/api/v2/actions/app_key_registrations/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/actions/app_key_registrations/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/actions/app_key_registrations/{id}", "v2_ok.json", 0},

		// Synthetics suites (V2)
		{"POST", "/api/v2/synthetics/suites/search", "v2_generic_list.json", 0},
		{"GET", "/api/v2/synthetics/suites/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/synthetics/suites", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/synthetics/suites/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/synthetics/suites/bulk-delete", "v2_ok.json", 0},

		// DORA
		{"PATCH", "/api/v2/dora/deployments/{id}", "v2_generic_data.json", 0},

		// Flaky Tests
		{"POST", "/api/v2/ci/tests/flaky", "v2_generic_list.json", 0},
		{"PATCH", "/api/v2/ci/tests/flaky", "v2_ok.json", 0},
		{"GET", "/api/v2/ci/tests/flaky", "v2_generic_list.json", 0},

		// OCI Integration
		{"GET", "/api/v2/integration/oci/tenancies", "v2_generic_list.json", 0},
		{"GET", "/api/v2/integration/oci/tenancies/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/integration/oci/tenancies", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/integration/oci/tenancies/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/integration/oci/tenancies/{id}", "v2_ok.json", 0},
		{"GET", "/api/v2/integration/oci/products", "v2_generic_list.json", 0},

		// Slack / PagerDuty / Webhooks (V1)
		{"GET", "/api/v1/integration/slack/configuration/accounts", "v2_generic_list.json", 0},
		{"GET", "/api/v1/integration/pagerduty/configuration/services", "v2_generic_list.json", 0},
		{"GET", "/api/v1/integration/webhooks/configuration/custom-variables", "v2_generic_list.json", 0},

		// Status Pages
		{"GET", "/api/v2/status-pages", "v2_generic_list.json", 0},
		{"GET", "/api/v2/status-pages/{id}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/status-pages", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/status-pages/{id}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/status-pages/{id}", "v2_ok.json", 0},
		{"GET", "/api/v2/status-pages/{id}/components", "v2_generic_list.json", 0},
		{"GET", "/api/v2/status-pages/{id}/components/{aid}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/status-pages/{id}/components", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/status-pages/{id}/components/{aid}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/status-pages/{id}/components/{aid}", "v2_ok.json", 0},
		{"GET", "/api/v2/status-pages/degradations", "v2_generic_list.json", 0},
		{"GET", "/api/v2/status-pages/{id}/degradations/{aid}", "v2_generic_data.json", 0},
		{"POST", "/api/v2/status-pages/{id}/degradations", "v2_generic_data.json", 0},
		{"PATCH", "/api/v2/status-pages/{id}/degradations/{aid}", "v2_generic_data.json", 0},
		{"DELETE", "/api/v2/status-pages/{id}/degradations/{aid}", "v2_ok.json", 0},

		// Metrics tags (V2)
		{"GET", "/api/v2/metrics/{name}/tags", "v2_generic_list.json", 0},
		{"GET", "/api/v2/metrics/{name}/all-tags", "v2_generic_list.json", 0},

		// SLO status (V2)
		{"GET", "/api/v2/slo/{id}/status", "v2_generic_data.json", 0},

		// Product Analytics
		{"POST", "/api/v2/product-analytics/events", "v2_ok.json", 0},
		{"POST", "/api/v2/product_analytics/events", "v2_ok.json", 0},

		// Usage (hyphen variants)
		{"GET", "/api/v1/usage/hourly-attribution", "v2_generic_list.json", 0},
		{"GET", "/api/v2/usage/cost-by-org", "v2_generic_list.json", 0},

		// Audit logs (GET variant)
		{"GET", "/api/v2/audit/events", "v2_generic_list.json", 0},

		// APM (raw endpoints)
		{"GET", "/api/v2/apm/services", "v2_generic_list.json", 0},
		{"GET", "/api/v2/apm/services/stats", "v2_generic_data.json", 0},
		{"GET", "/api/unstable/apm/entities", "v2_generic_list.json", 0},
		{"GET", "/api/v1/service_dependencies", "v2_generic_list.json", 0},
		{"GET", "/api/v1/trace/operation_names/{id}", "v2_generic_list.json", 0},
		{"GET", "/api/ui/apm/resources", "v2_generic_list.json", 0},
		{"GET", "/api/ui/apm/flow-map", "v2_generic_data.json", 0},

		// ServiceNow users/groups/services
		{"GET", "/api/v2/integration/servicenow/instances/{id}/users", "v2_generic_list.json", 0},
		{"GET", "/api/v2/integration/servicenow/instances/{id}/assignment-groups", "v2_generic_list.json", 0},
		{"GET", "/api/v2/integration/servicenow/instances/{id}/business-services", "v2_generic_list.json", 0},

		// Code Coverage
		{"GET", "/api/v2/code-coverage/branch", "v2_generic_data.json", 0},
		{"GET", "/api/v2/code-coverage/commit", "v2_generic_data.json", 0},

		// ---- Error response routes (use IDs starting with "err-" to trigger) ----
		// 404 responses
		{"GET", "/api/v1/monitor/999999999", "error_404.json", 404},
		{"GET", "/api/v1/dashboard/err-not-found", "error_404.json", 404},
		{"GET", "/api/v2/cases/err-not-found", "error_404.json", 404},

		// 403 responses
		{"GET", "/api/v1/monitor/err-forbidden", "error_403.json", 403},

		// 400 responses
		{"GET", "/api/v2/cases/err-bad-request", "error_400.json", 400},
	}

	// Build compiled routes
	var compiled []Route
	for _, rd := range routes {
		fixture, _ := fixtureFS.ReadFile("fixtures/" + rd.fixture)
		if fixture == nil {
			fixture = []byte(`{"data":[]}`)
		}
		// Convert pattern to regex: escape special chars, then replace {placeholders} with [^/]+
		regexStr := "^" + regexp.QuoteMeta(rd.pattern) + "$"
		regexStr = strings.ReplaceAll(regexStr, `\{id\}`, `[^/]+`)
		regexStr = strings.ReplaceAll(regexStr, `\{name\}`, `[^/]+`)
		regexStr = strings.ReplaceAll(regexStr, `\{host\}`, `[^/]+`)
		regexStr = strings.ReplaceAll(regexStr, `\{aid\}`, `[^/]+`)

		compiled = append(compiled, Route{
			Method:  rd.method,
			Pattern: rd.pattern,
			regex:   regexp.MustCompile(regexStr),
			Fixture: fixture,
			Status:  rd.status,
		})
	}
	return compiled
}
