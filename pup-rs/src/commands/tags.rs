use anyhow::Result;
use datadog_api_client::datadogV1::api_tags::{
    TagsAPI, ListHostTagsOptionalParams, GetHostTagsOptionalParams,
    CreateHostTagsOptionalParams, UpdateHostTagsOptionalParams, DeleteHostTagsOptionalParams,
};
use datadog_api_client::datadogV1::model::HostTags;

use crate::client;
use crate::config::Config;
use crate::formatter;

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
    formatter::print_json(&resp)
}

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
    formatter::print_json(&resp)
}

pub async fn add(cfg: &Config, hostname: &str, tags: Vec<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TagsAPI::with_client_and_config(dd_cfg, c),
        None => TagsAPI::with_config(dd_cfg),
    };
    let body = HostTags::new().tags(tags);
    let resp = api
        .create_host_tags(hostname.to_string(), body, CreateHostTagsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to add tags: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn update(cfg: &Config, hostname: &str, tags: Vec<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TagsAPI::with_client_and_config(dd_cfg, c),
        None => TagsAPI::with_config(dd_cfg),
    };
    let body = HostTags::new().tags(tags);
    let resp = api
        .update_host_tags(hostname.to_string(), body, UpdateHostTagsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to update tags: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn delete(cfg: &Config, hostname: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => TagsAPI::with_client_and_config(dd_cfg, c),
        None => TagsAPI::with_config(dd_cfg),
    };
    api.delete_host_tags(hostname.to_string(), DeleteHostTagsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete tags: {e:?}"))?;
    eprintln!("Tags deleted for host {hostname}.");
    Ok(())
}
