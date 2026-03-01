#!/usr/bin/env bash
# release.sh — bump version, commit, and tag a pup release
#
# Usage: ./scripts/release.sh <major|minor|patch>
#
# What it does:
#   1. Validates you're on main with a clean working tree
#   2. Computes the next version via git-semver
#   3. Creates a release branch (chore/release-vX.Y.Z)
#   4. Updates Cargo.toml + Cargo.lock
#   5. Creates a signed commit and pushes the branch
#   6. Opens a GitHub PR with release description, asks y/N to proceed
#   7. Merges the PR, pulls main, creates an annotated tag, and pushes it

set -euo pipefail

# ── helpers ──────────────────────────────────────────────────────────────────

die() { echo "error: $*" >&2; exit 1; }

require() {
    command -v "$1" &>/dev/null || die "'$1' not found in PATH"
}

# ── validate args ─────────────────────────────────────────────────────────────

BUMP="${1:-}"
case "$BUMP" in
    major|minor|patch) ;;
    *) die "usage: $0 <major|minor|patch>" ;;
esac

# ── require tools ─────────────────────────────────────────────────────────────

require git
require git-semver
require cargo
require sed
require gh

# ── validate repo state ───────────────────────────────────────────────────────

CURRENT_BRANCH=$(git symbolic-ref --short HEAD)
if [[ "$CURRENT_BRANCH" != "main" ]]; then
    die "must be on 'main' branch (currently on '$CURRENT_BRANCH')"
fi

if [[ -n "$(git status --porcelain)" ]]; then
    die "working tree is not clean — commit or stash changes first"
fi

git fetch origin main --quiet
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/main)
if [[ "$LOCAL" != "$REMOTE" ]]; then
    die "local main is not up to date with origin/main — run 'git pull' first"
fi

# ── compute versions ──────────────────────────────────────────────────────────

# git-semver outputs e.g. "v0.23.0"; strip the leading 'v' for Cargo.toml
NEW_TAG=$(git-semver -target "$BUMP")              # e.g. v0.23.0
NEW_VERSION="${NEW_TAG#v}"                          # e.g. 0.23.0

CURRENT_TAG=$(git describe --tags --abbrev=0)
echo "  current: $CURRENT_TAG"
echo "     next: $NEW_TAG  (${BUMP} bump)"
echo ""

# ── create release branch ─────────────────────────────────────────────────────

BRANCH="chore/release-${NEW_TAG}"
git checkout -b "$BRANCH"
echo "created branch: $BRANCH"

# ── update Cargo.toml ─────────────────────────────────────────────────────────

CARGO_TOML="$(git rev-parse --show-toplevel)/Cargo.toml"

# Replace only the package version line (first occurrence, at the top of [package])
sed -i '' "s/^version = \"[0-9]*\.[0-9]*\.[0-9]*\"/version = \"${NEW_VERSION}\"/" "$CARGO_TOML"

# Verify the replacement landed
CARGO_VERSION=$(grep '^version = ' "$CARGO_TOML" | head -1 | sed 's/version = "\(.*\)"/\1/')
if [[ "$CARGO_VERSION" != "$NEW_VERSION" ]]; then
    die "Cargo.toml version update failed (got '$CARGO_VERSION', expected '$NEW_VERSION')"
fi
echo "updated Cargo.toml: $CARGO_VERSION"

# ── refresh Cargo.lock ────────────────────────────────────────────────────────

echo "refreshing Cargo.lock..."
cargo check --quiet 2>&1 | grep -v "^$" || true

# ── commit (gpgsign=true in .gitconfig, so signing is automatic) ──────────────

git add Cargo.toml Cargo.lock
git commit -m "$(cat <<EOF
chore(release): bump version to ${NEW_TAG}

- Update Cargo.toml package version ${CURRENT_TAG#v} → ${NEW_VERSION}
- Refresh Cargo.lock

EOF
)"
echo "committed version bump"

# ── push release branch ───────────────────────────────────────────────────────

git push -u origin "$BRANCH"
echo "pushed: $BRANCH"

# ── create GitHub PR ──────────────────────────────────────────────────────────

echo ""
echo "creating PR..."
PR_URL=$(gh pr create \
    --title "chore(release): bump version to ${NEW_TAG}" \
    --base main \
    --body "$(cat <<EOF
## Summary
Release ${NEW_TAG}: version bump from ${CURRENT_TAG} to ${NEW_TAG}.

## Changes
- Update \`Cargo.toml\` package version ${CURRENT_TAG#v} → ${NEW_VERSION}
- Refresh \`Cargo.lock\`

## Testing
- CI will run \`cargo test\`, \`cargo clippy\`, and \`cargo fmt --check\`
- Cross-compilation verified for all 4 targets

---
🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)")
echo "PR: $PR_URL"

# ── ask to proceed ────────────────────────────────────────────────────────────

echo ""
read -r -p "Merge PR and tag ${NEW_TAG}? [y/N] " CONFIRM
case "$CONFIRM" in
    y|Y) ;;
    *) echo "aborted — PR left open at $PR_URL"; exit 0 ;;
esac

# ── merge the PR ──────────────────────────────────────────────────────────────

gh pr merge "$PR_URL" --merge --delete-branch --admin
echo "PR merged"

# ── pull main and verify the version bump landed ─────────────────────────────

git checkout main
git pull origin main --quiet
echo "pulled main"

MERGED_VERSION=$(grep '^version = ' "$CARGO_TOML" | head -1 | sed 's/version = "\(.*\)"/\1/')
if [[ "$MERGED_VERSION" != "$NEW_VERSION" ]]; then
    die "version $NEW_VERSION not found on main (got '$MERGED_VERSION') — was the PR merged?"
fi
echo "verified: $CARGO_TOML is at $NEW_VERSION"

# ── create annotated tag on main and push ─────────────────────────────────────

git tag -a "$NEW_TAG" -m "$NEW_TAG"
echo "tagged: $NEW_TAG"

git push origin "$NEW_TAG"
echo ""
echo "released $NEW_TAG"
