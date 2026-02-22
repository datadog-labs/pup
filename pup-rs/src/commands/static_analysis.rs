use anyhow::Result;

use crate::config::Config;
use crate::formatter;

// ---------------------------------------------------------------------------
// Static Analysis commands
//
// These commands are placeholders. The Datadog static analysis API endpoints
// are not yet available in the typed Rust client, and no Go reference
// implementation exists. When the API becomes available, these should be
// replaced with real implementations.
// ---------------------------------------------------------------------------

pub async fn ast_list(cfg: &Config) -> Result<()> {
    let placeholder = serde_json::json!({
        "data": [],
        "meta": {
            "message": "static analysis AST list - not yet implemented"
        }
    });
    formatter::output(cfg, &placeholder)
}

pub async fn ast_get(cfg: &Config, id: &str) -> Result<()> {
    let placeholder = serde_json::json!({
        "data": null,
        "meta": {
            "message": format!("static analysis AST get ({id}) - not yet implemented")
        }
    });
    formatter::output(cfg, &placeholder)
}

pub async fn custom_rulesets_list(cfg: &Config) -> Result<()> {
    let placeholder = serde_json::json!({
        "data": [],
        "meta": {
            "message": "static analysis custom rulesets list - not yet implemented"
        }
    });
    formatter::output(cfg, &placeholder)
}

pub async fn custom_rulesets_get(cfg: &Config, id: &str) -> Result<()> {
    let placeholder = serde_json::json!({
        "data": null,
        "meta": {
            "message": format!("static analysis custom rulesets get ({id}) - not yet implemented")
        }
    });
    formatter::output(cfg, &placeholder)
}

pub async fn sca_list(cfg: &Config) -> Result<()> {
    let placeholder = serde_json::json!({
        "data": [],
        "meta": {
            "message": "static analysis SCA list - not yet implemented"
        }
    });
    formatter::output(cfg, &placeholder)
}

pub async fn sca_get(cfg: &Config, id: &str) -> Result<()> {
    let placeholder = serde_json::json!({
        "data": null,
        "meta": {
            "message": format!("static analysis SCA get ({id}) - not yet implemented")
        }
    });
    formatter::output(cfg, &placeholder)
}

pub async fn coverage_list(cfg: &Config) -> Result<()> {
    let placeholder = serde_json::json!({
        "data": [],
        "meta": {
            "message": "static analysis coverage list - not yet implemented"
        }
    });
    formatter::output(cfg, &placeholder)
}

pub async fn coverage_get(cfg: &Config, id: &str) -> Result<()> {
    let placeholder = serde_json::json!({
        "data": null,
        "meta": {
            "message": format!("static analysis coverage get ({id}) - not yet implemented")
        }
    });
    formatter::output(cfg, &placeholder)
}
