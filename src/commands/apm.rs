use anyhow::{bail, Result};

use crate::config::Config;
use crate::formatter;
use crate::util;

/// Make an authenticated GET request to the DD API using reqwest directly.
/// APM commands use raw HTTP because they hit internal/unstable endpoints
/// not covered by the typed DD API client.
#[cfg(not(target_arch = "wasm32"))]
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
        bail!("APM API error (HTTP {status}): {body}");
    }

    Ok(resp.json().await?)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn services_list(cfg: &Config, env: String, from: String, to: String) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let path = format!("/api/v2/apm/services?start={from_ts}&end={to_ts}&filter[env]={env}");
    let data = raw_get(cfg, &path).await?;
    formatter::output(cfg, &data)
}

#[cfg(target_arch = "wasm32")]
pub async fn services_list(cfg: &Config, env: String, from: String, to: String) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let query = vec![
        ("start", from_ts.to_string()),
        ("end", to_ts.to_string()),
        ("filter[env]", env),
    ];
    let data = crate::api::get(cfg, "/api/v2/apm/services", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn services_stats(cfg: &Config, env: String, from: String, to: String) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let path = format!("/api/v2/apm/services/stats?start={from_ts}&end={to_ts}&filter[env]={env}");
    let data = raw_get(cfg, &path).await?;
    formatter::output(cfg, &data)
}

#[cfg(target_arch = "wasm32")]
pub async fn services_stats(cfg: &Config, env: String, from: String, to: String) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let query = vec![
        ("start", from_ts.to_string()),
        ("end", to_ts.to_string()),
        ("filter[env]", env),
    ];
    let data = crate::api::get(cfg, "/api/v2/apm/services/stats", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn entities_list(cfg: &Config, from: String, to: String) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let path = format!("/api/unstable/apm/entities?start={from_ts}&end={to_ts}");
    let data = raw_get(cfg, &path).await?;
    formatter::output(cfg, &data)
}

#[cfg(target_arch = "wasm32")]
pub async fn entities_list(cfg: &Config, from: String, to: String) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let query = vec![("start", from_ts.to_string()), ("end", to_ts.to_string())];
    let data = crate::api::get(cfg, "/api/unstable/apm/entities", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn dependencies_list(cfg: &Config, env: String, from: String, to: String) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let path = format!("/api/v1/service_dependencies?start={from_ts}&end={to_ts}&env={env}");
    let data = raw_get(cfg, &path).await?;
    formatter::output(cfg, &data)
}

#[cfg(target_arch = "wasm32")]
pub async fn dependencies_list(cfg: &Config, env: String, from: String, to: String) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let query = vec![
        ("start", from_ts.to_string()),
        ("end", to_ts.to_string()),
        ("env", env),
    ];
    let data = crate::api::get(cfg, "/api/v1/service_dependencies", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn services_operations(
    cfg: &Config,
    service: String,
    env: String,
    from: String,
    to: String,
) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let path =
        format!("/api/v1/trace/operation_names/{service}?env={env}&start={from_ts}&end={to_ts}");
    let data = raw_get(cfg, &path).await?;
    formatter::output(cfg, &data)
}

#[cfg(target_arch = "wasm32")]
pub async fn services_operations(
    cfg: &Config,
    service: String,
    env: String,
    from: String,
    to: String,
) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let path = format!("/api/v1/trace/operation_names/{service}");
    let query = vec![
        ("env", env),
        ("start", from_ts.to_string()),
        ("end", to_ts.to_string()),
    ];
    let data = crate::api::get(cfg, &path, &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn services_resources(
    cfg: &Config,
    service: String,
    operation: String,
    env: String,
    from: String,
    to: String,
) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let path = format!(
        "/api/ui/apm/resources?service={service}&operation={operation}&env={env}&start={from_ts}&end={to_ts}"
    );
    let data = raw_get(cfg, &path).await?;
    formatter::output(cfg, &data)
}

#[cfg(target_arch = "wasm32")]
pub async fn services_resources(
    cfg: &Config,
    service: String,
    operation: String,
    env: String,
    from: String,
    to: String,
) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let query = vec![
        ("service", service),
        ("operation", operation),
        ("env", env),
        ("start", from_ts.to_string()),
        ("end", to_ts.to_string()),
    ];
    let data = crate::api::get(cfg, "/api/ui/apm/resources", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn flow_map(
    cfg: &Config,
    query: String,
    limit: i64,
    from: String,
    to: String,
) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let path =
        format!("/api/ui/apm/flow-map?query={query}&limit={limit}&start={from_ts}&end={to_ts}");
    let data = raw_get(cfg, &path).await?;
    formatter::output(cfg, &data)
}

#[cfg(target_arch = "wasm32")]
pub async fn flow_map(
    cfg: &Config,
    query: String,
    limit: i64,
    from: String,
    to: String,
) -> Result<()> {
    let from_ts = util::parse_time_to_unix(&from)?;
    let to_ts = util::parse_time_to_unix(&to)?;
    let q = vec![
        ("query", query),
        ("limit", limit.to_string()),
        ("start", from_ts.to_string()),
        ("end", to_ts.to_string()),
    ];
    let data = crate::api::get(cfg, "/api/ui/apm/flow-map", &q).await?;
    crate::formatter::output(cfg, &data)
}
