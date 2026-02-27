use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_dashboards::{DashboardsAPI, ListDashboardsOptionalParams};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::model::Dashboard;

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = DashboardsAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_dashboards(ListDashboardsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list dashboards: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v1/dashboard", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn get(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = DashboardsAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .get_dashboard(id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get dashboard: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn get(cfg: &Config, id: &str) -> Result<()> {
    let data = crate::api::get(cfg, &format!("/api/v1/dashboard/{id}"), &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn create(cfg: &Config, file: &str) -> Result<()> {
    let body: Dashboard = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = DashboardsAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .create_dashboard(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create dashboard: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v1/dashboard", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn update(cfg: &Config, id: &str, file: &str) -> Result<()> {
    let body: Dashboard = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = DashboardsAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .update_dashboard(id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update dashboard: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn update(cfg: &Config, id: &str, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::put(cfg, &format!("/api/v1/dashboard/{id}"), &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn delete(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = DashboardsAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .delete_dashboard(id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete dashboard: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn delete(cfg: &Config, id: &str) -> Result<()> {
    let data = crate::api::delete(cfg, &format!("/api/v1/dashboard/{id}")).await?;
    crate::formatter::output(cfg, &data)
}
