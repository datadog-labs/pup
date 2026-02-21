use anyhow::Result;
use datadog_api_client::datadogV1::api_organizations::OrganizationsAPI;

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => OrganizationsAPI::with_client_and_config(dd_cfg, c),
        None => OrganizationsAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_orgs()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list orgs: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn get(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => OrganizationsAPI::with_client_and_config(dd_cfg, c),
        None => OrganizationsAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_org("current".to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get org: {e:?}"))?;
    formatter::print_json(&resp)
}
