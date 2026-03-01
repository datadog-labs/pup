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
const CHECK_UNTAGGED: &str = "untagged-monitors";
const CHECK_NO_RECOVERY: &str = "no-recovery-threshold";
const CHECK_FAST_RENOTIFY: &str = "fast-renotify-interval";
const CHECK_PAGER_BURDEN: &str = "pager-burden";

const ALL_CHECKS: &[&str] = &[
    CHECK_SILENT,
    CHECK_STALE,
    CHECK_MUTED,
    CHECK_UNTAGGED,
    CHECK_NO_RECOVERY,
    CHECK_FAST_RENOTIFY,
    CHECK_PAGER_BURDEN,
];

// ---- Notification handle helpers ----

/// Classification of a Datadog notification @-handle by impact level.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum HandleKind {
    /// Datadog On-Call (@oncall-<schedule>)
    DdOnCall,
    /// PagerDuty (@pagerduty or @pagerduty-<service>)
    PagerDuty,
    /// OpsGenie (@opsgenie-<team>)
    OpsGenie,
    /// VictorOps / Splunk On-Call (@victorops-<team>)
    VictorOps,
    /// Everything else: Slack, email, webhook, etc.
    Other,
}

impl HandleKind {
    fn is_pager(self) -> bool {
        !matches!(self, HandleKind::Other)
    }

    fn display(self) -> &'static str {
        match self {
            HandleKind::DdOnCall => "Datadog On-Call",
            HandleKind::PagerDuty => "PagerDuty",
            HandleKind::OpsGenie => "OpsGenie",
            HandleKind::VictorOps => "VictorOps",
            HandleKind::Other => "other",
        }
    }
}

/// Classify a bare handle name (without the leading `@`).
fn classify_handle(handle: &str) -> HandleKind {
    if handle.starts_with("oncall-") {
        HandleKind::DdOnCall
    } else if handle.starts_with("pagerduty") {
        HandleKind::PagerDuty
    } else if handle.starts_with("opsgenie-") {
        HandleKind::OpsGenie
    } else if handle.starts_with("victorops-") {
        HandleKind::VictorOps
    } else {
        HandleKind::Other
    }
}

/// Extract all `@handle` tokens from a monitor message.
/// A handle is `[a-zA-Z0-9_-]+` immediately following `@`.
/// Stops at whitespace, punctuation, or end-of-string.
fn extract_handles(msg: &str) -> Vec<&str> {
    let mut handles = Vec::new();
    let bytes = msg.as_bytes();
    let mut i = 0;
    while i < bytes.len() {
        if bytes[i] == b'@' {
            let start = i + 1;
            let len = bytes[start..]
                .iter()
                .take_while(|&&b| b.is_ascii_alphanumeric() || b == b'-' || b == b'_')
                .count();
            if len > 0 {
                handles.push(&msg[start..start + len]);
            }
            i = start + len;
        } else {
            i += 1;
        }
    }
    handles
}

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

/// Monitors with no tags — can't be filtered, routed, or grouped.
fn check_untagged_monitors(monitors: &[Monitor]) -> Finding {
    let resources: Vec<Resource> = monitors
        .iter()
        .filter(|m| m.tags.as_ref().map(|t| t.is_empty()).unwrap_or(true))
        .map(|m| Resource {
            id: m.id.unwrap_or(0),
            name: m.name.as_deref().unwrap_or("(unnamed)").to_string(),
            detail: "no tags set".to_string(),
        })
        .collect();

    Finding {
        check: CHECK_UNTAGGED,
        severity: Severity::Warning,
        count: resources.len(),
        resources,
        recommendation:
            "Add tags (e.g. team:, service:, env:) so monitors can be filtered, routed, and grouped",
    }
}

/// Monitors with a critical threshold but no critical recovery threshold — flapping risk.
fn check_no_recovery_threshold(monitors: &[Monitor]) -> Finding {
    let resources: Vec<Resource> = monitors
        .iter()
        .filter_map(|m| {
            let thresholds = m.options.as_ref()?.thresholds.as_ref()?;
            let has_critical = thresholds.critical.is_some();
            // critical_recovery is Option<Option<f64>>: outer None = absent, inner None = explicit null
            let has_critical_recovery = thresholds
                .critical_recovery
                .as_ref()
                .and_then(|r| r.as_ref())
                .is_some();
            if has_critical && !has_critical_recovery {
                Some(Resource {
                    id: m.id.unwrap_or(0),
                    name: m.name.as_deref().unwrap_or("(unnamed)").to_string(),
                    detail: "critical threshold set but no critical_recovery threshold".to_string(),
                })
            } else {
                None
            }
        })
        .collect();

    Finding {
        check: CHECK_NO_RECOVERY,
        severity: Severity::Info,
        count: resources.len(),
        resources,
        recommendation:
            "Set a critical_recovery threshold to add hysteresis and prevent alert flapping",
    }
}

/// Monitors configured with a renotify interval ≤60 min — will spam on-call if they fire.
/// This is a configuration audit regardless of current alert state.
fn check_fast_renotify_interval(monitors: &[Monitor]) -> Finding {
    let resources: Vec<Resource> = monitors
        .iter()
        .filter_map(|m| {
            // renotify_interval is Option<Option<i64>>
            let renotify = m.options.as_ref()?.renotify_interval.flatten()?;
            if renotify <= 0 || renotify > 60 {
                return None;
            }
            Some(Resource {
                id: m.id.unwrap_or(0),
                name: m.name.as_deref().unwrap_or("(unnamed)").to_string(),
                detail: format!("renotify_interval = {renotify} min"),
            })
        })
        .collect();

    Finding {
        check: CHECK_FAST_RENOTIFY,
        severity: Severity::Info,
        count: resources.len(),
        resources,
        recommendation:
            "Consider raising renotify_interval (>60 min) to avoid notification storms if this monitor fires",
    }
}

/// Monitors currently in ALERT that are actively paging through high-impact tools
/// (Datadog On-Call, PagerDuty, OpsGenie, VictorOps) or re-notifying via any channel.
/// Pager-tool entries are sorted first.
fn check_pager_burden(monitors: &[Monitor]) -> Finding {
    // Collect (is_pager, Resource) so we can sort before flattening.
    let mut tagged: Vec<(bool, Resource)> = monitors
        .iter()
        .filter(|m| matches!(m.overall_state, Some(MonitorOverallStates::ALERT)))
        .filter_map(|m| {
            let message = m.message.as_deref().unwrap_or("");
            let handles = extract_handles(message);
            if handles.is_empty() {
                return None; // silent-monitors catches no-notification monitors
            }

            let classified: Vec<(&str, HandleKind)> = handles
                .iter()
                .map(|&h| (h, classify_handle(h)))
                .collect();

            let pager_entries: Vec<_> = classified
                .iter()
                .filter(|(_, k)| k.is_pager())
                .collect();

            let renotify = m
                .options
                .as_ref()
                .and_then(|o| o.renotify_interval)
                .flatten()
                .unwrap_or(0);

            // Only include if actively paging via pager tool OR re-notifying frequently
            let is_pager = !pager_entries.is_empty();
            if !is_pager && (renotify <= 0 || renotify > 60) {
                return None;
            }

            let detail = if is_pager {
                // Deduplicate by tool name and build "Tool (@handle)" list
                let mut seen_tools = std::collections::HashSet::new();
                let tools: Vec<String> = pager_entries
                    .iter()
                    .filter(|(_, k)| seen_tools.insert(k.display()))
                    .map(|(h, k)| format!("{} (@{})", k.display(), h))
                    .collect();
                let mut s = format!("paging via {}", tools.join(", "));
                if renotify > 0 {
                    s.push_str(&format!("; re-notifying every {renotify} min"));
                }
                s
            } else {
                // Non-pager handles with short renotify
                let shown: Vec<_> = handles.iter().take(2).map(|h| format!("@{h}")).collect();
                format!("notifying {} every {renotify} min", shown.join(", "))
            };

            Some((
                is_pager,
                Resource {
                    id: m.id.unwrap_or(0),
                    name: m.name.as_deref().unwrap_or("(unnamed)").to_string(),
                    detail,
                },
            ))
        })
        .collect();

    // Pager-tool monitors first, then Slack/webhook re-notifiers
    tagged.sort_by_key(|(is_pager, _)| if *is_pager { 0u8 } else { 1u8 });
    let resources: Vec<Resource> = tagged.into_iter().map(|(_, r)| r).collect();

    Finding {
        check: CHECK_PAGER_BURDEN,
        severity: Severity::Warning,
        count: resources.len(),
        resources,
        recommendation: "Resolve active alerts — Datadog On-Call and PagerDuty pages are waking up on-call responders",
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
            CHECK_UNTAGGED => check_untagged_monitors(&monitors),
            CHECK_NO_RECOVERY => check_no_recovery_threshold(&monitors),
            CHECK_FAST_RENOTIFY => check_fast_renotify_interval(&monitors),
            CHECK_PAGER_BURDEN => check_pager_burden(&monitors),
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
        (
            CHECK_UNTAGGED,
            Severity::Warning,
            "Monitors without any tags — can't be filtered, routed, or grouped",
        ),
        (
            CHECK_NO_RECOVERY,
            Severity::Info,
            "Monitors with no critical_recovery threshold — flapping risk",
        ),
        (
            CHECK_FAST_RENOTIFY,
            Severity::Info,
            "Monitors configured with renotify_interval ≤60 min — will spam on-call if they fire",
        ),
        (
            CHECK_PAGER_BURDEN,
            Severity::Warning,
            "Monitors currently alerting via pager tools (DD On-Call, PagerDuty, OpsGenie, VictorOps)",
        ),
    ]
}

// ---- Tests ----

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn extract_handles_basic() {
        let handles = extract_handles("Alert! @pagerduty-prod @slack-alerts and @oncall-platform");
        assert_eq!(handles, vec!["pagerduty-prod", "slack-alerts", "oncall-platform"]);
    }

    #[test]
    fn extract_handles_no_handles() {
        assert!(extract_handles("no notifications here").is_empty());
    }

    #[test]
    fn extract_handles_email_like() {
        // Emails produce two tokens; neither will match pager prefixes
        let handles = extract_handles("notify user@example.com and @pagerduty-svc");
        assert!(handles.contains(&"pagerduty-svc"));
    }

    #[test]
    fn classify_pager_handles() {
        assert!(classify_handle("oncall-platform").is_pager());
        assert!(classify_handle("pagerduty-prod").is_pager());
        assert!(classify_handle("pagerduty").is_pager());
        assert!(classify_handle("opsgenie-sre").is_pager());
        assert!(classify_handle("victorops-team").is_pager());
        assert!(!classify_handle("slack-alerts").is_pager());
        assert!(!classify_handle("webhook-myapp").is_pager());
    }

    #[test]
    fn classify_handle_display_names() {
        assert_eq!(classify_handle("oncall-x").display(), "Datadog On-Call");
        assert_eq!(classify_handle("pagerduty-x").display(), "PagerDuty");
        assert_eq!(classify_handle("opsgenie-x").display(), "OpsGenie");
        assert_eq!(classify_handle("victorops-x").display(), "VictorOps");
        assert_eq!(classify_handle("slack-x").display(), "other");
    }
}
