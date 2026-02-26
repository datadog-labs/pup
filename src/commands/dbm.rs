use anyhow::Result;

use crate::client;
use crate::config::Config;
use crate::formatter;

/// Query DBM explain plans via the logs-analytics v1 endpoint.
///
/// Uses: POST /api/v1/logs-analytics/list?type=databasequery
/// Requires DD_API_KEY + DD_APP_KEY (OAuth2 not supported).
pub async fn explain_plans(
    cfg: &Config,
    query: String,
    source: Option<String>,
    limit: i32,
) -> Result<()> {
    if !cfg.has_api_keys() {
        anyhow::bail!(
            "dbm explain-plans requires API key authentication (DD_API_KEY + DD_APP_KEY).\n\
             This endpoint does not support bearer token auth."
        );
    }

    let full_query = match source {
        Some(src) => format!("{query} source:{src}"),
        None => query,
    };

    let body = serde_json::json!({
        "list": {
            "indexes": ["databasequery"],
            "search": { "query": full_query },
            "limit": limit
        }
    });

    let data = client::raw_post(
        cfg,
        "/api/v1/logs-analytics/list?type=databasequery",
        body,
    )
    .await?;

    formatter::output(cfg, &data)
}
