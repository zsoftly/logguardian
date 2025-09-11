#!/usr/bin/env bash

# Release Documentation Generator - Simplified and Fixed
# Generates CHANGELOG.md and RELEASE_NOTES.txt for releases

set -e

# Script configuration
VERSION=""
LATEST_TAG=""
FORCE_REGENERATE=false
DEBUG_MODE=false
COMMIT_CHANGES=false

usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Generate CHANGELOG.md and RELEASE_NOTES.md for a release version.

OPTIONS:
    -v, --version VERSION    Release version (required, format: v1.2.3 or 1.2.3)
    -t, --latest-tag TAG     Latest release tag for comparison (auto-detected if not provided)
    -f, --force             Force regeneration even if files exist
    -c, --commit            Commit and push changes (for CI/CD)
    -d, --debug             Enable debug mode with verbose output
    -h, --help              Show this help message

EXAMPLES:
    $0 --version v2.1.0
    $0 --version 2.1.0 --latest-tag v2.0.5 --debug

EOF
}

debug_log() {
    if [[ "$DEBUG_MODE" == true ]]; then
        echo "[DEBUG] $*" >&2
    fi
}

log_info() {
    echo "[INFO] $*" >&2
}

log_error() {
    echo "[ERROR] $*" >&2
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version) VERSION="$2"; shift 2 ;;
            -t|--latest-tag) LATEST_TAG="$2"; shift 2 ;;
            -f|--force) FORCE_REGENERATE=true; shift ;;
            -c|--commit) COMMIT_CHANGES=true; shift ;;
            -d|--debug) DEBUG_MODE=true; shift ;;
            -h|--help) usage; exit 0 ;;
            *) log_error "Unknown option: $1"; usage; exit 1 ;;
        esac
    done

    if [[ -z "$VERSION" ]]; then
        log_error "Version is required. Use --version or -v to specify."
        usage
        exit 1
    fi
}

# Validate version format
validate_version() {
    if [[ ! "$VERSION" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Invalid version format: $VERSION"
        exit 1
    fi
    log_info "Version format is valid: $VERSION"
}

# Get latest release tag if not provided
get_latest_tag() {
    if [[ -n "$LATEST_TAG" ]]; then
        log_info "Using provided latest tag: $LATEST_TAG"
        return
    fi
    
    LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    
    if [[ -n "$LATEST_TAG" ]]; then
        log_info "Found latest tag: $LATEST_TAG"
        debug_log "Tag date: $(git log -1 --format=%ai "$LATEST_TAG" 2>/dev/null || echo "unknown")"
    else
        log_info "No previous tags found - will include all commits"
    fi
}

# Get all commits and generate both files
generate_documentation() {
    local commit_range
    
    if [[ -n "$LATEST_TAG" ]]; then
        commit_range="${LATEST_TAG}..HEAD"
    else
        commit_range="HEAD"
    fi
    
    log_info "Getting commits using range: $commit_range"
    
    # Get raw commits and save to temporary file for processing
    local commits_file
    commits_file=$(mktemp)
    
    git log --pretty=format:"%s" --no-merges "$commit_range" 2>/dev/null > "$commits_file" || {
        log_error "Failed to get git log"
        rm -f "$commits_file"
        exit 1
    }
    
    local total_commits
    total_commits=$(wc -l < "$commits_file")
    debug_log "Total raw commits: $total_commits"
    
    if [[ $total_commits -eq 0 ]]; then
        log_info "No commits found"
        rm -f "$commits_file"
        return
    fi
    
    # Filter out automation commits
    local filtered_file
    filtered_file=$(mktemp)
    
    grep -v -E "^(Auto-generate changelog|Restore RELEASE_NOTES|Restore CHANGELOG|Merge (branch|pull request))" "$commits_file" > "$filtered_file" || cp "$commits_file" "$filtered_file"
    
    local filtered_commits
    filtered_commits=$(wc -l < "$filtered_file")
    debug_log "Commits after filtering: $filtered_commits"
    
    if [[ $filtered_commits -eq 0 ]]; then
        log_info "No notable changes after filtering"
        rm -f "$commits_file" "$filtered_file"
        return
    fi
    
    # Show first few commits for debugging
    if [[ "$DEBUG_MODE" == true ]]; then
        debug_log "First 10 filtered commits:"
        head -10 "$filtered_file" | nl >&2
    fi
    
    # Create temporary files for each category
    local features_file fixes_file other_file
    features_file=$(mktemp)
    fixes_file=$(mktemp)
    other_file=$(mktemp)
    
    # Categorize commits
    while IFS= read -r commit; do
        [[ -z "$commit" ]] && continue
        
        debug_log "Processing: $commit"
        
        if [[ "$commit" =~ ^(feat|feature)(\(.*\))?:?.* ]] || [[ "$commit" =~ ^(Enhance|Add|Implement|Create) ]]; then
            echo "$commit" >> "$features_file"
            debug_log "  -> FEATURE"
        elif [[ "$commit" =~ ^(fix|bug|hotfix)(\(.*\))?:?.* ]] || [[ "$commit" =~ (fix|bug|resolve|correct) ]]; then
            echo "$commit" >> "$fixes_file"
            debug_log "  -> FIX"
        else
            echo "$commit" >> "$other_file"
            debug_log "  -> OTHER"
        fi
    done < "$filtered_file"
    
    local feature_count fix_count other_count
    feature_count=$(wc -l < "$features_file")
    fix_count=$(wc -l < "$fixes_file")
    other_count=$(wc -l < "$other_file")
    
    debug_log "Final counts: Features=$feature_count, Fixes=$fix_count, Other=$other_count"
    
    # Generate CHANGELOG.md
    generate_changelog "$features_file" "$fixes_file" "$other_file"
    
    # Generate RELEASE_NOTES.txt
    generate_release_notes "$features_file" "$fixes_file" "$other_file"
    
    # Cleanup
    rm -f "$commits_file" "$filtered_file" "$features_file" "$fixes_file" "$other_file"
}

generate_changelog() {
    local features_file="$1"
    local fixes_file="$2"
    local other_file="$3"
    
    log_info "Generating CHANGELOG.md..."
    
    local temp_entry
    temp_entry=$(mktemp)
    
    echo "## [$VERSION] - $(date +%Y-%m-%d)" > "$temp_entry"
    echo "" >> "$temp_entry"
    
    # Add features section
    if [[ -s "$features_file" ]]; then
        echo "### Added" >> "$temp_entry"
        while IFS= read -r commit; do
            echo "- $commit" >> "$temp_entry"
        done < "$features_file"
        echo "" >> "$temp_entry"
        debug_log "Added $(wc -l < "$features_file") features to changelog"
    fi
    
    # Add fixes section
    if [[ -s "$fixes_file" ]]; then
        echo "### Fixed" >> "$temp_entry"
        while IFS= read -r commit; do
            echo "- $commit" >> "$temp_entry"
        done < "$fixes_file"
        echo "" >> "$temp_entry"
        debug_log "Added $(wc -l < "$fixes_file") fixes to changelog"
    fi
    
    # Add other changes section
    if [[ -s "$other_file" ]]; then
        echo "### Changed" >> "$temp_entry"
        while IFS= read -r commit; do
            echo "- $commit" >> "$temp_entry"
        done < "$other_file"
        echo "" >> "$temp_entry"
        debug_log "Added $(wc -l < "$other_file") other changes to changelog"
    fi
    
    # Update CHANGELOG.md
    if [[ -f CHANGELOG.md ]]; then
        local temp_changelog
        temp_changelog=$(mktemp)
        { head -n 1 CHANGELOG.md; echo ""; cat "$temp_entry"; tail -n +2 CHANGELOG.md; } > "$temp_changelog"
        mv "$temp_changelog" CHANGELOG.md
        log_info "Updated existing CHANGELOG.md"
    else
        {
            echo "# Changelog"
            echo ""
            cat "$temp_entry"
        } > CHANGELOG.md
        log_info "Created new CHANGELOG.md"
    fi
    
    rm -f "$temp_entry"
}

update_version_files() {
    local version_no_v="${VERSION#v}"  # Remove 'v' prefix if present
    
    log_info "Updating version to $VERSION in source files..."
    
    # Update VERSION file
    echo "$VERSION" > VERSION
    log_info "Updated VERSION file"
    
    # Update Lambda version if the const exists
    if [[ -f cmd/lambda/main.go ]] && grep -q 'const version = ' cmd/lambda/main.go 2>/dev/null; then
        sed -i "s/const version = .*/const version = \"${VERSION}\"/" cmd/lambda/main.go
        log_info "Updated Lambda version in cmd/lambda/main.go"
        debug_log "Lambda version line: $(grep 'const version = ' cmd/lambda/main.go)"
    fi
    
    # Update Container version if the var exists
    if [[ -f cmd/container/main.go ]] && grep -q 'var version = ' cmd/container/main.go 2>/dev/null; then
        sed -i "s/var version = .*/var version = \"${VERSION}\"/" cmd/container/main.go
        log_info "Updated Container version in cmd/container/main.go"
        debug_log "Container version line: $(grep 'var version = ' cmd/container/main.go)"
    fi
    
    # Update SAM template if it exists
    if [[ -f template.yaml ]]; then
        sed -i "s/SemanticVersion: .*/SemanticVersion: ${version_no_v}/" template.yaml
        log_info "Updated template.yaml with version $version_no_v"
    fi
    
    if [[ -f template.yml ]]; then
        sed -i "s/SemanticVersion: .*/SemanticVersion: ${version_no_v}/" template.yml
        log_info "Updated template.yml with version $version_no_v"
    fi
    
    # Update Makefile if VERSION variable exists
    if [[ -f Makefile ]] && grep -q '^VERSION ?=' Makefile 2>/dev/null; then
        sed -i "s/^VERSION ?=.*/VERSION ?= ${VERSION}/" Makefile
        log_info "Updated Makefile with version $VERSION"
    fi
}

generate_release_notes() {
    local features_file="$1"
    local fixes_file="$2"
    local other_file="$3"
    
    log_info "Generating RELEASE_NOTES.md..."
    
    # Get GitHub repo info
    local repo_url
    repo_url=$(git config --get remote.origin.url 2>/dev/null || echo "")
    local github_repo=""
    
    if [[ "$repo_url" =~ github\.com[:/]([^/]+/[^/]+)(\.git)?$ ]]; then
        github_repo="${BASH_REMATCH[1]%.git}"
    elif [[ "$repo_url" =~ ([^/]+/[^/]+)\.git$ ]]; then
        # Fallback for other Git URL formats
        github_repo="${BASH_REMATCH[1]}"
    fi
    
    # If we still couldn't extract it, try a simpler approach
    if [[ -z "$github_repo" && "$repo_url" =~ github\.com ]]; then
        # Extract zsoftly/logguardian from git@github.com:zsoftly/logguardian.git
        github_repo=$(echo "$repo_url" | sed -n 's/.*github\.com[:/]\([^/]*\/[^/]*\)\.git.*/\1/p')
    fi
    
    # Create release notes header
    {
        echo "# LogGuardian $VERSION Release Notes"
        echo ""
        echo "**Release Date:** $(date '+%B %d, %Y')"
        echo ""
        echo "## Overview"
        echo ""
        echo "LogGuardian is an enterprise-grade AWS CloudWatch log compliance automation tool that ensures log groups meet security and retention requirements."
        echo ""
        echo "This release can be deployed as:"
        echo "- Docker container via GitHub Container Registry (ghcr.io)"
        echo "- AWS Lambda function for serverless deployments"
        echo "- Standalone container binary"
        echo ""
    } > RELEASE_NOTES.md
    
    # Add features section
    echo "## 🚀 New Features" >> RELEASE_NOTES.md
    if [[ -s "$features_file" ]]; then
        while IFS= read -r commit; do
            echo "- $commit" >> RELEASE_NOTES.md
        done < "$features_file"
        debug_log "Added $(wc -l < "$features_file") features to release notes"
    else
        echo "- No new features in this release" >> RELEASE_NOTES.md
    fi
    echo "" >> RELEASE_NOTES.md
    
    # Add fixes section
    echo "## 🐛 Bug Fixes" >> RELEASE_NOTES.md
    if [[ -s "$fixes_file" ]]; then
        while IFS= read -r commit; do
            echo "- $commit" >> RELEASE_NOTES.md
        done < "$fixes_file"
        debug_log "Added $(wc -l < "$fixes_file") fixes to release notes"
    else
        echo "- No bug fixes in this release" >> RELEASE_NOTES.md
    fi
    echo "" >> RELEASE_NOTES.md
    
    # Add other changes section
    echo "## 🔧 Other Changes" >> RELEASE_NOTES.md
    if [[ -s "$other_file" ]]; then
        while IFS= read -r commit; do
            echo "- $commit" >> RELEASE_NOTES.md
        done < "$other_file"
        debug_log "Added $(wc -l < "$other_file") other changes to release notes"
    else
        echo "- No other changes in this release" >> RELEASE_NOTES.md
    fi
    echo "" >> RELEASE_NOTES.md
    
    # Add installation section
    echo "## Installation" >> RELEASE_NOTES.md
    echo "" >> RELEASE_NOTES.md
    echo "### Docker (Recommended)" >> RELEASE_NOTES.md
    echo '```bash' >> RELEASE_NOTES.md
    echo "docker pull ghcr.io/${github_repo:-zsoftly/logguardian}:${VERSION}" >> RELEASE_NOTES.md
    echo '```' >> RELEASE_NOTES.md
    echo "" >> RELEASE_NOTES.md
    echo "### AWS Lambda" >> RELEASE_NOTES.md
    echo "Download the Lambda deployment package (logguardian-compliance-${VERSION}.zip) from the release artifacts." >> RELEASE_NOTES.md
    echo "" >> RELEASE_NOTES.md
    echo "### Container Binary" >> RELEASE_NOTES.md
    echo "Download the standalone container binary (logguardian-container-${VERSION}) from the release artifacts." >> RELEASE_NOTES.md
    echo "" >> RELEASE_NOTES.md
    
    log_info "RELEASE_NOTES.md generated successfully"
}

# Check if files already exist
check_existing_files() {
    if [[ -f CHANGELOG.md ]] && [[ "$FORCE_REGENERATE" != true ]]; then
        log_error "CHANGELOG.md already exists. Use --force to regenerate."
        exit 1
    fi
    
    if [[ -f RELEASE_NOTES.md ]] && [[ "$FORCE_REGENERATE" != true ]]; then
        log_error "RELEASE_NOTES.md already exists. Use --force to regenerate."
        exit 1
    fi
}

# Commit generated changes
commit_changes() {
    local version="${VERSION}"
    
    log_info "Committing generated changes..."
    
    # Configure git if needed
    if [[ -n "${GITHUB_ACTIONS:-}" ]]; then
        git config --local user.email "action@github.com"
        git config --local user.name "GitHub Action"
    fi
    
    # Add all potentially modified files
    git add -A CHANGELOG.md RELEASE_NOTES.md VERSION 2>/dev/null || true
    git add -A cmd/lambda/main.go cmd/container/main.go 2>/dev/null || true
    git add -A template.yaml template.yml Makefile 2>/dev/null || true
    
    if git diff --staged --quiet; then
        log_info "No changes to commit"
        return 1
    else
        git commit -m "docs: Auto-generate release documentation for ${version}

- Generated CHANGELOG.md from commit history
- Created RELEASE_NOTES.md with installation instructions  
- Updated VERSION file to ${version}
- Updated version strings in source files"
        
        log_info "Changes committed successfully"
        
        # Push if in GitHub Actions
        if [[ -n "${GITHUB_ACTIONS:-}" ]]; then
            git push
            log_info "Changes pushed to remote"
        fi
        return 0
    fi
}

# Generate GitHub Actions summary
generate_github_summary() {
    local version="${VERSION}"
    local branch="${GITHUB_REF#refs/heads/}"
    
    # Only generate summary if running in GitHub Actions
    if [[ -z "${GITHUB_STEP_SUMMARY:-}" ]]; then
        return
    fi
    
    log_info "Generating GitHub Actions summary..."
    
    {
        echo "## 📝 Release Documentation Generated"
        echo ""
        echo "**Version:** ${version}"
        echo "**Branch:** ${branch}"
        echo ""
        echo "### Files Updated:"
        echo "- ✅ CHANGELOG.md"
        echo "- ✅ RELEASE_NOTES.md"
        echo "- ✅ VERSION"
        echo "- ✅ Version strings in source files"
        echo ""
        echo "### Next Steps:"
        echo "1. Pull the changes: \`git pull origin ${branch}\`"
        echo "2. Review the generated documentation"
        echo "3. Make any manual edits if needed"
        echo "4. Create and push the release tag:"
        echo "   \`\`\`bash"
        echo "   git tag -a ${version} -m \"Release ${version}\""
        echo "   git push origin ${version}"
        echo "   \`\`\`"
        echo "5. Merge release branch back to main after release"
    } >> "${GITHUB_STEP_SUMMARY}"
    
    log_info "GitHub summary generated"
}

# Main execution
main() {
    log_info "Starting release documentation generation"
    
    parse_args "$@"
    validate_version
    get_latest_tag
    check_existing_files
    
    # Update version in source files first
    update_version_files
    
    generate_documentation
    
    # Commit changes if requested
    if [[ "${COMMIT_CHANGES:-false}" == true ]]; then
        commit_changes || true
    fi
    
    # Generate GitHub summary if in GitHub Actions
    generate_github_summary
    
    log_info "Release documentation generation completed successfully!"
}

# Execute main function
main "$@"