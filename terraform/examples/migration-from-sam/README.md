# Migrate from SAM to Terraform

Step-by-step guide to migrate an existing SAM deployment to Terraform.

## Overview

This example helps you transition from SAM to Terraform **without recreating resources**.

## Migration Strategy

1. ✅ Import existing SAM resources into Terraform state
2. ✅ Verify Terraform matches existing infrastructure
3. ✅ Manage with Terraform going forward
4. ✅ Delete SAM stack (Terraform now owns resources)

## Prerequisites

- Existing SAM deployment of LogGuardian
- SAM stack name (e.g., `logguardian-prod`)
- Terraform >= 1.5.0
- AWS CLI configured

## Migration Steps

### Step 1: Identify SAM Resources
```bash
# List SAM stack resources
aws cloudformation describe-stack-resources \
  --stack-name logguardian-prod \
  --query 'StackResources[].{Type:ResourceType,ID:PhysicalResourceId}' \
  --output table

# Note which resources were created by SAM
```

### Step 2: Configure Terraform
```bash
# Copy example config
cp terraform.tfvars.example terraform.tfvars

# Edit to match SAM deployment
nano terraform.tfvars

# Set flags based on what SAM created:
# - sam_created_kms=true if SAM created KMS key
# - sam_created_config=true if SAM created Config service
# etc.
```

### Step 3: Initialize Terraform
```bash
terraform init
terraform plan  # Will show resources to CREATE
```

### Step 4: Import Existing Resources
```bash
# Run import script
./import.sh prod ca-central-1

# Follow prompts to import resources
```

### Step 5: Verify No Changes
```bash
terraform plan

# Should show:
# "No changes. Your infrastructure matches the configuration."
```

### Step 6: Delete SAM Stack
```bash
# Now that Terraform manages resources, delete SAM stack
aws cloudformation delete-stack --stack-name logguardian-prod

# Verify deletion (this removes the CloudFormation stack, not the resources)
aws cloudformation describe-stacks --stack-name logguardian-prod
```

## Troubleshooting

**Issue: "Resource already exists"**
```bash
# Import the resource first
terraform import module.logguardian.module.lambda.aws_lambda_function.logguardian logguardian-compliance-prod
```

**Issue: Plan shows changes after import**
```bash
# Compare Terraform config with SAM template
# Adjust terraform.tfvars to match SAM settings exactly
```

**Issue: Cannot delete SAM stack**
```bash
# Some resources might be "retained" in SAM
# Check CloudFormation console for DeletionPolicy
# These resources are safe - Terraform now manages them
```

## Rollback

If you need to rollback to SAM:
```bash
# 1. Remove Terraform state
rm -rf .terraform terraform.tfstate*

# 2. SAM stack still exists (resources unchanged)
# 3. Continue using SAM as before
```
