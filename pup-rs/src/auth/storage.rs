use anyhow::{Context, Result};
use std::path::PathBuf;

use super::types::{ClientCredentials, TokenSet};

// ---------------------------------------------------------------------------
// Storage trait
// ---------------------------------------------------------------------------

pub trait Storage: Send + Sync {
    fn backend_type(&self) -> BackendType;
    fn storage_location(&self) -> String;

    fn save_tokens(&self, site: &str, tokens: &TokenSet) -> Result<()>;
    fn load_tokens(&self, site: &str) -> Result<Option<TokenSet>>;
    fn delete_tokens(&self, site: &str) -> Result<()>;

    fn save_client_credentials(&self, site: &str, creds: &ClientCredentials) -> Result<()>;
    fn load_client_credentials(&self, site: &str) -> Result<Option<ClientCredentials>>;
    fn delete_client_credentials(&self, site: &str) -> Result<()>;
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum BackendType {
    Keychain,
    File,
}

impl std::fmt::Display for BackendType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            BackendType::Keychain => write!(f, "keychain"),
            BackendType::File => write!(f, "file"),
        }
    }
}

// ---------------------------------------------------------------------------
// File storage (~/.config/pup/)
// ---------------------------------------------------------------------------

pub struct FileStorage {
    base_dir: PathBuf,
}

impl FileStorage {
    pub fn new() -> Result<Self> {
        let base_dir = crate::config::config_dir()
            .context("could not determine config directory")?;
        std::fs::create_dir_all(&base_dir)
            .with_context(|| format!("failed to create config dir: {}", base_dir.display()))?;
        Ok(Self { base_dir })
    }
}

impl Storage for FileStorage {
    fn backend_type(&self) -> BackendType {
        BackendType::File
    }

    fn storage_location(&self) -> String {
        self.base_dir.display().to_string()
    }

    fn save_tokens(&self, site: &str, tokens: &TokenSet) -> Result<()> {
        let path = self.base_dir.join(format!("tokens_{}.json", sanitize(site)));
        let json = serde_json::to_string_pretty(tokens)?;
        std::fs::write(&path, json)
            .with_context(|| format!("failed to write tokens: {}", path.display()))?;
        // Restrict permissions on Unix
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            std::fs::set_permissions(&path, std::fs::Permissions::from_mode(0o600))?;
        }
        Ok(())
    }

    fn load_tokens(&self, site: &str) -> Result<Option<TokenSet>> {
        let path = self.base_dir.join(format!("tokens_{}.json", sanitize(site)));
        match std::fs::read_to_string(&path) {
            Ok(json) => Ok(Some(serde_json::from_str(&json)?)),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
            Err(e) => Err(e.into()),
        }
    }

    fn delete_tokens(&self, site: &str) -> Result<()> {
        let path = self.base_dir.join(format!("tokens_{}.json", sanitize(site)));
        match std::fs::remove_file(&path) {
            Ok(()) => Ok(()),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(()),
            Err(e) => Err(e.into()),
        }
    }

    fn save_client_credentials(&self, site: &str, creds: &ClientCredentials) -> Result<()> {
        let path = self.base_dir.join(format!("client_{}.json", sanitize(site)));
        let json = serde_json::to_string_pretty(creds)?;
        std::fs::write(&path, json)
            .with_context(|| format!("failed to write credentials: {}", path.display()))?;
        #[cfg(unix)]
        {
            use std::os::unix::fs::PermissionsExt;
            std::fs::set_permissions(&path, std::fs::Permissions::from_mode(0o600))?;
        }
        Ok(())
    }

    fn load_client_credentials(&self, site: &str) -> Result<Option<ClientCredentials>> {
        let path = self.base_dir.join(format!("client_{}.json", sanitize(site)));
        match std::fs::read_to_string(&path) {
            Ok(json) => Ok(Some(serde_json::from_str(&json)?)),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
            Err(e) => Err(e.into()),
        }
    }

    fn delete_client_credentials(&self, site: &str) -> Result<()> {
        let path = self.base_dir.join(format!("client_{}.json", sanitize(site)));
        match std::fs::remove_file(&path) {
            Ok(()) => Ok(()),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(()),
            Err(e) => Err(e.into()),
        }
    }
}

// ---------------------------------------------------------------------------
// Keychain storage (via keyring crate)
// ---------------------------------------------------------------------------

pub struct KeychainStorage;

const SERVICE_NAME: &str = "pup";

impl KeychainStorage {
    pub fn new() -> Result<Self> {
        // Test keychain availability by attempting an operation
        let entry = keyring::Entry::new(SERVICE_NAME, "__pup_test__")?;
        // Try a read — NotFound is fine, other errors mean keychain is unavailable
        match entry.get_password() {
            Ok(_) | Err(keyring::Error::NoEntry) => Ok(Self),
            Err(e) => Err(anyhow::anyhow!("keychain not available: {e}")),
        }
    }
}

impl Storage for KeychainStorage {
    fn backend_type(&self) -> BackendType {
        BackendType::Keychain
    }

    fn storage_location(&self) -> String {
        "OS keychain".to_string()
    }

    fn save_tokens(&self, site: &str, tokens: &TokenSet) -> Result<()> {
        let key = format!("tokens_{}", sanitize(site));
        let entry = keyring::Entry::new(SERVICE_NAME, &key)?;
        let json = serde_json::to_string(tokens)?;
        entry.set_password(&json)?;
        Ok(())
    }

    fn load_tokens(&self, site: &str) -> Result<Option<TokenSet>> {
        let key = format!("tokens_{}", sanitize(site));
        let entry = keyring::Entry::new(SERVICE_NAME, &key)?;
        match entry.get_password() {
            Ok(json) => Ok(Some(serde_json::from_str(&json)?)),
            Err(keyring::Error::NoEntry) => Ok(None),
            Err(e) => Err(e.into()),
        }
    }

    fn delete_tokens(&self, site: &str) -> Result<()> {
        let key = format!("tokens_{}", sanitize(site));
        let entry = keyring::Entry::new(SERVICE_NAME, &key)?;
        match entry.delete_credential() {
            Ok(()) | Err(keyring::Error::NoEntry) => Ok(()),
            Err(e) => Err(e.into()),
        }
    }

    fn save_client_credentials(&self, site: &str, creds: &ClientCredentials) -> Result<()> {
        let key = format!("client_{}", sanitize(site));
        let entry = keyring::Entry::new(SERVICE_NAME, &key)?;
        let json = serde_json::to_string(creds)?;
        entry.set_password(&json)?;
        Ok(())
    }

    fn load_client_credentials(&self, site: &str) -> Result<Option<ClientCredentials>> {
        let key = format!("client_{}", sanitize(site));
        let entry = keyring::Entry::new(SERVICE_NAME, &key)?;
        match entry.get_password() {
            Ok(json) => Ok(Some(serde_json::from_str(&json)?)),
            Err(keyring::Error::NoEntry) => Ok(None),
            Err(e) => Err(e.into()),
        }
    }

    fn delete_client_credentials(&self, site: &str) -> Result<()> {
        let key = format!("client_{}", sanitize(site));
        let entry = keyring::Entry::new(SERVICE_NAME, &key)?;
        match entry.delete_credential() {
            Ok(()) | Err(keyring::Error::NoEntry) => Ok(()),
            Err(e) => Err(e.into()),
        }
    }
}

// ---------------------------------------------------------------------------
// Factory — auto-detect backend, with fallback
// ---------------------------------------------------------------------------

use std::sync::Mutex;

static STORAGE: Mutex<Option<Box<dyn Storage>>> = Mutex::new(None);

pub fn get_storage() -> Result<&'static Mutex<Option<Box<dyn Storage>>>> {
    let mut guard = STORAGE.lock().unwrap();
    if guard.is_none() {
        let backend = detect_backend();
        *guard = Some(backend);
    }
    drop(guard);
    Ok(&STORAGE)
}

fn detect_backend() -> Box<dyn Storage> {
    // Check DD_TOKEN_STORAGE env var
    if let Ok(val) = std::env::var("DD_TOKEN_STORAGE") {
        match val.as_str() {
            "file" => return Box::new(FileStorage::new().expect("failed to create file storage")),
            "keychain" => {
                return Box::new(KeychainStorage::new().expect("keychain not available"))
            }
            _ => eprintln!("Warning: unknown DD_TOKEN_STORAGE={val:?}, auto-detecting"),
        }
    }

    // Try keychain first
    match KeychainStorage::new() {
        Ok(ks) => Box::new(ks),
        Err(_) => {
            eprintln!("Warning: OS keychain not available, using file storage (~/.config/pup/)");
            Box::new(FileStorage::new().expect("failed to create file storage"))
        }
    }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

fn sanitize(site: &str) -> String {
    site.chars()
        .map(|c| if c.is_alphanumeric() { c } else { '_' })
        .collect()
}
