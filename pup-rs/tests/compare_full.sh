#!/usr/bin/env bash
# compare_full.sh — Comprehensive end-to-end comparison: Go pup vs Rust pup-rs
# Tests all output formats (json, table, yaml) x modes (human, agent)
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

GO_BIN="$PROJECT_ROOT/pup"
RUST_BIN="$SCRIPT_DIR/../target/release/pup-rs"
MOCK_PORT="${MOCK_PORT:-19877}"

export PUP_MOCK_SERVER="http://localhost:$MOCK_PORT"
export DD_API_KEY="test-key"
export DD_APP_KEY="test-app-key"
export DD_SITE="datadoghq.com"

# Kill any existing mock on this port
lsof -ti:$MOCK_PORT 2>/dev/null | xargs kill 2>/dev/null || true
sleep 0.3

echo "Starting mock server on port $MOCK_PORT..."
"$SCRIPT_DIR/mockdd/mockdd" -port "$MOCK_PORT" &
MOCK_PID=$!
trap "kill $MOCK_PID 2>/dev/null; wait $MOCK_PID 2>/dev/null" EXIT
sleep 1

# Verify mock is up
if ! curl -s "http://localhost:$MOCK_PORT/api/v1/validate" > /dev/null 2>&1; then
    echo "FATAL: Mock server failed to start"
    exit 1
fi

echo '{}' > /tmp/pup_test_body.json

# ============================================================================
# COMPREHENSIVE COMMAND LIST
# Each command is tested in json/table/yaml x human/agent = 6 variants
# ============================================================================
declare -a COMMANDS=(
    # Monitors
    "monitors list"
    "monitors get 12345"
    "monitors search"
    "monitors delete 12345"
    # Dashboards
    "dashboards list"
    "dashboards get test-id-123"
    "dashboards delete test-id-123"
    # Metrics
    "metrics list"
    "metrics search --query avg:system.cpu.user"
    # SLOs
    "slos list"
    "slos get test-id-123"
    "slos delete test-id-123"
    # Synthetics
    "synthetics tests list"
    "synthetics locations list"
    # Events
    "events list"
    "events get 12345"
    # Downtime
    "downtime list"
    "downtime get test-id-123"
    # Tags
    "tags list"
    "tags get test-host"
    "tags delete test-host"
    # Users
    "users list"
    "users get test-id-123"
    "users roles list"
    # Infrastructure
    "infrastructure hosts list"
    # Audit logs
    "audit-logs list"
    # Security
    "security rules list"
    "security rules get test-id-123"
    # Organizations
    "organizations list"
    # Cloud
    "cloud aws list"
    "cloud gcp list"
    "cloud azure list"
    # Cases
    "cases get test-id-123"
    "cases projects list"
    "cases projects get test-id-123"
    # Service catalog
    "service-catalog list"
    "service-catalog get test-service"
    # API keys
    "api-keys list"
    "api-keys get test-id-123"
    "api-keys delete test-id-123"
    # App keys
    "app-keys list"
    "app-keys get test-id-123"
    # Notebooks
    "notebooks list"
    "notebooks get 12345"
    "notebooks delete 12345"
    # RUM
    "rum apps list"
    "rum playlists list"
    # On-call
    "on-call teams list"
    "on-call teams get test-id-123"
    "on-call teams delete test-id-123"
    # Fleet
    "fleet agents list"
    "fleet agents get test-id-123"
    "fleet agents versions"
    "fleet deployments list"
    "fleet deployments get test-id-123"
    "fleet schedules list"
    "fleet schedules get test-id-123"
    "fleet schedules delete test-id-123"
    # Data governance
    "data-governance scanner rules list"
    # Error tracking
    "error-tracking issues search"
    "error-tracking issues get test-id-123"
    # HAMR
    "hamr connections get"
    # Integrations
    "integrations jira accounts list"
    "integrations jira templates list"
    "integrations jira templates get 00000000-0000-0000-0000-000000000001"
    "integrations servicenow instances list"
    "integrations servicenow templates list"
    "integrations servicenow templates get 00000000-0000-0000-0000-000000000001"
    # Cost
    "cost projected"
    # Misc
    "misc ip-ranges"
    "misc status"
    # Investigations
    "investigations list"
    "investigations get test-id-123"
)

# Commands where Go and Rust use different argument styles (Go: --flag, Rust: positional)
# Format: "label|go_args|rust_args"
declare -a SPLIT_COMMANDS=(
    "rum apps get|rum apps get --app-id test-id-123|rum apps get test-id-123"
    "rum apps delete|rum apps delete --app-id test-id-123|rum apps delete test-id-123"
    "rum metrics get|rum metrics get --metric-id test-id-123|rum metrics get test-id-123"
    "rum metrics delete|rum metrics delete --metric-id test-id-123|rum metrics delete test-id-123"
    "cicd pipelines get|cicd pipelines get --pipeline-id test-id-123|cicd pipelines get test-id-123"
)

declare -a FORMATS=("json" "yaml" "table")
declare -a MODES=("human" "agent")

OUTDIR="/tmp/pup_full_compare"
rm -rf "$OUTDIR"
mkdir -p "$OUTDIR/diffs"

total=0
exact=0
diff_count=0
go_fail=0
rust_fail=0
both_fail=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

declare -a DIFF_DETAILS=()

for cmd in "${COMMANDS[@]}"; do
    for fmt in "${FORMATS[@]}"; do
        for mode in "${MODES[@]}"; do
            total=$((total + 1))
            safe="$(echo "${cmd}__${fmt}__${mode}" | tr ' /' '__')"

            agent_flag=""
            if [ "$mode" = "agent" ]; then
                agent_flag="--agent"
            fi

            # Run Go — capture both stdout and stderr
            go_out=$("$GO_BIN" $agent_flag --output "$fmt" $cmd 2>&1)
            go_rc=$?

            # Run Rust — capture both stdout and stderr
            rust_out=$("$RUST_BIN" $agent_flag --output "$fmt" $cmd 2>&1)
            rust_rc=$?

            # Save outputs
            echo "$go_out" > "$OUTDIR/go_${safe}.out"
            echo "$rust_out" > "$OUTDIR/rs_${safe}.out"

            if [ $go_rc -ne 0 ] && [ $rust_rc -ne 0 ]; then
                both_fail=$((both_fail + 1))
                continue
            elif [ $go_rc -ne 0 ]; then
                go_fail=$((go_fail + 1))
                printf "${YELLOW}GO_FAIL ${NC} %-50s fmt=%-5s mode=%s\n" "$cmd" "$fmt" "$mode"
                continue
            elif [ $rust_rc -ne 0 ]; then
                rust_fail=$((rust_fail + 1))
                printf "${YELLOW}RS_FAIL ${NC} %-50s fmt=%-5s mode=%s\n" "$cmd" "$fmt" "$mode"
                # Show the error
                echo "         RS error: $(echo "$rust_out" | head -1)"
                continue
            fi

            # Both succeeded — compare
            if [ "$go_out" = "$rust_out" ]; then
                exact=$((exact + 1))
            else
                diff_count=$((diff_count + 1))
                printf "${RED}DIFF    ${NC} %-50s fmt=%-5s mode=%s\n" "$cmd" "$fmt" "$mode"
                # Save diff
                diff <(echo "$go_out") <(echo "$rust_out") > "$OUTDIR/diffs/${safe}.diff" 2>&1
                # Show first few lines
                diff <(echo "$go_out") <(echo "$rust_out") | head -12 | sed 's/^/         /'
                DIFF_DETAILS+=("$cmd | $fmt | $mode")
            fi
        done
    done
done

# Run split commands (Go and Rust use different arg styles)
for entry in "${SPLIT_COMMANDS[@]}"; do
    IFS='|' read -r label go_cmd rust_cmd <<< "$entry"
    for fmt in "${FORMATS[@]}"; do
        for mode in "${MODES[@]}"; do
            total=$((total + 1))
            safe="$(echo "${label}__${fmt}__${mode}" | tr ' /' '__')"

            agent_flag=""
            if [ "$mode" = "agent" ]; then
                agent_flag="--agent"
            fi

            go_out=$("$GO_BIN" $agent_flag --output "$fmt" $go_cmd 2>&1)
            go_rc=$?
            rust_out=$("$RUST_BIN" $agent_flag --output "$fmt" $rust_cmd 2>&1)
            rust_rc=$?

            echo "$go_out" > "$OUTDIR/go_${safe}.out"
            echo "$rust_out" > "$OUTDIR/rs_${safe}.out"

            if [ $go_rc -ne 0 ] && [ $rust_rc -ne 0 ]; then
                both_fail=$((both_fail + 1))
                continue
            elif [ $go_rc -ne 0 ]; then
                go_fail=$((go_fail + 1))
                printf "${YELLOW}GO_FAIL ${NC} %-50s fmt=%-5s mode=%s\n" "$label" "$fmt" "$mode"
                continue
            elif [ $rust_rc -ne 0 ]; then
                rust_fail=$((rust_fail + 1))
                printf "${YELLOW}RS_FAIL ${NC} %-50s fmt=%-5s mode=%s\n" "$label" "$fmt" "$mode"
                echo "         RS error: $(echo "$rust_out" | head -1)"
                continue
            fi

            if [ "$go_out" = "$rust_out" ]; then
                exact=$((exact + 1))
            else
                diff_count=$((diff_count + 1))
                printf "${RED}DIFF    ${NC} %-50s fmt=%-5s mode=%s\n" "$label" "$fmt" "$mode"
                diff <(echo "$go_out") <(echo "$rust_out") > "$OUTDIR/diffs/${safe}.diff" 2>&1
                diff <(echo "$go_out") <(echo "$rust_out") | head -12 | sed 's/^/         /'
                DIFF_DETAILS+=("$label | $fmt | $mode")
            fi
        done
    done
done

echo ""
echo "============================================"
echo "     COMPREHENSIVE COMPARISON SUMMARY"
echo "============================================"
echo "Total test cases:     $total"
printf "${GREEN}Exact match:          $exact${NC}\n"
printf "${RED}Different output:     $diff_count${NC}\n"
printf "${YELLOW}Go-only failure:      $go_fail${NC}\n"
printf "${YELLOW}Rust-only failure:    $rust_fail${NC}\n"
echo "Both failed:          $both_fail"
echo ""

if [ $diff_count -gt 0 ]; then
    echo "--- DIFF DETAILS ---"
    for d in "${DIFF_DETAILS[@]}"; do
        echo "  $d"
    done
    echo ""
    echo "Diff files saved to: $OUTDIR/diffs/"
fi

if [ $rust_fail -gt 0 ]; then
    echo "--- RUST-ONLY FAILURES (needs investigation) ---"
    echo "Check $OUTDIR/ for rs_*.out files"
fi

if [ $diff_count -eq 0 ] && [ $rust_fail -eq 0 ]; then
    printf "\n${GREEN}PASS: All outputs match!${NC}\n"
else
    printf "\n${RED}FAIL: $diff_count diffs, $rust_fail Rust failures${NC}\n"
fi
