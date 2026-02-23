use anyhow::Result;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_product_analytics::ProductAnalyticsAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::ProductAnalyticsServerSideEventItem;

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

#[cfg(not(target_arch = "wasm32"))]
pub async fn events_send(cfg: &Config, file: &str) -> Result<()> {
    let body: ProductAnalyticsServerSideEventItem = util::read_json_file(file)?;
    let dd_cfg = client::make_dd_config(cfg);
    let api = match client::make_bearer_client(cfg) {
        Some(c) => ProductAnalyticsAPI::with_client_and_config(dd_cfg, c),
        None => ProductAnalyticsAPI::with_config(dd_cfg),
    };
    let resp = api
        .submit_product_analytics_event(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to send product analytics event: {e:?}"))?;
    formatter::output(cfg, &resp)
}

#[cfg(target_arch = "wasm32")]
pub async fn events_send(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/events", &body).await?;
    crate::formatter::output(cfg, &data)
}

// ---- Analytics ----

pub async fn analytics_scalar(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/analytics/scalar", &body).await?;
    formatter::output(cfg, &data)
}

pub async fn analytics_timeseries(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data =
        crate::api::post(cfg, "/api/v2/product-analytics/analytics/timeseries", &body).await?;
    formatter::output(cfg, &data)
}

// ---- Journey ----

pub async fn journey_funnel(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/journey/funnel", &body).await?;
    formatter::output(cfg, &data)
}

pub async fn journey_timeseries(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/journey/timeseries", &body).await?;
    formatter::output(cfg, &data)
}

pub async fn journey_scalar(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/journey/scalar", &body).await?;
    formatter::output(cfg, &data)
}

pub async fn journey_list(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/journey/list", &body).await?;
    formatter::output(cfg, &data)
}

pub async fn journey_drop_off_analysis(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(
        cfg,
        "/api/v2/product-analytics/journey/drop_off_analysis",
        &body,
    )
    .await?;
    formatter::output(cfg, &data)
}

// ---- Retention ----

pub async fn retention_grid(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/retention/grid", &body).await?;
    formatter::output(cfg, &data)
}

pub async fn retention_timeseries(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data =
        crate::api::post(cfg, "/api/v2/product-analytics/retention/timeseries", &body).await?;
    formatter::output(cfg, &data)
}

pub async fn retention_scalar(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/retention/scalar", &body).await?;
    formatter::output(cfg, &data)
}

pub async fn retention_list(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/retention/list", &body).await?;
    formatter::output(cfg, &data)
}

pub async fn retention_meta(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/retention/meta", &body).await?;
    formatter::output(cfg, &data)
}

// ---- Sankey ----

pub async fn sankey(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/sankey", &body).await?;
    formatter::output(cfg, &data)
}

// ---- Segment ----

pub async fn segment_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/product-analytics/segment", &[]).await?;
    formatter::output(cfg, &data)
}

pub async fn segment_create(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/segment", &body).await?;
    formatter::output(cfg, &data)
}

pub async fn segment_create_static(cfg: &Config, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let data = crate::api::post(cfg, "/api/v2/product-analytics/segment/static", &body).await?;
    formatter::output(cfg, &data)
}

pub async fn segment_get(cfg: &Config, id: &str) -> Result<()> {
    let path = format!("/api/v2/product-analytics/segment/{id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    formatter::output(cfg, &data)
}

pub async fn segment_update(cfg: &Config, id: &str, file: &str) -> Result<()> {
    let body: serde_json::Value = util::read_json_file(file)?;
    let path = format!("/api/v2/product-analytics/segment/{id}");
    let data = crate::api::put(cfg, &path, &body).await?;
    formatter::output(cfg, &data)
}

pub async fn segment_delete(cfg: &Config, id: &str) -> Result<()> {
    let path = format!("/api/v2/product-analytics/segment/{id}");
    crate::api::delete(cfg, &path).await?;
    println!("Successfully deleted segment {id}");
    Ok(())
}

pub async fn segment_templates_list(cfg: &Config) -> Result<()> {
    let data = crate::api::get(cfg, "/api/v2/product-analytics/segment/templates", &[]).await?;
    formatter::output(cfg, &data)
}

pub async fn segment_templates_get(cfg: &Config, id: &str) -> Result<()> {
    let path = format!("/api/v2/product-analytics/segment/templates/{id}");
    let data = crate::api::get(cfg, &path, &[]).await?;
    formatter::output(cfg, &data)
}
