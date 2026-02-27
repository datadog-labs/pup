use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_service_level_objectives::{
    DeleteSLOOptionalParams, GetSLOOptionalParams, ListSLOsOptionalParams,
    ServiceLevelObjectivesAPI,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::model::{ServiceLevelObjective, ServiceLevelObjectiveRequest};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceLevelObjectivesAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_slos(ListSLOsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list SLOs: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v1/slo", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn get(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceLevelObjectivesAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .get_slo(id.to_string(), GetSLOOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get SLO: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn get(cfg: &Config, id: &str) -> Result<()> {
    let data = crate::api::get(cfg, &format!("/api/v1/slo/{id}"), &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn create(cfg: &Config, file: &str) -> Result<()> {
    let body: ServiceLevelObjectiveRequest = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceLevelObjectivesAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .create_slo(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create SLO: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v1/slo", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn update(cfg: &Config, id: &str, file: &str) -> Result<()> {
    let body: ServiceLevelObjective = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceLevelObjectivesAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .update_slo(id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update SLO: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn update(cfg: &Config, id: &str, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::put(cfg, &format!("/api/v1/slo/{id}"), &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn delete(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceLevelObjectivesAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .delete_slo(id.to_string(), DeleteSLOOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete SLO: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn delete(cfg: &Config, id: &str) -> Result<()> {
    let data = crate::api::delete(cfg, &format!("/api/v1/slo/{id}")).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn status(cfg: &Config, id: &str, from_ts: i64, to_ts: i64) -> Result<()> {
    use datadog_api_client::datadogV2::api_service_level_objectives::{
        GetSloStatusOptionalParams, ServiceLevelObjectivesAPI as SloV2API,
    };

    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = SloV2API::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .get_slo_status(
            id.to_string(),
            from_ts,
            to_ts,
            GetSloStatusOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to get SLO status: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn status(cfg: &Config, id: &str, from_ts: i64, to_ts: i64) -> Result<()> {
    let query = vec![
        ("from_ts", from_ts.to_string()),
        ("to_ts", to_ts.to_string()),
    ];
    let data = crate::api::get(cfg, &format!("/api/v2/slo/{id}/status"), &query).await?;
    crate::formatter::output(cfg, &data)
}
