use anyhow::{bail, Result};
use chrono::Utc;
use datadog_api_client::datadogV2::api_error_tracking::{
    ErrorTrackingAPI, GetIssueOptionalParams, SearchIssuesOptionalParams,
};
use datadog_api_client::datadogV2::model::{
    IssuesSearchRequest, IssuesSearchRequestData, IssuesSearchRequestDataAttributes,
    IssuesSearchRequestDataType,
};

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn issues_search(cfg: &Config, query: Option<String>, _limit: i32) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("error tracking requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = ErrorTrackingAPI::with_config(dd_cfg);

    let now = Utc::now().timestamp_millis();
    let one_day_ago = now - 86_400_000; // 24 hours in millis

    let query_str = query.unwrap_or_else(|| "*".to_string());
    let attrs = IssuesSearchRequestDataAttributes::new(one_day_ago, query_str, now);
    let data = IssuesSearchRequestData::new(attrs, IssuesSearchRequestDataType::SEARCH_REQUEST);
    let body = IssuesSearchRequest::new(data);
    let params = SearchIssuesOptionalParams::default();

    let resp = api
        .search_issues(body, params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search issues: {e:?}"))?;
    let val = serde_json::to_value(&resp)?;
    if let Some(data) = val.get("data") {
        if data.as_array().map_or(false, |a| a.is_empty()) {
            println!("No error tracking issues found matching the specified criteria.");
            return Ok(());
        }
    }
    formatter::output(cfg, &resp)
}

pub async fn issues_get(cfg: &Config, issue_id: &str) -> Result<()> {
    if !cfg.has_api_keys() {
        bail!("error tracking requires API key authentication (DD_API_KEY + DD_APP_KEY)");
    }
    let dd_cfg = client::make_dd_config(cfg);
    let api = ErrorTrackingAPI::with_config(dd_cfg);
    let resp = api
        .get_issue(issue_id.to_string(), GetIssueOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get issue: {e:?}"))?;
    formatter::output(cfg, &resp)
}
