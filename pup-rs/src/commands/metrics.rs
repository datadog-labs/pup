use anyhow::Result;
use datadog_api_client::datadogV1::api_metrics::{
    ListActiveMetricsOptionalParams, MetricsAPI as MetricsV1API,
};
use datadog_api_client::datadogV1::model::MetricMetadata;
use datadog_api_client::datadogV2::api_metrics::MetricsAPI as MetricsV2API;
use datadog_api_client::datadogV2::model::MetricPayload;

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
            return formatter::output(cfg, &filtered);
        }
    }

    formatter::output(cfg, &resp)
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
    formatter::output(cfg, &resp)
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
    formatter::output(cfg, &resp)
}

pub async fn query(cfg: &Config, query: String, from: String, to: String) -> Result<()> {
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
    formatter::output(cfg, &resp)
}

pub async fn metadata_update(cfg: &Config, metric_name: &str, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => MetricsV1API::with_client_and_config(dd_cfg, c),
        None => MetricsV1API::with_config(dd_cfg),
    };
    let body: MetricMetadata = util::read_json_file(file)?;
    let resp = api
        .update_metric_metadata(metric_name.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update metric metadata: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn submit(cfg: &Config, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => MetricsV2API::with_client_and_config(dd_cfg, c),
        None => MetricsV2API::with_config(dd_cfg),
    };
    let body: MetricPayload = util::read_json_file(file)?;
    let resp = api
        .submit_metrics(body, datadog_api_client::datadogV2::api_metrics::SubmitMetricsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to submit metrics: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn tags_list(cfg: &Config, metric_name: &str) -> Result<()> {
    use datadog_api_client::datadogV2::api_metrics::ListTagsByMetricNameOptionalParams;

    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => MetricsV2API::with_client_and_config(dd_cfg, c),
        None => MetricsV2API::with_config(dd_cfg),
    };
    let resp = api
        .list_tags_by_metric_name(
            metric_name.to_string(),
            ListTagsByMetricNameOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to list tags for metric {metric_name}: {e:?}"))?;
    formatter::output(cfg, &resp)
}
