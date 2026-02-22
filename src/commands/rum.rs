use anyhow::{bail, Result};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_rum::{ListRUMEventsOptionalParams, RUMAPI};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_rum_metrics::RumMetricsAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_rum_replay_heatmaps::{
    ListReplayHeatmapSnapshotsOptionalParams, RumReplayHeatmapsAPI,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_rum_replay_playlists::{
    ListRumReplayPlaylistsOptionalParams, RumReplayPlaylistsAPI,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_rum_retention_filters::RumRetentionFiltersAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::{
    RUMApplicationCreate, RUMApplicationCreateAttributes, RUMApplicationCreateRequest,
    RUMApplicationCreateType, RUMApplicationUpdateRequest, RUMQueryFilter, RUMSearchEventsRequest,
    RUMSort, RumMetricCreateRequest, RumMetricUpdateRequest, RumRetentionFilterCreateRequest,
    RumRetentionFilterUpdateRequest,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
#[cfg(not(target_arch = "wasm32"))]
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn apps_list(cfg: &Config) -> Result<()> {
    // RUM apps is OAuth-excluded — require API keys
    if !cfg.has_api_keys() {
        bail!("RUM apps requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RUMAPI::with_config(dd_cfg);
    let resp = api
        .get_rum_applications()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list RUM apps: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn apps_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/rum/applications", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn apps_get(cfg: &Config, app_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM apps requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RUMAPI::with_config(dd_cfg);
    let resp = api
        .get_rum_application(app_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get RUM app: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn apps_get(cfg: &Config, app_id: &str) -> Result<()> {
    let path = format!("/api/v2/rum/applications/{app_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn apps_create(cfg: &Config, name: &str, app_type: Option<String>) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM apps requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RUMAPI::with_config(dd_cfg);
    let mut attrs = RUMApplicationCreateAttributes::new(name.to_string());
    if let Some(t) = app_type {
        attrs = attrs.type_(t);
    }
    let data = RUMApplicationCreate::new(attrs, RUMApplicationCreateType::RUM_APPLICATION_CREATE);
    let body = RUMApplicationCreateRequest::new(data);
    let resp = api
        .create_rum_application(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create RUM app: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn apps_create(cfg: &Config, name: &str, app_type: Option<String>) -> Result<()> {
    let mut attrs = serde_json::json!({ "name": name });
    if let Some(t) = app_type {
        attrs["type"] = serde_json::Value::String(t);
    }
    let body = serde_json::json!({
        "data": {
            "attributes": attrs,
            "type": "rum_application_create"
        }
    });
    let data = crate::api::post(cfg, "/api/v2/rum/applications", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn apps_delete(cfg: &Config, app_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM apps requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RUMAPI::with_config(dd_cfg);
    api.delete_rum_application(app_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete RUM app: {e:?}"))?;
    println!("Successfully deleted RUM application {app_id}");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn apps_delete(cfg: &Config, app_id: &str) -> Result<()> {
    let path = format!("/api/v2/rum/applications/{app_id}");
    crate::api::delete(cfg, &path).await?;
    println!("Successfully deleted RUM application {app_id}");
    Ok(())
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn events_list(cfg: &Config, from: String, to: String, limit: i32) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => RUMAPI::with_client_and_config(dd_cfg, c),
        None => RUMAPI::with_config(dd_cfg),
    };

    let from_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&from)?).unwrap();
    let to_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&to)?).unwrap();

    let params = ListRUMEventsOptionalParams::default()
        .filter_from(from_dt)
        .filter_to(to_dt)
        .page_limit(limit);

    let resp = api
        .list_rum_events(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list RUM events: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn events_list(cfg: &Config, from: String, to: String, limit: i32) -> Result<()> {
    let from_ms = crate::util::parse_time_to_unix_millis(&from)?;
    let to_ms = crate::util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();
    let query = vec![
        ("filter[from]", from_str),
        ("filter[to]", to_str),
        ("page[limit]", limit.to_string()),
    ];
    let data = crate::api::get(cfg, "/api/v2/rum/events", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn sessions_search(
    cfg: &Config,
    query: Option<String>,
    from: String,
    to: String,
    _limit: i32,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => RUMAPI::with_client_and_config(dd_cfg, c),
        None => RUMAPI::with_config(dd_cfg),
    };

    let from_str = chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&from)?)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&to)?)
        .unwrap()
        .to_rfc3339();

    let mut filter = RUMQueryFilter::new().from(from_str).to(to_str);
    if let Some(q) = query {
        filter = filter.query(q);
    }

    let body = RUMSearchEventsRequest::new()
        .filter(filter)
        .sort(RUMSort::TIMESTAMP_DESCENDING);

    let resp = api
        .search_rum_events(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search RUM sessions: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn sessions_search(
    cfg: &Config,
    query: Option<String>,
    from: String,
    to: String,
    _limit: i32,
) -> Result<()> {
    let from_ms = crate::util::parse_time_to_unix_millis(&from)?;
    let to_ms = crate::util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();
    let mut filter = serde_json::json!({ "from": from_str, "to": to_str });
    if let Some(q) = query {
        filter["query"] = serde_json::Value::String(q);
    }
    let body = serde_json::json!({
        "filter": filter,
        "sort": "-timestamp"
    });
    let data = crate::api::post(cfg, "/api/v2/rum/events/search", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn apps_update(cfg: &Config, app_id: &str, file: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM apps requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RUMAPI::with_config(dd_cfg);
    let body: RUMApplicationUpdateRequest = crate::util::read_json_file(file)?;
    let resp = api
        .update_rum_application(app_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update RUM app: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn apps_update(cfg: &Config, app_id: &str, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let path = format!("/api/v2/rum/applications/{app_id}");
    let data = crate::api::patch(cfg, &path, &body).await?;
    crate::formatter::output(cfg, &data)
}

// ---- RUM Metrics ----

#[cfg(not(target_arch = "wasm32"))]
pub async fn metrics_list(cfg: &Config) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM metrics requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumMetricsAPI::with_config(dd_cfg);
    let resp = api
        .list_rum_metrics()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list RUM metrics: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn metrics_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/rum/metrics", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn metrics_get(cfg: &Config, metric_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM metrics requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumMetricsAPI::with_config(dd_cfg);
    let resp = api
        .get_rum_metric(metric_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get RUM metric: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn metrics_get(cfg: &Config, metric_id: &str) -> Result<()> {
    let path = format!("/api/v2/rum/metrics/{metric_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn metrics_create(cfg: &Config, file: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM metrics requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumMetricsAPI::with_config(dd_cfg);
    let body: RumMetricCreateRequest = crate::util::read_json_file(file)?;
    let resp = api
        .create_rum_metric(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create RUM metric: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn metrics_create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/rum/metrics", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn metrics_update(cfg: &Config, metric_id: &str, file: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM metrics requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumMetricsAPI::with_config(dd_cfg);
    let body: RumMetricUpdateRequest = crate::util::read_json_file(file)?;
    let resp = api
        .update_rum_metric(metric_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update RUM metric: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn metrics_update(cfg: &Config, metric_id: &str, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let path = format!("/api/v2/rum/metrics/{metric_id}");
    let data = crate::api::patch(cfg, &path, &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn metrics_delete(cfg: &Config, metric_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM metrics requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumMetricsAPI::with_config(dd_cfg);
    api.delete_rum_metric(metric_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete RUM metric: {e:?}"))?;
    println!("RUM metric {metric_id} deleted.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn metrics_delete(cfg: &Config, metric_id: &str) -> Result<()> {
    let path = format!("/api/v2/rum/metrics/{metric_id}");
    crate::api::delete(cfg, &path).await?;
    println!("RUM metric {metric_id} deleted.");
    Ok(())
}

// ---- RUM Retention Filters ----

#[cfg(not(target_arch = "wasm32"))]
pub async fn retention_filters_list(cfg: &Config, app_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM retention filters requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumRetentionFiltersAPI::with_config(dd_cfg);
    let resp = api
        .list_retention_filters(app_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list RUM retention filters: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn retention_filters_list(cfg: &Config, app_id: &str) -> Result<()> {
    let path = format!("/api/v2/rum/applications/{app_id}/retention_filters");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn retention_filters_get(cfg: &Config, app_id: &str, filter_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM retention filters requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumRetentionFiltersAPI::with_config(dd_cfg);
    let resp = api
        .get_retention_filter(app_id.to_string(), filter_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get RUM retention filter: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn retention_filters_get(cfg: &Config, app_id: &str, filter_id: &str) -> Result<()> {
    let path = format!("/api/v2/rum/applications/{app_id}/retention_filters/{filter_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn retention_filters_create(cfg: &Config, app_id: &str, file: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM retention filters requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumRetentionFiltersAPI::with_config(dd_cfg);
    let body: RumRetentionFilterCreateRequest = crate::util::read_json_file(file)?;
    let resp = api
        .create_retention_filter(app_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create RUM retention filter: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn retention_filters_create(cfg: &Config, app_id: &str, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let path = format!("/api/v2/rum/applications/{app_id}/retention_filters");
    let data = crate::api::post(cfg, &path, &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn retention_filters_update(
    cfg: &Config,
    app_id: &str,
    filter_id: &str,
    file: &str,
) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM retention filters requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumRetentionFiltersAPI::with_config(dd_cfg);
    let body: RumRetentionFilterUpdateRequest = crate::util::read_json_file(file)?;
    let resp = api
        .update_retention_filter(app_id.to_string(), filter_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update RUM retention filter: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn retention_filters_update(
    cfg: &Config,
    app_id: &str,
    filter_id: &str,
    file: &str,
) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let path = format!("/api/v2/rum/applications/{app_id}/retention_filters/{filter_id}");
    let data = crate::api::patch(cfg, &path, &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn retention_filters_delete(cfg: &Config, app_id: &str, filter_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM retention filters requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumRetentionFiltersAPI::with_config(dd_cfg);
    api.delete_retention_filter(app_id.to_string(), filter_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete RUM retention filter: {e:?}"))?;
    println!("RUM retention filter {filter_id} deleted.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn retention_filters_delete(cfg: &Config, app_id: &str, filter_id: &str) -> Result<()> {
    let path = format!("/api/v2/rum/applications/{app_id}/retention_filters/{filter_id}");
    crate::api::delete(cfg, &path).await?;
    println!("RUM retention filter {filter_id} deleted.");
    Ok(())
}

// ---- RUM Sessions ----

#[cfg(not(target_arch = "wasm32"))]
pub async fn sessions_list(cfg: &Config, from: String, to: String, limit: i32) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => RUMAPI::with_client_and_config(dd_cfg, c),
        None => RUMAPI::with_config(dd_cfg),
    };

    let from_str = chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&from)?)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&to)?)
        .unwrap()
        .to_rfc3339();

    let filter = RUMQueryFilter::new()
        .from(from_str)
        .to(to_str)
        .query("@type:session".to_string());

    let body = RUMSearchEventsRequest::new()
        .filter(filter)
        .sort(RUMSort::TIMESTAMP_DESCENDING)
        .page(datadog_api_client::datadogV2::model::RUMQueryPageOptions::new().limit(limit));

    let resp = api
        .search_rum_events(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list RUM sessions: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn sessions_list(cfg: &Config, from: String, to: String, limit: i32) -> Result<()> {
    let from_ms = crate::util::parse_time_to_unix_millis(&from)?;
    let to_ms = crate::util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();
    let body = serde_json::json!({
        "filter": {
            "from": from_str,
            "to": to_str,
            "query": "@type:session"
        },
        "sort": "-timestamp",
        "page": {
            "limit": limit
        }
    });
    let data = crate::api::post(cfg, "/api/v2/rum/events/search", &body).await?;
    crate::formatter::output(cfg, &data)
}

// ---- RUM Playlists ----

#[cfg(not(target_arch = "wasm32"))]
pub async fn playlists_list(cfg: &Config) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM playlists requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumReplayPlaylistsAPI::with_config(dd_cfg);
    let resp = api
        .list_rum_replay_playlists(ListRumReplayPlaylistsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list RUM playlists: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn playlists_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/rum/replay/playlists", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn playlists_get(cfg: &Config, playlist_id: i32) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM playlists requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumReplayPlaylistsAPI::with_config(dd_cfg);
    let resp = api
        .get_rum_replay_playlist(playlist_id)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get RUM playlist: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn playlists_get(cfg: &Config, playlist_id: i32) -> Result<()> {
    let path = format!("/api/v2/rum/replay/playlists/{playlist_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

// ---- RUM Heatmaps ----

#[cfg(not(target_arch = "wasm32"))]
pub async fn heatmaps_query(cfg: &Config, view_name: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("RUM heatmaps requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = RumReplayHeatmapsAPI::with_config(dd_cfg);
    let resp = api
        .list_replay_heatmap_snapshots(
            view_name.to_string(),
            ListReplayHeatmapSnapshotsOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to query RUM heatmaps: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn heatmaps_query(cfg: &Config, view_name: &str) -> Result<()> {
    let query = vec![("view_name", view_name.to_string())];
    let data = crate::api::get(cfg, "/api/v2/rum/replay/heatmap/snapshots", &query).await?;
    crate::formatter::output(cfg, &data)
}
