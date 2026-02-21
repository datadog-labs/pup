use anyhow::Result;
use datadog_api_client::datadogV1::api_metrics::{
    MetricsAPI as MetricsV1API, ListActiveMetricsOptionalParams,
};

use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

pub async fn list(cfg: &Config, filter: Option<String>, from: String) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => MetricsV1API::with_client_and_config(dd_cfg, c),
        None => MetricsV1API::with_config(dd_cfg),
    };

    let from_ts = util::parse_time_to_unix(&from)?;
    let params = ListActiveMetricsOptionalParams::default();

    let resp = api
        .list_active_metrics(from_ts, params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list metrics: {e:?}"))?;

    // Client-side filter if provided
    if let Some(pattern) = filter {
        let pattern = pattern.to_lowercase();
        if let Some(metrics) = &resp.metrics {
            let filtered: Vec<_> = metrics
                .iter()
                .filter(|m| m.to_lowercase().contains(&pattern))
                .collect();
            return formatter::print_json(&filtered);
        }
    }

    formatter::print_json(&resp)
}

pub async fn search(cfg: &Config, query: String, from: String, to: String) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => MetricsV1API::with_client_and_config(dd_cfg, c),
        None => MetricsV1API::with_config(dd_cfg),
    };

    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;

    let resp = api
        .query_metrics(from_ts, to_ts, query)
        .await
        .map_err(|e| anyhow::anyhow!("failed to query metrics: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn metadata_get(cfg: &Config, metric_name: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => MetricsV1API::with_client_and_config(dd_cfg, c),
        None => MetricsV1API::with_config(dd_cfg),
    };
    let resp = api
        .get_metric_metadata(metric_name.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get metric metadata: {e:?}"))?;
    formatter::print_json(&resp)
}
