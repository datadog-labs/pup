use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_usage_metering::{
    GetHourlyUsageAttributionOptionalParams, GetUsageSummaryOptionalParams, UsageMeteringAPI,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::model::HourlyUsageAttributionUsageType;

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn summary(cfg: &Config, start: String, end: Option<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => UsageMeteringAPI::with_client_and_config(dd_cfg, c),
        None => UsageMeteringAPI::with_config(dd_cfg),
    };

    let start_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&start)?).unwrap();

    let mut params = GetUsageSummaryOptionalParams::default();
    if let Some(e) = end {
        let end_dt =
            chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&e)?).unwrap();
        params = params.end_month(end_dt);
    }

    let resp = api
        .get_usage_summary(start_dt, params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get usage summary: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn summary(cfg: &Config, start: String, end: Option<String>) -> Result<()> {
    let start_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&start)?).unwrap();
    let mut query = vec![("start_month", start_dt.to_rfc3339())];
    if let Some(e) = end {
        let end_dt =
            chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&e)?).unwrap();
        query.push(("end_month", end_dt.to_rfc3339()));
    }
    let data = crate::api::get(cfg, "/api/v1/usage/summary", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn hourly(cfg: &Config, start: String, end: Option<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => UsageMeteringAPI::with_client_and_config(dd_cfg, c),
        None => UsageMeteringAPI::with_config(dd_cfg),
    };

    let start_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&start)?).unwrap();

    let mut params = GetHourlyUsageAttributionOptionalParams::default();
    if let Some(e) = end {
        let end_dt =
            chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&e)?).unwrap();
        params = params.end_hr(end_dt);
    }

    let resp = api
        .get_hourly_usage_attribution(start_dt, HourlyUsageAttributionUsageType::API_USAGE, params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get hourly usage: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn hourly(cfg: &Config, start: String, end: Option<String>) -> Result<()> {
    let start_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&start)?).unwrap();
    let mut query = vec![
        ("start_hr", start_dt.to_rfc3339()),
        ("usage_type", "api_usage".to_string()),
    ];
    if let Some(e) = end {
        let end_dt =
            chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&e)?).unwrap();
        query.push(("end_hr", end_dt.to_rfc3339()));
    }
    let data = crate::api::get(cfg, "/api/v1/usage/hourly-attribution", &query).await?;
    crate::formatter::output(cfg, &data)
}
