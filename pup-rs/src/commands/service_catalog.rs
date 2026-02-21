use anyhow::Result;
use datadog_api_client::datadogV2::api_service_definition::{
    ServiceDefinitionAPI, ListServiceDefinitionsOptionalParams,
    GetServiceDefinitionOptionalParams,
};

use crate::client;
use crate::config::Config;
use crate::formatter;

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
    formatter::print_json(&resp)
}

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
    formatter::print_json(&resp)
}
