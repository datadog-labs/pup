# Troubleshooting Guide

Common issues and solutions for Pup CLI.

## Authentication Issues

### OAuth2 Login Fails

**Symptoms:**
```
Error: failed to complete OAuth login
```

**Common causes:**

1. **Network connectivity**
   ```bash
   # Test connectivity to Datadog
   curl -I https://datadoghq.com

   # Check DNS resolution
   nslookup datadoghq.com
   ```

2. **Firewall blocking localhost**
   - Callback server needs to bind to `127.0.0.1:<random-port>`
   - Check firewall allows connections to localhost
   - Try temporarily disabling firewall

3. **Browser doesn't open**
   ```
   ⚠️  Could not open browser automatically
   Please open this URL manually: https://...
   ```
   - Copy URL and paste in browser manually
   - Check `$BROWSER` environment variable
   - Try setting: `export BROWSER=chrome`

4. **Port already in use**
   - CLI automatically tries random available port
   - If error persists, check for port conflicts:
   ```bash
   # List processes listening on local ports
   lsof -i -P | grep LISTEN | grep 127.0.0.1
   ```

**Solutions:**
```bash
# Try with verbose logging
pup --verbose auth login

# Specify site explicitly
pup --site=datadoghq.com auth login

# Check authentication status
pup auth status
```

### Token Refresh Fails

**Symptoms:**
```
Error: failed to refresh access token
⚠️  Token expired. Run 'pup auth refresh' or 'pup auth login'
```

**Causes:**
- Refresh token expired (30-day lifetime)
- Network connectivity lost
- OAuth client revoked
- Invalid stored tokens

**Solutions:**
```bash
# Try manual refresh
pup auth refresh

# If refresh fails, re-authenticate
pup auth logout
pup auth login

# Check stored tokens (debug only)
ls -la ~/.config/pup/tokens_*.json
```

### Keychain Access Denied

**macOS symptoms:**
```
Warning: keychain access denied, using encrypted file storage
```

**Solutions:**

1. **Grant keychain access:**
   - Open "Keychain Access" app
   - Search for "pup"
   - Right-click → "Get Info"
   - Grant access to pup binary

2. **Use fallback storage:**
   - Pup automatically falls back to encrypted file
   - Check: `~/.config/pup/tokens.enc`
   - File permissions should be `0600`

### API Key Authentication Fails

**Symptoms:**
```
Error: authentication failed: 403 Forbidden
```

**Check environment variables:**
```bash
# Verify keys are set
echo $DD_API_KEY
echo $DD_APP_KEY
echo $DD_SITE

# Set if missing
export DD_API_KEY="your-api-key"
export DD_APP_KEY="your-app-key"
export DD_SITE="datadoghq.com"
```

**Validate keys:**
```bash
# Test with curl
curl -X GET "https://api.datadoghq.com/api/v1/validate" \
  -H "DD-API-KEY: ${DD_API_KEY}" \
  -H "DD-APPLICATION-KEY: ${DD_APP_KEY}"
```

### Error Tracking OAuth2 Scope Issues

**Symptoms:**
```
Error: 401 Unauthorized
# When using error-tracking commands with OAuth2

Error: OAuth error: invalid_scope
# During OAuth2 login after adding error_tracking_read scope
```

**Background:**

The `error_tracking_read` OAuth2 scope is [documented](https://docs.datadoghq.com/api/latest/scopes/) and required for Error Tracking API endpoints. Pup v0.2.0+ includes this scope in the default OAuth2 scopes list.

However, there may be scenarios where Datadog's OAuth2 authorization endpoint rejects the scope during login:

1. **Scope not available for Dynamic Client Registration (DCR)**: Some OAuth2 scopes may only be available for pre-registered OAuth applications, not for dynamically registered clients like pup uses.

2. **Org-level permissions required**: Your Datadog organization may need specific Error Tracking features or plan tiers enabled before the scope becomes available.

3. **Timing/rollout issues**: The scope might not yet be available in all Datadog regions or for all customers.

**Workaround - Use API Keys:**

If you encounter OAuth2 issues with error-tracking commands, use API key authentication instead:

```bash
# Logout from OAuth2
pup auth logout

# Set API keys
export DD_API_KEY="your-api-key"
export DD_APP_KEY="your-app-key"
export DD_SITE="datadoghq.com"

# Use error-tracking commands
pup error-tracking issues search
pup error-tracking issues get <issue-id>
```

**Testing OAuth2 scope availability:**

If you want to test whether the scope works for your organization:

```bash
# 1. Backup existing OAuth2 credentials
mkdir -p ~/.config/pup/backup
cp ~/.config/pup/tokens_*.json ~/.config/pup/backup/ 2>/dev/null || true
cp ~/.config/pup/client_*.json ~/.config/pup/backup/ 2>/dev/null || true

# 2. Logout and re-login to trigger new OAuth2 flow
pup auth logout
pup auth login

# 3. Test error-tracking command
pup error-tracking issues search --from=1d

# If you get "invalid_scope" error during login, the scope is not available
# If you get 401 during the command, there may be permission issues

# 4. Restore backup if needed
cp ~/.config/pup/backup/*.json ~/.config/pup/ 2>/dev/null || true
```

**Reporting scope issues:**

If you encounter OAuth2 scope problems with error-tracking:

1. Confirm your Datadog org has Error Tracking enabled
2. Verify API key authentication works: `pup error-tracking issues search`
3. Report to [Datadog Support](https://help.datadoghq.com/) if the scope should be available
4. Open a [GitHub issue](https://github.com/DataDog/pup/issues) if this is a pup-specific problem

## API Call Issues

### Rate Limiting

**Symptoms:**
```
Error: 429 Too Many Requests
Rate limit exceeded
```

**Solutions:**
- Wait before retrying
- Reduce number of concurrent requests
- Check your Datadog plan limits
- Use pagination with smaller page sizes

**Workaround:**
```bash
# Add delay between requests
for id in $(cat ids.txt); do
  pup monitors get "$id"
  sleep 1  # Wait 1 second between requests
done
```

### Timeout Errors

**Symptoms:**
```
Error: context deadline exceeded
Error: request timeout
```

**Causes:**
- Network latency
- Large result set
- Datadog API slow response

**Solutions:**
```bash
# Use pagination
pup monitors list --limit=100

# Use shorter time ranges
pup logs search --query="..." --from="30m"  # Instead of 24h

# Check network latency
ping api.datadoghq.com
```

### 404 Not Found

**Symptoms:**
```
Error: 404 Not Found
Resource not found: monitor 12345678
```

**Causes:**
- Resource deleted
- Wrong resource ID
- Wrong Datadog site
- Insufficient permissions

**Solutions:**
```bash
# Verify resource exists
pup monitors list | grep "12345678"

# Check you're on correct site
pup --verbose monitors get 12345678

# Try with different site
pup --site=datadoghq.eu monitors get 12345678
```

## Command Issues

### Command Not Found

**Symptoms:**
```
Error: unknown command "foo" for "pup"
```

**Solutions:**
```bash
# List available commands
pup --help

# Check command spelling
pup metrics --help

# Verify command exists
pup help metrics query
```

### Invalid Flags

**Symptoms:**
```
Error: unknown flag: --foo
```

**Solutions:**
```bash
# Check available flags
pup metrics query --help

# Common flag mistakes:
pup metrics query --query="..." --from="1h"  # Correct
pup metrics query -query="..." -from="1h"    # Wrong (single dash)
```

### Missing Required Flags

**Symptoms:**
```
Error: required flag "query" not set
```

**Solutions:**
```bash
# Check required flags in help
pup metrics query --help

# Provide required flags
pup metrics query --query="avg:system.cpu.user{*}" --from="1h"
```

## Build Issues

### Compilation Errors

**Symptoms:**
```
# go build
./cmd/foo.go:123: undefined: SomeType
```

**Solutions:**
```bash
# Clean and rebuild
go clean
go mod tidy
go build -o pup .

# Update dependencies
go get -u github.com/DataDog/datadog-api-client-go/v2
go mod tidy
```

### Missing Dependencies

**Symptoms:**
```
go: missing go.sum entry for module
```

**Solutions:**
```bash
# Download missing dependencies
go mod download

# Regenerate go.sum
go mod tidy

# Verify module checksums
go mod verify
```

### Test Failures

**Symptoms:**
```
FAIL: TestSomething
```

**Solutions:**
```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v ./pkg/auth/... -run TestOAuthFlow

# Run with race detection
go test -race ./...

# Check test coverage
go test -cover ./...
```

## Output Issues

### JSON Parse Errors

**Symptoms:**
```
Error: invalid character '<' looking for beginning of value
```

**Causes:**
- HTML error response instead of JSON
- API returned non-JSON
- Corrupted response

**Solutions:**
```bash
# Check raw response
pup --verbose monitors list

# Try different output format
pup monitors list --output=yaml
```

### Table Formatting Issues

**Symptoms:**
- Columns misaligned
- Text truncated
- Wide output

**Solutions:**
```bash
# Use JSON for complete output
pup monitors list --output=json | jq .

# Specify custom fields
pup monitors list --fields="id,name,status"

# Use YAML for readability
pup monitors list --output=yaml
```

## Performance Issues

### Slow Commands

**Causes:**
- Large result sets
- Wide time ranges
- Network latency
- Datadog API slow response

**Solutions:**
```bash
# Use pagination
pup monitors list --limit=50

# Narrow time range
pup logs search --from="30m"  # Instead of 24h

# Filter results
pup monitors list --tag="env:prod"  # Instead of all
```

### High Memory Usage

**Causes:**
- Loading large result sets
- Not using pagination
- Processing too much data

**Solutions:**
```bash
# Use streaming/pagination
pup monitors list --limit=100

# Process in batches
for page in {0..10}; do
  pup monitors list --offset=$((page * 100)) --limit=100
done
```

## Debug Mode

Enable verbose logging to troubleshoot issues:

```bash
# Global verbose flag
pup --verbose <command>

# Set log level via env var
export PUP_LOG_LEVEL=debug
pup <command>

# Trace HTTP requests
export DD_DEBUG=true
pup --verbose <command>
```

**Verbose output includes:**
- HTTP request details
- API endpoint URLs
- Authentication method used
- Response status codes
- Error stack traces

## Configuration Issues

### Config File Not Loaded

**Check locations:**
```bash
# Default location
ls -la ~/.config/pup/config.yaml

# Custom location
pup --config=/path/to/config.yaml <command>

# Verify config syntax
cat ~/.config/pup/config.yaml | yq .
```

### Environment Variable Conflicts

**Precedence order:**
1. Command flags (highest)
2. Environment variables
3. Config file
4. Defaults (lowest)

**Debug config:**
```bash
# Show resolved config
pup --verbose auth status

# Check env vars
env | grep DD_
env | grep PUP_
```

## Getting Help

### Documentation

1. **Check command help:**
   ```bash
   pup --help
   pup metrics --help
   pup metrics query --help
   ```

2. **Read documentation:**
   - [README.md](../README.md)
   - [COMMANDS.md](COMMANDS.md)
   - [EXAMPLES.md](EXAMPLES.md)
   - [OAUTH2.md](OAUTH2.md)

3. **Check API docs:**
   - [Datadog API Reference](https://docs.datadoghq.com/api/latest/)

### Reporting Issues

When opening a GitHub issue, include:

1. **Pup version:**
   ```bash
   pup --version
   ```

2. **Command that failed:**
   ```bash
   pup --verbose <command>
   ```

3. **Environment info:**
   ```bash
   # OS version
   uname -a

   # Go version
   go version

   # Environment variables (redact keys!)
   env | grep DD_SITE
   ```

4. **Error message and stack trace**
5. **Steps to reproduce**
6. **Expected vs actual behavior**

### Community Support

- **GitHub Issues:** [github.com/datadog-labs/pup/issues](https://github.com/datadog-labs/pup/issues)
- **Datadog Community:** [community.datadoghq.com](https://community.datadoghq.com/)

## Common Workarounds

### Bypass SSL Verification (Not Recommended)

Only for testing with self-signed certs:
```bash
export DD_SKIP_SSL_VALIDATION=true
pup <command>
```

### Use Proxy

```bash
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
pup <command>
```

### Override API Endpoint

For testing or custom deployments:
```bash
export DD_HOST=https://custom-api.example.com
pup <command>
```
