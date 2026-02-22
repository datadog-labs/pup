use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_usage_metering::{
    GetCostByOrgOptionalParams, GetMonthlyCostAttributionOptionalParams,
    GetProjectedCostOptionalParams, UsageMeteringAPI as UsageMeteringV2API,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn projected(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => UsageMeteringV2API::with_client_and_config(dd_cfg, c),
        None => UsageMeteringV2API::with_config(dd_cfg),
    };
    let resp = api
        .get_projected_cost(GetProjectedCostOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get projected cost: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn projected(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/usage/projected_cost", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn by_org(cfg: &Config, start_month: String, end_month: Option<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => UsageMeteringV2API::with_client_and_config(dd_cfg, c),
        None => UsageMeteringV2API::with_config(dd_cfg),
    };

    let start_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&start_month)?)
            .unwrap();

    let mut params = GetCostByOrgOptionalParams::default();
    if let Some(e) = end_month {
        let end_dt =
            chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&e)?).unwrap();
        params = params.end_month(end_dt);
    }

    let resp = api
        .get_cost_by_org(start_dt, params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get cost by org: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn by_org(cfg: &Config, start_month: String, end_month: Option<String>) -> Result<()> {
    let start_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&start_month)?)
            .unwrap();
    let mut query = vec![("start_month", start_dt.to_rfc3339())];
    if let Some(e) = end_month {
        let end_dt =
            chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&e)?).unwrap();
        query.push(("end_month", end_dt.to_rfc3339()));
    }
    let data = crate::api::get(cfg, "/api/v2/usage/cost_by_org", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn attribution(cfg: &Config, start: String, fields: Option<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => UsageMeteringV2API::with_client_and_config(dd_cfg, c),
        None => UsageMeteringV2API::with_config(dd_cfg),
    };

    let start_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&start)?).unwrap();

    let fields_str = fields.unwrap_or_else(|| "*".to_string());
    let params = GetMonthlyCostAttributionOptionalParams::default();

    let resp = api
        .get_monthly_cost_attribution(start_dt, fields_str, params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get cost attribution: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn attribution(cfg: &Config, start: String, fields: Option<String>) -> Result<()> {
    let start_dt =
        chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&start)?).unwrap();
    let fields_str = fields.unwrap_or_else(|| "*".to_string());
    let query = vec![
        ("start_month", start_dt.to_rfc3339()),
        ("fields", fields_str),
    ];
    let data = crate::api::get(cfg, "/api/v2/cost_by_tag/monthly_cost_attribution", &query).await?;
    crate::formatter::output(cfg, &data)
}
