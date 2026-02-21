use anyhow::{bail, Result};

use crate::config::Config;
use crate::formatter;

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

    let resp = req.header("Accept", "application/json").send().await?;
    if !resp.status().is_success() {
        let status = resp.status();
        let body = resp.text().await.unwrap_or_default();
        bail!("API error (HTTP {status}): {body}");
    }
    Ok(resp.json().await?)
}

pub async fn list(cfg: &Config, page_limit: i64, page_offset: i64) -> Result<()> {
    let path = format!(
        "/api/v2/bits-ai/investigations?page[limit]={page_limit}&page[offset]={page_offset}"
    );
    let data = raw_get(cfg, &path).await?;
    formatter::print_json(&data)
}

pub async fn get(cfg: &Config, investigation_id: &str) -> Result<()> {
    let path = format!("/api/v2/bits-ai/investigations/{investigation_id}");
    let data = raw_get(cfg, &path).await?;
    formatter::print_json(&data)
}
