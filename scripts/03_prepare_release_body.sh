#!/usr/bin/env bash
# Script 03: Prepare Release Body
# Prepare complete release body for GitHub release
# This script handles base64 decoding and assembly of the full release body with checksums and quick start

set -e

# Script variables
RELEASE_TAG="${1:-${GITHUB_REF_NAME}}"
GITHUB_REPOSITORY="${2:-${GITHUB_REPOSITORY}}"
NOTES_BASE64="${3}"
OUTPUT_FILE="${4:-release_body.md}"

# Helper functions
log_info() {
    echo "[INFO] $*" >&2
}

log_error() {
    echo "[ERROR] $*" >&2
    exit 1
}

# Validate inputs
if [[ -z "$RELEASE_TAG" ]]; then
    log_error "Release tag is required (arg 1 or GITHUB_REF_NAME env var)"
fi

if [[ -z "$GITHUB_REPOSITORY" ]]; then
    log_error "GitHub repository is required (arg 2 or GITHUB_REPOSITORY env var)"
fi

# Main release notes preparation
prepare_release_notes() {
    log_info "Preparing release notes for $RELEASE_TAG"
    
    # If we have base64 encoded notes, decode them
    if [[ -n "$NOTES_BASE64" ]]; then
        log_info "Decoding base64 release notes"
        echo "$NOTES_BASE64" | base64 -d > "$OUTPUT_FILE"
    elif [[ -f "RELEASE_NOTES.md" ]]; then
        log_info "Using RELEASE_NOTES.md file"
        cp RELEASE_NOTES.md "$OUTPUT_FILE"
    else
        log_info "Generating release notes from git history"
        generate_release_notes_from_git > "$OUTPUT_FILE"
    fi
    
    # Append checksums if available
    if [[ -f "artifacts/checksums.txt" ]]; then
        log_info "Appending checksums"
        {
            echo ""
            echo "## Checksums"
            echo '```'
            cat artifacts/checksums.txt
            echo '```'
        } >> "$OUTPUT_FILE"
    fi
    
    # Append quick start section
    log_info "Appending quick start section"
    append_quick_start
    
    log_info "Release notes prepared successfully: $OUTPUT_FILE"
}

# Generate release notes from git history
generate_release_notes_from_git() {
    local prev_tag
    prev_tag=$(git describe --tags --abbrev=0 "${RELEASE_TAG}^" 2>/dev/null || echo "")
    
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

# Append quick start section to release notes
append_quick_start() {
    cat >> "$OUTPUT_FILE" << EOF

## Quick Start

### Docker (Recommended)
\`\`\`bash
# Pull and run the container
docker pull ghcr.io/$GITHUB_REPOSITORY:$RELEASE_TAG
docker run --rm ghcr.io/$GITHUB_REPOSITORY:$RELEASE_TAG --help
\`\`\`

### AWS Lambda
\`\`\`bash
# Download and deploy
curl -L -o logguardian-compliance.zip \\
  https://github.com/$GITHUB_REPOSITORY/releases/download/$RELEASE_TAG/logguardian-compliance-$RELEASE_TAG.zip

aws lambda update-function-code \\
  --function-name logguardian-compliance \\
  --zip-file fileb://logguardian-compliance.zip
\`\`\`

See [Documentation](https://github.com/$GITHUB_REPOSITORY#documentation--support) for detailed usage instructions.
EOF
}

# Execute main function
prepare_release_notes