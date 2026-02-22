use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_tags::{
    CreateHostTagsOptionalParams, DeleteHostTagsOptionalParams, GetHostTagsOptionalParams,
    ListHostTagsOptionalParams, TagsAPI, UpdateHostTagsOptionalParams,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::model::HostTags;

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TagsAPI::with_client_and_config(dd_cfg, c),
        None => TagsAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_host_tags(ListHostTagsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list tags: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v1/tags/hosts", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn get(cfg: &Config, hostname: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TagsAPI::with_client_and_config(dd_cfg, c),
        None => TagsAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_host_tags(hostname.to_string(), GetHostTagsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get tags: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn get(cfg: &Config, hostname: &str) -> Result<()> {
    let data = crate::api::get(cfg, &format!("/api/v1/tags/hosts/{hostname}"), &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn add(cfg: &Config, hostname: &str, tags: Vec<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TagsAPI::with_client_and_config(dd_cfg, c),
        None => TagsAPI::with_config(dd_cfg),
    };
    let body = HostTags::new().tags(tags);
    let resp = api
        .create_host_tags(
            hostname.to_string(),
            body,
            CreateHostTagsOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to add tags: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn add(cfg: &Config, hostname: &str, tags: Vec<String>) -> Result<()> {
    let body = serde_json::json!({ "tags": tags });
    let data = crate::api::post(cfg, &format!("/api/v1/tags/hosts/{hostname}"), &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn update(cfg: &Config, hostname: &str, tags: Vec<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TagsAPI::with_client_and_config(dd_cfg, c),
        None => TagsAPI::with_config(dd_cfg),
    };
    let body = HostTags::new().tags(tags);
    let resp = api
        .update_host_tags(
            hostname.to_string(),
            body,
            UpdateHostTagsOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to update tags: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn update(cfg: &Config, hostname: &str, tags: Vec<String>) -> Result<()> {
    let body = serde_json::json!({ "tags": tags });
    let data = crate::api::put(cfg, &format!("/api/v1/tags/hosts/{hostname}"), &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn delete(cfg: &Config, hostname: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TagsAPI::with_client_and_config(dd_cfg, c),
        None => TagsAPI::with_config(dd_cfg),
    };
    api.delete_host_tags(
        hostname.to_string(),
        DeleteHostTagsOptionalParams::default(),
    )
    .await
    .map_err(|e| anyhow::anyhow!("failed to delete tags: {e:?}"))?;
    println!("Successfully deleted all tags from host {hostname}");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn delete(cfg: &Config, hostname: &str) -> Result<()> {
    crate::api::delete(cfg, &format!("/api/v1/tags/hosts/{hostname}")).await?;
    println!("Successfully deleted all tags from host {hostname}");
    Ok(())
}
