use anyhow::{bail, Result};

use crate::auth::{dcr, pkce, storage, types};
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

pub async fn login(cfg: &Config) -> Result<()> {
    let site = &cfg.site;

    // 1. Start callback server
    let mut server = crate::auth::callback::CallbackServer::new().await?;
    let redirect_uri = server.redirect_uri();
    eprintln!("Starting OAuth2 login for site: {site}");

    // 2. Load existing client credentials (lock released before any await)
    let existing_creds = with_storage(|store| store.load_client_credentials(site))?;

    let creds = match existing_creds {
        Some(creds) => {
            eprintln!("Using existing client registration: {}", creds.client_id);
            creds
        }
        None => {
            eprintln!("Registering new OAuth2 client...");
            let dcr_client = dcr::DcrClient::new(site);
            let scopes = types::default_scopes();
            let creds = dcr_client.register(&redirect_uri, &scopes).await?;
            with_storage(|store| store.save_client_credentials(site, &creds))?;
            eprintln!("Registered client: {}", creds.client_id);
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
    eprintln!("Opening browser for authorization...");
    eprintln!("If the browser doesn't open, visit:\n  {auth_url}");
    let _ = open::that(&auth_url);

    // 6. Wait for callback
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
    eprintln!("Exchanging authorization code for tokens...");
    let tokens = dcr_client
        .exchange_code(&result.code, &redirect_uri, &challenge.verifier, &creds)
        .await?;

    let location = with_storage(|store| {
        store.save_tokens(site, &tokens)?;
        Ok(store.storage_location())
    })?;

    eprintln!("Login successful! Tokens stored in {location}.");
    eprintln!("Token expires in {} hours.", tokens.expires_in / 3600);

    Ok(())
}

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

pub fn status(cfg: &Config) -> Result<()> {
    let site = &cfg.site;
    with_storage(|store| {
        println!("Site: {site}");
        println!(
            "Storage: {} ({})",
            store.backend_type(),
            store.storage_location()
        );

        match store.load_tokens(site)? {
            Some(tokens) => {
                println!("Authenticated: yes");
                println!("Token type: {}", tokens.token_type);
                if tokens.is_expired() {
                    println!("Token status: EXPIRED");
                } else {
                    let remaining =
                        (tokens.issued_at + tokens.expires_in) - chrono::Utc::now().timestamp();
                    println!("Token status: valid ({} minutes remaining)", remaining / 60);
                }
                if !tokens.client_id.is_empty() {
                    println!("Client ID: {}", tokens.client_id);
                }
            }
            None => {
                println!("Authenticated: no");
                if cfg.has_api_keys() {
                    println!("API keys: configured");
                }
                if cfg.has_bearer_token() {
                    println!("Bearer token: configured (DD_ACCESS_TOKEN)");
                }
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

pub async fn refresh(_cfg: &Config) -> Result<()> {
    anyhow::bail!("token refresh not yet implemented — use 'pup auth login' to re-authenticate")
}
