#!/usr/bin/env bash
# Interactive helper to add entries to CHANGELOG.md Unreleased section

set -e

CHANGELOG="CHANGELOG.md"

# Check if CHANGELOG.md exists
if [ ! -f "$CHANGELOG" ]; then
    echo "❌ CHANGELOG.md not found"
    exit 1
fi

# Check if Unreleased section exists
if ! grep -q "## \[Unreleased\]" "$CHANGELOG"; then
    echo "❌ No [Unreleased] section found in CHANGELOG.md"
    exit 1
fi

echo "Add entry to CHANGELOG.md [Unreleased] section"
echo ""
echo "Category:"
echo "  1) Added (new features)"
echo "  2) Changed (changes to existing functionality)"
echo "  3) Deprecated (soon-to-be removed features)"
echo "  4) Removed (removed features)"
echo "  5) Fixed (bug fixes)"
echo "  6) Security (security fixes)"
echo ""
read -p "Select category [1-6]: " CATEGORY

case $CATEGORY in
    1) SECTION="Added" ;;
    2) SECTION="Changed" ;;
    3) SECTION="Deprecated" ;;
    4) SECTION="Removed" ;;
    5) SECTION="Fixed" ;;
    6) SECTION="Security" ;;
    *)
        echo "❌ Invalid selection"
        exit 1
        ;;
esac

echo ""
read -p "Enter description (one line): " DESCRIPTION

if [ -z "$DESCRIPTION" ]; then
    echo "❌ Description required"
    exit 1
fi

# Find the Unreleased section and add entry
# This is a simplified implementation - for production use a proper CHANGELOG parser
echo ""
echo "Manual addition required:"
echo ""
echo "Add this line under the '### $SECTION' section in [Unreleased]:"
echo ""
echo "- $DESCRIPTION"
echo ""
echo "If the '### $SECTION' section doesn't exist, create it under '## [Unreleased]'"
echo ""
echo "Tip: Open CHANGELOG.md and add manually, or use a proper CHANGELOG tool like 'git-chglog'"
