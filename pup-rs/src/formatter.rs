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

/// Agent mode wrapper: { data, metadata }
#[derive(Serialize)]
struct AgentEnvelope<'a, T: Serialize> {
    data: &'a T,
    #[serde(skip_serializing_if = "Option::is_none")]
    metadata: Option<&'a Metadata>,
}

/// Format and print data to stdout.
pub fn format_and_print<T: Serialize>(
    data: &T,
    format: &OutputFormat,
    agent_mode: bool,
    meta: Option<&Metadata>,
) -> Result<()> {
    if agent_mode {
        // Agent mode always outputs JSON with envelope
        let envelope = AgentEnvelope {
            data,
            metadata: meta,
        };
        let json = serde_json::to_string_pretty(&envelope)?;
        println!("{json}");
        return Ok(());
    }

    match format {
        OutputFormat::Json => print_json(data),
        OutputFormat::Yaml => print_yaml(data),
        OutputFormat::Table => print_table(data),
    }
}

pub fn print_json<T: Serialize>(data: &T) -> Result<()> {
    let json = serde_json::to_string_pretty(data)?;
    println!("{json}");
    Ok(())
}

fn print_yaml<T: Serialize>(data: &T) -> Result<()> {
    let yaml = serde_yaml::to_string(data)?;
    print!("{yaml}");
    Ok(())
}

fn print_table<T: Serialize>(data: &T) -> Result<()> {
    // Convert to serde_json::Value to inspect structure
    let value = serde_json::to_value(data)?;
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
        "id", "title", "name", "type", "status", "state", "severity",
        "created_at", "updated_at", "created", "modified",
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
