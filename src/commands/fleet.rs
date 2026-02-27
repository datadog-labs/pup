use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_fleet_automation::{
    FleetAutomationAPI, GetFleetDeploymentOptionalParams, ListFleetAgentsOptionalParams,
    ListFleetDeploymentsOptionalParams,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn agents_list(cfg: &Config, page_size: Option<i64>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    let mut params = ListFleetAgentsOptionalParams::default();
    if let Some(ps) = page_size {
        params = params.page_size(ps);
    }
    let resp = api
        .list_fleet_agents(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list fleet agents: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn agents_list(cfg: &Config, page_size: Option<i64>) -> Result<()> {
    let mut query: Vec<(&str, String)> = Vec::new();
    if let Some(ps) = page_size {
        query.push(("page[size]", ps.to_string()));
    }
    let data = crate::api::get(cfg, "/api/v2/fleet/agents", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn agents_get(cfg: &Config, agent_key: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .get_fleet_agent_info(agent_key.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get fleet agent: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn agents_get(cfg: &Config, agent_key: &str) -> Result<()> {
    let path = format!("/api/v2/fleet/agents/{agent_key}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn agents_versions(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_fleet_agent_versions()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list agent versions: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn agents_versions(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/fleet/agents/versions", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn deployments_list(cfg: &Config, page_size: Option<i64>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    let mut params = ListFleetDeploymentsOptionalParams::default();
    if let Some(ps) = page_size {
        params = params.page_size(ps);
    }
    let resp = api
        .list_fleet_deployments(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list deployments: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn deployments_list(cfg: &Config, page_size: Option<i64>) -> Result<()> {
    let mut query: Vec<(&str, String)> = Vec::new();
    if let Some(ps) = page_size {
        query.push(("page[size]", ps.to_string()));
    }
    let data = crate::api::get(cfg, "/api/v2/fleet/deployments", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn deployments_get(cfg: &Config, deployment_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .get_fleet_deployment(
            deployment_id.to_string(),
            GetFleetDeploymentOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to get deployment: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn deployments_get(cfg: &Config, deployment_id: &str) -> Result<()> {
    let path = format!("/api/v2/fleet/deployments/{deployment_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn schedules_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_fleet_schedules()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list schedules: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn schedules_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/fleet/schedules", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn schedules_get(cfg: &Config, schedule_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .get_fleet_schedule(schedule_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get schedule: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn schedules_get(cfg: &Config, schedule_id: &str) -> Result<()> {
    let path = format!("/api/v2/fleet/schedules/{schedule_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn schedules_update(cfg: &Config, schedule_id: &str, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    let body = util::read_json_file(file)?;
    let resp = api
        .update_fleet_schedule(schedule_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update schedule: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn schedules_update(cfg: &Config, schedule_id: &str, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let path = format!("/api/v2/fleet/schedules/{schedule_id}");
    let data = crate::api::patch(cfg, &path, &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn schedules_delete(cfg: &Config, schedule_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    api.delete_fleet_schedule(schedule_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete schedule: {e:?}"))?;
    println!("Schedule '{schedule_id}' deleted successfully.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn schedules_delete(cfg: &Config, schedule_id: &str) -> Result<()> {
    let path = format!("/api/v2/fleet/schedules/{schedule_id}");
    crate::api::delete(cfg, &path).await?;
    println!("Schedule '{schedule_id}' deleted successfully.");
    Ok(())
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn deployments_cancel(cfg: &Config, deployment_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    api.cancel_fleet_deployment(deployment_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to cancel deployment: {e:?}"))?;
    println!("Fleet deployment {deployment_id} cancelled.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn deployments_cancel(cfg: &Config, deployment_id: &str) -> Result<()> {
    let path = format!("/api/v2/fleet/deployments/{deployment_id}");
    crate::api::delete(cfg, &path).await?;
    println!("Fleet deployment {deployment_id} cancelled.");
    Ok(())
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn deployments_configure(cfg: &Config, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    let body = util::read_json_file(file)?;
    let resp = api
        .create_fleet_deployment_configure(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to configure deployment: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn deployments_configure(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/fleet/deployments/configure", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn deployments_upgrade(cfg: &Config, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    let body = util::read_json_file(file)?;
    let resp = api
        .create_fleet_deployment_upgrade(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to upgrade deployment: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn deployments_upgrade(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/fleet/deployments/upgrade", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn schedules_create(cfg: &Config, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    let body = util::read_json_file(file)?;
    let resp = api
        .create_fleet_schedule(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create schedule: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn schedules_create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/fleet/schedules", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn schedules_trigger(cfg: &Config, schedule_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = FleetAutomationAPI::with_client_and_config(dd_cfg, dd_client);
    api.trigger_fleet_schedule(schedule_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to trigger schedule: {e:?}"))?;
    println!("Schedule {schedule_id} triggered.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn schedules_trigger(cfg: &Config, schedule_id: &str) -> Result<()> {
    let path = format!("/api/v2/fleet/schedules/{schedule_id}/trigger");
    let body = serde_json::json!({});
    crate::api::post(cfg, &path, &body).await?;
    println!("Schedule {schedule_id} triggered.");
    Ok(())
}
