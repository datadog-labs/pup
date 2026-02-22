use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_monitors::{
    DeleteMonitorOptionalParams, GetMonitorOptionalParams, ListMonitorsOptionalParams, MonitorsAPI,
    SearchMonitorsOptionalParams,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::model::Monitor;

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter::{self, Metadata};
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(
    cfg: &Config,
    name: Option<String>,
    tags: Option<String>,
    limit: i32,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = if let Some(http_client) = client::make_bearer_client(cfg) {
        MonitorsAPI::with_client_and_config(dd_cfg, http_client)
    } else {
        MonitorsAPI::with_config(dd_cfg)
    };

    let mut params = ListMonitorsOptionalParams::default();
    if let Some(name) = name {
        params = params.name(name);
    }
    if let Some(tags) = tags {
        params = params.monitor_tags(tags);
    }

    let limit = limit.clamp(1, 1000);
    params = params.page_size(limit).page(0);

    let monitors = api
        .list_monitors(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list monitors: {:?}", e))?;

    if monitors.is_empty() {
        eprintln!("No monitors found matching the specified criteria.");
        return Ok(());
    }

    let monitors: Vec<_> = monitors.into_iter().take(limit as usize).collect();
    let meta = Metadata {
        count: Some(monitors.len()),
        truncated: false,
        command: Some("monitors list".to_string()),
        next_action: None,
    };
    formatter::format_and_print(&monitors, &cfg.output_format, cfg.agent_mode, Some(&meta))?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn list(
    cfg: &Config,
    name: Option<String>,
    tags: Option<String>,
    limit: i32,
) -> Result<()> {
    let mut query = vec![];
    if let Some(n) = &name {
        query.push(("name", n.clone()));
    }
    if let Some(t) = &tags {
        query.push(("monitor_tags", t.clone()));
    }
    let limit = limit.clamp(1, 1000);
    query.push(("page_size", limit.to_string()));
    query.push(("page", "0".to_string()));
    let data = crate::api::get(cfg, "/api/v1/monitor", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn get(cfg: &Config, monitor_id: i64) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = if let Some(http_client) = client::make_bearer_client(cfg) {
        MonitorsAPI::with_client_and_config(dd_cfg, http_client)
    } else {
        MonitorsAPI::with_config(dd_cfg)
    };
    let resp = api
        .get_monitor(monitor_id, GetMonitorOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get monitor: {:?}", e))?;
    let meta = Metadata {
        count: None,
        truncated: false,
        command: Some("monitors get".to_string()),
        next_action: None,
    };
    formatter::format_and_print(&resp, &cfg.output_format, cfg.agent_mode, Some(&meta))
}

#[cfg(target_arch = "wasm32")]
pub async fn get(cfg: &Config, monitor_id: i64) -> Result<()> {
    let data = crate::api::get(cfg, &format!("/api/v1/monitor/{monitor_id}"), &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn create(cfg: &Config, file: &str) -> Result<()> {
    let body: Monitor = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let api = if let Some(http_client) = client::make_bearer_client(cfg) {
        MonitorsAPI::with_client_and_config(dd_cfg, http_client)
    } else {
        MonitorsAPI::with_config(dd_cfg)
    };
    let resp = api
        .create_monitor(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create monitor: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v1/monitor", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn update(cfg: &Config, monitor_id: i64, file: &str) -> Result<()> {
    let body: datadog_api_client::datadogV1::model::MonitorUpdateRequest =
        util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let api = if let Some(http_client) = client::make_bearer_client(cfg) {
        MonitorsAPI::with_client_and_config(dd_cfg, http_client)
    } else {
        MonitorsAPI::with_config(dd_cfg)
    };
    let resp = api
        .update_monitor(monitor_id, body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update monitor: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn update(cfg: &Config, monitor_id: i64, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::put(cfg, &format!("/api/v1/monitor/{monitor_id}"), &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn search(cfg: &Config, query: Option<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = if let Some(http_client) = client::make_bearer_client(cfg) {
        MonitorsAPI::with_client_and_config(dd_cfg, http_client)
    } else {
        MonitorsAPI::with_config(dd_cfg)
    };

    let mut params = SearchMonitorsOptionalParams::default();
    if let Some(q) = query {
        params = params.query(q);
    }

    let resp = api
        .search_monitors(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search monitors: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn search(cfg: &Config, query: Option<String>) -> Result<()> {
    let mut q = vec![];
    if let Some(qstr) = &query {
        q.push(("query", qstr.clone()));
    }
    let data = crate::api::get(cfg, "/api/v1/monitor/search", &q).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn delete(cfg: &Config, monitor_id: i64) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = if let Some(http_client) = client::make_bearer_client(cfg) {
        MonitorsAPI::with_client_and_config(dd_cfg, http_client)
    } else {
        MonitorsAPI::with_config(dd_cfg)
    };
    let resp = api
        .delete_monitor(monitor_id, DeleteMonitorOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete monitor: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn delete(cfg: &Config, monitor_id: i64) -> Result<()> {
    let data = crate::api::delete(cfg, &format!("/api/v1/monitor/{monitor_id}")).await?;
    crate::formatter::output(cfg, &data)
}
