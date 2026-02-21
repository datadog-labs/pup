use anyhow::{bail, Result};
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::sync::oneshot;

use super::dcr::DCR_REDIRECT_PORTS;

pub struct CallbackResult {
    pub code: String,
    pub state: String,
    pub error: Option<String>,
    pub error_description: Option<String>,
}

pub struct CallbackServer {
    port: u16,
    shutdown_tx: Option<oneshot::Sender<()>>,
}

impl CallbackServer {
    /// Find an available port from DCR_REDIRECT_PORTS and prepare the server.
    pub async fn new() -> Result<Self> {
        for &port in DCR_REDIRECT_PORTS {
            if tokio::net::TcpListener::bind(("127.0.0.1", port))
                .await
                .is_ok()
            {
                return Ok(Self {
                    port,
                    shutdown_tx: None,
                });
            }
        }
        bail!(
            "could not bind to any DCR redirect port ({:?})",
            DCR_REDIRECT_PORTS
        );
    }

    pub fn port(&self) -> u16 {
        self.port
    }

    pub fn redirect_uri(&self) -> String {
        format!("http://127.0.0.1:{}/oauth/callback", self.port)
    }

    /// Start the server and wait for the OAuth callback.
    pub async fn wait_for_callback(
        &mut self,
        timeout: std::time::Duration,
    ) -> Result<CallbackResult> {
        let (shutdown_tx, shutdown_rx) = oneshot::channel();
        let (result_tx, result_rx) = oneshot::channel::<CallbackResult>();
        self.shutdown_tx = Some(shutdown_tx);

        let port = self.port;
        let listener = tokio::net::TcpListener::bind(("127.0.0.1", port)).await?;

        tokio::spawn(async move {
            let result_tx = std::sync::Mutex::new(Some(result_tx));
            tokio::select! {
                _ = accept_loop(listener, result_tx) => {}
                _ = shutdown_rx => {}
            }
        });

        match tokio::time::timeout(timeout, result_rx).await {
            Ok(Ok(result)) => Ok(result),
            Ok(Err(_)) => bail!("callback channel closed unexpectedly"),
            Err(_) => bail!("OAuth callback timed out after {timeout:?}"),
        }
    }

    pub fn stop(&mut self) {
        if let Some(tx) = self.shutdown_tx.take() {
            let _ = tx.send(());
        }
    }
}

impl Drop for CallbackServer {
    fn drop(&mut self) {
        self.stop();
    }
}

async fn accept_loop(
    listener: tokio::net::TcpListener,
    result_tx: std::sync::Mutex<Option<oneshot::Sender<CallbackResult>>>,
) {
    loop {
        let Ok((mut stream, _)) = listener.accept().await else {
            return;
        };

        let mut buf = vec![0u8; 4096];
        let Ok(n) = stream.read(&mut buf).await else {
            continue;
        };

        let request = String::from_utf8_lossy(&buf[..n]);
        let Some(path_line) = request.lines().next() else {
            continue;
        };
        let parts: Vec<&str> = path_line.split_whitespace().collect();
        if parts.len() < 2 || !parts[1].starts_with("/oauth/callback") {
            let response = "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n";
            let _ = stream.write_all(response.as_bytes()).await;
            continue;
        }

        let query_string = parts[1].split('?').nth(1).unwrap_or("");
        let params: std::collections::HashMap<String, String> =
            url::form_urlencoded::parse(query_string.as_bytes())
                .map(|(k, v)| (k.to_string(), v.to_string()))
                .collect();

        let code = params.get("code").cloned().unwrap_or_default();
        let state = params.get("state").cloned().unwrap_or_default();
        let error = params.get("error").cloned();
        let error_description = params.get("error_description").cloned();

        let (status, body) = if error.is_some() {
            ("400 Bad Request", error_page(&error, &error_description))
        } else {
            ("200 OK", success_page())
        };
        let response = format!(
            "HTTP/1.1 {status}\r\nContent-Type: text/html\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{body}",
            body.len()
        );
        let _ = stream.write_all(response.as_bytes()).await;

        let result = CallbackResult {
            code,
            state,
            error,
            error_description,
        };
        if let Some(tx) = result_tx.lock().unwrap().take() {
            let _ = tx.send(result);
        }
        return;
    }
}

fn success_page() -> String {
    r#"<!DOCTYPE html>
<html><head><title>Pup - Authentication Successful</title>
<style>body{font-family:system-ui;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#f5f5f5}
.card{background:white;padding:2rem;border-radius:8px;box-shadow:0 2px 4px rgba(0,0,0,0.1);text-align:center}
h1{color:#632ca6}p{color:#555}</style></head>
<body><div class="card"><h1>Authentication Successful</h1>
<p>You can close this window and return to pup.</p></div></body></html>"#.to_string()
}

fn error_page(error: &Option<String>, desc: &Option<String>) -> String {
    let err = error.as_deref().unwrap_or("unknown_error");
    let desc = desc.as_deref().unwrap_or("An unknown error occurred.");
    format!(
        r#"<!DOCTYPE html>
<html><head><title>Pup - Authentication Failed</title>
<style>body{{font-family:system-ui;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#f5f5f5}}
.card{{background:white;padding:2rem;border-radius:8px;box-shadow:0 2px 4px rgba(0,0,0,0.1);text-align:center}}
h1{{color:#c00}}p{{color:#555}}</style></head>
<body><div class="card"><h1>Authentication Failed</h1>
<p><strong>{err}</strong></p><p>{desc}</p>
<p>Please close this window and try again.</p></div></body></html>"#
    )
}
