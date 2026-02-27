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
    LogsListRequestPage, LogsQueryFilter, LogsSort, LogsStorageTier,
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
    storage: Option<String>,
) -> Result<()> {
    if !cfg.has_bearer_token() && !cfg.has_api_keys() {
        bail!(
            "logs search requires authentication.\n\
             Set DD_ACCESS_TOKEN for OAuth or DD_API_KEY + DD_APP_KEY for API keys."
        );
    }

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;

    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = LogsAPI::with_client_and_config(dd_cfg, dd_client);

    let mut filter = LogsQueryFilter::new()
        .query(query)
        .from(from_ms.to_string())
        .to(to_ms.to_string());
    if let Some(s) = storage {
        let tier: LogsStorageTier = serde_json::from_value(serde_json::json!(s)).map_err(|_| {
            anyhow::anyhow!(
                "invalid storage tier '{s}'; valid values: indexes, online-archives, flex"
            )
        })?;
        filter = filter.storage_tier(tier);
    }

    let body = LogsListRequest::new()
        .filter(filter)
        .page(LogsListRequestPage::new().limit(limit))
        .sort(LogsSort::TIMESTAMP_DESCENDING);

    let params = ListLogsOptionalParams::default().body(body);

    let resp = api
        .list_logs(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search logs: {:?}", e))?;

    let meta = if cfg.agent_mode {
        let count = resp.data.as_ref().map(|d| d.len());
        let truncated = count.is_some_and(|c| c as i32 >= limit);
        Some(formatter::Metadata {
            count,
            truncated,
            command: Some("logs search".into()),
            next_action: if truncated {
                Some(format!(
                    "Results may be truncated at {limit}. Use --limit={} or narrow the --query",
                    limit + 1
                ))
            } else {
                None
            },
        })
    } else {
        None
    };
    formatter::format_and_print(&resp, &cfg.output_format, cfg.agent_mode, meta.as_ref())?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
    _storage: Option<String>,
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
pub async fn list(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
    storage: Option<String>,
) -> Result<()> {
    search(cfg, query, from, to, limit, storage).await
}

/// Alias for `search` with the same interface.
pub async fn query(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
    storage: Option<String>,
) -> Result<()> {
    search(cfg, query, from, to, limit, storage).await
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn aggregate(cfg: &Config, query: String, from: String, to: String) -> Result<()> {
    if !cfg.has_bearer_token() && !cfg.has_api_keys() {
        bail!(
            "logs aggregate requires authentication.\n\
             Set DD_ACCESS_TOKEN for OAuth or DD_API_KEY + DD_APP_KEY for API keys."
        );
    }

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;

    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = LogsAPI::with_client_and_config(dd_cfg, dd_client);

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
    if !cfg.has_bearer_token() && !cfg.has_api_keys() {
        bail!(
            "logs archives list requires authentication.\n\
             Set DD_ACCESS_TOKEN for OAuth or DD_API_KEY + DD_APP_KEY for API keys."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = LogsArchivesAPI::with_client_and_config(dd_cfg, dd_client);

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
    if !cfg.has_bearer_token() && !cfg.has_api_keys() {
        bail!(
            "logs archives get requires authentication.\n\
             Set DD_ACCESS_TOKEN for OAuth or DD_API_KEY + DD_APP_KEY for API keys."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = LogsArchivesAPI::with_client_and_config(dd_cfg, dd_client);

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
    if !cfg.has_bearer_token() && !cfg.has_api_keys() {
        bail!(
            "logs archives delete requires authentication.\n\
             Set DD_ACCESS_TOKEN for OAuth or DD_API_KEY + DD_APP_KEY for API keys."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = LogsArchivesAPI::with_client_and_config(dd_cfg, dd_client);

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
    if !cfg.has_bearer_token() && !cfg.has_api_keys() {
        bail!(
            "logs custom-destinations list requires authentication.\n\
             Set DD_ACCESS_TOKEN for OAuth or DD_API_KEY + DD_APP_KEY for API keys."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = LogsCustomDestinationsAPI::with_client_and_config(dd_cfg, dd_client);

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
    if !cfg.has_bearer_token() && !cfg.has_api_keys() {
        bail!(
            "logs custom-destinations get requires authentication.\n\
             Set DD_ACCESS_TOKEN for OAuth or DD_API_KEY + DD_APP_KEY for API keys."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = LogsCustomDestinationsAPI::with_client_and_config(dd_cfg, dd_client);

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
    if !cfg.has_bearer_token() && !cfg.has_api_keys() {
        bail!(
            "logs metrics list requires authentication.\n\
             Set DD_ACCESS_TOKEN for OAuth or DD_API_KEY + DD_APP_KEY for API keys."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = LogsMetricsAPI::with_client_and_config(dd_cfg, dd_client);

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
    if !cfg.has_bearer_token() && !cfg.has_api_keys() {
        bail!(
            "logs metrics get requires authentication.\n\
             Set DD_ACCESS_TOKEN for OAuth or DD_API_KEY + DD_APP_KEY for API keys."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = LogsMetricsAPI::with_client_and_config(dd_cfg, dd_client);

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
    if !cfg.has_bearer_token() && !cfg.has_api_keys() {
        bail!(
            "logs metrics delete requires authentication.\n\
             Set DD_ACCESS_TOKEN for OAuth or DD_API_KEY + DD_APP_KEY for API keys."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = LogsMetricsAPI::with_client_and_config(dd_cfg, dd_client);

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
pub async fn restriction_queries_list(cfg: &Config) -> Result<()> {
    let data = client::raw_get(cfg, "/api/v2/logs/config/restriction_queries").await?;
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
    let data = client::raw_get(cfg, &path).await?;
    formatter::output(cfg, &data)
}

#[cfg(target_arch = "wasm32")]
pub async fn restriction_queries_get(cfg: &Config, query_id: &str) -> Result<()> {
    let path = format!("/api/v2/logs/config/restriction_queries/{query_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}
