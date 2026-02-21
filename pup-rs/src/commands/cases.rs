use anyhow::Result;
use datadog_api_client::datadogV2::api_case_management::{
    CaseManagementAPI, SearchCasesOptionalParams,
};

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn search(cfg: &Config, _query: Option<String>, page_size: i64) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => CaseManagementAPI::with_client_and_config(dd_cfg, c),
        None => CaseManagementAPI::with_config(dd_cfg),
    };
    let params = SearchCasesOptionalParams::default().page_size(page_size);
    let resp = api
        .search_cases(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search cases: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn get(cfg: &Config, case_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => CaseManagementAPI::with_client_and_config(dd_cfg, c),
        None => CaseManagementAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_case(case_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get case: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn projects_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => CaseManagementAPI::with_client_and_config(dd_cfg, c),
        None => CaseManagementAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_projects()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list projects: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn projects_get(cfg: &Config, project_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => CaseManagementAPI::with_client_and_config(dd_cfg, c),
        None => CaseManagementAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_project(project_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get project: {e:?}"))?;
    formatter::print_json(&resp)
}
