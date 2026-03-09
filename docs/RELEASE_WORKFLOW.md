# Release Workflow

This document describes the complete release process for stackctl.

## Quick Reference

```bash
# Check what's changed
make version

# Test build locally
make release-snapshot

# Prepare release
make changelog-prepare
# (Then manually edit CHANGELOG.md)

# Create and push release
git add CHANGELOG.md
git commit -m "chore: prepare release vX.Y.Z"
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push && git push --tags
```

---

## Overview

stackctl uses a **semi-automated release process**:

1. **Manual**: Update CHANGELOG.md with human-friendly release notes
2. **Automatic**: GitHub Actions builds binaries and publishes release
3. **Automatic**: GoReleaser generates additional notes from git commits

This gives us both:
- Rich, curated release notes (CHANGELOG.md)
- Automated binary distribution (GoReleaser + GitHub Actions)

---

## Release Types

### Semantic Versioning (SemVer)

We follow [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** (X.0.0): Breaking changes, incompatible API changes
- **MINOR** (0.X.0): New features, backward compatible
- **PATCH** (0.0.X): Bug fixes, backward compatible

### Pre-releases

Use suffixes for pre-releases:
- **Release Candidate**: `v0.1.0-rc1`, `v0.1.0-rc2`
- **Beta**: `v0.2.0-beta1`
- **Alpha**: `v0.3.0-alpha1`

Pre-releases are marked as "Pre-release" on GitHub.

---

## Step-by-Step Release Process

### 1. Verify Current State

```bash
# Check version status
make version

# Verify working tree is clean
git status

# Run tests
make test

# Test local build
make release-snapshot
```

**Expected Output:**
- Latest tag displayed
- Unreleased changes shown
- Commits since last tag listed
- Suggested next version

### 2. Prepare CHANGELOG.md

Run the helper script:

```bash
make changelog-prepare
```

This will:
- Prompt for version number (e.g., `0.1.0`)
- Update CHANGELOG.md with version and date
- Create a new `[Unreleased]` section

**Manual Steps:**
1. Review the generated changes in `CHANGELOG.md`
2. Move items from `Unreleased` to the new version section
3. Organize items by category:
   - **Added**: New features
   - **Changed**: Changes to existing functionality
   - **Deprecated**: Soon-to-be removed features
   - **Removed**: Removed features
   - **Fixed**: Bug fixes
   - **Security**: Security fixes

**Example:**

```markdown
## [Unreleased]

### Planned
- Package refactoring
- CLI feature parity

## [0.1.0] - 2026-03-10

### Added
- Binary distribution system with GoReleaser
- Embedded templates using `go:embed`
- New installer script for pre-built binaries

### Changed
- Migrated from external template files to embedded FS

### Fixed
- Template rendering issues with special characters
```

### 3. Commit CHANGELOG Changes

```bash
git add CHANGELOG.md
git commit -m "chore: prepare release v0.1.0"
```

**Commit Message Convention:**
- Use `chore: prepare release vX.Y.Z` format
- This commit will NOT be included in the release changelog (filtered by GoReleaser)

### 4. Create Git Tag

```bash
# Create annotated tag
git tag -a v0.1.0 -m "Release v0.1.0"

# Verify tag
git tag -l -n9 v0.1.0
```

**Important:**
- Always use annotated tags (`-a` flag)
- Tag message should be brief: "Release vX.Y.Z"
- GoReleaser requires annotated tags

### 5. Push Changes

```bash
# Push commit
git push

# Push tag (triggers release)
git push --tags
```

**What Happens Next:**
1. GitHub Actions detects the `v*` tag
2. Workflow `.github/workflows/release.yml` starts
3. GoReleaser builds binaries for linux/amd64 and linux/arm64
4. GitHub Release is created with:
   - Binaries as downloadable assets
   - Auto-generated changelog from commits
   - Link to CHANGELOG.md for full release notes
   - Installation instructions

### 6. Verify Release

After 5-10 minutes:

1. **Check GitHub Actions**: Ensure workflow succeeded
   ```
   https://github.com/imran110219/stackctl/actions
   ```

2. **Check GitHub Releases**: Verify release is published
   ```
   https://github.com/imran110219/stackctl/releases
   ```

3. **Test Installation**: On a clean VM
   ```bash
   curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/install.sh | bash
   stackctl version
   ```

### 7. Post-Release Tasks

After successful release:

```bash
# Update ROADMAP.md
# Mark completed tasks as ✅ Done

# Update MEMORY.md if needed
# Add release to "Release Tracking" section

# Commit updates
git add docs/ROADMAP.md
git commit -m "docs: update roadmap after v0.1.0 release"
git push
```

---

## Make Commands Reference

### Development

```bash
make build          # Build binary
make build-dev      # Build with dev version info
make install        # Install to /usr/local/bin
make test           # Run tests
make clean          # Remove build artifacts
```

### Release Management

```bash
make version              # Show current version status
make changelog-prepare    # Prepare CHANGELOG for release
make release-check        # Validate GoReleaser config
make release-snapshot     # Build local snapshot (no publish)
make release-dry-run      # Full dry run of release process
```

### Helpers

```bash
make tidy           # Run go mod tidy
make fmt            # Format code
make lint           # Run linters
make dev            # Format, build, install (quick iteration)
```

---

## Scripts Reference

All scripts are in `scripts/` directory:

### `suggest-version.sh`

Analyzes commits since last tag and suggests next version based on conventional commits.

**Usage:**
```bash
./scripts/suggest-version.sh
```

**Logic:**
- Breaking change commits → MAJOR bump
- `feat:` commits → MINOR bump
- `fix:` commits → PATCH bump

### `changelog-prepare.sh`

Prepares CHANGELOG.md for a new release by adding version section.

**Usage:**
```bash
./scripts/changelog-prepare.sh
# Prompts for version number
```

**Actions:**
- Creates backup: `CHANGELOG.md.bak`
- Adds new `[Unreleased]` section
- Creates `[VERSION]` section with date

### `release-dry-run.sh`

Comprehensive dry run of release process.

**Usage:**
```bash
./scripts/release-dry-run.sh
```

**Checks:**
- Working directory status
- CHANGELOG.md completeness
- GoReleaser config validation
- Builds snapshot release
- Lists all artifacts

---

## Troubleshooting

### Release Failed: "tag already exists"

```bash
# Delete local and remote tag
git tag -d v0.1.0
git push origin :refs/tags/v0.1.0

# Recreate tag
git tag -a v0.1.0 -m "Release v0.1.0"
git push --tags
```

### Release Failed: GoReleaser Error

```bash
# Check config
make release-check

# Test locally
make release-snapshot

# Check logs in GitHub Actions
```

### Wrong Version Released

1. **Delete the GitHub Release** (via web UI)
2. **Delete the tag** (see above)
3. **Fix CHANGELOG.md**
4. **Recreate tag and push**

### Need to Amend Release Notes

**Option 1**: Edit on GitHub
- Go to the release on GitHub
- Click "Edit release"
- Update description

**Option 2**: Re-release (for major issues)
- Follow "Wrong Version Released" steps
- Fix issues
- Re-release

---

## Best Practices

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add new feature
fix: resolve bug
docs: update documentation
chore: maintenance tasks
refactor: code refactoring
test: add tests
ci: CI/CD changes
```

Benefits:
- Auto-generated changelogs
- Semantic version suggestions
- Clear commit history

### CHANGELOG.md Maintenance

**During Development:**
- Add entries to `[Unreleased]` section as you work
- Don't wait until release time

**At Release Time:**
- Move unreleased items to versioned section
- Add date
- Review for completeness

**Tips:**
- Focus on **user-facing changes**
- Skip internal refactors unless significant
- Link to issues/PRs when relevant
- Use clear, non-technical language

### Testing Before Release

Minimum testing checklist:

- [ ] `make test` passes
- [ ] `make release-snapshot` builds successfully
- [ ] Manual smoke test of key features
- [ ] CHANGELOG.md is complete and accurate

For major releases, add:

- [ ] Test on Ubuntu 22.04 VM
- [ ] Test on Ubuntu 24.04 VM
- [ ] Test installer script
- [ ] Review all documentation

---

## GitHub Actions Workflow

The release is automated via `.github/workflows/release.yml`:

```yaml
name: Release
on:
  push:
    tags:
      - 'v*'
```

**Process:**
1. Checkout code
2. Set up Go
3. Set up GoReleaser
4. Run GoReleaser
5. Publish to GitHub Releases

**Environment Variables:**
- `GITHUB_TOKEN`: Automatically provided
- `GITHUB_REPOSITORY_OWNER`: Auto-detected
- `GITHUB_REPOSITORY_NAME`: Auto-detected

---

## Future Enhancements

Potential improvements to release process:

1. **Automated Testing on Release**
   - Spin up test VMs
   - Run installation tests
   - Report results

2. **Changelog Generation from Commits**
   - Use tools like `git-chglog`
   - Auto-populate CHANGELOG.md
   - Still require manual review

3. **Release Announcements**
   - Post to Discord/Slack
   - Tweet from project account
   - Update project website

4. **Homebrew Formula**
   - Auto-update Homebrew tap
   - Formula generation in GoReleaser

5. **Debian Package**
   - Build `.deb` files
   - Maintain APT repository

---

## References

- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [GoReleaser Documentation](https://goreleaser.com/)
- [GitHub Actions](https://docs.github.com/en/actions)

---

## Questions?

If you have questions about the release process:

1. Check this document
2. Review `docs/RELEASE.md` (technical guide)
3. Check existing releases for examples
4. Open an issue for clarification
