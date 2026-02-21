#!/usr/bin/env bash
# generate_report.sh — Generate comprehensive HTML report comparing Go pup vs Rust pup-rs
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

GO_BIN="$PROJECT_ROOT/pup"
RUST_BIN="$SCRIPT_DIR/../target/release/pup-rs"
MOCK_PORT="${MOCK_PORT:-19880}"
REPORT="$PROJECT_ROOT/pup-rs/tests/parity_report.html"

export PUP_MOCK_SERVER="http://localhost:$MOCK_PORT"
export DD_API_KEY="test-key"
export DD_APP_KEY="test-app-key"
export DD_SITE="datadoghq.com"

# Start mock server
lsof -ti:$MOCK_PORT 2>/dev/null | xargs kill 2>/dev/null
sleep 0.3
"$SCRIPT_DIR/mockdd/mockdd" -port "$MOCK_PORT" &
MOCK_PID=$!
trap "kill $MOCK_PID 2>/dev/null; wait $MOCK_PID 2>/dev/null" EXIT
sleep 1

if ! curl -s "http://localhost:$MOCK_PORT/api/v1/validate" > /dev/null 2>&1; then
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
add "monitors delete" "monitors delete 12345"
# Dashboards
add "dashboards list" "dashboards list"
add "dashboards get" "dashboards get test-id-123"
add "dashboards delete" "dashboards delete test-id-123"
# Metrics
add "metrics list" "metrics list"
add "metrics search" "metrics search --query avg:system.cpu.user"
# SLOs
add "slos list" "slos list"
add "slos get" "slos get test-id-123"
add "slos delete" "slos delete test-id-123"
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
add "tags delete" "tags delete test-host"
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
add "api-keys delete" "api-keys delete test-id-123"
# App keys
add "app-keys list" "app-keys list"
add "app-keys get" "app-keys get test-id-123"
# Notebooks
add "notebooks list" "notebooks list"
add "notebooks get" "notebooks get 12345"
add "notebooks delete" "notebooks delete 12345"
# RUM
add "rum apps list" "rum apps list"
add "rum apps get" "rum apps get --app-id test-id-123" "rum apps get test-id-123"
add "rum apps delete" "rum apps delete --app-id test-id-123" "rum apps delete test-id-123"
add "rum metrics list" "rum metrics list"
add "rum metrics get" "rum metrics get --metric-id test-id-123" "rum metrics get test-id-123"
add "rum metrics delete" "rum metrics delete --metric-id test-id-123" "rum metrics delete test-id-123"
add "rum playlists list" "rum playlists list"
# CI/CD
add "cicd pipelines get" "cicd pipelines get --pipeline-id test-id-123" "cicd pipelines get test-id-123"
# On-call
add "on-call teams list" "on-call teams list"
add "on-call teams get" "on-call teams get test-id-123"
add "on-call teams delete" "on-call teams delete test-id-123"
# Fleet
add "fleet agents list" "fleet agents list"
add "fleet agents get" "fleet agents get test-id-123"
add "fleet agents versions" "fleet agents versions"
add "fleet deployments list" "fleet deployments list"
add "fleet deployments get" "fleet deployments get test-id-123"
add "fleet schedules list" "fleet schedules list"
add "fleet schedules get" "fleet schedules get test-id-123"
add "fleet schedules delete" "fleet schedules delete test-id-123"
# Data governance
add "data-governance scanner-rules list" "data-governance scanner-rules list"
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

declare -a MODES=("human" "agent")

# ============================================================================
# Collect results
# ============================================================================
declare -a RESULTS=()  # "label|mode|go_rc|rust_rc|status|safe|go_full_cmd|rust_full_cmd"

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

for entry in "${TEST_CASES[@]}"; do
    IFS='|' read -r label go_cmd rust_cmd <<< "$entry"
    for mode in "${MODES[@]}"; do
        total=$((total + 1))
        safe="$(echo "${label}_${mode}" | tr ' /' '__')"

        agent_flag=""
        if [ "$mode" = "agent" ]; then
            agent_flag="--agent"
        fi

        go_out=$("$GO_BIN" $agent_flag -o json $go_cmd 2>&1)
        go_rc=$?
        rust_out=$("$RUST_BIN" $agent_flag -o json $rust_cmd 2>&1)
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

        go_full="pup ${agent_flag} -o json ${go_cmd}"
        rust_full="pup-rs ${agent_flag} -o json ${rust_cmd}"
        RESULTS+=("${label}|${mode}|${go_rc}|${rust_rc}|${status}|${safe}|${go_full}|${rust_full}")
        printf "\r  %d/%d tests completed" "$total" "${#TEST_CASES[@]}"
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
.env-vars { background: #1c1c2e; border: 1px solid #30363d; border-radius: 4px; padding: 6px 10px; margin-bottom: 12px; font-family: 'SF Mono', 'Fira Code', monospace; font-size: 11px; color: #8b949e; overflow-x: auto; white-space: nowrap; }
.env-vars code { color: #d2a8ff; }
.env-label { background: #30363d; color: #8b949e; padding: 1px 6px; border-radius: 3px; font-size: 10px; font-weight: 600; margin-right: 6px; }
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
    IFS='|' read -r label mode go_rc rust_rc status safe <<< "$r"
    if [ "$status" != "match" ]; then
        has_issues=true
        break
    fi
done

if [ "$has_issues" = true ]; then
    # Differences
    if [ "$diff_count" -gt 0 ]; then
        echo '<div class="alert">' >> "$REPORT"
        echo '<h3>Output Differences (inspect these)</h3><ul>' >> "$REPORT"
        for r in "${RESULTS[@]}"; do
            IFS='|' read -r label mode go_rc rust_rc status safe go_full rust_full <<< "$r"
            if [ "$status" = "diff" ]; then
                echo "<li><strong>${label}</strong> (${mode} mode)</li>" >> "$REPORT"
            fi
        done
        echo '</ul></div>' >> "$REPORT"
    fi

    # Go-only failures
    if [ "$go_fail" -gt 0 ]; then
        echo '<div class="alert warning">' >> "$REPORT"
        echo '<h3>Go-Only Failures (Rust works, Go crashes)</h3><ul>' >> "$REPORT"
        for r in "${RESULTS[@]}"; do
            IFS='|' read -r label mode go_rc rust_rc status safe go_full rust_full <<< "$r"
            if [ "$status" = "go_fail" ]; then
                go_err=$(head -1 "$OUTDIR/go_${safe}.txt" 2>/dev/null)
                echo "<li><strong>${label}</strong> (${mode}) — <code>${go_err}</code></li>" >> "$REPORT"
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

ENV_DISPLAY="PUP_MOCK_SERVER=http://localhost:${MOCK_PORT} DD_API_KEY=test-key DD_APP_KEY=test-app-key DD_SITE=datadoghq.com"

for r in "${RESULTS[@]}"; do
    IFS='|' read -r label mode go_rc rust_rc status safe go_full rust_full <<< "$r"

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

    mode_badge="human"
    [ "$mode" = "agent" ] && mode_badge="agent"

    # Escape HTML in output
    go_escaped=$(echo "$go_out" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g' | head -80)
    rust_escaped=$(echo "$rust_out" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g' | head -80)

    # Escape HTML in commands
    go_cmd_escaped=$(echo "$go_full" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g')
    rust_cmd_escaped=$(echo "$rust_full" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g')

    go_box_class="go"
    rust_box_class="rust"
    [ "$go_rc" -ne 0 ] && go_box_class="error"
    [ "$rust_rc" -ne 0 ] && rust_box_class="error"

    cat >> "$REPORT" << EOF
<details data-status="${status}" class="${row_class}">
  <summary>
    <span class="badge ${badge_class}">${badge_text}</span>
    <span class="badge ${mode_badge}">${mode}</span>
    <strong>${label}</strong>
  </summary>
  <div class="detail-content">
    <div class="env-vars"><span class="env-label">ENV</span> <code>${ENV_DISPLAY}</code></div>
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
