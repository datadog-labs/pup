use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_service_definition::{
    GetServiceDefinitionOptionalParams, ListServiceDefinitionsOptionalParams, ServiceDefinitionAPI,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ServiceDefinitionAPI::with_client_and_config(dd_cfg, c),
        None => ServiceDefinitionAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_service_definitions(ListServiceDefinitionsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list services: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/services/definitions", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn get(cfg: &Config, service_name: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ServiceDefinitionAPI::with_client_and_config(dd_cfg, c),
        None => ServiceDefinitionAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_service_definition(
            service_name.to_string(),
            GetServiceDefinitionOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to get service: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn get(cfg: &Config, service_name: &str) -> Result<()> {
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/services/definitions/{service_name}"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}
