use anyhow::Result;
use datadog_api_client::datadogV2::api_status_pages::{
    GetComponentOptionalParams, GetDegradationOptionalParams, GetStatusPageOptionalParams,
    ListComponentsOptionalParams, ListDegradationsOptionalParams, ListStatusPagesOptionalParams,
    StatusPagesAPI,
};
use uuid::Uuid;

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn pages_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_status_pages(ListStatusPagesOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list status pages: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn pages_get(cfg: &Config, page_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let resp = api
        .get_status_page(uuid, GetStatusPageOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get status page: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn pages_delete(cfg: &Config, page_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    api.delete_status_page(uuid)
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete status page: {e:?}"))?;
    eprintln!("Status page {page_id} deleted.");
    Ok(())
}

pub async fn components_list(cfg: &Config, page_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let resp = api
        .list_components(uuid, ListComponentsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list components: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn components_get(cfg: &Config, page_id: &str, component_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let component_uuid = Uuid::parse_str(component_id)
        .map_err(|e| anyhow::anyhow!("invalid component UUID '{component_id}': {e}"))?;
    let resp = api
        .get_component(page_uuid, component_uuid, GetComponentOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get component: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn degradations_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_degradations(ListDegradationsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list degradations: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn degradations_get(cfg: &Config, page_id: &str, degradation_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let degradation_uuid = Uuid::parse_str(degradation_id)
        .map_err(|e| anyhow::anyhow!("invalid degradation UUID '{degradation_id}': {e}"))?;
    let resp = api
        .get_degradation(page_uuid, degradation_uuid, GetDegradationOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get degradation: {e:?}"))?;
    formatter::print_json(&resp)
}
