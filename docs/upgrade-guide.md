# LogGuardian Upgrade Guide

## Upgrade Guide (Latest Version)

### Changes
- Dashboard compliance metrics now working
- Separate Config rule controls (`CreateEncryptionConfigRule`, `CreateRetentionConfigRule`)
- Custom retention Config rule support

### Existing Deployments

#### Console Update
1. CloudFormation → your stack → **Update**
2. **Replace current template** → **Amazon S3 URL**
3. Get URL: SAR Console → LogGuardian → **Copy S3 URL** (latest version)
4. **Next** → **Next** → **Update**

#### CLI Update (Recommended - With Change Set Review)
```bash
# Get template URL (latest version)
TEMPLATE_URL=$(aws serverlessrepo create-cloud-formation-template \
  --application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
  --region ca-central-1 \
  --query 'TemplateUrl' --output text)

# Create a change set for review
aws cloudformation create-change-set \
  --stack-name serverlessrepo-LogGuardian \
  --change-set-name logguardian-upgrade-$(date +%Y%m%d-%H%M%S) \
  --template-url $TEMPLATE_URL \
  --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND

# Review the change set via CLI
aws cloudformation describe-change-set \
  --change-set-name logguardian-upgrade-$(date +%Y%m%d-%H%M%S) \
  --stack-name serverlessrepo-LogGuardian \
  --query 'Changes[*].[Type, ResourceChange.Action, ResourceChange.LogicalResourceId, ResourceChange.ResourceType]' \
  --output table

# Or review in the AWS Console:
# CloudFormation → Stack → Change Sets → Review changes

# If changes look good, execute the change set
aws cloudformation execute-change-set \
  --change-set-name logguardian-upgrade-$(date +%Y%m%d-%H%M%S) \
  --stack-name serverlessrepo-LogGuardian

# Monitor the update progress
aws cloudformation wait stack-update-complete \
  --stack-name serverlessrepo-LogGuardian
```

#### Direct Update (Without Review)
```bash
# Get template URL and update directly
TEMPLATE_URL=$(aws serverlessrepo create-cloud-formation-template \
  --application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
  --region ca-central-1 \
  --query 'TemplateUrl' --output text)

aws cloudformation update-stack \
  --stack-name serverlessrepo-LogGuardian \
  --template-url $TEMPLATE_URL \
  --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND
```

### New Deployments
Use latest version from SAR - metrics included automatically.

### Parameter Migration
Old `CreateConfigRules=true` becomes:
- `CreateEncryptionConfigRule=true`
- `CreateRetentionConfigRule=true`

For custom retention rule:
- `CreateRetentionConfigRule=false`
- `ExistingRetentionConfigRule=arn:aws:serverlessrepo:ca-central-1:410129828371:applications/CloudWatch-LogGroup-Retention-Monitor`

### Verification
Dashboard compliance metrics appear within 5 minutes after first Lambda execution.