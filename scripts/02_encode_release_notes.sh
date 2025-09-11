#!/usr/bin/env bash
# Script 02: Encode Release Notes
# Encode release notes to base64 for safe passing through GitHub Actions
# This script handles the encoding of RELEASE_NOTES.md or generates from git

set -e

# Script variables
RELEASE_TAG="${1:-${GITHUB_REF_NAME}}"
GITHUB_REPOSITORY="${2:-${GITHUB_REPOSITORY}}"

# Helper functions
log_info() {
    echo "[INFO] $*" >&2
}

# Generate release notes from git history
generate_from_git() {
    local prev_tag
    prev_tag=$(git describe --tags --abbrev=0 "${RELEASE_TAG}^" 2>/dev/null || echo "")
    
    {
        echo "# Release $RELEASE_TAG"
        echo ""
        echo "Release Date: $(date +%Y-%m-%d)"
        echo ""
        
        if [[ -n "$prev_tag" ]]; then
            echo "## Changes since $prev_tag"
            echo ""
            
            # Features
            echo "### 🚀 Features"
            local features
            features=$(git log "$prev_tag..$RELEASE_TAG" --grep="^feat" --pretty=format:"- %s (%h)" 2>/dev/null || echo "")
            if [[ -n "$features" ]]; then
                echo "$features"
            else
                echo "- No new features"
            fi
            echo ""
            echo ""
            
            # Bug fixes
            echo "### 🐛 Bug Fixes"
            local fixes
            fixes=$(git log "$prev_tag..$RELEASE_TAG" --grep="^fix" --pretty=format:"- %s (%h)" 2>/dev/null || echo "")
            if [[ -n "$fixes" ]]; then
                echo "$fixes"
            else
                echo "- No bug fixes"
            fi
            echo ""
            echo ""
            
            # Other changes
            echo "### 🔧 Other Changes"
            local other
            other=$(git log "$prev_tag..$RELEASE_TAG" --grep="^chore\|^docs\|^test\|^refactor" --pretty=format:"- %s (%h)" 2>/dev/null || echo "")
            if [[ -n "$other" ]]; then
                echo "$other"
            else
                echo "- No other changes"
            fi
            echo ""
        else
            echo "Initial release of LogGuardian - AWS CloudWatch Log Compliance Automation"
            echo ""
        fi
        
        echo ""
        echo "## Installation"
        echo ""
        echo "### Docker"
        echo '```bash'
        echo "docker pull ghcr.io/$GITHUB_REPOSITORY:$RELEASE_TAG"
        echo '```'
        echo ""
        echo "### Lambda"
        echo "Download the Lambda deployment package from the release artifacts below."
        echo ""
        echo "### Container Binary"
        echo "Download the standalone container binary from the release artifacts below."
    }
}

# Main logic
main() {
    if [[ -f "RELEASE_NOTES.md" ]]; then
        log_info "Found RELEASE_NOTES.md, encoding to base64"
        base64 -w0 < RELEASE_NOTES.md
    else
        log_info "No RELEASE_NOTES.md found, generating from git log"
        generate_from_git | base64 -w0
    fi
}

# Execute
main