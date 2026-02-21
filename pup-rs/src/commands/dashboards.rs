use anyhow::Result;
use datadog_api_client::datadogV1::api_dashboards::{DashboardsAPI, ListDashboardsOptionalParams};

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => DashboardsAPI::with_client_and_config(dd_cfg, c),
        None => DashboardsAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_dashboards(ListDashboardsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list dashboards: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn get(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => DashboardsAPI::with_client_and_config(dd_cfg, c),
        None => DashboardsAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_dashboard(id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get dashboard: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn delete(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => DashboardsAPI::with_client_and_config(dd_cfg, c),
        None => DashboardsAPI::with_config(dd_cfg),
    };
    let resp = api
        .delete_dashboard(id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete dashboard: {e:?}"))?;
    formatter::print_json(&resp)
}
