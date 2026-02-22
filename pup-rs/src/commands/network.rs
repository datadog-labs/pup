use anyhow::Result;

use crate::config::Config;
use crate::formatter;

pub fn list() -> Result<()> {
    anyhow::bail!("network commands are not yet implemented (API endpoints pending)")
}

pub async fn flows_list(cfg: &Config) -> Result<()> {
    let placeholder = serde_json::json!({
        "data": [],
        "meta": {
            "message": "Network flows list - API endpoint implementation pending"
        }
    });
    formatter::output(cfg, &placeholder)
}

pub async fn devices_list(cfg: &Config) -> Result<()> {
    let placeholder = serde_json::json!({
        "data": [],
        "meta": {
            "message": "Network devices list - API endpoint implementation pending"
        }
    });
    formatter::output(cfg, &placeholder)
}
