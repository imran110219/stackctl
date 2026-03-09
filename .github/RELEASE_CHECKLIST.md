# Release Checklist

Use this checklist when creating a new release.

## Pre-Release

- [ ] All tests passing: `make test`
- [ ] Code formatted: `make fmt`
- [ ] Build successful: `make release-snapshot`
- [ ] Version bumped correctly (see `make version`)
- [ ] CHANGELOG.md updated with version and date
- [ ] Known issues documented in CHANGELOG.md
- [ ] ROADMAP.md reflects current status

## Release

- [ ] Committed CHANGELOG.md: `git add CHANGELOG.md && git commit -m "chore: prepare release vX.Y.Z"`
- [ ] Created tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
- [ ] Pushed changes: `git push`
- [ ] Pushed tag: `git push --tags`
- [ ] GitHub Actions workflow succeeded
- [ ] GitHub Release published

## Post-Release

- [ ] Test installation on Ubuntu 22.04: `curl -fsSL https://raw.githubusercontent.com/imran110219/stackctl/main/install.sh | bash`
- [ ] Test installation on Ubuntu 24.04
- [ ] Verify `stackctl version` shows correct version
- [ ] Test basic commands: `stackctl doctor`, `stackctl status`
- [ ] Update ROADMAP.md (mark completed tasks as ✅)
- [ ] Update MEMORY.md if significant changes
- [ ] Update README.md if needed
- [ ] Update USER_GUIDE.md if needed

## Rollback (if needed)

- [ ] Delete GitHub Release (via web UI)
- [ ] Delete tag locally: `git tag -d vX.Y.Z`
- [ ] Delete tag remotely: `git push origin :refs/tags/vX.Y.Z`
- [ ] Fix issues
- [ ] Re-release with same or new version

---

**Pro Tip:** Use `make release-dry-run` to catch issues before pushing!
