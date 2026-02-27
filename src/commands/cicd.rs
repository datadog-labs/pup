use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_ci_visibility_pipelines::{
    CIVisibilityPipelinesAPI, SearchCIAppPipelineEventsOptionalParams,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_ci_visibility_tests::{
    CIVisibilityTestsAPI, ListCIAppTestEventsOptionalParams, SearchCIAppTestEventsOptionalParams,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_dora_metrics::DORAMetricsAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_test_optimization::{
    SearchFlakyTestsOptionalParams, TestOptimizationAPI,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::{
    CIAppPipelineEventsRequest, CIAppPipelinesQueryFilter, CIAppQueryPageOptions, CIAppSort,
    CIAppTestEventsRequest, CIAppTestsQueryFilter, DORADeploymentPatchRequest,
    FlakyTestsSearchRequest, UpdateFlakyTestsRequest,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
#[cfg(not(target_arch = "wasm32"))]
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn pipelines_list(
    cfg: &Config,
    query: Option<String>,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = CIVisibilityPipelinesAPI::with_client_and_config(dd_cfg, dd_client);

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();

    let mut filter = CIAppPipelinesQueryFilter::new().from(from_str).to(to_str);
    if let Some(q) = query {
        filter = filter.query(q);
    }

    let body = CIAppPipelineEventsRequest::new()
        .filter(filter)
        .page(CIAppQueryPageOptions::new().limit(limit))
        .sort(CIAppSort::TIMESTAMP_DESCENDING);

    let params = SearchCIAppPipelineEventsOptionalParams::default().body(body);
    let resp = api
        .search_ci_app_pipeline_events(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list pipelines: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn pipelines_list(
    cfg: &Config,
    query: Option<String>,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    let from_ms = crate::util::parse_time_to_unix_millis(&from)?;
    let to_ms = crate::util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();
    let mut filter = serde_json::json!({ "from": from_str, "to": to_str });
    if let Some(q) = query {
        filter["query"] = serde_json::Value::String(q);
    }
    let body = serde_json::json!({
        "filter": filter,
        "page": { "limit": limit },
        "sort": "-timestamp"
    });
    let data = crate::api::post(cfg, "/api/v2/ci/pipelines/events/search", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn tests_list(
    cfg: &Config,
    query: Option<String>,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = CIVisibilityTestsAPI::with_client_and_config(dd_cfg, dd_client);

    let from_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&from)?).unwrap();
    let to_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&to)?).unwrap();

    let mut params = ListCIAppTestEventsOptionalParams::default()
        .filter_from(from_dt)
        .filter_to(to_dt)
        .page_limit(limit);
    if let Some(q) = query {
        params = params.filter_query(q);
    }

    let resp = api
        .list_ci_app_test_events(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list tests: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn tests_list(
    cfg: &Config,
    query: Option<String>,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    let from_ms = crate::util::parse_time_to_unix_millis(&from)?;
    let to_ms = crate::util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();
    let mut q: Vec<(&str, String)> = vec![
        ("filter[from]", from_str),
        ("filter[to]", to_str),
        ("page[limit]", limit.to_string()),
    ];
    if let Some(qv) = query {
        q.push(("filter[query]", qv));
    }
    let data = crate::api::get(cfg, "/api/v2/ci/tests/events", &q).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn events_search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = CIVisibilityPipelinesAPI::with_client_and_config(dd_cfg, dd_client);

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();

    let filter = CIAppPipelinesQueryFilter::new()
        .from(from_str)
        .to(to_str)
        .query(query);

    let body = CIAppPipelineEventsRequest::new()
        .filter(filter)
        .page(CIAppQueryPageOptions::new().limit(limit))
        .sort(CIAppSort::TIMESTAMP_DESCENDING);

    let params = SearchCIAppPipelineEventsOptionalParams::default().body(body);
    let resp = api
        .search_ci_app_pipeline_events(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search pipeline events: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn events_search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    let from_ms = crate::util::parse_time_to_unix_millis(&from)?;
    let to_ms = crate::util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();
    let body = serde_json::json!({
        "filter": {
            "from": from_str,
            "to": to_str,
            "query": query
        },
        "page": { "limit": limit },
        "sort": "-timestamp"
    });
    let data = crate::api::post(cfg, "/api/v2/ci/pipelines/events/search", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn events_aggregate(cfg: &Config, query: String, from: String, to: String) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = CIVisibilityPipelinesAPI::with_client_and_config(dd_cfg, dd_client);

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();

    let filter = CIAppPipelinesQueryFilter::new()
        .from(from_str)
        .to(to_str)
        .query(query);

    let body = CIAppPipelineEventsRequest::new().filter(filter);

    let params = SearchCIAppPipelineEventsOptionalParams::default().body(body);
    let resp = api
        .search_ci_app_pipeline_events(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to aggregate pipeline events: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn events_aggregate(cfg: &Config, query: String, from: String, to: String) -> Result<()> {
    let from_ms = crate::util::parse_time_to_unix_millis(&from)?;
    let to_ms = crate::util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();
    let body = serde_json::json!({
        "filter": {
            "from": from_str,
            "to": to_str,
            "query": query
        }
    });
    let data = crate::api::post(cfg, "/api/v2/ci/pipelines/events/search", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn tests_search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = CIVisibilityTestsAPI::with_client_and_config(dd_cfg, dd_client);

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();

    let filter = CIAppTestsQueryFilter::new()
        .from(from_str)
        .to(to_str)
        .query(query);

    let body = CIAppTestEventsRequest::new()
        .filter(filter)
        .page(CIAppQueryPageOptions::new().limit(limit))
        .sort(CIAppSort::TIMESTAMP_DESCENDING);

    let params = SearchCIAppTestEventsOptionalParams::default().body(body);
    let resp = api
        .search_ci_app_test_events(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search test events: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn tests_search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
) -> Result<()> {
    let from_ms = crate::util::parse_time_to_unix_millis(&from)?;
    let to_ms = crate::util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();
    let body = serde_json::json!({
        "filter": {
            "from": from_str,
            "to": to_str,
            "query": query
        },
        "page": { "limit": limit },
        "sort": "-timestamp"
    });
    let data = crate::api::post(cfg, "/api/v2/ci/tests/events/search", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn tests_aggregate(cfg: &Config, query: String, from: String, to: String) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = CIVisibilityTestsAPI::with_client_and_config(dd_cfg, dd_client);

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();

    let filter = CIAppTestsQueryFilter::new()
        .from(from_str)
        .to(to_str)
        .query(query);

    let body = CIAppTestEventsRequest::new().filter(filter);

    let params = SearchCIAppTestEventsOptionalParams::default().body(body);
    let resp = api
        .search_ci_app_test_events(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to aggregate test events: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn tests_aggregate(cfg: &Config, query: String, from: String, to: String) -> Result<()> {
    let from_ms = crate::util::parse_time_to_unix_millis(&from)?;
    let to_ms = crate::util::parse_time_to_unix_millis(&to)?;
    let from_str = chrono::DateTime::from_timestamp_millis(from_ms)
        .unwrap()
        .to_rfc3339();
    let to_str = chrono::DateTime::from_timestamp_millis(to_ms)
        .unwrap()
        .to_rfc3339();
    let body = serde_json::json!({
        "filter": {
            "from": from_str,
            "to": to_str,
            "query": query
        }
    });
    let data = crate::api::post(cfg, "/api/v2/ci/tests/events/search", &body).await?;
    crate::formatter::output(cfg, &data)
}

// ---- Pipelines Get ----

#[cfg(not(target_arch = "wasm32"))]
pub async fn pipelines_get(cfg: &Config, pipeline_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = CIVisibilityPipelinesAPI::with_client_and_config(dd_cfg, dd_client);

    let filter = CIAppPipelinesQueryFilter::new().query(pipeline_id.to_string());

    let body = CIAppPipelineEventsRequest::new().filter(filter);

    let params = SearchCIAppPipelineEventsOptionalParams::default().body(body);
    let resp = api
        .search_ci_app_pipeline_events(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get pipeline: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn pipelines_get(cfg: &Config, pipeline_id: &str) -> Result<()> {
    let body = serde_json::json!({
        "filter": {
            "query": pipeline_id
        }
    });
    let data = crate::api::post(cfg, "/api/v2/ci/pipelines/events/search", &body).await?;
    crate::formatter::output(cfg, &data)
}

// ---- DORA Metrics ----

#[cfg(not(target_arch = "wasm32"))]
pub async fn dora_patch_deployment(cfg: &Config, deployment_id: &str, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = DORAMetricsAPI::with_client_and_config(dd_cfg, dd_client);
    let body: DORADeploymentPatchRequest = crate::util::read_json_file(file)?;
    api.patch_dora_deployment(deployment_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to patch DORA deployment: {e:?}"))?;
    println!("DORA deployment '{deployment_id}' patched successfully.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn dora_patch_deployment(cfg: &Config, deployment_id: &str, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let path = format!("/api/v2/dora/deployments/{deployment_id}");
    crate::api::patch(cfg, &path, &body).await?;
    println!("DORA deployment '{deployment_id}' patched successfully.");
    Ok(())
}

// ---- Flaky Tests ----

#[cfg(not(target_arch = "wasm32"))]
pub async fn flaky_tests_search(cfg: &Config, query: Option<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = TestOptimizationAPI::with_client_and_config(dd_cfg, dd_client);

    let mut body = FlakyTestsSearchRequest::new();
    if let Some(q) = query {
        use datadog_api_client::datadogV2::model::{
            FlakyTestsSearchFilter, FlakyTestsSearchRequestAttributes, FlakyTestsSearchRequestData,
            FlakyTestsSearchRequestDataType,
        };
        let filter = FlakyTestsSearchFilter::new().query(q);
        let attrs = FlakyTestsSearchRequestAttributes::new().filter(filter);
        let data = FlakyTestsSearchRequestData::new()
            .attributes(attrs)
            .type_(FlakyTestsSearchRequestDataType::SEARCH_FLAKY_TESTS_REQUEST);
        body = body.data(data);
    }

    let params = SearchFlakyTestsOptionalParams::default().body(body);
    let resp = api
        .search_flaky_tests(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search flaky tests: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn flaky_tests_search(cfg: &Config, query: Option<String>) -> Result<()> {
    let mut body = serde_json::json!({});
    if let Some(q) = query {
        body = serde_json::json!({
            "data": {
                "attributes": {
                    "filter": {
                        "query": q
                    }
                },
                "type": "search_flaky_tests_request"
            }
        });
    }
    let data = crate::api::post(cfg, "/api/v2/ci/tests/flaky", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn flaky_tests_update(cfg: &Config, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = TestOptimizationAPI::with_client_and_config(dd_cfg, dd_client);
    let body: UpdateFlakyTestsRequest = crate::util::read_json_file(file)?;
    let resp = api
        .update_flaky_tests(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update flaky tests: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn flaky_tests_update(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = crate::api::patch(cfg, "/api/v2/ci/tests/flaky", &body).await?;
    crate::formatter::output(cfg, &data)
}
