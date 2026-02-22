use anyhow::{Context, Result};
use std::path::PathBuf;

use super::types::{ClientCredentials, TokenSet};

// ---------------------------------------------------------------------------
// Storage trait
// ---------------------------------------------------------------------------

pub trait Storage: Send + Sync {
    #[allow(dead_code)]
    fn backend_type(&self) -> BackendType;
    fn storage_location(&self) -> String;

    fn save_tokens(&self, site: &str, tokens: &TokenSet) -> Result<()>;
    fn load_tokens(&self, site: &str) -> Result<Option<TokenSet>>;
    fn delete_tokens(&self, site: &str) -> Result<()>;

    fn save_client_credentials(&self, site: &str, creds: &ClientCredentials) -> Result<()>;
    fn load_client_credentials(&self, site: &str) -> Result<Option<ClientCredentials>>;
    fn delete_client_credentials(&self, site: &str) -> Result<()>;
}

#[allow(dead_code)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum BackendType {
    Keychain,
    File,
    #[cfg(feature = "browser")]
    LocalStorage,
}

impl std::fmt::Display for BackendType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            BackendType::Keychain => write!(f, "keychain"),
            BackendType::File => write!(f, "file"),
            #[cfg(feature = "browser")]
            BackendType::LocalStorage => write!(f, "localStorage"),
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
        let base_dir =
            crate::config::config_dir().context("could not determine config directory")?;
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
        let path = self
            .base_dir
            .join(format!("tokens_{}.json", sanitize(site)));
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
        let path = self
            .base_dir
            .join(format!("tokens_{}.json", sanitize(site)));
        match std::fs::read_to_string(&path) {
            Ok(json) => Ok(Some(serde_json::from_str(&json)?)),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
            Err(e) => Err(e.into()),
        }
    }

    fn delete_tokens(&self, site: &str) -> Result<()> {
        let path = self
            .base_dir
            .join(format!("tokens_{}.json", sanitize(site)));
        match std::fs::remove_file(&path) {
            Ok(()) => Ok(()),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(()),
            Err(e) => Err(e.into()),
        }
    }

    fn save_client_credentials(&self, site: &str, creds: &ClientCredentials) -> Result<()> {
        let path = self
            .base_dir
            .join(format!("client_{}.json", sanitize(site)));
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
        let path = self
            .base_dir
            .join(format!("client_{}.json", sanitize(site)));
        match std::fs::read_to_string(&path) {
            Ok(json) => Ok(Some(serde_json::from_str(&json)?)),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(None),
            Err(e) => Err(e.into()),
        }
    }

    fn delete_client_credentials(&self, site: &str) -> Result<()> {
        let path = self
            .base_dir
            .join(format!("client_{}.json", sanitize(site)));
        match std::fs::remove_file(&path) {
            Ok(()) => Ok(()),
            Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(()),
            Err(e) => Err(e.into()),
        }
    }
}

// ---------------------------------------------------------------------------
// Keychain storage (via keyring crate) — native only
// ---------------------------------------------------------------------------

#[cfg(not(target_arch = "wasm32"))]
pub struct KeychainStorage;

#[cfg(not(target_arch = "wasm32"))]
const SERVICE_NAME: &str = "pup";

#[cfg(not(target_arch = "wasm32"))]
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

#[cfg(not(target_arch = "wasm32"))]
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
// In-memory storage (WASM) — no persistent storage available
// ---------------------------------------------------------------------------

#[cfg(target_arch = "wasm32")]
pub struct InMemoryStorage;

#[cfg(target_arch = "wasm32")]
impl Storage for InMemoryStorage {
    fn backend_type(&self) -> BackendType {
        BackendType::File
    }

    fn storage_location(&self) -> String {
        "in-memory (WASM)".to_string()
    }

    fn save_tokens(&self, _site: &str, _tokens: &TokenSet) -> Result<()> {
        anyhow::bail!("token storage not available in WASM — use DD_ACCESS_TOKEN env var")
    }

    fn load_tokens(&self, _site: &str) -> Result<Option<TokenSet>> {
        Ok(None)
    }

    fn delete_tokens(&self, _site: &str) -> Result<()> {
        Ok(())
    }

    fn save_client_credentials(&self, _site: &str, _creds: &ClientCredentials) -> Result<()> {
        anyhow::bail!("client credential storage not available in WASM")
    }

    fn load_client_credentials(&self, _site: &str) -> Result<Option<ClientCredentials>> {
        Ok(None)
    }

    fn delete_client_credentials(&self, _site: &str) -> Result<()> {
        Ok(())
    }
}

// ---------------------------------------------------------------------------
// LocalStorage backend (browser WASM) — persists tokens across page reloads
// ---------------------------------------------------------------------------

#[cfg(feature = "browser")]
pub struct LocalStorageBackend;

#[cfg(feature = "browser")]
impl LocalStorageBackend {
    fn storage() -> Result<web_sys::Storage> {
        let window = web_sys::window().ok_or_else(|| anyhow::anyhow!("no global window object"))?;
        window
            .local_storage()
            .map_err(|_| anyhow::anyhow!("localStorage not available"))?
            .ok_or_else(|| anyhow::anyhow!("localStorage returned None"))
    }

    fn get_item(key: &str) -> Result<Option<String>> {
        let storage = Self::storage()?;
        storage
            .get_item(key)
            .map_err(|_| anyhow::anyhow!("failed to read from localStorage"))
    }

    fn set_item(key: &str, value: &str) -> Result<()> {
        let storage = Self::storage()?;
        storage
            .set_item(key, value)
            .map_err(|_| anyhow::anyhow!("failed to write to localStorage"))
    }

    fn remove_item(key: &str) -> Result<()> {
        let storage = Self::storage()?;
        storage
            .remove_item(key)
            .map_err(|_| anyhow::anyhow!("failed to remove from localStorage"))
    }
}

#[cfg(feature = "browser")]
impl Storage for LocalStorageBackend {
    fn backend_type(&self) -> BackendType {
        BackendType::LocalStorage
    }

    fn storage_location(&self) -> String {
        "browser localStorage".to_string()
    }

    fn save_tokens(&self, site: &str, tokens: &TokenSet) -> Result<()> {
        let key = format!("pup_tokens_{}", sanitize(site));
        let json = serde_json::to_string(tokens)?;
        Self::set_item(&key, &json)
    }

    fn load_tokens(&self, site: &str) -> Result<Option<TokenSet>> {
        let key = format!("pup_tokens_{}", sanitize(site));
        match Self::get_item(&key)? {
            Some(json) => Ok(Some(serde_json::from_str(&json)?)),
            None => Ok(None),
        }
    }

    fn delete_tokens(&self, site: &str) -> Result<()> {
        let key = format!("pup_tokens_{}", sanitize(site));
        Self::remove_item(&key)
    }

    fn save_client_credentials(&self, site: &str, creds: &ClientCredentials) -> Result<()> {
        let key = format!("pup_client_{}", sanitize(site));
        let json = serde_json::to_string(creds)?;
        Self::set_item(&key, &json)
    }

    fn load_client_credentials(&self, site: &str) -> Result<Option<ClientCredentials>> {
        let key = format!("pup_client_{}", sanitize(site));
        match Self::get_item(&key)? {
            Some(json) => Ok(Some(serde_json::from_str(&json)?)),
            None => Ok(None),
        }
    }

    fn delete_client_credentials(&self, site: &str) -> Result<()> {
        let key = format!("pup_client_{}", sanitize(site));
        Self::remove_item(&key)
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

#[cfg(not(target_arch = "wasm32"))]
fn detect_backend() -> Box<dyn Storage> {
    // Check DD_TOKEN_STORAGE env var
    if let Ok(val) = std::env::var("DD_TOKEN_STORAGE") {
        match val.as_str() {
            "file" => return Box::new(FileStorage::new().expect("failed to create file storage")),
            "keychain" => return Box::new(KeychainStorage::new().expect("keychain not available")),
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

#[cfg(all(target_arch = "wasm32", not(feature = "browser")))]
fn detect_backend() -> Box<dyn Storage> {
    Box::new(InMemoryStorage)
}

#[cfg(feature = "browser")]
fn detect_backend() -> Box<dyn Storage> {
    Box::new(LocalStorageBackend)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

fn sanitize(site: &str) -> String {
    site.chars()
        .map(|c| if c.is_alphanumeric() { c } else { '_' })
        .collect()
}
