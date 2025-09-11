# Changelog

## [v1.4.0] - 2025-09-11

### Added
- feat: Add Docker container support and update documentation
- feat: Add usage information to command line arguments for LogGuardian container
- feat: Enhance CI workflow with Docker build and test steps; add pre-commit script for local checks
- feat: Implement ServiceAdapter with retry logic and rate limiting; add tests for service adapter functionality
- Add dry-run compliance service and related tests

### Fixed
- fix: Update AWS region for dry-run container execution
- fix: Improve batch size parsing and handle errors gracefully

### Changed
- docs: Update containerization design document to reflect service adapter implementation details and authentication strategies
- docs: Update containerization design document with implementation status and decision dates
- docs: Update upgrade guide with detailed CLI change set review steps and parameter preservation instructions
- docs: Enhance CLI update instructions with change set review process
- docs: Update Terraform deployment examples and add warning for production validation

