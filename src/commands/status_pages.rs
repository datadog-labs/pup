use anyhow::{bail, Result};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_status_pages::{
    CreateComponentOptionalParams, CreateDegradationOptionalParams, CreateStatusPageOptionalParams,
    GetComponentOptionalParams, GetDegradationOptionalParams, GetStatusPageOptionalParams,
    ListComponentsOptionalParams, ListDegradationsOptionalParams, ListStatusPagesOptionalParams,
    StatusPagesAPI, UpdateComponentOptionalParams, UpdateDegradationOptionalParams,
    UpdateStatusPageOptionalParams,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::{
    CreateComponentRequest, CreateDegradationRequest, CreateStatusPageRequest,
    PatchComponentRequest, PatchDegradationRequest, PatchStatusPageRequest,
};
use uuid::Uuid;

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
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
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn pages_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/status_pages", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
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
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn pages_get(cfg: &Config, page_id: &str) -> Result<()> {
    let _uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let data = crate::api::get(cfg, &format!("/api/v2/status_pages/{page_id}"), &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
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
    println!("Status page {page_id} deleted.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn pages_delete(cfg: &Config, page_id: &str) -> Result<()> {
    let _uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    crate::api::delete(cfg, &format!("/api/v2/status_pages/{page_id}")).await?;
    println!("Status page {page_id} deleted.");
    Ok(())
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn pages_update(cfg: &Config, page_id: &str, file: &str) -> Result<()> {
    let page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let body: PatchStatusPageRequest = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let resp = api
        .update_status_page(page_uuid, body, UpdateStatusPageOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to update status page: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn pages_update(cfg: &Config, page_id: &str, file: &str) -> Result<()> {
    let _uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::patch(cfg, &format!("/api/v2/status_pages/{page_id}"), &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
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
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn components_list(cfg: &Config, page_id: &str) -> Result<()> {
    let _uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/status_pages/{page_id}/components"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
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
        .get_component(
            page_uuid,
            component_uuid,
            GetComponentOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to get component: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn components_get(cfg: &Config, page_id: &str, component_id: &str) -> Result<()> {
    let _page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let _component_uuid = Uuid::parse_str(component_id)
        .map_err(|e| anyhow::anyhow!("invalid component UUID '{component_id}': {e}"))?;
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/status_pages/{page_id}/components/{component_id}"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn components_update(
    cfg: &Config,
    page_id: &str,
    component_id: &str,
    file: &str,
) -> Result<()> {
    let page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let component_uuid = Uuid::parse_str(component_id)
        .map_err(|e| anyhow::anyhow!("invalid component UUID '{component_id}': {e}"))?;
    let body: PatchComponentRequest = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let resp = api
        .update_component(
            page_uuid,
            component_uuid,
            body,
            UpdateComponentOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to update component: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn components_update(
    cfg: &Config,
    page_id: &str,
    component_id: &str,
    file: &str,
) -> Result<()> {
    let _page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let _component_uuid = Uuid::parse_str(component_id)
        .map_err(|e| anyhow::anyhow!("invalid component UUID '{component_id}': {e}"))?;
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::patch(
        cfg,
        &format!("/api/v2/status_pages/{page_id}/components/{component_id}"),
        &body,
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
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
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn degradations_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/status_pages/degradations", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
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
        .get_degradation(
            page_uuid,
            degradation_uuid,
            GetDegradationOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to get degradation: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn degradations_get(cfg: &Config, page_id: &str, degradation_id: &str) -> Result<()> {
    let _page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let _degradation_uuid = Uuid::parse_str(degradation_id)
        .map_err(|e| anyhow::anyhow!("invalid degradation UUID '{degradation_id}': {e}"))?;
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/status_pages/{page_id}/degradations/{degradation_id}"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn degradations_create(cfg: &Config, page_id: &str, file: &str) -> Result<()> {
    let page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let body: CreateDegradationRequest = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let resp = api
        .create_degradation(page_uuid, body, CreateDegradationOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to create degradation: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn degradations_create(cfg: &Config, page_id: &str, file: &str) -> Result<()> {
    let _page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(
        cfg,
        &format!("/api/v2/status_pages/{page_id}/degradations"),
        &body,
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn degradations_update(
    cfg: &Config,
    page_id: &str,
    degradation_id: &str,
    file: &str,
) -> Result<()> {
    let page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let degradation_uuid = Uuid::parse_str(degradation_id)
        .map_err(|e| anyhow::anyhow!("invalid degradation UUID '{degradation_id}': {e}"))?;
    let body: PatchDegradationRequest = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let resp = api
        .update_degradation(
            page_uuid,
            degradation_uuid,
            body,
            UpdateDegradationOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to update degradation: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn degradations_update(
    cfg: &Config,
    page_id: &str,
    degradation_id: &str,
    file: &str,
) -> Result<()> {
    let _page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let _degradation_uuid = Uuid::parse_str(degradation_id)
        .map_err(|e| anyhow::anyhow!("invalid degradation UUID '{degradation_id}': {e}"))?;
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::patch(
        cfg,
        &format!("/api/v2/status_pages/{page_id}/degradations/{degradation_id}"),
        &body,
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn components_delete(cfg: &Config, page_id: &str, component_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let component_uuid = Uuid::parse_str(component_id)
        .map_err(|e| anyhow::anyhow!("invalid component UUID '{component_id}': {e}"))?;
    api.delete_component(page_uuid, component_uuid)
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete component: {e:?}"))?;
    println!("Component {component_id} deleted from page {page_id}.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn components_delete(cfg: &Config, page_id: &str, component_id: &str) -> Result<()> {
    let _page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let _component_uuid = Uuid::parse_str(component_id)
        .map_err(|e| anyhow::anyhow!("invalid component UUID '{component_id}': {e}"))?;
    crate::api::delete(
        cfg,
        &format!("/api/v2/status_pages/{page_id}/components/{component_id}"),
    )
    .await?;
    println!("Component {component_id} deleted from page {page_id}.");
    Ok(())
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn degradations_delete(cfg: &Config, page_id: &str, degradation_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let degradation_uuid = Uuid::parse_str(degradation_id)
        .map_err(|e| anyhow::anyhow!("invalid degradation UUID '{degradation_id}': {e}"))?;
    api.delete_degradation(page_uuid, degradation_uuid)
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete degradation: {e:?}"))?;
    println!("Degradation {degradation_id} deleted from page {page_id}.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn degradations_delete(cfg: &Config, page_id: &str, degradation_id: &str) -> Result<()> {
    let _page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let _degradation_uuid = Uuid::parse_str(degradation_id)
        .map_err(|e| anyhow::anyhow!("invalid degradation UUID '{degradation_id}': {e}"))?;
    crate::api::delete(
        cfg,
        &format!("/api/v2/status_pages/{page_id}/degradations/{degradation_id}"),
    )
    .await?;
    println!("Degradation {degradation_id} deleted from page {page_id}.");
    Ok(())
}

// ---------------------------------------------------------------------------
// Pages create
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn pages_create(cfg: &Config, file: &str) -> Result<()> {
    let body: CreateStatusPageRequest = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let resp = api
        .create_status_page(body, CreateStatusPageOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to create status page: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn pages_create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/status_pages", &body).await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Components create
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn components_create(cfg: &Config, page_id: &str, file: &str) -> Result<()> {
    let page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let body: CreateComponentRequest = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => StatusPagesAPI::with_client_and_config(dd_cfg, c),
        None => StatusPagesAPI::with_config(dd_cfg),
    };
    let resp = api
        .create_component(page_uuid, body, CreateComponentOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to create component: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn components_create(cfg: &Config, page_id: &str, file: &str) -> Result<()> {
    let _page_uuid = Uuid::parse_str(page_id)
        .map_err(|e| anyhow::anyhow!("invalid page UUID '{page_id}': {e}"))?;
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(
        cfg,
        &format!("/api/v2/status_pages/{page_id}/components"),
        &body,
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Third-party status pages (fetched from updog.ai, no DD auth needed)
// ---------------------------------------------------------------------------

use serde::Deserialize;

const THIRD_PARTY_OUTAGES_URL: &str = "https://updog.ai/data/third-party-outages.json";
const SPARKLINE_WIDTH: usize = 30;
const DAY_MS: i64 = 86_400_000;

const ANSI_GREEN: &str = "\x1b[32m";
const ANSI_RED: &str = "\x1b[31m";
const ANSI_DIM: &str = "\x1b[2m";
const ANSI_RESET: &str = "\x1b[0m";

#[derive(Deserialize, serde::Serialize, Clone)]
struct ThirdPartyOutagesResponse {
    data: ThirdPartyOutagesData,
}

#[derive(Deserialize, serde::Serialize, Clone)]
struct ThirdPartyOutagesData {
    attributes: ThirdPartyOutagesAttributes,
}

#[derive(Deserialize, serde::Serialize, Clone)]
struct ThirdPartyOutagesAttributes {
    provider_data: Vec<ThirdPartyProvider>,
}

#[derive(Deserialize, serde::Serialize, Clone)]
struct ThirdPartyProvider {
    provider_name: String,
    #[serde(default)]
    provider_service: String,
    display_name: String,
    #[serde(default)]
    outages: Vec<ThirdPartyOutage>,
    monitoring_start_date: i64,
    // Remaining fields kept for JSON/YAML passthrough
    #[serde(skip_serializing_if = "Option::is_none")]
    integration_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    status_url: Option<String>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    monitored_api_patterns: Vec<String>,
}

#[derive(Deserialize, serde::Serialize, Clone)]
struct ThirdPartyOutage {
    start: i64,
    #[serde(default)]
    end: i64,
    status: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    impacted_region: String,
}

fn filter_providers(
    providers: Vec<ThirdPartyProvider>,
    search: Option<&str>,
    active_only: bool,
) -> Vec<ThirdPartyProvider> {
    providers
        .into_iter()
        .filter(|p| {
            if let Some(q) = search {
                let q = q.to_lowercase();
                if !p.provider_name.to_lowercase().contains(&q)
                    && !p.display_name.to_lowercase().contains(&q)
                {
                    return false;
                }
            }
            if active_only {
                let has_active = p.outages.iter().any(|o| o.status != "resolved");
                if !has_active {
                    return false;
                }
            }
            true
        })
        .collect()
}

fn provider_current_status(provider: &ThirdPartyProvider) -> &str {
    for outage in &provider.outages {
        if outage.status != "resolved" {
            return &outage.status;
        }
    }
    "operational"
}

fn build_sparkline(provider: &ThirdPartyProvider, now_ms: i64) -> String {
    let mut s = String::new();
    for i in (0..SPARKLINE_WIDTH).rev() {
        let bucket_start = now_ms - (i as i64 + 1) * DAY_MS;
        let bucket_end = now_ms - i as i64 * DAY_MS;

        if bucket_end <= provider.monitoring_start_date {
            s.push_str(ANSI_DIM);
            s.push('·');
            s.push_str(ANSI_RESET);
            continue;
        }

        let has_outage = provider.outages.iter().any(|o| {
            let outage_end = if o.end == 0 { now_ms } else { o.end };
            o.start < bucket_end && outage_end > bucket_start
        });

        if has_outage {
            s.push_str(ANSI_RED);
        } else {
            s.push_str(ANSI_GREEN);
        }
        s.push('█');
        s.push_str(ANSI_RESET);
    }
    s
}

fn format_third_party_table(providers: &[ThirdPartyProvider]) -> String {
    let now_ms = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap_or_default()
        .as_millis() as i64;

    // Compute visible column widths (ANSI codes must not influence these).
    let wp = providers
        .iter()
        .map(|p| p.provider_name.len())
        .max()
        .unwrap_or(0)
        .max("PROVIDER".len());
    let wd = providers
        .iter()
        .map(|p| p.display_name.len())
        .max()
        .unwrap_or(0)
        .max("DISPLAY NAME".len());
    let ws = providers
        .iter()
        .map(|p| p.provider_service.len())
        .max()
        .unwrap_or(0)
        .max("SERVICE".len());
    let wst = providers
        .iter()
        .map(|p| provider_current_status(p).len())
        .max()
        .unwrap_or(0)
        .max("STATUS".len());
    // UPTIME column is always exactly SPARKLINE_WIDTH visible chars.
    let wu = SPARKLINE_WIDTH;

    let sep = format!(
        "+-{}-+-{}-+-{}-+-{}-+-{}-+",
        "-".repeat(wp),
        "-".repeat(wd),
        "-".repeat(ws),
        "-".repeat(wu),
        "-".repeat(wst),
    );
    let hsep = sep.replace('-', "=");

    let mut s = String::new();
    s.push_str(&sep);
    s.push('\n');
    s.push_str(&format!(
        "| {:<wp$} | {:<wd$} | {:<ws$} | {:<wu$} | {:<wst$} |\n",
        "PROVIDER",
        "DISPLAY NAME",
        "SERVICE",
        "UPTIME",
        "STATUS",
        wp = wp,
        wd = wd,
        ws = ws,
        wu = wu,
        wst = wst,
    ));
    s.push_str(&hsep);
    s.push('\n');

    for p in providers {
        let sparkline = build_sparkline(p, now_ms);
        let status = provider_current_status(p);
        // The sparkline is exactly `wu` visible chars; ANSI bytes are invisible so
        // we cannot use a format-width specifier for it — assemble that column manually.
        s.push_str(&format!(
            "| {:<wp$} | {:<wd$} | {:<ws$} | {sparkline} | {:<wst$} |\n",
            p.provider_name,
            p.display_name,
            p.provider_service,
            status,
            wp = wp,
            wd = wd,
            ws = ws,
            wst = wst,
        ));
    }

    s.push_str(&sep);
    s
}

pub async fn third_party_list(cfg: &Config, search: Option<&str>, active: bool) -> Result<()> {
    let client = reqwest::Client::new();
    let resp = client
        .get(THIRD_PARTY_OUTAGES_URL)
        .header("Accept", "application/json")
        .send()
        .await?;
    if !resp.status().is_success() {
        let status = resp.status();
        bail!("failed to fetch third-party outages from updog.ai (HTTP {status})");
    }
    let parsed: ThirdPartyOutagesResponse = resp.json().await?;
    let providers = filter_providers(parsed.data.attributes.provider_data, search, active);

    if cfg.output_format == crate::config::OutputFormat::Table {
        if providers.is_empty() {
            println!("No results found");
        } else {
            println!("{}", format_third_party_table(&providers));
        }
        return Ok(());
    }

    formatter::output(cfg, &providers)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_provider(
        name: &str,
        display: &str,
        monitoring_start: i64,
        outages: Vec<ThirdPartyOutage>,
    ) -> ThirdPartyProvider {
        ThirdPartyProvider {
            provider_name: name.to_string(),
            display_name: display.to_string(),
            provider_service: String::new(),
            outages,
            monitoring_start_date: monitoring_start,
            integration_id: None,
            status_url: None,
            monitored_api_patterns: vec![],
        }
    }

    fn make_outage(start: i64, end: i64, status: &str) -> ThirdPartyOutage {
        ThirdPartyOutage {
            start,
            end,
            status: status.to_string(),
            impacted_region: String::new(),
        }
    }

    #[test]
    fn test_provider_current_status_operational() {
        let p = make_provider("a", "A", 0, vec![make_outage(1, 2, "resolved")]);
        assert_eq!(provider_current_status(&p), "operational");
    }

    #[test]
    fn test_provider_current_status_active() {
        let p = make_provider("a", "A", 0, vec![make_outage(1, 0, "active")]);
        assert_eq!(provider_current_status(&p), "active");
    }

    #[test]
    fn test_provider_current_status_no_outages() {
        let p = make_provider("a", "A", 0, vec![]);
        assert_eq!(provider_current_status(&p), "operational");
    }

    #[test]
    fn test_filter_no_filter() {
        let providers = vec![
            make_provider("aws-s3", "Amazon S3", 0, vec![]),
            make_provider("stripe", "Stripe", 0, vec![]),
        ];
        let result = filter_providers(providers, None, false);
        assert_eq!(result.len(), 2);
    }

    #[test]
    fn test_filter_by_provider_name() {
        let providers = vec![
            make_provider("aws-s3", "Amazon S3", 0, vec![]),
            make_provider("stripe", "Stripe", 0, vec![]),
        ];
        let result = filter_providers(providers, Some("aws"), false);
        assert_eq!(result.len(), 1);
        assert_eq!(result[0].provider_name, "aws-s3");
    }

    #[test]
    fn test_filter_by_display_name() {
        let providers = vec![
            make_provider("aws-s3", "Amazon S3", 0, vec![]),
            make_provider("stripe", "Stripe", 0, vec![]),
        ];
        let result = filter_providers(providers, Some("amazon"), false);
        assert_eq!(result.len(), 1);
        assert_eq!(result[0].provider_name, "aws-s3");
    }

    #[test]
    fn test_filter_active_only() {
        let providers = vec![
            make_provider("aws-s3", "Amazon S3", 0, vec![make_outage(1, 0, "active")]),
            make_provider("stripe", "Stripe", 0, vec![make_outage(1, 2, "resolved")]),
        ];
        let result = filter_providers(providers, None, true);
        assert_eq!(result.len(), 1);
        assert_eq!(result[0].provider_name, "aws-s3");
    }

    #[test]
    fn test_filter_no_match() {
        let providers = vec![make_provider("aws-s3", "Amazon S3", 0, vec![])];
        let result = filter_providers(providers, Some("nonexistent"), false);
        assert!(result.is_empty());
    }

    #[test]
    fn test_sparkline_all_green() {
        let now_ms = 1_705_276_800_000i64; // 2024-01-15 00:00:00 UTC
        let p = make_provider("test", "Test", now_ms - 60 * DAY_MS, vec![]);
        let sparkline = build_sparkline(&p, now_ms);
        assert!(!sparkline.contains(ANSI_RED));
        assert!(sparkline.contains(ANSI_GREEN));
        assert!(!sparkline.contains('·'));
    }

    #[test]
    fn test_sparkline_with_active_outage() {
        let now_ms = 1_705_276_800_000i64;
        let p = make_provider(
            "test",
            "Test",
            now_ms - 60 * DAY_MS,
            vec![make_outage(now_ms - DAY_MS / 2, 0, "active")],
        );
        let sparkline = build_sparkline(&p, now_ms);
        assert!(sparkline.contains(ANSI_RED));
    }

    #[test]
    fn test_sparkline_dim_dots_before_monitoring() {
        let now_ms = 1_705_276_800_000i64;
        // Monitoring started 10 days ago → first 20 buckets should be dim dots
        let p = make_provider("test", "Test", now_ms - 10 * DAY_MS, vec![]);
        let sparkline = build_sparkline(&p, now_ms);
        let plain: String = {
            let mut out = String::new();
            let mut in_escape = false;
            for c in sparkline.chars() {
                if c == '\x1b' {
                    in_escape = true;
                    continue;
                }
                if in_escape {
                    if c == 'm' {
                        in_escape = false;
                    }
                    continue;
                }
                out.push(c);
            }
            out
        };
        let dot_count = plain.chars().filter(|&c| c == '·').count();
        assert!(dot_count >= 19, "expected ≥19 dim dots, got {dot_count}");
        assert!(sparkline.contains(ANSI_DIM));
    }

    #[test]
    fn test_sparkline_outage_outside_window() {
        let now_ms = 1_705_276_800_000i64;
        // Outage 45 days ago — outside 30-day window
        let p = make_provider(
            "test",
            "Test",
            now_ms - 60 * DAY_MS,
            vec![make_outage(
                now_ms - 45 * DAY_MS,
                now_ms - 44 * DAY_MS,
                "resolved",
            )],
        );
        let sparkline = build_sparkline(&p, now_ms);
        assert!(!sparkline.contains(ANSI_RED));
    }

    #[test]
    fn test_format_third_party_table_headers() {
        let providers = vec![make_provider("aws-s3", "Amazon S3", 0, vec![])];
        let output = format_third_party_table(&providers);
        for header in &["PROVIDER", "DISPLAY NAME", "SERVICE", "UPTIME", "STATUS"] {
            assert!(output.contains(header), "missing header: {header}");
        }
    }

    #[test]
    fn test_format_third_party_table_data() {
        let now_ms = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_millis() as i64;
        let providers = vec![make_provider(
            "aws-s3",
            "Amazon S3",
            now_ms - 60 * DAY_MS,
            vec![make_outage(now_ms - DAY_MS / 2, 0, "active")],
        )];
        let output = format_third_party_table(&providers);
        assert!(output.contains("aws-s3"));
        assert!(output.contains("Amazon S3"));
        assert!(output.contains("active"));
        assert!(output.contains('█'));
    }
}
