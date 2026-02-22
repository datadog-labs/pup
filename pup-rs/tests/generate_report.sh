#!/usr/bin/env bash
# generate_report.sh — Generate comprehensive HTML report comparing Go pup vs Rust pup-rs
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

GO_BIN="$PROJECT_ROOT/pup"
RUST_BIN="$SCRIPT_DIR/../target/release/pup-rs"
MOCK_PORT="${MOCK_PORT:-19880}"
REPORT="$PROJECT_ROOT/pup-rs/tests/parity_report.html"

# Base env vars for mock server auth (NOT exported to avoid contaminating agent detection)
MOCK_URL="http://localhost:$MOCK_PORT"

# All env vars that can trigger agent mode — must be cleared for human mode tests
AGENT_ENV_VARS=(
    FORCE_AGENT_MODE CLAUDECODE CLAUDE_CODE CURSOR_AGENT
    CODEX OPENAI_CODEX OPENCODE AIDER CLINE
    WINDSURF_AGENT GITHUB_COPILOT AMAZON_Q AWS_Q_DEVELOPER
    GEMINI_CODE_ASSIST SRC_CODY
)

# Start mock server
lsof -ti:$MOCK_PORT 2>/dev/null | xargs kill 2>/dev/null
sleep 0.3
"$SCRIPT_DIR/mockdd/mockdd" -port "$MOCK_PORT" &
MOCK_PID=$!
trap "kill $MOCK_PID 2>/dev/null; wait $MOCK_PID 2>/dev/null" EXIT
sleep 1

if ! curl -s "$MOCK_URL/api/v1/validate" > /dev/null 2>&1; then
    echo "FATAL: Mock server failed to start"
    exit 1
fi

echo '{}' > /tmp/pup_test_body.json

# ============================================================================
# Command definitions: "label|go_cmd|rust_cmd"
# When go_cmd == rust_cmd, just use "cmd" shorthand
# ============================================================================
declare -a TEST_CASES=()

add() {
    local label="$1" go_cmd="$2" rust_cmd="${3:-$2}"
    TEST_CASES+=("$label|$go_cmd|$rust_cmd")
}

# Monitors
add "monitors list" "monitors list"
add "monitors get" "monitors get 12345"
add "monitors search" "monitors search"
add "monitors delete" "monitors delete -y 12345"
# Dashboards
add "dashboards list" "dashboards list"
add "dashboards get" "dashboards get test-id-123"
add "dashboards delete" "dashboards delete -y test-id-123"
# Metrics
add "metrics list" "metrics list"
add "metrics search" "metrics search --query avg:system.cpu.user"
# SLOs
add "slos list" "slos list"
add "slos get" "slos get test-id-123"
add "slos delete" "slos delete -y test-id-123"
# Synthetics
add "synthetics tests list" "synthetics tests list"
add "synthetics locations list" "synthetics locations list"
# Events
add "events list" "events list"
add "events get" "events get 12345"
# Downtime
add "downtime list" "downtime list"
add "downtime get" "downtime get test-id-123"
# Tags
add "tags list" "tags list"
add "tags get" "tags get test-host"
add "tags delete" "tags delete -y test-host"
# Users
add "users list" "users list"
add "users get" "users get test-id-123"
add "users roles list" "users roles list"
# Infrastructure
add "infrastructure hosts list" "infrastructure hosts list"
# Audit logs
add "audit-logs list" "audit-logs list"
# Security
add "security rules list" "security rules list"
add "security rules get" "security rules get test-id-123"
# Organizations
add "organizations list" "organizations list"
# Cloud
add "cloud aws list" "cloud aws list"
add "cloud gcp list" "cloud gcp list"
add "cloud azure list" "cloud azure list"
# Cases
add "cases get" "cases get test-id-123"
add "cases projects list" "cases projects list"
add "cases projects get" "cases projects get test-id-123"
# Service catalog
add "service-catalog list" "service-catalog list"
add "service-catalog get" "service-catalog get test-service"
# API keys
add "api-keys list" "api-keys list"
add "api-keys get" "api-keys get test-id-123"
add "api-keys delete" "api-keys delete -y test-id-123"
# App keys
add "app-keys list" "app-keys list"
add "app-keys get" "app-keys get test-id-123"
# Notebooks
add "notebooks list" "notebooks list"
add "notebooks get" "notebooks get 12345"
add "notebooks delete" "notebooks delete -y 12345"
# RUM
add "rum apps list" "rum apps list"
add "rum apps get" "rum apps get --app-id test-id-123" "rum apps get test-id-123"
add "rum apps delete" "rum apps delete -y --app-id test-id-123" "rum apps delete -y test-id-123"
add "rum metrics list" "rum metrics list"
add "rum metrics get" "rum metrics get --metric-id test-id-123" "rum metrics get test-id-123"
add "rum metrics delete" "rum metrics delete -y --metric-id test-id-123" "rum metrics delete -y test-id-123"
add "rum playlists list" "rum playlists list"
# CI/CD
add "cicd pipelines get" "cicd pipelines get --pipeline-id test-id-123" "cicd pipelines get test-id-123"
# On-call
add "on-call teams list" "on-call teams list"
add "on-call teams get" "on-call teams get test-id-123"
add "on-call teams delete" "on-call teams delete -y test-id-123"
# Fleet
add "fleet agents list" "fleet agents list"
add "fleet agents get" "fleet agents get test-id-123"
add "fleet agents versions" "fleet agents versions"
add "fleet deployments list" "fleet deployments list"
add "fleet deployments get" "fleet deployments get test-id-123"
add "fleet schedules list" "fleet schedules list"
add "fleet schedules get" "fleet schedules get test-id-123"
add "fleet schedules delete" "fleet schedules delete -y test-id-123"
# Data governance
add "data-governance scanner rules list" "data-governance scanner-rules list" "data-governance scanner rules list"
# Error tracking
add "error-tracking issues search" "error-tracking issues search"
add "error-tracking issues get" "error-tracking issues get test-id-123"
# HAMR
add "hamr connections get" "hamr connections get"
# Integrations
add "integrations jira accounts list" "integrations jira accounts list"
add "integrations jira templates list" "integrations jira templates list"
add "integrations jira templates get" "integrations jira templates get 00000000-0000-0000-0000-000000000001"
add "integrations servicenow instances list" "integrations servicenow instances list"
add "integrations servicenow templates list" "integrations servicenow templates list"
add "integrations servicenow templates get" "integrations servicenow templates get 00000000-0000-0000-0000-000000000001"
# Cost
add "cost projected" "cost projected"
# Misc
add "misc ip-ranges" "misc ip-ranges"
add "misc status" "misc status"
# Investigations
add "investigations list" "investigations list"
add "investigations get" "investigations get test-id-123"

# ============================================================================
# Help output tests
# ============================================================================
add "help: top-level" "--help" "--help"
add "help: monitors" "monitors --help" "monitors --help"
add "help: monitors list" "monitors list --help" "monitors list --help"
add "help: dashboards" "dashboards --help" "dashboards --help"
add "help: logs" "logs --help" "logs --help"
add "help: logs search" "logs search --help" "logs search --help"
add "help: metrics" "metrics --help" "metrics --help"
add "help: slos" "slos --help" "slos --help"
add "help: synthetics" "synthetics --help" "synthetics --help"
add "help: events" "events --help" "events --help"
add "help: downtime" "downtime --help" "downtime --help"
add "help: tags" "tags --help" "tags --help"
add "help: users" "users --help" "users --help"
add "help: infrastructure" "infrastructure --help" "infrastructure --help"
add "help: audit-logs" "audit-logs --help" "audit-logs --help"
add "help: security" "security --help" "security --help"
add "help: organizations" "organizations --help" "organizations --help"
add "help: cloud" "cloud --help" "cloud --help"
add "help: cases" "cases --help" "cases --help"
add "help: service-catalog" "service-catalog --help" "service-catalog --help"
add "help: api-keys" "api-keys --help" "api-keys --help"
add "help: app-keys" "app-keys --help" "app-keys --help"
add "help: notebooks" "notebooks --help" "notebooks --help"
add "help: rum" "rum --help" "rum --help"
add "help: cicd" "cicd --help" "cicd --help"
add "help: on-call" "on-call --help" "on-call --help"
add "help: fleet" "fleet --help" "fleet --help"
add "help: data-governance" "data-governance --help" "data-governance --help"
add "help: error-tracking" "error-tracking --help" "error-tracking --help"
add "help: hamr" "hamr --help" "hamr --help"
add "help: integrations" "integrations --help" "integrations --help"
add "help: cost" "cost --help" "cost --help"
add "help: misc" "misc --help" "misc --help"
add "help: investigations" "investigations --help" "investigations --help"
add "help: apm" "apm --help" "apm --help"
add "help: product-analytics" "product-analytics --help" "product-analytics --help"
add "help: static-analysis" "static-analysis --help" "static-analysis --help"
add "help: auth" "auth --help" "auth --help"

# ============================================================================
# Error handling tests
# ============================================================================
# Missing required arguments
add "error: monitors get (no id)" "monitors get" "monitors get"
add "error: dashboards get (no id)" "dashboards get" "dashboards get"
add "error: logs search (no query)" "logs search" "logs search"

# Invalid/non-existent resources (404 from mock)
add "error: monitors get 404" "monitors get 999999999" "monitors get 999999999"
add "error: dashboards get 404" "dashboards get err-not-found" "dashboards get err-not-found"
add "error: cases get 404" "cases get err-not-found" "cases get err-not-found"

# Invalid flag values
add "error: invalid output format" "-o xml monitors list" "-o xml monitors list"
add "error: unknown flag" "monitors list --bogus-flag" "monitors list --bogus-flag"

# Nonexistent subcommand
add "error: unknown subcommand" "monitors bogus" "monitors bogus"
add "error: unknown top-level" "nonexistent" "nonexistent"

# Human-mode help is excluded (Cobra vs clap text formatting is expected to differ).
# Only agent-mode help is tested (both emit JSON schema).
# API commands are tested in both modes.
declare -a MODES=("human" "agent")

# Help tests only run in agent mode (human mode Cobra vs clap always differs)
declare -a HELP_ONLY_AGENT=1

# ============================================================================
# Helper: build clean env for running a CLI command
# Clears ALL agent-detection env vars, then sets only what's needed.
# ============================================================================
run_clean() {
    local bin="$1" mode="$2"
    shift 2
    # Build env: start with base vars, clear all agent env vars
    local -a env_args=(
        "PATH=$PATH"
        "HOME=$HOME"
        "PUP_MOCK_SERVER=$MOCK_URL"
        "DD_API_KEY=test-key"
        "DD_APP_KEY=test-app-key"
        "DD_SITE=datadoghq.com"
    )
    if [ "$mode" = "agent" ]; then
        env_args+=("FORCE_AGENT_MODE=1")
    fi
    # env -i gives a clean slate, then we add only what's needed
    env -i "${env_args[@]}" "$bin" "$@" 2>&1
}

# ============================================================================
# Collect results
# ============================================================================
# Fields: label|mode|go_rc|rust_rc|status|safe|go_full_cmd|rust_full_cmd|env_display
declare -a RESULTS=()

total=0
exact=0
diff_count=0
go_fail=0
rust_fail=0
both_fail=0

OUTDIR="/tmp/pup_report_data"
rm -rf "$OUTDIR"
mkdir -p "$OUTDIR"

echo "Running comparison tests..."

num_tests=$(( ${#TEST_CASES[@]} * ${#MODES[@]} ))

for entry in "${TEST_CASES[@]}"; do
    IFS='|' read -r label go_cmd rust_cmd <<< "$entry"
    for mode in "${MODES[@]}"; do
        # Skip human mode for help and error tests (Cobra vs clap always differs)
        if [ "$mode" = "human" ]; then
            case "$label" in
                help:*|error:*) continue ;;
            esac
        fi
        total=$((total + 1))
        safe="$(echo "${label}_${mode}" | tr ' /' '__')"

        # Build the actual commands (no --agent flag; mode via env var)
        # shellcheck disable=SC2086
        go_out=$(run_clean "$GO_BIN" "$mode" -o json $go_cmd)
        go_rc=$?
        # shellcheck disable=SC2086
        rust_out=$(run_clean "$RUST_BIN" "$mode" -o json $rust_cmd)
        rust_rc=$?

        echo "$go_out" > "$OUTDIR/go_${safe}.txt"
        echo "$rust_out" > "$OUTDIR/rs_${safe}.txt"

        if [ $go_rc -ne 0 ] && [ $rust_rc -ne 0 ]; then
            both_fail=$((both_fail + 1))
            status="both_fail"
        elif [ $go_rc -ne 0 ]; then
            go_fail=$((go_fail + 1))
            status="go_fail"
        elif [ $rust_rc -ne 0 ]; then
            rust_fail=$((rust_fail + 1))
            status="rust_fail"
        elif [ "$go_out" = "$rust_out" ]; then
            exact=$((exact + 1))
            status="match"
        else
            diff_count=$((diff_count + 1))
            status="diff"
        fi

        go_full="pup -o json ${go_cmd}"
        rust_full="pup-rs -o json ${rust_cmd}"

        # Build env display string for this test
        env_display="PUP_MOCK_SERVER=${MOCK_URL}|DD_API_KEY=test-key|DD_APP_KEY=test-app-key|DD_SITE=datadoghq.com"
        if [ "$mode" = "agent" ]; then
            env_display="${env_display}|FORCE_AGENT_MODE=1"
        fi

        RESULTS+=("${label}|${mode}|${go_rc}|${rust_rc}|${status}|${safe}|${go_full}|${rust_full}|${env_display}")
        printf "\r  %d/%d tests completed" "$total" "$num_tests"
    done
done

echo ""
echo "Generating HTML report..."

# ============================================================================
# Generate HTML
# ============================================================================
cat > "$REPORT" << 'HTMLHEAD'
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Pup CLI Parity Report — Go vs Rust</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0d1117; color: #c9d1d9; line-height: 1.5; padding: 20px; }
h1 { color: #f0f6fc; margin-bottom: 8px; font-size: 24px; }
h2 { color: #f0f6fc; margin: 32px 0 16px; font-size: 20px; border-bottom: 1px solid #30363d; padding-bottom: 8px; }
h3 { color: #f0f6fc; margin: 24px 0 8px; font-size: 16px; }
.subtitle { color: #8b949e; margin-bottom: 24px; }
.summary { display: flex; gap: 16px; flex-wrap: wrap; margin-bottom: 24px; }
.stat { background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 16px 24px; text-align: center; min-width: 140px; }
.stat .num { font-size: 32px; font-weight: 700; }
.stat .label { font-size: 12px; color: #8b949e; text-transform: uppercase; letter-spacing: 0.5px; }
.stat.match .num { color: #3fb950; }
.stat.diff .num { color: #f85149; }
.stat.go-fail .num { color: #d29922; }
.stat.rust-fail .num { color: #f97583; }
.stat.both-fail .num { color: #8b949e; }
.badge { display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 12px; font-weight: 600; }
.badge.match { background: #238636; color: #fff; }
.badge.diff { background: #da3633; color: #fff; }
.badge.go-fail { background: #9e6a03; color: #fff; }
.badge.rust-fail { background: #b62324; color: #fff; }
.badge.both-fail { background: #484f58; color: #fff; }
.badge.human { background: #1f6feb; color: #fff; }
.badge.agent { background: #8957e5; color: #fff; }
.alert { background: #f8514926; border: 1px solid #f85149; border-radius: 8px; padding: 16px; margin-bottom: 24px; }
.alert h3 { color: #f85149; margin: 0 0 8px; }
.alert ul { padding-left: 20px; }
.alert li { margin: 4px 0; }
.alert.warning { background: #d2992226; border-color: #d29922; }
.alert.warning h3 { color: #d29922; }
table { width: 100%; border-collapse: collapse; margin: 16px 0; }
th { background: #161b22; color: #f0f6fc; text-align: left; padding: 10px 12px; font-size: 13px; border-bottom: 2px solid #30363d; position: sticky; top: 0; }
td { padding: 8px 12px; border-bottom: 1px solid #21262d; font-size: 13px; vertical-align: top; }
tr:hover { background: #161b22; }
tr.diff-row { background: #f8514910; }
tr.go-fail-row { background: #d2992210; }
tr.rust-fail-row { background: #b6232410; }
.cmd { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 12px; white-space: nowrap; }
.cmd-box { background: #161b22; border: 1px solid #30363d; border-radius: 4px; padding: 6px 10px; margin: 4px 0 8px; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 12px; color: #79c0ff; overflow-x: auto; white-space: nowrap; }
.env-table { border-collapse: collapse; margin-bottom: 12px; width: auto; }
.env-table td { padding: 3px 10px 3px 0; border: none; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 11px; }
.env-table .env-key { color: #d2a8ff; white-space: nowrap; }
.env-table .env-val { color: #79c0ff; }
.env-table .env-eq { color: #484f58; padding: 0 2px; }
.diff-id { display: inline-block; background: #30363d; color: #8b949e; font-size: 10px; font-family: 'SF Mono', 'Fira Code', monospace; padding: 2px 6px; border-radius: 4px; margin-right: 6px; }
.diff-view { background: #0d1117; border: 1px solid #30363d; border-radius: 6px; padding: 0; margin: 8px 0; max-height: 400px; overflow: auto; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 11px; }
.diff-line { padding: 1px 12px; white-space: pre-wrap; word-break: break-all; }
.diff-line.add { background: #1a3a2a; color: #3fb950; }
.diff-line.del { background: #3d1a1a; color: #f85149; }
.diff-line.ctx { color: #8b949e; }
.diff-line.hdr { color: #79c0ff; background: #161b22; font-weight: 600; }
.output-box { background: #0d1117; border: 1px solid #30363d; border-radius: 6px; padding: 12px; margin: 4px 0; max-height: 300px; overflow: auto; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 11px; white-space: pre-wrap; word-break: break-all; }
.output-box.go { border-left: 3px solid #3fb950; }
.output-box.rust { border-left: 3px solid #1f6feb; }
.output-box.error { border-left: 3px solid #f85149; color: #f85149; }
details { margin: 8px 0; }
details summary { cursor: pointer; padding: 8px 12px; background: #161b22; border: 1px solid #30363d; border-radius: 6px; font-size: 13px; }
details summary:hover { background: #1c2128; }
details[open] summary { border-radius: 6px 6px 0 0; }
.detail-content { border: 1px solid #30363d; border-top: none; border-radius: 0 0 6px 6px; padding: 16px; background: #0d1117; }
.side-by-side { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.side-by-side .col-header { font-weight: 600; font-size: 12px; color: #8b949e; text-transform: uppercase; margin-bottom: 4px; }
.timestamp { color: #484f58; font-size: 11px; }
.filter-bar { margin: 16px 0; display: flex; gap: 8px; flex-wrap: wrap; }
.filter-btn { background: #21262d; border: 1px solid #30363d; color: #c9d1d9; padding: 6px 14px; border-radius: 20px; cursor: pointer; font-size: 12px; transition: all 0.15s; }
.filter-btn:hover { background: #30363d; }
.filter-btn.active { background: #1f6feb; border-color: #1f6feb; color: #fff; }
</style>
</head>
<body>
<h1>Pup CLI Parity Report</h1>
<p class="subtitle">Go <code>pup</code> vs Rust <code>pup-rs</code> — End-to-end output comparison against mock Datadog API</p>
HTMLHEAD

# Write timestamp
echo "<p class=\"timestamp\">Generated: $(date -u '+%Y-%m-%d %H:%M:%S UTC')</p>" >> "$REPORT"

# Summary stats
cat >> "$REPORT" << EOF
<h2>Summary</h2>
<div class="summary">
  <div class="stat"><div class="num">${total}</div><div class="label">Total Tests</div></div>
  <div class="stat match"><div class="num">${exact}</div><div class="label">Exact Match</div></div>
  <div class="stat diff"><div class="num">${diff_count}</div><div class="label">Different</div></div>
  <div class="stat go-fail"><div class="num">${go_fail}</div><div class="label">Go-Only Fail</div></div>
  <div class="stat rust-fail"><div class="num">${rust_fail}</div><div class="label">Rust-Only Fail</div></div>
  <div class="stat both-fail"><div class="num">${both_fail}</div><div class="label">Both Fail</div></div>
</div>
EOF

# Issues section at the top
has_issues=false
for r in "${RESULTS[@]}"; do
    IFS='|' read -r label mode go_rc rust_rc status safe go_full rust_full env_display <<< "$r"
    if [ "$status" != "match" ]; then
        has_issues=true
        break
    fi
done

if [ "$has_issues" = true ]; then
    # Differences — with deterministic IDs based on command+mode
    if [ "$diff_count" -gt 0 ]; then
        echo '<div class="alert">' >> "$REPORT"
        echo '<h3>Output Differences</h3><table style="width:100%;border-collapse:collapse">' >> "$REPORT"
        echo '<tr><th>ID</th><th>Command</th><th style="width:80px">Mode</th></tr>' >> "$REPORT"
        for r in "${RESULTS[@]}"; do
            IFS='|' read -r label mode go_rc rust_rc status safe go_full rust_full env_display <<< "$r"
            if [ "$status" = "diff" ]; then
                echo "<tr><td><a href=\"#${safe}\" style=\"color:#f85149;font-weight:700;font-family:monospace;font-size:12px\">${safe}</a></td><td>${label}</td><td>${mode}</td></tr>" >> "$REPORT"
            fi
        done
        echo '</table></div>' >> "$REPORT"
    fi

    # Go-only failures
    if [ "$go_fail" -gt 0 ]; then
        echo '<div class="alert warning">' >> "$REPORT"
        echo '<h3>Go-Only Failures (Rust works, Go crashes)</h3><ul>' >> "$REPORT"
        for r in "${RESULTS[@]}"; do
            IFS='|' read -r label mode go_rc rust_rc status safe go_full rust_full env_display <<< "$r"
            if [ "$status" = "go_fail" ]; then
                go_err=$(head -1 "$OUTDIR/go_${safe}.txt" 2>/dev/null)
                echo "<li><a href=\"#${safe}\" style=\"color:#d29922;font-weight:700;font-family:monospace;font-size:12px\">${safe}</a> — <code>${go_err}</code></li>" >> "$REPORT"
            fi
        done
        echo '</ul></div>' >> "$REPORT"
    fi

    # Rust-only failures
    if [ "$rust_fail" -gt 0 ]; then
        echo '<div class="alert" style="background:#b6232410;border-color:#b62324">' >> "$REPORT"
        echo '<h3 style="color:#f97583">Rust-Only Failures</h3><ul>' >> "$REPORT"
        for r in "${RESULTS[@]}"; do
            IFS='|' read -r label mode go_rc rust_rc status safe go_full rust_full env_display <<< "$r"
            if [ "$status" = "rust_fail" ]; then
                rs_err=$(head -1 "$OUTDIR/rs_${safe}.txt" 2>/dev/null)
                echo "<li><a href=\"#${safe}\" style=\"color:#f97583;font-weight:700;font-family:monospace;font-size:12px\">${safe}</a> — <code>${rs_err}</code></li>" >> "$REPORT"
            fi
        done
        echo '</ul></div>' >> "$REPORT"
    fi

    # Both fail
    if [ "$both_fail" -gt 0 ]; then
        echo '<div class="alert" style="background:#484f5810;border-color:#484f58">' >> "$REPORT"
        echo '<h3 style="color:#8b949e">Both Fail</h3><ul>' >> "$REPORT"
        for r in "${RESULTS[@]}"; do
            IFS='|' read -r label mode go_rc rust_rc status safe go_full rust_full env_display <<< "$r"
            if [ "$status" = "both_fail" ]; then
                echo "<li><a href=\"#${safe}\" style=\"color:#8b949e;font-weight:700;font-family:monospace;font-size:12px\">${safe}</a> — <strong>${label}</strong> (${mode})</li>" >> "$REPORT"
            fi
        done
        echo '</ul></div>' >> "$REPORT"
    fi
fi

# Filter bar
cat >> "$REPORT" << 'FILTERBAR'
<div class="filter-bar">
  <button class="filter-btn active" onclick="filterRows('all')">All</button>
  <button class="filter-btn" onclick="filterRows('match')">Match</button>
  <button class="filter-btn" onclick="filterRows('diff')">Diff</button>
  <button class="filter-btn" onclick="filterRows('go_fail')">Go Fail</button>
  <button class="filter-btn" onclick="filterRows('rust_fail')">Rust Fail</button>
  <button class="filter-btn" onclick="filterRows('both_fail')">Both Fail</button>
</div>
FILTERBAR

# Detail rows
echo '<h2>Detailed Results</h2>' >> "$REPORT"

for r in "${RESULTS[@]}"; do
    IFS='|' read -r label mode go_rc rust_rc status safe go_full rust_full env_display <<< "$r"

    go_out=$(cat "$OUTDIR/go_${safe}.txt" 2>/dev/null)
    rust_out=$(cat "$OUTDIR/rs_${safe}.txt" 2>/dev/null)

    row_class=""
    [ "$status" = "diff" ] && row_class="diff-row"
    [ "$status" = "go_fail" ] && row_class="go-fail-row"
    [ "$status" = "rust_fail" ] && row_class="rust-fail-row"

    badge_class="$status"
    badge_text="Match"
    [ "$status" = "diff" ] && badge_text="Diff"
    [ "$status" = "go_fail" ] && badge_text="Go Fail"
    [ "$status" = "rust_fail" ] && badge_text="Rust Fail"
    [ "$status" = "both_fail" ] && badge_text="Both Fail"

    # Assign deterministic ID based on command+mode for non-match entries
    id_html=""
    anchor=""
    if [ "$status" != "match" ]; then
        id_html="<span class=\"diff-id\">${safe}</span>"
        anchor=" id=\"${safe}\""
    fi

    mode_badge="human"
    [ "$mode" = "agent" ] && mode_badge="agent"

    # Escape HTML in commands
    go_cmd_escaped=$(echo "$go_full" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g')
    rust_cmd_escaped=$(echo "$rust_full" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g')

    # Build env table
    env_table_rows="<tr><td class=\"env-key\">PUP_MOCK_SERVER</td><td class=\"env-eq\">=</td><td class=\"env-val\">${MOCK_URL}</td></tr>"
    env_table_rows="${env_table_rows}<tr><td class=\"env-key\">DD_API_KEY</td><td class=\"env-eq\">=</td><td class=\"env-val\">test-key</td></tr>"
    env_table_rows="${env_table_rows}<tr><td class=\"env-key\">DD_APP_KEY</td><td class=\"env-eq\">=</td><td class=\"env-val\">test-app-key</td></tr>"
    env_table_rows="${env_table_rows}<tr><td class=\"env-key\">DD_SITE</td><td class=\"env-eq\">=</td><td class=\"env-val\">datadoghq.com</td></tr>"
    if [ "$mode" = "agent" ]; then
        env_table_rows="${env_table_rows}<tr><td class=\"env-key\">FORCE_AGENT_MODE</td><td class=\"env-eq\">=</td><td class=\"env-val\">1</td></tr>"
    fi

    # For diffs and failures: generate unified diff view with colored lines
    diff_html=""
    if [ "$status" != "match" ]; then
        diff_lines=$(diff -u "$OUTDIR/go_${safe}.txt" "$OUTDIR/rs_${safe}.txt" 2>/dev/null | head -100 | while IFS= read -r line; do
            escaped=$(echo "$line" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g')
            case "$line" in
                @@*) echo "<div class=\"diff-line hdr\">${escaped}</div>" ;;
                +*) echo "<div class=\"diff-line add\">${escaped}</div>" ;;
                -*) echo "<div class=\"diff-line del\">${escaped}</div>" ;;
                *) echo "<div class=\"diff-line ctx\">${escaped}</div>" ;;
            esac
        done)
        diff_html="<h3 style=\"font-size:13px;margin:12px 0 4px;color:#8b949e\">Unified Diff (Go → Rust)</h3><div class=\"diff-view\">${diff_lines}</div>"
    fi

    # Side-by-side output (escaped)
    go_escaped=$(echo "$go_out" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g' | head -80)
    rust_escaped=$(echo "$rust_out" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g' | head -80)

    go_box_class="go"
    rust_box_class="rust"
    [ "$go_rc" -ne 0 ] && go_box_class="error"
    [ "$rust_rc" -ne 0 ] && rust_box_class="error"

    cat >> "$REPORT" << EOF
<details data-status="${status}" class="${row_class}"${anchor}>
  <summary>
    ${id_html}<span class="badge ${badge_class}">${badge_text}</span>
    <span class="badge ${mode_badge}">${mode}</span>
    <strong>${label}</strong>
  </summary>
  <div class="detail-content">
    <table class="env-table">${env_table_rows}</table>
    <div class="side-by-side">
      <div>
        <div class="col-header">Go (exit ${go_rc})</div>
        <div class="cmd-box">\$ ${go_cmd_escaped}</div>
        <div class="output-box ${go_box_class}">${go_escaped}</div>
      </div>
      <div>
        <div class="col-header">Rust (exit ${rust_rc})</div>
        <div class="cmd-box">\$ ${rust_cmd_escaped}</div>
        <div class="output-box ${rust_box_class}">${rust_escaped}</div>
      </div>
    </div>
    ${diff_html}
  </div>
</details>
EOF
done

# JavaScript for filtering
cat >> "$REPORT" << 'JSBLOCK'
<script>
function filterRows(status) {
  document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
  event.target.classList.add('active');
  document.querySelectorAll('details').forEach(d => {
    if (status === 'all' || d.dataset.status === status) {
      d.style.display = '';
    } else {
      d.style.display = 'none';
    }
  });
}
// Auto-expand diffs and failures on load
document.querySelectorAll('details[data-status="diff"], details[data-status="rust_fail"], details[data-status="go_fail"]').forEach(d => d.open = true);
</script>
</body>
</html>
JSBLOCK

echo ""
echo "Report generated: $REPORT"
echo "  Total: $total | Match: $exact | Diff: $diff_count | Go-fail: $go_fail | Rust-fail: $rust_fail | Both-fail: $both_fail"
