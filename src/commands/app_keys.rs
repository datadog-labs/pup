use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_action_connection::{
    ActionConnectionAPI, ListAppKeyRegistrationsOptionalParams,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

// ---------------------------------------------------------------------------
// List app key registrations
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(cfg: &Config, page_size: i64, page_number: i64) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ActionConnectionAPI::with_client_and_config(dd_cfg, c),
        None => ActionConnectionAPI::with_config(dd_cfg),
    };

    let mut params = ListAppKeyRegistrationsOptionalParams::default();
    if page_size > 0 {
        params.page_size = Some(page_size);
    }
    if page_number > 0 {
        params.page_number = Some(page_number);
    }

    let resp = api
        .list_app_key_registrations(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list app key registrations: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn list(cfg: &Config, page_size: i64, page_number: i64) -> Result<()> {
    let mut query: Vec<(&str, String)> = Vec::new();
    if page_size > 0 {
        query.push(("page[size]", page_size.to_string()));
    }
    if page_number > 0 {
        query.push(("page[number]", page_number.to_string()));
    }
    let data = crate::api::get(
        cfg,
        "/api/v2/integration/action_connections/app-keys",
        &query,
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Get app key registration
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn get(cfg: &Config, key_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ActionConnectionAPI::with_client_and_config(dd_cfg, c),
        None => ActionConnectionAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_app_key_registration(key_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get app key registration: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn get(cfg: &Config, key_id: &str) -> Result<()> {
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/integration/action_connections/app-keys/{key_id}"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Register app key for Action Connections
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn register(cfg: &Config, key_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ActionConnectionAPI::with_client_and_config(dd_cfg, c),
        None => ActionConnectionAPI::with_config(dd_cfg),
    };
    let resp = api
        .register_app_key(key_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to register app key: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn register(cfg: &Config, key_id: &str) -> Result<()> {
    let body = serde_json::json!({});
    let data = crate::api::post(
        cfg,
        &format!("/api/v2/integration/action_connections/app-keys/{key_id}"),
        &body,
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Unregister app key from Action Connections
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn unregister(cfg: &Config, key_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ActionConnectionAPI::with_client_and_config(dd_cfg, c),
        None => ActionConnectionAPI::with_config(dd_cfg),
    };
    api.unregister_app_key(key_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to unregister app key: {e:?}"))?;
    println!("Successfully unregistered app key {key_id}");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn unregister(cfg: &Config, key_id: &str) -> Result<()> {
    crate::api::delete(
        cfg,
        &format!("/api/v2/integration/action_connections/app-keys/{key_id}"),
    )
    .await?;
    println!("Successfully unregistered app key {key_id}");
    Ok(())
}
