use anyhow::{bail, Result};
use serde::Deserialize;
use std::path::PathBuf;

/// Runtime configuration with precedence: flag > env > file > default.
pub struct Config {
    pub api_key: Option<String>,
    pub app_key: Option<String>,
    pub access_token: Option<String>,
    pub site: String,
    pub output_format: OutputFormat,
    pub auto_approve: bool,
    pub agent_mode: bool,
}

#[derive(Clone, Debug, PartialEq)]
pub enum OutputFormat {
    Json,
    Table,
    Yaml,
}

impl std::fmt::Display for OutputFormat {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            OutputFormat::Json => write!(f, "json"),
            OutputFormat::Table => write!(f, "table"),
            OutputFormat::Yaml => write!(f, "yaml"),
        }
    }
}

impl std::str::FromStr for OutputFormat {
    type Err = anyhow::Error;
    fn from_str(s: &str) -> Result<Self> {
        match s.to_lowercase().as_str() {
            "json" => Ok(OutputFormat::Json),
            "table" => Ok(OutputFormat::Table),
            "yaml" => Ok(OutputFormat::Yaml),
            _ => bail!("invalid output format: {s:?} (expected json, table, or yaml)"),
        }
    }
}

/// Config file structure (~/.config/pup/config.yaml)
#[derive(Deserialize, Default)]
struct FileConfig {
    api_key: Option<String>,
    app_key: Option<String>,
    access_token: Option<String>,
    site: Option<String>,
    output: Option<String>,
    auto_approve: Option<bool>,
}

impl Config {
    /// Load configuration with precedence: flag overrides > env > file > defaults.
    /// Flag overrides are applied by the caller after this returns.
    pub fn from_env() -> Result<Self> {
        let file_cfg = load_config_file().unwrap_or_default();

        let cfg = Config {
            api_key: env_or("DD_API_KEY", file_cfg.api_key),
            app_key: env_or("DD_APP_KEY", file_cfg.app_key),
            access_token: env_or("DD_ACCESS_TOKEN", file_cfg.access_token),
            site: env_or("DD_SITE", file_cfg.site).unwrap_or_else(|| "datadoghq.com".into()),
            output_format: env_or("DD_OUTPUT", file_cfg.output)
                .and_then(|s| s.parse().ok())
                .unwrap_or(OutputFormat::Json),
            auto_approve: env_bool("DD_AUTO_APPROVE")
                || env_bool("DD_CLI_AUTO_APPROVE")
                || file_cfg.auto_approve.unwrap_or(false),
            agent_mode: false, // set by caller from --agent flag or useragent detection
        };

        Ok(cfg)
    }

    /// Validate that sufficient auth credentials are configured.
    pub fn validate_auth(&self) -> Result<()> {
        if self.access_token.is_none() && (self.api_key.is_none() || self.app_key.is_none()) {
            bail!(
                "authentication required: set DD_ACCESS_TOKEN for bearer auth, \
                 run 'pup auth login' for OAuth2, \
                 or set DD_API_KEY and DD_APP_KEY for API key auth"
            );
        }
        Ok(())
    }

    pub fn has_api_keys(&self) -> bool {
        self.api_key.is_some() && self.app_key.is_some()
    }

    pub fn has_bearer_token(&self) -> bool {
        self.access_token.is_some()
    }

    /// Returns the API host (e.g., "api.datadoghq.com").
    pub fn api_host(&self) -> String {
        if self.site.contains("oncall") {
            self.site.clone()
        } else {
            format!("api.{}", self.site)
        }
    }
}

/// Config file path: ~/.config/pup/config.yaml
pub fn config_dir() -> Option<PathBuf> {
    dirs::config_dir().map(|d| d.join("pup"))
}

fn load_config_file() -> Option<FileConfig> {
    let path = config_dir()?.join("config.yaml");
    let contents = std::fs::read_to_string(path).ok()?;
    serde_yaml::from_str(&contents).ok()
}

fn env_or(key: &str, fallback: Option<String>) -> Option<String> {
    std::env::var(key)
        .ok()
        .filter(|s| !s.is_empty())
        .or(fallback)
}

fn env_bool(key: &str) -> bool {
    matches!(
        std::env::var(key)
            .unwrap_or_default()
            .to_lowercase()
            .as_str(),
        "true" | "1"
    )
}
