use anyhow::{bail, Result};
use chrono::Utc;
use regex::Regex;

/// Parses a time string into Unix milliseconds.
///
/// Supported formats:
///   - "now" (case-insensitive)
///   - Relative short: "1h", "30m", "7d", "5s", "1w"
///   - Relative long: "5min", "5mins", "5minute", "5minutes", "2hr", "2hours", "3days", "1week"
///   - With spaces: "5 minutes", "2 hours"
///   - With leading minus: "-5m", "-2h"
///   - Unix timestamp (all digits, assumed milliseconds)
///   - RFC3339: "2024-01-01T00:00:00Z"
///
/// All relative times are interpreted as "ago from now".
/// Returns second-aligned milliseconds (Unix seconds * 1000) to match Go behavior.
pub fn parse_time_to_unix_millis(input: &str) -> Result<i64> {
    let input = input.trim();

    // "now" (case-insensitive)
    if input.eq_ignore_ascii_case("now") {
        return Ok(now_millis());
    }

    // Unix timestamp (all digits)
    if !input.is_empty() && input.chars().all(|c| c.is_ascii_digit()) {
        return Ok(input.parse()?);
    }

    // RFC3339 timestamp
    if input.contains('T') {
        let dt = chrono::DateTime::parse_from_rfc3339(input)?;
        return Ok(dt.timestamp() * 1000);
    }

    // Relative time — strip leading minus
    let stripped = input.trim_start_matches('-').trim();

    let re = Regex::new(
        r"(?i)^(\d+)\s*(s|sec|secs|second|seconds|m|min|mins|minute|minutes|h|hr|hrs|hour|hours|d|day|days|w|week|weeks)$",
    )
    .unwrap();

    if let Some(caps) = re.captures(stripped) {
        let num: i64 = caps[1].parse()?;
        let unit = caps[2].to_lowercase();
        let seconds = match unit.as_str() {
            "s" | "sec" | "secs" | "second" | "seconds" => num,
            "m" | "min" | "mins" | "minute" | "minutes" => num * 60,
            "h" | "hr" | "hrs" | "hour" | "hours" => num * 3600,
            "d" | "day" | "days" => num * 86400,
            "w" | "week" | "weeks" => num * 7 * 86400,
            _ => bail!("unknown time unit: {}", unit),
        };
        // Second-aligned: Unix seconds * 1000 (matches Go behavior)
        return Ok((Utc::now().timestamp() - seconds) * 1000);
    }

    bail!(
        "unable to parse time: {input:?}\n\
         Expected: now, 1h, 30m, 7d, 5minutes, RFC3339, or Unix timestamp"
    )
}

/// Convenience: parse to Unix seconds.
pub fn parse_time_to_unix(input: &str) -> Result<i64> {
    Ok(parse_time_to_unix_millis(input)? / 1000)
}

fn now_millis() -> i64 {
    Utc::now().timestamp() * 1000
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_now() {
        let ms = parse_time_to_unix_millis("now").unwrap();
        let diff = (Utc::now().timestamp() * 1000 - ms).abs();
        assert!(diff < 2000, "now should be within 2s: diff={diff}ms");
    }

    #[test]
    fn test_now_case_insensitive() {
        assert!(parse_time_to_unix_millis("NOW").is_ok());
        assert!(parse_time_to_unix_millis("Now").is_ok());
    }

    #[test]
    fn test_relative_short() {
        let ms = parse_time_to_unix_millis("1h").unwrap();
        let expected = (Utc::now().timestamp() - 3600) * 1000;
        assert!((ms - expected).abs() < 2000);
    }

    #[test]
    fn test_relative_long() {
        let ms = parse_time_to_unix_millis("5minutes").unwrap();
        let expected = (Utc::now().timestamp() - 300) * 1000;
        assert!((ms - expected).abs() < 2000);
    }

    #[test]
    fn test_relative_with_spaces() {
        let ms = parse_time_to_unix_millis("5 minutes").unwrap();
        let expected = (Utc::now().timestamp() - 300) * 1000;
        assert!((ms - expected).abs() < 2000);
    }

    #[test]
    fn test_relative_with_minus() {
        let ms = parse_time_to_unix_millis("-30m").unwrap();
        let expected = (Utc::now().timestamp() - 1800) * 1000;
        assert!((ms - expected).abs() < 2000);
    }

    #[test]
    fn test_unix_timestamp() {
        let ms = parse_time_to_unix_millis("1700000000000").unwrap();
        assert_eq!(ms, 1700000000000);
    }

    #[test]
    fn test_rfc3339() {
        let ms = parse_time_to_unix_millis("2024-01-01T00:00:00Z").unwrap();
        assert_eq!(ms, 1704067200000);
    }

    #[test]
    fn test_invalid() {
        assert!(parse_time_to_unix_millis("invalid").is_err());
        assert!(parse_time_to_unix_millis("").is_err());
    }
}
