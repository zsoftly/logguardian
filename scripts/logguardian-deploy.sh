#!/bin/bash

# LogGuardian Production Deployment Script
# Optimized for single-region deployment with optional infrastructure support
# Usage: ./logguardian-deploy.sh [OPTIONS]

set -e

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TEMPLATE_FILE="$PROJECT_ROOT/template.yaml"

# Default values
DEFAULT_ENVIRONMENT="prod"
DEFAULT_REGION="ca-central-1"
DEFAULT_S3_EXPIRATION_DAYS=90
DEFAULT_LAMBDA_MEMORY=512
DEFAULT_LAMBDA_TIMEOUT=900
DEFAULT_RETENTION_DAYS=365

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
    cat << USAGE_EOF
LogGuardian Production Deployment Script

USAGE:
    $0 [OPTIONS]

COMMANDS:
    deploy              Deploy LogGuardian (default)
    status              Show deployment status
    cleanup             Clean up deployment
    validate            Validate template before deployment
    help                Show this help message

OPTIONS:
    -e, --environment ENV       Environment name [default: $DEFAULT_ENVIRONMENT]
    -r, --region REGION         AWS region [default: $DEFAULT_REGION]
    -s, --stack-name NAME       Stack name [default: logguardian-ENV-REGION]
    -t, --template FILE         Template file path [default: $TEMPLATE_FILE]
    -v, --verbose               Verbose output
    -h, --help                  Show help

INFRASTRUCTURE OPTIONS:
    --create-kms-key BOOL           Create new KMS key [default: true]
    --existing-kms-key-arn ARN      Use existing KMS key ARN
    --create-config-service BOOL    Create AWS Config service [default: true]
    --existing-config-bucket NAME   Use existing Config bucket
    --existing-config-role-arn ARN  Use existing Config service role ARN
    --create-config-rules BOOL      Create Config rules [default: true]
    --existing-encryption-rule NAME Use existing encryption rule
    --existing-retention-rule NAME  Use existing retention rule
    --create-eventbridge BOOL      Create EventBridge rules [default: true]
    --create-dashboard BOOL         Create CloudWatch dashboard [default: true]

CONFIGURATION OPTIONS:
    --retention-days DAYS           Log retention in days [default: $DEFAULT_RETENTION_DAYS]
    --s3-expiration-days DAYS       S3 data expiration days [default: $DEFAULT_S3_EXPIRATION_DAYS]
    --lambda-memory MB              Lambda memory size [default: $DEFAULT_LAMBDA_MEMORY]
    --lambda-timeout SEC            Lambda timeout [default: $DEFAULT_LAMBDA_TIMEOUT]
    --enable-staggered BOOL         Enable staggered scheduling [default: true]

TAGGING OPTIONS:
    --product-name NAME             Product name for tagging [default: LogGuardian]
    --customer-tag-prefix NAME      Customer resource tag prefix
    --owner NAME                    Owner/Team name [default: ZSoftly]
    --managed-by TOOL               Management tool [default: SAM]

EXAMPLES:
    # Basic deployment to production
    $0 deploy -e prod -r ca-central-1

    # Enterprise deployment with existing infrastructure
    $0 deploy -e prod -r us-east-1 \\
      --create-kms-key false \\
      --existing-kms-key-arn arn:aws:kms:us-east-1:ACCOUNT:key/KEY-ID \\
      --create-config-service false \\
      --existing-config-bucket enterprise-config \\
      --existing-config-role-arn arn:aws:iam::ACCOUNT:role/ConfigRole \\
      --customer-tag-prefix "Enterprise-LogGuardian" \\
      --owner "Enterprise-Security"

    # Development deployment with minimal infrastructure
    $0 deploy -e dev -r ca-central-1 \\
      --retention-days 7 \\
      --s3-expiration-days 3 \\
      --lambda-memory 256 \\
      --create-dashboard false

    # Check deployment status
    $0 status -e prod -r ca-central-1

    # Clean up deployment
    $0 cleanup -e dev -r ca-central-1

USAGE_EOF
}

# Parse command line arguments
parse_args() {
    COMMAND="deploy"
    ENVIRONMENT="$DEFAULT_ENVIRONMENT"
    REGION="$DEFAULT_REGION"
    STACK_NAME=""
    TEMPLATE="$TEMPLATE_FILE"
    VERBOSE=false
    
    # Infrastructure parameters
    CREATE_KMS_KEY="true"
    EXISTING_KMS_KEY_ARN=""
    CREATE_CONFIG_SERVICE="true"
    EXISTING_CONFIG_BUCKET=""
    EXISTING_CONFIG_SERVICE_ROLE_ARN=""
    CREATE_CONFIG_RULES="true"
    EXISTING_ENCRYPTION_CONFIG_RULE=""
    EXISTING_RETENTION_CONFIG_RULE=""
    CREATE_EVENTBRIDGE_RULES="true"
    CREATE_MONITORING_DASHBOARD="true"
    
    # Configuration parameters
    S3_EXPIRATION_DAYS="$DEFAULT_S3_EXPIRATION_DAYS"
    LAMBDA_MEMORY="$DEFAULT_LAMBDA_MEMORY"
    LAMBDA_TIMEOUT="$DEFAULT_LAMBDA_TIMEOUT"
    DEFAULT_RETENTION_DAYS_VAL="$DEFAULT_RETENTION_DAYS"
    ENABLE_STAGGERED_SCHEDULING="true"
    
    # Tagging parameters
    PRODUCT_NAME="LogGuardian"
    CUSTOMER_TAG_PREFIX=""
    OWNER="ZSoftly"
    MANAGED_BY="SAM"

    while [[ $# -gt 0 ]]; do
        case $1 in
            deploy|status|cleanup|validate|help)
                COMMAND="$1"
                shift
                ;;
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -r|--region)
                REGION="$2"
                shift 2
                ;;
            -s|--stack-name)
                STACK_NAME="$2"
                shift 2
                ;;
            -t|--template)
                TEMPLATE="$2"
                shift 2
                ;;
            --create-kms-key)
                CREATE_KMS_KEY="$2"
                shift 2
                ;;
            --existing-kms-key-arn)
                EXISTING_KMS_KEY_ARN="$2"
                CREATE_KMS_KEY="false"
                shift 2
                ;;
            --create-config-service)
                CREATE_CONFIG_SERVICE="$2"
                shift 2
                ;;
            --existing-config-bucket)
                EXISTING_CONFIG_BUCKET="$2"
                shift 2
                ;;
            --existing-config-role-arn)
                EXISTING_CONFIG_SERVICE_ROLE_ARN="$2"
                shift 2
                ;;
            --create-config-rules)
                CREATE_CONFIG_RULES="$2"
                shift 2
                ;;
            --existing-encryption-rule)
                EXISTING_ENCRYPTION_CONFIG_RULE="$2"
                shift 2
                ;;
            --existing-retention-rule)
                EXISTING_RETENTION_CONFIG_RULE="$2"
                shift 2
                ;;
            --create-eventbridge)
                CREATE_EVENTBRIDGE_RULES="$2"
                shift 2
                ;;
            --create-dashboard)
                CREATE_MONITORING_DASHBOARD="$2"
                shift 2
                ;;
            --retention-days)
                DEFAULT_RETENTION_DAYS_VAL="$2"
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
            --enable-staggered)
                ENABLE_STAGGERED_SCHEDULING="$2"
                shift 2
                ;;
            --product-name)
                PRODUCT_NAME="$2"
                shift 2
                ;;
            --customer-tag-prefix)
                CUSTOMER_TAG_PREFIX="$2"
                shift 2
                ;;
            --owner)
                OWNER="$2"
                shift 2
                ;;
            --managed-by)
                MANAGED_BY="$2"
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

    # Set default stack name if not provided
    if [ -z "$STACK_NAME" ]; then
        STACK_NAME="logguardian-$ENVIRONMENT-$REGION"
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
    
    sam validate --template "$template_file" --region "$REGION"
    log_success "Template validation passed"
}

# Build parameter overrides for optimized template
build_parameter_overrides() {
    local params="Environment=$ENVIRONMENT"
    
    # Infrastructure parameters
    params="$params CreateKMSKey=$CREATE_KMS_KEY"
    if [ -n "$EXISTING_KMS_KEY_ARN" ]; then
        params="$params ExistingKMSKeyArn=$EXISTING_KMS_KEY_ARN"
    fi
    
    params="$params CreateConfigService=$CREATE_CONFIG_SERVICE"
    if [ -n "$EXISTING_CONFIG_BUCKET" ]; then
        params="$params ExistingConfigBucket=$EXISTING_CONFIG_BUCKET"
    fi
    if [ -n "$EXISTING_CONFIG_SERVICE_ROLE_ARN" ]; then
        params="$params ExistingConfigServiceRoleArn=$EXISTING_CONFIG_SERVICE_ROLE_ARN"
    fi
    
    params="$params CreateConfigRules=$CREATE_CONFIG_RULES"
    if [ -n "$EXISTING_ENCRYPTION_CONFIG_RULE" ]; then
        params="$params ExistingEncryptionConfigRule=$EXISTING_ENCRYPTION_CONFIG_RULE"
    fi
    if [ -n "$EXISTING_RETENTION_CONFIG_RULE" ]; then
        params="$params ExistingRetentionConfigRule=$EXISTING_RETENTION_CONFIG_RULE"
    fi
    
    params="$params CreateEventBridgeRules=$CREATE_EVENTBRIDGE_RULES"
    params="$params CreateMonitoringDashboard=$CREATE_MONITORING_DASHBOARD"
    
    # Configuration parameters
    params="$params S3ExpirationDays=$S3_EXPIRATION_DAYS"
    params="$params LambdaMemorySize=$LAMBDA_MEMORY"
    params="$params LambdaTimeout=$LAMBDA_TIMEOUT"
    params="$params DefaultRetentionDays=$DEFAULT_RETENTION_DAYS_VAL"
    params="$params EnableStaggeredScheduling=$ENABLE_STAGGERED_SCHEDULING"
    
    # Tagging parameters
    local product_name="$PRODUCT_NAME"
    if [ -n "$CUSTOMER_TAG_PREFIX" ]; then
        params="$params CustomerTagPrefix=$CUSTOMER_TAG_PREFIX"
        product_name="$CUSTOMER_TAG_PREFIX"
    fi
    
    params="$params ProductName=$product_name"
    params="$params Owner=$OWNER"
    params="$params ManagedBy=$MANAGED_BY"
    
    echo "$params"
}

# Deploy LogGuardian
deploy() {
    log_info "Starting LogGuardian deployment"
    log_info "Environment: $ENVIRONMENT"
    log_info "Region: $REGION"
    log_info "Stack name: $STACK_NAME"
    
    build_lambda
    validate_template
    
    local parameter_overrides=$(build_parameter_overrides)
    local stack_tags=""
    
    # Build stack-level tags following AWS best practices
    local product_name="$PRODUCT_NAME"
    if [ -n "$CUSTOMER_TAG_PREFIX" ]; then
        product_name="$CUSTOMER_TAG_PREFIX"
    fi
    
    # Get version from VERSION file (required)
    if [ ! -f "${SCRIPT_DIR}/../VERSION" ]; then
        log_error "VERSION file not found. Cannot determine application version."
        exit 1
    fi
    version=$(cat "${SCRIPT_DIR}/../VERSION")
    stack_tags="Product=$product_name Owner=$OWNER Environment=$ENVIRONMENT ManagedBy=$MANAGED_BY Application=LogGuardian Version=$version CreatedBy=SAM-Deploy"
    
    if [ "$VERBOSE" = "true" ]; then
        log_info "Parameter overrides: $parameter_overrides"
        log_info "Stack tags: $stack_tags"
    fi
    
    log_info "Deploying to AWS..."
    sam deploy \
        --template-file "$TEMPLATE" \
        --stack-name "$STACK_NAME" \
        --parameter-overrides $parameter_overrides \
        --tags "$stack_tags" \
        --capabilities CAPABILITY_NAMED_IAM \
        --region "$REGION" \
        --no-confirm-changeset
        
    log_success "Successfully deployed LogGuardian to $REGION"
    echo ""
    show_deployment_status
}

# Show deployment status
show_deployment_status() {
    log_info "Checking deployment status"
    echo ""
    
    echo "Region: $REGION"
    echo "Stack: $STACK_NAME"
    
    # Check if stack exists
    if aws cloudformation describe-stacks --stack-name "$STACK_NAME" --region "$REGION" >/dev/null 2>&1; then
        # Get stack status
        local stack_status=$(aws cloudformation describe-stacks \
            --stack-name "$STACK_NAME" \
            --region "$REGION" \
            --query 'Stacks[0].StackStatus' \
            --output text)
        
        echo "Status: $stack_status"
        
        # Get outputs
        local outputs=$(aws cloudformation describe-stacks \
            --stack-name "$STACK_NAME" \
            --region "$REGION" \
            --query 'Stacks[0].Outputs' \
            --output table 2>/dev/null)
            
        if [ -n "$outputs" ] && [ "$outputs" != "None" ]; then
            echo ""
            echo "Stack Outputs:"
            echo "$outputs"
        fi
    else
        echo "Status: NOT_DEPLOYED"
    fi
    
    echo ""
}

# Clean up deployment
cleanup() {
    log_warning "This will delete the LogGuardian stack: $STACK_NAME"
    log_warning "Region: $REGION"
    
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cleanup cancelled"
        exit 0
    fi
    
    log_info "Deleting stack: $STACK_NAME in region: $REGION"
    
    if aws cloudformation describe-stacks --stack-name "$STACK_NAME" --region "$REGION" >/dev/null 2>&1; then
        aws cloudformation delete-stack --stack-name "$STACK_NAME" --region "$REGION"
        log_info "Delete initiated for $STACK_NAME"
        log_success "Cleanup command sent. Stack is being deleted..."
        log_info "Use 'status' command to monitor deletion progress"
    else
        log_warning "Stack $STACK_NAME not found in region $REGION"
    fi
}

# Main execution
main() {
    parse_args "$@"
    
    case $COMMAND in
        deploy)
            deploy
            ;;
        status)
            show_deployment_status
            ;;
        cleanup)
            cleanup
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
