# Performance Benchmarks: Go pup vs Rust pup-rs

Benchmarks run on: 2026-02-21, macOS 26.2 (Darwin 25.2.0), Apple M3 Max, arm64

## Binary Size

| Metric | Go | Rust (release) | Rust (stripped) | Improvement |
|--------|-----|----------------|-----------------|-------------|
| Size (bytes) | 39,156,994 | 32,836,320 | 26,925,712 | 31% smaller (stripped vs Go) |
| Size (MB) | 37.3 MB | 31.3 MB | 25.7 MB | |

The Go binary includes the Go runtime and GC. The Rust release binary includes debug symbols; stripping removes them for a 31% reduction vs Go.

## Startup Time (100 iterations of `--help`)

Measured with `time` over 100 sequential invocations. Median of 3 trials reported.

| Metric | Go | Rust | Improvement |
|--------|-----|------|-------------|
| Total (100 runs) | 0.932s | 0.783s | 16% faster |
| Per invocation | 9.3ms | 7.8ms | 1.5ms faster |

Both CLIs start fast enough that the difference is imperceptible to users. The Go runtime initialization and Rust's lack of a GC contribute to the gap.

## Memory Usage (peak RSS, `monitors list` via mock server)

Measured with `/usr/bin/time -l` against a local mock Datadog API server. Average of 3 runs reported.

| Metric | Go | Rust | Improvement |
|--------|-----|------|-------------|
| Peak RSS (bytes) | 20,245,163 | 15,127,893 | 25% less |
| Peak RSS (MB) | 19.3 MB | 14.4 MB | 4.9 MB less |

The Go runtime pre-allocates heap and stack space for goroutines and the GC. Rust's ownership model and lack of a GC result in a tighter memory footprint.

## Command Coverage

| Metric | Go | Rust |
|--------|-----|------|
| Command groups | 47 | 48 |
| Leaf subcommands | 271 | 283 |
| API output parity | -- | 155/155 (100%) |
| Agent schema parity | -- | 390/390 descriptions |

The Rust version includes 12 additional leaf subcommands (e.g., `dashboards create/update`, `monitors create/update`, `slos create/update`, `downtime create`, `completions`) that were not present in the Go version.

## Methodology

- **Binary size:** Measured with `wc -c`. Rust release built with `cargo build --release`. Stripped with `strip`.
- **Startup time:** `time (for i in $(seq 100); do ./binary --help > /dev/null 2>&1; done)`. No warmup; cold cache on first trial, warm thereafter.
- **Memory:** `/usr/bin/time -l` reporting `maximum resident set size`. Commands run against `mockdd` (Go mock server on localhost:19881) to isolate CLI overhead from network latency.
- **Command coverage:** Go counts from `cmd/root.go` registration. Rust counts from `clap` subcommand enum. API output parity validated via `compare_outputs.sh` diffing JSON responses from both binaries against the same mock server.
