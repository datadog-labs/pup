# Testing Guide

Test strategy, coverage requirements, and CI/CD documentation for Pup.

## Coverage Requirements

**Minimum threshold: 80%** - PRs that drop coverage below 80% will fail CI.

## Running Tests Locally

```bash
# Run all tests with race detection
go test -v -race ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

## Test Organization

### Package Tests (pkg/)

High-coverage unit tests for all packages:

```
pkg/auth/callback/    94.0% coverage
pkg/auth/dcr/         88.1% coverage
pkg/auth/oauth/       91.4% coverage
pkg/auth/storage/     81.8% coverage
pkg/auth/types/      100.0% coverage
pkg/client/           95.5% coverage
pkg/config/          100.0% coverage
pkg/formatter/        93.8% coverage
pkg/util/             96.9% coverage

Average: 93.9% coverage
```

### Command Tests (cmd/)

Structural tests for all commands:
- 26 test files (one per command)
- 163 test functions
- Tests verify: command structure, flags, hierarchy, parent-child relationships

## Test Patterns

### Table-Driven Tests

Preferred pattern for multiple test cases:

```go
func TestParseTimeParam(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        want     time.Time
        wantErr  bool
    }{
        {
            name:    "relative hour",
            input:   "1h",
            want:    time.Now().Add(-1 * time.Hour),
            wantErr: false,
        },
        {
            name:    "relative minutes",
            input:   "30m",
            want:    time.Now().Add(-30 * time.Minute),
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   "invalid",
            want:    time.Time{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseTimeParam(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("parseTimeParam() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && !timesApproxEqual(got, tt.want) {
                t.Errorf("parseTimeParam() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Mock API Responses

Use httptest for mocking Datadog API:

```go
func TestMetricsQuery(t *testing.T) {
    // Setup mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validate request
        assert.Equal(t, "GET", r.Method)
        assert.Contains(t, r.URL.Path, "/api/v2/query/timeseries")

        // Return mock response
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(mockMetricsResponse)
    }))
    defer server.Close()

    // Test with mock server
    client := newTestClient(server.URL)
    result, err := client.QueryMetrics(context.Background(), query)

    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Command Structure Tests

Test command hierarchy and flags:

```go
func TestMetricsCommands(t *testing.T) {
    tests := []struct {
        name     string
        cmd      *cobra.Command
        wantSubs []string
    }{
        {
            name:     "metrics command",
            cmd:      metricsCmd,
            wantSubs: []string{"query", "list", "get", "search"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Verify command has expected subcommands
            subCmds := tt.cmd.Commands()
            assert.Len(t, subCmds, len(tt.wantSubs))

            for _, wantName := range tt.wantSubs {
                found := false
                for _, subCmd := range subCmds {
                    if subCmd.Name() == wantName {
                        found = true
                        break
                    }
                }
                assert.True(t, found, "missing subcommand: %s", wantName)
            }
        })
    }
}
```

## CI/CD Pipeline

> **Note:** Pup uses Datadog CI products for enhanced monitoring and analytics.

GitHub Actions workflow runs on all branches with 4 parallel jobs:

### 1. Test and Coverage

```yaml
- name: Run tests with coverage
  run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

- name: Check coverage threshold
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$coverage < 80" | bc -l) )); then
      echo "Coverage $coverage% is below 80% threshold"
      exit 1
    fi

- name: Upload coverage report
  uses: actions/upload-artifact@v3
  with:
    name: coverage-report
    path: |
      coverage.out
      coverage.html
    retention-days: 30
```

**On Pull Requests:**
- Runs all tests with race detection
- Generates coverage reports (text, HTML)
- Checks coverage meets 80% threshold (fails if below)
- Posts PR comment with coverage breakdown
- Uploads coverage artifacts (retained 30 days)

**On Main Branch:**
- All PR checks plus:
- Updates coverage badge in README.md
- Stores badge data in `.github/badges/coverage.json`

### 2. Lint

```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v3
  with:
    version: latest
    args: --timeout=5m
```

Enforces Go style and best practices.

### 3. Build

```yaml
- name: Build
  run: go build -o pup .

- name: Verify binary
  run: ./pup --version
```

Verifies project builds and binary executes.

### 4. Datadog Static Analysis (SAST)

```yaml
- name: Run Datadog Static Analysis
  run: datadog-ci sast scan --service=pup --env=ci
```

Scans code for security vulnerabilities and quality issues on pull requests.

See **[DATADOG_CI.md](DATADOG_CI.md)** for:
- Test Visibility with orchestrion
- Code Coverage upload to Datadog
- CI Pipeline Visibility tracking
- SAST configuration and results

## Coverage Badge

README.md displays live coverage badge:
- Updates automatically on main branch pushes
- Color-coded: green (80%+), yellow (70-80%), red (<70%)
- Badge data: `.github/badges/coverage.json`
- Display: shields.io endpoint

## PR Coverage Comments

Every PR receives automated comment:

```markdown
## 📊 Code Coverage Report

![Coverage](https://img.shields.io/badge/coverage-93.9%25-brightgreen)

✅ Coverage meets the 80% threshold

### Package Coverage Details

| Package | Coverage |
|---------|----------|
| pkg/auth/callback | 94.0% |
| pkg/auth/dcr | 88.1% |
| pkg/auth/oauth | 91.4% |
| ... | ... |

**Overall:** 93.9%
**Commit:** abc123def
```

## Integration Testing

Integration tests with mocked Datadog API (planned):

```go
func TestEndToEndMetricsQuery(t *testing.T) {
    // Setup mock Datadog API
    mockAPI := setupMockDatadogAPI(t)
    defer mockAPI.Close()

    // Execute command
    cmd := exec.Command("pup", "metrics", "query",
        "--query=avg:system.cpu.user{*}",
        "--from=1h",
        "--api-url="+mockAPI.URL)

    output, err := cmd.CombinedOutput()
    assert.NoError(t, err)
    assert.Contains(t, string(output), "system.cpu.user")
}
```

## OAuth2 Testing

Test OAuth2 flow components:

```go
func TestOAuthFlow(t *testing.T) {
    // Test PKCE generation
    verifier, challenge := generatePKCE()
    assert.Len(t, verifier, 43)
    assert.Len(t, challenge, 43)

    // Test authorization URL generation
    authURL := buildAuthURL(challenge, state)
    assert.Contains(t, authURL, "code_challenge=")
    assert.Contains(t, authURL, "code_challenge_method=S256")

    // Test token exchange
    mockServer := setupMockOAuthServer(t)
    defer mockServer.Close()

    tokens, err := exchangeCodeForTokens(code, verifier, mockServer.URL)
    assert.NoError(t, err)
    assert.NotEmpty(t, tokens.AccessToken)
}
```

## Best Practices

**Do:**
- Write tests before fixing bugs (TDD for bug fixes)
- Use table-driven tests for multiple cases
- Mock external dependencies
- Test error paths, not just happy paths
- Use meaningful test names (`TestParseTimeParam_InvalidInput`)
- Assert specific error types when possible

**Don't:**
- Skip tests or use `t.Skip()` without good reason
- Test implementation details (test behavior, not internals)
- Make tests depend on each other
- Use sleeps for timing (use channels or mock time)
- Commit failing or commented-out tests

## Test Coverage Goals

**Current Status:**
- pkg/ average: 93.9% ✅
- Overall target: 80% ✅

**Future Goals:**
- Integration test suite
- E2E tests with mock API server
- Performance benchmarks
- Fuzz testing for parsers

## Troubleshooting Tests

**Test fails intermittently:**
- Check for race conditions (run with `-race`)
- Look for time-dependent assertions
- Ensure proper cleanup in `defer` statements

**Coverage report inaccurate:**
- Run with `-covermode=atomic` for accurate concurrent coverage
- Check for files in `.gitignore` not counted
- Verify test files have `_test.go` suffix

**Tests slow:**
- Profile tests: `go test -cpuprofile=cpu.prof`
- Check for unnecessary sleeps
- Mock external dependencies
- Use parallel tests: `t.Parallel()`
