use anyhow::Result;
use datadog_api_client::datadogV1::api_authentication::AuthenticationAPI;
use datadog_api_client::datadogV1::api_ip_ranges::IPRangesAPI;

use crate::client;
use crate::config::Config;
use crate::formatter;

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

pub async fn status(cfg: &Config) -> Result<()> {
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
