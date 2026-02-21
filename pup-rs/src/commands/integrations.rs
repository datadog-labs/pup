use anyhow::Result;
use datadog_api_client::datadogV2::api_jira_integration::JiraIntegrationAPI;
use datadog_api_client::datadogV2::api_service_now_integration::ServiceNowIntegrationAPI;
use uuid::Uuid;

use crate::client;
use crate::config::Config;
use crate::formatter;

// ---- Jira ----

pub async fn jira_accounts_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => JiraIntegrationAPI::with_client_and_config(dd_cfg, c),
        None => JiraIntegrationAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_jira_accounts()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list Jira accounts: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn jira_templates_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => JiraIntegrationAPI::with_client_and_config(dd_cfg, c),
        None => JiraIntegrationAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_jira_issue_templates()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list Jira templates: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn jira_templates_get(cfg: &Config, template_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => JiraIntegrationAPI::with_client_and_config(dd_cfg, c),
        None => JiraIntegrationAPI::with_config(dd_cfg),
    };
    let uuid = Uuid::parse_str(template_id)
        .map_err(|e| anyhow::anyhow!("invalid template UUID '{template_id}': {e}"))?;
    let resp = api
        .get_jira_issue_template(uuid)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get Jira template: {e:?}"))?;
    formatter::print_json(&resp)
}

// ---- ServiceNow ----

pub async fn servicenow_instances_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ServiceNowIntegrationAPI::with_client_and_config(dd_cfg, c),
        None => ServiceNowIntegrationAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_service_now_instances()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list ServiceNow instances: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn servicenow_templates_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ServiceNowIntegrationAPI::with_client_and_config(dd_cfg, c),
        None => ServiceNowIntegrationAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_service_now_templates()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list ServiceNow templates: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn servicenow_templates_get(cfg: &Config, template_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ServiceNowIntegrationAPI::with_client_and_config(dd_cfg, c),
        None => ServiceNowIntegrationAPI::with_config(dd_cfg),
    };
    let uuid = Uuid::parse_str(template_id)
        .map_err(|e| anyhow::anyhow!("invalid template UUID '{template_id}': {e}"))?;
    let resp = api
        .get_service_now_template(uuid)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get ServiceNow template: {e:?}"))?;
    formatter::print_json(&resp)
}
