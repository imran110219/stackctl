#!/usr/bin/env bash
# Perform a dry run of the release process

set -e

echo "=========================================="
echo "stackctl Release Dry Run"
echo "=========================================="
echo ""

# Check if goreleaser is installed
SKIP_GORELEASER=false
if ! command -v goreleaser >/dev/null 2>&1; then
    echo "⚠️  GoReleaser not installed (optional for development)"
    echo ""
    echo "GoReleaser is only needed for:"
    echo "  - Local release testing (this command)"
    echo "  - Building snapshot releases"
    echo ""
    echo "For actual releases, GitHub Actions handles everything!"
    echo ""
    echo "To install GoReleaser (optional):"
    echo "  macOS:   brew install goreleaser"
    echo "  Linux:   https://goreleaser.com/install/"
    echo ""
    echo "Continuing with basic checks (without GoReleaser build test)..."
    echo ""
    SKIP_GORELEASER=true
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    echo "⚠ Warning: Working directory has uncommitted changes"
    echo ""
fi

# Get current branch
BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo "Current branch: $BRANCH"
echo ""

# Get latest tag
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "none")
echo "Latest tag: $LATEST_TAG"
echo ""

# Show commits since last tag
echo "Commits since last tag:"
if [ "$LATEST_TAG" != "none" ]; then
    git log "$LATEST_TAG..HEAD" --oneline | head -10
else
    git log --oneline | head -10
fi
echo ""

# Check CHANGELOG.md
echo "Checking CHANGELOG.md..."
if [ -f "CHANGELOG.md" ]; then
    echo "✓ CHANGELOG.md exists"

    if grep -q "## \[Unreleased\]" CHANGELOG.md; then
        echo "✓ [Unreleased] section found"

        # Check if there's content in Unreleased
        UNRELEASED_CONTENT=$(sed -n '/## \[Unreleased\]/,/## \[/p' CHANGELOG.md | grep -E '^- ' || true)
        if [ -n "$UNRELEASED_CONTENT" ]; then
            echo "⚠ Unreleased section has content - should be moved to a version section before release"
        else
            echo "✓ Unreleased section is clean"
        fi
    else
        echo "⚠ No [Unreleased] section found"
    fi
else
    echo "❌ CHANGELOG.md not found"
fi
echo ""

# Suggest next version
echo "Suggested next version:"
./scripts/suggest-version.sh 2>/dev/null || echo "  (could not determine)"
echo ""

# Test GoReleaser config
if [ "$SKIP_GORELEASER" = false ]; then
    echo "Testing GoReleaser configuration..."
    if goreleaser check; then
        echo "✓ GoReleaser config is valid"
    else
        echo "❌ GoReleaser config has errors"
        exit 1
    fi
    echo ""

    # Build a snapshot
    echo "Building snapshot release..."
    echo "(This creates local builds without publishing)"
    echo ""

    if goreleaser release --snapshot --clean --skip=publish; then
        echo ""
        echo "✓ Snapshot build successful!"
        echo ""
        echo "Build artifacts in dist/:"
        ls -lh dist/ | grep -E '^-' | awk '{print "  " $9 " (" $5 ")"}'
    else
        echo "❌ Snapshot build failed"
        exit 1
    fi
else
    echo "⏭️  Skipping GoReleaser build test (not installed)"
    echo ""
    echo "Testing basic Go build instead..."
    if go build -o /tmp/stackctl-test ./cmd/stackctl; then
        echo "✓ Go build successful!"
        /tmp/stackctl-test version 2>/dev/null || echo "  (version info will be injected by GoReleaser)"
        rm /tmp/stackctl-test
    else
        echo "❌ Go build failed"
        exit 1
    fi
fi

echo ""
echo "=========================================="
echo "Dry Run Complete!"
echo "=========================================="
echo ""
echo "To create a real release:"
echo "  1. Update CHANGELOG.md: make changelog-prepare"
echo "  2. Commit changes: git add CHANGELOG.md && git commit -m 'chore: prepare release vX.Y.Z'"
echo "  3. Create tag: git tag -a vX.Y.Z -m 'Release vX.Y.Z'"
echo "  4. Push: git push && git push --tags"
echo ""
echo "GitHub Actions will automatically build and publish the release."
