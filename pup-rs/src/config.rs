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
        if let Ok(mock) = std::env::var("PUP_MOCK_SERVER") {
            let host = mock
                .trim_start_matches("http://")
                .trim_start_matches("https://");
            return host.to_string();
        }
        if self.site.contains("oncall") {
            self.site.clone()
        } else {
            format!("api.{}", self.site)
        }
    }

    /// Returns the full API base URL (e.g., "https://api.datadoghq.com").
    /// Respects PUP_MOCK_SERVER for testing.
    pub fn api_base_url(&self) -> String {
        if let Ok(mock) = std::env::var("PUP_MOCK_SERVER") {
            return mock;
        }
        format!("https://{}", self.api_host())
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

#[cfg(test)]
mod tests {
    use super::*;

    fn make_cfg(api_key: Option<&str>, app_key: Option<&str>, token: Option<&str>) -> Config {
        Config {
            api_key: api_key.map(String::from),
            app_key: app_key.map(String::from),
            access_token: token.map(String::from),
            site: "datadoghq.com".into(),
            output_format: OutputFormat::Json,
            auto_approve: false,
            agent_mode: false,
        }
    }

    #[test]
    fn test_output_format_parse() {
        assert_eq!("json".parse::<OutputFormat>().unwrap(), OutputFormat::Json);
        assert_eq!("JSON".parse::<OutputFormat>().unwrap(), OutputFormat::Json);
        assert_eq!("table".parse::<OutputFormat>().unwrap(), OutputFormat::Table);
        assert_eq!("yaml".parse::<OutputFormat>().unwrap(), OutputFormat::Yaml);
        assert!("xml".parse::<OutputFormat>().is_err());
    }

    #[test]
    fn test_output_format_display() {
        assert_eq!(OutputFormat::Json.to_string(), "json");
        assert_eq!(OutputFormat::Table.to_string(), "table");
        assert_eq!(OutputFormat::Yaml.to_string(), "yaml");
    }

    #[test]
    fn test_validate_auth_api_keys() {
        let cfg = make_cfg(Some("key"), Some("app"), None);
        assert!(cfg.validate_auth().is_ok());
    }

    #[test]
    fn test_validate_auth_bearer() {
        let cfg = make_cfg(None, None, Some("token"));
        assert!(cfg.validate_auth().is_ok());
    }

    #[test]
    fn test_validate_auth_none() {
        let cfg = make_cfg(None, None, None);
        assert!(cfg.validate_auth().is_err());
    }

    #[test]
    fn test_validate_auth_partial_keys() {
        let cfg = make_cfg(Some("key"), None, None);
        assert!(cfg.validate_auth().is_err());
    }

    #[test]
    fn test_has_api_keys() {
        assert!(make_cfg(Some("k"), Some("a"), None).has_api_keys());
        assert!(!make_cfg(Some("k"), None, None).has_api_keys());
        assert!(!make_cfg(None, None, None).has_api_keys());
    }

    #[test]
    fn test_has_bearer_token() {
        assert!(make_cfg(None, None, Some("t")).has_bearer_token());
        assert!(!make_cfg(None, None, None).has_bearer_token());
    }

    #[test]
    fn test_api_host_standard() {
        let cfg = make_cfg(None, None, Some("t"));
        assert_eq!(cfg.api_host(), "api.datadoghq.com");
    }

    #[test]
    fn test_api_host_eu() {
        let mut cfg = make_cfg(None, None, Some("t"));
        cfg.site = "datadoghq.eu".into();
        assert_eq!(cfg.api_host(), "api.datadoghq.eu");
    }

    #[test]
    fn test_api_host_oncall() {
        let mut cfg = make_cfg(None, None, Some("t"));
        cfg.site = "navy.oncall.datadoghq.com".into();
        assert_eq!(cfg.api_host(), "navy.oncall.datadoghq.com");
    }

    #[test]
    fn test_env_or_with_fallback() {
        assert_eq!(
            env_or("__PUP_TEST_NONEXISTENT__", Some("fallback".into())),
            Some("fallback".into())
        );
    }

    #[test]
    fn test_env_or_no_fallback() {
        assert_eq!(env_or("__PUP_TEST_NONEXISTENT__", None), None);
    }
}
