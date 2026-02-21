use anyhow::{bail, Result};
use datadog_api_client::datadogV1::api_events::{EventsAPI as EventsV1API, ListEventsOptionalParams};
use datadog_api_client::datadogV2::api_events::{EventsAPI as EventsV2API, SearchEventsOptionalParams};
use datadog_api_client::datadogV2::model::{
    EventsListRequest, EventsQueryFilter, EventsRequestPage, EventsSort,
};

use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

pub async fn list(cfg: &Config, start: i64, end: i64, tags: Option<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => EventsV1API::with_client_and_config(dd_cfg, c),
        None => EventsV1API::with_config(dd_cfg),
    };

    // Default to last hour if not specified
    let now = chrono::Utc::now().timestamp();
    let start = if start == 0 { now - 3600 } else { start };
    let end = if end == 0 { now } else { end };

    let mut params = ListEventsOptionalParams::default();
    if let Some(t) = tags {
        params = params.tags(t);
    }
    let resp = api
        .list_events(start, end, params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list events: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    // Events search is OAuth-excluded — require API keys
    if !cfg.has_api_keys() {
        bail!(
            "events search requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let dd_cfg = client::make_dd_config(cfg);
    let api = EventsV2API::with_config(dd_cfg);

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;

    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();

    let body = EventsListRequest::new()
        .filter(
            EventsQueryFilter::new()
                .query(query)
                .from(from_str)
                .to(to_str),
        )
        .page(EventsRequestPage::new().limit(limit))
        .sort(EventsSort::TIMESTAMP_DESCENDING);

    let params = SearchEventsOptionalParams::default().body(body);
    let resp = api
        .search_events(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search events: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn get(cfg: &Config, id: i64) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => EventsV1API::with_client_and_config(dd_cfg, c),
        None => EventsV1API::with_config(dd_cfg),
    };
    let resp = api
        .get_event(id)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get event: {e:?}"))?;
    formatter::print_json(&resp)
}
