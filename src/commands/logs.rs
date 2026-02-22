use anyhow::{bail, Result};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_logs::{ListLogsOptionalParams, LogsAPI};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_logs_archives::LogsArchivesAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_logs_custom_destinations::LogsCustomDestinationsAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_logs_metrics::LogsMetricsAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::{
    LogsAggregateRequest, LogsAggregationFunction, LogsCompute, LogsListRequest,
    LogsListRequestPage, LogsQueryFilter, LogsSort,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    // Logs search API doesn't support OAuth/bearer - force API keys
    if !cfg.has_api_keys() {
        bail!(
            "logs search requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    // Force API key auth only - do NOT use bearer middleware
    let api = LogsAPI::with_config(dd_cfg);

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;

    let body = LogsListRequest::new()
        .filter(
            LogsQueryFilter::new()
                .query(query)
                .from(from_ms.to_string())
                .to(to_ms.to_string()),
        )
        .page(LogsListRequestPage::new().limit(limit))
        .sort(LogsSort::TIMESTAMP_DESCENDING);

    let params = ListLogsOptionalParams::default().body(body);

    let resp = api
        .list_logs(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search logs: {:?}", e))?;

    formatter::output(cfg, &resp)?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;
    let body = serde_json::json!({
        "filter": {
            "query": query,
            "from": from_ms.to_string(),
            "to": to_ms.to_string()
        },
        "page": { "limit": limit },
        "sort": "-timestamp"
    });
    let data = crate::api::post(cfg, "/api/v2/logs/events/search", &body).await?;
    crate::formatter::output(cfg, &data)
}

/// Alias for `search` with the same interface.
pub async fn list(cfg: &Config, query: String, from: String, to: String, limit: i32) -> Result<()> {
    search(cfg, query, from, to, limit).await
}

/// Alias for `search` with the same interface.
pub async fn query(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    search(cfg, query, from, to, limit).await
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn aggregate(cfg: &Config, query: String, from: String, to: String) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!(
            "logs aggregate requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let api = LogsAPI::with_config(dd_cfg);

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;

    let body = LogsAggregateRequest::new()
        .filter(
            LogsQueryFilter::new()
                .query(query)
                .from(from_ms.to_string())
                .to(to_ms.to_string()),
        )
        .compute(vec![LogsCompute::new(LogsAggregationFunction::COUNT)]);

    let resp = api
        .aggregate_logs(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to aggregate logs: {:?}", e))?;

    formatter::output(cfg, &resp)?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn aggregate(cfg: &Config, query: String, from: String, to: String) -> Result<()> {
    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;
    let body = serde_json::json!({
        "filter": {
            "query": query,
            "from": from_ms.to_string(),
            "to": to_ms.to_string()
        },
        "compute": [{ "type": "count" }]
    });
    let data = crate::api::post(cfg, "/api/v2/logs/analytics/aggregate", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn archives_list(cfg: &Config) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!(
            "logs archives list requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let api = LogsArchivesAPI::with_config(dd_cfg);

    let resp = api
        .list_logs_archives()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list log archives: {:?}", e))?;

    formatter::output(cfg, &resp)?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn archives_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/logs/config/archives", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn archives_get(cfg: &Config, archive_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!(
            "logs archives get requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let api = LogsArchivesAPI::with_config(dd_cfg);

    let resp = api
        .get_logs_archive(archive_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get log archive: {:?}", e))?;

    formatter::output(cfg, &resp)?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn archives_get(cfg: &Config, archive_id: &str) -> Result<()> {
    let path = format!("/api/v2/logs/config/archives/{archive_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn archives_delete(cfg: &Config, archive_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!(
            "logs archives delete requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let api = LogsArchivesAPI::with_config(dd_cfg);

    api.delete_logs_archive(archive_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete log archive: {:?}", e))?;

    println!("Log archive {archive_id} deleted.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn archives_delete(cfg: &Config, archive_id: &str) -> Result<()> {
    let path = format!("/api/v2/logs/config/archives/{archive_id}");
    crate::api::delete(cfg, &path).await?;
    println!("Log archive {archive_id} deleted.");
    Ok(())
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn custom_destinations_list(cfg: &Config) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!(
            "logs custom-destinations list requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let api = LogsCustomDestinationsAPI::with_config(dd_cfg);

    let resp = api
        .list_logs_custom_destinations()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list custom destinations: {:?}", e))?;

    formatter::output(cfg, &resp)?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn custom_destinations_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/logs/config/custom_destinations", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn custom_destinations_get(cfg: &Config, destination_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!(
            "logs custom-destinations get requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let api = LogsCustomDestinationsAPI::with_config(dd_cfg);

    let resp = api
        .get_logs_custom_destination(destination_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get custom destination: {:?}", e))?;

    formatter::output(cfg, &resp)?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn custom_destinations_get(cfg: &Config, destination_id: &str) -> Result<()> {
    let path = format!("/api/v2/logs/config/custom_destinations/{destination_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn metrics_list(cfg: &Config) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!(
            "logs metrics list requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let api = LogsMetricsAPI::with_config(dd_cfg);

    let resp = api
        .list_logs_metrics()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list log-based metrics: {:?}", e))?;

    formatter::output(cfg, &resp)?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn metrics_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/logs/config/metrics", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn metrics_get(cfg: &Config, metric_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!(
            "logs metrics get requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let api = LogsMetricsAPI::with_config(dd_cfg);

    let resp = api
        .get_logs_metric(metric_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get log-based metric: {:?}", e))?;

    formatter::output(cfg, &resp)?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn metrics_get(cfg: &Config, metric_id: &str) -> Result<()> {
    let path = format!("/api/v2/logs/config/metrics/{metric_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn metrics_delete(cfg: &Config, metric_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!(
            "logs metrics delete requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let api = LogsMetricsAPI::with_config(dd_cfg);

    api.delete_logs_metric(metric_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete log-based metric: {:?}", e))?;

    println!("Log-based metric {metric_id} deleted.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn metrics_delete(cfg: &Config, metric_id: &str) -> Result<()> {
    let path = format!("/api/v2/logs/config/metrics/{metric_id}");
    crate::api::delete(cfg, &path).await?;
    println!("Log-based metric {metric_id} deleted.");
    Ok(())
}

// ---------------------------------------------------------------------------
// Restriction Queries (raw HTTP - not available in typed client)
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
async fn raw_get(cfg: &Config, path: &str) -> Result<serde_json::Value> {
    let url = format!("{}{}", cfg.api_base_url(), path);
    let client = reqwest::Client::new();
    let mut req = client.get(&url);

    if let Some(token) = &cfg.access_token {
        req = req.header("Authorization", format!("Bearer {token}"));
    } else if let (Some(api_key), Some(app_key)) = (&cfg.api_key, &cfg.app_key) {
        req = req
            .header("DD-API-KEY", api_key.as_str())
            .header("DD-APPLICATION-KEY", app_key.as_str());
    } else {
        bail!("no authentication configured");
    }

    let resp = req.header("Accept", "application/json").send().await?;
    if !resp.status().is_success() {
        let status = resp.status();
        let body = resp.text().await.unwrap_or_default();
        bail!("API error (HTTP {status}): {body}");
    }
    Ok(resp.json().await?)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn restriction_queries_list(cfg: &Config) -> Result<()> {
    let data = raw_get(cfg, "/api/v2/logs/config/restriction_queries").await?;
    formatter::output(cfg, &data)
}

#[cfg(target_arch = "wasm32")]
pub async fn restriction_queries_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/logs/config/restriction_queries", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn restriction_queries_get(cfg: &Config, query_id: &str) -> Result<()> {
    let path = format!("/api/v2/logs/config/restriction_queries/{query_id}");
    let data = raw_get(cfg, &path).await?;
    formatter::output(cfg, &data)
}

#[cfg(target_arch = "wasm32")]
pub async fn restriction_queries_get(cfg: &Config, query_id: &str) -> Result<()> {
    let path = format!("/api/v2/logs/config/restriction_queries/{query_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}
