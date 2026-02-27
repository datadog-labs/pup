#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;
use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_slack_integration::SlackIntegrationAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_webhooks_integration::WebhooksIntegrationAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_jira_integration::JiraIntegrationAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_service_now_integration::ServiceNowIntegrationAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::{
    JiraIssueTemplateCreateRequest, JiraIssueTemplateUpdateRequest,
    ServiceNowTemplateCreateRequest, ServiceNowTemplateUpdateRequest,
};

// ---- Jira ----

#[cfg(not(target_arch = "wasm32"))]
pub async fn jira_accounts_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = JiraIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_jira_accounts()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list Jira accounts: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn jira_accounts_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/integration/jira/accounts", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn jira_templates_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = JiraIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_jira_issue_templates()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list Jira templates: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn jira_templates_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/integration/jira/issue_templates", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn jira_templates_get(cfg: &Config, template_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = JiraIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let uuid = util::parse_uuid(template_id, "template")?;
    let resp = api
        .get_jira_issue_template(uuid)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get Jira template: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn jira_templates_get(cfg: &Config, template_id: &str) -> Result<()> {
    util::parse_uuid(template_id, "template")?;
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/integration/jira/issue_templates/{template_id}"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn jira_accounts_delete(cfg: &Config, account_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = JiraIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let uuid = util::parse_uuid(account_id, "account")?;
    api.delete_jira_account(uuid)
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete Jira account: {e:?}"))?;
    println!("Jira account {account_id} deleted.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn jira_accounts_delete(cfg: &Config, account_id: &str) -> Result<()> {
    util::parse_uuid(account_id, "account")?;
    crate::api::delete(
        cfg,
        &format!("/api/v2/integration/jira/accounts/{account_id}"),
    )
    .await?;
    println!("Jira account {account_id} deleted.");
    Ok(())
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn jira_templates_create(cfg: &Config, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = JiraIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let body: JiraIssueTemplateCreateRequest = crate::util::read_json_file(file)?;
    let resp = api
        .create_jira_issue_template(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create Jira template: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn jira_templates_create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/integration/jira/issue_templates", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn jira_templates_update(cfg: &Config, template_id: &str, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = JiraIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let uuid = util::parse_uuid(template_id, "template")?;
    let body: JiraIssueTemplateUpdateRequest = crate::util::read_json_file(file)?;
    let resp = api
        .update_jira_issue_template(uuid, body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update Jira template: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn jira_templates_update(cfg: &Config, template_id: &str, file: &str) -> Result<()> {
    util::parse_uuid(template_id, "template")?;
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = crate::api::patch(
        cfg,
        &format!("/api/v2/integration/jira/issue_templates/{template_id}"),
        &body,
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn jira_templates_delete(cfg: &Config, template_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = JiraIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let uuid = util::parse_uuid(template_id, "template")?;
    api.delete_jira_issue_template(uuid)
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete Jira template: {e:?}"))?;
    println!("Jira template {template_id} deleted.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn jira_templates_delete(cfg: &Config, template_id: &str) -> Result<()> {
    util::parse_uuid(template_id, "template")?;
    crate::api::delete(
        cfg,
        &format!("/api/v2/integration/jira/issue_templates/{template_id}"),
    )
    .await?;
    println!("Jira template {template_id} deleted.");
    Ok(())
}

// ---- ServiceNow ----

#[cfg(not(target_arch = "wasm32"))]
pub async fn servicenow_instances_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceNowIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_service_now_instances()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list ServiceNow instances: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn servicenow_instances_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/integration/servicenow/instances", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn servicenow_templates_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceNowIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_service_now_templates()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list ServiceNow templates: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn servicenow_templates_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/integration/servicenow/templates", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn servicenow_templates_get(cfg: &Config, template_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceNowIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let uuid = util::parse_uuid(template_id, "template")?;
    let resp = api
        .get_service_now_template(uuid)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get ServiceNow template: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn servicenow_templates_get(cfg: &Config, template_id: &str) -> Result<()> {
    util::parse_uuid(template_id, "template")?;
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/integration/servicenow/templates/{template_id}"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn servicenow_templates_create(cfg: &Config, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceNowIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let body: ServiceNowTemplateCreateRequest = crate::util::read_json_file(file)?;
    let resp = api
        .create_service_now_template(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create ServiceNow template: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn servicenow_templates_create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/integration/servicenow/templates", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn servicenow_templates_update(
    cfg: &Config,
    template_id: &str,
    file: &str,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceNowIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let uuid = util::parse_uuid(template_id, "template")?;
    let body: ServiceNowTemplateUpdateRequest = crate::util::read_json_file(file)?;
    let resp = api
        .update_service_now_template(uuid, body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update ServiceNow template: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn servicenow_templates_update(
    cfg: &Config,
    template_id: &str,
    file: &str,
) -> Result<()> {
    util::parse_uuid(template_id, "template")?;
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = crate::api::patch(
        cfg,
        &format!("/api/v2/integration/servicenow/templates/{template_id}"),
        &body,
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn servicenow_templates_delete(cfg: &Config, template_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceNowIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let uuid = util::parse_uuid(template_id, "template")?;
    api.delete_service_now_template(uuid)
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete ServiceNow template: {e:?}"))?;
    println!("ServiceNow template {template_id} deleted.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn servicenow_templates_delete(cfg: &Config, template_id: &str) -> Result<()> {
    util::parse_uuid(template_id, "template")?;
    crate::api::delete(
        cfg,
        &format!("/api/v2/integration/servicenow/templates/{template_id}"),
    )
    .await?;
    println!("ServiceNow template {template_id} deleted.");
    Ok(())
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn servicenow_users_list(cfg: &Config, instance_name: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceNowIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_service_now_users(util::parse_uuid(instance_name, "instance")?)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list ServiceNow users: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn servicenow_users_list(cfg: &Config, instance_name: &str) -> Result<()> {
    util::parse_uuid(instance_name, "instance")?;
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/integration/servicenow/instances/{instance_name}/users"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn servicenow_assignment_groups_list(cfg: &Config, instance_name: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceNowIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_service_now_assignment_groups(util::parse_uuid(instance_name, "instance")?)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list ServiceNow assignment groups: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn servicenow_assignment_groups_list(cfg: &Config, instance_name: &str) -> Result<()> {
    util::parse_uuid(instance_name, "instance")?;
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/integration/servicenow/instances/{instance_name}/assignment_groups"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn servicenow_business_services_list(cfg: &Config, instance_name: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = ServiceNowIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_service_now_business_services(util::parse_uuid(instance_name, "instance")?)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list ServiceNow business services: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn servicenow_business_services_list(cfg: &Config, instance_name: &str) -> Result<()> {
    util::parse_uuid(instance_name, "instance")?;
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/integration/servicenow/instances/{instance_name}/business_services"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

// ---- Slack ----

#[cfg(not(target_arch = "wasm32"))]
pub async fn slack_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = SlackIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .get_slack_integration_channels("main".to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list Slack channels: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn slack_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(
        cfg,
        "/api/v1/integration/slack/configuration/accounts/main/channels",
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

// ---- PagerDuty ----

pub async fn pagerduty_list(_cfg: &Config) -> Result<()> {
    anyhow::bail!(
        "listing PagerDuty services is not supported by the current API version \
         - use 'get' with a specific service name instead"
    )
}

// ---- Webhooks ----

#[cfg(not(target_arch = "wasm32"))]
pub async fn webhooks_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = WebhooksIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .get_webhooks_integration("main".to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list webhooks: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn webhooks_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(
        cfg,
        "/api/v1/integration/webhooks/configuration/webhooks/main",
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}
