#!/bin/bash

# Update version script for LogGuardian
# Usage: ./scripts/update-version.sh <new-version>

set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <new-version>"
    echo "Example: $0 1.2.0"
    exit 1
fi

NEW_VERSION=$1
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Strip 'v' prefix if provided, then validate
CLEAN_VERSION="${NEW_VERSION#v}"

# Validate version format (without v prefix)
if ! echo "$CLEAN_VERSION" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' > /dev/null; then
    echo "Error: Version must be in format X.Y.Z or vX.Y.Z (e.g., 1.2.0 or v1.2.0)"
    exit 1
fi

# VERSION file always has 'v' prefix
VERSION_WITH_V="v${CLEAN_VERSION}"
# template.yaml always uses semantic version without 'v' prefix
SEMANTIC_VERSION="${CLEAN_VERSION}"

echo "Updating version to $VERSION_WITH_V..."

# Update VERSION file (with v prefix)
echo "$VERSION_WITH_V" > "$PROJECT_ROOT/VERSION"
echo "✅ Updated VERSION file to $VERSION_WITH_V"

# Update template.yaml (without v prefix for SAR)
sed -i "s/SemanticVersion: .*/SemanticVersion: $SEMANTIC_VERSION/" "$PROJECT_ROOT/template.yaml"
echo "✅ Updated template.yaml to $SEMANTIC_VERSION"

echo ""
echo "Version updated to $VERSION_WITH_V"
echo ""
echo "Next steps:"
echo "1. Review the changes: git diff"
echo "2. Commit: git add -A && git commit -m \"chore: Bump version to $VERSION_WITH_V\""
echo "3. Build: make build"
echo "4. Package: make package"
echo "5. Tag: git tag -a $VERSION_WITH_V -m \"Release version $VERSION_WITH_V\""
echo "6. Push: git push origin <branch> && git push origin $VERSION_WITH_V"
echo "7. Publish to SAR: make publish"