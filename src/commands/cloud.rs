use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_aws_integration::{
    AWSIntegrationAPI, ListAWSAccountsOptionalParams,
};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_azure_integration::AzureIntegrationAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_gcp_integration::GCPIntegrationAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_oci_integration::OCIIntegrationAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::{
    CreateTenancyConfigRequest, UpdateTenancyConfigRequest,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;

#[cfg(not(target_arch = "wasm32"))]
pub async fn aws_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = AWSIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_aws_accounts(ListAWSAccountsOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list AWS accounts: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn aws_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v1/integration/aws", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn gcp_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = GCPIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_gcp_integration()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list GCP integrations: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn gcp_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v1/integration/gcp", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn azure_list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = AzureIntegrationAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_azure_integration()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list Azure integrations: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn azure_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v1/integration/azure", &[]).await?;
    crate::formatter::output(cfg, &data)
}

// ---------------------------------------------------------------------------
// OCI tenancy management
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
fn make_oci_api(cfg: &Config) -> OCIIntegrationAPI {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    OCIIntegrationAPI::with_client_and_config(dd_cfg, dd_client)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn oci_tenancies_list(cfg: &Config) -> Result<()> {
    let api = make_oci_api(cfg);
    let resp = api
        .get_tenancy_configs()
        .await
        .map_err(|e| anyhow::anyhow!("failed to list OCI tenancies: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn oci_tenancies_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/integration/oci/tenancy_configs", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn oci_tenancies_get(cfg: &Config, tenancy_id: &str) -> Result<()> {
    let api = make_oci_api(cfg);
    let resp = api
        .get_tenancy_config(tenancy_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to get OCI tenancy: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn oci_tenancies_get(cfg: &Config, tenancy_id: &str) -> Result<()> {
    let data = crate::api::get(
        cfg,
        &format!("/api/v2/integration/oci/tenancy_configs/{tenancy_id}"),
        &[],
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn oci_tenancies_create(cfg: &Config, file: &str) -> Result<()> {
    let api = make_oci_api(cfg);
    let body: CreateTenancyConfigRequest = crate::util::read_json_file(file)?;
    let resp = api
        .create_tenancy_config(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create OCI tenancy: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn oci_tenancies_create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/integration/oci/tenancy_configs", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn oci_tenancies_update(cfg: &Config, tenancy_id: &str, file: &str) -> Result<()> {
    let api = make_oci_api(cfg);
    let body: UpdateTenancyConfigRequest = crate::util::read_json_file(file)?;
    let resp = api
        .update_tenancy_config(tenancy_id.to_string(), body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update OCI tenancy: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn oci_tenancies_update(cfg: &Config, tenancy_id: &str, file: &str) -> Result<()> {
    let body: serde_json::Value = crate::util::read_json_file(file)?;
    let data = crate::api::put(
        cfg,
        &format!("/api/v2/integration/oci/tenancy_configs/{tenancy_id}"),
        &body,
    )
    .await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn oci_tenancies_delete(cfg: &Config, tenancy_id: &str) -> Result<()> {
    let api = make_oci_api(cfg);
    api.delete_tenancy_config(tenancy_id.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete OCI tenancy: {e:?}"))?;
    println!("OCI tenancy '{tenancy_id}' deleted.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn oci_tenancies_delete(cfg: &Config, tenancy_id: &str) -> Result<()> {
    crate::api::delete(
        cfg,
        &format!("/api/v2/integration/oci/tenancy_configs/{tenancy_id}"),
    )
    .await?;
    println!("OCI tenancy '{tenancy_id}' deleted.");
    Ok(())
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn oci_products_list(cfg: &Config, product_keys: &str) -> Result<()> {
    let api = make_oci_api(cfg);
    let resp = api
        .list_tenancy_products(product_keys.to_string())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list OCI products: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn oci_products_list(cfg: &Config, product_keys: &str) -> Result<()> {
    let query = vec![("product_keys", product_keys.to_string())];
    let data = crate::api::get(cfg, "/api/v2/integration/oci/tenancy_products", &query).await?;
    crate::formatter::output(cfg, &data)
}
