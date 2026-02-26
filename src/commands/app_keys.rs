use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_key_management::{
    KeyManagementAPI, ListApplicationKeysOptionalParams,
    ListCurrentUserApplicationKeysOptionalParams,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::ApplicationKeysSort;

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

#[cfg(not(target_arch = "wasm32"))]
fn parse_sort(s: &str) -> Result<ApplicationKeysSort> {
    match s {
        "created_at" => Ok(ApplicationKeysSort::CREATED_AT_ASCENDING),
        "-created_at" => Ok(ApplicationKeysSort::CREATED_AT_DESCENDING),
        "last4" => Ok(ApplicationKeysSort::LAST4_ASCENDING),
        "-last4" => Ok(ApplicationKeysSort::LAST4_DESCENDING),
        "name" => Ok(ApplicationKeysSort::NAME_ASCENDING),
        "-name" => Ok(ApplicationKeysSort::NAME_DESCENDING),
        _ => anyhow::bail!(
            "invalid --sort value: {s:?}\nExpected: name, -name, created_at, -created_at, last4, -last4"
        ),
    }
}

// ---------------------------------------------------------------------------
// List application keys (current user)
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(
    cfg: &Config,
    page_size: i64,
    page_number: i64,
    filter: &str,
    sort: &str,
) -> Result<()> {
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
    if !filter.is_empty() {
        params.filter = Some(filter.to_string());
    }
    if !sort.is_empty() {
        params.sort = Some(parse_sort(sort)?);
    }

    let resp = api
        .list_current_user_application_keys(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list application keys: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn list(
    cfg: &Config,
    page_size: i64,
    page_number: i64,
    filter: &str,
    sort: &str,
) -> Result<()> {
    let mut query: Vec<(&str, String)> = Vec::new();
    if page_size > 0 {
        query.push(("page[size]", page_size.to_string()));
    }
    if page_number > 0 {
        query.push(("page[number]", page_number.to_string()));
    }
    if !filter.is_empty() {
        query.push(("filter", filter.to_string()));
    }
    if !sort.is_empty() {
        query.push(("sort", sort.to_string()));
    }
    let data = crate::api::get(cfg, "/api/v2/current_user/application_keys", &query).await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// List all application keys (org-wide, requires API keys)
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn list_all(
    cfg: &Config,
    page_size: i64,
    page_number: i64,
    filter: &str,
    sort: &str,
) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => KeyManagementAPI::with_client_and_config(dd_cfg, c),
        None => KeyManagementAPI::with_config(dd_cfg),
    };

    let mut params = ListApplicationKeysOptionalParams::default();
    if page_size > 0 {
        params.page_size = Some(page_size);
    }
    if page_number > 0 {
        params.page_number = Some(page_number);
    }
    if !filter.is_empty() {
        params.filter = Some(filter.to_string());
    }
    if !sort.is_empty() {
        params.sort = Some(parse_sort(sort)?);
    }

    let resp = api
        .list_application_keys(params)
        .await
        .map_err(|e| anyhow::anyhow!("failed to list all application keys: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn list_all(
    cfg: &Config,
    page_size: i64,
    page_number: i64,
    filter: &str,
    sort: &str,
) -> Result<()> {
    let mut query: Vec<(&str, String)> = Vec::new();
    if page_size > 0 {
        query.push(("page[size]", page_size.to_string()));
    }
    if page_number > 0 {
        query.push(("page[number]", page_number.to_string()));
    }
    if !filter.is_empty() {
        query.push(("filter", filter.to_string()));
    }
    if !sort.is_empty() {
        query.push(("sort", sort.to_string()));
    }
    let data = crate::api::get(cfg, "/api/v2/application_keys", &query).await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Get application key details (current user)
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
// Create application key (current user)
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn create(cfg: &Config, name: &str, scopes: &str) -> Result<()> {
    use datadog_api_client::datadogV2::model::{
        ApplicationKeyCreateAttributes, ApplicationKeyCreateData, ApplicationKeyCreateRequest,
        ApplicationKeysType,
    };

    let mut attrs = ApplicationKeyCreateAttributes::new(name.to_string());
    if !scopes.is_empty() {
        let scope_list: Vec<String> = scopes
            .split(',')
            .map(|s| s.trim().to_string())
            .filter(|s| !s.is_empty())
            .collect();
        attrs.scopes = Some(Some(scope_list));
    }

    let body = ApplicationKeyCreateRequest::new(ApplicationKeyCreateData::new(
        attrs,
        ApplicationKeysType::APPLICATION_KEYS,
    ));

    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => KeyManagementAPI::with_client_and_config(dd_cfg, c),
        None => KeyManagementAPI::with_config(dd_cfg),
    };
    let resp = api
        .create_current_user_application_key(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create application key: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn create(cfg: &Config, name: &str, scopes: &str) -> Result<()> {
    let mut attrs = serde_json::json!({ "name": name });
    if !scopes.is_empty() {
        let scope_list: Vec<&str> = scopes
            .split(',')
            .map(|s| s.trim())
            .filter(|s| !s.is_empty())
            .collect();
        attrs["scopes"] = serde_json::json!(scope_list);
    }
    let body = serde_json::json!({
        "data": {
            "type": "application_keys",
            "attributes": attrs,
        }
    });
    let data = crate::api::post(cfg, "/api/v2/current_user/application_keys", &body).await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// Update application key (current user)
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub async fn update(cfg: &Config, key_id: &str, name: &str, scopes: &str) -> Result<()> {
    use datadog_api_client::datadogV2::model::{
        ApplicationKeyUpdateAttributes, ApplicationKeyUpdateData, ApplicationKeyUpdateRequest,
        ApplicationKeysType,
    };

    let mut attrs = ApplicationKeyUpdateAttributes::new();
    if !name.is_empty() {
        attrs.name = Some(name.to_string());
    }
    if !scopes.is_empty() {
        let scope_list: Vec<String> = scopes
            .split(',')
            .map(|s| s.trim().to_string())
            .filter(|s| !s.is_empty())
            .collect();
        attrs.scopes = Some(Some(scope_list));
    }

    let body = ApplicationKeyUpdateRequest::new(ApplicationKeyUpdateData::new(
        attrs,
        key_id.to_string(),
        ApplicationKeysType::APPLICATION_KEYS,
    ));

    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => KeyManagementAPI::with_client_and_config(dd_cfg, c),
        None => KeyManagementAPI::with_config(dd_cfg),
    };
    let resp = api
        .update_current_user_application_key(key_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update application key: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn update(cfg: &Config, key_id: &str, name: &str, scopes: &str) -> Result<()> {
    let mut attrs = serde_json::json!({});
    if !name.is_empty() {
        attrs["name"] = serde_json::json!(name);
    }
    if !scopes.is_empty() {
        let scope_list: Vec<&str> = scopes
            .split(',')
            .map(|s| s.trim())
            .filter(|s| !s.is_empty())
            .collect();
        attrs["scopes"] = serde_json::json!(scope_list);
    }
    let body = serde_json::json!({
        "data": {
            "type": "application_keys",
            "id": key_id,
            "attributes": attrs,
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
// Delete application key (current user)
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
