#!/usr/bin/env bash
# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache License Version 2.0.
# This product includes software developed at Datadog (https://www.datadoghq.com/).
# Copyright 2024-present Datadog, Inc.

# run_exhaustive.sh -- End-to-end exhaustive comparison of Go pup vs Rust pup-rs.
#
# Workflow:
#   1. Build all binaries (Go pup, Rust pup-rs, mock server, helper tools).
#   2. Start the mock Datadog API server.
#   3. Generate command lists by walking --help output of both CLIs.
#   4. Run every Go command and capture the request log.
#   5. Run every Rust command and capture the request log.
#   6. Compare the two logs for endpoint parity.
#
# Exit code 0 means full parity; non-zero means gaps remain.
#
# Environment variables:
#   MOCK_PORT   -- port for the mock server (default: 19876)
#   SKIP_BUILD  -- if non-empty, skip the build steps
#   KEEP_LOGS   -- if non-empty, keep intermediate log files
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
RUST_ROOT="$SCRIPT_DIR/.."
MOCK_PORT="${MOCK_PORT:-19876}"
MOCK_URL="http://localhost:$MOCK_PORT"

# -- Colours -----------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { printf "${CYAN}[INFO]${NC}  %s\n" "$*"; }
warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$*"; }
err()   { printf "${RED}[ERR]${NC}   %s\n" "$*" >&2; }
ok()    { printf "${GREEN}[OK]${NC}    %s\n" "$*"; }

# -- Build all binaries ------------------------------------------------------
if [[ -z "${SKIP_BUILD:-}" ]]; then
    echo "=== Building all binaries ==="

    # Build Go pup
    info "Building Go pup..."
    cd "$PROJECT_ROOT"
    go build -o "$PROJECT_ROOT/pup" .

    # Build Rust pup-rs
    info "Building Rust pup-rs..."
    cd "$RUST_ROOT"
    cargo build --release 2>/dev/null
    RUST_BIN="$RUST_ROOT/target/release/pup-rs"

    # Build mock server
    info "Building mock server..."
    cd "$SCRIPT_DIR/mockdd"
    go build -o "$SCRIPT_DIR/mockdd/mockdd" .

    # Build helper tools
    info "Building gen_commands..."
    cd "$SCRIPT_DIR/gen_commands"
    go build -o "$SCRIPT_DIR/gen_commands_bin" .

    info "Building compare..."
    cd "$SCRIPT_DIR/compare"
    go build -o "$SCRIPT_DIR/compare_bin" .
else
    info "Skipping build (SKIP_BUILD set)"
    RUST_BIN="$RUST_ROOT/target/release/pup-rs"
fi

echo ""
echo "=== Starting mock server on port $MOCK_PORT ==="

# Kill any existing mock server on this port.
pkill -f "mockdd -port $MOCK_PORT" 2>/dev/null || true
sleep 0.5

"$SCRIPT_DIR/mockdd/mockdd" -port "$MOCK_PORT" &
MOCK_PID=$!

# Clean up mock server on exit.
cleanup() {
    if [[ -n "${MOCK_PID:-}" ]]; then
        info "Stopping mock server (PID $MOCK_PID)..."
        kill "$MOCK_PID" 2>/dev/null || true
        wait "$MOCK_PID" 2>/dev/null || true
    fi
    if [[ -z "${KEEP_LOGS:-}" ]]; then
        rm -f /tmp/pup_mock_requests.jsonl
    fi
}
trap cleanup EXIT

# Wait for server to be ready.
for i in $(seq 1 20); do
    if curl -s "http://localhost:$MOCK_PORT/api/v1/validate" > /dev/null 2>&1; then
        ok "Mock server ready (PID $MOCK_PID)"
        break
    fi
    if ! kill -0 "$MOCK_PID" 2>/dev/null; then
        err "Mock server exited prematurely"
        exit 1
    fi
    sleep 0.25
done

if ! curl -s "http://localhost:$MOCK_PORT/api/v1/validate" > /dev/null 2>&1; then
    err "Mock server did not become ready in time"
    exit 1
fi

echo ""
echo "=== Generating command lists ==="
cd "$SCRIPT_DIR"
./gen_commands_bin "$PROJECT_ROOT/pup" "$RUST_BIN"

# -- Common environment for both CLIs ----------------------------------------
export DD_API_KEY="test-key"
export DD_APP_KEY="test-app-key"
export DD_SITE="datadoghq.com"
export PUP_MOCK_SERVER="$MOCK_URL"

echo ""
echo "=== Running Go commands ==="
GO_LOG="/tmp/pup_mock_go.jsonl"
> /tmp/pup_mock_requests.jsonl

GO_TOTAL=0
GO_FAILED=0
while IFS= read -r cmd; do
    [[ -z "$cmd" || "$cmd" == \#* ]] && continue
    GO_TOTAL=$((GO_TOTAL + 1))
    if ! eval "$cmd" >/dev/null 2>&1; then
        GO_FAILED=$((GO_FAILED + 1))
    fi
done < go_commands.txt

cp /tmp/pup_mock_requests.jsonl "$GO_LOG"
GO_COUNT=$(wc -l < "$GO_LOG" | tr -d ' ')
ok "Go: ran $GO_TOTAL commands ($GO_FAILED failed), $GO_COUNT requests logged"

echo ""
echo "=== Running Rust commands ==="
RUST_LOG="/tmp/pup_mock_rust.jsonl"
> /tmp/pup_mock_requests.jsonl

RUST_TOTAL=0
RUST_FAILED=0
while IFS= read -r cmd; do
    [[ -z "$cmd" || "$cmd" == \#* ]] && continue
    RUST_TOTAL=$((RUST_TOTAL + 1))
    if ! eval "$cmd" >/dev/null 2>&1; then
        RUST_FAILED=$((RUST_FAILED + 1))
    fi
done < rust_commands.txt

cp /tmp/pup_mock_requests.jsonl "$RUST_LOG"
RUST_COUNT=$(wc -l < "$RUST_LOG" | tr -d ' ')
ok "Rust: ran $RUST_TOTAL commands ($RUST_FAILED failed), $RUST_COUNT requests logged"

echo ""
echo "=== Comparing requests ==="
./compare_bin "$GO_LOG" "$RUST_LOG"
