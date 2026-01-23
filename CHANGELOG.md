# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Auto-generate self-signed client certificates for secure OPC-UA connections
- Automatic certificate storage and reuse in plugin data directory
- Server certificate auto-trust for simplified secure connections
- Support for Sign and SignAndEncrypt security modes without manual certificate configuration

### Changed

- Security modes now work out-of-the-box without requiring user-provided certificates
- Certificates are automatically regenerated 30 days before expiration

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
