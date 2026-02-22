use anyhow::{bail, Result};

use crate::config::Config;
use crate::formatter;

async fn raw_get(cfg: &Config, path: &str) -> Result<serde_json::Value> {
    let url = format!("{}{}", cfg.api_base_url(), path);
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

async fn raw_post(cfg: &Config, path: &str, body: serde_json::Value) -> Result<serde_json::Value> {
    let url = format!("{}{}", cfg.api_base_url(), path);
    let client = reqwest::Client::new();
    let mut req = client.post(&url);

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
        .header("Content-Type", "application/json")
        .header("Accept", "application/json")
        .json(&body)
        .send()
        .await?;
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
    formatter::output(cfg, &data)
}

pub async fn get(cfg: &Config, investigation_id: &str) -> Result<()> {
    let path = format!("/api/v2/bits-ai/investigations/{investigation_id}");
    let data = raw_get(cfg, &path).await?;
    formatter::output(cfg, &data)
}

pub async fn trigger(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = raw_post(cfg, "/api/v2/bits-ai/investigations", body).await?;
    formatter::output(cfg, &data)
}
