use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_sensitive_data_scanner::SensitiveDataScannerAPI;

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

#[cfg(not(target_arch = "wasm32"))]
pub async fn scanner_rules_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SensitiveDataScannerAPI::with_client_and_config(dd_cfg, c),
        None => SensitiveDataScannerAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_scanning_groups()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list scanner rules: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn scanner_rules_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/sensitive-data-scanner/config", &[]).await?;
    crate::formatter::output(cfg, &data)
}
