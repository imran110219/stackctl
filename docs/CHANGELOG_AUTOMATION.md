# Changelog & Release Automation

This document provides a quick overview of the changelog automation system.

## Quick Start

```bash
# 1. Check current status
make version

# 2. Prepare for release
make changelog-prepare
# Enter version when prompted (e.g., 0.1.0)

# 3. Edit CHANGELOG.md manually
# Move unreleased items to the new version section

# 4. Commit and tag
git add CHANGELOG.md
git commit -m "chore: prepare release v0.1.0"
git tag -a v0.1.0 -m "Release v0.1.0"

# 5. Push (triggers automated release)
git push && git push --tags
```

## What Gets Automated?

### ✅ Automated
- Binary building (GoReleaser)
- GitHub Release creation
- Asset packaging (binaries, checksums)
- Release notes from git commits
- Installation script updates

### 📝 Manual (You Control)
- CHANGELOG.md content (rich, human-friendly)
- Version numbers
- Release timing
- What goes in each release

## Why This Approach?

| Aspect | Fully Automated | Our Hybrid Approach |
|--------|-----------------|---------------------|
| **Release notes** | Generic commit list | Curated, user-focused |
| **Context** | Just code changes | Why changes matter |
| **Known issues** | ❌ Not tracked | ✅ Documented |
| **AI awareness** | ❌ Not preserved | ✅ Full context in repo |
| **Speed** | ✅ Fastest | ⚡ Fast (1 min manual) |
| **Quality** | ⚠️ Variable | ✅ Consistent |

## Files Created

### Core Files
- **`CHANGELOG.md`** - Release history with rich context
- **`Makefile`** - 20+ development and release commands
- **`.goreleaser.yml`** - Enhanced with CHANGELOG link

### Scripts (`scripts/`)
- **`suggest-version.sh`** - Analyzes commits, suggests next version
- **`changelog-prepare.sh`** - Prepares CHANGELOG for release
- **`changelog-add.sh`** - Interactive helper for adding entries
- **`release-dry-run.sh`** - Full release simulation

### Documentation (`docs/`)
- **`RELEASE_WORKFLOW.md`** - Complete release guide
- **`CHANGELOG_AUTOMATION.md`** - This file
- **`.github/RELEASE_CHECKLIST.md`** - Quick checklist

### Updated Files
- **`MEMORY.md`** - Now references CHANGELOG.md
- **`ROADMAP.md`** - Marked v0.1.0-rc1 as released

## Make Commands

### Essential
```bash
make version            # What's changed? What's next?
make release-snapshot   # Test build locally
make changelog-prepare  # Prepare for release
make release-dry-run    # Full simulation
```

### Development
```bash
make build             # Build binary
make test              # Run tests
make dev               # Format, build, install (quick!)
```

### Helpers
```bash
make help              # Show all commands
make list              # List all targets
```

## Typical Development Workflow

### During Development

As you work, add entries to CHANGELOG.md:

```markdown
## [Unreleased]

### Planned
- Package refactoring
- CLI feature parity

### Added
- New feature X
- New command Y

### Fixed
- Bug in Z
```

### When Ready to Release

1. **Check status**: `make version`
2. **Test build**: `make release-dry-run`
3. **Prepare changelog**: `make changelog-prepare`
4. **Edit CHANGELOG.md**: Move unreleased items to version section
5. **Commit**: `git add CHANGELOG.md && git commit -m "chore: prepare release vX.Y.Z"`
6. **Tag**: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
7. **Push**: `git push && git push --tags`
8. **Wait**: GitHub Actions builds and publishes (5-10 minutes)
9. **Verify**: Check GitHub Releases
10. **Test**: Install on clean VM and test

### After Release

Update roadmap and memory:

```bash
# Mark completed tasks
vim docs/ROADMAP.md

# Commit updates
git add docs/ROADMAP.md
git commit -m "docs: update roadmap after vX.Y.Z"
git push
```

## For AI Assistants (Claude)

When you start a new conversation:

1. **Read `CHANGELOG.md`** to understand:
   - What versions were released
   - When they were released
   - What features were added
   - What's planned next

2. **Read `docs/ROADMAP.md`** for:
   - Current sprint priorities
   - Long-term vision

3. **Read `MEMORY.md`** for:
   - Project architecture
   - Key decisions

This gives you full context immediately, with no network calls needed.

## Troubleshooting

### "make: command not found"
Install make: `sudo apt install build-essential`

### "goreleaser: command not found"
Not needed for development. Only CI/CD needs it.
For local testing: https://goreleaser.com/install/

### "scripts/suggest-version.sh: Permission denied"
Fix permissions: `chmod +x scripts/*.sh`

### Want to undo changelog-prepare?
Restore backup: `mv CHANGELOG.md.bak CHANGELOG.md`

## Benefits of This System

### For Developers
- ✅ Fast: 1 minute to prepare release
- ✅ Automated: Binary builds happen automatically
- ✅ Tested: Dry-run before real release
- ✅ Reversible: Can undo mistakes

### For Users
- ✅ Clear: Rich, curated release notes
- ✅ Trustworthy: Checksums and signatures
- ✅ Easy: One-line installer

### For AI Assistants
- ✅ Fast: No network calls needed
- ✅ Rich: Full context, not just version numbers
- ✅ Reliable: Version controlled, always accurate
- ✅ Future-proof: Works offline, no API limits

## Next Steps

1. **Familiarize yourself**: Run `make version` and `make help`
2. **Test it out**: Try `make release-snapshot`
3. **Read the guide**: See `docs/RELEASE_WORKFLOW.md`
4. **Practice**: Do a dry run with `make release-dry-run`

## References

- [RELEASE_WORKFLOW.md](./RELEASE_WORKFLOW.md) - Complete release guide
- [ROADMAP.md](./ROADMAP.md) - Project roadmap and priorities
- [Keep a Changelog](https://keepachangelog.com/) - Changelog standard
- [Semantic Versioning](https://semver.org/) - Version numbering
- [GoReleaser Docs](https://goreleaser.com/) - Release automation

---

**Pro Tip**: Use `make dev` during development for quick iteration (format + build + install in one command)!
