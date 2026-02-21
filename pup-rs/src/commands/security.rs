use anyhow::Result;
use datadog_api_client::datadogV2::api_security_monitoring::{
    SecurityMonitoringAPI, ListSecurityMonitoringRulesOptionalParams,
    SearchSecurityMonitoringSignalsOptionalParams,
};
use datadog_api_client::datadogV2::model::{
    SecurityMonitoringSignalListRequest, SecurityMonitoringSignalListRequestFilter,
    SecurityMonitoringSignalListRequestPage, SecurityMonitoringSignalsSort,
};

use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

pub async fn rules_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SecurityMonitoringAPI::with_client_and_config(dd_cfg, c),
        None => SecurityMonitoringAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_security_monitoring_rules(ListSecurityMonitoringRulesOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list rules: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn rules_get(cfg: &Config, rule_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SecurityMonitoringAPI::with_client_and_config(dd_cfg, c),
        None => SecurityMonitoringAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_security_monitoring_rule(rule_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get rule: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn signals_search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SecurityMonitoringAPI::with_client_and_config(dd_cfg, c),
        None => SecurityMonitoringAPI::with_config(dd_cfg),
    };

    let from_dt = chrono::DateTime::from_timestamp_millis(
        util::parse_time_to_unix_millis(&from)?,
    )
    .unwrap();
    let to_dt = chrono::DateTime::from_timestamp_millis(
        util::parse_time_to_unix_millis(&to)?,
    )
    .unwrap();

    let body = SecurityMonitoringSignalListRequest::new()
        .filter(
            SecurityMonitoringSignalListRequestFilter::new()
                .query(query)
                .from(from_dt)
                .to(to_dt),
        )
        .page(SecurityMonitoringSignalListRequestPage::new().limit(limit))
        .sort(SecurityMonitoringSignalsSort::TIMESTAMP_DESCENDING);

    let params = SearchSecurityMonitoringSignalsOptionalParams::default().body(body);
    let resp = api
        .search_security_monitoring_signals(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search signals: {e:?}"))?;
    formatter::print_json(&resp)
}
