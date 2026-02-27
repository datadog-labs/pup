use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_key_management::{
    GetAPIKeyOptionalParams, KeyManagementAPI, ListAPIKeysOptionalParams,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = KeyManagementAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_api_keys(ListAPIKeysOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list API keys: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/api_keys", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn get(cfg: &Config, key_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = KeyManagementAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .get_api_key(key_id.to_string(), GetAPIKeyOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get API key: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn get(cfg: &Config, key_id: &str) -> Result<()> {
    let data = crate::api::get(cfg, &format!("/api/v2/api_keys/{key_id}"), &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn create(cfg: &Config, name: &str) -> Result<()> {
    use datadog_api_client::datadogV2::model::{
        APIKeyCreateAttributes, APIKeyCreateData, APIKeyCreateRequest, APIKeysType,
    };
    let body = APIKeyCreateRequest::new(APIKeyCreateData::new(
        APIKeyCreateAttributes::new(name.to_string()),
        APIKeysType::API_KEYS,
    ));
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = KeyManagementAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .create_api_key(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create API key: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn create(cfg: &Config, name: &str) -> Result<()> {
    let body = serde_json::json!({
        "data": {
            "type": "api_keys",
            "attributes": {
                "name": name,
            }
        }
    });
    let data = crate::api::post(cfg, "/api/v2/api_keys", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn delete(cfg: &Config, key_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = KeyManagementAPI::with_client_and_config(dd_cfg, dd_client);
    api.delete_api_key(key_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete API key: {e:?}"))?;
    println!("Successfully deleted API key {key_id}");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn delete(cfg: &Config, key_id: &str) -> Result<()> {
    crate::api::delete(cfg, &format!("/api/v2/api_keys/{key_id}")).await?;
    println!("Successfully deleted API key {key_id}");
    Ok(())
}
