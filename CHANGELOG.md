# Changelog

All notable changes to stackctl will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Package refactoring (split internal/stackctl into layers)
- CLI feature parity (modules list/info, logs, restart, exec, destroy)
- Installation testing on clean Ubuntu VMs

## [0.1.0-rc1] - 2026-03-09

### Added
- Binary distribution system with GoReleaser
- Embedded templates using `go:embed` (no external files needed)
- GitHub Actions workflow for automated releases
- New `install.sh` script for downloading pre-built binaries
- Support for Linux amd64 and arm64 architectures
- `stackctl version` command with build info

### Changed
- Migrated from external template files to embedded FS
- Refactored template rendering to support `fs.FS`
- Updated `findTemplatesDir()` to use embedded templates with `STACKCTL_TEMPLATES` override for development

### Infrastructure
- Created `.goreleaser.yml` configuration
- Added `.github/workflows/release.yml` for automated builds
- Created `docs/RELEASE.md` with release instructions
- Created `docs/INSTALL_TESTING.md` with testing procedures

### Known Issues
- Installation not yet tested on clean Ubuntu VMs (blocked until release)
- Documentation (README.md, USER_GUIDE.md) needs updates post-testing

### Release Assets
- Binary: `stackctl_linux_amd64`, `stackctl_linux_arm64`
- Installer: `curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/install.sh | bash`
- GitHub Release: https://github.com/imran110219/stackctl/releases/tag/v0.1.0-rc1

---

## Release History

<!-- Keep this section updated after each release -->
- **v0.1.0-rc1** (2026-03-09): First public release candidate with binary distribution
- **Next**: v0.1.0 after VM testing and documentation updates
