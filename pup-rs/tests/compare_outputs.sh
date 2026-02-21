#!/usr/bin/env bash
# compare_outputs.sh — Compare actual output of Go pup vs Rust pup-rs
# across all output formats (json, table, yaml) and modes (human, agent).
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

GO_BIN="$PROJECT_ROOT/pup"
RUST_BIN="$SCRIPT_DIR/../target/release/pup-rs"
MOCK_PORT="${MOCK_PORT:-19876}"

export PUP_MOCK_SERVER="http://localhost:$MOCK_PORT"
export DD_API_KEY="test-key"
export DD_APP_KEY="test-app-key"
export DD_SITE="datadoghq.com"

# Ensure mock server is running
if ! curl -s "http://localhost:$MOCK_PORT/api/v1/validate" > /dev/null 2>&1; then
    echo "Starting mock server..."
    "$SCRIPT_DIR/mockdd/mockdd" -port "$MOCK_PORT" &
    MOCK_PID=$!
    trap "kill $MOCK_PID 2>/dev/null" EXIT
    sleep 1
fi

echo '{}' > /tmp/pup_test_body.json

# Commands that reliably work against the mock for BOTH CLIs
declare -a COMMANDS=(
    "dashboards list"
    "dashboards get test-id-123"
    "dashboards delete test-id-123"
    "tags list"
    "tags get test-host"
    "users list"
    "users get test-id-123"
    "downtime list"
    "downtime get test-id-123"
    "slos list"
    "slos get test-id-123"
    "slos delete test-id-123"
    "api-keys list"
    "api-keys get test-id-123"
    "api-keys delete test-id-123"
    "events list"
    "cases get test-id-123"
    "incidents list"
    "incidents get test-id-123"
    "service-catalog list"
    "misc ip-ranges"
    "organizations list"
    "security rules list"
    "security rules get test-id-123"
    "audit-logs list"
    "cloud aws list"
    "cloud gcp list"
    "cloud azure list"
    "on-call teams list"
    "fleet agents list"
)

declare -a FORMATS=("json" "table" "yaml")
declare -a MODES=("human" "agent")

OUTDIR="/tmp/pup_output_compare"
rm -rf "$OUTDIR"
mkdir -p "$OUTDIR"

total=0
exact=0
diff_count=0
go_only=0
rust_only=0
both_fail=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

for cmd in "${COMMANDS[@]}"; do
    for fmt in "${FORMATS[@]}"; do
        for mode in "${MODES[@]}"; do
            total=$((total + 1))
            safe="$(echo "${cmd}__${fmt}__${mode}" | tr ' /' '__')"

            agent_flag=""
            if [ "$mode" = "agent" ]; then
                agent_flag="--agent"
            fi

            # Run Go
            go_out=$("$GO_BIN" $agent_flag --output "$fmt" $cmd 2>/dev/null)
            go_rc=$?

            # Run Rust
            rust_out=$("$RUST_BIN" $agent_flag --output "$fmt" $cmd 2>/dev/null)
            rust_rc=$?

            # Save outputs
            echo "$go_out" > "$OUTDIR/go_${safe}.out"
            echo "$rust_out" > "$OUTDIR/rs_${safe}.out"

            if [ $go_rc -ne 0 ] && [ $rust_rc -ne 0 ]; then
                both_fail=$((both_fail + 1))
                continue
            elif [ $go_rc -ne 0 ]; then
                go_only=$((go_only + 1))
                printf "${YELLOW}GO_FAIL ${NC} %-45s fmt=%-5s mode=%s\n" "$cmd" "$fmt" "$mode"
                continue
            elif [ $rust_rc -ne 0 ]; then
                rust_only=$((rust_only + 1))
                printf "${YELLOW}RS_FAIL ${NC} %-45s fmt=%-5s mode=%s\n" "$cmd" "$fmt" "$mode"
                continue
            fi

            # Both succeeded — compare
            if [ "$go_out" = "$rust_out" ]; then
                exact=$((exact + 1))
                printf "${GREEN}MATCH   ${NC} %-45s fmt=%-5s mode=%s\n" "$cmd" "$fmt" "$mode"
            else
                diff_count=$((diff_count + 1))
                printf "${RED}DIFF    ${NC} %-45s fmt=%-5s mode=%s\n" "$cmd" "$fmt" "$mode"
                # Show first difference
                diff_result=$(diff <(echo "$go_out") <(echo "$rust_out") | head -8)
                echo "$diff_result" | sed 's/^/         /'
            fi
        done
    done
done

echo ""
echo "============================================"
echo "         OUTPUT COMPARISON SUMMARY"
echo "============================================"
echo "Total test cases:   $total"
printf "${GREEN}Exact match:        $exact${NC}\n"
printf "${RED}Different output:   $diff_count${NC}\n"
printf "${YELLOW}Go-only success:    $go_only${NC}\n"
printf "${YELLOW}Rust-only success:  $rust_only${NC}\n"
echo "Both failed:        $both_fail"
echo ""

if [ $diff_count -eq 0 ] && [ $go_only -eq 0 ]; then
    printf "${GREEN}PASS: All outputs match!${NC}\n"
else
    echo "Diff files saved to: $OUTDIR"
    printf "\nTo inspect a diff: diff $OUTDIR/go_<name>.out $OUTDIR/rs_<name>.out\n"
fi
