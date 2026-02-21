use anyhow::Result;
use serde::Serialize;

use crate::config::OutputFormat;

/// Agent mode metadata envelope.
#[derive(Serialize)]
pub struct Metadata {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub count: Option<usize>,
    #[serde(skip_serializing_if = "std::ops::Not::not")]
    pub truncated: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub command: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub next_action: Option<String>,
}

/// Agent mode wrapper: { status, data, metadata }
#[derive(Serialize)]
struct AgentEnvelope<'a, T: Serialize> {
    status: &'static str,
    data: &'a T,
    #[serde(skip_serializing_if = "Option::is_none")]
    metadata: Option<&'a Metadata>,
}

/// Non-agent success wrapper: { status, data }
#[derive(Serialize)]
struct SuccessEnvelope<'a, T: Serialize> {
    status: &'static str,
    data: &'a T,
}

/// Recursively sort all JSON object keys alphabetically.
fn sort_json_value(v: serde_json::Value) -> serde_json::Value {
    match v {
        serde_json::Value::Object(map) => {
            let mut sorted: std::collections::BTreeMap<String, serde_json::Value> =
                std::collections::BTreeMap::new();
            for (k, val) in map {
                sorted.insert(k, sort_json_value(val));
            }
            serde_json::Value::Object(sorted.into_iter().collect())
        }
        serde_json::Value::Array(arr) => {
            serde_json::Value::Array(arr.into_iter().map(sort_json_value).collect())
        }
        other => other,
    }
}

/// Go's encoding/json escapes <, >, and & for HTML safety.
/// Apply the same escaping to match Go output exactly.
fn go_html_escape(json: &str) -> String {
    json.replace('&', "\\u0026")
        .replace('<', "\\u003c")
        .replace('>', "\\u003e")
}

/// Format and print data to stdout.
pub fn format_and_print<T: Serialize>(
    data: &T,
    format: &OutputFormat,
    agent_mode: bool,
    meta: Option<&Metadata>,
) -> Result<()> {
    if agent_mode {
        // Sort inner data keys but preserve envelope field order (status first)
        let sorted_data = sort_json_value(serde_json::to_value(data)?);
        let envelope = AgentEnvelope {
            status: "success",
            data: &sorted_data,
            metadata: meta,
        };
        let json = go_html_escape(&serde_json::to_string_pretty(&envelope)?);
        println!("{json}");
        return Ok(());
    }

    match format {
        OutputFormat::Json => print_json(data),
        OutputFormat::Yaml => print_yaml(data),
        OutputFormat::Table => print_table(data),
    }
}

/// Convenience: format and print using config settings (respects -o flag and agent mode).
pub fn output<T: Serialize>(cfg: &crate::config::Config, data: &T) -> Result<()> {
    format_and_print(data, &cfg.output_format, cfg.agent_mode, None)
}

pub fn print_json<T: Serialize>(data: &T) -> Result<()> {
    let sorted_data = sort_json_value(serde_json::to_value(data)?);
    let envelope = SuccessEnvelope {
        status: "success",
        data: &sorted_data,
    };
    let json = go_html_escape(&serde_json::to_string_pretty(&envelope)?);
    println!("{json}");
    Ok(())
}

fn print_yaml<T: Serialize>(data: &T) -> Result<()> {
    let sorted_data = sort_json_value(serde_json::to_value(data)?);
    let envelope = SuccessEnvelope {
        status: "success",
        data: &sorted_data,
    };
    let yaml = serde_yaml::to_string(&envelope)?;
    print!("{yaml}");
    Ok(())
}

fn print_table<T: Serialize>(data: &T) -> Result<()> {
    // Wrap in success envelope first (matches Go behavior)
    let sorted_data = sort_json_value(serde_json::to_value(data)?);
    let envelope = SuccessEnvelope {
        status: "success",
        data: &sorted_data,
    };
    let value = serde_json::to_value(&envelope)?;
    let rows = extract_rows(&value);

    if rows.is_empty() {
        println!("No results found");
        return Ok(());
    }

    // Collect headers from all rows
    let mut headers: Vec<String> = Vec::new();
    let mut header_set = std::collections::HashSet::new();
    for row in &rows {
        if let serde_json::Value::Object(map) = row {
            for key in map.keys() {
                if header_set.insert(key.clone()) {
                    headers.push(key.clone());
                }
            }
        }
    }

    // Prioritize common fields
    let priority = [
        "id",
        "title",
        "name",
        "type",
        "status",
        "state",
        "severity",
        "created_at",
        "updated_at",
        "created",
        "modified",
    ];
    let mut final_headers: Vec<String> = Vec::new();
    for &p in &priority {
        if header_set.contains(p) {
            final_headers.push(p.to_string());
        }
    }
    for h in &headers {
        if final_headers.len() >= 12 {
            break;
        }
        if !final_headers.contains(h) {
            final_headers.push(h.clone());
        }
    }

    let mut table = comfy_table::Table::new();
    table.set_header(&final_headers);

    for row in &rows {
        let cells: Vec<String> = final_headers
            .iter()
            .map(|h| {
                if let serde_json::Value::Object(map) = row {
                    format_cell(map.get(h.as_str()))
                } else {
                    String::new()
                }
            })
            .collect();
        table.add_row(cells);
    }

    println!("{table}");
    Ok(())
}

/// Extract displayable rows from a JSON value.
/// Handles: arrays, objects with "data" field, single objects.
fn extract_rows(value: &serde_json::Value) -> Vec<&serde_json::Value> {
    match value {
        serde_json::Value::Array(arr) => arr.iter().collect(),
        serde_json::Value::Object(map) => {
            // API responses often wrap data: { "data": [...], "meta": ... }
            if let Some(data) = map.get("data") {
                return extract_rows(data);
            }
            vec![value]
        }
        _ => vec![],
    }
}

fn format_cell(value: Option<&serde_json::Value>) -> String {
    match value {
        None | Some(serde_json::Value::Null) => String::new(),
        Some(serde_json::Value::String(s)) => {
            if s.len() > 50 {
                format!("{}...", &s[..47])
            } else {
                s.clone()
            }
        }
        Some(serde_json::Value::Number(n)) => n.to_string(),
        Some(serde_json::Value::Bool(b)) => b.to_string(),
        Some(serde_json::Value::Array(arr)) => {
            if arr.is_empty() {
                "[]".to_string()
            } else if arr.len() <= 3 {
                format!(
                    "[{}]",
                    arr.iter()
                        .map(|v| format_cell(Some(v)))
                        .collect::<Vec<_>>()
                        .join(", ")
                )
            } else {
                format!("[{} items]", arr.len())
            }
        }
        Some(serde_json::Value::Object(map)) => format!("{{{} fields}}", map.len()),
    }
}

/// Format an API error with contextual guidance.
#[allow(dead_code)]
pub fn format_api_error(operation: &str, status: Option<u16>, body: Option<&str>) -> String {
    let mut msg = format!("failed to {operation}");

    if let Some(code) = status {
        msg.push_str(&format!(" (HTTP {code})"));
    }

    if let Some(body) = body {
        if !body.is_empty() {
            msg.push_str(&format!("\nAPI response: {body}"));
        }
    }

    if let Some(code) = status {
        let hint = match code {
            500.. => "API server error — try again later",
            429 => "rate limited — wait and retry",
            403 => "access denied — check permissions",
            401 => "authentication failed — check credentials or run 'pup auth login'",
            404 => "resource not found — verify the ID",
            400 => "invalid request — check parameters",
            _ => "",
        };
        if !hint.is_empty() {
            msg.push_str(&format!("\nHint: {hint}"));
        }
    }

    msg
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_format_cell_string() {
        assert_eq!(format_cell(Some(&serde_json::json!("hello"))), "hello");
    }

    #[test]
    fn test_format_cell_long_string() {
        let long = "a".repeat(60);
        let result = format_cell(Some(&serde_json::json!(long)));
        assert_eq!(result.len(), 50);
        assert!(result.ends_with("..."));
    }

    #[test]
    fn test_format_cell_number() {
        assert_eq!(format_cell(Some(&serde_json::json!(42))), "42");
        assert_eq!(format_cell(Some(&serde_json::json!(3.14))), "3.14");
    }

    #[test]
    fn test_format_cell_null() {
        assert_eq!(format_cell(Some(&serde_json::Value::Null)), "");
        assert_eq!(format_cell(None), "");
    }

    #[test]
    fn test_format_cell_array() {
        assert_eq!(format_cell(Some(&serde_json::json!([]))), "[]");
        assert_eq!(format_cell(Some(&serde_json::json!([1, 2]))), "[1, 2]");
        assert_eq!(
            format_cell(Some(&serde_json::json!([1, 2, 3, 4, 5]))),
            "[5 items]"
        );
    }

    #[test]
    fn test_format_cell_object() {
        assert_eq!(
            format_cell(Some(&serde_json::json!({"a": 1, "b": 2}))),
            "{2 fields}"
        );
    }

    #[test]
    fn test_extract_rows_array() {
        let val = serde_json::json!([{"id": 1}, {"id": 2}]);
        assert_eq!(extract_rows(&val).len(), 2);
    }

    #[test]
    fn test_extract_rows_data_wrapper() {
        let val = serde_json::json!({"data": [{"id": 1}], "meta": {}});
        assert_eq!(extract_rows(&val).len(), 1);
    }

    #[test]
    fn test_extract_rows_single_object() {
        let val = serde_json::json!({"id": 1, "name": "test"});
        assert_eq!(extract_rows(&val).len(), 1);
    }

    #[test]
    fn test_format_api_error_basic() {
        let msg = format_api_error("list monitors", None, None);
        assert_eq!(msg, "failed to list monitors");
    }

    #[test]
    fn test_format_api_error_with_status() {
        let msg = format_api_error("list monitors", Some(403), None);
        assert!(msg.contains("HTTP 403"));
        assert!(msg.contains("access denied"));
    }

    #[test]
    fn test_format_api_error_with_body() {
        let msg = format_api_error("get user", Some(404), Some("not found"));
        assert!(msg.contains("not found"));
        assert!(msg.contains("resource not found"));
    }

    #[test]
    fn test_format_api_error_server_error() {
        let msg = format_api_error("query", Some(500), None);
        assert!(msg.contains("API server error"));
    }

    #[test]
    fn test_format_api_error_rate_limit() {
        let msg = format_api_error("query", Some(429), None);
        assert!(msg.contains("rate limited"));
    }
}
