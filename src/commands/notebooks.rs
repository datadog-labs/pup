use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::api_notebooks::{ListNotebooksOptionalParams, NotebooksAPI};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV1::model::{NotebookCreateRequest, NotebookUpdateRequest};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn list(cfg: &Config) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = NotebooksAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .list_notebooks(ListNotebooksOptionalParams::default())
        .await
        .map_err(|e| anyhow::anyhow!("failed to list notebooks: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v1/notebooks", &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn get(cfg: &Config, notebook_id: i64) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = NotebooksAPI::with_client_and_config(dd_cfg, dd_client);
    let resp = api
        .get_notebook(notebook_id)
        .await
        .map_err(|e| anyhow::anyhow!("failed to get notebook: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn get(cfg: &Config, notebook_id: i64) -> Result<()> {
    let data = crate::api::get(cfg, &format!("/api/v1/notebooks/{notebook_id}"), &[]).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn delete(cfg: &Config, notebook_id: i64) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = NotebooksAPI::with_client_and_config(dd_cfg, dd_client);
    api.delete_notebook(notebook_id)
        .await
        .map_err(|e| anyhow::anyhow!("failed to delete notebook: {e:?}"))?;
    println!("Successfully deleted notebook {notebook_id}");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn delete(cfg: &Config, notebook_id: i64) -> Result<()> {
    crate::api::delete(cfg, &format!("/api/v1/notebooks/{notebook_id}")).await?;
    println!("Successfully deleted notebook {notebook_id}");
    Ok(())
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn create(cfg: &Config, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = NotebooksAPI::with_client_and_config(dd_cfg, dd_client);
    let body: NotebookCreateRequest = util::read_json_file(file)?;
    let resp = api
        .create_notebook(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to create notebook: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v1/notebooks", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn update(cfg: &Config, notebook_id: i64, file: &str) -> Result<()> {
    let dd_cfg = client::make_dd_config(cfg);
    let dd_client = client::make_dd_client(cfg);
    let api = NotebooksAPI::with_client_and_config(dd_cfg, dd_client);
    let body: NotebookUpdateRequest = util::read_json_file(file)?;
    let resp = api
        .update_notebook(notebook_id, body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to update notebook: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn update(cfg: &Config, notebook_id: i64, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::put(cfg, &format!("/api/v1/notebooks/{notebook_id}"), &body).await?;
    crate::formatter::output(cfg, &data)
}
