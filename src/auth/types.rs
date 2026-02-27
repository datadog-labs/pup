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
        "error_tracking_read",
        "monitors_read",
        "monitors_write",
        "monitors_downtime",
        "apm_read",
        "slos_read",
        "slos_write",
        "slos_corrections",
        "incident_read",
        "incident_write",
        "incident_notification_settings_read",
        "incident_settings_read",
        "incident_settings_write",
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
        "logs_generate_metrics",
        "logs_read_archives",
        "logs_read_config",
        "logs_read_data",
        "logs_read_index_data",
        "logs_write_archives",
        "metrics_read",
        "oci_configuration_edit",
        "oci_configuration_read",
        "oci_configurations_manage",
        "timeseries_query",
        "usage_read",
    ]
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_token(issued_ago_secs: i64, expires_in: i64) -> TokenSet {
        TokenSet {
            access_token: "test".into(),
            refresh_token: "refresh".into(),
            token_type: "Bearer".into(),
            expires_in,
            issued_at: chrono::Utc::now().timestamp() - issued_ago_secs,
            scope: String::new(),
            client_id: String::new(),
        }
    }

    #[test]
    fn test_token_not_expired() {
        let token = make_token(0, 3600); // issued now, expires in 1h
        assert!(!token.is_expired());
    }

    #[test]
    fn test_token_expired() {
        let token = make_token(7200, 3600); // issued 2h ago, expires in 1h = expired 1h ago
        assert!(token.is_expired());
    }

    #[test]
    fn test_token_expires_within_buffer() {
        let token = make_token(3400, 3600); // issued 3400s ago, expires at 3600s = 200s left < 300s buffer
        assert!(token.is_expired());
    }

    #[test]
    fn test_token_just_outside_buffer() {
        let token = make_token(3000, 3600); // issued 3000s ago, expires at 3600s = 600s left > 300s buffer
        assert!(!token.is_expired());
    }

    #[test]
    fn test_default_scopes() {
        let scopes = default_scopes();
        assert!(scopes.len() > 20);
        assert!(scopes.contains(&"dashboards_read"));
        assert!(scopes.contains(&"monitors_read"));
        assert!(scopes.contains(&"logs_read_data"));
    }

    #[test]
    fn test_token_serialization_roundtrip() {
        let token = make_token(0, 3600);
        let json = serde_json::to_string(&token).unwrap();
        let parsed: TokenSet = serde_json::from_str(&json).unwrap();
        assert_eq!(parsed.access_token, "test");
        assert_eq!(parsed.token_type, "Bearer");
    }
}
