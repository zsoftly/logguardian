#!/bin/bash

# LogGuardian CloudFormation Template Validation Script
# Validates all CloudFormation templates for syntax and best practices

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"

# Template files to validate
TEMPLATES=(
  "00-logguardian-simple.yaml"
  "01-logguardian-main.yaml"
  "04-kms-key.yaml"
  "02-iam-roles.yaml"
  "03-lambda-function.yaml"
  "05-config-rules.yaml"
  "06-eventbridge-rules.yaml"
  "07-monitoring.yaml"
  "08-logguardian-stacksets.yaml"
)

# Referenced templates for relationship checking
REFERENCED_TEMPLATES=(
  "04-kms-key.yaml"
  "02-iam-roles.yaml"
  "03-lambda-function.yaml"
  "05-config-rules.yaml"
  "06-eventbridge-rules.yaml"
  "07-monitoring.yaml"
)

print_color $CYAN "LogGuardian CloudFormation Template Validation"
print_color $CYAN "================================================="

# Check if AWS CLI is available
if ! command -v aws &> /dev/null; then
    print_color $RED "[ERROR] AWS CLI not found. Please install AWS CLI to validate templates."
    exit 1
fi

# Check if jq is available for JSON processing
if ! command -v jq &> /dev/null; then
    print_color $YELLOW "[WARN] jq not found. JSON validation will be limited."
    JQ_AVAILABLE=false
else
    JQ_AVAILABLE=true
fi

VALIDATION_ERRORS=0
VALIDATION_WARNINGS=0

# Function to validate a single template
validate_template() {
    local template_file=$1
    local template_path="$SCRIPT_DIR/$template_file"
    
    
    print_color $CYAN "[INFO] Validating: $template_file"
    
    if [[ ! -f "$template_path" ]]; then
        print_color $RED "[ERROR] Template file not found: $template_path"
        ((VALIDATION_ERRORS++))
        return 1
    fi
    
    # AWS CloudFormation syntax validation
    if aws cloudformation validate-template --template-body "file://$template_path" > /dev/null 2>&1; then
        print_color $GREEN "[PASS] CloudFormation syntax valid"
    else
        print_color $RED "[FAIL] CloudFormation syntax validation failed"
        aws cloudformation validate-template --template-body "file://$template_path" 2>&1 || true
        ((VALIDATION_ERRORS++))
        return 1
    fi
    
    # Check for common best practices
    check_best_practices "$template_path"
    
    return 0
}

# Function to check CloudFormation best practices
check_best_practices() {
    local template_path=$1
    local template_content=$(cat "$template_path")
    
    # Check for hardcoded values that should be parameters
    if echo "$template_content" | grep -q "ca-central-1\|ca-west-1\|eu-central-1" && ! echo "$template_content" | grep -q "\${AWS::Region}"; then
        print_color $YELLOW "[WARN] Template may contain hardcoded regions"
        ((VALIDATION_WARNINGS++))
    fi
    
    # Check for Description field
    if ! echo "$template_content" | grep -q "^Description:"; then
        print_color $YELLOW "[WARN] Template missing Description field"
        ((VALIDATION_WARNINGS++))
    fi
    
    # Check for proper resource naming
    if echo "$template_content" | grep -q "Type: AWS::" && ! echo "$template_content" | grep -q "Properties:"; then
        print_color $YELLOW "[WARN] Resources may be missing Properties"
        ((VALIDATION_WARNINGS++))
    fi
    
    # Check for tags on resources
    if echo "$template_content" | grep -q "Type: AWS::" && ! echo "$template_content" | grep -q "Tags:"; then
        print_color $YELLOW "[INFO] Consider adding tags to resources for better organization"
    fi
    
    print_color $GREEN "[PASS] Best practices check completed"
}

# Function to validate parameter files
validate_parameters() {
    local param_dir="$SCRIPT_DIR/parameters"
    
    if [[ ! -d "$param_dir" ]]; then
        print_color $YELLOW "[WARN] Parameters directory not found: $param_dir"
        return 0
    fi
    
    
    print_color $CYAN "Validating parameter files"
    
    for param_file in "$param_dir"/*.json; do
        if [[ -f "$param_file" ]]; then
            local filename=$(basename "$param_file")
            print_color $CYAN "  [INFO] Validating: $filename"
            
            if $JQ_AVAILABLE; then
                if jq empty "$param_file" 2>/dev/null; then
                    print_color $GREEN "  [PASS] JSON syntax valid"
                    
                    # Check parameter structure
                    if jq -e '.[0] | has("ParameterKey") and has("ParameterValue")' "$param_file" > /dev/null 2>&1; then
                        print_color $GREEN "  [PASS] Parameter structure valid"
                    else
                        print_color $RED "  [FAIL] Invalid parameter structure"
                        ((VALIDATION_ERRORS++))
                    fi
                else
                    print_color $RED "  [FAIL] Invalid JSON syntax"
                    ((VALIDATION_ERRORS++))
                fi
            else
                # Basic JSON validation without jq
                if python3 -m json.tool "$param_file" > /dev/null 2>&1; then
                    print_color $GREEN "  [PASS] JSON syntax valid"
                elif python -m json.tool "$param_file" > /dev/null 2>&1; then
                    print_color $GREEN "  [PASS] JSON syntax valid"
                else
                    print_color $RED "  [FAIL] Invalid JSON syntax"
                    ((VALIDATION_ERRORS++))
                fi
            fi
        fi
    done
}

# Function to check template relationships
check_template_relationships() {
    
    print_color $CYAN "Checking template relationships"
    
    # Check that main template references exist
    local main_template="$SCRIPT_DIR/01-logguardian-main.yaml"
    if [[ -f "$main_template" ]]; then
        local main_content=$(cat "$main_template")
        
        # Check nested template references
        for template in "kms-key.yaml" "iam-roles.yaml" "lambda-function.yaml" "config-rules.yaml" "eventbridge-rules.yaml" "monitoring.yaml"; do
            if echo "$main_content" | grep -q "$template"; then
                if [[ -f "$SCRIPT_DIR/$template" ]]; then
                    print_color $GREEN "[PASS] Referenced template exists: $template"
                else
                    print_color $RED "[FAIL] Referenced template missing: $template"
                    ((VALIDATION_ERRORS++))
                fi
            fi
        done
    fi
}

# Main validation process
main() {
    print_color $CYAN "Starting validation of ${#TEMPLATES[@]} templates...\\n"
    
    # Validate each template
    for template in "${TEMPLATES[@]}"; do
        validate_template "$template"
    done
    
    # Validate parameter files
    validate_parameters
    
    # Check template relationships
    check_template_relationships
    
    # Summary
    print_color $CYAN "\\nValidation Summary"
    print_color $CYAN "====================="
    
    if [[ $VALIDATION_ERRORS -eq 0 ]]; then
        print_color $GREEN "[SUCCESS] All templates passed validation!"
    else
        print_color $RED "[ERROR] $VALIDATION_ERRORS validation error(s) found"
    fi
    
    if [[ $VALIDATION_WARNINGS -gt 0 ]]; then
        print_color $YELLOW "[WARN] $VALIDATION_WARNINGS warning(s) found"
    fi
    
    print_color $CYAN "\\nTips:"
    print_color $CYAN "  • Run './90-validate-templates.sh' to validate all templates"
    print_color $CYAN "  • Use 'aws cloudformation estimate-cost' to estimate deployment costs"
    print_color $CYAN "  • Test in development environment before production deployment"
    
    # Exit with error code if validation failed
    if [[ $VALIDATION_ERRORS -gt 0 ]]; then
        exit 1
    fi
}

# Run main function
main "$@"
