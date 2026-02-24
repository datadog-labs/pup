use anyhow::{bail, Result};
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::api_spans::SpansAPI;
#[cfg(not(target_arch = "wasm32"))]
use datadog_api_client::datadogV2::model::{
    SpansAggregateData, SpansAggregateRequest, SpansAggregateRequestAttributes,
    SpansAggregateRequestType, SpansAggregationFunction, SpansCompute, SpansGroupBy,
    SpansListRequest, SpansListRequestAttributes, SpansListRequestData, SpansListRequestPage,
    SpansListRequestType, SpansQueryFilter, SpansSort,
};

#[cfg(not(target_arch = "wasm32"))]
use crate::client;
use crate::config::Config;
use crate::formatter;
use crate::util;

/// Validate the sort parameter.
fn validate_sort(sort: &str) -> Result<()> {
    match sort {
        "timestamp" | "-timestamp" => Ok(()),
        _ => bail!(
            "invalid --sort value: {sort:?}\nExpected: timestamp (ascending) or -timestamp (descending)"
        ),
    }
}

/// Parse a compute string like "count", "avg(@duration)", "percentile(@duration, 99)"
/// into a (function_name, Option<metric>) pair as raw strings.
fn parse_compute_raw(input: &str) -> Result<(String, Option<String>)> {
    let input = input.trim();
    if input.is_empty() {
        bail!("--compute is required");
    }

    // Simple aggregations without a metric
    if input == "count" {
        return Ok(("count".into(), None));
    }

    // func(@field) pattern
    if let Some(paren) = input.find('(') {
        let func = &input[..paren];
        let rest = input[paren + 1..].trim_end_matches(')').trim();

        // Handle percentile(@field, N)
        if func == "percentile" {
            let parts: Vec<&str> = rest.splitn(2, ',').collect();
            if parts.len() != 2 {
                bail!("percentile requires field and value: percentile(@duration, 99)");
            }
            let metric = parts[0].trim().to_string();
            let pct: u32 = parts[1]
                .trim()
                .parse()
                .map_err(|_| anyhow::anyhow!("invalid percentile value: {}", parts[1].trim()))?;
            let agg_name = match pct {
                75 => "pc75",
                90 => "pc90",
                95 => "pc95",
                98 => "pc98",
                99 => "pc99",
                _ => bail!("unsupported percentile: {pct} (supported: 75, 90, 95, 98, 99)"),
            };
            return Ok((agg_name.into(), Some(metric)));
        }

        let metric = rest.to_string();
        let agg_name = match func {
            "avg" | "sum" | "min" | "max" | "median" | "cardinality" => func.to_string(),
            "count" => bail!("count does not accept a field argument; use just 'count'"),
            _ => bail!("unknown aggregation function: {func}"),
        };
        return Ok((agg_name, Some(metric)));
    }

    bail!(
        "invalid --compute format: {input:?}\n\
         Expected: count, avg(@duration), sum(@duration), percentile(@duration, 99), etc."
    )
}

/// Parse a compute string into (SpansAggregationFunction, Option<metric>).
#[cfg(not(target_arch = "wasm32"))]
fn parse_compute(input: &str) -> Result<(SpansAggregationFunction, Option<String>)> {
    let (func, metric) = parse_compute_raw(input)?;
    let agg = match func.as_str() {
        "count" => SpansAggregationFunction::COUNT,
        "avg" => SpansAggregationFunction::AVG,
        "sum" => SpansAggregationFunction::SUM,
        "min" => SpansAggregationFunction::MIN,
        "max" => SpansAggregationFunction::MAX,
        "median" => SpansAggregationFunction::MEDIAN,
        "cardinality" => SpansAggregationFunction::CARDINALITY,
        "pc75" => SpansAggregationFunction::PERCENTILE_75,
        "pc90" => SpansAggregationFunction::PERCENTILE_90,
        "pc95" => SpansAggregationFunction::PERCENTILE_95,
        "pc98" => SpansAggregationFunction::PERCENTILE_98,
        "pc99" => SpansAggregationFunction::PERCENTILE_99,
        _ => bail!("unknown aggregation function: {func}"),
    };
    Ok((agg, metric))
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
    sort: String,
) -> Result<()> {
    validate_sort(&sort)?;

    let dd_cfg = client::make_dd_config(cfg);
    let api = if let Some(bearer_client) = client::make_bearer_client(cfg) {
        SpansAPI::with_client_and_config(dd_cfg, bearer_client)
    } else {
        SpansAPI::with_config(dd_cfg)
    };

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;

    let page_limit = limit.min(1000);
    let spans_sort = match sort.as_str() {
        "timestamp" => SpansSort::TIMESTAMP_ASCENDING,
        _ => SpansSort::TIMESTAMP_DESCENDING,
    };

    let body = SpansListRequest::new().data(
        SpansListRequestData::new()
            .type_(SpansListRequestType::SEARCH_REQUEST)
            .attributes(
                SpansListRequestAttributes::new()
                    .filter(
                        SpansQueryFilter::new()
                            .query(query)
                            .from(from_ms.to_string())
                            .to(to_ms.to_string()),
                    )
                    .page(SpansListRequestPage::new().limit(page_limit))
                    .sort(spans_sort),
            ),
    );

    let resp = api
        .list_spans(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to search spans: {:?}", e))?;

    let meta = if cfg.agent_mode {
        let count = resp.data.as_ref().map(|d| d.len());
        let truncated = count.is_some_and(|c| c as i32 >= page_limit);
        Some(formatter::Metadata {
            count,
            truncated,
            command: Some("traces search".into()),
            next_action: if truncated {
                Some(format!(
                    "Results may be truncated at {page_limit}. Use --limit={} or narrow the --query",
                    page_limit + 1
                ))
            } else {
                None
            },
        })
    } else {
        None
    };
    formatter::format_and_print(&resp, &cfg.output_format, cfg.agent_mode, meta.as_ref())?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn search(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    limit: i32,
    sort: String,
) -> Result<()> {
    validate_sort(&sort)?;

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;
    let body = serde_json::json!({
        "data": {
            "attributes": {
                "filter": {
                    "query": query,
                    "from": from_ms.to_string(),
                    "to": to_ms.to_string()
                },
                "page": { "limit": limit },
                "sort": sort
            },
            "type": "search_request"
        }
    });
    let data = crate::api::post(cfg, "/api/v2/spans/events/search", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(not(target_arch = "wasm32"))]
pub async fn aggregate(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    compute: String,
    group_by: Option<String>,
) -> Result<()> {
    let (agg_fn, metric) = parse_compute(&compute)?;

    let dd_cfg = client::make_dd_config(cfg);
    let api = if let Some(bearer_client) = client::make_bearer_client(cfg) {
        SpansAPI::with_client_and_config(dd_cfg, bearer_client)
    } else {
        SpansAPI::with_config(dd_cfg)
    };

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;

    let mut spans_compute = SpansCompute::new(agg_fn);
    if let Some(m) = metric {
        spans_compute = spans_compute.metric(m);
    }

    let mut attrs = SpansAggregateRequestAttributes::new()
        .compute(vec![spans_compute])
        .filter(
            SpansQueryFilter::new()
                .query(query)
                .from(from_ms.to_string())
                .to(to_ms.to_string()),
        );

    if let Some(facet) = group_by {
        attrs = attrs.group_by(vec![SpansGroupBy::new(facet)]);
    }

    let body = SpansAggregateRequest::new().data(
        SpansAggregateData::new()
            .type_(SpansAggregateRequestType::AGGREGATE_REQUEST)
            .attributes(attrs),
    );

    let resp = api
        .aggregate_spans(body)
        .await
        .map_err(|e| anyhow::anyhow!("failed to aggregate spans: {:?}", e))?;

    let meta = if cfg.agent_mode {
        Some(formatter::Metadata {
            count: None,
            truncated: false,
            command: Some("traces aggregate".into()),
            next_action: None,
        })
    } else {
        None
    };
    formatter::format_and_print(&resp, &cfg.output_format, cfg.agent_mode, meta.as_ref())?;
    Ok(())
}

#[cfg(target_arch = "wasm32")]
pub async fn aggregate(
    cfg: &Config,
    query: String,
    from: String,
    to: String,
    compute: String,
    group_by: Option<String>,
) -> Result<()> {
    let (func, metric) = parse_compute_raw(&compute)?;

    let from_ms = util::parse_time_to_unix_millis(&from)?;
    let to_ms = util::parse_time_to_unix_millis(&to)?;

    let mut compute_obj = serde_json::json!({ "aggregation": func });
    if let Some(m) = metric {
        compute_obj["metric"] = serde_json::Value::String(m);
    }

    let mut body = serde_json::json!({
        "data": {
            "attributes": {
                "filter": {
                    "query": query,
                    "from": from_ms.to_string(),
                    "to": to_ms.to_string()
                },
                "compute": [compute_obj]
            },
            "type": "aggregate_request"
        }
    });

    if let Some(facet) = group_by {
        body["data"]["attributes"]["group_by"] = serde_json::json!([{ "facet": facet }]);
    }

    let data = crate::api::post(cfg, "/api/v2/spans/analytics/aggregate", &body).await?;
    crate::formatter::output(cfg, &data)
}

#[cfg(all(test, not(target_arch = "wasm32")))]
mod tests {
    use super::*;
    use datadog_api_client::datadogV2::model::SpansAggregationFunction;

    #[test]
    fn test_parse_compute_count() {
        let (agg, metric) = parse_compute("count").unwrap();
        assert_eq!(agg, SpansAggregationFunction::COUNT);
        assert!(metric.is_none());
    }

    #[test]
    fn test_parse_compute_avg() {
        let (agg, metric) = parse_compute("avg(@duration)").unwrap();
        assert_eq!(agg, SpansAggregationFunction::AVG);
        assert_eq!(metric.unwrap(), "@duration");
    }

    #[test]
    fn test_parse_compute_sum() {
        let (agg, metric) = parse_compute("sum(@duration)").unwrap();
        assert_eq!(agg, SpansAggregationFunction::SUM);
        assert_eq!(metric.unwrap(), "@duration");
    }

    #[test]
    fn test_parse_compute_min() {
        let (agg, metric) = parse_compute("min(@duration)").unwrap();
        assert_eq!(agg, SpansAggregationFunction::MIN);
        assert_eq!(metric.unwrap(), "@duration");
    }

    #[test]
    fn test_parse_compute_max() {
        let (agg, metric) = parse_compute("max(@duration)").unwrap();
        assert_eq!(agg, SpansAggregationFunction::MAX);
        assert_eq!(metric.unwrap(), "@duration");
    }

    #[test]
    fn test_parse_compute_median() {
        let (agg, metric) = parse_compute("median(@duration)").unwrap();
        assert_eq!(agg, SpansAggregationFunction::MEDIAN);
        assert_eq!(metric.unwrap(), "@duration");
    }

    #[test]
    fn test_parse_compute_cardinality() {
        let (agg, metric) = parse_compute("cardinality(@usr.id)").unwrap();
        assert_eq!(agg, SpansAggregationFunction::CARDINALITY);
        assert_eq!(metric.unwrap(), "@usr.id");
    }

    #[test]
    fn test_parse_compute_percentile_99() {
        let (agg, metric) = parse_compute("percentile(@duration, 99)").unwrap();
        assert_eq!(agg, SpansAggregationFunction::PERCENTILE_99);
        assert_eq!(metric.unwrap(), "@duration");
    }

    #[test]
    fn test_parse_compute_percentile_95() {
        let (agg, metric) = parse_compute("percentile(@duration, 95)").unwrap();
        assert_eq!(agg, SpansAggregationFunction::PERCENTILE_95);
        assert_eq!(metric.unwrap(), "@duration");
    }

    #[test]
    fn test_parse_compute_percentile_90() {
        let (agg, metric) = parse_compute("percentile(@duration, 90)").unwrap();
        assert_eq!(agg, SpansAggregationFunction::PERCENTILE_90);
        assert_eq!(metric.unwrap(), "@duration");
    }

    #[test]
    fn test_parse_compute_empty() {
        assert!(parse_compute("").is_err());
    }

    #[test]
    fn test_parse_compute_invalid() {
        assert!(parse_compute("invalid").is_err());
    }

    #[test]
    fn test_parse_compute_unknown_function() {
        assert!(parse_compute("foo(@bar)").is_err());
    }

    #[test]
    fn test_parse_compute_unsupported_percentile() {
        assert!(parse_compute("percentile(@duration, 50)").is_err());
    }

    #[test]
    fn test_parse_compute_percentile_missing_value() {
        assert!(parse_compute("percentile(@duration)").is_err());
    }

    #[test]
    fn test_parse_compute_count_with_field_rejected() {
        let err = parse_compute("count(@duration)").unwrap_err();
        assert!(err.to_string().contains("does not accept a field"));
    }

    #[test]
    fn test_validate_sort_valid() {
        assert!(validate_sort("timestamp").is_ok());
        assert!(validate_sort("-timestamp").is_ok());
    }

    #[test]
    fn test_validate_sort_invalid() {
        assert!(validate_sort("garbage").is_err());
        assert!(validate_sort("").is_err());
        assert!(validate_sort("asc").is_err());
    }
}
