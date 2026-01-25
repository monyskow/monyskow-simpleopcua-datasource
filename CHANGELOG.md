# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Certificate generation UI with "Generate Certificate" button for secure connections
- Client certificates are now stored in Grafana's encrypted storage (secureJsonData)
- Certificates persist across Grafana restarts and datasource updates
- Certificate status indicator showing configured/not configured state
- Support for Sign and SignAndEncrypt security modes with auto-generated certificates
- GitHub Sponsors link for plugin support

### Changed

- Security modes now support certificate auto-generation from the UI
- Certificates are automatically cached during the plugin session for performance
- Certificate lifecycle management - valid for 3 years with automatic renewal when expiring soon

## [1.0.0] - 2024-12-25

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
