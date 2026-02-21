use anyhow::{bail, Result};
use datadog_api_client::datadogV2::api_logs::{ListLogsOptionalParams, LogsAPI};
use datadog_api_client::datadogV2::model::{
    LogsListRequest, LogsListRequestPage, LogsQueryFilter, LogsSort,
};

use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

pub async fn search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    // Logs search API doesn't support OAuth/bearer - force API keys
    if !cfg.has_api_keys() {
        bail!(
            "logs search requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    // Force API key auth only - do NOT use bearer middleware
    let api = LogsAPI::with_config(dd_cfg);

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;

    let body = LogsListRequest::new()
        .filter(
            LogsQueryFilter::new()
                .query(query)
                .from(from_ms.to_string())
                .to(to_ms.to_string()),
        )
        .page(LogsListRequestPage::new().limit(limit))
        .sort(LogsSort::TIMESTAMP_DESCENDING);

    let params = ListLogsOptionalParams::default().body(body);

    let resp = api
        .list_logs(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search logs: {:?}", e))?;

    formatter::print_json(&resp)?;
    Ok(())
}
