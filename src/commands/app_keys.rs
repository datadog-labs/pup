use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_key_management::{
    KeyManagementAPI, ListApplicationKeysOptionalParams,
    ListCurrentUserApplicationKeysOptionalParams,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::{
    ApplicationKeyCreateAttributes, ApplicationKeyCreateData, ApplicationKeyCreateRequest,
    ApplicationKeyUpdateAttributes, ApplicationKeyUpdateData, ApplicationKeyUpdateRequest,
    ApplicationKeysSort, ApplicationKeysType,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

#[cfg(not(target_arch = "wasm32"))]
fn parse_sort(s: &str) -> ApplicationKeysSort {
    match s {
        "created_at" => ApplicationKeysSort::CREATED_AT_ASCENDING,
        "-created_at" => ApplicationKeysSort::CREATED_AT_DESCENDING,
        "name" => ApplicationKeysSort::NAME_ASCENDING,
        "-name" => ApplicationKeysSort::NAME_DESCENDING,
        "last4" => ApplicationKeysSort::LAST4_ASCENDING,
        "-last4" => ApplicationKeysSort::LAST4_DESCENDING,
        _ => ApplicationKeysSort::UnparsedObject(datadog_api_client::datadog::UnparsedObject {
            value: serde_json::Value::String(s.to_string()),
        }),
    }
}

// ---------------------------------------------------------------------------
// List (current user by default, org-wide with --all)
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(
    cfg: &Config,
    all: bool,
    filter: &Option<String>,
    sort: &Option<String>,
    page_size: i64,
    page_number: i64,
) -> Result<()> {
    if all {
        return list_all(cfg, filter, sort, page_size, page_number).await;
    }

    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => KeyManagementAPI::with_client_and_config(dd_cfg, c),
        None => KeyManagementAPI::with_config(dd_cfg),
    };

    let mut params = ListCurrentUserApplicationKeysOptionalParams::default();
    if page_size > 0 {
        params.page_size = Some(page_size);
    }
    if page_number > 0 {
        params.page_number = Some(page_number);
    }
    if let Some(f) = filter {
        params.filter = Some(f.clone());
    }
    if let Some(s) = sort {
        params.sort = Some(parse_sort(s));
    }

    let resp = api
        .list_current_user_application_keys(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list application keys: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(not(target_arch = "wasm32"))]
async fn list_all(
    cfg: &Config,
    filter: &Option<String>,
    sort: &Option<String>,
    page_size: i64,
    page_number: i64,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    // Org-wide list requires API keys, not OAuth
    let api = KeyManagementAPI::with_config(dd_cfg);

    let mut params = ListApplicationKeysOptionalParams::default();
    if page_size > 0 {
        params.page_size = Some(page_size);
    }
    if page_number > 0 {
        params.page_number = Some(page_number);
    }
    if let Some(f) = filter {
        params.filter = Some(f.clone());
    }
    if let Some(s) = sort {
        params.sort = Some(parse_sort(s));
    }

    let resp = api
        .list_application_keys(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list all application keys: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn list(
    cfg: &Config,
    all: bool,
    filter: &Option<String>,
    _sort: &Option<String>,
    page_size: i64,
    page_number: i64,
) -> Result<()> {
    let path = if all {
        "/api/v2/application_keys"
    } else {
        "/api/v2/current_user/application_keys"
    };
    let mut query: Vec<(&str, String)> = Vec::new();
    if page_size > 0 {
        query.push(("page[size]", page_size.to_string()));
    }
    if page_number > 0 {
        query.push(("page[number]", page_number.to_string()));
    }
    if let Some(f) = filter {
        query.push(("filter", f.clone()));
    }
    let data = crate::api::get(cfg, path, &query).await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn get(cfg: &Config, key_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => KeyManagementAPI::with_client_and_config(dd_cfg, c),
        None => KeyManagementAPI::with_config(dd_cfg),
    };
    let resp = api
        .get_current_user_application_key(key_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get application key: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn get(cfg: &Config, key_id: &str) -> Result<()> {
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/current_user/application_keys/{key_id}"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn create(cfg: &Config, name: &str, scopes: &Option<String>) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => KeyManagementAPI::with_client_and_config(dd_cfg, c),
        None => KeyManagementAPI::with_config(dd_cfg),
    };

    let mut attrs = ApplicationKeyCreateAttributes::new(name.to_string());
    if let Some(s) = scopes {
        let scope_list: Vec<String> = s.split(',').map(|v| v.trim().to_string()).collect();
        attrs.scopes = Some(Some(scope_list));
    }

    let body = ApplicationKeyCreateRequest::new(ApplicationKeyCreateData::new(
        attrs,
        ApplicationKeysType::APPLICATION_KEYS,
    ));

    let resp = api
        .create_current_user_application_key(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create application key: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn create(cfg: &Config, name: &str, scopes: &Option<String>) -> Result<()> {
    let mut attrs = serde_json::json!({ "name": name });
    if let Some(s) = scopes {
        let scope_list: Vec<&str> = s.split(',').map(|v| v.trim()).collect();
        attrs["scopes"] = serde_json::json!(scope_list);
    }
    let body = serde_json::json!({
        "data": {
            "attributes": attrs,
            "type": "application_keys",
        }
    });
    let data = crate::api::post(cfg, "/api/v2/current_user/application_keys", &body).await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn update(
    cfg: &Config,
    key_id: &str,
    name: &Option<String>,
    scopes: &Option<String>,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => KeyManagementAPI::with_client_and_config(dd_cfg, c),
        None => KeyManagementAPI::with_config(dd_cfg),
    };

    let mut attrs = ApplicationKeyUpdateAttributes::new();
    if let Some(n) = name {
        attrs.name = Some(n.clone());
    }
    if let Some(s) = scopes {
        let scope_list: Vec<String> = s.split(',').map(|v| v.trim().to_string()).collect();
        attrs.scopes = Some(Some(scope_list));
    }

    let body = ApplicationKeyUpdateRequest::new(ApplicationKeyUpdateData::new(
        attrs,
        key_id.to_string(),
        ApplicationKeysType::APPLICATION_KEYS,
    ));

    let resp = api
        .update_current_user_application_key(key_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update application key: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn update(
    cfg: &Config,
    key_id: &str,
    name: &Option<String>,
    scopes: &Option<String>,
) -> Result<()> {
    let mut attrs = serde_json::json!({});
    if let Some(n) = name {
        attrs["name"] = serde_json::json!(n);
    }
    if let Some(s) = scopes {
        let scope_list: Vec<&str> = s.split(',').map(|v| v.trim()).collect();
        attrs["scopes"] = serde_json::json!(scope_list);
    }
    let body = serde_json::json!({
        "data": {
            "attributes": attrs,
            "id": key_id,
            "type": "application_keys",
        }
    });
    let data = crate::api::patch(
        cfg,
        &format!("/api/v2/current_user/application_keys/{key_id}"),
        &body,
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn delete(cfg: &Config, key_id: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => KeyManagementAPI::with_client_and_config(dd_cfg, c),
        None => KeyManagementAPI::with_config(dd_cfg),
    };
    api.delete_current_user_application_key(key_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete application key: {e:?}"))?;
    println!("Successfully deleted application key {key_id}");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn delete(cfg: &Config, key_id: &str) -> Result<()> {
    crate::api::delete(
        cfg,
        &format!("/api/v2/current_user/application_keys/{key_id}"),
    )
    .await?;
    println!("Successfully deleted application key {key_id}");
    Ok(())
}
