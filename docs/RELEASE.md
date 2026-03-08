# Release Process

This document explains how to create and publish releases for stackctl.

## Overview

Releases are automated using [GoReleaser](https://goreleaser.com/) and GitHub Actions. When you push a version tag, GitHub Actions automatically:

1. Builds binaries for Linux (amd64 and arm64)
2. Creates archives with documentation
3. Generates checksums
4. Creates a GitHub Release with changelog
5. Uploads all artifacts

## Prerequisites

- Write access to the repository
- Ability to push tags
- All tests passing on main branch

## Creating a Release

### 1. Ensure main branch is ready

```bash
# Make sure you're on main and up to date
git checkout main
git pull origin main

# Verify everything builds
go build -v ./cmd/stackctl

# Run tests (if available)
go test ./...

# Check git status is clean
git status
```

### 2. Create and push a version tag

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backwards compatible)
- **PATCH**: Bug fixes (backwards compatible)

```bash
# Create a new tag (e.g., v0.1.0, v1.0.0, v1.2.3)
git tag -a v0.1.0 -m "Release v0.1.0"

# Push the tag to trigger the release workflow
git push origin v0.1.0
```

### 3. Monitor the release

1. Go to the **Actions** tab in GitHub
2. Watch the "Release" workflow run
3. If successful, a new release will appear in the **Releases** section

### 4. Verify the release

Check that the release includes:
- ✅ Linux amd64 and arm64 binaries
- ✅ Archives (.tar.gz files)
- ✅ `checksums.txt` file
- ✅ Generated changelog
- ✅ Installation instructions in the release notes

## Testing Releases Locally

### Option 1: Use GoReleaser (Recommended)

Install GoReleaser:
```bash
# macOS
brew install goreleaser

# Linux
# See https://goreleaser.com/install/
```

Build without publishing:
```bash
# Snapshot build (no tag required)
goreleaser release --snapshot --clean

# Check the dist/ directory for artifacts
ls -lh dist/
```

### Option 2: Manual Build

Simulate what GoReleaser does:
```bash
# Build for Linux amd64
GOOS=linux GOARCH=amd64 go build \
  -trimpath \
  -ldflags="-s -w -X main.version=v0.1.0-test -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o dist/stackctl_linux_amd64 \
  ./cmd/stackctl

# Build for Linux arm64
GOOS=linux GOARCH=arm64 go build \
  -trimpath \
  -ldflags="-s -w -X main.version=v0.1.0-test -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o dist/stackctl_linux_arm64 \
  ./cmd/stackctl
```

## Pre-release Testing

Before creating an official release:

1. **Create a pre-release tag** (e.g., `v0.1.0-rc1`)
   ```bash
   git tag -a v0.1.0-rc1 -m "Release candidate 1"
   git push origin v0.1.0-rc1
   ```

2. **GitHub will automatically mark it as a pre-release**

3. **Test the binaries** on a clean Ubuntu VM:
   ```bash
   # Download the binary
   wget https://github.com/YOUR_ORG/stackctl/releases/download/v0.1.0-rc1/stackctl_v0.1.0-rc1_linux_amd64.tar.gz

   # Extract and test
   tar -xzf stackctl_*.tar.gz
   cd stackctl_*/
   ./stackctl --help
   ```

4. **If issues are found**, fix them and create a new RC tag (e.g., `v0.1.0-rc2`)

5. **When ready**, create the final release tag (e.g., `v0.1.0`)

## Troubleshooting

### Build fails with "templates not found"

The `templates_embed.go` file must be present with the `//go:embed all:templates` directive. Verify:
```bash
grep "go:embed all:templates" templates_embed.go
```

### GoReleaser fails on GitHub Actions

Check the workflow logs in the Actions tab. Common issues:
- Missing `GITHUB_TOKEN` (should be automatic)
- Invalid `.goreleaser.yml` syntax
- Go version mismatch

### Binary doesn't include templates

Verify the build includes embedded files:
```bash
# Check binary size (should be ~7-8MB with templates)
ls -lh stackctl

# Test that it works without external templates
./stackctl --help
```

## Version Variables

The following variables are injected at build time by GoReleaser:

- `version`: Git tag (e.g., `v0.1.0`)
- `commit`: Git commit hash
- `date`: Build timestamp (RFC3339 format)
- `builtBy`: Always `goreleaser` for official releases

For development builds, these default to:
- `version`: `dev`
- `commit`: `none`
- `date`: `unknown`
- `builtBy`: `manual`

## Next Steps

After successful releases:

1. Update the `install.sh` script to download from GitHub Releases (Priority 1, Task 4)
2. Test installation on clean Ubuntu 22.04 and 24.04 VMs (Priority 1, Task 5)
3. Update README.md and USER_GUIDE.md with new installation instructions (Priority 1, Task 6)

## References

- [GoReleaser Documentation](https://goreleaser.com/)
- [Semantic Versioning](https://semver.org/)
- [GitHub Actions: Creating Releases](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases)
