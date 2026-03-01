use anyhow::Result;

use crate::config::Config;
use crate::ops::vet;

pub async fn run(
    cfg: &Config,
    tags: Option<String>,
    check: Option<String>,
    severity: Option<String>,
) -> Result<()> {
    let result = vet::run(cfg, tags, check, severity).await?;

    if cfg.agent_mode {
        let meta = crate::formatter::Metadata {
            count: Some(result.findings.len()),
            truncated: false,
            command: Some("vet".to_string()),
            next_action: if result.critical > 0 {
                Some("Address CRITICAL findings first".to_string())
            } else {
                None
            },
        };
        crate::formatter::format_and_print(&result, &cfg.output_format, true, Some(&meta))?;
        return Ok(());
    }

    // Human-readable output
    if result.findings.is_empty() {
        println!("All checks passed.");
    } else {
        for finding in &result.findings {
            println!(
                "\n{}: {} ({} found)",
                finding.severity.label(),
                finding.check,
                finding.count
            );
            for r in &finding.resources {
                println!("  - #{} \"{}\" ({})", r.id, r.name, r.detail);
            }
            println!("  -> {}", finding.recommendation);
        }
    }

    if !result.passed.is_empty() {
        println!("\nPASSED: {}", result.passed.join(", "));
    }

    println!(
        "\nSummary: {} critical, {} warning, {} passed",
        result.critical,
        result.warnings,
        result.passed.len()
    );

    // Exit non-zero if there are critical findings
    if result.critical > 0 {
        std::process::exit(1);
    }

    Ok(())
}

pub fn list_checks() {
    println!("{:<25} {:<10} DESCRIPTION", "CHECK", "SEVERITY");
    println!("{}", "-".repeat(80));
    for (name, severity, desc) in vet::list_checks() {
        println!("{:<25} {:<10} {}", name, severity.label(), desc);
    }
}
