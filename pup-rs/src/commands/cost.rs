use anyhow::Result;
use datadog_api_client::datadogV2::api_usage_metering::{
    UsageMeteringAPI as UsageMeteringV2API, GetProjectedCostOptionalParams,
    GetCostByOrgOptionalParams,
};

use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

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
    formatter::print_json(&resp)
}

pub async fn by_org(cfg: &Config, start_month: String, end_month: Option<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => UsageMeteringV2API::with_client_and_config(dd_cfg, c),
        None => UsageMeteringV2API::with_config(dd_cfg),
    };

    let start_dt = chrono::DateTime::from_timestamp_millis(
        util::parse_time_to_unix_millis(&start_month)?,
    )
    .unwrap();

    let mut params = GetCostByOrgOptionalParams::default();
    if let Some(e) = end_month {
        let end_dt =
            chrono::DateTime::from_timestamp_millis(util::parse_time_to_unix_millis(&e)?)
                .unwrap();
        params = params.end_month(end_dt);
    }

    let resp = api
        .get_cost_by_org(start_dt, params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get cost by org: {e:?}"))?;
    formatter::print_json(&resp)
}
