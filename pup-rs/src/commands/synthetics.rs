use anyhow::Result;
use datadog_api_client::datadogV1::api_synthetics::{
    ListTestsOptionalParams, SearchTestsOptionalParams, SyntheticsAPI,
};
use datadog_api_client::datadogV2::api_synthetics::{
    SearchSuitesOptionalParams, SyntheticsAPI as SyntheticsV2API,
};
use datadog_api_client::datadogV2::model::{
    DeletedSuitesRequestDelete, DeletedSuitesRequestDeleteAttributes,
    DeletedSuitesRequestDeleteRequest, SuiteCreateEditRequest,
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
    formatter::output(cfg, &resp)
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
    formatter::output(cfg, &resp)
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
    formatter::output(cfg, &resp)
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
    formatter::output(cfg, &resp)
}

// ---- Suites (V2 API) ----

pub async fn suites_list(cfg: &Config, query: Option<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SyntheticsV2API::with_client_and_config(dd_cfg, c),
        None => SyntheticsV2API::with_config(dd_cfg),
    };
    let mut params = SearchSuitesOptionalParams::default();
    if let Some(q) = query {
        params = params.query(q);
    }
    let resp = api
        .search_suites(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list synthetic suites: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn suites_get(cfg: &Config, suite_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SyntheticsV2API::with_client_and_config(dd_cfg, c),
        None => SyntheticsV2API::with_config(dd_cfg),
    };
    let resp = api
        .get_synthetics_suite(suite_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get synthetic suite: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn suites_create(cfg: &Config, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SyntheticsV2API::with_client_and_config(dd_cfg, c),
        None => SyntheticsV2API::with_config(dd_cfg),
    };
    let body: SuiteCreateEditRequest = crate::util::read_json_file(file)?;
    let resp = api
        .create_synthetics_suite(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create synthetic suite: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn suites_update(cfg: &Config, suite_id: &str, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SyntheticsV2API::with_client_and_config(dd_cfg, c),
        None => SyntheticsV2API::with_config(dd_cfg),
    };
    let body: SuiteCreateEditRequest = crate::util::read_json_file(file)?;
    let resp = api
        .edit_synthetics_suite(suite_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update synthetic suite: {e:?}"))?;
    formatter::output(cfg, &resp)
}

pub async fn suites_delete(cfg: &Config, suite_ids: Vec<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => SyntheticsV2API::with_client_and_config(dd_cfg, c),
        None => SyntheticsV2API::with_config(dd_cfg),
    };
    let attrs = DeletedSuitesRequestDeleteAttributes::new(suite_ids);
    let data = DeletedSuitesRequestDelete::new(attrs);
    let body = DeletedSuitesRequestDeleteRequest::new(data);
    let resp = api
        .delete_synthetics_suites(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete synthetic suites: {e:?}"))?;
    formatter::output(cfg, &resp)
}
