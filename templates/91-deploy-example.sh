#!/bin/bash

# LogGuardian Deployment Example Script
# Demonstrates how to deploy LogGuardian using CloudFormation templates

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Configuration
DEPLOYMENT_BUCKET="${DEPLOYMENT_BUCKET:-my-logguardian-deployment-bucket}"
AWS_REGION="${AWS_REGION:-ca-central-1}"
STACK_NAME_PREFIX="logguardian"

# Help function
show_help() {
    cat << EOF
LogGuardian Deployment Script

Usage: $0 [OPTIONS] ENVIRONMENT

ENVIRONMENT:
    dev, staging, prod

OPTIONS:
    -b, --bucket BUCKET     S3 bucket for deployment artifacts (default: $DEPLOYMENT_BUCKET)
    -r, --region REGION     AWS region (default: $AWS_REGION)
    -t, --template TYPE     Template type: simple or full (default: simple)
    -s, --stack-name NAME   Custom stack name prefix (default: $STACK_NAME_PREFIX)
    -h, --help             Show this help message

EXAMPLES:
    # Simple deployment to dev
    $0 dev

    # Full deployment to production
    $0 -t full prod

    # Custom bucket and region
    $0 -b my-bucket -r ca-west-1 staging

ENVIRONMENT VARIABLES:
    DEPLOYMENT_BUCKET       S3 bucket for deployment artifacts
    AWS_REGION             Target AWS region

EOF
}

# Parse command line arguments
TEMPLATE_TYPE="simple"
ENVIRONMENT=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -b|--bucket)
            DEPLOYMENT_BUCKET="$2"
            shift 2
            ;;
        -r|--region)
            AWS_REGION="$2"
            shift 2
            ;;
        -t|--template)
            TEMPLATE_TYPE="$2"
            shift 2
            ;;
        -s|--stack-name)
            STACK_NAME_PREFIX="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        dev|staging|prod)
            ENVIRONMENT="$1"
            shift
            ;;
        *)
            print_color $RED "[ERROR] Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Validate required parameters
if [[ -z "$ENVIRONMENT" ]]; then
    print_color $RED "[ERROR] Environment is required (dev, staging, prod)"
    show_help
    exit 1
fi

if [[ ! "$ENVIRONMENT" =~ ^(dev|staging|prod)$ ]]; then
    print_color $RED "[ERROR] Invalid environment. Must be dev, staging, or prod"
    exit 1
fi

if [[ ! "$TEMPLATE_TYPE" =~ ^(simple|full)$ ]]; then
    print_color $RED "[ERROR] Invalid template type. Must be simple or full"
    exit 1
fi

# Set derived variables
STACK_NAME="${STACK_NAME_PREFIX}-${ENVIRONMENT}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
TEMPLATES_DIR="$SCRIPT_DIR"

print_color $CYAN "LogGuardian Deployment"
print_color $CYAN "========================="
print_color $CYAN "Environment: $ENVIRONMENT"
print_color $CYAN "Template Type: $TEMPLATE_TYPE"
print_color $CYAN "Stack Name: $STACK_NAME"
print_color $CYAN "Deployment Bucket: $DEPLOYMENT_BUCKET"
print_color $CYAN "AWS Region: $AWS_REGION"
echo

# Check prerequisites
check_prerequisites() {
    print_color $CYAN "🔍 Checking prerequisites..."
    
    # Check AWS CLI
    if ! command -v aws &> /dev/null; then
        print_color $RED "[ERROR] AWS CLI not found. Please install AWS CLI."
        exit 1
    fi
    
    # Check AWS credentials
    if ! aws sts get-caller-identity > /dev/null 2>&1; then
        print_color $RED "[ERROR] AWS credentials not configured. Please run 'aws configure'."
        exit 1
    fi
    
    # Check if deployment bucket exists
    if ! aws s3 ls "s3://$DEPLOYMENT_BUCKET" > /dev/null 2>&1; then
        print_color $YELLOW "[WARN]  Deployment bucket doesn't exist. Creating..."
        if aws s3 mb "s3://$DEPLOYMENT_BUCKET" --region "$AWS_REGION"; then
            print_color $GREEN "[PASS] Created deployment bucket: $DEPLOYMENT_BUCKET"
        else
            print_color $RED "[ERROR] Failed to create deployment bucket"
            exit 1
        fi
    else
        print_color $GREEN "[PASS] Deployment bucket exists: $DEPLOYMENT_BUCKET"
    fi
    
    # Check if Lambda package exists
    local lambda_package="$SCRIPT_DIR/../dist/logguardian-compliance.zip"
    if [[ ! -f "$lambda_package" ]]; then
        print_color $YELLOW "[WARN]  Lambda package not found. Building..."
        cd "$SCRIPT_DIR/.."
        make package
        cd "$TEMPLATES_DIR"
    fi
    
    print_color $GREEN "[PASS] Prerequisites check completed"
}

# Upload artifacts to S3
upload_artifacts() {
    print_color $CYAN "📤 Uploading deployment artifacts..."
    
    # Upload Lambda package
    local lambda_package="$SCRIPT_DIR/../dist/logguardian-compliance.zip"
    if aws s3 cp "$lambda_package" "s3://$DEPLOYMENT_BUCKET/logguardian-compliance.zip"; then
        print_color $GREEN "[PASS] Uploaded Lambda package"
    else
        print_color $RED "[ERROR] Failed to upload Lambda package"
        exit 1
    fi
    
    # Upload templates for full deployment
    if [[ "$TEMPLATE_TYPE" == "full" ]]; then
        if aws s3 sync "$TEMPLATES_DIR/" "s3://$DEPLOYMENT_BUCKET/templates/" --exclude "*.sh" --exclude "README.md"; then
            print_color $GREEN "[PASS] Uploaded CloudFormation templates"
        else
            print_color $RED "[ERROR] Failed to upload templates"
            exit 1
        fi
    fi
}

# Deploy CloudFormation stack
deploy_stack() {
    print_color $CYAN "Deploying CloudFormation stack..."
    
    local template_file
    local parameters_file="$TEMPLATES_DIR/parameters/${ENVIRONMENT}-parameters.json"
    
    if [[ "$TEMPLATE_TYPE" == "simple" ]]; then
        template_file="$TEMPLATES_DIR/00-logguardian-simple.yaml"
    else
        template_file="$TEMPLATES_DIR/01-logguardian-main.yaml"
    fi
    
    # Update parameters file with deployment bucket
    local temp_params=$(mktemp)
    if ! command -v jq &> /dev/null; then
        print_color $RED "[ERROR] 'jq' is required but not installed. Please install jq."
        exit 1
    fi
    if ! jq --arg bucket "$DEPLOYMENT_BUCKET" '
        map(if .ParameterKey == "DeploymentBucket" then .ParameterValue = $bucket else . end) |
        map(if .ParameterKey == "TemplatesBucket" then .ParameterValue = $bucket else . end)
    ' "$parameters_file" > "$temp_params"; then
        print_color $RED "[ERROR] Failed to process parameters file with jq. Please check your JSON syntax."
        rm -f "$temp_params"
        exit 1
    fi
    
    print_color $CYAN "Template: $(basename "$template_file")"
    print_color $CYAN "Parameters: $(basename "$parameters_file")"
    
    # Check if stack exists
    if aws cloudformation describe-stacks --stack-name "$STACK_NAME" --region "$AWS_REGION" > /dev/null 2>&1; then
        print_color $YELLOW "[WARN]  Stack exists. Updating..."
        local action="update-stack"
    else
        print_color $CYAN "📋 Creating new stack..."
        local action="create-stack"
    fi
    
    # Deploy stack
    local stack_id
    if stack_id=$(aws cloudformation "$action" \
        --stack-name "$STACK_NAME" \
        --template-body "file://$template_file" \
        --parameters "file://$temp_params" \
        --capabilities CAPABILITY_NAMED_IAM \
        --region "$AWS_REGION" \
        --output text --query 'StackId' 2>/dev/null); then
        
        print_color $GREEN "[PASS] Stack deployment initiated: $stack_id"
        
        # Wait for stack completion
        print_color $CYAN "⏳ Waiting for stack deployment to complete..."
        if aws cloudformation wait "stack-${action%-stack}-complete" \
            --stack-name "$STACK_NAME" \
            --region "$AWS_REGION"; then
            print_color $GREEN "[PASS] Stack deployment completed successfully!"
        else
            print_color $RED "[ERROR] Stack deployment failed"
            # Show stack events for debugging
            aws cloudformation describe-stack-events \
                --stack-name "$STACK_NAME" \
                --region "$AWS_REGION" \
                --query 'StackEvents[?ResourceStatus==`CREATE_FAILED` || ResourceStatus==`UPDATE_FAILED`].[LogicalResourceId,ResourceStatusReason]' \
                --output table
            exit 1
        fi
    else
        print_color $RED "[ERROR] Failed to initiate stack deployment"
        exit 1
    fi
    
    # Clean up temp file
    rm -f "$temp_params"
}

# Show deployment outputs
show_outputs() {
    print_color $CYAN " Deployment Outputs:"
    
    aws cloudformation describe-stacks \
        --stack-name "$STACK_NAME" \
        --region "$AWS_REGION" \
        --query 'Stacks[0].Outputs[?OutputKey!=`null`].[OutputKey,OutputValue,Description]' \
        --output table
}

# Main deployment function
main() {
    check_prerequisites
    
    # Validate templates
    print_color $CYAN "🔍 Validating CloudFormation templates..."
    if "$TEMPLATES_DIR/90-validate-templates.sh"; then
        print_color $GREEN "[PASS] Template validation passed"
    else
        print_color $RED "[ERROR] Template validation failed"
        exit 1
    fi
    
    upload_artifacts
    deploy_stack
    show_outputs
    
    print_color $GREEN "\\n🎉 LogGuardian deployment completed successfully!"
    print_color $CYAN "\\n💡 Next steps:"
    print_color $CYAN "  • Check the CloudWatch dashboard for monitoring"
    print_color $CYAN "  • Review Lambda function logs for any issues"
    print_color $CYAN "  • Verify Config rules are running and detecting non-compliant resources"
    print_color $CYAN "  • Test the EventBridge schedule trigger"
}

# Run main function
main "$@"
