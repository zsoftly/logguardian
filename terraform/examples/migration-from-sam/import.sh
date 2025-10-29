#!/bin/bash
# Import existing SAM resources into Terraform state
# Run this AFTER: terraform init && terraform plan

set -e

ENVIRONMENT=${1:-prod}
AWS_REGION=${2:-ca-central-1}

echo "Ì¥Ñ Importing SAM resources into Terraform..."
echo "Environment: $ENVIRONMENT"
echo "Region: $AWS_REGION"
echo ""

# Function to import resource
import_resource() {
    local resource_address=$1
    local resource_id=$2
    
    echo "Importing: $resource_address"
    if terraform import "$resource_address" "$resource_id" 2>/dev/null; then
        echo "‚úÖ Imported: $resource_address"
    else
        echo "‚ö†Ô∏è  Skipped: $resource_address (may not exist or already imported)"
    fi
    echo ""
}

# Get AWS Account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# Import KMS Key (if SAM created it)
read -p "Did SAM create the KMS key? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    KMS_KEY_ID=$(aws kms list-aliases --query "Aliases[?AliasName=='alias/logguardian-cloudwatch-logs'].TargetKeyId" --output text)
    if [ -n "$KMS_KEY_ID" ]; then
        import_resource "module.logguardian.module.kms.aws_kms_key.logguardian[0]" "$KMS_KEY_ID"
        import_resource "module.logguardian.module.kms.aws_kms_alias.logguardian[0]" "alias/logguardian-cloudwatch-logs"
    fi
fi

# Import Lambda Function
LAMBDA_NAME="logguardian-compliance-${ENVIRONMENT}"
import_resource "module.logguardian.module.lambda.aws_lambda_function.logguardian" "$LAMBDA_NAME"

# Import Lambda Log Group
LOG_GROUP="/aws/lambda/$LAMBDA_NAME"
import_resource "module.logguardian.module.lambda.aws_cloudwatch_log_group.lambda" "$LOG_GROUP"

# Import IAM Role
IAM_ROLE="LogGuardian-LambdaRole-${ENVIRONMENT}"
import_resource "module.logguardian.module.iam.aws_iam_role.lambda_execution" "$IAM_ROLE"

# Import EventBridge Rules (if SAM created them)
read -p "Did SAM create EventBridge rules? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    import_resource "module.logguardian.module.eventbridge.aws_cloudwatch_event_rule.encryption[0]" "logguardian-encryption-schedule-${ENVIRONMENT}"
    import_resource "module.logguardian.module.eventbridge.aws_cloudwatch_event_rule.retention[0]" "logguardian-retention-schedule-${ENVIRONMENT}"
fi

# Import Config Rules (if SAM created them)
read -p "Did SAM create Config rules? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    import_resource "module.logguardian.module.config.aws_config_config_rule.encryption[0]" "logguardian-encryption-${ENVIRONMENT}"
    import_resource "module.logguardian.module.config.aws_config_config_rule.retention[0]" "logguardian-retention-${ENVIRONMENT}"
fi

echo ""
echo "‚úÖ Import complete!"
echo ""
echo "Next steps:"
echo "1. Run: terraform plan"
echo "2. Verify no changes are needed"
echo "3. If clean: SAM ‚Üí Terraform migration successful!"
echo "4. Delete SAM stack: aws cloudformation delete-stack --stack-name logguardian-${ENVIRONMENT}"
