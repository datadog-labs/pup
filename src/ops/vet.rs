use anyhow::Result;
use datadog_api_client::datadogV1::api_monitors::{ListMonitorsOptionalParams, MonitorsAPI};
use datadog_api_client::datadogV1::model::{Monitor, MonitorOverallStates};
use serde::Serialize;
use std::time::{SystemTime, UNIX_EPOCH};

use crate::client;
use crate::config::Config;

// ---- Output types ----

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "lowercase")]
pub enum Severity {
    Critical,
    Warning,
    Info,
}

impl Severity {
    pub fn label(&self) -> &'static str {
        match self {
            Severity::Critical => "CRITICAL",
            Severity::Warning => "WARNING",
            Severity::Info => "INFO",
        }
    }
}

#[derive(Debug, Serialize)]
pub struct Resource {
    pub id: i64,
    pub name: String,
    pub detail: String,
}

#[derive(Debug, Serialize)]
pub struct Finding {
    pub check: &'static str,
    pub severity: Severity,
    pub count: usize,
    pub resources: Vec<Resource>,
    pub recommendation: &'static str,
}

#[derive(Debug, Serialize)]
pub struct VetResult {
    pub findings: Vec<Finding>,
    pub passed: Vec<&'static str>,
    pub critical: usize,
    pub warnings: usize,
    pub infos: usize,
}

// ---- Check names ----

const CHECK_SILENT: &str = "silent-monitors";
const CHECK_STALE: &str = "stale-monitors";
const CHECK_MUTED: &str = "muted-forgotten";

const ALL_CHECKS: &[&str] = &[CHECK_SILENT, CHECK_STALE, CHECK_MUTED];

// ---- Check implementations ----

/// Monitors with no @-mention or notification channel in their message.
fn check_silent_monitors(monitors: &[Monitor]) -> Finding {
    let resources: Vec<Resource> = monitors
        .iter()
        .filter(|m| {
            let msg = m.message.as_deref().unwrap_or("");
            !msg.contains('@')
        })
        .map(|m| Resource {
            id: m.id.unwrap_or(0),
            name: m.name.as_deref().unwrap_or("(unnamed)").to_string(),
            detail: "no notification channel (@-mention) in message".to_string(),
        })
        .collect();

    Finding {
        check: CHECK_SILENT,
        severity: Severity::Critical,
        count: resources.len(),
        resources,
        recommendation: "Add @mention or notification channel so alerts reach on-call responders",
    }
}

/// Monitors currently in "No Data" state.
fn check_stale_monitors(monitors: &[Monitor]) -> Finding {
    let resources: Vec<Resource> = monitors
        .iter()
        .filter(|m| matches!(m.overall_state, Some(MonitorOverallStates::NO_DATA)))
        .map(|m| Resource {
            id: m.id.unwrap_or(0),
            name: m.name.as_deref().unwrap_or("(unnamed)").to_string(),
            detail: "monitor is in No Data state".to_string(),
        })
        .collect();

    Finding {
        check: CHECK_STALE,
        severity: Severity::Warning,
        count: resources.len(),
        resources,
        recommendation:
            "Investigate missing data source or delete if the monitor is no longer needed",
    }
}

/// Monitors muted indefinitely or with a silence expiry >30 days out.
#[allow(deprecated)] // MonitorOptions::silenced is deprecated in the DD API but still functional
fn check_muted_forgotten(monitors: &[Monitor]) -> Finding {
    let now_secs = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0);
    let thirty_days_secs: i64 = 30 * 24 * 60 * 60;

    let resources: Vec<Resource> = monitors
        .iter()
        .filter_map(|m| {
            let silenced = m.options.as_ref()?.silenced.as_ref()?;
            if silenced.is_empty() {
                return None;
            }

            let has_indefinite = silenced.values().any(|v| v.is_none());
            let max_until = silenced.values().filter_map(|v| *v).max().unwrap_or(0);

            let detail = if has_indefinite {
                "muted indefinitely (no expiry set)".to_string()
            } else if max_until > now_secs + thirty_days_secs {
                let days_remaining = (max_until - now_secs) / 86400;
                format!("muted for {days_remaining} more days")
            } else {
                return None;
            };

            Some(Resource {
                id: m.id.unwrap_or(0),
                name: m.name.as_deref().unwrap_or("(unnamed)").to_string(),
                detail,
            })
        })
        .collect();

    Finding {
        check: CHECK_MUTED,
        severity: Severity::Warning,
        count: resources.len(),
        resources,
        recommendation: "Review and unmute, or delete if the monitor is no longer needed",
    }
}

// ---- Entry point ----

pub async fn run(
    cfg: &Config,
    tags: Option<String>,
    check: Option<String>,
    severity_filter: Option<String>,
) -> Result<VetResult> {
    let checks_to_run: Vec<&str> = match &check {
        Some(c) => {
            if !ALL_CHECKS.contains(&c.as_str()) {
                anyhow::bail!(
                    "unknown check '{}'. Available: {}",
                    c,
                    ALL_CHECKS.join(", ")
                );
            }
            vec![c.as_str()]
        }
        None => ALL_CHECKS.to_vec(),
    };

    // Single API call for all checks
    let dd_cfg = client::make_dd_config(cfg);
    let api = if let Some(http_client) = client::make_bearer_client(cfg) {
        MonitorsAPI::with_client_and_config(dd_cfg, http_client)
    } else {
        MonitorsAPI::with_config(dd_cfg)
    };

    let mut params = ListMonitorsOptionalParams::default().page_size(1000).page(0);
    if let Some(t) = tags {
        params = params.monitor_tags(t);
    }

    let monitors = api
        .list_monitors(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list monitors: {:?}", e))?;

    let min_severity: Option<Severity> = severity_filter.as_deref().map(|s| match s {
        "critical" => Severity::Critical,
        "warning" => Severity::Warning,
        _ => Severity::Info,
    });

    let mut findings: Vec<Finding> = Vec::new();
    let mut passed: Vec<&'static str> = Vec::new();

    for &name in &checks_to_run {
        let finding = match name {
            CHECK_SILENT => check_silent_monitors(&monitors),
            CHECK_STALE => check_stale_monitors(&monitors),
            CHECK_MUTED => check_muted_forgotten(&monitors),
            _ => unreachable!(),
        };

        if finding.count == 0 {
            passed.push(finding.check);
        } else if let Some(min) = min_severity {
            let include = matches!(
                (min, finding.severity),
                (Severity::Critical, Severity::Critical)
                    | (Severity::Warning, Severity::Critical | Severity::Warning)
                    | (Severity::Info, _)
            );
            if include {
                findings.push(finding);
            } else {
                passed.push(finding.check);
            }
        } else {
            findings.push(finding);
        }
    }

    let critical = findings
        .iter()
        .filter(|f| f.severity == Severity::Critical)
        .count();
    let warnings = findings
        .iter()
        .filter(|f| f.severity == Severity::Warning)
        .count();
    let infos = findings
        .iter()
        .filter(|f| f.severity == Severity::Info)
        .count();

    Ok(VetResult {
        findings,
        passed,
        critical,
        warnings,
        infos,
    })
}

/// List all available checks with descriptions.
pub fn list_checks() -> Vec<(&'static str, Severity, &'static str)> {
    vec![
        (
            CHECK_SILENT,
            Severity::Critical,
            "Monitors with no notification channels — alerts fire into the void",
        ),
        (
            CHECK_STALE,
            Severity::Warning,
            "Monitors in \"No Data\" state — abandoned or misconfigured data source",
        ),
        (
            CHECK_MUTED,
            Severity::Warning,
            "Monitors muted indefinitely or for >30 days — meant to be temporary",
        ),
    ]
}
