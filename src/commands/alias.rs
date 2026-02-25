use anyhow::{bail, Context, Result};
use std::collections::BTreeMap;
use std::path::PathBuf;

use crate::config;

fn aliases_path() -> Result<PathBuf> {
    let dir = config::config_dir().context("could not determine config directory")?;
    Ok(dir.join("aliases.yaml"))
}

fn load_aliases() -> Result<BTreeMap<String, String>> {
    let path = aliases_path()?;
    match std::fs::read_to_string(&path) {
        Ok(contents) => Ok(serde_yaml::from_str(&contents).unwrap_or_default()),
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(BTreeMap::new()),
        Err(e) => Err(e.into()),
    }
}

fn save_aliases(aliases: &BTreeMap<String, String>) -> Result<()> {
    let path = aliases_path()?;
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }
    let yaml = serde_yaml::to_string(aliases)?;
    std::fs::write(&path, yaml)?;
    Ok(())
}

pub fn list(cfg: &crate::config::Config) -> Result<()> {
    let aliases = load_aliases()?;
    match cfg.output_format {
        crate::config::OutputFormat::Table => {
            if aliases.is_empty() {
                println!("No aliases configured.");
            } else {
                for (name, command) in &aliases {
                    println!("{name} = {command}");
                }
            }
        }
        _ => {
            let items: Vec<serde_json::Value> = aliases
                .iter()
                .map(|(name, command)| serde_json::json!({"name": name, "command": command}))
                .collect();
            crate::formatter::format_and_print(&items, &cfg.output_format, cfg.agent_mode, None)?;
        }
    }
    Ok(())
}

pub fn set(name: String, command: String) -> Result<()> {
    let mut aliases = load_aliases()?;
    aliases.insert(name.clone(), command.clone());
    save_aliases(&aliases)?;
    println!("Alias set: {name} = {command}");
    Ok(())
}

pub fn delete(names: Vec<String>) -> Result<()> {
    let mut aliases = load_aliases()?;
    for name in &names {
        if aliases.remove(name).is_none() {
            bail!("alias not found: {name}");
        }
    }
    save_aliases(&aliases)?;
    println!("Deleted {} alias(es).", names.len());
    Ok(())
}

pub fn import(file: &str) -> Result<()> {
    let contents = std::fs::read_to_string(file)
        .with_context(|| format!("failed to read alias file: {file}"))?;

    // Try YAML first, then JSON
    let imported: BTreeMap<String, String> = if file.ends_with(".json") {
        serde_json::from_str(&contents)
            .with_context(|| format!("failed to parse JSON from {file}"))?
    } else {
        serde_yaml::from_str(&contents)
            .with_context(|| format!("failed to parse YAML from {file}"))?
    };

    if imported.is_empty() {
        bail!("no aliases found in {file}");
    }

    let mut aliases = load_aliases()?;
    let count = imported.len();
    for (name, command) in imported {
        aliases.insert(name, command);
    }
    save_aliases(&aliases)?;
    println!("Imported {count} alias(es) from {file}.");
    Ok(())
}
