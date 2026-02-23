use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_authentication::AuthenticationAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_ip_ranges::IPRangesAPI;

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

#[cfg(not(target_arch = "wasm32"))]
pub async fn ip_ranges(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => IPRangesAPI::with_client_and_config(dd_cfg, c),
        None => IPRangesAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_ip_ranges()
        .await
        .map_err(|e| anyhow::anyhow!("failed to get IP ranges: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn ip_ranges(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v1/ip_ranges", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn status(cfg: &Config) -> Result<()> {
    if !cfg.has_bearer_token() && !cfg.has_api_keys() {
        let transformed = serde_json::json!({
            "message": "no credentials configured",
            "status": "unauthenticated"
        });
        return formatter::output(cfg, &transformed);
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => AuthenticationAPI::with_client_and_config(dd_cfg, c),
        None => AuthenticationAPI::with_config(dd_cfg),
    };
    let _resp = api
        .validate()
        .await
        .map_err(|e| anyhow::anyhow!("failed to validate API keys: {e:?}"))?;
    let transformed = serde_json::json!({
        "message": "API is operational",
        "status": "ok"
    });
    formatter::output(cfg, &transformed)
}

#[cfg(target_arch = "wasm32")]
pub async fn status(cfg: &Config) -> Result<()> {
    let _data = crate::api::get(cfg, "/api/v1/validate", &[]).await?;
    let transformed = serde_json::json!({
        "message": "API is operational",
        "status": "ok"
    });
    crate::formatter::output(cfg, &transformed)
}
