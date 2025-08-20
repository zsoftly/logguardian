# LogGuardian Resource Tagging Strategy

This document outlines the comprehensive tagging strategy for LogGuardian deployments to ensure proper resource management, cost tracking, and compliance.

## Standard Tags Applied to All Resources

| Tag Key | Description | Example Values | Source |
|---------|-------------|----------------|---------|
| `Product` | Product/service name | `LogGuardian`, `ACME-LogGuardian` | Parameter |
| `Owner` | Team/person responsible | `ZSoftly`, `Platform-Team`, `Security-Team` | Parameter |
| `Environment` | Deployment environment | `dev`, `staging`, `prod`, `sandbox` | Parameter |
| `ManagedBy` | Infrastructure management tool | `SAM`, `CloudFormation`, `Terraform` | Parameter |
| `Application` | Application name | `LogGuardian` | Fixed |
| `Version` | Application version | `<current-version>` | Template-driven |
| `CreatedBy` | Template that created resource | `SAM-Template` | Fixed |

## Tag Implementation

### Template Level (Global Tags)
```yaml
# In template.yaml - Applied to all resources
Globals:
  Tags:
    Product: !Ref ProductName
    Owner: !Ref Owner
    Environment: !Ref Environment
    ManagedBy: !Ref ManagedBy
    Application: LogGuardian
    Version: !Ref ApplicationVersion  # Dynamically set from template
    CreatedBy: "SAM-Template"
```

### Resource-Specific Tags
Individual resources may have additional specific tags:

```yaml
# KMS Key gets additional description
LogGuardianKMSKey:
  Type: AWS::KMS::Key
  Properties:
    Description: !Sub "${ProductName} KMS key for ${Environment} environment"
    # Global tags automatically applied

# Lambda Function gets runtime-specific tags
LogGuardianFunction:
  Type: AWS::Serverless::Function
  Properties:
    Tags:
      Runtime: provided.al2023
      Language: Go
      # Global tags automatically applied
```

## Deployment Script Integration

### Default Tagging
```bash
# Default tags applied by deployment script
./scripts/logguardian-deploy.sh deploy-single -e sandbox -r ca-central-1
# Results in:
# Product=LogGuardian, Owner=ZSoftly, Environment=sandbox, ManagedBy=SAM
```

### Custom Tagging
```bash
# Custom tags for enterprise deployment
./scripts/logguardian-deploy.sh deploy-prod \
  -e prod \
  -r ca-central-1 \
  --product-name "Enterprise-LogGuardian" \
  --owner "Platform-Engineering" \
  --managed-by "SAM"
```

### Customer-Specific Tagging
```bash
# Customer with their own tagging strategy
./scripts/logguardian-deploy.sh deploy-customer \
  -e prod \
  -r ca-central-1 \
  --customer-tag-prefix "ACME-Compliance-LogGuardian" \
  --owner "ACME-Security-Team" \
  --existing-kms-key alias/acme-compliance-logs
```

## Tag-Based Cost Tracking

### Cost Allocation Tags
Use these tags for cost allocation and chargeback:

```bash
# Finance team cost allocation
aws ce get-cost-and-usage \
  --time-period Start=2024-01-01,End=2024-02-01 \
  --granularity MONTHLY \
  --metrics BlendedCost \
  --group-by Type=DIMENSION,Key=TAG:Product \
  --group-by Type=DIMENSION,Key=TAG:Environment \
  --group-by Type=DIMENSION,Key=TAG:Owner
```

### Resource Filtering by Tags
```bash
# Find all LogGuardian resources in production
aws resourcegroupstaggingapi get-resources \
  --tag-filters Key=Product,Values=LogGuardian Key=Environment,Values=prod

# Find resources by owner
aws resourcegroupstaggingapi get-resources \
  --tag-filters Key=Owner,Values=Platform-Team
```

## Environment-Specific Tagging Examples

### Development Environment
```bash
./scripts/logguardian-deploy.sh deploy-dev \
  -r ca-central-1 \
  --owner "DevOps-Team" \
  --product-name "LogGuardian-Dev"

# Results in resources tagged with:
# Product=LogGuardian-Dev
# Owner=DevOps-Team  
# Environment=dev
# ManagedBy=SAM
```

### Staging Environment
```bash
./scripts/logguardian-deploy.sh deploy-staging \
  -r ca-central-1,ca-west-1 \
  --owner "QA-Team" \
  --product-name "LogGuardian-Staging"

# Results in resources tagged with:
# Product=LogGuardian-Staging
# Owner=QA-Team
# Environment=staging
# ManagedBy=SAM
```

### Production Environment
```bash
./scripts/logguardian-deploy.sh deploy-prod \
  -r ca-central-1,ca-west-1,us-east-1 \
  --owner "Platform-Engineering" \
  --product-name "LogGuardian-Production"

# Results in resources tagged with:
# Product=LogGuardian-Production
# Owner=Platform-Engineering
# Environment=prod
# ManagedBy=SAM
```

## Customer Integration Tagging

### Enterprise Customer
```bash
# Large enterprise with specific tagging requirements
./scripts/logguardian-deploy.sh deploy-customer \
  -e prod \
  -r us-east-1 \
  --customer-tag-prefix "MegaCorp-Compliance-LogGuardian" \
  --owner "MegaCorp-InfoSec-Team" \
  --existing-kms-key alias/megacorp-sox-compliance \
  --managed-by "SAM"

# Additional tags can be applied after deployment:
aws resourcegroupstaggingapi tag-resources \
  --resource-arn-list $(aws cloudformation describe-stack-resources \
    --stack-name logguardian-prod-us-east-1 \
    --query 'StackResources[].PhysicalResourceId' --output text) \
  --tags CostCenter=12345,Project=SOX-Compliance,BusinessUnit=InfoSec
```

### Multi-Account Customer
```bash
# Customer deploying across multiple accounts
./scripts/logguardian-deploy.sh deploy-customer \
  -e prod \
  -r ca-central-1 \
  --customer-tag-prefix "GlobalBank-LogGuardian" \
  --owner "GlobalBank-CloudOps" \
  --existing-kms-key alias/globalbank-audit-logs

# Additional account-specific tags
export ACCOUNT_TYPE="Production"
export BUSINESS_UNIT="Risk-Management"
```

## Tag Validation and Compliance

### Validation Script
```bash
#!/bin/bash
# validate-tags.sh - Ensure all resources have required tags

REQUIRED_TAGS=("Product" "Owner" "Environment" "ManagedBy")
STACK_NAME=$1
REGION=$2

echo "Validating tags for stack: $STACK_NAME in region: $REGION"

# Get all resources in stack
resources=$(aws cloudformation list-stack-resources \
  --stack-name $STACK_NAME \
  --region $REGION \
  --query 'StackResourceSummaries[].PhysicalResourceId' \
  --output text)

for resource in $resources; do
  echo "Checking resource: $resource"
  
  # Get resource tags
  tags=$(aws resourcegroupstaggingapi get-resources \
    --resource-arn-list $resource \
    --query 'ResourceTagMappingList[0].Tags' \
    --output json)
  
  # Check required tags
  for tag in "${REQUIRED_TAGS[@]}"; do
    if ! echo $tags | jq -e ".[] | select(.Key == \"$tag\")" > /dev/null; then
      echo "❌ Missing required tag: $tag on resource $resource"
    else
      echo "✅ Found required tag: $tag"
    fi
  done
  echo ""
done
```

### Policy Enforcement
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Deny",
      "Action": [
        "ec2:RunInstances",
        "lambda:CreateFunction",
        "s3:CreateBucket",
        "kms:CreateKey"
      ],
      "Resource": "*",
      "Condition": {
        "Null": {
          "aws:RequestedRegion": "false",
          "aws:RequestTag/Product": "true",
          "aws:RequestTag/Owner": "true",
          "aws:RequestTag/Environment": "true"
        }
      }
    }
  ]
}
```

## Tag Monitoring and Reporting

### Daily Tag Compliance Report
```bash
#!/bin/bash
# daily-tag-report.sh

echo "LogGuardian Tag Compliance Report - $(date)"
echo "================================================"

environments=("dev" "staging" "prod" "sandbox")
regions=("ca-central-1" "ca-west-1" "us-east-1")

for env in "${environments[@]}"; do
  for region in "${regions[@]}"; do
    echo ""
    echo "Environment: $env, Region: $region"
    echo "--------------------------------"
    
    # Check if stack exists
    if aws cloudformation describe-stacks \
       --stack-name logguardian-$env-$region \
       --region $region >/dev/null 2>&1; then
      
      # Get stack tags
      stack_tags=$(aws cloudformation describe-stacks \
        --stack-name logguardian-$env-$region \
        --region $region \
        --query 'Stacks[0].Tags' \
        --output table)
      
      echo "Stack Tags:"
      echo "$stack_tags"
    else
      echo "Stack not deployed"
    fi
  done
done
```

## Best Practices

1. **Consistency**: Use the deployment script parameters for consistent tagging
2. **Automation**: Let the script handle standard tags, add custom ones as needed
3. **Validation**: Run tag validation after deployments
4. **Cost Tracking**: Use tags for cost allocation and chargeback
5. **Compliance**: Enforce tagging policies via IAM and Service Control Policies
6. **Documentation**: Keep tag meanings and usage documented
7. **Regular Audits**: Run regular reports to ensure tag compliance

This tagging strategy ensures all LogGuardian resources are properly tagged for management, cost tracking, and compliance across all deployment scenarios.
