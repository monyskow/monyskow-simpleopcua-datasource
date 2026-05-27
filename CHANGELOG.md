# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2026-05-27

### Added

- `npm run server:dev` script for quick local development against a single Grafana version + auth combo (#67)

### Fixed

- Client certificate `KeyUsage` now includes `ContentCommitment` (nonRepudiation) per OPC-UA Part 6 §6.2.2. Strict-compliant servers (e.g. ProSys Simulation Server) reject certs without this bit during `OpenSecureChannel`, which surfaced as `connect: connect: EOF` (#69)
- README links in the Documentation table now use absolute GitHub URLs so they resolve when the README is rendered on the Grafana plugin catalog (#76)

### Changed

- Documentation reorganized: lean `README.md` with link table, topical files moved to `docs/installation.md`, `docs/configuration.md`, `docs/usage.md`, `docs/development.md`, `docs/testing.md`, `docs/troubleshooting.md`. `CONTRIBUTING.md` is now a redirect stub (#74)
- Repo hygiene: untracked ~100MB of accidentally-committed test artifacts and release zips, tightened `.gitignore`, archived completed plan files (#72)

### Internal

- CI: pinned `golangci-lint` to v2.12.2 and bumped `golangci-lint-action` to v7 for Go 1.25 support (#70, #71)

## [1.0.0] - 2026-01-30

### Added

- Initial release
- OPC-UA Data Access (DA) support
- Node browsing with tree view
- Anonymous authentication
- Username/Password authentication
- Certificate authentication
- Security policy and mode configuration
- Connection timeout configuration
- Template variable support
- Health check implementation
- Certificate generation UI with "Generate Certificate" button for secure connections
- Client certificates stored in Grafana's encrypted storage (secureJsonData)
- Certificates persist across Grafana restarts and datasource updates
- Certificate status indicator showing configured/not configured state
- Support for Sign and SignAndEncrypt security modes with auto-generated certificates
- GitHub Sponsors link for plugin support

### Changed

- Security modes now support certificate auto-generation from the UI
- Certificates are automatically cached during the plugin session for performance
- Certificate lifecycle management - valid for 3 years with automatic renewal when expiring soon
