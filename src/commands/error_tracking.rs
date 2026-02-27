use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use chrono::Utc;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_error_tracking::{
    ErrorTrackingAPI, GetIssueOptionalParams, SearchIssuesOptionalParams,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::{
    IssuesSearchRequest, IssuesSearchRequestData, IssuesSearchRequestDataAttributes,
    IssuesSearchRequestDataType,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

#[cfg(not(target_arch = "wasm32"))]
pub async fn issues_search(cfg: &Config, query: Option<String>, _limit: i32) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ErrorTrackingAPI::with_client_and_config(dd_cfg, c),
        None => ErrorTrackingAPI::with_config(dd_cfg),
    };

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
        if data.as_array().is_some_and(|a| a.is_empty()) {
            println!("No error tracking issues found matching the specified criteria.");
            return Ok(());
        }
    }
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn issues_search(cfg: &Config, query: Option<String>, _limit: i32) -> Result<()> {
    let now = chrono::Utc::now().timestamp_millis();
    let one_day_ago = now - 86_400_000;
    let query_str = query.unwrap_or_else(|| "*".to_string());
    let body = serde_json::json!({
        "data": {
            "attributes": {
                "start": one_day_ago,
                "query": query_str,
                "end": now,
            },
            "type": "search_request",
        }
    });
    let data = crate::api::post(cfg, "/api/v2/error-tracking/issues/search", &body).await?;
    if let Some(arr) = data.get("data").and_then(|d| d.as_array()) {
        if arr.is_empty() {
            println!("No error tracking issues found matching the specified criteria.");
            return Ok(());
        }
    }
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn issues_get(cfg: &Config, issue_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ErrorTrackingAPI::with_client_and_config(dd_cfg, c),
        None => ErrorTrackingAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_issue(issue_id.to_string(), GetIssueOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get issue: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn issues_get(cfg: &Config, issue_id: &str) -> Result<()> {
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/error-tracking/issues/{issue_id}"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}
