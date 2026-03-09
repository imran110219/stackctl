#!/usr/bin/env bash
# Suggest next version based on conventional commits since last tag

set -e

# Get the latest tag
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo "Latest tag: $LATEST_TAG"

# Parse version (remove 'v' prefix)
VERSION=${LATEST_TAG#v}

# Split version into parts
IFS='.' read -r -a VERSION_PARTS <<< "$VERSION"
MAJOR="${VERSION_PARTS[0]}"
MINOR="${VERSION_PARTS[1]}"
PATCH="${VERSION_PARTS[2]%-*}"  # Remove any suffix like -rc1

# Check commits since last tag for breaking changes, features, fixes
HAS_BREAKING=false
HAS_FEAT=false
HAS_FIX=false

while IFS= read -r commit; do
    if [[ $commit =~ ^.*!:|BREAKING[[:space:]]CHANGE ]]; then
        HAS_BREAKING=true
    elif [[ $commit =~ ^feat ]]; then
        HAS_FEAT=true
    elif [[ $commit =~ ^fix ]]; then
        HAS_FIX=true
    fi
done < <(git log "$LATEST_TAG..HEAD" --pretty=format:"%s" 2>/dev/null || true)

# Determine next version
if [ "$HAS_BREAKING" = true ]; then
    NEXT_MAJOR=$((MAJOR + 1))
    echo "v${NEXT_MAJOR}.0.0 (breaking changes detected)"
elif [ "$HAS_FEAT" = true ]; then
    NEXT_MINOR=$((MINOR + 1))
    echo "v${MAJOR}.${NEXT_MINOR}.0 (new features detected)"
elif [ "$HAS_FIX" = true ]; then
    NEXT_PATCH=$((PATCH + 1))
    echo "v${MAJOR}.${MINOR}.${NEXT_PATCH} (bug fixes detected)"
else
    echo "v${MAJOR}.${MINOR}.${PATCH} (no conventional commits found)"
fi

# Show release type guidance
echo ""
echo "Release type guidance:"
echo "  - MAJOR (breaking): New version breaks backward compatibility"
echo "  - MINOR (feature): New features, backward compatible"
echo "  - PATCH (fix): Bug fixes, no new features"
echo ""
echo "For pre-release, append: -rc1, -rc2, etc."
echo "Example: v0.1.0-rc1 → v0.1.0"
