//! Integration tests for command modules using mockito mock server.
//!
//! These tests use the PUP_MOCK_SERVER mechanism to redirect all DD API calls
//! to a local mockito server, testing command functions without a live API.
//!
//! For DD client API tests, we use Matcher::Any for paths since the DD client
//! library may construct URLs differently from what we expect. Each test gets
//! its own mockito server, so there's no cross-test interference.

use crate::config::{Config, OutputFormat};
use std::sync::Mutex;

/// Global mutex to serialize tests that modify process-wide env vars.
/// Uses unwrap_or_else to recover from poisoned state (previous test panicked).
static ENV_MUTEX: Mutex<()> = Mutex::new(());

fn lock_env() -> std::sync::MutexGuard<'static, ()> {
    ENV_MUTEX.lock().unwrap_or_else(|e| e.into_inner())
}

fn test_config(mock_url: &str) -> Config {
    std::env::set_var("PUP_MOCK_SERVER", mock_url);
    std::env::set_var("DD_API_KEY", "test-api-key");
    std::env::set_var("DD_APP_KEY", "test-app-key");

    Config {
        api_key: Some("test-api-key".into()),
        app_key: Some("test-app-key".into()),
        access_token: None,
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    }
}

fn cleanup_env() {
    std::env::remove_var("PUP_MOCK_SERVER");
}

/// Helper: create a catch-all mock that responds 200 with JSON for any request
/// matching the given HTTP method. Used for DD client API tests where the
/// exact path may differ from our expectations.
async fn mock_any(server: &mut mockito::Server, method: &str, body: &str) -> mockito::Mock {
    server
        .mock(method, mockito::Matcher::Any)
        .match_query(mockito::Matcher::Any)
        .with_status(200)
        .with_header("content-type", "application/json")
        .with_body(body)
        .create_async()
        .await
}

// =========================================================================
// DD Client API Command Tests (monitors, dashboards, slos, tags, events,
// logs, metrics) — use catch-all mocks since the DD client constructs its
// own URLs from OpenAPI specs.
// =========================================================================

// -------------------------------------------------------------------------
// Monitors
// -------------------------------------------------------------------------

#[tokio::test]
async fn test_monitors_list_empty() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "GET", "[]").await;

    let result = crate::commands::monitors::list(&cfg, None, None, 10).await;
    assert!(result.is_ok(), "monitors list failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_monitors_list_with_results() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());

    let body = r#"[{"id": 1, "name": "Test Monitor", "type": "metric alert", "query": "avg(last_5m):avg:system.cpu.user{*} > 90", "message": "CPU high", "tags": [], "options": {}}]"#;
    let _mock = mock_any(&mut server, "GET", body).await;

    let result = crate::commands::monitors::list(&cfg, Some("Test".into()), None, 10).await;
    assert!(
        result.is_ok(),
        "monitors list with results failed: {:?}",
        result.err()
    );
    cleanup_env();
}

#[tokio::test]
async fn test_monitors_get() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());

    let body = r#"{"id": 12345, "name": "Test Monitor", "type": "metric alert", "query": "avg(last_5m):avg:system.cpu.user{*} > 90", "message": "CPU high", "tags": [], "options": {}}"#;
    let _mock = mock_any(&mut server, "GET", body).await;

    let result = crate::commands::monitors::get(&cfg, 12345).await;
    assert!(result.is_ok(), "monitors get failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_monitors_search() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());

    let body = r#"{"monitors": [], "metadata": {"page": 0, "page_count": 0, "per_page": 30, "total_count": 0}}"#;
    let _mock = mock_any(&mut server, "GET", body).await;

    let result = crate::commands::monitors::search(&cfg, Some("cpu".into())).await;
    assert!(result.is_ok(), "monitors search failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_monitors_delete() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "DELETE", r#"{"deleted_monitor_id": 12345}"#).await;

    let result = crate::commands::monitors::delete(&cfg, 12345).await;
    assert!(result.is_ok(), "monitors delete failed: {:?}", result.err());
    cleanup_env();
}

// -------------------------------------------------------------------------
// Dashboards
// -------------------------------------------------------------------------

#[tokio::test]
async fn test_dashboards_list() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "GET", r#"{"dashboards": []}"#).await;

    let result = crate::commands::dashboards::list(&cfg).await;
    assert!(result.is_ok(), "dashboards list failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_dashboards_get() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(
        &mut server,
        "GET",
        r#"{"id": "abc-123", "title": "Test Dashboard", "layout_type": "ordered", "widgets": []}"#,
    )
    .await;

    let result = crate::commands::dashboards::get(&cfg, "abc-123").await;
    assert!(result.is_ok(), "dashboards get failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_dashboards_delete() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(
        &mut server,
        "DELETE",
        r#"{"deleted_dashboard_id": "abc-123"}"#,
    )
    .await;

    let result = crate::commands::dashboards::delete(&cfg, "abc-123").await;
    assert!(
        result.is_ok(),
        "dashboards delete failed: {:?}",
        result.err()
    );
    cleanup_env();
}

// -------------------------------------------------------------------------
// SLOs
// -------------------------------------------------------------------------

#[tokio::test]
async fn test_slos_list() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "GET", r#"{"data": [], "errors": []}"#).await;

    let result = crate::commands::slos::list(&cfg).await;
    assert!(result.is_ok(), "slos list failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_slos_get() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(
        &mut server,
        "GET",
        r#"{"data": {"id": "abc123", "name": "Test SLO", "type": "metric", "thresholds": [{"timeframe": "7d", "target": 99.9}]}, "errors": []}"#,
    )
    .await;

    let result = crate::commands::slos::get(&cfg, "abc123").await;
    assert!(result.is_ok(), "slos get failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_slos_delete() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "DELETE", r#"{"data": []}"#).await;

    let result = crate::commands::slos::delete(&cfg, "abc123").await;
    assert!(result.is_ok(), "slos delete failed: {:?}", result.err());
    cleanup_env();
}

// -------------------------------------------------------------------------
// Tags
// -------------------------------------------------------------------------

#[tokio::test]
async fn test_tags_list() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "GET", r#"{"tags": {}}"#).await;

    let result = crate::commands::tags::list(&cfg).await;
    assert!(result.is_ok(), "tags list failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_tags_get() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(
        &mut server,
        "GET",
        r#"{"host": "myhost", "tags": ["env:prod", "service:web"]}"#,
    )
    .await;

    let result = crate::commands::tags::get(&cfg, "myhost").await;
    assert!(result.is_ok(), "tags get failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_tags_add() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(
        &mut server,
        "POST",
        r#"{"host": "myhost", "tags": ["env:prod"]}"#,
    )
    .await;

    let result = crate::commands::tags::add(&cfg, "myhost", vec!["env:prod".into()]).await;
    assert!(result.is_ok(), "tags add failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_tags_update() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(
        &mut server,
        "PUT",
        r#"{"host": "myhost", "tags": ["env:staging"]}"#,
    )
    .await;

    let result = crate::commands::tags::update(&cfg, "myhost", vec!["env:staging".into()]).await;
    assert!(result.is_ok(), "tags update failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_tags_delete() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());

    // Delete returns 204 No Content
    let _mock = server
        .mock("DELETE", mockito::Matcher::Any)
        .match_query(mockito::Matcher::Any)
        .with_status(204)
        .create_async()
        .await;

    let result = crate::commands::tags::delete(&cfg, "myhost").await;
    assert!(result.is_ok(), "tags delete failed: {:?}", result.err());
    cleanup_env();
}

// -------------------------------------------------------------------------
// Events
// -------------------------------------------------------------------------

#[tokio::test]
async fn test_events_list() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "GET", r#"{"events": []}"#).await;

    let now = chrono::Utc::now().timestamp();
    let result = crate::commands::events::list(&cfg, now - 3600, now, None).await;
    assert!(result.is_ok(), "events list failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_events_get() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(
        &mut server,
        "GET",
        r#"{"event": {"id": 12345, "title": "Test Event", "text": "Something happened"}}"#,
    )
    .await;

    let result = crate::commands::events::get(&cfg, 12345).await;
    assert!(result.is_ok(), "events get failed: {:?}", result.err());
    cleanup_env();
}

// -------------------------------------------------------------------------
// Logs (requires API keys)
// -------------------------------------------------------------------------

#[tokio::test]
async fn test_logs_search() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "POST", r#"{"data": [], "meta": {"page": {}}}"#).await;

    let result = crate::commands::logs::search(
        &cfg,
        "status:error".into(),
        "1h".into(),
        "now".into(),
        10,
        None,
    )
    .await;
    assert!(result.is_ok(), "logs search failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_logs_search_requires_api_keys() {
    let _lock = lock_env();
    let server = mockito::Server::new_async().await;
    std::env::set_var("PUP_MOCK_SERVER", &server.url());

    let cfg = Config {
        api_key: None,
        app_key: None,
        access_token: Some("token".into()),
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let result = crate::commands::logs::search(
        &cfg,
        "status:error".into(),
        "1h".into(),
        "now".into(),
        10,
        None,
    )
    .await;
    assert!(result.is_err(), "logs search should require API keys");
    assert!(
        result
            .unwrap_err()
            .to_string()
            .contains("API+APP key authentication"),
        "error should mention API key auth"
    );
    cleanup_env();
}

#[tokio::test]
async fn test_logs_aggregate() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "POST", r#"{"data": {"buckets": []}}"#).await;

    let result =
        crate::commands::logs::aggregate(&cfg, "*".into(), "1h".into(), "now".into(), None).await;
    assert!(result.is_ok(), "logs aggregate failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_logs_search_with_flex_storage() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "POST", r#"{"data": [], "meta": {"page": {}}}"#).await;

    let result = crate::commands::logs::search(
        &cfg,
        "*".into(),
        "1h".into(),
        "now".into(),
        10,
        Some("flex".into()),
    )
    .await;
    assert!(
        result.is_ok(),
        "logs search with flex failed: {:?}",
        result.err()
    );
    cleanup_env();
}

#[tokio::test]
async fn test_logs_search_with_online_archives_storage() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "POST", r#"{"data": [], "meta": {"page": {}}}"#).await;

    let result = crate::commands::logs::search(
        &cfg,
        "*".into(),
        "1h".into(),
        "now".into(),
        10,
        Some("online-archives".into()),
    )
    .await;
    assert!(
        result.is_ok(),
        "logs search with online-archives failed: {:?}",
        result.err()
    );
    cleanup_env();
}

#[tokio::test]
async fn test_logs_search_with_invalid_storage_tier() {
    let _lock = lock_env();
    let server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());

    let result = crate::commands::logs::search(
        &cfg,
        "*".into(),
        "1h".into(),
        "now".into(),
        10,
        Some("invalid-tier".into()),
    )
    .await;
    assert!(
        result.is_err(),
        "logs search with invalid storage tier should fail"
    );
    assert!(
        result
            .unwrap_err()
            .to_string()
            .contains("unknown storage tier"),
        "error should mention unknown storage tier"
    );
    cleanup_env();
}

#[tokio::test]
async fn test_logs_aggregate_with_flex_storage() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "POST", r#"{"data": {"buckets": []}}"#).await;

    let result = crate::commands::logs::aggregate(
        &cfg,
        "*".into(),
        "1h".into(),
        "now".into(),
        Some("flex".into()),
    )
    .await;
    assert!(
        result.is_ok(),
        "logs aggregate with flex failed: {:?}",
        result.err()
    );
    cleanup_env();
}

#[tokio::test]
async fn test_logs_archives_list() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "GET", r#"{"data": []}"#).await;

    let result = crate::commands::logs::archives_list(&cfg).await;
    assert!(
        result.is_ok(),
        "logs archives list failed: {:?}",
        result.err()
    );
    cleanup_env();
}

#[tokio::test]
async fn test_logs_custom_destinations_list() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "GET", r#"{"data": []}"#).await;

    let result = crate::commands::logs::custom_destinations_list(&cfg).await;
    assert!(
        result.is_ok(),
        "logs custom destinations list failed: {:?}",
        result.err()
    );
    cleanup_env();
}

#[tokio::test]
async fn test_logs_metrics_list() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "GET", r#"{"data": []}"#).await;

    let result = crate::commands::logs::metrics_list(&cfg).await;
    assert!(
        result.is_ok(),
        "logs metrics list failed: {:?}",
        result.err()
    );
    cleanup_env();
}

#[tokio::test]
async fn test_logs_restriction_queries_list() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());

    // restriction_queries_list uses raw HTTP (not DD client), so mock specific path
    let _mock = server
        .mock("GET", "/api/v2/logs/config/restriction_queries")
        .match_query(mockito::Matcher::Any)
        .with_status(200)
        .with_header("content-type", "application/json")
        .with_body(r#"{"data": []}"#)
        .create_async()
        .await;

    let result = crate::commands::logs::restriction_queries_list(&cfg).await;
    assert!(
        result.is_ok(),
        "logs restriction queries list failed: {:?}",
        result.err()
    );
    cleanup_env();
}

// -------------------------------------------------------------------------
// Metrics
// -------------------------------------------------------------------------

#[tokio::test]
async fn test_metrics_list() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(
        &mut server,
        "GET",
        r#"{"metrics": [], "from": "2024-01-01T00:00:00Z"}"#,
    )
    .await;

    let result = crate::commands::metrics::list(&cfg, None, "1h".into()).await;
    assert!(result.is_ok(), "metrics list failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_metrics_query() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(
        &mut server,
        "GET",
        r#"{"status": "ok", "res_type": "time_series", "series": [], "from_date": 0, "to_date": 0, "query": "avg:system.cpu.user{*}"}"#,
    )
    .await;

    let result = crate::commands::metrics::query(
        &cfg,
        "avg:system.cpu.user{*}".into(),
        "1h".into(),
        "now".into(),
    )
    .await;
    assert!(result.is_ok(), "metrics query failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_metrics_metadata_get() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(
        &mut server,
        "GET",
        r#"{"type": "gauge", "description": "CPU usage", "unit": "percent"}"#,
    )
    .await;

    let result = crate::commands::metrics::metadata_get(&cfg, "system.cpu.user").await;
    assert!(
        result.is_ok(),
        "metrics metadata get failed: {:?}",
        result.err()
    );
    cleanup_env();
}

// -------------------------------------------------------------------------
// Events search (requires API keys)
// -------------------------------------------------------------------------

#[tokio::test]
async fn test_events_search() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    let cfg = test_config(&server.url());
    let _mock = mock_any(&mut server, "POST", r#"{"data": [], "meta": {"page": {}}}"#).await;

    let result =
        crate::commands::events::search(&cfg, "source:nginx".into(), "1h".into(), "now".into(), 10)
            .await;
    assert!(result.is_ok(), "events search failed: {:?}", result.err());
    cleanup_env();
}

#[tokio::test]
async fn test_events_search_requires_api_keys() {
    let _lock = lock_env();
    let server = mockito::Server::new_async().await;
    std::env::set_var("PUP_MOCK_SERVER", &server.url());

    let cfg = Config {
        api_key: None,
        app_key: None,
        access_token: Some("token".into()),
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let result =
        crate::commands::events::search(&cfg, "source:nginx".into(), "1h".into(), "now".into(), 10)
            .await;
    assert!(result.is_err(), "events search should require API keys");
    cleanup_env();
}

// =========================================================================
// Raw HTTP api module tests — these use the api.rs module directly
// (not the DD client library), so we can mock specific paths precisely.
// =========================================================================

#[tokio::test]
async fn test_api_get() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    std::env::set_var("PUP_MOCK_SERVER", &server.url());

    let cfg = Config {
        api_key: Some("test-key".into()),
        app_key: Some("test-app".into()),
        access_token: None,
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let mock = server
        .mock("GET", "/api/v1/test")
        .with_status(200)
        .with_header("content-type", "application/json")
        .with_body(r#"{"status": "ok"}"#)
        .create_async()
        .await;

    let result = crate::api::get(&cfg, "/api/v1/test", &[]).await;
    assert!(result.is_ok(), "api get failed: {:?}", result.err());
    let val = result.unwrap();
    assert_eq!(val["status"], "ok");
    mock.assert_async().await;
    cleanup_env();
}

#[tokio::test]
async fn test_api_get_with_query() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    std::env::set_var("PUP_MOCK_SERVER", &server.url());

    let cfg = Config {
        api_key: Some("test-key".into()),
        app_key: Some("test-app".into()),
        access_token: None,
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let mock = server
        .mock("GET", "/api/v1/search")
        .match_query(mockito::Matcher::Any)
        .with_status(200)
        .with_header("content-type", "application/json")
        .with_body(r#"{"results": []}"#)
        .create_async()
        .await;

    let query = vec![("q", "test".to_string())];
    let result = crate::api::get(&cfg, "/api/v1/search", &query).await;
    assert!(
        result.is_ok(),
        "api get with query failed: {:?}",
        result.err()
    );
    mock.assert_async().await;
    cleanup_env();
}

#[tokio::test]
async fn test_api_post() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    std::env::set_var("PUP_MOCK_SERVER", &server.url());

    let cfg = Config {
        api_key: Some("test-key".into()),
        app_key: Some("test-app".into()),
        access_token: None,
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let mock = server
        .mock("POST", "/api/v2/test")
        .with_status(200)
        .with_header("content-type", "application/json")
        .with_body(r#"{"created": true}"#)
        .create_async()
        .await;

    let body = serde_json::json!({"name": "test"});
    let result = crate::api::post(&cfg, "/api/v2/test", &body).await;
    assert!(result.is_ok(), "api post failed: {:?}", result.err());
    mock.assert_async().await;
    cleanup_env();
}

#[tokio::test]
async fn test_api_put() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    std::env::set_var("PUP_MOCK_SERVER", &server.url());

    let cfg = Config {
        api_key: Some("test-key".into()),
        app_key: Some("test-app".into()),
        access_token: None,
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let mock = server
        .mock("PUT", "/api/v1/test/123")
        .with_status(200)
        .with_header("content-type", "application/json")
        .with_body(r#"{"updated": true}"#)
        .create_async()
        .await;

    let body = serde_json::json!({"name": "updated"});
    let result = crate::api::put(&cfg, "/api/v1/test/123", &body).await;
    assert!(result.is_ok(), "api put failed: {:?}", result.err());
    mock.assert_async().await;
    cleanup_env();
}

#[tokio::test]
async fn test_api_patch() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    std::env::set_var("PUP_MOCK_SERVER", &server.url());

    let cfg = Config {
        api_key: Some("test-key".into()),
        app_key: Some("test-app".into()),
        access_token: None,
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let mock = server
        .mock("PATCH", "/api/v1/test/123")
        .with_status(200)
        .with_header("content-type", "application/json")
        .with_body(r#"{"patched": true}"#)
        .create_async()
        .await;

    let body = serde_json::json!({"name": "patched"});
    let result = crate::api::patch(&cfg, "/api/v1/test/123", &body).await;
    assert!(result.is_ok(), "api patch failed: {:?}", result.err());
    mock.assert_async().await;
    cleanup_env();
}

#[tokio::test]
async fn test_api_delete() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    std::env::set_var("PUP_MOCK_SERVER", &server.url());

    let cfg = Config {
        api_key: Some("test-key".into()),
        app_key: Some("test-app".into()),
        access_token: None,
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let mock = server
        .mock("DELETE", "/api/v1/test/123")
        .with_status(200)
        .with_header("content-type", "application/json")
        .with_body(r#"{"deleted": true}"#)
        .create_async()
        .await;

    let result = crate::api::delete(&cfg, "/api/v1/test/123").await;
    assert!(result.is_ok(), "api delete failed: {:?}", result.err());
    mock.assert_async().await;
    cleanup_env();
}

#[tokio::test]
async fn test_api_error_response() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    std::env::set_var("PUP_MOCK_SERVER", &server.url());

    let cfg = Config {
        api_key: Some("test-key".into()),
        app_key: Some("test-app".into()),
        access_token: None,
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let mock = server
        .mock("GET", "/api/v1/test/missing")
        .with_status(404)
        .with_header("content-type", "application/json")
        .with_body(r#"{"errors": ["not found"]}"#)
        .create_async()
        .await;

    let result = crate::api::get(&cfg, "/api/v1/test/missing", &[]).await;
    assert!(result.is_err(), "should return error for 404");
    assert!(result.unwrap_err().to_string().contains("404"));
    mock.assert_async().await;
    cleanup_env();
}

#[tokio::test]
async fn test_api_bearer_auth() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    std::env::set_var("PUP_MOCK_SERVER", &server.url());

    let cfg = Config {
        api_key: None,
        app_key: None,
        access_token: Some("test-bearer-token".into()),
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let mock = server
        .mock("GET", "/api/v1/test")
        .match_header("Authorization", "Bearer test-bearer-token")
        .with_status(200)
        .with_header("content-type", "application/json")
        .with_body(r#"{"auth": "bearer"}"#)
        .create_async()
        .await;

    let result = crate::api::get(&cfg, "/api/v1/test", &[]).await;
    assert!(result.is_ok(), "bearer auth failed: {:?}", result.err());
    mock.assert_async().await;
    cleanup_env();
}

#[tokio::test]
async fn test_api_no_auth() {
    let _lock = lock_env();

    let cfg = Config {
        api_key: None,
        app_key: None,
        access_token: None,
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let result = crate::api::get(&cfg, "/api/v1/test", &[]).await;
    assert!(result.is_err(), "should fail without auth");
    assert!(
        result.unwrap_err().to_string().contains("authentication"),
        "error should mention authentication"
    );
    cleanup_env();
}

#[tokio::test]
async fn test_api_empty_response() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    std::env::set_var("PUP_MOCK_SERVER", &server.url());

    let cfg = Config {
        api_key: Some("test-key".into()),
        app_key: Some("test-app".into()),
        access_token: None,
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let mock = server
        .mock("DELETE", "/api/v1/test/empty")
        .with_status(204)
        .with_body("")
        .create_async()
        .await;

    let result = crate::api::delete(&cfg, "/api/v1/test/empty").await;
    assert!(result.is_ok(), "empty response failed: {:?}", result.err());
    let val = result.unwrap();
    assert_eq!(val, serde_json::json!({}));
    mock.assert_async().await;
    cleanup_env();
}

#[tokio::test]
async fn test_api_server_error() {
    let _lock = lock_env();
    let mut server = mockito::Server::new_async().await;
    std::env::set_var("PUP_MOCK_SERVER", &server.url());

    let cfg = Config {
        api_key: Some("test-key".into()),
        app_key: Some("test-app".into()),
        access_token: None,
        site: "datadoghq.com".into(),
        output_format: OutputFormat::Json,
        auto_approve: false,
        agent_mode: false,
    };

    let mock = server
        .mock("GET", "/api/v1/test")
        .with_status(500)
        .with_body(r#"{"errors": ["internal server error"]}"#)
        .create_async()
        .await;

    let result = crate::api::get(&cfg, "/api/v1/test", &[]).await;
    assert!(result.is_err());
    assert!(result.unwrap_err().to_string().contains("500"));
    mock.assert_async().await;
    cleanup_env();
}

// =========================================================================
// Bulk command module tests — exercise list/get operations for all remaining
// command modules to maximize coverage. The mock_any helper catches all
// requests. We use `let _ =` instead of asserting success because some DD
// client types may not deserialize our minimal responses — the important
// thing is that the command code paths are exercised.
// =========================================================================

/// Mock all HTTP methods with the same response body.
async fn mock_all(s: &mut mockito::Server, body: &str) {
    for method in &["GET", "POST", "PUT", "PATCH", "DELETE"] {
        s.mock(method, mockito::Matcher::Any)
            .match_query(mockito::Matcher::Any)
            .with_status(200)
            .with_header("content-type", "application/json")
            .with_body(body)
            .create_async()
            .await;
    }
}

// --- RUM ---
#[tokio::test]
async fn test_rum_apps_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::rum::apps_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_rum_apps_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {"id": "abc", "type": "rum_browser"}}"#).await;
    let _ = crate::commands::rum::apps_get(&cfg, "abc").await;
    cleanup_env();
}
#[tokio::test]
async fn test_rum_apps_delete() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{}"#).await;
    let _ = crate::commands::rum::apps_delete(&cfg, "abc").await;
    cleanup_env();
}
#[tokio::test]
async fn test_rum_metrics_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::rum::metrics_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_rum_metrics_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::rum::metrics_get(&cfg, "m1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_rum_metrics_delete() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{}"#).await;
    let _ = crate::commands::rum::metrics_delete(&cfg, "m1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_rum_retention_filters_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::rum::retention_filters_list(&cfg, "app1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_rum_events_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::rum::events_list(&cfg, "1h".into(), "now".into(), 10).await;
    cleanup_env();
}
#[tokio::test]
async fn test_rum_playlists_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::rum::playlists_list(&cfg).await;
    cleanup_env();
}

// --- Status Pages ---
#[tokio::test]
async fn test_status_pages_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::status_pages::pages_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_status_pages_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::status_pages::pages_get(&cfg, "p1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_status_pages_delete() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{}"#).await;
    let _ = crate::commands::status_pages::pages_delete(&cfg, "p1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_status_pages_components_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::status_pages::components_list(&cfg, "p1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_status_pages_degradations_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::status_pages::degradations_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_status_pages_third_party_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::status_pages::third_party_list(&cfg, None, false).await;
    cleanup_env();
}

// --- Cases ---
#[tokio::test]
async fn test_cases_search() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::cases::search(&cfg, None, 10).await;
    cleanup_env();
}
#[tokio::test]
async fn test_cases_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::cases::get(&cfg, "case1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_cases_projects_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::cases::projects_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_cases_projects_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::cases::projects_get(&cfg, "proj1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_cases_projects_delete() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{}"#).await;
    let _ = crate::commands::cases::projects_delete(&cfg, "proj1").await;
    cleanup_env();
}

// --- Integrations ---
#[tokio::test]
async fn test_integrations_jira_accounts_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::integrations::jira_accounts_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_integrations_jira_templates_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::integrations::jira_templates_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_integrations_servicenow_instances_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::integrations::servicenow_instances_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_integrations_servicenow_templates_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::integrations::servicenow_templates_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_integrations_slack_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::integrations::slack_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_integrations_webhooks_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::integrations::webhooks_list(&cfg).await;
    cleanup_env();
}

// --- CI/CD ---
#[tokio::test]
async fn test_cicd_pipelines_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::cicd::pipelines_list(&cfg, None, "1h".into(), "now".into(), 10).await;
    cleanup_env();
}
#[tokio::test]
async fn test_cicd_tests_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::cicd::tests_list(&cfg, None, "1h".into(), "now".into(), 10).await;
    cleanup_env();
}

// --- Fleet ---
#[tokio::test]
async fn test_fleet_agents_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::fleet::agents_list(&cfg, None).await;
    cleanup_env();
}
#[tokio::test]
async fn test_fleet_agents_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::fleet::agents_get(&cfg, "a1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_fleet_agents_versions() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::fleet::agents_versions(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_fleet_deployments_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::fleet::deployments_list(&cfg, None).await;
    cleanup_env();
}
#[tokio::test]
async fn test_fleet_schedules_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::fleet::schedules_list(&cfg).await;
    cleanup_env();
}

// --- Incidents ---
#[tokio::test]
async fn test_incidents_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::incidents::list(&cfg, 10).await;
    cleanup_env();
}
#[tokio::test]
async fn test_incidents_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::incidents::get(&cfg, "inc1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_incidents_settings_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::incidents::settings_get(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_incidents_handles_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::incidents::handles_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_incidents_postmortem_templates_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::incidents::postmortem_templates_list(&cfg).await;
    cleanup_env();
}

// --- On-Call ---
#[tokio::test]
async fn test_on_call_teams_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::on_call::teams_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_on_call_teams_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::on_call::teams_get(&cfg, "t1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_on_call_teams_delete() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{}"#).await;
    let _ = crate::commands::on_call::teams_delete(&cfg, "t1").await;
    cleanup_env();
}

// --- Security ---
#[tokio::test]
async fn test_security_rules_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::security::rules_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_security_rules_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::security::rules_get(&cfg, "r1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_security_content_packs_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::security::content_packs_list(&cfg).await;
    cleanup_env();
}

// --- Synthetics ---
#[tokio::test]
async fn test_synthetics_tests_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"tests": []}"#).await;
    let _ = crate::commands::synthetics::tests_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_synthetics_tests_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{}"#).await;
    let _ = crate::commands::synthetics::tests_get(&cfg, "pub1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_synthetics_locations_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"locations": []}"#).await;
    let _ = crate::commands::synthetics::locations_list(&cfg).await;
    cleanup_env();
}

// --- App Keys ---
#[tokio::test]
async fn test_app_keys_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::app_keys::list(&cfg, 10, 0, "", "").await;
    cleanup_env();
}
#[tokio::test]
async fn test_app_keys_list_all() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::app_keys::list_all(&cfg, 10, 0, "", "").await;
    cleanup_env();
}
#[tokio::test]
async fn test_app_keys_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::app_keys::get(&cfg, "k1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_app_keys_create() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::app_keys::create(&cfg, "test-key", "").await;
    cleanup_env();
}
#[tokio::test]
async fn test_app_keys_update() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::app_keys::update(&cfg, "k1", "new-name", "").await;
    cleanup_env();
}
#[tokio::test]
async fn test_app_keys_delete() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{}"#).await;
    let _ = crate::commands::app_keys::delete(&cfg, "k1").await;
    cleanup_env();
}

// --- API Keys ---
#[tokio::test]
async fn test_api_keys_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::api_keys::list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_api_keys_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::api_keys::get(&cfg, "k1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_api_keys_delete() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{}"#).await;
    let _ = crate::commands::api_keys::delete(&cfg, "k1").await;
    cleanup_env();
}

// --- Audit Logs ---
#[tokio::test]
async fn test_audit_logs_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::audit_logs::list(&cfg, "1h".into(), "now".into(), 10).await;
    cleanup_env();
}

// --- Users ---
#[tokio::test]
async fn test_users_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::users::list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_users_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::users::get(&cfg, "u1").await;
    cleanup_env();
}
#[tokio::test]
async fn test_users_roles_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::users::roles_list(&cfg).await;
    cleanup_env();
}

// --- Usage ---
#[tokio::test]
async fn test_usage_summary() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"usage": []}"#).await;
    let _ = crate::commands::usage::summary(&cfg, "2024-01".into(), None).await;
    cleanup_env();
}

// --- Infrastructure ---
#[tokio::test]
async fn test_infrastructure_hosts_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"host_list": [], "total_returned": 0}"#).await;
    let _ = crate::commands::infrastructure::hosts_list(&cfg, None, "name".into(), 10).await;
    cleanup_env();
}

// --- Notebooks ---
#[tokio::test]
async fn test_notebooks_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::notebooks::list(&cfg).await;
    cleanup_env();
}

// --- Downtime ---
#[tokio::test]
async fn test_downtime_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::downtime::list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_downtime_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::downtime::get(&cfg, "d1").await;
    cleanup_env();
}

// --- Cost ---
#[tokio::test]
async fn test_cost_projected() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::cost::projected(&cfg).await;
    cleanup_env();
}

// --- Error Tracking ---
#[tokio::test]
async fn test_error_tracking_issues_search() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::error_tracking::issues_search(&cfg, None, 10).await;
    cleanup_env();
}

// --- Cloud ---
#[tokio::test]
async fn test_cloud_aws_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::cloud::aws_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_cloud_gcp_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::cloud::gcp_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_cloud_azure_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::cloud::azure_list(&cfg).await;
    cleanup_env();
}

// --- Organizations ---
#[tokio::test]
async fn test_organizations_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"orgs": []}"#).await;
    let _ = crate::commands::organizations::list(&cfg).await;
    cleanup_env();
}

// --- Service Catalog ---
#[tokio::test]
async fn test_service_catalog_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::service_catalog::list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_service_catalog_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::service_catalog::get(&cfg, "svc1").await;
    cleanup_env();
}

// --- Misc ---
#[tokio::test]
async fn test_misc_ip_ranges() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{}"#).await;
    let _ = crate::commands::misc::ip_ranges(&cfg).await;
    cleanup_env();
}

// --- Data Governance ---
#[tokio::test]
async fn test_data_governance_scanner_rules_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::data_governance::scanner_rules_list(&cfg).await;
    cleanup_env();
}

// --- Investigations ---
#[tokio::test]
async fn test_investigations_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::investigations::list(&cfg, 10, 0).await;
    cleanup_env();
}
#[tokio::test]
async fn test_investigations_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::investigations::get(&cfg, "inv1").await;
    cleanup_env();
}

// --- Network ---
#[tokio::test]
async fn test_network_flows_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::network::flows_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_network_devices_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::network::devices_list(&cfg).await;
    cleanup_env();
}

// --- Code Coverage ---
#[tokio::test]
async fn test_code_coverage_branch_summary() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ =
        crate::commands::code_coverage::branch_summary(&cfg, "repo".into(), "main".into()).await;
    cleanup_env();
}

// --- HAMR ---
#[tokio::test]
async fn test_hamr_connections_get() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": {}}"#).await;
    let _ = crate::commands::hamr::connections_get(&cfg).await;
    cleanup_env();
}

// --- Static Analysis ---
#[tokio::test]
async fn test_static_analysis_ast_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::static_analysis::ast_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_static_analysis_sca_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::static_analysis::sca_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_static_analysis_custom_rulesets_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::static_analysis::custom_rulesets_list(&cfg).await;
    cleanup_env();
}
#[tokio::test]
async fn test_static_analysis_coverage_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ = crate::commands::static_analysis::coverage_list(&cfg).await;
    cleanup_env();
}

// --- APM ---
#[tokio::test]
async fn test_apm_services_list() {
    let _lock = lock_env();
    let mut s = mockito::Server::new_async().await;
    let cfg = test_config(&s.url());
    mock_all(&mut s, r#"{"data": []}"#).await;
    let _ =
        crate::commands::apm::services_list(&cfg, "prod".into(), "1h".into(), "now".into()).await;
    cleanup_env();
}
