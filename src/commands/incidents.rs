use anyhow::{bail, Result};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_incidents::{
    CreateGlobalIncidentHandleOptionalParams, GetIncidentOptionalParams, IncidentsAPI,
    ListGlobalIncidentHandlesOptionalParams, ListIncidentAttachmentsOptionalParams,
    ListIncidentsOptionalParams, UpdateGlobalIncidentHandleOptionalParams,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

// ---------------------------------------------------------------------------
// Helper: build an IncidentsAPI with bearer-token support
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
fn make_api(cfg: &Config) -> IncidentsAPI {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    IncidentsAPI::with_client_and_config(dd_cfg, dd_client)
}

// ---------------------------------------------------------------------------
// Core incident operations
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(cfg: &Config, limit: i64) -> Result<()> {
    let api = make_api(cfg);
    let params = ListIncidentsOptionalParams::default().page_size(limit);
    let resp = api
        .list_incidents(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list incidents: {:?}", e))?;
    formatter::output(cfg, &resp)?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn list(cfg: &Config, limit: i64) -> Result<()> {
    let query_params = vec![("page[size]", limit.to_string())];
    let data = crate::api::get(cfg, "/api/v2/incidents", &query_params).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn get(cfg: &Config, incident_id: &str) -> Result<()> {
    let api = make_api(cfg);
    let resp = api
        .get_incident(
            incident_id.to_string(),
            GetIncidentOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to get incident: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn get(cfg: &Config, incident_id: &str) -> Result<()> {
    let path = format!("/api/v2/incidents/{incident_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Attachments
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn attachments_list(cfg: &Config, incident_id: &str) -> Result<()> {
    let api = make_api(cfg);
    let resp = api
        .list_incident_attachments(
            incident_id.to_string(),
            ListIncidentAttachmentsOptionalParams::default(),
        )
        .await
        .map_err(|e| anyhow::anyhow!("failed to list incident attachments: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn attachments_list(cfg: &Config, incident_id: &str) -> Result<()> {
    let path = format!("/api/v2/incidents/{incident_id}/attachments");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn attachments_delete(
    cfg: &Config,
    incident_id: &str,
    attachment_id: &str,
) -> Result<()> {
    let url = format!(
        "{}/api/v2/incidents/{}/attachments/{}",
        cfg.api_base_url(),
        incident_id,
        attachment_id
    );
    let client = reqwest::Client::new();
    let mut req = client.delete(&url);

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
        bail!("failed to delete incident attachment (HTTP {status}): {body}");
    }
    println!("Incident attachment {attachment_id} deleted from incident {incident_id}.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn attachments_delete(
    cfg: &Config,
    incident_id: &str,
    attachment_id: &str,
) -> Result<()> {
    let path = format!("/api/v2/incidents/{incident_id}/attachments/{attachment_id}");
    crate::api::delete(cfg, &path).await?;
    println!("Incident attachment {attachment_id} deleted from incident {incident_id}.");
    Ok(())
}

// ---------------------------------------------------------------------------
// Global incident settings
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn settings_get(cfg: &Config) -> Result<()> {
    let api = make_api(cfg);
    let resp = api
        .get_global_incident_settings()
        .await
        .map_err(|e| anyhow::anyhow!("failed to get incident settings: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn settings_get(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/incidents/config/settings", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn settings_update(cfg: &Config, file: &str) -> Result<()> {
    let body = util::read_json_file(file)?;
    let api = make_api(cfg);
    let resp = api
        .update_global_incident_settings(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update incident settings: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn settings_update(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::put(cfg, "/api/v2/incidents/config/settings", &body).await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Global incident handles
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn handles_list(cfg: &Config) -> Result<()> {
    let api = make_api(cfg);
    let params = ListGlobalIncidentHandlesOptionalParams::default();
    let resp = api
        .list_global_incident_handles(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list incident handles: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn handles_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/incidents/config/handles", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn handles_create(cfg: &Config, file: &str) -> Result<()> {
    let body = util::read_json_file(file)?;
    let api = make_api(cfg);
    let resp = api
        .create_global_incident_handle(body, CreateGlobalIncidentHandleOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to create incident handle: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn handles_create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/incidents/config/handles", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn handles_update(cfg: &Config, file: &str) -> Result<()> {
    let body = util::read_json_file(file)?;
    let api = make_api(cfg);
    let resp = api
        .update_global_incident_handle(body, UpdateGlobalIncidentHandleOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to update incident handle: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn handles_update(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::patch(cfg, "/api/v2/incidents/config/handles", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn handles_delete(cfg: &Config, _handle_id: &str) -> Result<()> {
    let api = make_api(cfg);
    api.delete_global_incident_handle()
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete incident handle: {:?}", e))?;
    println!("Incident handle deleted.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn handles_delete(cfg: &Config, _handle_id: &str) -> Result<()> {
    crate::api::delete(cfg, "/api/v2/incidents/config/handles").await?;
    println!("Incident handle deleted.");
    Ok(())
}

// ---------------------------------------------------------------------------
// Postmortem templates
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn postmortem_templates_list(cfg: &Config) -> Result<()> {
    let api = make_api(cfg);
    let resp = api
        .list_incident_postmortem_templates()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list postmortem templates: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn postmortem_templates_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/incidents/config/postmortem-templates", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn postmortem_templates_get(cfg: &Config, template_id: &str) -> Result<()> {
    let api = make_api(cfg);
    let resp = api
        .get_incident_postmortem_template(template_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get postmortem template: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn postmortem_templates_get(cfg: &Config, template_id: &str) -> Result<()> {
    let path = format!("/api/v2/incidents/config/postmortem-templates/{template_id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn postmortem_templates_create(cfg: &Config, file: &str) -> Result<()> {
    let body = util::read_json_file(file)?;
    let api = make_api(cfg);
    let resp = api
        .create_incident_postmortem_template(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create postmortem template: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn postmortem_templates_create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data =
        crate::api::post(cfg, "/api/v2/incidents/config/postmortem-templates", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn postmortem_templates_update(
    cfg: &Config,
    template_id: &str,
    file: &str,
) -> Result<()> {
    let body = util::read_json_file(file)?;
    let api = make_api(cfg);
    let resp = api
        .update_incident_postmortem_template(template_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update postmortem template: {:?}", e))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn postmortem_templates_update(
    cfg: &Config,
    template_id: &str,
    file: &str,
) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let path = format!("/api/v2/incidents/config/postmortem-templates/{template_id}");
    let data = crate::api::patch(cfg, &path, &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn postmortem_templates_delete(cfg: &Config, template_id: &str) -> Result<()> {
    let api = make_api(cfg);
    api.delete_incident_postmortem_template(template_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete postmortem template: {:?}", e))?;
    println!("Postmortem template {template_id} deleted.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn postmortem_templates_delete(cfg: &Config, template_id: &str) -> Result<()> {
    let path = format!("/api/v2/incidents/config/postmortem-templates/{template_id}");
    crate::api::delete(cfg, &path).await?;
    println!("Postmortem template {template_id} deleted.");
    Ok(())
}
