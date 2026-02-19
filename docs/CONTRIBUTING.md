# Contributing Guide

Guidelines for contributing to Pup.

## Getting Started

```bash
# Clone repository
git clone https://github.com/datadog-labs/pup.git
cd pup

# Install dependencies
go mod download

# Build
go build -o pup .

# Run tests
go test ./...

# Run with local changes
go run main.go <command>
```

## Development Workflow

### 1. Create Branch

```bash
git checkout -b <type>/<short-description>
```

**Branch prefixes:**
- `feat/` - New features
- `fix/` - Bug fixes
- `refactor/` - Code refactoring
- `docs/` - Documentation updates
- `test/` - Test additions/updates
- `chore/` - Maintenance tasks
- `perf/` - Performance improvements

**Examples:**
```bash
git checkout -b feat/oauth2-token-refresh
git checkout -b fix/metrics-query-timeout
git checkout -b docs/update-readme-oauth
```

### 2. Make Changes

Follow code style guidelines:

**Go Style:**
- Follow standard Go conventions
- Use `gofmt` to format code
- Run `golangci-lint run` before committing
- Keep functions small and focused
- Use clear, descriptive names

**Error Handling:**
```go
// Good: wrap errors with context
if err != nil {
    return fmt.Errorf("failed to query metrics: %w", err)
}

// Bad: lose context
if err != nil {
    return err
}

// Bad: expose secrets
if err != nil {
    return fmt.Errorf("auth failed with key %s: %w", apiKey, err)
}
```

**Testing:**
- Write unit tests for all public functions
- Use table-driven tests for multiple cases
- Mock external dependencies (Datadog API)
- Maintain >80% coverage (CI enforced)

Example table-driven test:
```go
func TestParseTimeParam(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    time.Time
        wantErr bool
    }{
        {"relative hour", "1h", time.Now().Add(-1 * time.Hour), false},
        {"invalid", "bad", time.Time{}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseTimeParam(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            // Assert got matches want
        })
    }
}
```

### 3. Commit Changes

**Stage specific files** (avoid `git add .`):
```bash
git add pkg/auth/oauth/client.go pkg/auth/oauth/client_test.go
```

**Commit with conventional format:**
```bash
git commit -m "$(cat <<'EOF'
<type>(<scope>): <subject>

<body describing what and why>

- Key change 1
- Key change 2
- Key change 3

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"
```

**Commit types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting)
- `refactor` - Code refactoring (no behavior change)
- `test` - Test additions or changes
- `chore` - Build process or tooling changes

**Example:**
```bash
git commit -m "$(cat <<'EOF'
feat(auth): add OAuth2 authentication with PKCE

Implement OAuth2 authentication flow including:
- Dynamic Client Registration (DCR)
- PKCE code challenge generation
- Secure token storage via OS keychain
- Automatic token refresh

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"
```

### 4. Create Pull Request

Use `gh` CLI for efficiency:

```bash
gh pr create \
  --title "<type>(<scope>): <clear, concise title>" \
  --body "$(cat <<'EOF'
## Summary
Brief overview of what this PR does (1-2 sentences).

## Changes
- Specific change 1 with file reference (file.go:123)
- Specific change 2 with file reference
- Specific change 3 with file reference

## Testing
- Test scenarios covered
- How to verify the changes
- Coverage percentage

## Related Issues
Closes #<issue-number>
Fixes #<issue-number>

---
🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)" \
  --label "<labels>"
```

**PR title guidelines:**
- Keep under 70 characters
- Use imperative mood ("add" not "added")
- Be specific about what changed

**PR body guidelines:**
- **Summary**: What and why in 1-2 sentences
- **Changes**: Bulleted list with file references
- **Testing**: How changes were tested
- **Related Issues**: Use `Closes #N` or `Fixes #N`
- **Breaking Changes**: Clearly marked if any
- **Screenshots**: For CLI output changes

**Example PR:**
```bash
gh pr create \
  --title "feat(auth): implement OAuth2 token refresh with PKCE" \
  --body "$(cat <<'EOF'
## Summary
Implements automatic OAuth2 token refresh using PKCE flow to maintain authentication without user intervention.

## Changes
- Added token refresher in pkg/auth/refresh/refresher.go:45
- Implemented background refresh scheduler
- Added unit tests in pkg/auth/refresh/refresher_test.go
- Updated OAuth client to use refresh tokens in pkg/auth/oauth/client.go:123

## Testing
- Unit tests verify refresh token exchange (98% coverage)
- Integration tests validate automatic refresh before expiration
- Manual test: verified token auto-refreshes after 50 minutes
- All existing tests pass

## Related Issues
Closes #42

---
🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)" \
  --label "enhancement,auth"
```

## Code Review Process

1. **Automated Checks**: CI runs tests, linting, coverage checks
2. **Human Review**: Maintainer reviews code quality and design
3. **Address Feedback**: Make requested changes
4. **Approval**: Once approved, PR can be merged
5. **Merge**: Squash and merge to keep history clean

## Testing Requirements

All PRs must:
- Pass all existing tests
- Add tests for new functionality
- Maintain ≥80% code coverage
- Pass `golangci-lint` checks
- Build successfully

See [TESTING.md](TESTING.md) for detailed testing guidelines.

## Security Guidelines

**Never commit:**
- API keys or secrets
- OAuth tokens or client secrets
- Environment variables with credentials
- Test data with real user information

**Always:**
- Validate user inputs to prevent injection
- Use parameterized queries for any data storage
- Wrap errors without exposing sensitive data
- Use HTTPS for all external requests

**OAuth2 Security:**
- Use PKCE S256 for code challenge
- Validate state parameter to prevent CSRF
- Never log or print access/refresh tokens
- Use OS keychain for primary token storage
- Encrypt fallback file storage with AES-256-GCM

## Documentation

When adding features:
- Update relevant documentation files
- Add usage examples to EXAMPLES.md
- Update COMMANDS.md if adding new commands
- Include inline code comments for complex logic

## License

All contributions must be compatible with Apache 2.0 license.

By contributing, you agree that your contributions will be licensed under Apache License 2.0.
