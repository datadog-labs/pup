use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_high_availability_multi_region::HighAvailabilityMultiRegionAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::HamrOrgConnectionRequest;

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn connections_get(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = HighAvailabilityMultiRegionAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .get_hamr_org_connection()
        .await
        .map_err(|e| anyhow::anyhow!("failed to get HAMR connection: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn connections_get(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/hamr/connections", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn connections_create(cfg: &Config, file: &str) -> Result<()> {
    let body: HamrOrgConnectionRequest = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = HighAvailabilityMultiRegionAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .create_hamr_org_connection(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create HAMR connection: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn connections_create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/hamr/connections", &body).await?;
    crate::formatter::output(cfg, &data)
}
