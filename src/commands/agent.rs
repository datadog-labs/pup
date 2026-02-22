use anyhow::{bail, Result};

pub fn schema() -> Result<()> {
    bail!("agent schema generation not yet implemented")
}

pub fn guide() -> Result<()> {
    println!("Datadog Agent Management Guide");
    println!("==============================");
    println!();
    println!("The Datadog Agent collects metrics, traces, and logs from your hosts");
    println!("and sends them to Datadog for monitoring and analysis.");
    println!();
    println!("COMMON OPERATIONS:");
    println!("  Install:    See https://docs.datadoghq.com/agent/");
    println!("  Start:      sudo datadog-agent start");
    println!("  Stop:       sudo datadog-agent stop");
    println!("  Restart:    sudo datadog-agent restart");
    println!("  Status:     datadog-agent status");
    println!("  Config:     /etc/datadog-agent/datadog.yaml");
    println!();
    println!("FLEET MANAGEMENT:");
    println!("  Use 'pup fleet' commands to manage agents at scale:");
    println!("  pup fleet agents list       - List all fleet agents");
    println!("  pup fleet deployments list  - List deployments");
    println!("  pup fleet schedules list    - List schedules");
    println!();
    println!("DOCUMENTATION:");
    println!("  https://docs.datadoghq.com/agent/");
    Ok(())
}
