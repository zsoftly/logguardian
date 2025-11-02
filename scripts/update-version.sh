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

# Strip 'v' prefix if provided (for convenience)
VERSION="${NEW_VERSION#v}"

# Validate version format (semantic versioning X.Y.Z)
if ! echo "$VERSION" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' > /dev/null; then
    echo "Error: Version must be in format X.Y.Z (e.g., 1.2.0)"
    exit 1
fi

echo "Updating version to $VERSION..."

# Update VERSION file (pure semantic versioning)
echo "$VERSION" > "$PROJECT_ROOT/VERSION"
echo "✅ Updated VERSION file to $VERSION"

# Update template.yaml (same format)
sed -i "s/SemanticVersion: .*/SemanticVersion: $VERSION/" "$PROJECT_ROOT/template.yaml"
echo "✅ Updated template.yaml to $VERSION"

echo ""
echo "Version updated to $VERSION"
echo ""
echo "Next steps:"
echo "1. Review the changes: git diff"
echo "2. Commit: git add -A && git commit -m \"chore: Bump version to $VERSION\""
echo "3. Build: make build"
echo "4. Package: make package"
echo "5. Tag: git tag -a $VERSION -m \"Release version $VERSION\""
echo "6. Push: git push origin <branch> && git push origin $VERSION"
echo "7. Publish to SAR: make publish"