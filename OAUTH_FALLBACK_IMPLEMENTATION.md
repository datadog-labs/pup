# OAuth Fallback Implementation

## Summary

Implemented automatic detection and fallback to API keys for Datadog API endpoints that don't support OAuth authentication. This ensures users get clear, actionable error messages and automatic fallback behavior when OAuth can't be used.

## Problem

Based on analysis of the `datadog-api-spec` repository (see `/Users/cody.lee/pup-oauth-analysis.csv`), 28 out of 132 pup commands (21%) use API endpoints that don't support OAuth authentication:

- **Logs API** (all endpoints) - missing `logs_read_data` scope
- **RUM API** (all endpoints) - missing `rum_apps_read/write` scopes
- **API/App Keys Management** - missing `api_keys_read/write` scopes

Previously, users with OAuth authentication would get generic API errors when trying to use these endpoints.

## Solution

### 1. Authentication Validator (`pkg/client/auth_validator.go`)

Created a new module that:
- Maintains a registry of endpoints that don't support OAuth
- Validates authentication type matches endpoint requirements
- Provides clear error messages with actionable steps

**Key Functions:**
```go
// Check if endpoint requires API keys
RequiresAPIKeyFallback(method, path string) bool

// Validate auth matches endpoint requirements
ValidateEndpointAuth(ctx, cfg, method, path string) error

// Get current authentication type
GetAuthType(ctx context.Context) AuthType
```

### 2. Client Updates (`pkg/client/client.go`)

Enhanced client creation with:
```go
// Force API key authentication
NewWithAPIKeys(cfg *config.Config) (*Client, error)

// Unified client creation with auth options
NewWithOptions(cfg *config.Config, forceAPIKeys bool) (*Client, error)
```

Added validation to `RawRequest()` method to check auth before making requests.

### 3. Command Layer (`cmd/root.go`)

Added smart client factory:
```go
// Creates appropriate client based on endpoint
getClientForEndpoint(method, path string) (*Client, error)
```

This function:
1. Checks if endpoint supports OAuth
2. If OAuth not supported and API keys available → uses API keys
3. If OAuth not supported and API keys missing → returns clear error
4. If OAuth supported → uses standard client (OAuth or API keys)

### 4. Updated Commands

Modified commands that use non-OAuth endpoints:

**Logs Commands:**
- `logs search` → uses `getClientForEndpoint("POST", "/api/v2/logs/events/search")`
- `logs list` → uses `getClientForEndpoint("POST", "/api/v2/logs/events")`
- `logs query` → uses `getClientForEndpoint("POST", "/api/v2/logs/events")`

**RUM Commands:**
- `rum apps list/get/create/update/delete` → uses `getClientForEndpoint()` with appropriate paths
- `rum metrics list/get` → uses `getClientForEndpoint()` with appropriate paths
- `rum retention-filters list/get` → uses `getClientForEndpoint()` with appropriate paths
- `rum sessions list/search` → uses `getClientForEndpoint()` with appropriate paths

**API Keys Commands:**
- `api-keys list/get/create/delete` → uses `getClientForEndpoint()` with appropriate paths

## Error Messages

### Before (with OAuth)
```
Error: failed to list logs: 401 Unauthorized
```

### After (with OAuth, no API keys)
```
Error: endpoint POST /api/v2/logs/events/search does not support OAuth authentication.
Please set DD_API_KEY and DD_APP_KEY environment variables.
Reason: Logs API missing OAuth implementation in spec
```

### After (with OAuth + API keys)
✅ **Automatically uses API keys** - no error, seamless fallback

## Testing

Comprehensive test coverage in `pkg/client/auth_validator_test.go`:

- ✅ `TestGetAuthType` - detects OAuth vs API keys vs none
- ✅ `TestRequiresAPIKeyFallback` - endpoint detection
- ✅ `TestValidateEndpointAuth` - validation logic
- ✅ `TestGetEndpointRequirement` - endpoint matching with IDs
- ✅ `TestGetAuthTypeDescription` - human-readable descriptions

All tests passing.

## User Experience

### Scenario 1: OAuth + API Keys Set
```bash
pup auth login  # OAuth authentication
export DD_API_KEY="..." DD_APP_KEY="..."
pup logs search --query="status:error" --from="1h"
# ✅ Works! Uses API keys automatically
```

### Scenario 2: OAuth Only (no API keys)
```bash
pup auth login  # OAuth authentication
pup logs search --query="status:error" --from="1h"
# ❌ Clear error: "endpoint does not support OAuth, please set DD_API_KEY and DD_APP_KEY"
```

### Scenario 3: API Keys Only
```bash
export DD_API_KEY="..." DD_APP_KEY="..."
pup logs search --query="status:error" --from="1h"
# ✅ Works! Uses API keys
```

### Scenario 4: OAuth-Supported Endpoint
```bash
pup auth login
pup monitors list
# ✅ Works! Uses OAuth token
```

## Endpoints Registry

The validator maintains a registry of 28 endpoint patterns that require API keys:

**Logs API (11 endpoints):**
- POST `/api/v2/logs/events/search`
- POST `/api/v2/logs/events`
- POST `/api/v2/logs/analytics/aggregate`
- GET `/api/v2/logs/config/archives*`
- GET `/api/v2/logs/config/custom_destinations*`
- GET `/api/v2/logs/config/metrics*`

**RUM API (10 endpoints):**
- GET/POST/PATCH/DELETE `/api/v2/rum/applications*`
- GET `/api/v2/rum/metrics*`
- GET `/api/v2/rum/retention_filters*`
- POST `/api/v2/rum/events/search`

**API/App Keys (7 endpoints):**
- GET/POST/DELETE `/api/v2/api_keys*`
- GET/POST/DELETE `/api/v2/app_keys*`

Pattern matching supports both exact paths and paths with IDs (e.g., `/api/v2/rum/applications/abc123`).

## Future Improvements

1. **Automatic Refresh**: When OAuth token expires and endpoint doesn't support OAuth, automatically try API keys
2. **Warning Messages**: Show warning when OAuth is used but endpoint doesn't support it (before fallback)
3. **Telemetry**: Track which endpoints are most affected by OAuth limitations
4. **Upstream**: Work with Datadog API team to add OAuth support to remaining endpoints

## References

- CSV Analysis: `/Users/cody.lee/pup-oauth-analysis.csv`
- API Spec Repo: `../datadog-api-spec`
- Implementation Branch: `feat/oauth-fallback-validation`
- Commit: `2aee04e`
