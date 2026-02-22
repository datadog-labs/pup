use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_audit::{
    AuditAPI, ListAuditLogsOptionalParams, SearchAuditLogsOptionalParams,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::{
    AuditLogsQueryFilter, AuditLogsQueryPageOptions, AuditLogsSearchEventsRequest, AuditLogsSort,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(cfg: &Config, from: String, to: String, limit: i32) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => AuditAPI::with_client_and_config(dd_cfg, c),
        None => AuditAPI::with_config(dd_cfg),
    };

    let from_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&from)?).unwrap();
    let to_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&to)?).unwrap();

    let params = ListAuditLogsOptionalParams::default()
        .filter_from(from_dt)
        .filter_to(to_dt)
        .page_limit(limit);

    let resp = api
        .list_audit_logs(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list audit logs: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn list(cfg: &Config, from: String, to: String, limit: i32) -> Result<()> {
    let from_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&from)?).unwrap();
    let to_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&to)?).unwrap();
    let query = vec![
        ("filter[from]", from_dt.to_rfc3339()),
        ("filter[to]", to_dt.to_rfc3339()),
        ("page[limit]", limit.to_string()),
    ];
    let data = crate::api::get(cfg, "/api/v2/audit/events", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => AuditAPI::with_client_and_config(dd_cfg, c),
        None => AuditAPI::with_config(dd_cfg),
    };

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;

    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();

    let body = AuditLogsSearchEventsRequest::new()
        .filter(
            AuditLogsQueryFilter::new()
                .query(query)
                .from(from_str)
                .to(to_str),
        )
        .page(AuditLogsQueryPageOptions::new().limit(limit))
        .sort(AuditLogsSort::TIMESTAMP_DESCENDING);

    let params = SearchAuditLogsOptionalParams::default().body(body);
    let resp = api
        .search_audit_logs(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search audit logs: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();
    let body = serde_json::json!({
        "filter": {
            "query": query,
            "from": from_str,
            "to": to_str,
        },
        "page": {
            "limit": limit,
        },
        "sort": "timestamp",
    });
    let data = crate::api::post(cfg, "/api/v2/audit/events/search", &body).await?;
    crate::formatter::output(cfg, &data)
}
