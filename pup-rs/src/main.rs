mod auth;
mod client;
mod commands;
mod config;
mod formatter;
mod useragent;
mod util;
mod version;

use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(name = "pup", version = version::VERSION, about = "Datadog API CLI (Rust)")]
struct Cli {
    /// Output format (json, table, yaml)
    #[arg(short, long, global = true, default_value = "json")]
    output: String,
    /// Auto-approve destructive operations
    #[arg(short = 'y', long = "yes", global = true)]
    yes: bool,
    /// Enable agent mode
    #[arg(long, global = true)]
    agent: bool,
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Manage monitors
    Monitors { #[command(subcommand)] action: MonitorActions },
    /// Search and analyze logs
    Logs { #[command(subcommand)] action: LogActions },
    /// Manage incidents
    Incidents { #[command(subcommand)] action: IncidentActions },
    /// Manage dashboards
    Dashboards { #[command(subcommand)] action: DashboardActions },
    /// Query and manage metrics
    Metrics { #[command(subcommand)] action: MetricActions },
    /// Manage SLOs
    Slos { #[command(subcommand)] action: SloActions },
    /// Manage synthetics tests
    Synthetics { #[command(subcommand)] action: SyntheticsActions },
    /// Manage events
    Events { #[command(subcommand)] action: EventActions },
    /// Manage downtimes
    Downtime { #[command(subcommand)] action: DowntimeActions },
    /// Manage host tags
    Tags { #[command(subcommand)] action: TagActions },
    /// Manage users and roles
    Users { #[command(subcommand)] action: UserActions },
    /// Manage infrastructure hosts
    Infrastructure { #[command(subcommand)] action: InfraActions },
    /// Search audit logs
    #[command(name = "audit-logs")]
    AuditLogs { #[command(subcommand)] action: AuditLogActions },
    /// Manage security monitoring
    Security { #[command(subcommand)] action: SecurityActions },
    /// Manage organizations
    Organizations { #[command(subcommand)] action: OrgActions },
    /// Manage cloud integrations
    Cloud { #[command(subcommand)] action: CloudActions },
    /// Manage cases
    Cases { #[command(subcommand)] action: CaseActions },
    /// Manage service catalog
    #[command(name = "service-catalog")]
    ServiceCatalog { #[command(subcommand)] action: ServiceCatalogActions },
    /// Authentication (OAuth2)
    Auth { #[command(subcommand)] action: AuthActions },
    /// Show version information
    Version,
    /// Validate configuration
    Test,
}

// ---- Monitors ----
#[derive(Subcommand)]
enum MonitorActions {
    /// List monitors
    List {
        #[arg(long)] name: Option<String>,
        #[arg(long)] tags: Option<String>,
        #[arg(long, default_value_t = 200)] limit: i32,
    },
}

// ---- Logs ----
#[derive(Subcommand)]
enum LogActions {
    /// Search logs (forces API key auth)
    Search {
        #[arg(long)] query: String,
        #[arg(long, default_value = "1h")] from: String,
        #[arg(long, default_value = "now")] to: String,
        #[arg(long, default_value_t = 50)] limit: i32,
    },
}

// ---- Incidents ----
#[derive(Subcommand)]
enum IncidentActions {
    /// List incidents (unstable)
    List { #[arg(long, default_value_t = 50)] limit: i64 },
}

// ---- Dashboards ----
#[derive(Subcommand)]
enum DashboardActions {
    /// List all dashboards
    List,
    /// Get dashboard details
    Get { id: String },
    /// Delete a dashboard
    Delete { id: String },
}

// ---- Metrics ----
#[derive(Subcommand)]
enum MetricActions {
    /// List active metrics
    List {
        #[arg(long)] filter: Option<String>,
        #[arg(long, default_value = "1h")] from: String,
    },
    /// Query metrics (v1 API)
    Search {
        #[arg(long)] query: String,
        #[arg(long, default_value = "1h")] from: String,
        #[arg(long, default_value = "now")] to: String,
    },
    /// Get metric metadata
    Metadata {
        #[command(subcommand)] action: MetricMetadataActions,
    },
}

#[derive(Subcommand)]
enum MetricMetadataActions {
    /// Get metric metadata
    Get { metric_name: String },
}

// ---- SLOs ----
#[derive(Subcommand)]
enum SloActions {
    /// List all SLOs
    List,
    /// Get SLO details
    Get { id: String },
    /// Delete an SLO
    Delete { id: String },
}

// ---- Synthetics ----
#[derive(Subcommand)]
enum SyntheticsActions {
    /// Manage synthetic tests
    Tests { #[command(subcommand)] action: SyntheticsTestActions },
    /// Manage synthetic locations
    Locations { #[command(subcommand)] action: SyntheticsLocationActions },
}

#[derive(Subcommand)]
enum SyntheticsTestActions {
    /// List all tests
    List,
    /// Get test details
    Get { public_id: String },
    /// Search tests
    Search {
        #[arg(long)] text: Option<String>,
        #[arg(long, default_value_t = 50)] count: i64,
        #[arg(long, default_value_t = 0)] start: i64,
    },
}

#[derive(Subcommand)]
enum SyntheticsLocationActions {
    /// List test locations
    List,
}

// ---- Events ----
#[derive(Subcommand)]
enum EventActions {
    /// List events (v1 API)
    List {
        #[arg(long, default_value_t = 0)] start: i64,
        #[arg(long, default_value_t = 0)] end: i64,
        #[arg(long)] tags: Option<String>,
    },
    /// Search events (v2 API, requires API keys)
    Search {
        #[arg(long)] query: String,
        #[arg(long, default_value = "1h")] from: String,
        #[arg(long, default_value = "now")] to: String,
        #[arg(long, default_value_t = 100)] limit: i32,
    },
    /// Get event details
    Get { event_id: i64 },
}

// ---- Downtime ----
#[derive(Subcommand)]
enum DowntimeActions {
    /// List downtimes
    List,
    /// Get downtime details
    Get { id: String },
    /// Cancel a downtime
    Cancel { id: String },
}

// ---- Tags ----
#[derive(Subcommand)]
enum TagActions {
    /// List all host tags
    List,
    /// Get tags for a host
    Get { hostname: String },
    /// Add tags to a host
    Add { hostname: String, tags: Vec<String> },
    /// Update tags for a host
    Update { hostname: String, tags: Vec<String> },
    /// Delete all tags for a host
    Delete { hostname: String },
}

// ---- Users ----
#[derive(Subcommand)]
enum UserActions {
    /// List users
    List,
    /// Get user details
    Get { user_id: String },
    /// Manage roles
    Roles { #[command(subcommand)] action: UserRoleActions },
}

#[derive(Subcommand)]
enum UserRoleActions {
    /// List roles
    List,
}

// ---- Infrastructure ----
#[derive(Subcommand)]
enum InfraActions {
    /// Manage hosts
    Hosts { #[command(subcommand)] action: InfraHostActions },
}

#[derive(Subcommand)]
enum InfraHostActions {
    /// List hosts
    List {
        #[arg(long)] filter: Option<String>,
        #[arg(long, default_value = "status")] sort: String,
        #[arg(long, default_value_t = 100)] count: i64,
    },
}

// ---- Audit Logs ----
#[derive(Subcommand)]
enum AuditLogActions {
    /// List audit logs
    List {
        #[arg(long, default_value = "1h")] from: String,
        #[arg(long, default_value = "now")] to: String,
        #[arg(long, default_value_t = 100)] limit: i32,
    },
    /// Search audit logs
    Search {
        #[arg(long)] query: String,
        #[arg(long, default_value = "1h")] from: String,
        #[arg(long, default_value = "now")] to: String,
        #[arg(long, default_value_t = 100)] limit: i32,
    },
}

// ---- Security ----
#[derive(Subcommand)]
enum SecurityActions {
    /// Manage security rules
    Rules { #[command(subcommand)] action: SecurityRuleActions },
    /// Search security signals
    Signals { #[command(subcommand)] action: SecuritySignalActions },
}

#[derive(Subcommand)]
enum SecurityRuleActions {
    /// List rules
    List,
    /// Get rule details
    Get { rule_id: String },
}

#[derive(Subcommand)]
enum SecuritySignalActions {
    /// Search signals
    Search {
        #[arg(long)] query: String,
        #[arg(long, default_value = "1h")] from: String,
        #[arg(long, default_value = "now")] to: String,
        #[arg(long, default_value_t = 100)] limit: i32,
    },
}

// ---- Organizations ----
#[derive(Subcommand)]
enum OrgActions {
    /// List organizations
    List,
    /// Get current organization
    Get,
}

// ---- Cloud ----
#[derive(Subcommand)]
enum CloudActions {
    /// List AWS integrations
    Aws { #[command(subcommand)] action: CloudSubActions },
    /// List GCP integrations
    Gcp { #[command(subcommand)] action: CloudSubActions },
    /// List Azure integrations
    Azure { #[command(subcommand)] action: CloudSubActions },
}

#[derive(Subcommand)]
enum CloudSubActions {
    /// List integrations
    List,
}

// ---- Cases ----
#[derive(Subcommand)]
enum CaseActions {
    /// Search cases
    Search {
        #[arg(long)] query: Option<String>,
        #[arg(long, default_value_t = 50)] page_size: i64,
    },
    /// Get case details
    Get { case_id: String },
    /// Manage projects
    Projects { #[command(subcommand)] action: CaseProjectActions },
}

#[derive(Subcommand)]
enum CaseProjectActions {
    /// List projects
    List,
    /// Get project details
    Get { project_id: String },
}

// ---- Service Catalog ----
#[derive(Subcommand)]
enum ServiceCatalogActions {
    /// List services
    List,
    /// Get service details
    Get { service_name: String },
}

// ---- Auth ----
#[derive(Subcommand)]
enum AuthActions {
    /// Login via OAuth2
    Login,
    /// Logout
    Logout,
    /// Show auth status
    Status,
    /// Print access token
    Token,
}

// ---- Main ----

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let cli = Cli::parse();
    let mut cfg = config::Config::from_env()?;

    // Apply flag overrides
    if let Ok(fmt) = cli.output.parse() {
        cfg.output_format = fmt;
    }
    if cli.yes {
        cfg.auto_approve = true;
    }
    cfg.agent_mode = cli.agent || useragent::is_agent_mode();
    if cfg.agent_mode {
        cfg.auto_approve = true;
    }

    match cli.command {
        // --- Monitors ---
        Commands::Monitors { action } => {
            cfg.validate_auth()?;
            match action {
                MonitorActions::List { name, tags, limit } => {
                    commands::monitors::list(&cfg, name, tags, limit).await?;
                }
            }
        }
        // --- Logs ---
        Commands::Logs { action } => {
            cfg.validate_auth()?;
            match action {
                LogActions::Search { query, from, to, limit } => {
                    commands::logs::search(&cfg, query, from, to, limit).await?;
                }
            }
        }
        // --- Incidents ---
        Commands::Incidents { action } => {
            cfg.validate_auth()?;
            match action {
                IncidentActions::List { limit } => {
                    commands::incidents::list(&cfg, limit).await?;
                }
            }
        }
        // --- Dashboards ---
        Commands::Dashboards { action } => {
            cfg.validate_auth()?;
            match action {
                DashboardActions::List => commands::dashboards::list(&cfg).await?,
                DashboardActions::Get { id } => commands::dashboards::get(&cfg, &id).await?,
                DashboardActions::Delete { id } => commands::dashboards::delete(&cfg, &id).await?,
            }
        }
        // --- Metrics ---
        Commands::Metrics { action } => {
            cfg.validate_auth()?;
            match action {
                MetricActions::List { filter, from } => {
                    commands::metrics::list(&cfg, filter, from).await?;
                }
                MetricActions::Search { query, from, to } => {
                    commands::metrics::search(&cfg, query, from, to).await?;
                }
                MetricActions::Metadata { action } => match action {
                    MetricMetadataActions::Get { metric_name } => {
                        commands::metrics::metadata_get(&cfg, &metric_name).await?;
                    }
                },
            }
        }
        // --- SLOs ---
        Commands::Slos { action } => {
            cfg.validate_auth()?;
            match action {
                SloActions::List => commands::slos::list(&cfg).await?,
                SloActions::Get { id } => commands::slos::get(&cfg, &id).await?,
                SloActions::Delete { id } => commands::slos::delete(&cfg, &id).await?,
            }
        }
        // --- Synthetics ---
        Commands::Synthetics { action } => {
            cfg.validate_auth()?;
            match action {
                SyntheticsActions::Tests { action } => match action {
                    SyntheticsTestActions::List => commands::synthetics::tests_list(&cfg).await?,
                    SyntheticsTestActions::Get { public_id } => {
                        commands::synthetics::tests_get(&cfg, &public_id).await?;
                    }
                    SyntheticsTestActions::Search { text, count, start } => {
                        commands::synthetics::tests_search(&cfg, text, count, start).await?;
                    }
                },
                SyntheticsActions::Locations { action } => match action {
                    SyntheticsLocationActions::List => {
                        commands::synthetics::locations_list(&cfg).await?;
                    }
                },
            }
        }
        // --- Events ---
        Commands::Events { action } => {
            cfg.validate_auth()?;
            match action {
                EventActions::List { start, end, tags } => {
                    commands::events::list(&cfg, start, end, tags).await?;
                }
                EventActions::Search { query, from, to, limit } => {
                    commands::events::search(&cfg, query, from, to, limit).await?;
                }
                EventActions::Get { event_id } => {
                    commands::events::get(&cfg, event_id).await?;
                }
            }
        }
        // --- Downtime ---
        Commands::Downtime { action } => {
            cfg.validate_auth()?;
            match action {
                DowntimeActions::List => commands::downtime::list(&cfg).await?,
                DowntimeActions::Get { id } => commands::downtime::get(&cfg, &id).await?,
                DowntimeActions::Cancel { id } => commands::downtime::cancel(&cfg, &id).await?,
            }
        }
        // --- Tags ---
        Commands::Tags { action } => {
            cfg.validate_auth()?;
            match action {
                TagActions::List => commands::tags::list(&cfg).await?,
                TagActions::Get { hostname } => commands::tags::get(&cfg, &hostname).await?,
                TagActions::Add { hostname, tags } => {
                    commands::tags::add(&cfg, &hostname, tags).await?;
                }
                TagActions::Update { hostname, tags } => {
                    commands::tags::update(&cfg, &hostname, tags).await?;
                }
                TagActions::Delete { hostname } => {
                    commands::tags::delete(&cfg, &hostname).await?;
                }
            }
        }
        // --- Users ---
        Commands::Users { action } => {
            cfg.validate_auth()?;
            match action {
                UserActions::List => commands::users::list(&cfg).await?,
                UserActions::Get { user_id } => commands::users::get(&cfg, &user_id).await?,
                UserActions::Roles { action } => match action {
                    UserRoleActions::List => commands::users::roles_list(&cfg).await?,
                },
            }
        }
        // --- Infrastructure ---
        Commands::Infrastructure { action } => {
            cfg.validate_auth()?;
            match action {
                InfraActions::Hosts { action } => match action {
                    InfraHostActions::List { filter, sort, count } => {
                        commands::infrastructure::hosts_list(&cfg, filter, sort, count).await?;
                    }
                },
            }
        }
        // --- Audit Logs ---
        Commands::AuditLogs { action } => {
            cfg.validate_auth()?;
            match action {
                AuditLogActions::List { from, to, limit } => {
                    commands::audit_logs::list(&cfg, from, to, limit).await?;
                }
                AuditLogActions::Search { query, from, to, limit } => {
                    commands::audit_logs::search(&cfg, query, from, to, limit).await?;
                }
            }
        }
        // --- Security ---
        Commands::Security { action } => {
            cfg.validate_auth()?;
            match action {
                SecurityActions::Rules { action } => match action {
                    SecurityRuleActions::List => commands::security::rules_list(&cfg).await?,
                    SecurityRuleActions::Get { rule_id } => {
                        commands::security::rules_get(&cfg, &rule_id).await?;
                    }
                },
                SecurityActions::Signals { action } => match action {
                    SecuritySignalActions::Search { query, from, to, limit } => {
                        commands::security::signals_search(&cfg, query, from, to, limit).await?;
                    }
                },
            }
        }
        // --- Organizations ---
        Commands::Organizations { action } => {
            cfg.validate_auth()?;
            match action {
                OrgActions::List => commands::organizations::list(&cfg).await?,
                OrgActions::Get => commands::organizations::get(&cfg).await?,
            }
        }
        // --- Cloud ---
        Commands::Cloud { action } => {
            cfg.validate_auth()?;
            match action {
                CloudActions::Aws { action } => match action {
                    CloudSubActions::List => commands::cloud::aws_list(&cfg).await?,
                },
                CloudActions::Gcp { action } => match action {
                    CloudSubActions::List => commands::cloud::gcp_list(&cfg).await?,
                },
                CloudActions::Azure { action } => match action {
                    CloudSubActions::List => commands::cloud::azure_list(&cfg).await?,
                },
            }
        }
        // --- Cases ---
        Commands::Cases { action } => {
            cfg.validate_auth()?;
            match action {
                CaseActions::Search { query, page_size } => {
                    commands::cases::search(&cfg, query, page_size).await?;
                }
                CaseActions::Get { case_id } => commands::cases::get(&cfg, &case_id).await?,
                CaseActions::Projects { action } => match action {
                    CaseProjectActions::List => commands::cases::projects_list(&cfg).await?,
                    CaseProjectActions::Get { project_id } => {
                        commands::cases::projects_get(&cfg, &project_id).await?;
                    }
                },
            }
        }
        // --- Service Catalog ---
        Commands::ServiceCatalog { action } => {
            cfg.validate_auth()?;
            match action {
                ServiceCatalogActions::List => commands::service_catalog::list(&cfg).await?,
                ServiceCatalogActions::Get { service_name } => {
                    commands::service_catalog::get(&cfg, &service_name).await?;
                }
            }
        }
        // --- Auth ---
        Commands::Auth { action } => match action {
            AuthActions::Login => commands::auth::login(&cfg).await?,
            AuthActions::Logout => commands::auth::logout(&cfg).await?,
            AuthActions::Status => commands::auth::status(&cfg)?,
            AuthActions::Token => commands::auth::token(&cfg)?,
        },
        // --- Utility ---
        Commands::Version => println!("{}", version::build_info()),
        Commands::Test => commands::test::run(&cfg)?,
    }

    Ok(())
}
