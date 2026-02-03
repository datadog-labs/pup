# Changelog

All notable changes to Pup will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- OAuth2 authentication with PKCE support
- Keychain storage for secure token management (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- Multi-architecture release builds (Linux, macOS, Windows on amd64/arm64)
- SBOM (Software Bill of Materials) generation
- Code signing with cosign
- Commands: `auth`, `monitors`, `dashboards`, `slos`, `incidents`
- Automatic fallback from OAuth2 to API key authentication
- Comprehensive LLM-friendly help text

### Changed
- Project renamed from "fetch" to "pup"
- Configuration directory moved from `~/.config/fetch` to `~/.config/pup`

### Security
- OAuth2 PKCE flow for secure authentication
- Secure token storage in OS keychain
- Signed release artifacts

## [0.1.0] - Initial Development

### Added
- Initial Go-based CLI wrapper for Datadog APIs
- Basic commands for monitors, dashboards, SLOs, and incidents
- API key authentication support
