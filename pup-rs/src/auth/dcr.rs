use anyhow::{bail, Context, Result};
use chrono::Utc;
use serde::{Deserialize, Serialize};

use super::types::{ClientCredentials, TokenSet};

pub const DCR_CLIENT_NAME: &str = "datadog-api-claude-plugin";
pub const DCR_REDIRECT_PORTS: &[u16] = &[8000, 8080, 8888, 9000];

pub fn get_redirect_uris() -> Vec<String> {
    DCR_REDIRECT_PORTS
        .iter()
        .map(|port| format!("http://127.0.0.1:{port}/oauth/callback"))
        .collect()
}

/// DCR + token exchange client.
pub struct DcrClient {
    site: String,
    http: reqwest::Client,
}

#[derive(Serialize)]
struct RegistrationRequest {
    client_name: String,
    redirect_uris: Vec<String>,
    grant_types: Vec<String>,
}

#[derive(Deserialize)]
struct RegistrationResponse {
    client_id: String,
    client_name: String,
    redirect_uris: Vec<String>,
}

#[derive(Deserialize)]
struct TokenResponse {
    access_token: String,
    token_type: String,
    expires_in: i64,
    #[serde(default)]
    refresh_token: String,
    #[serde(default)]
    scope: String,
}

impl DcrClient {
    pub fn new(site: &str) -> Self {
        Self {
            site: site.to_string(),
            http: reqwest::Client::builder()
                .timeout(std::time::Duration::from_secs(30))
                .build()
                .expect("failed to build HTTP client"),
        }
    }

    /// Dynamic Client Registration (RFC 7591).
    pub async fn register(
        &self,
        redirect_uri: &str,
        _scopes: &[&str],
    ) -> Result<ClientCredentials> {
        let url = format!("https://api.{}/api/v2/oauth2/register", self.site);

        let body = RegistrationRequest {
            client_name: DCR_CLIENT_NAME.to_string(),
            redirect_uris: vec![redirect_uri.to_string()],
            grant_types: vec!["authorization_code".to_string(), "refresh_token".to_string()],
        };

        let resp = self
            .http
            .post(&url)
            .json(&body)
            .send()
            .await
            .context("DCR registration request failed")?;

        if resp.status() != reqwest::StatusCode::CREATED {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            bail!("DCR registration failed (HTTP {status}): {body}");
        }

        let reg: RegistrationResponse = resp.json().await.context("failed to parse DCR response")?;

        Ok(ClientCredentials {
            client_id: reg.client_id,
            client_name: reg.client_name,
            redirect_uris: reg.redirect_uris,
            registered_at: Utc::now().timestamp(),
            site: self.site.clone(),
        })
    }

    /// Exchange authorization code for tokens.
    pub async fn exchange_code(
        &self,
        code: &str,
        redirect_uri: &str,
        code_verifier: &str,
        creds: &ClientCredentials,
    ) -> Result<TokenSet> {
        let params = [
            ("grant_type", "authorization_code"),
            ("client_id", &creds.client_id),
            ("code", code),
            ("redirect_uri", redirect_uri),
            ("code_verifier", code_verifier),
        ];
        self.request_tokens(&params, &creds.client_id).await
    }

    /// Refresh an access token.
    pub async fn refresh_token(
        &self,
        refresh_token: &str,
        creds: &ClientCredentials,
    ) -> Result<TokenSet> {
        let params = [
            ("grant_type", "refresh_token"),
            ("client_id", &creds.client_id),
            ("refresh_token", refresh_token),
            ("redirect_uri", ""),  // not needed for refresh
            ("code_verifier", ""), // not needed for refresh
        ];
        self.request_tokens(&params, &creds.client_id).await
    }

    async fn request_tokens(&self, params: &[(&str, &str)], client_id: &str) -> Result<TokenSet> {
        let url = format!("https://api.{}/oauth2/v1/token", self.site);

        // Filter out empty params
        let form_params: Vec<(&str, &str)> = params
            .iter()
            .filter(|(_, v)| !v.is_empty())
            .copied()
            .collect();

        let resp = self
            .http
            .post(&url)
            .form(&form_params)
            .send()
            .await
            .context("token request failed")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            bail!("token exchange failed (HTTP {status}): {body}");
        }

        let token_resp: TokenResponse = resp.json().await.context("failed to parse token response")?;

        Ok(TokenSet {
            access_token: token_resp.access_token,
            refresh_token: token_resp.refresh_token,
            token_type: token_resp.token_type,
            expires_in: token_resp.expires_in,
            issued_at: Utc::now().timestamp(),
            scope: token_resp.scope,
            client_id: client_id.to_string(),
        })
    }

    /// Build the authorization URL for the browser.
    pub fn build_authorization_url(
        &self,
        client_id: &str,
        redirect_uri: &str,
        state: &str,
        challenge: &super::pkce::PkceChallenge,
        scopes: &[&str],
    ) -> String {
        let scope = scopes.join(" ");
        let params = url::form_urlencoded::Serializer::new(String::new())
            .append_pair("response_type", "code")
            .append_pair("client_id", client_id)
            .append_pair("redirect_uri", redirect_uri)
            .append_pair("state", state)
            .append_pair("scope", &scope)
            .append_pair("code_challenge", &challenge.challenge)
            .append_pair("code_challenge_method", &challenge.method)
            .finish();
        format!("https://app.{}/oauth2/v1/authorize?{params}", self.site)
    }
}
