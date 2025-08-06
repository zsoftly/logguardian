#!/bin/bash

# LogGuardian Multi-Region Deployment Script
# Provides functions for deploying and managing LogGuardian across multiple regions
# Usage: source this script or run individual functions

set -e

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TEMPLATE_FILE="$PROJECT_ROOT/template.yaml"

# Default values
DEFAULT_ENVIRONMENT="sandbox"
DEFAULT_REGIONS=("ca-central-1" "ca-west-1")
DEFAULT_S3_EXPIRATION_DAYS=14
DEFAULT_LAMBDA_MEMORY=256
DEFAULT_LAMBDA_TIMEOUT=900
DEFAULT_RETENTION_DAYS=30

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  INFO:${NC} $1"
}

log_success() {
    echo -e "${GREEN}✅ SUCCESS:${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠️  WARNING:${NC} $1"
}

log_error() {
    echo -e "${RED}❌ ERROR:${NC} $1"
}

# Usage function
show_usage() {
    cat << EOF
LogGuardian Multi-Region Deployment Script

USAGE:
    $0 COMMAND [OPTIONS]

COMMANDS:
    deploy-single       Deploy to a single region
    deploy-multi        Deploy to multiple regions
    deploy-dev          Deploy development environment
    deploy-staging      Deploy staging environment  
    deploy-prod         Deploy production environment
    deploy-customer     Deploy with customer infrastructure options
    status              Show deployment status across regions
    cleanup             Clean up deployments
    validate            Validate template before deployment
    help                Show this help message

GLOBAL OPTIONS:
    -e, --environment ENV       Environment (dev/staging/prod/sandbox) [default: $DEFAULT_ENVIRONMENT]
    -r, --regions REGIONS       Comma-separated regions [default: ${DEFAULT_REGIONS[*]}]
    -t, --template FILE         Template file path [default: $TEMPLATE_FILE]
    -s, --stack-prefix PREFIX   Stack name prefix [default: logguardian]
    -v, --verbose               Verbose output
    -h, --help                  Show help

EXAMPLES:
    # Deploy to sandbox in two regions
    $0 deploy-multi -e sandbox -r ca-central-1,ca-west-1

    # Deploy production with custom settings
    $0 deploy-prod -r us-east-1,eu-west-1 --retention-days 2555

    # Deploy with customer's existing KMS key
    $0 deploy-customer -e prod -r ca-central-1 --existing-kms-key alias/customer-logs

    # Check deployment status
    $0 status -e prod

    # Clean up sandbox deployments
    $0 cleanup -e sandbox

EOF
}

# Parse command line arguments
parse_args() {
    COMMAND=""
    ENVIRONMENT="$DEFAULT_ENVIRONMENT"
    REGIONS_STRING=""
    TEMPLATE="$TEMPLATE_FILE"
    STACK_PREFIX="logguardian"
    VERBOSE=false
    
    # Deployment parameters
    S3_EXPIRATION_DAYS="$DEFAULT_S3_EXPIRATION_DAYS"
    LAMBDA_MEMORY="$DEFAULT_LAMBDA_MEMORY"
    LAMBDA_TIMEOUT="$DEFAULT_LAMBDA_TIMEOUT"
    RETENTION_DAYS="$DEFAULT_RETENTION_DAYS"
    CREATE_DASHBOARD="true"
    ENABLE_STAGGERED="true"
    
    # Customer infrastructure parameters
    EXISTING_KMS_KEY=""
    CREATE_KMS_KEY="true"
    EXISTING_CONFIG_RULE=""
    CREATE_CONFIG_RULES="true"
    CUSTOMER_TAG_PREFIX=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            deploy-single|deploy-multi|deploy-dev|deploy-staging|deploy-prod|deploy-customer|status|cleanup|validate|help)
                COMMAND="$1"
                shift
                ;;
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -r|--regions)
                REGIONS_STRING="$2"
                shift 2
                ;;
            -t|--template)
                TEMPLATE="$2"
                shift 2
                ;;
            -s|--stack-prefix)
                STACK_PREFIX="$2"
                shift 2
                ;;
            --s3-expiration-days)
                S3_EXPIRATION_DAYS="$2"
                shift 2
                ;;
            --lambda-memory)
                LAMBDA_MEMORY="$2"
                shift 2
                ;;
            --lambda-timeout)
                LAMBDA_TIMEOUT="$2"
                shift 2
                ;;
            --retention-days)
                RETENTION_DAYS="$2"
                shift 2
                ;;
            --create-dashboard)
                CREATE_DASHBOARD="$2"
                shift 2
                ;;
            --enable-staggered)
                ENABLE_STAGGERED="$2"
                shift 2
                ;;
            --existing-kms-key)
                EXISTING_KMS_KEY="$2"
                CREATE_KMS_KEY="false"
                shift 2
                ;;
            --existing-config-rule)
                EXISTING_CONFIG_RULE="$2"
                CREATE_CONFIG_RULES="false"
                shift 2
                ;;
            --customer-tag-prefix)
                CUSTOMER_TAG_PREFIX="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done

    # Set regions array
    if [ -n "$REGIONS_STRING" ]; then
        IFS=',' read -ra REGIONS <<< "$REGIONS_STRING"
    else
        REGIONS=("${DEFAULT_REGIONS[@]}")
    fi

    # Validate command
    if [ -z "$COMMAND" ]; then
        log_error "No command specified"
        show_usage
        exit 1
    fi
}

# Build Lambda function if needed
build_lambda() {
    log_info "Checking if Lambda build is needed..."
    
    if [ ! -f "$PROJECT_ROOT/build/bootstrap" ]; then
        log_info "Building Lambda function..."
        cd "$PROJECT_ROOT"
        make build
        log_success "Lambda function built"
    else
        log_info "Lambda function already built"
    fi
}

# Validate template
validate_template() {
    local template_file="${1:-$TEMPLATE}"
    
    log_info "Validating SAM template: $template_file"
    
    if [ ! -f "$template_file" ]; then
        log_error "Template file not found: $template_file"
        exit 1
    fi
    
    sam validate --template "$template_file" --region "${REGIONS[0]}"
    log_success "Template validation passed"
}

# Build parameter overrides
build_parameter_overrides() {
    local params="Environment=$ENVIRONMENT"
    params="$params S3ExpirationDays=$S3_EXPIRATION_DAYS"
    params="$params CreateMonitoringDashboard=$CREATE_DASHBOARD"
    params="$params EnableStaggeredScheduling=$ENABLE_STAGGERED"
    params="$params DefaultRetentionDays=$RETENTION_DAYS"
    params="$params LambdaMemorySize=$LAMBDA_MEMORY"
    params="$params LambdaTimeout=$LAMBDA_TIMEOUT"
    
    # Add KMS key parameters if specified
    if [ -n "$EXISTING_KMS_KEY" ]; then
        params="$params KMSKeyAlias=$EXISTING_KMS_KEY"
    fi
    
    # Add customer-specific parameters
    if [ -n "$CUSTOMER_TAG_PREFIX" ]; then
        params="$params CustomerTagPrefix=$CUSTOMER_TAG_PREFIX"
    fi
    
    echo "$params"
}

# Deploy to a single region
deploy_to_region() {
    local region="$1"
    local stack_name="$2"
    local confirm_changeset="${3:-true}"
    
    log_info "Deploying to region: $region"
    log_info "Stack name: $stack_name"
    
    local parameter_overrides=$(build_parameter_overrides)
    local confirm_flag=""
    
    if [ "$confirm_changeset" = "false" ]; then
        confirm_flag="--no-confirm-changeset"
    fi
    
    if [ "$VERBOSE" = "true" ]; then
        log_info "Parameter overrides: $parameter_overrides"
    fi
    
    sam deploy \
        --template-file "$TEMPLATE" \
        --stack-name "$stack_name" \
        --parameter-overrides $parameter_overrides \
        --capabilities CAPABILITY_IAM \
        --region "$region" \
        $confirm_flag
        
    log_success "Successfully deployed to $region"
}

# Deploy to multiple regions
deploy_multi_region() {
    log_info "Starting multi-region deployment"
    log_info "Environment: $ENVIRONMENT"
    log_info "Regions: ${REGIONS[*]}"
    
    build_lambda
    validate_template
    
    for region in "${REGIONS[@]}"; do
        local stack_name="$STACK_PREFIX-$ENVIRONMENT-$region"
        echo "======================================"
        deploy_to_region "$region" "$stack_name" "false"
        echo ""
    done
    
    log_success "Multi-region deployment complete!"
    echo ""
    show_deployment_status
}

# Deploy single region
deploy_single_region() {
    local region="${REGIONS[0]}"
    local stack_name="$STACK_PREFIX-$ENVIRONMENT-$region"
    
    log_info "Starting single region deployment"
    log_info "Environment: $ENVIRONMENT"
    log_info "Region: $region"
    
    build_lambda
    validate_template
    
    deploy_to_region "$region" "$stack_name" "true"
    
    echo ""
    show_deployment_status
}

# Development environment deployment
deploy_dev_environment() {
    ENVIRONMENT="dev"
    S3_EXPIRATION_DAYS=3
    RETENTION_DAYS=7
    CREATE_DASHBOARD="false"
    LAMBDA_MEMORY=128
    
    log_info "Deploying development environment with optimized settings"
    deploy_multi_region
}

# Staging environment deployment
deploy_staging_environment() {
    ENVIRONMENT="staging"
    S3_EXPIRATION_DAYS=14
    RETENTION_DAYS=30
    CREATE_DASHBOARD="true"
    LAMBDA_MEMORY=256
    
    log_info "Deploying staging environment"
    deploy_multi_region
}

# Production environment deployment
deploy_prod_environment() {
    ENVIRONMENT="prod"
    S3_EXPIRATION_DAYS=90
    RETENTION_DAYS=365
    CREATE_DASHBOARD="true"
    LAMBDA_MEMORY=512
    LAMBDA_TIMEOUT=900
    
    log_info "Deploying production environment with full monitoring"
    deploy_multi_region
}

# Deploy with customer infrastructure options
deploy_customer_infrastructure() {
    log_info "Deploying with customer infrastructure integration"
    
    if [ -n "$EXISTING_KMS_KEY" ]; then
        log_info "Using existing KMS key: $EXISTING_KMS_KEY"
    fi
    
    if [ -n "$EXISTING_CONFIG_RULE" ]; then
        log_info "Using existing Config rule: $EXISTING_CONFIG_RULE"
    fi
    
    deploy_multi_region
}

# Show deployment status
show_deployment_status() {
    log_info "Checking deployment status across regions"
    echo ""
    
    for region in "${REGIONS[@]}"; do
        local stack_name="$STACK_PREFIX-$ENVIRONMENT-$region"
        
        echo "Region: $region"
        echo "Stack: $stack_name"
        
        # Check if stack exists
        if aws cloudformation describe-stacks --stack-name "$stack_name" --region "$region" >/dev/null 2>&1; then
            # Get stack status
            local stack_status=$(aws cloudformation describe-stacks \
                --stack-name "$stack_name" \
                --region "$region" \
                --query 'Stacks[0].StackStatus' \
                --output text)
            
            echo "Status: $stack_status"
            
            # Get dashboard URL if it exists
            local dashboard_url=$(aws cloudformation describe-stacks \
                --stack-name "$stack_name" \
                --region "$region" \
                --query 'Stacks[0].Outputs[?OutputKey==`DashboardURL`].OutputValue' \
                --output text 2>/dev/null)
                
            if [ -n "$dashboard_url" ] && [ "$dashboard_url" != "None" ]; then
                echo "Dashboard: $dashboard_url"
            else
                echo "Dashboard: Not created"
            fi
        else
            echo "Status: NOT_DEPLOYED"
        fi
        
        echo ""
    done
}

# Clean up deployments
cleanup_deployments() {
    log_warning "This will delete LogGuardian stacks in environment: $ENVIRONMENT"
    log_warning "Regions: ${REGIONS[*]}"
    
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cleanup cancelled"
        exit 0
    fi
    
    for region in "${REGIONS[@]}"; do
        local stack_name="$STACK_PREFIX-$ENVIRONMENT-$region"
        
        log_info "Deleting stack: $stack_name in region: $region"
        
        if aws cloudformation describe-stacks --stack-name "$stack_name" --region "$region" >/dev/null 2>&1; then
            aws cloudformation delete-stack --stack-name "$stack_name" --region "$region"
            log_info "Delete initiated for $stack_name"
        else
            log_warning "Stack $stack_name not found in region $region"
        fi
    done
    
    log_success "Cleanup commands sent. Stacks are being deleted..."
    log_info "Use 'status' command to monitor deletion progress"
}

# Main execution
main() {
    parse_args "$@"
    
    case $COMMAND in
        deploy-single)
            deploy_single_region
            ;;
        deploy-multi)
            deploy_multi_region
            ;;
        deploy-dev)
            deploy_dev_environment
            ;;
        deploy-staging)
            deploy_staging_environment
            ;;
        deploy-prod)
            deploy_prod_environment
            ;;
        deploy-customer)
            deploy_customer_infrastructure
            ;;
        status)
            show_deployment_status
            ;;
        cleanup)
            cleanup_deployments
            ;;
        validate)
            validate_template
            ;;
        help)
            show_usage
            ;;
        *)
            log_error "Unknown command: $COMMAND"
            show_usage
            exit 1
            ;;
    esac
}

# Run main if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
