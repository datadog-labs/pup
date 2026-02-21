use chrono::Utc;
use serde::{Deserialize, Serialize};

/// OAuth2 token set (JSON cross-compatible with Go/TypeScript versions).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TokenSet {
    pub access_token: String,
    pub refresh_token: String,
    #[serde(default = "default_token_type")]
    pub token_type: String,
    pub expires_in: i64,
    pub issued_at: i64,
    #[serde(default)]
    pub scope: String,
    #[serde(default)]
    pub client_id: String,
}

fn default_token_type() -> String {
    "Bearer".to_string()
}

impl TokenSet {
    /// Returns true if the token is expired or will expire within 5 minutes.
    pub fn is_expired(&self) -> bool {
        let now = Utc::now().timestamp();
        let expires_at = self.issued_at + self.expires_in;
        now >= (expires_at - 300) // 5-minute safety buffer
    }
}

/// DCR client credentials (cross-compatible with Go/TypeScript versions).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClientCredentials {
    pub client_id: String,
    pub client_name: String,
    pub redirect_uris: Vec<String>,
    pub registered_at: i64,
    pub site: String,
}

/// Default OAuth scopes requested during login.
pub fn default_scopes() -> Vec<&'static str> {
    vec![
        "dashboards_read",
        "dashboards_write",
        "monitors_read",
        "monitors_write",
        "monitors_downtime",
        "apm_read",
        "slos_read",
        "slos_write",
        "slos_corrections",
        "incident_read",
        "incident_write",
        "synthetics_read",
        "synthetics_write",
        "synthetics_global_variable_read",
        "synthetics_global_variable_write",
        "synthetics_private_location_read",
        "synthetics_private_location_write",
        "security_monitoring_signals_read",
        "security_monitoring_rules_read",
        "security_monitoring_findings_read",
        "security_monitoring_suppressions_read",
        "security_monitoring_filters_read",
        "rum_apps_read",
        "rum_apps_write",
        "rum_retention_filters_read",
        "rum_retention_filters_write",
        "hosts_read",
        "user_access_read",
        "user_self_profile_read",
        "cases_read",
        "cases_write",
        "events_read",
        "logs_read_data",
        "logs_read_index_data",
        "metrics_read",
        "timeseries_query",
        "usage_read",
    ]
}
