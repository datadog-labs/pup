# Homebrew Tap Setup Guide

This guide documents the setup required to enable automatic Homebrew formula publishing to the `DataDog/homebrew-pack` tap using dd-octo-sts for secure, short-lived token access.

## Overview

When a new release is tagged, the release workflow will:
1. Request an OIDC token from GitHub Actions
2. Exchange it with dd-octo-sts for a scoped, short-lived GitHub token
3. Use GoReleaser to build binaries and push formula to `homebrew-pack`
4. Token automatically expires after 1 hour and is revoked after the workflow completes

Users can then install via: `brew install datadog/pack/pup`

## Why dd-octo-sts?

**Security advantages over Personal Access Tokens (PATs):**
- ✅ **Short-lived tokens**: 1-hour expiration, auto-revoked after CI run
- ✅ **No credential storage**: No long-lived secrets to manage or rotate
- ✅ **Scoped access**: Limited to specific workflow, tags, and repository
- ✅ **Least privilege**: Only grants `contents: write` on `homebrew-pack`
- ✅ **Audit trail**: All token exchanges logged and traceable

## Prerequisites

### 1. Repository Setup

The `DataDog/homebrew-pack` repository must:
- ✅ Exist at https://github.com/DataDog/homebrew-pack
- ✅ Be public (or have appropriate access configured)
- ✅ Have the trust policy merged to the default branch
- ✅ Follow Homebrew tap naming conventions (`homebrew-*` prefix)

### 2. Tag Protection (Recommended)

**Strongly recommended** for security best practices:
- ✅ Protect version tags (`v*.*.*`) to prevent unauthorized releases
- ✅ Aligns with dd-octo-sts security guardrails (privileged permissions on protected refs)
- ✅ Ensures only authorized users can create releases

See [Step 3: Protect Release Tags](#step-3-protect-release-tags-recommended) below for setup instructions.

## Setup Instructions

### Step 1: Add Trust Policy to homebrew-pack

1. Clone the `homebrew-pack` repository:
   ```bash
   git clone https://github.com/DataDog/homebrew-pack.git
   cd homebrew-pack
   ```

2. Create the trust policy directory if it doesn't exist:
   ```bash
   mkdir -p .github/chainguard
   ```

3. Create `.github/chainguard/pup-release.sts.yaml` with the following content:
   ```yaml
   # Trust policy for pup release workflow to push Homebrew formulas
   issuer: https://token.actions.githubusercontent.com

   # Allow releases from semantic version tags (v1.2.3, v0.1.0, etc.)
   subject_pattern: repo:DataDog/pup:ref:refs/tags/v[0-9]+\.[0-9]+\.[0-9]+

   # Defense-in-depth: additional claim validation
   claim_pattern:
     repository: DataDog/pup
     ref: refs/tags/v[0-9]+\.[0-9]+\.[0-9]+
     ref_type: tag
     event_name: push
     job_workflow_ref: DataDog/pup/\.github/workflows/release\.yml@refs/tags/v[0-9]+\.[0-9]+\.[0-9]+

   # Grant write access to push formula updates
   permissions:
     contents: write
   ```

   **Note**: A copy of this policy is available at `docs/homebrew-pack-trust-policy.yaml` in the pup repository.

4. Commit and create a pull request:
   ```bash
   git checkout -b add-pup-release-policy
   git add .github/chainguard/pup-release.sts.yaml
   git commit -m "feat(policy): add trust policy for pup release workflow"
   git push -u origin add-pup-release-policy
   gh pr create --title "feat(policy): add trust policy for pup release workflow" \
     --body "Adds dd-octo-sts trust policy to allow pup release workflow to push formula updates securely using short-lived tokens."
   ```

5. **Wait for Trust Policy Validation check to pass** (automated check by dd-octo-sts)

6. **Merge the PR to the default branch** (policy must be on default branch to work)

### Step 2: Verify pup Workflow Configuration

The `pup` repository workflow is already configured (see `.github/workflows/release.yml`):

```yaml
- name: Get Homebrew tap token via dd-octo-sts
  uses: DataDog/dd-octo-sts-action@acaa02eee7e3bb0839e4272dacb37b8f3b58ba80 # v1.0.3
  id: octo-sts
  with:
    scope: DataDog/homebrew-pack
    policy: pup-release

- name: Run GoReleaser
  uses: goreleaser/goreleaser-action@v6
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    HOMEBREW_TAP_TOKEN: ${{ steps.octo-sts.outputs.token }}
```

**No GitHub secrets required!** The workflow uses OIDC federation automatically.

### Step 3: Protect Release Tags (Recommended)

**Why protect tags?**
- Prevents unauthorized release creation
- Aligns with dd-octo-sts best practice: privileged permissions on protected refs
- Ensures only designated users/teams can trigger releases

**How to set up tag protection:**

1. Go to: https://github.com/DataDog/pup/settings/rules/new

2. Create a **Tag ruleset**:
   - **Ruleset name**: `Protect Release Tags`
   - **Enforcement status**: Active
   - **Target**: Tags
   - **Tag name pattern**: `v[0-9]*.[0-9]*.[0-9]*`

3. **Configure protections**:
   - ✅ **Restrict creations**: Check this box
     - Add authorized users/teams who can create releases
     - Example: Add release engineers, maintainers team
   - ✅ **Restrict deletions**: Check this box (prevent accidental deletion)
   - ✅ **Require signed commits**: Optional but recommended

4. **Add bypass list** (users/teams who can create releases):
   - Add your GitHub username
   - Add release maintainers
   - Consider creating a dedicated "releases" team

5. **Save the ruleset**

**Alternative: Protected Environments** (simpler but less granular)

If tag rulesets are too complex, use a protected environment:

1. Go to: https://github.com/DataDog/pup/settings/environments/new
2. Create environment named `release`
3. Add required reviewers
4. Update workflow to use environment:
   ```yaml
   jobs:
     goreleaser:
       runs-on: ubuntu-latest
       environment: release  # Add this line
   ```

**Note**: The setup works without tag protection, but it's strongly recommended for production releases.

## Verification Checklist

Before creating your first release with Homebrew tap publishing:

- [ ] `DataDog/homebrew-pack` repository exists and is public
- [ ] Trust policy merged to default branch in `homebrew-pack`
- [ ] Trust Policy Validation check passed on the policy PR
- [ ] Release workflow in `pup` includes dd-octo-sts-action step
- [ ] GoReleaser config includes `brews` section (`.goreleaser.yml`)
- [ ] (Recommended) Release tags protected via repository rulesets

## Testing the Setup

### Test with a Pre-release

To test without affecting production:

1. **Ensure the trust policy is merged to the default branch** in `homebrew-pack`

2. Create a pre-release tag in the `pup` repository:
   ```bash
   cd /path/to/pup
   git tag -a v0.9.0-beta.1 -m "Test release for Homebrew tap"
   git push origin v0.9.0-beta.1
   ```

   **Note**: If you protected tags, ensure you have permission to create them.

3. Monitor the GitHub Actions workflow:
   - Go to: https://github.com/DataDog/pup/actions
   - Check the "Release" workflow run
   - Verify the "Get Homebrew tap token via dd-octo-sts" step succeeds
   - Verify GoReleaser successfully pushes to `homebrew-pack`

4. Verify the formula was created:
   - Check: https://github.com/DataDog/homebrew-pack/blob/main/Formula/pup.rb
   - The formula should be auto-generated with version `0.9.0-beta.1`

5. Test installation (optional):
   ```bash
   brew tap datadog/pack
   brew install pup
   pup version
   ```

6. Clean up the test release if needed:
   ```bash
   git tag -d v0.9.0-beta.1
   git push origin :refs/tags/v0.9.0-beta.1
   # Manually delete the GitHub release if created
   ```

## First Production Release

Once testing is successful:

1. Create a production release tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. The workflow will:
   - Get short-lived token via dd-octo-sts
   - Build binaries for all platforms
   - Create GitHub release with artifacts
   - Push `pup.rb` formula to `DataDog/homebrew-pack`

3. Users can install via:
   ```bash
   brew tap datadog/pack
   brew install pup
   ```

## Troubleshooting

### Error: "failed to exchange OIDC token"

**Cause**: Trust policy not found or not on default branch.

**Solution**:
1. Verify policy exists at `.github/chainguard/pup-release.sts.yaml` in `homebrew-pack`
2. Ensure it's merged to the default branch (usually `main`)
3. Check the policy filename exactly matches (`.sts.yaml` not `.sts.yml`)

### Error: "OIDC token validation failed"

**Cause**: Claims in OIDC token don't match trust policy patterns.

**Solution**:
1. Check the dd-octo-sts-action step logs - it prints claims on failure
2. Verify the tag matches pattern: `v[0-9]+\.[0-9]+\.[0-9]+` (e.g., `v1.2.3`)
3. Ensure workflow file is `.github/workflows/release.yml` (not renamed)
4. Compare printed claims against `claim_pattern` in trust policy

### Error: "insufficient permissions" or "permission denied"

**Cause**: Trust policy doesn't grant required permissions or tags are protected.

**Solution**:
1. Verify trust policy includes: `permissions: { contents: write }`
2. Ensure policy is on the default branch
3. If tags are protected, verify you're in the bypass list
4. Wait a few minutes after merging - policy cache may need to refresh

### Error: "workflow not found" or "policy not found"

**Cause**: Incorrect scope or policy name in workflow.

**Solution**:
1. Verify `scope: DataDog/homebrew-pack` in workflow
2. Verify `policy: pup-release` matches filename (without `.sts.yaml`)
3. Check for typos in repository owner/name

### Error: "resource not accessible by integration" when pushing tags

**Cause**: Tag protection enabled but you're not in the bypass list.

**Solution**:
1. Go to: https://github.com/DataDog/pup/settings/rules
2. Find the "Protect Release Tags" ruleset
3. Add your username to the bypass list
4. Or temporarily disable the ruleset for testing

### Formula not updating with new versions

**Cause**: Previous step might have failed silently.

**Solution**:
1. Check full workflow logs for dd-octo-sts and GoReleaser steps
2. Ensure the tag matches the version pattern (e.g., `v1.2.3`)
3. Verify GoReleaser config includes `brews` section
4. Check if `homebrew-pack` repository has any branch protection rules blocking pushes

## Policy Maintenance

### Updating the Trust Policy

To modify the trust policy (e.g., change permissions, add constraints):

1. Edit `.github/chainguard/pup-release.sts.yaml` in `homebrew-pack`
2. Create a PR with the changes
3. Wait for Trust Policy Validation check to pass
4. Merge to default branch
5. Changes take effect immediately for new workflow runs

### Common Policy Updates

**Allow pre-release tags** (e.g., `v1.0.0-beta.1`):
```yaml
subject_pattern: repo:DataDog/pup:ref:refs/tags/v[0-9]+\.[0-9]+\.[0-9]+(-[a-z]+\.[0-9]+)?
claim_pattern:
  ref: refs/tags/v[0-9]+\.[0-9]+\.[0-9]+(-[a-z]+\.[0-9]+)?
```

**Restrict to specific major versions**:
```yaml
subject_pattern: repo:DataDog/pup:ref:refs/tags/v1\.[0-9]+\.[0-9]+
```

**Add workflow approval via protected environment**:
```yaml
claim_pattern:
  environment: release  # Add this line
```

## Security Notes

- ✅ **No secrets to rotate**: Tokens are short-lived and auto-revoked
- ✅ **Scoped access**: Policy restricts to specific workflow and tags
- ✅ **Audit trail**: All token exchanges logged in dd-octo-sts service
- ✅ **Defense in depth**: Multiple claim validations (`subject_pattern` + `claim_pattern`)
- ✅ **Protected refs recommended**: Tag protection adds extra security layer

### Security Best Practices

1. **Protect version tags**: Use repository rulesets to control who can create releases
2. **Keep patterns specific**: Avoid overly broad regex patterns in trust policies
3. **Review policy changes**: Always review trust policy PRs carefully
4. **Monitor workflow runs**: Check dd-octo-sts step logs for anomalies
5. **Least privilege**: Only grant minimum required permissions
6. **Regular audits**: Periodically review who has release tag creation permissions

## References

- [dd-octo-sts User Guide (Confluence)](https://datadoghq.atlassian.net/wiki/spaces/SECENG/pages/5138645099)
- [dd-octo-sts GitHub Action](https://github.com/DataDog/dd-octo-sts-action)
- [GitHub Tag Protection Rulesets](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/managing-repository-settings/managing-rulesets-for-a-repository)
- [GoReleaser Homebrew Tap Documentation](https://goreleaser.com/customization/homebrew/)
- [Homebrew Tap Creation Guide](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- Slack: [#sdlc-security](https://dd.enterprise.slack.com/archives/C027P1CK07N)

## Comparison: dd-octo-sts vs PAT

| Feature | dd-octo-sts | Personal Access Token |
|---------|-------------|----------------------|
| Token lifetime | 1 hour, auto-revoked | Indefinite until manually revoked |
| Secret storage | None (OIDC federation) | Requires GitHub secret |
| Scope | Workflow + tag specific | Broad access |
| Rotation | Automatic | Manual |
| Audit trail | Complete | Limited |
| Setup complexity | Medium (trust policy + tag protection) | Simple (create + add secret) |
| Security posture | ✅ Excellent | ⚠️ Acceptable |
| Recommended for | Datadog repos | External/simple use cases |
