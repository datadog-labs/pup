use anyhow::{bail, Result};
#[cfg(not(feature = "browser"))]
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
#[cfg(not(feature = "browser"))]
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
    /// Load configuration with precedence: flag overrides > env > file > keychain > defaults.
    /// Flag overrides are applied by the caller after this returns.
    #[cfg(not(feature = "browser"))]
    pub fn from_env() -> Result<Self> {
        let file_cfg = load_config_file().unwrap_or_default();

        let access_token = env_or("DD_ACCESS_TOKEN", file_cfg.access_token);
        let site = env_or("DD_SITE", file_cfg.site).unwrap_or_else(|| "datadoghq.com".into());

        // If no token from env/file, try loading from keychain/storage (where `pup auth login` saves)
        #[cfg(not(target_arch = "wasm32"))]
        let access_token = access_token.or_else(|| load_token_from_storage(&site));

        let cfg = Config {
            api_key: env_or("DD_API_KEY", file_cfg.api_key),
            app_key: env_or("DD_APP_KEY", file_cfg.app_key),
            access_token,
            site,
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

    /// Create configuration from explicit parameters (no env vars or filesystem).
    /// Used by the browser WASM build where `std::env` is unavailable.
    #[cfg(feature = "browser")]
    pub fn from_params(
        site: String,
        access_token: Option<String>,
        api_key: Option<String>,
        app_key: Option<String>,
    ) -> Self {
        Config {
            api_key,
            app_key,
            access_token,
            site,
            output_format: OutputFormat::Json,
            auto_approve: false,
            agent_mode: false,
        }
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

    /// Validate that both DD_API_KEY and DD_APP_KEY are configured.
    /// Used for endpoints that require API key auth and do not accept OAuth2 tokens.
    pub fn validate_api_and_app_keys(&self) -> Result<()> {
        if self.api_key.is_none() || self.app_key.is_none() {
            bail!(
                "this command requires both DD_API_KEY and DD_APP_KEY — \
                 OAuth2 bearer tokens are not supported here"
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
        #[cfg(not(feature = "browser"))]
        {
            if let Ok(mock) = std::env::var("PUP_MOCK_SERVER") {
                let host = mock
                    .trim_start_matches("http://")
                    .trim_start_matches("https://");
                return host.to_string();
            }
        }
        if self.site.contains("oncall") {
            self.site.clone()
        } else {
            format!("api.{}", self.site)
        }
    }

    /// Returns the full API base URL (e.g., "https://api.datadoghq.com").
    /// Respects PUP_MOCK_SERVER for testing (native/WASI only).
    pub fn api_base_url(&self) -> String {
        #[cfg(not(feature = "browser"))]
        {
            if let Ok(mock) = std::env::var("PUP_MOCK_SERVER") {
                return mock;
            }
        }
        format!("https://{}", self.api_host())
    }
}

/// Config file path: ~/.config/pup/config.yaml
#[cfg(not(target_arch = "wasm32"))]
pub fn config_dir() -> Option<PathBuf> {
    dirs::config_dir().map(|d| d.join("pup"))
}

/// WASI: use PUP_CONFIG_DIR env var or return None
#[cfg(all(target_arch = "wasm32", not(feature = "browser")))]
pub fn config_dir() -> Option<PathBuf> {
    std::env::var("PUP_CONFIG_DIR").ok().map(PathBuf::from)
}

/// Browser WASM: no filesystem access
#[cfg(feature = "browser")]
pub fn config_dir() -> Option<PathBuf> {
    None
}

#[cfg(not(feature = "browser"))]
fn load_config_file() -> Option<FileConfig> {
    let path = config_dir()?.join("config.yaml");
    let contents = std::fs::read_to_string(path).ok()?;
    serde_yaml::from_str(&contents).ok()
}

/// Try to load a valid (non-expired) access token from keychain/file storage.
/// Returns None silently on any error — callers fall through to other auth methods.
#[cfg(all(not(feature = "browser"), not(target_arch = "wasm32")))]
fn load_token_from_storage(site: &str) -> Option<String> {
    let guard = crate::auth::storage::get_storage().ok()?;
    let lock = guard.lock().ok()?;
    let store = lock.as_ref()?;
    let tokens = store.load_tokens(site).ok()??;
    if tokens.is_expired() {
        return None;
    }
    Some(tokens.access_token)
}

#[cfg(not(feature = "browser"))]
fn env_or(key: &str, fallback: Option<String>) -> Option<String> {
    std::env::var(key)
        .ok()
        .filter(|s| !s.is_empty())
        .or(fallback)
}

#[cfg(not(feature = "browser"))]
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
    use crate::test_utils::ENV_LOCK;

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
        assert_eq!(
            "table".parse::<OutputFormat>().unwrap(),
            OutputFormat::Table
        );
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
    fn test_validate_api_and_app_keys_ok() {
        let cfg = make_cfg(Some("key"), Some("app"), None);
        assert!(cfg.validate_api_and_app_keys().is_ok());
    }

    #[test]
    fn test_validate_api_and_app_keys_bearer_only_fails() {
        let cfg = make_cfg(None, None, Some("token"));
        assert!(cfg.validate_api_and_app_keys().is_err());
    }

    #[test]
    fn test_validate_api_and_app_keys_missing_app_key_fails() {
        let cfg = make_cfg(Some("key"), None, None);
        assert!(cfg.validate_api_and_app_keys().is_err());
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
        let _guard = ENV_LOCK.lock().unwrap_or_else(|p| p.into_inner());
        std::env::remove_var("PUP_MOCK_SERVER");
        let cfg = make_cfg(None, None, Some("t"));
        assert_eq!(cfg.api_host(), "api.datadoghq.com");
    }

    #[test]
    fn test_api_host_eu() {
        let _guard = ENV_LOCK.lock().unwrap_or_else(|p| p.into_inner());
        std::env::remove_var("PUP_MOCK_SERVER");
        let mut cfg = make_cfg(None, None, Some("t"));
        cfg.site = "datadoghq.eu".into();
        assert_eq!(cfg.api_host(), "api.datadoghq.eu");
    }

    #[test]
    fn test_api_host_oncall() {
        let _guard = ENV_LOCK.lock().unwrap_or_else(|p| p.into_inner());
        std::env::remove_var("PUP_MOCK_SERVER");
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

    #[test]
    fn test_api_base_url_standard() {
        let _guard = ENV_LOCK.lock().unwrap_or_else(|p| p.into_inner());
        std::env::remove_var("PUP_MOCK_SERVER");
        let cfg = make_cfg(None, None, Some("t"));
        assert_eq!(cfg.api_base_url(), "https://api.datadoghq.com");
    }

    #[test]
    fn test_api_base_url_eu() {
        let _guard = ENV_LOCK.lock().unwrap_or_else(|p| p.into_inner());
        std::env::remove_var("PUP_MOCK_SERVER");
        let mut cfg = make_cfg(None, None, Some("t"));
        cfg.site = "datadoghq.eu".into();
        assert_eq!(cfg.api_base_url(), "https://api.datadoghq.eu");
    }

    #[test]
    fn test_api_base_url_oncall() {
        let _guard = ENV_LOCK.lock().unwrap_or_else(|p| p.into_inner());
        std::env::remove_var("PUP_MOCK_SERVER");
        let mut cfg = make_cfg(None, None, Some("t"));
        cfg.site = "navy.oncall.datadoghq.com".into();
        assert_eq!(cfg.api_base_url(), "https://navy.oncall.datadoghq.com");
    }

    #[test]
    fn test_api_base_url_mock_server() {
        let _guard = ENV_LOCK.lock().unwrap_or_else(|p| p.into_inner());
        std::env::set_var("PUP_MOCK_SERVER", "http://127.0.0.1:1234");
        let cfg = make_cfg(None, None, Some("t"));
        assert_eq!(cfg.api_base_url(), "http://127.0.0.1:1234");
        std::env::remove_var("PUP_MOCK_SERVER");
    }

    #[test]
    fn test_api_host_mock_server() {
        let _guard = ENV_LOCK.lock().unwrap_or_else(|p| p.into_inner());
        std::env::set_var("PUP_MOCK_SERVER", "http://127.0.0.1:5678");
        let cfg = make_cfg(None, None, Some("t"));
        assert_eq!(cfg.api_host(), "127.0.0.1:5678");
        std::env::remove_var("PUP_MOCK_SERVER");
    }

    #[test]
    fn test_env_bool_true() {
        std::env::set_var("__PUP_TEST_BOOL_TRUE__", "true");
        assert!(env_bool("__PUP_TEST_BOOL_TRUE__"));
        std::env::remove_var("__PUP_TEST_BOOL_TRUE__");
    }

    #[test]
    fn test_env_bool_one() {
        std::env::set_var("__PUP_TEST_BOOL_ONE__", "1");
        assert!(env_bool("__PUP_TEST_BOOL_ONE__"));
        std::env::remove_var("__PUP_TEST_BOOL_ONE__");
    }

    #[test]
    fn test_env_bool_false() {
        std::env::set_var("__PUP_TEST_BOOL_FALSE__", "false");
        assert!(!env_bool("__PUP_TEST_BOOL_FALSE__"));
        std::env::remove_var("__PUP_TEST_BOOL_FALSE__");
    }

    #[test]
    fn test_env_bool_missing() {
        assert!(!env_bool("__PUP_TEST_BOOL_MISSING__"));
    }

    #[test]
    fn test_config_dir_returns_path() {
        let dir = config_dir();
        // On native builds, dirs::config_dir() should return Some
        assert!(dir.is_some());
        assert!(dir.unwrap().ends_with("pup"));
    }

    #[test]
    fn test_env_or_with_env_value() {
        std::env::set_var("__PUP_TEST_ENV_OR__", "env-value");
        assert_eq!(
            env_or("__PUP_TEST_ENV_OR__", Some("fallback".into())),
            Some("env-value".into())
        );
        std::env::remove_var("__PUP_TEST_ENV_OR__");
    }

    #[test]
    fn test_env_or_empty_env_uses_fallback() {
        std::env::set_var("__PUP_TEST_ENV_EMPTY__", "");
        assert_eq!(
            env_or("__PUP_TEST_ENV_EMPTY__", Some("fallback".into())),
            Some("fallback".into())
        );
        std::env::remove_var("__PUP_TEST_ENV_EMPTY__");
    }
}
