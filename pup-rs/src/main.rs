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
    /// Output format
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
    Monitors {
        #[command(subcommand)]
        action: MonitorActions,
    },
    /// Search and analyze logs
    Logs {
        #[command(subcommand)]
        action: LogActions,
    },
    /// Manage incidents
    Incidents {
        #[command(subcommand)]
        action: IncidentActions,
    },
    /// Authentication (OAuth2)
    Auth {
        #[command(subcommand)]
        action: AuthActions,
    },
    /// Show version information
    Version,
    /// Validate configuration
    Test,
}

#[derive(Subcommand)]
enum MonitorActions {
    /// List monitors with optional filtering
    List {
        #[arg(long)]
        name: Option<String>,
        #[arg(long)]
        tags: Option<String>,
        #[arg(long, default_value_t = 200)]
        limit: i32,
    },
}

#[derive(Subcommand)]
enum LogActions {
    /// Search logs (forces API key auth)
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
}

#[derive(Subcommand)]
enum IncidentActions {
    /// List incidents (unstable operation)
    List {
        #[arg(long, default_value_t = 50)]
        limit: i64,
    },
}

#[derive(Subcommand)]
enum AuthActions {
    /// Login via OAuth2 (DCR + PKCE)
    Login,
    /// Logout and remove stored tokens
    Logout,
    /// Show current authentication status
    Status,
    /// Print the current access token
    Token,
}

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
        Commands::Monitors { action } => {
            cfg.validate_auth()?;
            match action {
                MonitorActions::List { name, tags, limit } => {
                    commands::monitors::list(&cfg, name, tags, limit).await?;
                }
            }
        }
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
            }
        }
        Commands::Incidents { action } => {
            cfg.validate_auth()?;
            match action {
                IncidentActions::List { limit } => {
                    commands::incidents::list(&cfg, limit).await?;
                }
            }
        }
        Commands::Auth { action } => match action {
            AuthActions::Login => commands::auth::login(&cfg).await?,
            AuthActions::Logout => commands::auth::logout(&cfg).await?,
            AuthActions::Status => commands::auth::status(&cfg)?,
            AuthActions::Token => commands::auth::token(&cfg)?,
        },
        Commands::Version => {
            println!("{}", version::build_info());
        }
        Commands::Test => {
            commands::test::run(&cfg)?;
        }
    }

    Ok(())
}
