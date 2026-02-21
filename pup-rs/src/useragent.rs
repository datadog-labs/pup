use crate::version;

pub struct AgentInfo {
    pub name: String,
    pub detected: bool,
}

struct AgentDetector {
    name: &'static str,
    env_vars: &'static [&'static str],
}

/// Table-driven AI agent detection, checked in priority order.
static AGENT_DETECTORS: &[AgentDetector] = &[
    AgentDetector {
        name: "claude-code",
        env_vars: &["CLAUDECODE", "CLAUDE_CODE"],
    },
    AgentDetector {
        name: "cursor",
        env_vars: &["CURSOR_AGENT"],
    },
    AgentDetector {
        name: "codex",
        env_vars: &["CODEX", "OPENAI_CODEX"],
    },
    AgentDetector {
        name: "opencode",
        env_vars: &["OPENCODE"],
    },
    AgentDetector {
        name: "aider",
        env_vars: &["AIDER"],
    },
    AgentDetector {
        name: "cline",
        env_vars: &["CLINE"],
    },
    AgentDetector {
        name: "windsurf",
        env_vars: &["WINDSURF_AGENT"],
    },
    AgentDetector {
        name: "github-copilot",
        env_vars: &["GITHUB_COPILOT"],
    },
    AgentDetector {
        name: "amazon-q",
        env_vars: &["AMAZON_Q", "AWS_Q_DEVELOPER"],
    },
    AgentDetector {
        name: "gemini-code",
        env_vars: &["GEMINI_CODE_ASSIST"],
    },
    AgentDetector {
        name: "sourcegraph-cody",
        env_vars: &["SRC_CODY"],
    },
];

fn is_env_truthy(key: &str) -> bool {
    match std::env::var(key) {
        Ok(val) => matches!(val.to_lowercase().as_str(), "1" | "true"),
        Err(_) => false,
    }
}

pub fn detect_agent_info() -> AgentInfo {
    for detector in AGENT_DETECTORS {
        for env_var in detector.env_vars {
            if is_env_truthy(env_var) {
                return AgentInfo {
                    name: detector.name.to_string(),
                    detected: true,
                };
            }
        }
    }
    AgentInfo {
        name: String::new(),
        detected: false,
    }
}

pub fn is_agent_mode() -> bool {
    is_env_truthy("FORCE_AGENT_MODE") || detect_agent_info().detected
}

pub fn get() -> String {
    let agent = detect_agent_info();
    let base = format!(
        "pup/{} (rust; os {}; arch {}",
        version::VERSION,
        std::env::consts::OS,
        std::env::consts::ARCH,
    );
    if agent.detected {
        format!("{}; ai-agent {})", base, agent.name)
    } else {
        format!("{})", base)
    }
}
