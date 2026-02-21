use anyhow::Result;
use datadog_api_client::datadogV2::api_downtimes::{
    DowntimesAPI, ListDowntimesOptionalParams, GetDowntimeOptionalParams,
};

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => DowntimesAPI::with_client_and_config(dd_cfg, c),
        None => DowntimesAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_downtimes(ListDowntimesOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list downtimes: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn get(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => DowntimesAPI::with_client_and_config(dd_cfg, c),
        None => DowntimesAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_downtime(id.to_string(), GetDowntimeOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get downtime: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn cancel(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => DowntimesAPI::with_client_and_config(dd_cfg, c),
        None => DowntimesAPI::with_config(dd_cfg),
    };
    api.cancel_downtime(id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to cancel downtime: {e:?}"))?;
    eprintln!("Downtime {id} cancelled.");
    Ok(())
}
