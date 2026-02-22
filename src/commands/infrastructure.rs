use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_hosts::{HostsAPI, ListHostsOptionalParams};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

#[cfg(not(target_arch = "wasm32"))]
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
    let mut params = ListHostsOptionalParams::default()
        .count(count)
        .sort_field(sort);
    if let Some(f) = filter {
        params = params.filter(f);
    }
    let resp = api
        .list_hosts(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list hosts: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn hosts_list(
    cfg: &Config,
    filter: Option<String>,
    sort: String,
    count: i64,
) -> Result<()> {
    let mut query = vec![("count", count.to_string()), ("sort_field", sort)];
    if let Some(f) = filter {
        query.push(("filter", f));
    }
    let data = crate::api::get(cfg, "/api/v1/hosts", &query).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn hosts_get(cfg: &Config, hostname: &str) -> Result<()> {
    // The V1 HostsAPI does not have a direct get-host method.
    // Use list_hosts with a filter to find the specific host.
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => HostsAPI::with_client_and_config(dd_cfg, c),
        None => HostsAPI::with_config(dd_cfg),
    };
    let params = ListHostsOptionalParams::default()
        .filter(hostname.to_string())
        .count(1);
    let resp = api
        .list_hosts(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get host {hostname}: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn hosts_get(cfg: &Config, hostname: &str) -> Result<()> {
    let query = vec![("filter", hostname.to_string()), ("count", "1".to_string())];
    let data = crate::api::get(cfg, "/api/v1/hosts", &query).await?;
    crate::formatter::output(cfg, &data)
}
