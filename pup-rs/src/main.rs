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
    /// Manage API keys
    #[command(name = "api-keys")]
    ApiKeys { #[command(subcommand)] action: ApiKeyActions },
    /// Manage application keys
    #[command(name = "app-keys")]
    AppKeys { #[command(subcommand)] action: AppKeyActions },
    /// Query usage data
    Usage { #[command(subcommand)] action: UsageActions },
    /// Manage notebooks
    Notebooks { #[command(subcommand)] action: NotebookActions },
    /// RUM (Real User Monitoring)
    Rum { #[command(subcommand)] action: RumActions },
    /// CI/CD visibility
    Cicd { #[command(subcommand)] action: CicdActions },
    /// Manage on-call teams
    #[command(name = "on-call")]
    OnCall { #[command(subcommand)] action: OnCallActions },
    /// Fleet automation
    Fleet { #[command(subcommand)] action: FleetActions },
    /// Data governance
    #[command(name = "data-governance")]
    DataGovernance { #[command(subcommand)] action: DataGovActions },
    /// Error tracking
    #[command(name = "error-tracking")]
    ErrorTracking { #[command(subcommand)] action: ErrorTrackingActions },
    /// Code coverage
    #[command(name = "code-coverage")]
    CodeCoverage { #[command(subcommand)] action: CodeCoverageActions },
    /// HAMR (High Availability Multi-Region)
    Hamr { #[command(subcommand)] action: HamrActions },
    /// Status pages
    #[command(name = "status-pages")]
    StatusPages { #[command(subcommand)] action: StatusPageActions },
    /// Manage integrations (Jira, ServiceNow)
    Integrations { #[command(subcommand)] action: IntegrationActions },
    /// Cloud cost management
    Cost { #[command(subcommand)] action: CostActions },
    /// Miscellaneous (IP ranges)
    Misc { #[command(subcommand)] action: MiscActions },
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

// ---- API Keys ----
#[derive(Subcommand)]
enum ApiKeyActions {
    /// List API keys
    List,
    /// Get API key details
    Get { key_id: String },
}

// ---- App Keys ----
#[derive(Subcommand)]
enum AppKeyActions {
    /// List application keys
    List,
    /// Get application key details
    Get { key_id: String },
}

// ---- Usage ----
#[derive(Subcommand)]
enum UsageActions {
    /// Get usage summary
    Summary {
        #[arg(long, default_value = "30d")] start: String,
        #[arg(long)] end: Option<String>,
    },
    /// Get hourly usage
    Hourly {
        #[arg(long, default_value = "1d")] start: String,
        #[arg(long)] end: Option<String>,
    },
}

// ---- Notebooks ----
#[derive(Subcommand)]
enum NotebookActions {
    /// List notebooks
    List,
    /// Get notebook details
    Get { notebook_id: i64 },
}

// ---- RUM ----
#[derive(Subcommand)]
enum RumActions {
    /// Manage RUM applications
    Apps { #[command(subcommand)] action: RumAppActions },
    /// List RUM events
    Events {
        #[arg(long, default_value = "1h")] from: String,
        #[arg(long, default_value = "now")] to: String,
        #[arg(long, default_value_t = 100)] limit: i32,
    },
}

#[derive(Subcommand)]
enum RumAppActions {
    /// List RUM apps (requires API keys)
    List,
    /// Get RUM app details (requires API keys)
    Get { app_id: String },
}

// ---- CI/CD ----
#[derive(Subcommand)]
enum CicdActions {
    /// List CI pipelines
    Pipelines {
        #[arg(long)] query: Option<String>,
        #[arg(long, default_value = "1h")] from: String,
        #[arg(long, default_value = "now")] to: String,
        #[arg(long, default_value_t = 50)] limit: i32,
    },
    /// List CI tests
    Tests {
        #[arg(long)] query: Option<String>,
        #[arg(long, default_value = "1h")] from: String,
        #[arg(long, default_value = "now")] to: String,
        #[arg(long, default_value_t = 50)] limit: i32,
    },
}

// ---- On-Call ----
#[derive(Subcommand)]
enum OnCallActions {
    /// Manage on-call teams
    Teams { #[command(subcommand)] action: OnCallTeamActions },
}

#[derive(Subcommand)]
enum OnCallTeamActions {
    /// List teams
    List,
    /// Get team details
    Get { team_id: String },
    /// Delete a team
    Delete { team_id: String },
    /// List team memberships
    Memberships {
        team_id: String,
        #[arg(long, default_value_t = 100)] page_size: i64,
    },
}

// ---- Fleet ----
#[derive(Subcommand)]
enum FleetActions {
    /// Manage fleet agents
    Agents { #[command(subcommand)] action: FleetAgentActions },
    /// Manage fleet deployments
    Deployments { #[command(subcommand)] action: FleetDeploymentActions },
    /// Manage fleet schedules
    Schedules { #[command(subcommand)] action: FleetScheduleActions },
}

#[derive(Subcommand)]
enum FleetAgentActions {
    /// List agents
    List { #[arg(long)] page_size: Option<i64> },
    /// Get agent details
    Get { agent_key: String },
    /// List agent versions
    Versions,
}

#[derive(Subcommand)]
enum FleetDeploymentActions {
    /// List deployments
    List { #[arg(long)] page_size: Option<i64> },
    /// Get deployment details
    Get { deployment_id: String },
}

#[derive(Subcommand)]
enum FleetScheduleActions {
    /// List schedules
    List,
    /// Get schedule details
    Get { schedule_id: String },
}

// ---- Data Governance ----
#[derive(Subcommand)]
enum DataGovActions {
    /// List scanner rules
    #[command(name = "scanner-rules")]
    ScannerRules { #[command(subcommand)] action: DataGovScannerActions },
}

#[derive(Subcommand)]
enum DataGovScannerActions {
    /// List rules
    List,
}

// ---- Error Tracking ----
#[derive(Subcommand)]
enum ErrorTrackingActions {
    /// Manage error tracking issues
    Issues { #[command(subcommand)] action: ErrorTrackingIssueActions },
}

#[derive(Subcommand)]
enum ErrorTrackingIssueActions {
    /// Search issues
    Search {
        #[arg(long)] query: Option<String>,
        #[arg(long, default_value_t = 50)] limit: i32,
    },
    /// Get issue details
    Get { issue_id: String },
}

// ---- Code Coverage ----
#[derive(Subcommand)]
enum CodeCoverageActions {
    /// Get branch coverage summary
    #[command(name = "branch-summary")]
    BranchSummary {
        #[arg(long)] repo: String,
        #[arg(long)] branch: String,
    },
    /// Get commit coverage summary
    #[command(name = "commit-summary")]
    CommitSummary {
        #[arg(long)] repo: String,
        #[arg(long)] commit: String,
    },
}

// ---- HAMR ----
#[derive(Subcommand)]
enum HamrActions {
    /// Manage HAMR connections
    Connections { #[command(subcommand)] action: HamrConnectionActions },
}

#[derive(Subcommand)]
enum HamrConnectionActions {
    /// Get HAMR org connection
    Get,
}

// ---- Status Pages ----
#[derive(Subcommand)]
enum StatusPageActions {
    /// Manage status pages
    Pages { #[command(subcommand)] action: StatusPagePageActions },
    /// Manage page components
    Components { #[command(subcommand)] action: StatusPageComponentActions },
    /// Manage degradations
    Degradations { #[command(subcommand)] action: StatusPageDegradationActions },
}

#[derive(Subcommand)]
enum StatusPagePageActions {
    /// List status pages
    List,
    /// Get status page details
    Get { page_id: String },
    /// Delete a status page
    Delete { page_id: String },
}

#[derive(Subcommand)]
enum StatusPageComponentActions {
    /// List components
    List { page_id: String },
    /// Get component details
    Get { page_id: String, component_id: String },
}

#[derive(Subcommand)]
enum StatusPageDegradationActions {
    /// List degradations
    List,
    /// Get degradation details
    Get { page_id: String, degradation_id: String },
}

// ---- Integrations ----
#[derive(Subcommand)]
enum IntegrationActions {
    /// Jira integration
    Jira { #[command(subcommand)] action: JiraActions },
    /// ServiceNow integration
    Servicenow { #[command(subcommand)] action: ServiceNowActions },
}

#[derive(Subcommand)]
enum JiraActions {
    /// List Jira accounts
    Accounts,
    /// Manage Jira templates
    Templates { #[command(subcommand)] action: JiraTemplateActions },
}

#[derive(Subcommand)]
enum JiraTemplateActions {
    /// List templates
    List,
    /// Get template details
    Get { template_id: String },
}

#[derive(Subcommand)]
enum ServiceNowActions {
    /// List ServiceNow instances
    Instances,
    /// Manage ServiceNow templates
    Templates { #[command(subcommand)] action: ServiceNowTemplateActions },
}

#[derive(Subcommand)]
enum ServiceNowTemplateActions {
    /// List templates
    List,
    /// Get template details
    Get { template_id: String },
}

// ---- Cost ----
#[derive(Subcommand)]
enum CostActions {
    /// Get projected cost
    Projected,
    /// Get cost by org
    #[command(name = "by-org")]
    ByOrg {
        #[arg(long)] start_month: String,
        #[arg(long)] end_month: Option<String>,
    },
}

// ---- Misc ----
#[derive(Subcommand)]
enum MiscActions {
    /// Get IP ranges
    #[command(name = "ip-ranges")]
    IpRanges,
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
        // --- API Keys ---
        Commands::ApiKeys { action } => {
            cfg.validate_auth()?;
            match action {
                ApiKeyActions::List => commands::api_keys::list(&cfg).await?,
                ApiKeyActions::Get { key_id } => commands::api_keys::get(&cfg, &key_id).await?,
            }
        }
        // --- App Keys ---
        Commands::AppKeys { action } => {
            cfg.validate_auth()?;
            match action {
                AppKeyActions::List => commands::app_keys::list(&cfg).await?,
                AppKeyActions::Get { key_id } => commands::app_keys::get(&cfg, &key_id).await?,
            }
        }
        // --- Usage ---
        Commands::Usage { action } => {
            cfg.validate_auth()?;
            match action {
                UsageActions::Summary { start, end } => {
                    commands::usage::summary(&cfg, start, end).await?;
                }
                UsageActions::Hourly { start, end } => {
                    commands::usage::hourly(&cfg, start, end).await?;
                }
            }
        }
        // --- Notebooks ---
        Commands::Notebooks { action } => {
            cfg.validate_auth()?;
            match action {
                NotebookActions::List => commands::notebooks::list(&cfg).await?,
                NotebookActions::Get { notebook_id } => {
                    commands::notebooks::get(&cfg, notebook_id).await?;
                }
            }
        }
        // --- RUM ---
        Commands::Rum { action } => {
            cfg.validate_auth()?;
            match action {
                RumActions::Apps { action } => match action {
                    RumAppActions::List => commands::rum::apps_list(&cfg).await?,
                    RumAppActions::Get { app_id } => commands::rum::apps_get(&cfg, &app_id).await?,
                },
                RumActions::Events { from, to, limit } => {
                    commands::rum::events_list(&cfg, from, to, limit).await?;
                }
            }
        }
        // --- CI/CD ---
        Commands::Cicd { action } => {
            cfg.validate_auth()?;
            match action {
                CicdActions::Pipelines { query, from, to, limit } => {
                    commands::cicd::pipelines_list(&cfg, query, from, to, limit).await?;
                }
                CicdActions::Tests { query, from, to, limit } => {
                    commands::cicd::tests_list(&cfg, query, from, to, limit).await?;
                }
            }
        }
        // --- On-Call ---
        Commands::OnCall { action } => {
            cfg.validate_auth()?;
            match action {
                OnCallActions::Teams { action } => match action {
                    OnCallTeamActions::List => commands::on_call::teams_list(&cfg).await?,
                    OnCallTeamActions::Get { team_id } => {
                        commands::on_call::teams_get(&cfg, &team_id).await?;
                    }
                    OnCallTeamActions::Delete { team_id } => {
                        commands::on_call::teams_delete(&cfg, &team_id).await?;
                    }
                    OnCallTeamActions::Memberships { team_id, page_size } => {
                        commands::on_call::memberships_list(&cfg, &team_id, page_size).await?;
                    }
                },
            }
        }
        // --- Fleet ---
        Commands::Fleet { action } => {
            cfg.validate_auth()?;
            match action {
                FleetActions::Agents { action } => match action {
                    FleetAgentActions::List { page_size } => {
                        commands::fleet::agents_list(&cfg, page_size).await?;
                    }
                    FleetAgentActions::Get { agent_key } => {
                        commands::fleet::agents_get(&cfg, &agent_key).await?;
                    }
                    FleetAgentActions::Versions => commands::fleet::agents_versions(&cfg).await?,
                },
                FleetActions::Deployments { action } => match action {
                    FleetDeploymentActions::List { page_size } => {
                        commands::fleet::deployments_list(&cfg, page_size).await?;
                    }
                    FleetDeploymentActions::Get { deployment_id } => {
                        commands::fleet::deployments_get(&cfg, &deployment_id).await?;
                    }
                },
                FleetActions::Schedules { action } => match action {
                    FleetScheduleActions::List => commands::fleet::schedules_list(&cfg).await?,
                    FleetScheduleActions::Get { schedule_id } => {
                        commands::fleet::schedules_get(&cfg, &schedule_id).await?;
                    }
                },
            }
        }
        // --- Data Governance ---
        Commands::DataGovernance { action } => {
            cfg.validate_auth()?;
            match action {
                DataGovActions::ScannerRules { action } => match action {
                    DataGovScannerActions::List => {
                        commands::data_governance::scanner_rules_list(&cfg).await?;
                    }
                },
            }
        }
        // --- Error Tracking ---
        Commands::ErrorTracking { action } => {
            cfg.validate_auth()?;
            match action {
                ErrorTrackingActions::Issues { action } => match action {
                    ErrorTrackingIssueActions::Search { query, limit } => {
                        commands::error_tracking::issues_search(&cfg, query, limit).await?;
                    }
                    ErrorTrackingIssueActions::Get { issue_id } => {
                        commands::error_tracking::issues_get(&cfg, &issue_id).await?;
                    }
                },
            }
        }
        // --- Code Coverage ---
        Commands::CodeCoverage { action } => {
            cfg.validate_auth()?;
            match action {
                CodeCoverageActions::BranchSummary { repo, branch } => {
                    commands::code_coverage::branch_summary(&cfg, repo, branch).await?;
                }
                CodeCoverageActions::CommitSummary { repo, commit } => {
                    commands::code_coverage::commit_summary(&cfg, repo, commit).await?;
                }
            }
        }
        // --- HAMR ---
        Commands::Hamr { action } => {
            cfg.validate_auth()?;
            match action {
                HamrActions::Connections { action } => match action {
                    HamrConnectionActions::Get => commands::hamr::connections_get(&cfg).await?,
                },
            }
        }
        // --- Status Pages ---
        Commands::StatusPages { action } => {
            cfg.validate_auth()?;
            match action {
                StatusPageActions::Pages { action } => match action {
                    StatusPagePageActions::List => commands::status_pages::pages_list(&cfg).await?,
                    StatusPagePageActions::Get { page_id } => {
                        commands::status_pages::pages_get(&cfg, &page_id).await?;
                    }
                    StatusPagePageActions::Delete { page_id } => {
                        commands::status_pages::pages_delete(&cfg, &page_id).await?;
                    }
                },
                StatusPageActions::Components { action } => match action {
                    StatusPageComponentActions::List { page_id } => {
                        commands::status_pages::components_list(&cfg, &page_id).await?;
                    }
                    StatusPageComponentActions::Get { page_id, component_id } => {
                        commands::status_pages::components_get(&cfg, &page_id, &component_id).await?;
                    }
                },
                StatusPageActions::Degradations { action } => match action {
                    StatusPageDegradationActions::List => {
                        commands::status_pages::degradations_list(&cfg).await?;
                    }
                    StatusPageDegradationActions::Get { page_id, degradation_id } => {
                        commands::status_pages::degradations_get(&cfg, &page_id, &degradation_id).await?;
                    }
                },
            }
        }
        // --- Integrations ---
        Commands::Integrations { action } => {
            cfg.validate_auth()?;
            match action {
                IntegrationActions::Jira { action } => match action {
                    JiraActions::Accounts => commands::integrations::jira_accounts_list(&cfg).await?,
                    JiraActions::Templates { action } => match action {
                        JiraTemplateActions::List => commands::integrations::jira_templates_list(&cfg).await?,
                        JiraTemplateActions::Get { template_id } => {
                            commands::integrations::jira_templates_get(&cfg, &template_id).await?;
                        }
                    },
                },
                IntegrationActions::Servicenow { action } => match action {
                    ServiceNowActions::Instances => {
                        commands::integrations::servicenow_instances_list(&cfg).await?;
                    }
                    ServiceNowActions::Templates { action } => match action {
                        ServiceNowTemplateActions::List => {
                            commands::integrations::servicenow_templates_list(&cfg).await?;
                        }
                        ServiceNowTemplateActions::Get { template_id } => {
                            commands::integrations::servicenow_templates_get(&cfg, &template_id).await?;
                        }
                    },
                },
            }
        }
        // --- Cost ---
        Commands::Cost { action } => {
            cfg.validate_auth()?;
            match action {
                CostActions::Projected => commands::cost::projected(&cfg).await?,
                CostActions::ByOrg { start_month, end_month } => {
                    commands::cost::by_org(&cfg, start_month, end_month).await?;
                }
            }
        }
        // --- Misc ---
        Commands::Misc { action } => {
            cfg.validate_auth()?;
            match action {
                MiscActions::IpRanges => commands::misc::ip_ranges(&cfg).await?,
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
