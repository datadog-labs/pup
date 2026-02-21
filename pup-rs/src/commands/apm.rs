use anyhow::{bail, Result};

use crate::config::Config;
use crate::formatter;
use crate::util;

/// Make an authenticated GET request to the DD API using reqwest directly.
/// APM commands use raw HTTP because they hit internal/unstable endpoints
/// not covered by the typed DD API client.
async fn raw_get(cfg: &Config, path: &str) -> Result<serde_json::Value> {
    let url = format!("https://{}{}", cfg.api_host(), path);
    let client = reqwest::Client::new();
    let mut req = client.get(&url);

    if let Some(token) = &cfg.access_token {
        req = req.header("Authorization", format!("Bearer {token}"));
    } else if let (Some(api_key), Some(app_key)) = (&cfg.api_key, &cfg.app_key) {
        req = req
            .header("DD-API-KEY", api_key.as_str())
            .header("DD-APPLICATION-KEY", app_key.as_str());
    } else {
        bail!("no authentication configured");
    }

    let resp = req
        .header("Accept", "application/json")
        .send()
        .await?;

    if !resp.status().is_success() {
        let status = resp.status();
        let body = resp.text().await.unwrap_or_default();
        bail!("APM API error (HTTP {status}): {body}");
    }

    Ok(resp.json().await?)
}

pub async fn services_list(cfg: &Config, env: String, from: String, to: String) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let path = format!(
        "/api/v2/apm/services?start={from_ts}&end={to_ts}&filter[env]={env}"
    );
    let data = raw_get(cfg, &path).await?;
    formatter::print_json(&data)
}

pub async fn services_stats(
    cfg: &Config,
    env: String,
    from: String,
    to: String,
) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let path = format!(
        "/api/v2/apm/services/stats?start={from_ts}&end={to_ts}&filter[env]={env}"
    );
    let data = raw_get(cfg, &path).await?;
    formatter::print_json(&data)
}
