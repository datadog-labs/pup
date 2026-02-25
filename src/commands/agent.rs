use anyhow::Result;

pub fn guide() -> Result<()> {
    println!("datadog-agent (Datadog-Agent) — Operational Reference");
    println!("======================================================");
    println!();
    println!("NOTE: This guide is about the datadog-agent host daemon that collects");
    println!("metrics, traces, and logs from your infrastructure — NOT an AI agent.");
    println!();
    println!("The datadog-agent binary runs on each monitored host and ships data to");
    println!("Datadog. Manage it at scale with 'pup fleet' commands.");
    println!();
    println!("COMMON datadog-agent OPERATIONS:");
    println!("  Install:    https://docs.datadoghq.com/agent/");
    println!("  Start:      sudo datadog-agent start");
    println!("  Stop:       sudo datadog-agent stop");
    println!("  Restart:    sudo datadog-agent restart");
    println!("  Status:     datadog-agent status");
    println!("  Config:     /etc/datadog-agent/datadog.yaml");
    println!();
    println!("FLEET MANAGEMENT (via pup):");
    println!("  pup fleet agents list        List all datadog-agent instances");
    println!("  pup fleet agents versions    Show available datadog-agent versions");
    println!("  pup fleet deployments list   List agent deployment tasks");
    println!("  pup fleet schedules list     List agent schedule tasks");
    println!();
    println!("DOCUMENTATION:");
    println!("  https://docs.datadoghq.com/agent/");
    Ok(())
}
