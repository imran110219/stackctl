#!/usr/bin/env bash
# Prepare CHANGELOG.md for a release by updating Unreleased section

set -e

CHANGELOG="CHANGELOG.md"

# Check if CHANGELOG.md exists
if [ ! -f "$CHANGELOG" ]; then
    echo "❌ CHANGELOG.md not found"
    exit 1
fi

# Prompt for version
echo "Preparing changelog for release..."
echo ""
read -p "Enter version (e.g., 0.1.0 or 0.1.0-rc2): " VERSION

if [ -z "$VERSION" ]; then
    echo "❌ Version required"
    exit 1
fi

# Get today's date
DATE=$(date +%Y-%m-%d)

# Check if Unreleased section exists
if ! grep -q "## \[Unreleased\]" "$CHANGELOG"; then
    echo "❌ No [Unreleased] section found in CHANGELOG.md"
    exit 1
fi

# Create backup
cp "$CHANGELOG" "${CHANGELOG}.bak"
echo "✓ Created backup: ${CHANGELOG}.bak"

# Replace [Unreleased] with version and date
# This is a simple approach - manually verify the output!
sed -i.tmp "s/## \[Unreleased\]/## [Unreleased]\n\n### Planned\n- (Add planned features here)\n\n## [$VERSION] - $DATE/" "$CHANGELOG"
rm "${CHANGELOG}.tmp"

echo ""
echo "✓ Updated CHANGELOG.md:"
echo "  - Added new [Unreleased] section"
echo "  - Created [$VERSION] section with date $DATE"
echo ""
echo "Next steps:"
echo "  1. Review CHANGELOG.md (backup at ${CHANGELOG}.bak)"
echo "  2. Move items from Unreleased to [$VERSION] section"
echo "  3. Commit: git add CHANGELOG.md && git commit -m 'chore: prepare release v$VERSION'"
echo "  4. Tag: git tag -a v$VERSION -m 'Release v$VERSION'"
echo "  5. Push: git push && git push --tags"
echo ""
echo "To restore backup: mv ${CHANGELOG}.bak $CHANGELOG"
