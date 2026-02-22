#[allow(dead_code)]
mod api;
mod auth;
mod client;
mod commands;
mod config;
mod formatter;
mod useragent;
mod util;
mod version;

use clap::{CommandFactory, Parser, Subcommand};

#[derive(Parser)]
#[command(name = "pup", version = version::VERSION, about = "Datadog API CLI")]
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
    #[command(
        after_help = "EXAMPLES:\n  # List all monitors\n  pup monitors list\n\n  # Filter monitors by name\n  pup monitors list --name=\"CPU\"\n\n  # Filter monitors by tags\n  pup monitors list --tags=\"env:production,team:backend\"\n\n  # Get detailed information about a specific monitor\n  pup monitors get 12345678\n\n  # Delete a monitor with confirmation prompt\n  pup monitors delete 12345678\n\n  # Delete a monitor without confirmation (automation)\n  pup monitors delete 12345678 --yes"
    )]
    Monitors {
        #[command(subcommand)]
        action: MonitorActions,
    },
    /// Search and analyze logs
    #[command(
        after_help = "EXAMPLES:\n  # Search for error logs in the last hour\n  pup logs search --query=\"status:error\" --from=\"1h\"\n\n  # Search Flex logs specifically\n  pup logs search --query=\"status:error\" --from=\"1h\" --storage=\"flex\"\n\n  # Query logs from a specific service\n  pup logs query --query=\"service:web-app\" --from=\"4h\" --to=\"now\"\n\n  # Query online archives\n  pup logs query --query=\"service:web-app\" --from=\"30d\" --storage=\"online-archives\"\n\n  # Aggregate logs by status\n  pup logs aggregate --query=\"*\" --compute=\"count\" --group-by=\"status\"\n\n  # List log archives\n  pup logs archives list\n\n  # List log-based metrics\n  pup logs metrics list\n\n  # List custom destinations\n  pup logs custom-destinations list\n\n  # List restriction queries\n  pup logs restriction-queries list"
    )]
    Logs {
        #[command(subcommand)]
        action: LogActions,
    },
    /// Manage incidents
    #[command(
        after_help = "EXAMPLES:\n  # List all incidents\n  pup incidents list\n\n  # List incidents with a limit\n  pup incidents list --limit=10\n\n  # Get incident details\n  pup incidents get <incident-id>\n\n  # List incident attachments\n  pup incidents attachments list <incident-id>\n\n  # View global incident settings\n  pup incidents settings get\n\n  # List postmortem templates\n  pup incidents postmortem-templates list"
    )]
    Incidents {
        #[command(subcommand)]
        action: IncidentActions,
    },
    /// Manage dashboards
    #[command(
        after_help = "EXAMPLES:\n  # List all dashboards\n  pup dashboards list\n\n  # Get detailed dashboard configuration\n  pup dashboards get abc-def-123\n\n  # Get dashboard and save to file\n  pup dashboards get abc-def-123 > dashboard.json\n\n  # Delete a dashboard with confirmation\n  pup dashboards delete abc-def-123\n\n  # Delete a dashboard without confirmation (automation)\n  pup dashboards delete abc-def-123 --yes"
    )]
    Dashboards {
        #[command(subcommand)]
        action: DashboardActions,
    },
    /// Query and manage metrics
    #[command(
        after_help = "EXAMPLES:\n  # Query metrics\n  pup metrics query --query=\"avg:system.cpu.user{*}\" --from=\"1h\" --to=\"now\"\n  pup metrics query --query=\"sum:app.requests{env:prod} by {service}\" --from=\"4h\"\n\n  # List metrics\n  pup metrics list\n  pup metrics list --filter=\"system.*\"\n\n  # Get metric metadata\n  pup metrics metadata get system.cpu.user\n\n  # Submit custom metrics\n  pup metrics submit --name=\"custom.metric\" --value=123 --tags=\"env:prod,team:backend\"\n\n  # List metric tags\n  pup metrics tags list system.cpu.user"
    )]
    Metrics {
        #[command(subcommand)]
        action: MetricActions,
    },
    /// Manage Service Level Objectives
    #[command(
        after_help = "EXAMPLES:\n  # List all SLOs\n  pup slos list\n\n  # Get detailed SLO information\n  pup slos get abc-123-def\n\n  # Get SLO history and status\n  pup slos get abc-123-def | jq '.data'\n\n  # Delete an SLO with confirmation\n  pup slos delete abc-123-def\n\n  # Delete an SLO without confirmation (automation)\n  pup slos delete abc-123-def --yes"
    )]
    Slos {
        #[command(subcommand)]
        action: SloActions,
    },
    /// Manage synthetic monitoring
    #[command(
        after_help = "EXAMPLES:\n  # List all synthetic tests\n  pup synthetics tests list\n\n  # Search tests by creator or team\n  pup synthetics tests search --text='creator:\"Jane Doe\"'\n  pup synthetics tests search --text=\"team:my-team\"\n\n  # Get test details\n  pup synthetics tests get test-id\n\n  # List available locations\n  pup synthetics locations list"
    )]
    Synthetics {
        #[command(subcommand)]
        action: SyntheticsActions,
    },
    /// Manage Datadog events
    #[command(
        after_help = "EXAMPLES:\n  # List recent events\n  pup events list\n\n  # Search for deployment events\n  pup events search --query=\"tags:deployment\"\n\n  # Get specific event\n  pup events get 1234567890"
    )]
    Events {
        #[command(subcommand)]
        action: EventActions,
    },
    /// Manage monitor downtimes
    #[command(
        after_help = "EXAMPLES:\n  # List all active downtimes\n  pup downtime list\n\n  # Get downtime details\n  pup downtime get abc-123-def\n\n  # Cancel a downtime\n  pup downtime cancel abc-123-def"
    )]
    Downtime {
        #[command(subcommand)]
        action: DowntimeActions,
    },
    /// Manage host tags
    #[command(
        after_help = "EXAMPLES:\n  # List all host tags\n  pup tags list\n\n  # Get tags for a host\n  pup tags get my-host\n\n  # Add tags to a host\n  pup tags add my-host env:prod team:backend"
    )]
    Tags {
        #[command(subcommand)]
        action: TagActions,
    },
    /// Manage users and access
    #[command(
        after_help = "EXAMPLES:\n  # List all users\n  pup users list\n\n  # Get user details\n  pup users get user-id\n\n  # List roles\n  pup users roles list"
    )]
    Users {
        #[command(subcommand)]
        action: UserActions,
    },
    /// Manage infrastructure monitoring
    #[command(
        after_help = "EXAMPLES:\n  # List all hosts\n  pup infrastructure hosts list\n\n  # Search for hosts by tag\n  pup infrastructure hosts list --filter=\"env:production\"\n\n  # Get host details\n  pup infrastructure hosts get my-host"
    )]
    Infrastructure {
        #[command(subcommand)]
        action: InfraActions,
    },
    /// Query audit logs
    #[command(
        name = "audit-logs",
        after_help = "EXAMPLES:\n  # List recent audit logs\n  pup audit-logs list\n\n  # Search for specific user actions\n  pup audit-logs search --query=\"@usr.name:admin@example.com\"\n\n  # Search for failed actions\n  pup audit-logs search --query=\"@evt.outcome:error\""
    )]
    AuditLogs {
        #[command(subcommand)]
        action: AuditLogActions,
    },
    /// Manage security monitoring
    #[command(
        after_help = "EXAMPLES:\n  # List security monitoring rules\n  pup security rules list\n\n  # Get rule details\n  pup security rules get rule-id\n\n  # List security signals\n  pup security signals list"
    )]
    Security {
        #[command(subcommand)]
        action: SecurityActions,
    },
    /// Manage organization settings
    #[command(
        after_help = "EXAMPLES:\n  # Get organization details\n  pup organizations get\n\n  # List child organizations\n  pup organizations list"
    )]
    Organizations {
        #[command(subcommand)]
        action: OrgActions,
    },
    /// Manage cloud integrations
    #[command(
        after_help = "EXAMPLES:\n  # List AWS integrations\n  pup cloud aws list\n\n  # List GCP integrations\n  pup cloud gcp list\n\n  # List Azure integrations\n  pup cloud azure list"
    )]
    Cloud {
        #[command(subcommand)]
        action: CloudActions,
    },
    /// Manage case management cases and projects
    #[command(
        after_help = "EXAMPLES:\n  # Search cases\n  pup cases search --query=\"bug\"\n\n  # Get case details\n  pup cases get case-123\n\n  # Create a new case\n  pup cases create --title=\"Bug report\" --type-id=\"type-uuid\" --priority=P2\n\n  # List projects\n  pup cases projects list"
    )]
    Cases {
        #[command(subcommand)]
        action: CaseActions,
    },
    /// Manage service catalog
    #[command(
        name = "service-catalog",
        after_help = "EXAMPLES:\n  # List all services\n  pup service-catalog list\n\n  # Get service details\n  pup service-catalog get service-name"
    )]
    ServiceCatalog {
        #[command(subcommand)]
        action: ServiceCatalogActions,
    },
    /// Manage API keys
    #[command(
        name = "api-keys",
        after_help = "EXAMPLES:\n  # List all API keys\n  pup api-keys list\n\n  # Get API key details\n  pup api-keys get key-id\n\n  # Create new API key\n  pup api-keys create --name=\"Production Key\"\n\n  # Delete an API key (with confirmation prompt)\n  pup api-keys delete key-id"
    )]
    ApiKeys {
        #[command(subcommand)]
        action: ApiKeyActions,
    },
    /// Manage application keys
    #[command(
        name = "app-keys",
        after_help = "EXAMPLES:\n  # List your application keys\n  pup app-keys list\n\n  # List all application keys in the org (requires API keys)\n  pup app-keys list --all\n\n  # Get application key details\n  pup app-keys get <app-key-id>\n\n  # Create a new application key\n  pup app-keys create --name=\"My Key\"\n\n  # Create a scoped application key\n  pup app-keys create --name=\"Read Only\" --scopes=\"dashboards_read,metrics_read\"\n\n  # Update an application key name\n  pup app-keys update <app-key-id> --name=\"New Name\"\n\n  # Delete an application key\n  pup app-keys delete <app-key-id>"
    )]
    AppKeys {
        #[command(subcommand)]
        action: AppKeyActions,
    },
    /// Query usage and billing information
    #[command(
        after_help = "EXAMPLES:\n  # Get usage summary for the last 30 days\n  pup usage summary\n\n  # Get usage summary for a specific period\n  pup usage summary --start=\"60d\" --end=\"now\"\n\n  # Get hourly usage for the last day\n  pup usage hourly\n\n  # Get hourly usage for a specific period\n  pup usage hourly --start=\"2d\" --end=\"1d\""
    )]
    Usage {
        #[command(subcommand)]
        action: UsageActions,
    },
    /// Manage notebooks
    #[command(
        after_help = "EXAMPLES:\n  # List all notebooks\n  pup notebooks list\n\n  # Get notebook details\n  pup notebooks get notebook-id\n\n  # Create a notebook from file\n  pup notebooks create --body @notebook.json\n\n  # Create from stdin\n  cat notebook.json | pup notebooks create --body -\n\n  # Update a notebook\n  pup notebooks update 12345 --body @updated.json\n\n  # Delete a notebook\n  pup notebooks delete 12345"
    )]
    Notebooks {
        #[command(subcommand)]
        action: NotebookActions,
    },
    /// Manage Real User Monitoring (RUM)
    #[command(
        after_help = "EXAMPLES:\n  # List all RUM applications\n  pup rum apps list\n\n  # Get RUM application details\n  pup rum apps get --app-id=\"abc-123-def\"\n\n  # Create a new browser RUM application\n  pup rum apps create --name=\"my-web-app\" --type=\"browser\"\n\n  # List RUM custom metrics\n  pup rum metrics list\n\n  # List retention filters\n  pup rum retention-filters list\n\n  # Query session replay data\n  pup rum sessions list --from=\"1h\""
    )]
    Rum {
        #[command(subcommand)]
        action: RumActions,
    },
    /// Manage CI/CD visibility
    #[command(
        after_help = "EXAMPLES:\n  # List recent pipelines\n  pup cicd pipelines list\n\n  # Get pipeline details\n  pup cicd pipelines get --pipeline-id=\"abc-123\"\n\n  # Search for failed pipelines\n  pup cicd events search --query=\"@ci.status:error\" --from=\"1h\"\n\n  # Aggregate by status\n  pup cicd events aggregate --query=\"*\" --compute=\"count\" --group-by=\"@ci.status\"\n\n  # List recent test events\n  pup cicd tests list --from=\"1h\"\n\n  # Search flaky tests\n  pup cicd flaky-tests search --query=\"flaky_test_state:active\""
    )]
    Cicd {
        #[command(subcommand)]
        action: CicdActions,
    },
    /// Manage teams and on-call operations
    #[command(
        name = "on-call",
        after_help = "EXAMPLES:\n  # List all teams\n  pup on-call teams list\n\n  # Create a new team\n  pup on-call teams create --name=\"SRE Team\" --handle=\"sre-team\"\n\n  # Add a member to a team\n  pup on-call teams memberships add <team-id> --user-id=<uuid> --role=member\n\n  # List team members\n  pup on-call teams memberships list <team-id>"
    )]
    OnCall {
        #[command(subcommand)]
        action: OnCallActions,
    },
    /// Manage Fleet Automation
    #[command(
        after_help = "EXAMPLES:\n  # List fleet agents\n  pup fleet agents list\n\n  # Get agent details\n  pup fleet agents get <agent-key>\n\n  # List deployments\n  pup fleet deployments list\n\n  # Deploy a configuration change\n  pup fleet deployments configure --file=config.json\n\n  # List schedules\n  pup fleet schedules list"
    )]
    Fleet {
        #[command(subcommand)]
        action: FleetActions,
    },
    /// Manage data governance
    #[command(
        name = "data-governance",
        after_help = "EXAMPLES:\n  # List scanning rules\n  pup data-governance scanner rules list\n\n  # Get rule details\n  pup data-governance scanner rules get rule-id"
    )]
    DataGovernance {
        #[command(subcommand)]
        action: DataGovActions,
    },
    /// Manage error tracking
    #[command(
        name = "error-tracking",
        after_help = "EXAMPLES:\n  # Search error issues\n  pup error-tracking issues search\n\n  # Get issue details\n  pup error-tracking issues get issue-id"
    )]
    ErrorTracking {
        #[command(subcommand)]
        action: ErrorTrackingActions,
    },
    /// Query code coverage data
    #[command(
        name = "code-coverage",
        after_help = "EXAMPLES:\n  # Get branch coverage summary\n  pup code-coverage branch-summary --repo=\"my-org/my-repo\" --branch=\"main\"\n\n  # Get commit coverage summary\n  pup code-coverage commit-summary --repo=\"my-org/my-repo\" --commit=\"abc123def\""
    )]
    CodeCoverage {
        #[command(subcommand)]
        action: CodeCoverageActions,
    },
    /// Manage High Availability Multi-Region (HAMR)
    #[command(
        after_help = "EXAMPLES:\n  # Get HAMR connection status\n  pup hamr connections get\n\n  # Create a HAMR connection\n  pup hamr connections create --file=connection.json"
    )]
    Hamr {
        #[command(subcommand)]
        action: HamrActions,
    },
    /// Manage status pages
    #[command(
        name = "status-pages",
        after_help = "EXAMPLES:\n  # List all status pages\n  pup status-pages pages list\n\n  # Get status page details\n  pup status-pages pages get <page-id>\n\n  # Create a status page\n  pup status-pages pages create --file=page.json\n\n  # List status page components\n  pup status-pages components list <page-id>\n\n  # List degradations\n  pup status-pages degradations list <page-id>\n\n  # View third-party outage signals\n  pup status-pages third-party list"
    )]
    StatusPages {
        #[command(subcommand)]
        action: StatusPageActions,
    },
    /// Manage third-party integrations
    #[command(
        after_help = "EXAMPLES:\n  # List Slack integrations\n  pup integrations slack list\n\n  # List PagerDuty integrations\n  pup integrations pagerduty list\n\n  # List webhooks\n  pup integrations webhooks list"
    )]
    Integrations {
        #[command(subcommand)]
        action: IntegrationActions,
    },
    /// Manage cost and billing data
    #[command(
        after_help = "EXAMPLES:\n  # Get projected costs for current month\n  pup cost projected\n\n  # Get cost attribution by team tag\n  pup cost attribution --start-month=2024-01 --fields=team\n\n  # Get actual costs for a specific month\n  pup cost by-org --start-month=2024-01"
    )]
    Cost {
        #[command(subcommand)]
        action: CostActions,
    },
    /// Miscellaneous API operations
    #[command(
        after_help = "EXAMPLES:\n  # Get Datadog IP ranges\n  pup misc ip-ranges\n\n  # Check API status\n  pup misc status"
    )]
    Misc {
        #[command(subcommand)]
        action: MiscActions,
    },
    /// Manage APM services and entities
    #[command(
        after_help = "EXAMPLES:\n  # List services with stats\n  pup apm services stats --start $(date -d '1 hour ago' +%s) --end $(date +%s)\n\n  # Query entities with filtering\n  pup apm entities list --start $(date -d '1 hour ago' +%s) --end $(date +%s) --env prod\n\n  # View service dependencies\n  pup apm dependencies list --env prod --start $(date -d '1 hour ago' +%s) --end $(date +%s)"
    )]
    Apm {
        #[command(subcommand)]
        action: ApmActions,
    },
    /// Manage Bits AI investigations
    #[command(
        after_help = "EXAMPLES:\n  # Trigger investigation from a monitor alert\n  pup investigations trigger --type=monitor_alert --monitor-id=123456 --event-id=\"evt-abc\" --event-ts=1706918956000\n\n  # Get investigation details\n  pup investigations get <investigation-id>\n\n  # List investigations\n  pup investigations list --page-limit=20"
    )]
    Investigations {
        #[command(subcommand)]
        action: InvestigationActions,
    },
    /// Manage network monitoring
    #[command(
        after_help = "EXAMPLES:\n  # List network devices/monitors\n  pup network list\n\n  # List network flows\n  pup network flows list\n\n  # List network devices\n  pup network devices list"
    )]
    Network {
        #[command(subcommand)]
        action: NetworkActions,
    },
    /// Manage observability pipelines
    #[command(
        name = "obs-pipelines",
        after_help = "EXAMPLES:\n  # List observability pipelines\n  pup obs-pipelines list\n\n  # Get pipeline details\n  pup obs-pipelines get <pipeline-id>"
    )]
    ObsPipelines {
        #[command(subcommand)]
        action: ObsPipelinesActions,
    },
    /// Manage service scorecards
    #[command(
        after_help = "EXAMPLES:\n  # List scorecards\n  pup scorecards list\n\n  # Get scorecard details\n  pup scorecards get <scorecard-id>"
    )]
    Scorecards {
        #[command(subcommand)]
        action: ScorecardsActions,
    },
    /// Query APM traces
    #[command(after_help = "EXAMPLES:\n  # List recent traces\n  pup traces list")]
    Traces {
        #[command(subcommand)]
        action: TracesActions,
    },
    /// Agent tooling: schema, guide, and diagnostics for AI coding assistants
    #[command(
        name = "agent",
        after_help = "EXAMPLES:\n  # Output command schema as JSON (for AI coding assistants)\n  pup agent schema\n\n  # Output the comprehensive steering guide\n  pup agent guide"
    )]
    Agent {
        #[command(subcommand)]
        action: AgentActions,
    },
    /// Create shortcuts for pup commands
    #[command(
        after_help = "EXAMPLES:\n  # List your aliases\n  pup alias list\n\n  # Create a shortcut for a frequently used command\n  pup alias set prod-monitors \"monitors list --tags=env:production\"\n\n  # Delete an alias\n  pup alias delete prod-monitors\n\n  # Import aliases from a YAML file\n  pup alias import --file=aliases.yaml"
    )]
    Alias {
        #[command(subcommand)]
        action: AliasActions,
    },
    /// Send product analytics events
    #[command(
        name = "product-analytics",
        after_help = "EXAMPLES:\n  # Send a basic event\n  pup product-analytics events send --app-id=my-app --event=button_clicked\n\n  # Send event with properties and user context\n  pup product-analytics events send --app-id=my-app --event=purchase_completed --properties='{\"amount\":99.99,\"currency\":\"USD\"}' --user-id=user-123"
    )]
    ProductAnalytics {
        #[command(subcommand)]
        action: ProductAnalyticsActions,
    },
    /// Manage static analysis
    #[command(
        name = "static-analysis",
        after_help = "EXAMPLES:\n  # List custom rulesets\n  pup static-analysis custom-rulesets list\n\n  # Get ruleset details\n  pup static-analysis custom-rulesets get abc-123"
    )]
    StaticAnalysis {
        #[command(subcommand)]
        action: StaticAnalysisActions,
    },
    /// OAuth2 authentication commands
    #[command(
        after_help = "EXAMPLES:\n  # Login with OAuth2\n  pup auth login\n\n  # Check authentication status\n  pup auth status\n\n  # Refresh access token\n  pup auth refresh\n\n  # Logout and clear credentials\n  pup auth logout\n\n  # Login to different Datadog site\n  DD_SITE=datadoghq.eu pup auth login"
    )]
    Auth {
        #[command(subcommand)]
        action: AuthActions,
    },
    /// Generate shell completions
    #[command(
        after_help = "EXAMPLES:\n  # Generate bash completions\n  pup completions bash > /etc/bash_completion.d/pup\n\n  # Generate zsh completions\n  pup completions zsh > ~/.zfunc/_pup\n\n  # Generate fish completions\n  pup completions fish > ~/.config/fish/completions/pup.fish"
    )]
    Completions {
        /// Shell to generate completions for
        shell: clap_complete::Shell,
    },
    /// Print version information
    #[command(after_help = "EXAMPLES:\n  # Print version\n  pup version")]
    Version,
    /// Test connection and credentials
    #[command(after_help = "EXAMPLES:\n  # Test connection and credentials\n  pup test")]
    Test,
}

// ---- Monitors ----
#[derive(Subcommand)]
enum MonitorActions {
    /// List monitors (limited results)
    List {
        #[arg(long)]
        name: Option<String>,
        #[arg(long)]
        tags: Option<String>,
        #[arg(long, default_value_t = 200)]
        limit: i32,
    },
    /// Get monitor details
    Get { monitor_id: i64 },
    /// Create a monitor from JSON file
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update a monitor from JSON file
    Update {
        monitor_id: i64,
        #[arg(long)]
        file: String,
    },
    /// Search monitors
    Search {
        #[arg(long)]
        query: Option<String>,
    },
    /// Delete a monitor
    Delete { monitor_id: i64 },
}

// ---- Logs ----
#[derive(Subcommand)]
enum LogActions {
    /// Search logs (v1 API)
    Search {
        #[arg(long)]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 50)]
        limit: i32,
    },
    /// List logs (v2 API)
    List {
        #[arg(long, default_value = "*")]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 50)]
        limit: i32,
    },
    /// Query logs (v2 API)
    Query {
        #[arg(long)]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 50)]
        limit: i32,
    },
    /// Aggregate logs (v2 API)
    Aggregate {
        #[arg(long, default_value = "*")]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
    },
    /// Manage log archives
    Archives {
        #[command(subcommand)]
        action: LogArchiveActions,
    },
    /// Manage custom log destinations
    #[command(name = "custom-destinations")]
    CustomDestinations {
        #[command(subcommand)]
        action: LogCustomDestinationActions,
    },
    /// Manage log-based metrics
    Metrics {
        #[command(subcommand)]
        action: LogMetricActions,
    },
    /// Manage log restriction queries
    #[command(name = "restriction-queries")]
    RestrictionQueries {
        #[command(subcommand)]
        action: LogRestrictionQueryActions,
    },
}

#[derive(Subcommand)]
enum LogRestrictionQueryActions {
    /// List restriction queries
    List,
    /// Get restriction query details
    Get { query_id: String },
}

#[derive(Subcommand)]
enum LogArchiveActions {
    /// List all log archives
    List,
    /// Get log archive details
    Get { archive_id: String },
    /// Delete a log archive
    Delete { archive_id: String },
}

#[derive(Subcommand)]
enum LogCustomDestinationActions {
    /// List custom log destinations
    List,
    /// Get custom destination details
    Get { destination_id: String },
}

#[derive(Subcommand)]
enum LogMetricActions {
    /// List log-based metrics
    List,
    /// Get log-based metric details
    Get { metric_id: String },
    /// Delete a log-based metric
    Delete { metric_id: String },
}

// ---- Incidents ----
#[derive(Subcommand)]
enum IncidentActions {
    /// List all incidents
    List {
        #[arg(long, default_value_t = 50)]
        limit: i64,
    },
    /// Get incident details
    Get { incident_id: String },
    /// Manage incident attachments
    Attachments {
        #[command(subcommand)]
        action: IncidentAttachmentActions,
    },
    /// Manage global incident settings
    Settings {
        #[command(subcommand)]
        action: IncidentSettingsActions,
    },
    /// Manage global incident handles
    Handles {
        #[command(subcommand)]
        action: IncidentHandleActions,
    },
    /// Manage incident postmortem templates
    #[command(name = "postmortem-templates")]
    PostmortemTemplates {
        #[command(subcommand)]
        action: IncidentPostmortemActions,
    },
}

#[derive(Subcommand)]
enum IncidentAttachmentActions {
    /// List incident attachments
    List { incident_id: String },
    /// Delete an incident attachment
    Delete {
        incident_id: String,
        attachment_id: String,
    },
}

#[derive(Subcommand)]
enum IncidentSettingsActions {
    /// Get global incident settings
    Get,
    /// Update global incident settings
    Update {
        #[arg(long)]
        file: String,
    },
}

#[derive(Subcommand)]
enum IncidentHandleActions {
    /// List global incident handles
    List,
    /// Create global incident handle
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update global incident handle
    Update {
        #[arg(long)]
        file: String,
    },
    /// Delete global incident handle
    Delete { handle_id: String },
}

#[derive(Subcommand)]
enum IncidentPostmortemActions {
    /// List postmortem templates
    List,
    /// Get postmortem template
    Get { template_id: String },
    /// Create postmortem template
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update postmortem template
    Update {
        template_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete postmortem template
    Delete { template_id: String },
}

// ---- Dashboards ----
#[derive(Subcommand)]
enum DashboardActions {
    /// List all dashboards
    List,
    /// Get dashboard details
    Get { id: String },
    /// Create a dashboard from JSON file
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update a dashboard from JSON file
    Update {
        id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete a dashboard
    Delete { id: String },
}

// ---- Metrics ----
#[derive(Subcommand)]
enum MetricActions {
    /// List all available metrics
    List {
        #[arg(long)]
        filter: Option<String>,
        #[arg(long, default_value = "1h")]
        from: String,
    },
    /// Search metrics (v1 API)
    Search {
        #[arg(long)]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
    },
    /// Query time-series metrics data (v2 API)
    Query {
        #[arg(long)]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
    },
    /// Submit custom metrics to Datadog
    Submit {
        #[arg(long)]
        file: String,
    },
    /// Manage metric metadata
    Metadata {
        #[command(subcommand)]
        action: MetricMetadataActions,
    },
    /// Manage metric tags
    Tags {
        #[command(subcommand)]
        action: MetricTagActions,
    },
}

#[derive(Subcommand)]
enum MetricTagActions {
    /// List tags for a metric
    List { metric_name: String },
}

#[derive(Subcommand)]
enum MetricMetadataActions {
    /// Get metric metadata
    Get { metric_name: String },
    /// Update metric metadata
    Update {
        metric_name: String,
        #[arg(long)]
        file: String,
    },
}

// ---- SLOs ----
#[derive(Subcommand)]
enum SloActions {
    /// List all SLOs
    List,
    /// Get SLO details
    Get { id: String },
    /// Create an SLO from JSON file
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update an SLO from JSON file
    Update {
        id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete an SLO
    Delete { id: String },
    /// Get SLO status
    Status {
        id: String,
        #[arg(long)]
        from_ts: i64,
        #[arg(long)]
        to_ts: i64,
    },
}

// ---- Synthetics ----
#[derive(Subcommand)]
enum SyntheticsActions {
    /// Manage synthetic tests
    Tests {
        #[command(subcommand)]
        action: SyntheticsTestActions,
    },
    /// Manage test locations
    Locations {
        #[command(subcommand)]
        action: SyntheticsLocationActions,
    },
    /// Manage synthetic test suites
    Suites {
        #[command(subcommand)]
        action: SyntheticsSuiteActions,
    },
}

#[derive(Subcommand)]
enum SyntheticsTestActions {
    /// List synthetic tests
    List,
    /// Get test details
    Get { public_id: String },
    /// Search synthetic tests
    Search {
        #[arg(long)]
        text: Option<String>,
        #[arg(long, default_value_t = 50)]
        count: i64,
        #[arg(long, default_value_t = 0)]
        start: i64,
    },
}

#[derive(Subcommand)]
enum SyntheticsLocationActions {
    /// List available locations
    List,
}

#[derive(Subcommand)]
enum SyntheticsSuiteActions {
    /// Search synthetic suites
    List {
        #[arg(long)]
        query: Option<String>,
    },
    /// Get suite details
    Get { suite_id: String },
    /// Create a synthetic suite
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update a synthetic suite
    Update {
        suite_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete synthetic suites
    Delete {
        /// Suite IDs to delete
        suite_ids: Vec<String>,
    },
}

// ---- Events ----
#[derive(Subcommand)]
enum EventActions {
    /// List recent events
    List {
        #[arg(long, default_value_t = 0)]
        start: i64,
        #[arg(long, default_value_t = 0)]
        end: i64,
        #[arg(long)]
        tags: Option<String>,
    },
    /// Search events
    Search {
        #[arg(long)]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 100)]
        limit: i32,
    },
    /// Get event details
    Get { event_id: i64 },
}

// ---- Downtime ----
#[derive(Subcommand)]
enum DowntimeActions {
    /// List all downtimes
    List,
    /// Get downtime details
    Get { id: String },
    /// Create a downtime from JSON file
    Create {
        #[arg(long)]
        file: String,
    },
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
    /// Update host tags
    Update { hostname: String, tags: Vec<String> },
    /// Delete all tags from a host
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
    Roles {
        #[command(subcommand)]
        action: UserRoleActions,
    },
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
    Hosts {
        #[command(subcommand)]
        action: InfraHostActions,
    },
}

#[derive(Subcommand)]
enum InfraHostActions {
    /// List hosts
    List {
        #[arg(long)]
        filter: Option<String>,
        #[arg(long, default_value = "status")]
        sort: String,
        #[arg(long, default_value_t = 100)]
        count: i64,
    },
    /// Get host details
    Get { hostname: String },
}

// ---- Audit Logs ----
#[derive(Subcommand)]
enum AuditLogActions {
    /// List recent audit logs
    List {
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 100)]
        limit: i32,
    },
    /// Search audit logs
    Search {
        #[arg(long)]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 100)]
        limit: i32,
    },
}

// ---- Security ----
#[derive(Subcommand)]
enum SecurityActions {
    /// Manage security rules
    Rules {
        #[command(subcommand)]
        action: SecurityRuleActions,
    },
    /// Manage security signals
    Signals {
        #[command(subcommand)]
        action: SecuritySignalActions,
    },
    /// Manage security findings
    Findings {
        #[command(subcommand)]
        action: SecurityFindingActions,
    },
    /// Manage security content packs
    #[command(name = "content-packs")]
    ContentPacks {
        #[command(subcommand)]
        action: SecurityContentPackActions,
    },
    /// Manage entity risk scores
    #[command(name = "risk-scores")]
    RiskScores {
        #[command(subcommand)]
        action: SecurityRiskScoreActions,
    },
}

#[derive(Subcommand)]
enum SecurityRuleActions {
    /// List security rules
    List,
    /// Get rule details
    Get { rule_id: String },
    /// Bulk export security monitoring rules
    #[command(name = "bulk-export")]
    BulkExport {
        /// Rule IDs to export
        rule_ids: Vec<String>,
    },
}

#[derive(Subcommand)]
enum SecuritySignalActions {
    /// Search security signals
    Search {
        #[arg(long)]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 100)]
        limit: i32,
    },
}

#[derive(Subcommand)]
enum SecurityFindingActions {
    /// Search security findings
    Search {
        #[arg(long)]
        query: Option<String>,
        #[arg(long, default_value_t = 100)]
        limit: i64,
    },
}

#[derive(Subcommand)]
enum SecurityContentPackActions {
    /// List content pack states
    List,
    /// Activate a content pack
    Activate { pack_id: String },
    /// Deactivate a content pack
    Deactivate { pack_id: String },
}

#[derive(Subcommand)]
enum SecurityRiskScoreActions {
    /// List entity risk scores
    List {
        #[arg(long)]
        query: Option<String>,
    },
}

// ---- Organizations ----
#[derive(Subcommand)]
enum OrgActions {
    /// List organizations
    List,
    /// Get organization details
    Get,
}

// ---- Cloud ----
#[derive(Subcommand)]
enum CloudActions {
    /// Manage AWS integrations
    Aws {
        #[command(subcommand)]
        action: CloudAwsActions,
    },
    /// Manage GCP integrations
    Gcp {
        #[command(subcommand)]
        action: CloudGcpActions,
    },
    /// Manage Azure integrations
    Azure {
        #[command(subcommand)]
        action: CloudAzureActions,
    },
    /// Manage OCI integrations
    Oci {
        #[command(subcommand)]
        action: CloudOciActions,
    },
}

#[derive(Subcommand)]
enum CloudAwsActions {
    /// List AWS integrations
    List,
}

#[derive(Subcommand)]
enum CloudGcpActions {
    /// List GCP integrations
    List,
}

#[derive(Subcommand)]
enum CloudAzureActions {
    /// List Azure integrations
    List,
}

#[derive(Subcommand)]
enum CloudOciActions {
    /// Manage OCI tenancy configurations
    Tenancies {
        #[command(subcommand)]
        action: CloudOciTenancyActions,
    },
    /// Manage OCI products
    Products {
        #[command(subcommand)]
        action: CloudOciProductActions,
    },
}

#[derive(Subcommand)]
enum CloudOciProductActions {
    /// List OCI tenancy products
    List {
        /// Product keys to query
        product_keys: String,
    },
}

#[derive(Subcommand)]
enum CloudOciTenancyActions {
    /// List OCI tenancy configurations
    List,
    /// Get OCI tenancy configuration
    Get { tenancy_id: String },
    /// Create OCI tenancy configuration
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update OCI tenancy configuration
    Update {
        tenancy_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete OCI tenancy configuration
    Delete { tenancy_id: String },
}

// ---- Cases ----
#[derive(Subcommand)]
enum CaseActions {
    /// Search cases
    Search {
        #[arg(long)]
        query: Option<String>,
        #[arg(long, default_value_t = 50)]
        page_size: i64,
    },
    /// Get case details
    Get { case_id: String },
    /// Create a new case
    Create {
        #[arg(long)]
        file: String,
    },
    /// Archive a case
    Archive { case_id: String },
    /// Unarchive a case
    Unarchive { case_id: String },
    /// Assign a case to a user
    Assign {
        case_id: String,
        #[arg(long)]
        user_id: String,
    },
    /// Update case priority
    #[command(name = "update-priority")]
    UpdatePriority {
        case_id: String,
        #[arg(long)]
        priority: String,
    },
    /// Update case status
    #[command(name = "update-status")]
    UpdateStatus {
        case_id: String,
        #[arg(long)]
        status: String,
    },
    /// Manage case projects
    Projects {
        #[command(subcommand)]
        action: CaseProjectActions,
    },
    /// Move a case to a different project
    Move {
        case_id: String,
        #[arg(long)]
        project_id: String,
    },
    /// Update case title
    #[command(name = "update-title")]
    UpdateTitle {
        case_id: String,
        #[arg(long)]
        title: String,
    },
    /// Manage Jira integrations for cases
    Jira {
        #[command(subcommand)]
        action: CaseJiraActions,
    },
    /// Manage ServiceNow integrations for cases
    Servicenow {
        #[command(subcommand)]
        action: CaseServicenowActions,
    },
}

#[derive(Subcommand)]
enum CaseProjectActions {
    /// List all projects
    List,
    /// Get project details
    Get { project_id: String },
    /// Create a new project
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        key: String,
    },
    /// Delete a project
    Delete { project_id: String },
    /// Update a project
    Update {
        project_id: String,
        #[arg(long)]
        file: String,
    },
    /// Manage project notification rules
    #[command(name = "notification-rules")]
    NotificationRules {
        #[command(subcommand)]
        action: CaseNotificationRuleActions,
    },
}

#[derive(Subcommand)]
enum CaseJiraActions {
    /// Create a Jira issue for a case
    #[command(name = "create-issue")]
    CreateIssue {
        case_id: String,
        #[arg(long)]
        file: String,
    },
    /// Link a Jira issue to a case
    Link {
        case_id: String,
        #[arg(long)]
        file: String,
    },
    /// Unlink a Jira issue from a case
    Unlink { case_id: String },
}

#[derive(Subcommand)]
enum CaseServicenowActions {
    /// Create a ServiceNow ticket for a case
    #[command(name = "create-ticket")]
    CreateTicket {
        case_id: String,
        #[arg(long)]
        file: String,
    },
}

#[derive(Subcommand)]
enum CaseNotificationRuleActions {
    /// List notification rules for a project
    List { project_id: String },
    /// Create a notification rule
    Create {
        project_id: String,
        #[arg(long)]
        file: String,
    },
    /// Update a notification rule
    Update {
        project_id: String,
        rule_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete a notification rule
    Delete { project_id: String, rule_id: String },
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
    /// Create new API key
    Create {
        #[arg(long)]
        name: String,
    },
    /// Delete an API key (DESTRUCTIVE)
    Delete { key_id: String },
}

// ---- App Keys ----
#[derive(Subcommand)]
enum AppKeyActions {
    /// List application keys
    List {
        /// List all org keys (requires API keys, not OAuth)
        #[arg(long)]
        all: bool,
        /// Filter by key name
        #[arg(long)]
        filter: Option<String>,
        /// Sort field (name, -name, created_at, -created_at)
        #[arg(long)]
        sort: Option<String>,
        /// Number of results per page
        #[arg(long, default_value = "10")]
        page_size: i64,
        /// Page number (0-indexed)
        #[arg(long, default_value = "0")]
        page_number: i64,
    },
    /// Get application key details
    Get { key_id: String },
    /// Create a new application key
    Create {
        /// Application key name
        #[arg(long)]
        name: String,
        /// Comma-separated authorization scopes
        #[arg(long)]
        scopes: Option<String>,
    },
    /// Update an application key
    Update {
        key_id: String,
        /// New name for the application key
        #[arg(long)]
        name: Option<String>,
        /// Comma-separated authorization scopes
        #[arg(long)]
        scopes: Option<String>,
    },
    /// Delete an application key (DESTRUCTIVE)
    Delete { key_id: String },
}

// ---- Usage ----
#[derive(Subcommand)]
enum UsageActions {
    /// Get usage summary
    Summary {
        #[arg(long, default_value = "30d")]
        start: String,
        #[arg(long)]
        end: Option<String>,
    },
    /// Get hourly usage
    Hourly {
        #[arg(long, default_value = "1d")]
        start: String,
        #[arg(long)]
        end: Option<String>,
    },
}

// ---- Notebooks ----
#[derive(Subcommand)]
enum NotebookActions {
    /// List notebooks
    List,
    /// Get notebook details
    Get { notebook_id: i64 },
    /// Create a new notebook
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update a notebook
    Update {
        notebook_id: i64,
        #[arg(long)]
        file: String,
    },
    /// Delete a notebook
    Delete { notebook_id: i64 },
}

// ---- RUM ----
#[derive(Subcommand)]
enum RumActions {
    /// Manage RUM applications
    Apps {
        #[command(subcommand)]
        action: RumAppActions,
    },
    /// List RUM events
    Events {
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 100)]
        limit: i32,
    },
    /// Query RUM session replay data
    Sessions {
        #[command(subcommand)]
        action: RumSessionActions,
    },
    /// Manage RUM custom metrics
    Metrics {
        #[command(subcommand)]
        action: RumMetricActions,
    },
    /// Manage RUM retention filters
    #[command(name = "retention-filters")]
    RetentionFilters {
        #[command(subcommand)]
        action: RumRetentionFilterActions,
    },
    /// Manage session replay playlists
    Playlists {
        #[command(subcommand)]
        action: RumPlaylistActions,
    },
    /// Query RUM interaction heatmaps
    Heatmaps {
        #[command(subcommand)]
        action: RumHeatmapActions,
    },
}

#[derive(Subcommand)]
enum RumAppActions {
    /// List all RUM applications
    List,
    /// Get RUM application details
    Get { app_id: String },
    /// Create a new RUM application
    Create {
        #[arg(long)]
        name: String,
        #[arg(long, name = "type")]
        app_type: Option<String>,
    },
    /// Update a RUM application
    Update {
        app_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete a RUM application
    Delete { app_id: String },
}

#[derive(Subcommand)]
enum RumSessionActions {
    /// Search RUM sessions
    Search {
        #[arg(long)]
        query: Option<String>,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 100)]
        limit: i32,
    },
    /// List RUM sessions
    List {
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 100)]
        limit: i32,
    },
}

#[derive(Subcommand)]
enum RumMetricActions {
    /// List all RUM custom metrics
    List,
    /// Get RUM custom metric details
    Get { metric_id: String },
    /// Create a RUM custom metric
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update a RUM custom metric
    Update {
        metric_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete a RUM custom metric
    Delete { metric_id: String },
}

#[derive(Subcommand)]
enum RumRetentionFilterActions {
    /// List all retention filters
    List { app_id: String },
    /// Get retention filter details
    Get { app_id: String, filter_id: String },
    /// Create a retention filter
    Create {
        app_id: String,
        #[arg(long)]
        file: String,
    },
    /// Update a retention filter
    Update {
        app_id: String,
        filter_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete a retention filter
    Delete { app_id: String, filter_id: String },
}

#[derive(Subcommand)]
enum RumPlaylistActions {
    /// List session replay playlists
    List,
    /// Get playlist details
    Get { playlist_id: i32 },
}

#[derive(Subcommand)]
enum RumHeatmapActions {
    /// Query heatmap data
    Query {
        #[arg(long)]
        view_name: String,
    },
}

// ---- CI/CD ----
#[derive(Subcommand)]
enum CicdActions {
    /// Manage CI pipelines
    Pipelines {
        #[command(subcommand)]
        action: CicdPipelineActions,
    },
    /// Query CI test events
    Tests {
        #[command(subcommand)]
        action: CicdTestActions,
    },
    /// Query CI/CD events
    Events {
        #[command(subcommand)]
        action: CicdEventActions,
    },
    /// Manage DORA metrics
    Dora {
        #[command(subcommand)]
        action: CicdDoraActions,
    },
    /// Manage flaky tests
    #[command(name = "flaky-tests")]
    FlakyTests {
        #[command(subcommand)]
        action: CicdFlakyTestActions,
    },
}

#[derive(Subcommand)]
enum CicdPipelineActions {
    /// List CI pipelines
    List {
        #[arg(long)]
        query: Option<String>,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 50)]
        limit: i32,
    },
    /// Get pipeline details
    Get { pipeline_id: String },
}

#[derive(Subcommand)]
enum CicdTestActions {
    /// List CI test events
    List {
        #[arg(long)]
        query: Option<String>,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 50)]
        limit: i32,
    },
    /// Search CI test events
    Search {
        #[arg(long)]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 50)]
        limit: i32,
    },
    /// Aggregate CI test events
    Aggregate {
        #[arg(long)]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
    },
}

#[derive(Subcommand)]
enum CicdEventActions {
    /// Search CI/CD events
    Search {
        #[arg(long)]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
        #[arg(long, default_value_t = 50)]
        limit: i32,
    },
    /// Aggregate CI/CD events
    Aggregate {
        #[arg(long)]
        query: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
    },
}

#[derive(Subcommand)]
enum CicdDoraActions {
    /// Patch a DORA deployment
    #[command(name = "patch-deployment")]
    PatchDeployment {
        deployment_id: String,
        #[arg(long)]
        file: String,
    },
}

#[derive(Subcommand)]
enum CicdFlakyTestActions {
    /// Search flaky tests
    Search {
        #[arg(long)]
        query: Option<String>,
    },
    /// Update flaky tests
    Update {
        #[arg(long)]
        file: String,
    },
}

// ---- On-Call ----
#[derive(Subcommand)]
enum OnCallActions {
    /// Manage teams
    Teams {
        #[command(subcommand)]
        action: OnCallTeamActions,
    },
}

#[derive(Subcommand)]
enum OnCallTeamActions {
    /// List all teams
    List,
    /// Get team details
    Get { team_id: String },
    /// Create a new team
    Create {
        #[arg(long)]
        name: String,
        #[arg(long)]
        handle: String,
    },
    /// Update team details
    Update {
        team_id: String,
        #[arg(long)]
        name: String,
        #[arg(long)]
        handle: String,
    },
    /// Delete a team
    Delete { team_id: String },
    /// Manage team memberships
    Memberships {
        #[command(subcommand)]
        action: OnCallMembershipActions,
    },
}

#[derive(Subcommand)]
enum OnCallMembershipActions {
    /// List team members
    List {
        team_id: String,
        #[arg(long, default_value_t = 100)]
        page_size: i64,
    },
    /// Add a member to team
    Add {
        team_id: String,
        #[arg(long)]
        user_id: String,
        #[arg(long)]
        role: Option<String>,
    },
    /// Update member role
    Update {
        team_id: String,
        user_id: String,
        #[arg(long)]
        role: String,
    },
    /// Remove member from team
    Remove { team_id: String, user_id: String },
}

// ---- Fleet ----
#[derive(Subcommand)]
enum FleetActions {
    /// Manage fleet agents
    Agents {
        #[command(subcommand)]
        action: FleetAgentActions,
    },
    /// Manage fleet deployments
    Deployments {
        #[command(subcommand)]
        action: FleetDeploymentActions,
    },
    /// Manage fleet schedules
    Schedules {
        #[command(subcommand)]
        action: FleetScheduleActions,
    },
}

#[derive(Subcommand)]
enum FleetAgentActions {
    /// List fleet agents
    List {
        #[arg(long)]
        page_size: Option<i64>,
    },
    /// Get fleet agent details
    Get { agent_key: String },
    /// List available agent versions
    Versions,
}

#[derive(Subcommand)]
enum FleetDeploymentActions {
    /// List fleet deployments
    List {
        #[arg(long)]
        page_size: Option<i64>,
    },
    /// Get fleet deployment details
    Get { deployment_id: String },
    /// Cancel a fleet deployment
    Cancel { deployment_id: String },
    /// Create a configuration deployment
    Configure {
        #[arg(long)]
        file: String,
    },
    /// Create an upgrade deployment
    Upgrade {
        #[arg(long)]
        file: String,
    },
}

#[derive(Subcommand)]
enum FleetScheduleActions {
    /// List fleet schedules
    List,
    /// Get fleet schedule details
    Get { schedule_id: String },
    /// Create a fleet schedule
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update a fleet schedule
    Update {
        schedule_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete a fleet schedule
    Delete { schedule_id: String },
    /// Trigger a fleet schedule
    Trigger { schedule_id: String },
}

// ---- Data Governance ----
#[derive(Subcommand)]
enum DataGovActions {
    /// Manage sensitive data scanner
    Scanner {
        #[command(subcommand)]
        action: DataGovScannerActions,
    },
}

#[derive(Subcommand)]
enum DataGovScannerActions {
    /// Manage scanning rules
    Rules {
        #[command(subcommand)]
        action: DataGovScannerRuleActions,
    },
}

#[derive(Subcommand)]
enum DataGovScannerRuleActions {
    /// List scanning rules
    List,
}

// ---- Error Tracking ----
#[derive(Subcommand)]
enum ErrorTrackingActions {
    /// Manage error issues
    Issues {
        #[command(subcommand)]
        action: ErrorTrackingIssueActions,
    },
}

#[derive(Subcommand)]
enum ErrorTrackingIssueActions {
    /// Search error issues
    Search {
        #[arg(long)]
        query: Option<String>,
        #[arg(long, default_value_t = 50)]
        limit: i32,
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
        #[arg(long)]
        repo: String,
        #[arg(long)]
        branch: String,
    },
    /// Get commit coverage summary
    #[command(name = "commit-summary")]
    CommitSummary {
        #[arg(long)]
        repo: String,
        #[arg(long)]
        commit: String,
    },
}

// ---- HAMR ----
#[derive(Subcommand)]
enum HamrActions {
    /// Manage HAMR organization connections
    Connections {
        #[command(subcommand)]
        action: HamrConnectionActions,
    },
}

#[derive(Subcommand)]
enum HamrConnectionActions {
    /// Get HAMR organization connection
    Get,
    /// Create HAMR organization connection
    Create {
        #[arg(long)]
        file: String,
    },
}

// ---- Status Pages ----
#[derive(Subcommand)]
enum StatusPageActions {
    /// Manage status pages
    Pages {
        #[command(subcommand)]
        action: StatusPagePageActions,
    },
    /// Manage status page components
    Components {
        #[command(subcommand)]
        action: StatusPageComponentActions,
    },
    /// Manage status page degradations
    Degradations {
        #[command(subcommand)]
        action: StatusPageDegradationActions,
    },
    /// View third-party service outage signals
    #[command(name = "third-party")]
    ThirdParty {
        #[command(subcommand)]
        action: StatusPageThirdPartyActions,
    },
}

#[derive(Subcommand)]
enum StatusPagePageActions {
    /// List all status pages
    List,
    /// Get status page details
    Get { page_id: String },
    /// Create a status page
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update a status page
    Update {
        page_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete a status page
    Delete { page_id: String },
}

#[derive(Subcommand)]
enum StatusPageComponentActions {
    /// List components for a page
    List { page_id: String },
    /// Get component details
    Get {
        page_id: String,
        component_id: String,
    },
    /// Create a component
    Create {
        page_id: String,
        #[arg(long)]
        file: String,
    },
    /// Update a component
    Update {
        page_id: String,
        component_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete a component
    Delete {
        page_id: String,
        component_id: String,
    },
}

#[derive(Subcommand)]
enum StatusPageDegradationActions {
    /// List degradations
    List,
    /// Get degradation details
    Get {
        page_id: String,
        degradation_id: String,
    },
    /// Create a degradation
    Create {
        page_id: String,
        #[arg(long)]
        file: String,
    },
    /// Update a degradation
    Update {
        page_id: String,
        degradation_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete a degradation
    Delete {
        page_id: String,
        degradation_id: String,
    },
}

#[derive(Subcommand)]
enum StatusPageThirdPartyActions {
    /// List third-party status pages
    List,
}

// ---- Integrations ----
#[derive(Subcommand)]
enum IntegrationActions {
    /// Manage Jira integration
    Jira {
        #[command(subcommand)]
        action: JiraActions,
    },
    /// Manage ServiceNow integration
    Servicenow {
        #[command(subcommand)]
        action: ServiceNowActions,
    },
    /// Manage Slack integration
    Slack {
        #[command(subcommand)]
        action: SlackActions,
    },
    /// Manage PagerDuty integration
    Pagerduty {
        #[command(subcommand)]
        action: PagerdutyActions,
    },
    /// Manage webhooks
    Webhooks {
        #[command(subcommand)]
        action: WebhooksActions,
    },
}

#[derive(Subcommand)]
enum JiraActions {
    /// Manage Jira accounts
    Accounts {
        #[command(subcommand)]
        action: JiraAccountActions,
    },
    /// Manage Jira issue templates
    Templates {
        #[command(subcommand)]
        action: JiraTemplateActions,
    },
}

#[derive(Subcommand)]
enum JiraAccountActions {
    /// List Jira accounts
    List,
    /// Delete a Jira account
    Delete { account_id: String },
}

#[derive(Subcommand)]
enum JiraTemplateActions {
    /// List Jira issue templates
    List,
    /// Get Jira issue template
    Get { template_id: String },
    /// Create Jira issue template
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update Jira issue template
    Update {
        template_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete Jira issue template
    Delete { template_id: String },
}

#[derive(Subcommand)]
enum ServiceNowActions {
    /// Manage ServiceNow instances
    Instances {
        #[command(subcommand)]
        action: ServiceNowInstanceActions,
    },
    /// Manage ServiceNow templates
    Templates {
        #[command(subcommand)]
        action: ServiceNowTemplateActions,
    },
    /// Manage ServiceNow users
    Users {
        #[command(subcommand)]
        action: ServiceNowUserActions,
    },
    /// Manage ServiceNow assignment groups
    #[command(name = "assignment-groups")]
    AssignmentGroups {
        #[command(subcommand)]
        action: ServiceNowAssignmentGroupActions,
    },
    /// Manage ServiceNow business services
    #[command(name = "business-services")]
    BusinessServices {
        #[command(subcommand)]
        action: ServiceNowBusinessServiceActions,
    },
}

#[derive(Subcommand)]
enum ServiceNowInstanceActions {
    /// List ServiceNow instances
    List,
}

#[derive(Subcommand)]
enum ServiceNowUserActions {
    /// List ServiceNow users
    List { instance_name: String },
}

#[derive(Subcommand)]
enum ServiceNowAssignmentGroupActions {
    /// List ServiceNow assignment groups
    List { instance_name: String },
}

#[derive(Subcommand)]
enum ServiceNowBusinessServiceActions {
    /// List ServiceNow business services
    List { instance_name: String },
}

#[derive(Subcommand)]
enum ServiceNowTemplateActions {
    /// List ServiceNow templates
    List,
    /// Get ServiceNow template
    Get { template_id: String },
    /// Create ServiceNow template
    Create {
        #[arg(long)]
        file: String,
    },
    /// Update ServiceNow template
    Update {
        template_id: String,
        #[arg(long)]
        file: String,
    },
    /// Delete ServiceNow template
    Delete { template_id: String },
}

#[derive(Subcommand)]
enum SlackActions {
    /// List Slack channels
    List,
}

#[derive(Subcommand)]
enum PagerdutyActions {
    /// List PagerDuty services
    List,
}

#[derive(Subcommand)]
enum WebhooksActions {
    /// List webhooks
    List,
}

// ---- Cost ----
#[derive(Subcommand)]
enum CostActions {
    /// Get projected end-of-month costs
    Projected,
    /// Get costs by organization
    #[command(name = "by-org")]
    ByOrg {
        #[arg(long)]
        start_month: String,
        #[arg(long)]
        end_month: Option<String>,
    },
    /// Get cost attribution by tags
    Attribution {
        #[arg(long)]
        start: String,
        #[arg(long)]
        fields: Option<String>,
    },
}

// ---- Misc ----
#[derive(Subcommand)]
enum MiscActions {
    /// Get Datadog IP ranges
    #[command(name = "ip-ranges")]
    IpRanges,
    /// Check API status
    Status,
}

// ---- APM ----
#[derive(Subcommand)]
enum ApmActions {
    /// Manage APM services
    Services {
        #[command(subcommand)]
        action: ApmServiceActions,
    },
    /// Manage APM entities
    Entities {
        #[command(subcommand)]
        action: ApmEntityActions,
    },
    /// Manage service dependencies
    Dependencies {
        #[command(subcommand)]
        action: ApmDependencyActions,
    },
    /// View service flow map
    #[command(name = "flow-map")]
    FlowMap {
        #[arg(long)]
        query: String,
        #[arg(long, default_value_t = 100)]
        limit: i64,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
    },
}

#[derive(Subcommand)]
enum ApmServiceActions {
    /// List APM services
    List {
        #[arg(long)]
        env: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
    },
    /// List services with performance statistics
    Stats {
        #[arg(long)]
        env: String,
        #[arg(long)]
        from: String,
        #[arg(long)]
        to: String,
    },
    /// List operations for a service
    Operations {
        #[arg(long)]
        service: String,
        #[arg(long)]
        env: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
    },
    /// List resources (endpoints) for a service operation
    Resources {
        #[arg(long)]
        service: String,
        #[arg(long)]
        operation: String,
        #[arg(long)]
        env: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
    },
}

#[derive(Subcommand)]
enum ApmEntityActions {
    /// Query APM entities
    List {
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
    },
}

#[derive(Subcommand)]
enum ApmDependencyActions {
    /// List service dependencies
    List {
        #[arg(long)]
        env: String,
        #[arg(long, default_value = "1h")]
        from: String,
        #[arg(long, default_value = "now")]
        to: String,
    },
}

// ---- Investigations ----
#[derive(Subcommand)]
enum InvestigationActions {
    /// List investigations
    List {
        #[arg(long, default_value_t = 50)]
        page_limit: i64,
        #[arg(long, default_value_t = 0)]
        page_offset: i64,
    },
    /// Get investigation details
    Get { investigation_id: String },
    /// Trigger a new investigation
    Trigger {
        #[arg(long)]
        file: String,
    },
}

// ---- Network (placeholder) ----
#[derive(Subcommand)]
enum NetworkActions {
    /// List network devices/monitors
    List,
    /// Query network flows
    Flows {
        #[command(subcommand)]
        action: NetworkFlowActions,
    },
    /// List network devices
    Devices {
        #[command(subcommand)]
        action: NetworkDeviceActions,
    },
}

#[derive(Subcommand)]
enum NetworkFlowActions {
    /// List network flows
    List,
}

#[derive(Subcommand)]
enum NetworkDeviceActions {
    /// List network devices
    List,
}

// ---- Obs Pipelines (placeholder) ----
#[derive(Subcommand)]
enum ObsPipelinesActions {
    /// List observability pipelines
    List,
    /// Get pipeline details
    Get { pipeline_id: String },
}

// ---- Scorecards (placeholder) ----
#[derive(Subcommand)]
enum ScorecardsActions {
    /// List scorecards
    List,
    /// Get scorecard details
    Get { scorecard_id: String },
}

// ---- Traces (placeholder) ----
#[derive(Subcommand)]
enum TracesActions {
    /// List traces
    List,
}

// ---- Agent (placeholder) ----
#[derive(Subcommand)]
enum AgentActions {
    /// Output command schema as JSON
    Schema,
    /// Output the comprehensive steering guide
    Guide,
}

// ---- Alias ----
#[derive(Subcommand)]
enum AliasActions {
    /// List your aliases
    List,
    /// Create a shortcut for a pup command
    Set { name: String, command: String },
    /// Delete set aliases
    Delete { names: Vec<String> },
    /// Import aliases from a YAML file
    Import {
        #[arg(long)]
        file: String,
    },
}

// ---- Product Analytics ----
#[derive(Subcommand)]
enum ProductAnalyticsActions {
    /// Send product analytics events
    Events {
        #[command(subcommand)]
        action: ProductAnalyticsEventActions,
    },
}

#[derive(Subcommand)]
enum ProductAnalyticsEventActions {
    /// Send a product analytics event
    Send {
        #[arg(long)]
        file: String,
    },
}

// ---- Static Analysis ----
#[derive(Subcommand)]
enum StaticAnalysisActions {
    /// AST analysis
    Ast {
        #[command(subcommand)]
        action: StaticAnalysisAstActions,
    },
    /// Custom security rulesets
    #[command(name = "custom-rulesets")]
    CustomRulesets {
        #[command(subcommand)]
        action: StaticAnalysisCustomRulesetActions,
    },
    /// Software Composition Analysis
    Sca {
        #[command(subcommand)]
        action: StaticAnalysisScaActions,
    },
    /// Code coverage analysis
    Coverage {
        #[command(subcommand)]
        action: StaticAnalysisCoverageActions,
    },
}

#[derive(Subcommand)]
enum StaticAnalysisAstActions {
    /// List AST analyses
    List,
    /// Get AST analysis details
    Get { id: String },
}

#[derive(Subcommand)]
enum StaticAnalysisCustomRulesetActions {
    /// List custom rulesets
    List,
    /// Get custom ruleset details
    Get { id: String },
}

#[derive(Subcommand)]
enum StaticAnalysisScaActions {
    /// List SCA results
    List,
    /// Get SCA scan details
    Get { id: String },
}

#[derive(Subcommand)]
enum StaticAnalysisCoverageActions {
    /// List coverage analyses
    List,
    /// Get coverage analysis details
    Get { id: String },
}

// ---- Auth ----
#[derive(Subcommand)]
enum AuthActions {
    /// Login via OAuth2
    Login,
    /// Logout and clear tokens
    Logout,
    /// Check authentication status
    Status,
    /// Print access token
    Token,
    /// Refresh access token
    Refresh,
}

// ---- Agent-mode JSON schema for --help ----

fn build_agent_schema(cmd: &clap::Command) -> serde_json::Value {
    let mut root = serde_json::Map::new();
    root.insert("version".into(), serde_json::json!("dev"));
    root.insert(
        "description".into(),
        serde_json::json!(
            "Pup - Datadog API CLI. Provides OAuth2 + API key authentication for querying metrics, logs, monitors, traces, and 30+ other Datadog API domains."
        ),
    );
    let mut auth = serde_json::Map::new();
    auth.insert("oauth".into(), serde_json::json!("pup auth login"));
    auth.insert(
        "api_keys".into(),
        serde_json::json!("Set DD_API_KEY + DD_APP_KEY + DD_SITE environment variables"),
    );
    root.insert("auth".into(), serde_json::Value::Object(auth));

    // Global flags — hardcoded to match Go ordering and descriptions exactly
    root.insert(
        "global_flags".into(),
        serde_json::json!([
            {
                "name": "--agent",
                "type": "bool",
                "default": "false",
                "description": "Enable agent mode (auto-detected for AI coding assistants)"
            },
            {
                "name": "--output",
                "type": "string",
                "default": "json",
                "description": "Output format (json, table, yaml)"
            },
            {
                "name": "--yes",
                "type": "bool",
                "default": "false",
                "description": "Skip confirmation prompts (auto-approve all operations)"
            }
        ]),
    );

    // Operational knowledge sections — critical for AI agent effectiveness
    root.insert("anti_patterns".into(), serde_json::json!([
        "Don't omit --from on time-series queries; you'll get unexpected time ranges or errors",
        "Don't use --limit=1000 as a first step; start with small limits and refine queries",
        "Don't list all monitors/logs without filters in large organizations (>10k monitors)",
        "Don't assume APM durations are in seconds or milliseconds; they are in NANOSECONDS",
        "Don't fetch raw logs to count them; use 'pup logs aggregate --compute=count' instead",
        "Don't use --from=30d unless you specifically need a month of data; it's slow",
        "Don't retry failed requests without checking the error; 401 means re-authenticate, 403 means missing permissions",
        "Don't use 'pup metrics query' without specifying an aggregation (avg, sum, max, min, count)",
        "Don't pipe large JSON responses through multiple jq transforms; use query filters at the API level"
    ]));

    root.insert("best_practices".into(), serde_json::json!([
        "Always specify --from to set a time range; most commands default to 1h but be explicit",
        "Start with narrow time ranges (1h) then widen if needed; large ranges are slow and expensive",
        "Filter by service first when investigating issues: --query='service:<name>'",
        "Use --limit to control result size; default varies by command (50-200)",
        "For monitors, use --tags to filter rather than listing all and parsing locally",
        "APM durations are in NANOSECONDS: 1 second = 1000000000, 5ms = 5000000",
        "Use 'pup logs aggregate' for counts and distributions instead of fetching all logs and counting locally",
        "Prefer JSON output (default) for structured parsing; use --output=table only for human display",
        "Chain narrow queries: first aggregate to find patterns, then search for specific examples",
        "Use 'pup monitors search' for full-text search, 'pup monitors list' for tag/name filtering"
    ]));

    root.insert("query_syntax".into(), serde_json::json!({
        "apm": "service:<name> resource_name:<path> @duration:>5000000000 (nanoseconds!) status:error operation_name:<op>. Duration is always in nanoseconds",
        "events": "sources:nagios,pagerduty status:error priority:normal tags:env:prod",
        "logs": "status:error, service:web-app, @attr:val, host:i-*, \"exact phrase\", AND/OR/NOT operators, -status:info (negation), wildcards with *",
        "metrics": "<aggregation>:<metric_name>{<filter>} by {<group>}. Example: avg:system.cpu.user{env:prod} by {host}. Aggregations: avg, sum, min, max, count",
        "monitors": "Use --name for substring search, --tags for tag filtering (comma-separated). Search via --query for full-text search",
        "rum": "@type:error @session.type:user @view.url_path:/checkout @action.type:click service:<app-name>",
        "security": "@workflow.rule.type:log_detection source:cloudtrail @network.client.ip:10.0.0.0/8 status:critical",
        "traces": "service:<name> resource_name:<path> @duration:>5s (shorthand) env:production"
    }));

    root.insert("time_formats".into(), serde_json::json!({
        "relative": ["5s", "30m", "1h", "4h", "1d", "7d", "1w", "30d", "5min", "2hours", "3days"],
        "absolute": ["Unix timestamp in milliseconds", "RFC3339 (2024-01-01T00:00:00Z)"],
        "examples": [
            "--from=1h (1 hour ago)",
            "--from=30m --to=now",
            "--from=7d --to=1d (7 days ago to 1 day ago)",
            "--from=2024-01-01T00:00:00Z --to=2024-01-02T00:00:00Z",
            "--from=\"5 minutes\""
        ]
    }));

    root.insert("workflows".into(), serde_json::json!([
        {
            "name": "Investigate errors",
            "steps": [
                "pup logs search --query=\"status:error\" --from=1h --limit=20",
                "pup logs aggregate --query=\"status:error\" --from=1h --compute=\"count\" --group-by=\"service\"",
                "pup monitors list --tags=\"env:production\" --limit=50"
            ]
        },
        {
            "name": "Performance investigation",
            "steps": [
                "pup metrics query --query=\"avg:trace.servlet.request.duration{env:prod} by {service}\" --from=1h",
                "pup logs search --query=\"@duration:>5000000000\" --from=1h --limit=20",
                "pup apm services list"
            ]
        },
        {
            "name": "Monitor status check",
            "steps": [
                "pup monitors list --tags=\"env:production\" --limit=500",
                "pup monitors search --query=\"status:Alert\"",
                "pup monitors get <monitor_id>"
            ]
        },
        {
            "name": "Security audit",
            "steps": [
                "pup audit-logs search --query=\"*\" --from=1d --limit=100",
                "pup security rules list",
                "pup security signals list --query=\"status:critical\" --from=1d"
            ]
        },
        {
            "name": "Service health overview",
            "steps": [
                "pup slos list",
                "pup monitors list --tags=\"team:<team_name>\"",
                "pup incidents list --query=\"status:active\""
            ]
        }
    ]));

    // Commands — sorted alphabetically to match Go
    let mut commands: Vec<serde_json::Value> = cmd
        .get_subcommands()
        .filter(|s| s.get_name() != "help")
        .map(|s| build_command_schema(s, ""))
        .collect();
    commands.sort_by(|a, b| {
        let an = a.get("name").and_then(|v| v.as_str()).unwrap_or("");
        let bn = b.get("name").and_then(|v| v.as_str()).unwrap_or("");
        an.cmp(bn)
    });
    root.insert("commands".into(), serde_json::Value::Array(commands));

    serde_json::Value::Object(root)
}

fn build_command_schema(cmd: &clap::Command, parent_path: &str) -> serde_json::Value {
    let mut obj = serde_json::Map::new();
    let name = cmd.get_name().to_string();
    let full_path = if parent_path.is_empty() {
        name.clone()
    } else {
        format!("{parent_path} {name}")
    };

    obj.insert("name".into(), serde_json::json!(name));
    obj.insert("full_path".into(), serde_json::json!(full_path));

    if let Some(about) = cmd.get_about() {
        obj.insert("description".into(), serde_json::json!(about.to_string()));
    }

    // Determine read_only based on command name
    let is_write = name == "delete"
        || name == "create"
        || name == "update"
        || name == "cancel"
        || name == "trigger"
        || name == "set"
        || name == "add"
        || name == "remove"
        || name == "assign"
        || name == "archive"
        || name == "unarchive"
        || name == "activate"
        || name == "deactivate"
        || name.starts_with("update-")
        || name.starts_with("create-")
        || name == "submit"
        || name == "send"
        || name == "import"
        || name == "register"
        || name == "unregister"
        || name.contains("delete")
        || name.contains("patch");
    obj.insert("read_only".into(), serde_json::json!(!is_write));

    // Flags (non-global arguments specific to this command)
    let flags: Vec<serde_json::Value> = cmd
        .get_arguments()
        .filter(|a| {
            let id = a.get_id().as_str();
            id != "help" && id != "version" && !a.is_global_set()
        })
        .map(|a| {
            let mut flag = serde_json::Map::new();
            let flag_name = if let Some(long) = a.get_long() {
                format!("--{long}")
            } else {
                // Positional arg
                a.get_id().to_string()
            };
            flag.insert("name".into(), serde_json::json!(flag_name));
            let type_str = if a.get_action().takes_values() {
                "string"
            } else {
                "bool"
            };
            flag.insert("type".into(), serde_json::json!(type_str));
            if let Some(def) = a.get_default_values().first() {
                flag.insert(
                    "default".into(),
                    serde_json::json!(def.to_str().unwrap_or("").to_string()),
                );
            }
            if let Some(help) = a.get_help() {
                flag.insert("description".into(), serde_json::json!(help.to_string()));
            }
            serde_json::Value::Object(flag)
        })
        .collect();
    if !flags.is_empty() {
        obj.insert("flags".into(), serde_json::Value::Array(flags));
    }

    // Subcommands — sorted alphabetically to match Go
    let mut subs: Vec<serde_json::Value> = cmd
        .get_subcommands()
        .filter(|s| s.get_name() != "help")
        .map(|s| build_command_schema(s, &full_path))
        .collect();
    subs.sort_by(|a, b| {
        let an = a.get("name").and_then(|v| v.as_str()).unwrap_or("");
        let bn = b.get("name").and_then(|v| v.as_str()).unwrap_or("");
        an.cmp(bn)
    });
    if !subs.is_empty() {
        obj.insert("subcommands".into(), serde_json::Value::Array(subs));
    }

    serde_json::Value::Object(obj)
}

// ---- Main ----

#[cfg(not(target_arch = "wasm32"))]
#[tokio::main]
async fn main() -> anyhow::Result<()> {
    main_inner().await
}

#[cfg(target_arch = "wasm32")]
#[tokio::main(flavor = "current_thread")]
async fn main() -> anyhow::Result<()> {
    main_inner().await
}

async fn main_inner() -> anyhow::Result<()> {
    // In agent mode, intercept --help to return a JSON schema instead of plain text.
    let args: Vec<String> = std::env::args().collect();
    let has_help = args.iter().any(|a| a == "--help" || a == "-h");
    if has_help && useragent::is_agent_mode() {
        let cmd = Cli::command();
        let schema = build_agent_schema(&cmd);
        println!("{}", serde_json::to_string_pretty(&schema).unwrap());
        return Ok(());
    }

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
                MonitorActions::Get { monitor_id } => {
                    commands::monitors::get(&cfg, monitor_id).await?;
                }
                MonitorActions::Create { file } => {
                    commands::monitors::create(&cfg, &file).await?;
                }
                MonitorActions::Update { monitor_id, file } => {
                    commands::monitors::update(&cfg, monitor_id, &file).await?;
                }
                MonitorActions::Search { query } => {
                    commands::monitors::search(&cfg, query).await?;
                }
                MonitorActions::Delete { monitor_id } => {
                    commands::monitors::delete(&cfg, monitor_id).await?;
                }
            }
        }
        // --- Logs ---
        Commands::Logs { action } => {
            cfg.validate_auth()?;
            match action {
                LogActions::Search {
                    query,
                    from,
                    to,
                    limit,
                } => {
                    commands::logs::search(&cfg, query, from, to, limit).await?;
                }
                LogActions::List {
                    query,
                    from,
                    to,
                    limit,
                } => {
                    commands::logs::list(&cfg, query, from, to, limit).await?;
                }
                LogActions::Query {
                    query,
                    from,
                    to,
                    limit,
                } => {
                    commands::logs::query(&cfg, query, from, to, limit).await?;
                }
                LogActions::Aggregate { query, from, to } => {
                    commands::logs::aggregate(&cfg, query, from, to).await?;
                }
                LogActions::Archives { action } => match action {
                    LogArchiveActions::List => commands::logs::archives_list(&cfg).await?,
                    LogArchiveActions::Get { archive_id } => {
                        commands::logs::archives_get(&cfg, &archive_id).await?;
                    }
                    LogArchiveActions::Delete { archive_id } => {
                        commands::logs::archives_delete(&cfg, &archive_id).await?;
                    }
                },
                LogActions::CustomDestinations { action } => match action {
                    LogCustomDestinationActions::List => {
                        commands::logs::custom_destinations_list(&cfg).await?;
                    }
                    LogCustomDestinationActions::Get { destination_id } => {
                        commands::logs::custom_destinations_get(&cfg, &destination_id).await?;
                    }
                },
                LogActions::Metrics { action } => match action {
                    LogMetricActions::List => commands::logs::metrics_list(&cfg).await?,
                    LogMetricActions::Get { metric_id } => {
                        commands::logs::metrics_get(&cfg, &metric_id).await?;
                    }
                    LogMetricActions::Delete { metric_id } => {
                        commands::logs::metrics_delete(&cfg, &metric_id).await?;
                    }
                },
                LogActions::RestrictionQueries { action } => match action {
                    LogRestrictionQueryActions::List => {
                        commands::logs::restriction_queries_list(&cfg).await?;
                    }
                    LogRestrictionQueryActions::Get { query_id } => {
                        commands::logs::restriction_queries_get(&cfg, &query_id).await?;
                    }
                },
            }
        }
        // --- Incidents ---
        Commands::Incidents { action } => {
            cfg.validate_auth()?;
            match action {
                IncidentActions::List { limit } => {
                    commands::incidents::list(&cfg, limit).await?;
                }
                IncidentActions::Get { incident_id } => {
                    commands::incidents::get(&cfg, &incident_id).await?;
                }
                IncidentActions::Attachments { action } => match action {
                    IncidentAttachmentActions::List { incident_id } => {
                        commands::incidents::attachments_list(&cfg, &incident_id).await?;
                    }
                    IncidentAttachmentActions::Delete {
                        incident_id,
                        attachment_id,
                    } => {
                        commands::incidents::attachments_delete(&cfg, &incident_id, &attachment_id)
                            .await?;
                    }
                },
                IncidentActions::Settings { action } => match action {
                    IncidentSettingsActions::Get => {
                        commands::incidents::settings_get(&cfg).await?;
                    }
                    IncidentSettingsActions::Update { file } => {
                        commands::incidents::settings_update(&cfg, &file).await?;
                    }
                },
                IncidentActions::Handles { action } => match action {
                    IncidentHandleActions::List => {
                        commands::incidents::handles_list(&cfg).await?;
                    }
                    IncidentHandleActions::Create { file } => {
                        commands::incidents::handles_create(&cfg, &file).await?;
                    }
                    IncidentHandleActions::Update { file } => {
                        commands::incidents::handles_update(&cfg, &file).await?;
                    }
                    IncidentHandleActions::Delete { handle_id } => {
                        commands::incidents::handles_delete(&cfg, &handle_id).await?;
                    }
                },
                IncidentActions::PostmortemTemplates { action } => match action {
                    IncidentPostmortemActions::List => {
                        commands::incidents::postmortem_templates_list(&cfg).await?;
                    }
                    IncidentPostmortemActions::Get { template_id } => {
                        commands::incidents::postmortem_templates_get(&cfg, &template_id).await?;
                    }
                    IncidentPostmortemActions::Create { file } => {
                        commands::incidents::postmortem_templates_create(&cfg, &file).await?;
                    }
                    IncidentPostmortemActions::Update { template_id, file } => {
                        commands::incidents::postmortem_templates_update(&cfg, &template_id, &file)
                            .await?;
                    }
                    IncidentPostmortemActions::Delete { template_id } => {
                        commands::incidents::postmortem_templates_delete(&cfg, &template_id)
                            .await?;
                    }
                },
            }
        }
        // --- Dashboards ---
        Commands::Dashboards { action } => {
            cfg.validate_auth()?;
            match action {
                DashboardActions::List => commands::dashboards::list(&cfg).await?,
                DashboardActions::Get { id } => commands::dashboards::get(&cfg, &id).await?,
                DashboardActions::Create { file } => {
                    commands::dashboards::create(&cfg, &file).await?;
                }
                DashboardActions::Update { id, file } => {
                    commands::dashboards::update(&cfg, &id, &file).await?;
                }
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
                MetricActions::Query { query, from, to } => {
                    commands::metrics::query(&cfg, query, from, to).await?;
                }
                MetricActions::Submit { file } => {
                    commands::metrics::submit(&cfg, &file).await?;
                }
                MetricActions::Metadata { action } => match action {
                    MetricMetadataActions::Get { metric_name } => {
                        commands::metrics::metadata_get(&cfg, &metric_name).await?;
                    }
                    MetricMetadataActions::Update { metric_name, file } => {
                        commands::metrics::metadata_update(&cfg, &metric_name, &file).await?;
                    }
                },
                MetricActions::Tags { action } => match action {
                    MetricTagActions::List { metric_name } => {
                        commands::metrics::tags_list(&cfg, &metric_name).await?;
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
                SloActions::Create { file } => commands::slos::create(&cfg, &file).await?,
                SloActions::Update { id, file } => {
                    commands::slos::update(&cfg, &id, &file).await?;
                }
                SloActions::Delete { id } => commands::slos::delete(&cfg, &id).await?,
                SloActions::Status { id, from_ts, to_ts } => {
                    commands::slos::status(&cfg, &id, from_ts, to_ts).await?;
                }
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
                SyntheticsActions::Suites { action } => match action {
                    SyntheticsSuiteActions::List { query } => {
                        commands::synthetics::suites_list(&cfg, query).await?;
                    }
                    SyntheticsSuiteActions::Get { suite_id } => {
                        commands::synthetics::suites_get(&cfg, &suite_id).await?;
                    }
                    SyntheticsSuiteActions::Create { file } => {
                        commands::synthetics::suites_create(&cfg, &file).await?;
                    }
                    SyntheticsSuiteActions::Update { suite_id, file } => {
                        commands::synthetics::suites_update(&cfg, &suite_id, &file).await?;
                    }
                    SyntheticsSuiteActions::Delete { suite_ids } => {
                        commands::synthetics::suites_delete(&cfg, suite_ids).await?;
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
                EventActions::Search {
                    query,
                    from,
                    to,
                    limit,
                } => {
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
                DowntimeActions::Create { file } => {
                    commands::downtime::create(&cfg, &file).await?;
                }
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
                    InfraHostActions::List {
                        filter,
                        sort,
                        count,
                    } => {
                        commands::infrastructure::hosts_list(&cfg, filter, sort, count).await?;
                    }
                    InfraHostActions::Get { hostname } => {
                        commands::infrastructure::hosts_get(&cfg, &hostname).await?;
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
                AuditLogActions::Search {
                    query,
                    from,
                    to,
                    limit,
                } => {
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
                    SecurityRuleActions::BulkExport { rule_ids } => {
                        commands::security::rules_bulk_export(&cfg, rule_ids).await?;
                    }
                },
                SecurityActions::Signals { action } => match action {
                    SecuritySignalActions::Search {
                        query,
                        from,
                        to,
                        limit,
                    } => {
                        commands::security::signals_search(&cfg, query, from, to, limit).await?;
                    }
                },
                SecurityActions::Findings { action } => match action {
                    SecurityFindingActions::Search { query, limit } => {
                        commands::security::findings_search(&cfg, query, limit).await?;
                    }
                },
                SecurityActions::ContentPacks { action } => match action {
                    SecurityContentPackActions::List => {
                        commands::security::content_packs_list(&cfg).await?;
                    }
                    SecurityContentPackActions::Activate { pack_id } => {
                        commands::security::content_packs_activate(&cfg, &pack_id).await?;
                    }
                    SecurityContentPackActions::Deactivate { pack_id } => {
                        commands::security::content_packs_deactivate(&cfg, &pack_id).await?;
                    }
                },
                SecurityActions::RiskScores { action } => match action {
                    SecurityRiskScoreActions::List { query } => {
                        commands::security::risk_scores_list(&cfg, query).await?;
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
                    CloudAwsActions::List => commands::cloud::aws_list(&cfg).await?,
                },
                CloudActions::Gcp { action } => match action {
                    CloudGcpActions::List => commands::cloud::gcp_list(&cfg).await?,
                },
                CloudActions::Azure { action } => match action {
                    CloudAzureActions::List => commands::cloud::azure_list(&cfg).await?,
                },
                CloudActions::Oci { action } => match action {
                    CloudOciActions::Tenancies { action } => match action {
                        CloudOciTenancyActions::List => {
                            commands::cloud::oci_tenancies_list(&cfg).await?;
                        }
                        CloudOciTenancyActions::Get { tenancy_id } => {
                            commands::cloud::oci_tenancies_get(&cfg, &tenancy_id).await?;
                        }
                        CloudOciTenancyActions::Create { file } => {
                            commands::cloud::oci_tenancies_create(&cfg, &file).await?;
                        }
                        CloudOciTenancyActions::Update { tenancy_id, file } => {
                            commands::cloud::oci_tenancies_update(&cfg, &tenancy_id, &file).await?;
                        }
                        CloudOciTenancyActions::Delete { tenancy_id } => {
                            commands::cloud::oci_tenancies_delete(&cfg, &tenancy_id).await?;
                        }
                    },
                    CloudOciActions::Products { action } => match action {
                        CloudOciProductActions::List { product_keys } => {
                            commands::cloud::oci_products_list(&cfg, &product_keys).await?;
                        }
                    },
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
                CaseActions::Create { file } => {
                    commands::cases::create(&cfg, &file).await?;
                }
                CaseActions::Archive { case_id } => {
                    commands::cases::archive(&cfg, &case_id).await?;
                }
                CaseActions::Unarchive { case_id } => {
                    commands::cases::unarchive(&cfg, &case_id).await?;
                }
                CaseActions::Assign { case_id, user_id } => {
                    commands::cases::assign(&cfg, &case_id, &user_id).await?;
                }
                CaseActions::UpdatePriority { case_id, priority } => {
                    commands::cases::update_priority(&cfg, &case_id, &priority).await?;
                }
                CaseActions::UpdateStatus { case_id, status } => {
                    commands::cases::update_status(&cfg, &case_id, &status).await?;
                }
                CaseActions::Move {
                    case_id,
                    project_id,
                } => {
                    commands::cases::move_to_project(&cfg, &case_id, &project_id).await?;
                }
                CaseActions::UpdateTitle { case_id, title } => {
                    commands::cases::update_title(&cfg, &case_id, &title).await?;
                }
                CaseActions::Projects { action } => match action {
                    CaseProjectActions::List => commands::cases::projects_list(&cfg).await?,
                    CaseProjectActions::Get { project_id } => {
                        commands::cases::projects_get(&cfg, &project_id).await?;
                    }
                    CaseProjectActions::Create { name, key } => {
                        commands::cases::projects_create(&cfg, &name, &key).await?;
                    }
                    CaseProjectActions::Delete { project_id } => {
                        commands::cases::projects_delete(&cfg, &project_id).await?;
                    }
                    CaseProjectActions::Update { project_id, file } => {
                        commands::cases::projects_update(&cfg, &project_id, &file).await?;
                    }
                    CaseProjectActions::NotificationRules { action } => match action {
                        CaseNotificationRuleActions::List { project_id } => {
                            commands::cases::projects_notification_rules_list(&cfg, &project_id)
                                .await?;
                        }
                        CaseNotificationRuleActions::Create { project_id, file } => {
                            commands::cases::projects_notification_rules_create(
                                &cfg,
                                &project_id,
                                &file,
                            )
                            .await?;
                        }
                        CaseNotificationRuleActions::Update {
                            project_id,
                            rule_id,
                            file,
                        } => {
                            commands::cases::projects_notification_rules_update(
                                &cfg,
                                &project_id,
                                &rule_id,
                                &file,
                            )
                            .await?;
                        }
                        CaseNotificationRuleActions::Delete {
                            project_id,
                            rule_id,
                        } => {
                            commands::cases::projects_notification_rules_delete(
                                &cfg,
                                &project_id,
                                &rule_id,
                            )
                            .await?;
                        }
                    },
                },
                CaseActions::Jira { action } => match action {
                    CaseJiraActions::CreateIssue { case_id, file } => {
                        commands::cases::jira_create_issue(&cfg, &case_id, &file).await?;
                    }
                    CaseJiraActions::Link { case_id, file } => {
                        commands::cases::jira_link(&cfg, &case_id, &file).await?;
                    }
                    CaseJiraActions::Unlink { case_id } => {
                        commands::cases::jira_unlink(&cfg, &case_id).await?;
                    }
                },
                CaseActions::Servicenow { action } => match action {
                    CaseServicenowActions::CreateTicket { case_id, file } => {
                        commands::cases::servicenow_create_ticket(&cfg, &case_id, &file).await?;
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
                ApiKeyActions::Create { name } => {
                    commands::api_keys::create(&cfg, &name).await?;
                }
                ApiKeyActions::Delete { key_id } => {
                    commands::api_keys::delete(&cfg, &key_id).await?;
                }
            }
        }
        // --- App Keys ---
        Commands::AppKeys { action } => {
            cfg.validate_auth()?;
            match action {
                AppKeyActions::List {
                    all,
                    filter,
                    sort,
                    page_size,
                    page_number,
                } => {
                    commands::app_keys::list(&cfg, all, &filter, &sort, page_size, page_number)
                        .await?
                }
                AppKeyActions::Get { key_id } => {
                    commands::app_keys::get(&cfg, &key_id).await?
                }
                AppKeyActions::Create { name, scopes } => {
                    commands::app_keys::create(&cfg, &name, &scopes).await?
                }
                AppKeyActions::Update {
                    key_id,
                    name,
                    scopes,
                } => commands::app_keys::update(&cfg, &key_id, &name, &scopes).await?,
                AppKeyActions::Delete { key_id } => {
                    if !cfg.auto_approve {
                        eprint!("Delete application key {key_id}? Type 'yes' to confirm: ");
                        let mut input = String::new();
                        std::io::stdin().read_line(&mut input)?;
                        if input.trim() != "yes" {
                            println!("Operation cancelled.");
                            return Ok(());
                        }
                    }
                    commands::app_keys::delete(&cfg, &key_id).await?
                }
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
                NotebookActions::Create { file } => {
                    commands::notebooks::create(&cfg, &file).await?;
                }
                NotebookActions::Update { notebook_id, file } => {
                    commands::notebooks::update(&cfg, notebook_id, &file).await?;
                }
                NotebookActions::Delete { notebook_id } => {
                    commands::notebooks::delete(&cfg, notebook_id).await?;
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
                    RumAppActions::Create { name, app_type } => {
                        commands::rum::apps_create(&cfg, &name, app_type).await?;
                    }
                    RumAppActions::Update { app_id, file } => {
                        commands::rum::apps_update(&cfg, &app_id, &file).await?;
                    }
                    RumAppActions::Delete { app_id } => {
                        commands::rum::apps_delete(&cfg, &app_id).await?;
                    }
                },
                RumActions::Events { from, to, limit } => {
                    commands::rum::events_list(&cfg, from, to, limit).await?;
                }
                RumActions::Sessions { action } => match action {
                    RumSessionActions::Search {
                        query,
                        from,
                        to,
                        limit,
                    } => {
                        commands::rum::sessions_search(&cfg, query, from, to, limit).await?;
                    }
                    RumSessionActions::List { from, to, limit } => {
                        commands::rum::sessions_list(&cfg, from, to, limit).await?;
                    }
                },
                RumActions::Metrics { action } => match action {
                    RumMetricActions::List => commands::rum::metrics_list(&cfg).await?,
                    RumMetricActions::Get { metric_id } => {
                        commands::rum::metrics_get(&cfg, &metric_id).await?;
                    }
                    RumMetricActions::Create { file } => {
                        commands::rum::metrics_create(&cfg, &file).await?;
                    }
                    RumMetricActions::Update { metric_id, file } => {
                        commands::rum::metrics_update(&cfg, &metric_id, &file).await?;
                    }
                    RumMetricActions::Delete { metric_id } => {
                        commands::rum::metrics_delete(&cfg, &metric_id).await?;
                    }
                },
                RumActions::RetentionFilters { action } => match action {
                    RumRetentionFilterActions::List { app_id } => {
                        commands::rum::retention_filters_list(&cfg, &app_id).await?;
                    }
                    RumRetentionFilterActions::Get { app_id, filter_id } => {
                        commands::rum::retention_filters_get(&cfg, &app_id, &filter_id).await?;
                    }
                    RumRetentionFilterActions::Create { app_id, file } => {
                        commands::rum::retention_filters_create(&cfg, &app_id, &file).await?;
                    }
                    RumRetentionFilterActions::Update {
                        app_id,
                        filter_id,
                        file,
                    } => {
                        commands::rum::retention_filters_update(&cfg, &app_id, &filter_id, &file)
                            .await?;
                    }
                    RumRetentionFilterActions::Delete { app_id, filter_id } => {
                        commands::rum::retention_filters_delete(&cfg, &app_id, &filter_id).await?;
                    }
                },
                RumActions::Playlists { action } => match action {
                    RumPlaylistActions::List => commands::rum::playlists_list(&cfg).await?,
                    RumPlaylistActions::Get { playlist_id } => {
                        commands::rum::playlists_get(&cfg, playlist_id).await?;
                    }
                },
                RumActions::Heatmaps { action } => match action {
                    RumHeatmapActions::Query { view_name } => {
                        commands::rum::heatmaps_query(&cfg, &view_name).await?;
                    }
                },
            }
        }
        // --- CI/CD ---
        Commands::Cicd { action } => {
            cfg.validate_auth()?;
            match action {
                CicdActions::Pipelines { action } => match action {
                    CicdPipelineActions::List {
                        query,
                        from,
                        to,
                        limit,
                    } => {
                        commands::cicd::pipelines_list(&cfg, query, from, to, limit).await?;
                    }
                    CicdPipelineActions::Get { pipeline_id } => {
                        commands::cicd::pipelines_get(&cfg, &pipeline_id).await?;
                    }
                },
                CicdActions::Tests { action } => match action {
                    CicdTestActions::List {
                        query,
                        from,
                        to,
                        limit,
                    } => {
                        commands::cicd::tests_list(&cfg, query, from, to, limit).await?;
                    }
                    CicdTestActions::Search {
                        query,
                        from,
                        to,
                        limit,
                    } => {
                        commands::cicd::tests_search(&cfg, query, from, to, limit).await?;
                    }
                    CicdTestActions::Aggregate { query, from, to } => {
                        commands::cicd::tests_aggregate(&cfg, query, from, to).await?;
                    }
                },
                CicdActions::Events { action } => match action {
                    CicdEventActions::Search {
                        query,
                        from,
                        to,
                        limit,
                    } => {
                        commands::cicd::events_search(&cfg, query, from, to, limit).await?;
                    }
                    CicdEventActions::Aggregate { query, from, to } => {
                        commands::cicd::events_aggregate(&cfg, query, from, to).await?;
                    }
                },
                CicdActions::Dora { action } => match action {
                    CicdDoraActions::PatchDeployment {
                        deployment_id,
                        file,
                    } => {
                        commands::cicd::dora_patch_deployment(&cfg, &deployment_id, &file).await?;
                    }
                },
                CicdActions::FlakyTests { action } => match action {
                    CicdFlakyTestActions::Search { query } => {
                        commands::cicd::flaky_tests_search(&cfg, query).await?;
                    }
                    CicdFlakyTestActions::Update { file } => {
                        commands::cicd::flaky_tests_update(&cfg, &file).await?;
                    }
                },
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
                    OnCallTeamActions::Create { name, handle } => {
                        commands::on_call::teams_create(&cfg, &name, &handle).await?;
                    }
                    OnCallTeamActions::Update {
                        team_id,
                        name,
                        handle,
                    } => {
                        commands::on_call::teams_update(&cfg, &team_id, &name, &handle).await?;
                    }
                    OnCallTeamActions::Delete { team_id } => {
                        commands::on_call::teams_delete(&cfg, &team_id).await?;
                    }
                    OnCallTeamActions::Memberships { action } => match action {
                        OnCallMembershipActions::List { team_id, page_size } => {
                            commands::on_call::memberships_list(&cfg, &team_id, page_size).await?;
                        }
                        OnCallMembershipActions::Add {
                            team_id,
                            user_id,
                            role,
                        } => {
                            commands::on_call::memberships_add(&cfg, &team_id, &user_id, role)
                                .await?;
                        }
                        OnCallMembershipActions::Update {
                            team_id,
                            user_id,
                            role,
                        } => {
                            commands::on_call::memberships_update(&cfg, &team_id, &user_id, &role)
                                .await?;
                        }
                        OnCallMembershipActions::Remove { team_id, user_id } => {
                            commands::on_call::memberships_remove(&cfg, &team_id, &user_id).await?;
                        }
                    },
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
                    FleetDeploymentActions::Cancel { deployment_id } => {
                        commands::fleet::deployments_cancel(&cfg, &deployment_id).await?;
                    }
                    FleetDeploymentActions::Configure { file } => {
                        commands::fleet::deployments_configure(&cfg, &file).await?;
                    }
                    FleetDeploymentActions::Upgrade { file } => {
                        commands::fleet::deployments_upgrade(&cfg, &file).await?;
                    }
                },
                FleetActions::Schedules { action } => match action {
                    FleetScheduleActions::List => commands::fleet::schedules_list(&cfg).await?,
                    FleetScheduleActions::Get { schedule_id } => {
                        commands::fleet::schedules_get(&cfg, &schedule_id).await?;
                    }
                    FleetScheduleActions::Create { file } => {
                        commands::fleet::schedules_create(&cfg, &file).await?;
                    }
                    FleetScheduleActions::Update { schedule_id, file } => {
                        commands::fleet::schedules_update(&cfg, &schedule_id, &file).await?;
                    }
                    FleetScheduleActions::Delete { schedule_id } => {
                        commands::fleet::schedules_delete(&cfg, &schedule_id).await?;
                    }
                    FleetScheduleActions::Trigger { schedule_id } => {
                        commands::fleet::schedules_trigger(&cfg, &schedule_id).await?;
                    }
                },
            }
        }
        // --- Data Governance ---
        Commands::DataGovernance { action } => {
            cfg.validate_auth()?;
            match action {
                DataGovActions::Scanner { action } => match action {
                    DataGovScannerActions::Rules { action } => match action {
                        DataGovScannerRuleActions::List => {
                            commands::data_governance::scanner_rules_list(&cfg).await?;
                        }
                    },
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
                    HamrConnectionActions::Create { file } => {
                        commands::hamr::connections_create(&cfg, &file).await?;
                    }
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
                    StatusPagePageActions::Create { file } => {
                        commands::status_pages::pages_create(&cfg, &file).await?;
                    }
                    StatusPagePageActions::Update { page_id, file } => {
                        commands::status_pages::pages_update(&cfg, &page_id, &file).await?;
                    }
                    StatusPagePageActions::Delete { page_id } => {
                        commands::status_pages::pages_delete(&cfg, &page_id).await?;
                    }
                },
                StatusPageActions::Components { action } => match action {
                    StatusPageComponentActions::List { page_id } => {
                        commands::status_pages::components_list(&cfg, &page_id).await?;
                    }
                    StatusPageComponentActions::Get {
                        page_id,
                        component_id,
                    } => {
                        commands::status_pages::components_get(&cfg, &page_id, &component_id)
                            .await?;
                    }
                    StatusPageComponentActions::Create { page_id, file } => {
                        commands::status_pages::components_create(&cfg, &page_id, &file).await?;
                    }
                    StatusPageComponentActions::Update {
                        page_id,
                        component_id,
                        file,
                    } => {
                        commands::status_pages::components_update(
                            &cfg,
                            &page_id,
                            &component_id,
                            &file,
                        )
                        .await?;
                    }
                    StatusPageComponentActions::Delete {
                        page_id,
                        component_id,
                    } => {
                        commands::status_pages::components_delete(&cfg, &page_id, &component_id)
                            .await?;
                    }
                },
                StatusPageActions::Degradations { action } => match action {
                    StatusPageDegradationActions::List => {
                        commands::status_pages::degradations_list(&cfg).await?;
                    }
                    StatusPageDegradationActions::Get {
                        page_id,
                        degradation_id,
                    } => {
                        commands::status_pages::degradations_get(&cfg, &page_id, &degradation_id)
                            .await?;
                    }
                    StatusPageDegradationActions::Create { page_id, file } => {
                        commands::status_pages::degradations_create(&cfg, &page_id, &file).await?;
                    }
                    StatusPageDegradationActions::Update {
                        page_id,
                        degradation_id,
                        file,
                    } => {
                        commands::status_pages::degradations_update(
                            &cfg,
                            &page_id,
                            &degradation_id,
                            &file,
                        )
                        .await?;
                    }
                    StatusPageDegradationActions::Delete {
                        page_id,
                        degradation_id,
                    } => {
                        commands::status_pages::degradations_delete(
                            &cfg,
                            &page_id,
                            &degradation_id,
                        )
                        .await?;
                    }
                },
                StatusPageActions::ThirdParty { action } => match action {
                    StatusPageThirdPartyActions::List => {
                        commands::status_pages::third_party_list(&cfg).await?;
                    }
                },
            }
        }
        // --- Integrations ---
        Commands::Integrations { action } => {
            cfg.validate_auth()?;
            match action {
                IntegrationActions::Jira { action } => match action {
                    JiraActions::Accounts { action } => match action {
                        JiraAccountActions::List => {
                            commands::integrations::jira_accounts_list(&cfg).await?
                        }
                        JiraAccountActions::Delete { account_id } => {
                            commands::integrations::jira_accounts_delete(&cfg, &account_id).await?;
                        }
                    },
                    JiraActions::Templates { action } => match action {
                        JiraTemplateActions::List => {
                            commands::integrations::jira_templates_list(&cfg).await?
                        }
                        JiraTemplateActions::Get { template_id } => {
                            commands::integrations::jira_templates_get(&cfg, &template_id).await?;
                        }
                        JiraTemplateActions::Create { file } => {
                            commands::integrations::jira_templates_create(&cfg, &file).await?;
                        }
                        JiraTemplateActions::Update { template_id, file } => {
                            commands::integrations::jira_templates_update(
                                &cfg,
                                &template_id,
                                &file,
                            )
                            .await?;
                        }
                        JiraTemplateActions::Delete { template_id } => {
                            commands::integrations::jira_templates_delete(&cfg, &template_id)
                                .await?;
                        }
                    },
                },
                IntegrationActions::Servicenow { action } => match action {
                    ServiceNowActions::Instances { action } => match action {
                        ServiceNowInstanceActions::List => {
                            commands::integrations::servicenow_instances_list(&cfg).await?;
                        }
                    },
                    ServiceNowActions::Templates { action } => match action {
                        ServiceNowTemplateActions::List => {
                            commands::integrations::servicenow_templates_list(&cfg).await?;
                        }
                        ServiceNowTemplateActions::Get { template_id } => {
                            commands::integrations::servicenow_templates_get(&cfg, &template_id)
                                .await?;
                        }
                        ServiceNowTemplateActions::Create { file } => {
                            commands::integrations::servicenow_templates_create(&cfg, &file)
                                .await?;
                        }
                        ServiceNowTemplateActions::Update { template_id, file } => {
                            commands::integrations::servicenow_templates_update(
                                &cfg,
                                &template_id,
                                &file,
                            )
                            .await?;
                        }
                        ServiceNowTemplateActions::Delete { template_id } => {
                            commands::integrations::servicenow_templates_delete(&cfg, &template_id)
                                .await?;
                        }
                    },
                    ServiceNowActions::Users { action } => match action {
                        ServiceNowUserActions::List { instance_name } => {
                            commands::integrations::servicenow_users_list(&cfg, &instance_name)
                                .await?;
                        }
                    },
                    ServiceNowActions::AssignmentGroups { action } => match action {
                        ServiceNowAssignmentGroupActions::List { instance_name } => {
                            commands::integrations::servicenow_assignment_groups_list(
                                &cfg,
                                &instance_name,
                            )
                            .await?;
                        }
                    },
                    ServiceNowActions::BusinessServices { action } => match action {
                        ServiceNowBusinessServiceActions::List { instance_name } => {
                            commands::integrations::servicenow_business_services_list(
                                &cfg,
                                &instance_name,
                            )
                            .await?;
                        }
                    },
                },
                IntegrationActions::Slack { action } => match action {
                    SlackActions::List => commands::integrations::slack_list(&cfg).await?,
                },
                IntegrationActions::Pagerduty { action } => match action {
                    PagerdutyActions::List => {
                        commands::integrations::pagerduty_list(&cfg).await?;
                    }
                },
                IntegrationActions::Webhooks { action } => match action {
                    WebhooksActions::List => commands::integrations::webhooks_list(&cfg).await?,
                },
            }
        }
        // --- Cost ---
        Commands::Cost { action } => {
            cfg.validate_auth()?;
            match action {
                CostActions::Projected => commands::cost::projected(&cfg).await?,
                CostActions::ByOrg {
                    start_month,
                    end_month,
                } => {
                    commands::cost::by_org(&cfg, start_month, end_month).await?;
                }
                CostActions::Attribution { start, fields } => {
                    commands::cost::attribution(&cfg, start, fields).await?;
                }
            }
        }
        // --- Misc ---
        Commands::Misc { action } => {
            cfg.validate_auth()?;
            match action {
                MiscActions::IpRanges => commands::misc::ip_ranges(&cfg).await?,
                MiscActions::Status => commands::misc::status(&cfg).await?,
            }
        }
        // --- APM ---
        Commands::Apm { action } => {
            cfg.validate_auth()?;
            match action {
                ApmActions::Services { action } => match action {
                    ApmServiceActions::List { env, from, to } => {
                        commands::apm::services_list(&cfg, env, from, to).await?;
                    }
                    ApmServiceActions::Stats { env, from, to } => {
                        commands::apm::services_stats(&cfg, env, from, to).await?;
                    }
                    ApmServiceActions::Operations {
                        service,
                        env,
                        from,
                        to,
                    } => {
                        commands::apm::services_operations(&cfg, service, env, from, to).await?;
                    }
                    ApmServiceActions::Resources {
                        service,
                        operation,
                        env,
                        from,
                        to,
                    } => {
                        commands::apm::services_resources(&cfg, service, operation, env, from, to)
                            .await?;
                    }
                },
                ApmActions::Entities { action } => match action {
                    ApmEntityActions::List { from, to } => {
                        commands::apm::entities_list(&cfg, from, to).await?;
                    }
                },
                ApmActions::Dependencies { action } => match action {
                    ApmDependencyActions::List { env, from, to } => {
                        commands::apm::dependencies_list(&cfg, env, from, to).await?;
                    }
                },
                ApmActions::FlowMap {
                    query,
                    limit,
                    from,
                    to,
                } => {
                    commands::apm::flow_map(&cfg, query, limit, from, to).await?;
                }
            }
        }
        // --- Investigations ---
        Commands::Investigations { action } => {
            cfg.validate_auth()?;
            match action {
                InvestigationActions::List {
                    page_limit,
                    page_offset,
                } => {
                    commands::investigations::list(&cfg, page_limit, page_offset).await?;
                }
                InvestigationActions::Get { investigation_id } => {
                    commands::investigations::get(&cfg, &investigation_id).await?;
                }
                InvestigationActions::Trigger { file } => {
                    commands::investigations::trigger(&cfg, &file).await?;
                }
            }
        }
        // --- Network (placeholder) ---
        Commands::Network { action } => match action {
            NetworkActions::List => commands::network::list()?,
            NetworkActions::Flows { action } => match action {
                NetworkFlowActions::List => {
                    cfg.validate_auth()?;
                    commands::network::flows_list(&cfg).await?;
                }
            },
            NetworkActions::Devices { action } => match action {
                NetworkDeviceActions::List => {
                    cfg.validate_auth()?;
                    commands::network::devices_list(&cfg).await?;
                }
            },
        },
        // --- Obs Pipelines (placeholder) ---
        Commands::ObsPipelines { action } => match action {
            ObsPipelinesActions::List => commands::obs_pipelines::list()?,
            ObsPipelinesActions::Get { pipeline_id } => {
                commands::obs_pipelines::get(&pipeline_id)?;
            }
        },
        // --- Scorecards (placeholder) ---
        Commands::Scorecards { action } => match action {
            ScorecardsActions::List => commands::scorecards::list()?,
            ScorecardsActions::Get { scorecard_id } => {
                commands::scorecards::get(&scorecard_id)?;
            }
        },
        // --- Traces (placeholder) ---
        Commands::Traces { action } => match action {
            TracesActions::List => commands::traces::list()?,
        },
        // --- Agent (placeholder) ---
        Commands::Agent { action } => match action {
            AgentActions::Schema => commands::agent::schema()?,
            AgentActions::Guide => commands::agent::guide()?,
        },
        // --- Alias ---
        Commands::Alias { action } => match action {
            AliasActions::List => commands::alias::list()?,
            AliasActions::Set { name, command } => commands::alias::set(name, command)?,
            AliasActions::Delete { names } => commands::alias::delete(names)?,
            AliasActions::Import { file } => commands::alias::import(&file)?,
        },
        // --- Product Analytics ---
        Commands::ProductAnalytics { action } => {
            cfg.validate_auth()?;
            match action {
                ProductAnalyticsActions::Events { action } => match action {
                    ProductAnalyticsEventActions::Send { file } => {
                        commands::product_analytics::events_send(&cfg, &file).await?;
                    }
                },
            }
        }
        // --- Static Analysis ---
        Commands::StaticAnalysis { action } => {
            cfg.validate_auth()?;
            match action {
                StaticAnalysisActions::Ast { action } => match action {
                    StaticAnalysisAstActions::List => {
                        commands::static_analysis::ast_list(&cfg).await?;
                    }
                    StaticAnalysisAstActions::Get { id } => {
                        commands::static_analysis::ast_get(&cfg, &id).await?;
                    }
                },
                StaticAnalysisActions::CustomRulesets { action } => match action {
                    StaticAnalysisCustomRulesetActions::List => {
                        commands::static_analysis::custom_rulesets_list(&cfg).await?;
                    }
                    StaticAnalysisCustomRulesetActions::Get { id } => {
                        commands::static_analysis::custom_rulesets_get(&cfg, &id).await?;
                    }
                },
                StaticAnalysisActions::Sca { action } => match action {
                    StaticAnalysisScaActions::List => {
                        commands::static_analysis::sca_list(&cfg).await?;
                    }
                    StaticAnalysisScaActions::Get { id } => {
                        commands::static_analysis::sca_get(&cfg, &id).await?;
                    }
                },
                StaticAnalysisActions::Coverage { action } => match action {
                    StaticAnalysisCoverageActions::List => {
                        commands::static_analysis::coverage_list(&cfg).await?;
                    }
                    StaticAnalysisCoverageActions::Get { id } => {
                        commands::static_analysis::coverage_get(&cfg, &id).await?;
                    }
                },
            }
        }
        // --- Auth ---
        Commands::Auth { action } => match action {
            AuthActions::Login => commands::auth::login(&cfg).await?,
            AuthActions::Logout => commands::auth::logout(&cfg).await?,
            AuthActions::Status => commands::auth::status(&cfg)?,
            AuthActions::Token => commands::auth::token(&cfg)?,
            AuthActions::Refresh => commands::auth::refresh(&cfg).await?,
        },
        // --- Utility ---
        Commands::Completions { shell } => {
            clap_complete::generate(shell, &mut Cli::command(), "pup", &mut std::io::stdout());
        }
        Commands::Version => println!("{}", version::build_info()),
        Commands::Test => commands::test::run(&cfg)?,
    }

    Ok(())
}
