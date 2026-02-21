use anyhow::Result;
use datadog_api_client::datadogV2::api_users::{UsersAPI, ListUsersOptionalParams};
use datadog_api_client::datadogV2::api_roles::{RolesAPI, ListRolesOptionalParams};

use crate::client;
use crate::config::Config;
use crate::formatter;

pub async fn list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => UsersAPI::with_client_and_config(dd_cfg, c),
        None => UsersAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_users(ListUsersOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list users: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn get(cfg: &Config, id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => UsersAPI::with_client_and_config(dd_cfg, c),
        None => UsersAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_user(id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get user: {e:?}"))?;
    formatter::print_json(&resp)
}

pub async fn roles_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => RolesAPI::with_client_and_config(dd_cfg, c),
        None => RolesAPI::with_config(dd_cfg),
    };
    let resp = api
        .list_roles(ListRolesOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list roles: {e:?}"))?;
    formatter::print_json(&resp)
}
