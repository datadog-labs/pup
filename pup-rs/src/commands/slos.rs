use anyhow::Result;
use datadog_api_client::datadogV1::api_service_level_objectives::{
    ServiceLevelObjectivesAPI, ListSLOsOptionalParams, GetSLOOptionalParams, DeleteSLOOptionalParams,
};

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ServiceLevelObjectivesAPI::with_client_and_config(dd_cfg, c),
        None => ServiceLevelObjectivesAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_slos(ListSLOsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list SLOs: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn get(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ServiceLevelObjectivesAPI::with_client_and_config(dd_cfg, c),
        None => ServiceLevelObjectivesAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_slo(id.to_string(), GetSLOOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get SLO: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn delete(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ServiceLevelObjectivesAPI::with_client_and_config(dd_cfg, c),
        None => ServiceLevelObjectivesAPI::with_config(dd_cfg),
    };
    let resp = api
        .delete_slo(id.to_string(), DeleteSLOOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete SLO: {e:?}"))?;
    formatter::print_json(&resp)
}
