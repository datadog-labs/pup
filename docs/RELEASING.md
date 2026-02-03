# Release Process

This document describes how to create a new release of Pup.

## Prerequisites

- Push access to the repository
- Git configured with your name and email
- GoReleaser installed locally (for testing): `brew install goreleaser`

## Release Workflow

The release process is fully automated via GitHub Actions and GoReleaser.

### 1. Update Version Information

Update the CHANGELOG.md file with the new version:

```bash
# Edit CHANGELOG.md
vim CHANGELOG.md
```

Move changes from `[Unreleased]` to a new version section:

```markdown
## [1.0.0] - 2024-02-03

### Added
- New feature description

### Changed
- Change description

### Fixed
- Bug fix description
```

Commit the changelog:

```bash
git add CHANGELOG.md
git commit -m "docs: update changelog for v1.0.0"
git push origin main
```

### 2. Create and Push a Tag

Tags must follow semantic versioning with a `v` prefix (e.g., `v1.0.0`).

```bash
# Create an annotated tag
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag to trigger the release workflow
git push origin v1.0.0
```

### 3. Automated Release Process

Once the tag is pushed, GitHub Actions will automatically:

1. **Run Tests**: Execute `go test ./...` to ensure everything passes
2. **Build Binaries**: Cross-compile for multiple platforms:
   - Linux: amd64, arm64, arm (v7)
   - macOS (Darwin): amd64, arm64
   - Windows: amd64
3. **Create Archives**: Generate tar.gz (Linux/macOS) and zip (Windows) archives
4. **Generate SBOM**: Create Software Bill of Materials for each archive
5. **Sign Artifacts**: Sign all artifacts with cosign (keyless signing)
6. **Create Checksums**: Generate SHA256 checksums for all artifacts
7. **Create Release**: Publish a GitHub release with all artifacts

### 4. Release Artifacts

Each release includes:

#### Binaries
- `pup_1.0.0_Linux_x86_64.tar.gz`
- `pup_1.0.0_Linux_arm64.tar.gz`
- `pup_1.0.0_Linux_armv7.tar.gz`
- `pup_1.0.0_Darwin_x86_64.tar.gz`
- `pup_1.0.0_Darwin_arm64.tar.gz`
- `pup_1.0.0_Windows_x86_64.zip`

#### Source
- `pup_1.0.0_source.tar.gz` - Full source code archive

#### Security
- `pup_1.0.0_checksums.txt` - SHA256 checksums
- `*.sbom.json` - Software Bill of Materials for each archive
- `*.sig` - Cosign signatures for each artifact
- `*.pem` - Cosign certificates for verification

### 5. Verify the Release

After the workflow completes (usually 5-10 minutes):

1. Visit https://github.com/DataDog/pup/releases
2. Verify the release is published with all artifacts
3. Check that signatures and SBOMs are present
4. Test downloading and verifying an artifact:

```bash
# Download a release
curl -LO https://github.com/DataDog/pup/releases/download/v1.0.0/pup_1.0.0_Linux_x86_64.tar.gz
curl -LO https://github.com/DataDog/pup/releases/download/v1.0.0/pup_1.0.0_Linux_x86_64.tar.gz.sig
curl -LO https://github.com/DataDog/pup/releases/download/v1.0.0/pup_1.0.0_Linux_x86_64.tar.gz.pem

# Verify signature with cosign
cosign verify-blob \
  --certificate pup_1.0.0_Linux_x86_64.tar.gz.pem \
  --signature pup_1.0.0_Linux_x86_64.tar.gz.sig \
  pup_1.0.0_Linux_x86_64.tar.gz

# Extract and test
tar xzf pup_1.0.0_Linux_x86_64.tar.gz
./pup --version
```

## Testing Releases Locally

Before pushing a tag, you can test the release process locally:

```bash
# Install goreleaser
brew install goreleaser

# Test without publishing
goreleaser release --snapshot --clean

# Check the dist/ directory
ls -lh dist/
```

## Versioning Guidelines

Follow [Semantic Versioning](https://semver.org/):

- **Major version (v1.0.0 → v2.0.0)**: Breaking changes
- **Minor version (v1.0.0 → v1.1.0)**: New features, backwards compatible
- **Patch version (v1.0.0 → v1.0.1)**: Bug fixes, backwards compatible

### Pre-releases

For testing, you can create pre-release versions:

```bash
git tag -a v1.0.0-rc.1 -m "Release candidate 1"
git push origin v1.0.0-rc.1
```

GoReleaser will automatically mark these as pre-releases on GitHub.

## Rollback

If a release has issues:

1. Delete the GitHub release (this does NOT delete the tag)
2. Delete the local tag: `git tag -d v1.0.0`
3. Delete the remote tag: `git push origin :refs/tags/v1.0.0`
4. Fix the issues
5. Create a new patch release with fixes

## Troubleshooting

### Release workflow fails

1. Check the GitHub Actions logs: https://github.com/DataDog/pup/actions
2. Common issues:
   - **Tests failing**: Fix tests before releasing
   - **Build errors**: Ensure `go build` works locally
   - **Permission errors**: Check repository settings > Actions > General > Workflow permissions

### Signature verification fails

Ensure you're using the correct certificate and signature files from the release.

### Missing artifacts

Check that `.goreleaser.yml` includes all desired platforms and architectures.

## Support

For issues with the release process:
- GitHub Issues: https://github.com/DataDog/pup/issues
- Internal Slack: #datadog-pup (if available)
