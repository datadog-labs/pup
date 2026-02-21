use async_trait::async_trait;
use reqwest_middleware::{ClientBuilder, ClientWithMiddleware, Middleware, Next};
use task_local_extensions::Extensions;

use crate::config::Config;

// ---------------------------------------------------------------------------
// Bearer token middleware
// ---------------------------------------------------------------------------

struct BearerAuthMiddleware {
    token: String,
}

#[async_trait]
impl Middleware for BearerAuthMiddleware {
    async fn handle(
        &self,
        mut req: reqwest::Request,
        extensions: &mut Extensions,
        next: Next<'_>,
    ) -> reqwest_middleware::Result<reqwest::Response> {
        req.headers_mut().insert(
            reqwest::header::AUTHORIZATION,
            format!("Bearer {}", self.token).parse().unwrap(),
        );
        next.run(req, extensions).await
    }
}

// ---------------------------------------------------------------------------
// DD Configuration builder
// ---------------------------------------------------------------------------

/// Creates a DD API Configuration with all unstable ops enabled.
/// `Configuration::new()` reads DD_API_KEY, DD_APP_KEY, DD_SITE from env.
pub fn make_dd_config(_cfg: &Config) -> datadog_api_client::datadog::Configuration {
    let mut dd_cfg = datadog_api_client::datadog::Configuration::new();

    // Enable all 63 unstable operations (snake_case in Rust client)
    for op in UNSTABLE_OPS {
        dd_cfg.set_unstable_operation_enabled(op, true);
    }

    dd_cfg
}

/// Creates a reqwest middleware client with bearer token injection.
/// Returns None if no bearer token is configured.
pub fn make_bearer_client(cfg: &Config) -> Option<ClientWithMiddleware> {
    let token = cfg.access_token.as_ref()?;
    let reqwest_client = reqwest::Client::builder()
        .build()
        .expect("failed to build reqwest client");
    let client = ClientBuilder::new(reqwest_client)
        .with(BearerAuthMiddleware {
            token: token.clone(),
        })
        .build();
    Some(client)
}

// ---------------------------------------------------------------------------
// Auth type detection
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum AuthType {
    None,
    OAuth,
    ApiKeys,
}

impl std::fmt::Display for AuthType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            AuthType::None => write!(f, "None"),
            AuthType::OAuth => write!(f, "OAuth2 Bearer Token"),
            AuthType::ApiKeys => write!(f, "API Keys (DD_API_KEY + DD_APP_KEY)"),
        }
    }
}

pub fn get_auth_type(cfg: &Config) -> AuthType {
    if cfg.has_bearer_token() {
        AuthType::OAuth
    } else if cfg.has_api_keys() {
        AuthType::ApiKeys
    } else {
        AuthType::None
    }
}

// ---------------------------------------------------------------------------
// OAuth-excluded endpoint validation
// ---------------------------------------------------------------------------

struct EndpointRequirement {
    path: &'static str,
    method: &'static str,
}

/// Returns true if the endpoint doesn't support OAuth and requires API key fallback.
pub fn requires_api_key_fallback(method: &str, path: &str) -> bool {
    find_endpoint_requirement(method, path).is_some()
}

fn find_endpoint_requirement(method: &str, path: &str) -> Option<&'static EndpointRequirement> {
    OAUTH_EXCLUDED_ENDPOINTS.iter().find(|req| {
        if req.method != method {
            return false;
        }
        // Trailing "/" means prefix match (for ID-parameterized paths)
        if req.path.ends_with('/') {
            path.starts_with(&req.path[..req.path.len() - 1])
        } else {
            req.path == path
        }
    })
}

// ---------------------------------------------------------------------------
// Static tables
// ---------------------------------------------------------------------------

/// All 63 unstable operations (snake_case for the Rust DD client).
static UNSTABLE_OPS: &[&str] = &[
    // Incidents (16)
    "v2.list_incidents",
    "v2.get_incident",
    "v2.create_incident",
    "v2.update_incident",
    "v2.delete_incident",
    "v2.create_global_incident_handle",
    "v2.delete_global_incident_handle",
    "v2.get_global_incident_settings",
    "v2.list_global_incident_handles",
    "v2.update_global_incident_handle",
    "v2.update_global_incident_settings",
    "v2.create_incident_postmortem_template",
    "v2.delete_incident_postmortem_template",
    "v2.get_incident_postmortem_template",
    "v2.list_incident_postmortem_templates",
    "v2.update_incident_postmortem_template",
    // Fleet Automation (14)
    "v2.list_fleet_agents",
    "v2.get_fleet_agent_info",
    "v2.list_fleet_agent_versions",
    "v2.list_fleet_deployments",
    "v2.get_fleet_deployment",
    "v2.create_fleet_deployment_configure",
    "v2.create_fleet_deployment_upgrade",
    "v2.cancel_fleet_deployment",
    "v2.list_fleet_schedules",
    "v2.get_fleet_schedule",
    "v2.create_fleet_schedule",
    "v2.update_fleet_schedule",
    "v2.delete_fleet_schedule",
    "v2.trigger_fleet_schedule",
    // ServiceNow (9)
    "v2.create_service_now_template",
    "v2.delete_service_now_template",
    "v2.get_service_now_template",
    "v2.list_service_now_assignment_groups",
    "v2.list_service_now_business_services",
    "v2.list_service_now_instances",
    "v2.list_service_now_templates",
    "v2.list_service_now_users",
    "v2.update_service_now_template",
    // Jira (7)
    "v2.create_jira_issue_template",
    "v2.delete_jira_account",
    "v2.delete_jira_issue_template",
    "v2.get_jira_issue_template",
    "v2.list_jira_accounts",
    "v2.list_jira_issue_templates",
    "v2.update_jira_issue_template",
    // Cases (5)
    "v2.create_case_jira_issue",
    "v2.link_jira_issue_to_case",
    "v2.unlink_jira_issue",
    "v2.create_case_service_now_ticket",
    "v2.move_case_to_project",
    // Content Packs (3)
    "v2.activate_content_pack",
    "v2.deactivate_content_pack",
    "v2.get_content_packs_states",
    // Code Coverage (2)
    "v2.get_code_coverage_branch_summary",
    "v2.get_code_coverage_commit_summary",
    // OCI Integration (2)
    "v2.create_tenancy_config",
    "v2.get_tenancy_configs",
    // HAMR (2)
    "v2.create_hamr_org_connection",
    "v2.get_hamr_org_connection",
    // Entity Risk Scores (1)
    "v2.list_entity_risk_scores",
    // SLO Status (1)
    "v2.get_slo_status",
    // Flaky Tests (1)
    "v2.update_flaky_tests",
];

/// Endpoints that don't support OAuth (52 patterns across 7 API groups).
/// Trailing "/" means prefix match for ID-parameterized paths.
static OAUTH_EXCLUDED_ENDPOINTS: &[EndpointRequirement] = &[
    // Logs API (11)
    EndpointRequirement { path: "/api/v2/logs/events", method: "POST" },
    EndpointRequirement { path: "/api/v2/logs/events/search", method: "POST" },
    EndpointRequirement { path: "/api/v2/logs/analytics/aggregate", method: "POST" },
    EndpointRequirement { path: "/api/v2/logs/config/archives", method: "GET" },
    EndpointRequirement { path: "/api/v2/logs/config/archives/", method: "GET" },
    EndpointRequirement { path: "/api/v2/logs/config/archives/", method: "DELETE" },
    EndpointRequirement { path: "/api/v2/logs/config/custom_destinations", method: "GET" },
    EndpointRequirement { path: "/api/v2/logs/config/custom_destinations/", method: "GET" },
    EndpointRequirement { path: "/api/v2/logs/config/metrics", method: "GET" },
    EndpointRequirement { path: "/api/v2/logs/config/metrics/", method: "GET" },
    EndpointRequirement { path: "/api/v2/logs/config/metrics/", method: "DELETE" },
    // RUM API (10)
    EndpointRequirement { path: "/api/v2/rum/applications", method: "GET" },
    EndpointRequirement { path: "/api/v2/rum/applications/", method: "GET" },
    EndpointRequirement { path: "/api/v2/rum/applications", method: "POST" },
    EndpointRequirement { path: "/api/v2/rum/applications/", method: "PATCH" },
    EndpointRequirement { path: "/api/v2/rum/applications/", method: "DELETE" },
    EndpointRequirement { path: "/api/v2/rum/metrics", method: "GET" },
    EndpointRequirement { path: "/api/v2/rum/metrics/", method: "GET" },
    EndpointRequirement { path: "/api/v2/rum/retention_filters", method: "GET" },
    EndpointRequirement { path: "/api/v2/rum/retention_filters/", method: "GET" },
    EndpointRequirement { path: "/api/v2/rum/events/search", method: "POST" },
    // API/App Keys (8)
    EndpointRequirement { path: "/api/v2/api_keys", method: "GET" },
    EndpointRequirement { path: "/api/v2/api_keys/", method: "GET" },
    EndpointRequirement { path: "/api/v2/api_keys", method: "POST" },
    EndpointRequirement { path: "/api/v2/api_keys/", method: "DELETE" },
    EndpointRequirement { path: "/api/v2/app_keys", method: "GET" },
    EndpointRequirement { path: "/api/v2/app_keys/", method: "GET" },
    EndpointRequirement { path: "/api/v2/app_keys/", method: "POST" },
    EndpointRequirement { path: "/api/v2/app_keys/", method: "DELETE" },
    // Events (1)
    EndpointRequirement { path: "/api/v2/events/search", method: "POST" },
    // Error Tracking (2)
    EndpointRequirement { path: "/api/v2/error_tracking/issues/search", method: "POST" },
    EndpointRequirement { path: "/api/v2/error_tracking/issues/", method: "GET" },
    // Fleet Automation (15)
    EndpointRequirement { path: "/api/v2/fleet/agents", method: "GET" },
    EndpointRequirement { path: "/api/v2/fleet/agents/", method: "GET" },
    EndpointRequirement { path: "/api/v2/fleet/agents/versions", method: "GET" },
    EndpointRequirement { path: "/api/v2/fleet/deployments", method: "GET" },
    EndpointRequirement { path: "/api/v2/fleet/deployments/", method: "GET" },
    EndpointRequirement { path: "/api/v2/fleet/deployments/configure", method: "POST" },
    EndpointRequirement { path: "/api/v2/fleet/deployments/upgrade", method: "POST" },
    EndpointRequirement { path: "/api/v2/fleet/deployments/", method: "POST" },
    EndpointRequirement { path: "/api/v2/fleet/deployments/", method: "DELETE" },
    EndpointRequirement { path: "/api/v2/fleet/schedules", method: "GET" },
    EndpointRequirement { path: "/api/v2/fleet/schedules/", method: "GET" },
    EndpointRequirement { path: "/api/v2/fleet/schedules", method: "POST" },
    EndpointRequirement { path: "/api/v2/fleet/schedules/", method: "PATCH" },
    EndpointRequirement { path: "/api/v2/fleet/schedules/", method: "DELETE" },
    EndpointRequirement { path: "/api/v2/fleet/schedules/", method: "POST" },
    // Notebooks (5)
    EndpointRequirement { path: "/api/v1/notebooks", method: "GET" },
    EndpointRequirement { path: "/api/v1/notebooks", method: "POST" },
    EndpointRequirement { path: "/api/v1/notebooks/", method: "GET" },
    EndpointRequirement { path: "/api/v1/notebooks/", method: "PUT" },
    EndpointRequirement { path: "/api/v1/notebooks/", method: "DELETE" },
];
