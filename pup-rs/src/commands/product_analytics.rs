use anyhow::Result;
use datadog_api_client::datadogV2::api_product_analytics::ProductAnalyticsAPI;
use datadog_api_client::datadogV2::model::ProductAnalyticsServerSideEventItem;

use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

pub async fn events_send(cfg: &Config, file: &str) -> Result<()> {
    let body: ProductAnalyticsServerSideEventItem = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ProductAnalyticsAPI::with_client_and_config(dd_cfg, c),
        None => ProductAnalyticsAPI::with_config(dd_cfg),
    };
    let resp = api
        .submit_product_analytics_event(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to send product analytics event: {e:?}"))?;
    formatter::output(cfg, &resp)
}
