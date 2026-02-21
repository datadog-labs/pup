use anyhow::Result;
use datadog_api_client::datadogV1::api_synthetics::{
    SyntheticsAPI, ListTestsOptionalParams, SearchTestsOptionalParams,
};

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn tests_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SyntheticsAPI::with_client_and_config(dd_cfg, c),
        None => SyntheticsAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_tests(ListTestsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list tests: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn tests_get(cfg: &Config, public_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SyntheticsAPI::with_client_and_config(dd_cfg, c),
        None => SyntheticsAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_test(public_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get test: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn tests_search(
    cfg: &Config,
    text: Option<String>,
    count: i64,
    start: i64,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SyntheticsAPI::with_client_and_config(dd_cfg, c),
        None => SyntheticsAPI::with_config(dd_cfg),
    };

    let mut params = SearchTestsOptionalParams::default();
    if let Some(t) = text {
        params = params.text(t);
    }
    if count != 50 {
        params = params.count(count);
    }
    if start != 0 {
        params = params.start(start);
    }

    let resp = api
        .search_tests(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search tests: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn locations_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SyntheticsAPI::with_client_and_config(dd_cfg, c),
        None => SyntheticsAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_locations()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list locations: {e:?}"))?;
    formatter::print_json(&resp)
}
