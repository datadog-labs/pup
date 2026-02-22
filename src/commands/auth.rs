use anyhow::{bail, Result};

use crate::auth::storage;
use crate::config::Config;

/// Helper to run a closure with the storage lock held (non-async to avoid holding lock across await).
fn with_storage<F, R>(f: F) -> Result<R>
where
    F: FnOnce(&mut dyn storage::Storage) -> Result<R>,
{
    let guard = storage::get_storage()?;
    let mut lock = guard.lock().unwrap();
    let store = lock.as_mut().unwrap();
    f(&mut **store)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn login(cfg: &Config) -> Result<()> {
    use crate::auth::{dcr, pkce, types};

    let site = &cfg.site;

    // 1. Start callback server
    let mut server = crate::auth::callback::CallbackServer::new().await?;
    let redirect_uri = server.redirect_uri();
    eprintln!("\n🔐 Starting OAuth2 login for site: {site}\n");
    eprintln!("📡 Callback server started on: {redirect_uri}");

    // 2. Load existing client credentials (lock released before any await)
    let existing_creds = with_storage(|store| store.load_client_credentials(site))?;

    let creds = match existing_creds {
        Some(creds) => {
            eprintln!("✓ Using existing client registration");
            creds
        }
        None => {
            eprintln!("📝 Registering new OAuth2 client...");
            let dcr_client = dcr::DcrClient::new(site);
            let scopes = types::default_scopes();
            let creds = dcr_client.register(&redirect_uri, &scopes).await?;
            with_storage(|store| store.save_client_credentials(site, &creds))?;
            eprintln!("✓ Registered client: {}", creds.client_id);
            creds
        }
    };

    // 3. Generate PKCE challenge + state
    let challenge = pkce::generate_pkce_challenge()?;
    let state = pkce::generate_state()?;

    // 4. Build authorization URL
    let dcr_client = dcr::DcrClient::new(site);
    let scopes = types::default_scopes();
    let auth_url = dcr_client.build_authorization_url(
        &creds.client_id,
        &redirect_uri,
        &state,
        &challenge,
        &scopes,
    );

    // 5. Open browser
    eprintln!("\n🌐 Opening browser for authentication...");
    eprintln!("If the browser doesn't open, visit: {auth_url}");
    let _ = open::that(&auth_url);

    // 6. Wait for callback
    eprintln!("\n⏳ Waiting for authorization...");
    let result = server
        .wait_for_callback(std::time::Duration::from_secs(300))
        .await?;

    if let Some(err) = &result.error {
        let desc = result.error_description.as_deref().unwrap_or("");
        bail!("OAuth error: {err}: {desc}");
    }

    if result.state != state {
        bail!("OAuth state mismatch (possible CSRF attack)");
    }

    // 7. Exchange code for tokens
    eprintln!("🔄 Exchanging authorization code for tokens...");
    let tokens = dcr_client
        .exchange_code(&result.code, &redirect_uri, &challenge.verifier, &creds)
        .await?;

    let location = with_storage(|store| {
        store.save_tokens(site, &tokens)?;
        Ok(store.storage_location())
    })?;

    let expires_at = chrono::DateTime::from_timestamp(tokens.issued_at + tokens.expires_in, 0)
        .map(|dt| dt.with_timezone(&chrono::Local).to_rfc3339())
        .unwrap_or_else(|| format!("in {} hours", tokens.expires_in / 3600));
    let display_location = if location.contains("keychain") || location.contains("Keychain") {
        "macOS Keychain (secure)".to_string()
    } else {
        location
    };
    eprintln!("\n✅ Login successful!");
    eprintln!("   Access token expires: {expires_at}");
    eprintln!("   Token stored in: {display_location}");

    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn login(_cfg: &Config) -> Result<()> {
    bail!(
        "OAuth login is not available in WASM builds.\n\
         Use DD_ACCESS_TOKEN env var for bearer token auth,\n\
         or DD_API_KEY + DD_APP_KEY for API key auth."
    )
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn logout(cfg: &Config) -> Result<()> {
    let site = &cfg.site;
    with_storage(|store| {
        store.delete_tokens(site)?;
        store.delete_client_credentials(site)?;
        Ok(())
    })?;
    eprintln!("Logged out from {site}. Tokens and client credentials removed.");
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn logout(_cfg: &Config) -> Result<()> {
    bail!(
        "OAuth logout is not available in WASM builds.\n\
         Token storage is not available — credentials are read from environment variables."
    )
}

pub fn status(cfg: &Config) -> Result<()> {
    let site = &cfg.site;

    // In WASM, just report env var status
    #[cfg(target_arch = "wasm32")]
    {
        if cfg.has_bearer_token() || cfg.has_api_keys() {
            println!("✅ Authenticated for site: {site}");
        } else {
            println!("❌ Not authenticated for site: {site}");
        }
        return Ok(());
    }

    #[cfg(not(target_arch = "wasm32"))]
    with_storage(|store| {
        match store.load_tokens(site)? {
            Some(tokens) => {
                let expires_at_ts = tokens.issued_at + tokens.expires_in;
                let now = chrono::Utc::now().timestamp();
                let remaining_secs = expires_at_ts - now;

                let (status, remaining_str) = if tokens.is_expired() {
                    ("expired".to_string(), "expired".to_string())
                } else {
                    let mins = remaining_secs / 60;
                    let secs = remaining_secs % 60;
                    ("valid".to_string(), format!("{mins}m{secs}s"))
                };

                if tokens.is_expired() {
                    eprintln!("⚠️  Token expired for site: {site}");
                } else {
                    eprintln!("✅ Authenticated for site: {site}");
                    eprintln!("   Token expires in: {remaining_str}");
                }

                let expires_at = chrono::DateTime::from_timestamp(expires_at_ts, 0)
                    .map(|dt| dt.with_timezone(&chrono::Local).to_rfc3339())
                    .unwrap_or_default();

                let json = serde_json::json!({
                    "authenticated": true,
                    "expires_at": expires_at,
                    "has_refresh": !tokens.refresh_token.is_empty(),
                    "site": site,
                    "status": status,
                    "token_type": tokens.token_type,
                });
                println!("{}", serde_json::to_string_pretty(&json).unwrap());
            }
            None => {
                eprintln!("❌ Not authenticated for site: {site}");
                let json = serde_json::json!({
                    "authenticated": false,
                    "site": site,
                    "status": "no token",
                });
                println!("{}", serde_json::to_string_pretty(&json).unwrap());
            }
        }
        Ok(())
    })
}

pub fn token(cfg: &Config) -> Result<()> {
    if let Some(token) = &cfg.access_token {
        println!("{token}");
        return Ok(());
    }

    #[cfg(target_arch = "wasm32")]
    bail!("no token available — set DD_ACCESS_TOKEN env var");

    #[cfg(not(target_arch = "wasm32"))]
    {
        let site = &cfg.site;
        with_storage(|store| match store.load_tokens(site)? {
            Some(tokens) => {
                if tokens.is_expired() {
                    bail!("token is expired — run 'pup auth login' to refresh");
                }
                println!("{}", tokens.access_token);
                Ok(())
            }
            None => bail!("no token available — run 'pup auth login' or set DD_ACCESS_TOKEN"),
        })
    }
}

pub async fn refresh(_cfg: &Config) -> Result<()> {
    anyhow::bail!("token refresh not yet implemented — use 'pup auth login' to re-authenticate")
}
