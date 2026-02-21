use anyhow::Result;
use datadog_api_client::datadogV1::api_hosts::{HostsAPI, ListHostsOptionalParams};

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn hosts_list(
    cfg: &Config,
    filter: Option<String>,
    sort: String,
    count: i64,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => HostsAPI::with_client_and_config(dd_cfg, c),
        None => HostsAPI::with_config(dd_cfg),
    };
    let mut params = ListHostsOptionalParams::default().count(count).sort_field(sort);
    if let Some(f) = filter {
        params = params.filter(f);
    }
    let resp = api
        .list_hosts(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list hosts: {e:?}"))?;
    formatter::print_json(&resp)
}
