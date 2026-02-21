/// Version is set at build time via env var or defaults to Cargo package version.
pub const VERSION: &str = env!("CARGO_PKG_VERSION");

pub fn build_info() -> String {
    format!(
        "Pup {} (rust {}; {} {})",
        VERSION,
        rustc_version(),
        std::env::consts::OS,
        std::env::consts::ARCH,
    )
}

fn rustc_version() -> &'static str {
    option_env!("RUSTC_VERSION").unwrap_or("unknown")
}
