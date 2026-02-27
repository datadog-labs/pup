use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_downtimes::{
    DowntimesAPI, GetDowntimeOptionalParams, ListDowntimesOptionalParams,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = DowntimesAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_downtimes(ListDowntimesOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list downtimes: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/downtime", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn get(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = DowntimesAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .get_downtime(id.to_string(), GetDowntimeOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get downtime: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn get(cfg: &Config, id: &str) -> Result<()> {
    let data = crate::api::get(cfg, &format!("/api/v2/downtime/{id}"), &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn create(cfg: &Config, file: &str) -> Result<()> {
    let body: datadog_api_client::datadogV2::model::DowntimeCreateRequest =
        crate::util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = DowntimesAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .create_downtime(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create downtime: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/downtime", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn cancel(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = DowntimesAPI::with_client_and_config(dd_cfg, dd_client);
    api.cancel_downtime(id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to cancel downtime: {e:?}"))?;
    println!("Downtime {id} cancelled.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn cancel(cfg: &Config, id: &str) -> Result<()> {
    crate::api::delete(cfg, &format!("/api/v2/downtime/{id}")).await?;
    println!("Downtime {id} cancelled.");
    Ok(())
}
