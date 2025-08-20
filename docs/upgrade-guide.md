# LogGuardian Upgrade Guide

## v1.0.4 Upgrade (Current → Latest)

### Changes
- Dashboard compliance metrics now working
- Separate Config rule controls (`CreateEncryptionConfigRule`, `CreateRetentionConfigRule`)
- Custom retention Config rule support

### Existing Deployments

#### Console Update
1. CloudFormation → your stack → **Update**
2. **Replace current template** → **Amazon S3 URL**
3. Get URL: SAR Console → LogGuardian → **Copy S3 URL** (v1.0.4)
4. **Next** → **Next** → **Update**

#### CLI Update
```bash
# Get template URL
TEMPLATE_URL=$(aws serverlessrepo create-cloud-formation-template \
  --application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
  --semantic-version 1.0.4 \
  --region ca-central-1 \
  --query 'TemplateUrl' --output text)

# Update stack
aws cloudformation update-stack \
  --stack-name your-stack-name \
  --template-url $TEMPLATE_URL \
  --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND
```

### New Deployments
Use v1.0.4 from SAR - metrics included automatically.

### Parameter Migration
Old `CreateConfigRules=true` becomes:
- `CreateEncryptionConfigRule=true`
- `CreateRetentionConfigRule=true`

For custom retention rule:
- `CreateRetentionConfigRule=false`
- `ExistingRetentionConfigRule=arn:aws:serverlessrepo:ca-central-1:410129828371:applications/CloudWatch-LogGroup-Retention-Monitor`

### Verification
Dashboard compliance metrics appear within 5 minutes after first Lambda execution.