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

**IMPORTANT**: Before upgrading, capture your current parameter values to preserve your configuration.

##### Step 1: Capture Current Parameter Values

First, save your existing stack parameters to ensure they are preserved during the upgrade:

```bash
STACK_NAME="serverlessrepo-LogGuardian"  # Update if your stack name is different

# Save current parameters to a file
aws cloudformation describe-stacks \
  --stack-name $STACK_NAME \
  --query 'Stacks[0].Parameters' > current-params.json

# View current parameters to verify
cat current-params.json | jq -r '.[] | "\(.ParameterKey)=\(.ParameterValue)"'
```

##### Step 2: Get Latest Template URL

```bash
TEMPLATE_URL=$(aws serverlessrepo create-cloud-formation-template \
  --application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
  --region ca-central-1 \
  --query 'TemplateUrl' --output text)
```

##### Step 3: Create Change Set for Review

Create a change set that preserves your existing parameters:

```bash
CHANGE_SET_NAME="logguardian-upgrade-$(date +%Y%m%d-%H%M%S)"

aws cloudformation create-change-set \
  --stack-name $STACK_NAME \
  --change-set-name $CHANGE_SET_NAME \
  --template-url $TEMPLATE_URL \
  --parameters file://current-params.json \
  --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND
```

##### Step 4: Review Proposed Changes

Review the changes that will be applied:

```bash
# Review resource changes
aws cloudformation describe-change-set \
  --change-set-name $CHANGE_SET_NAME \
  --stack-name $STACK_NAME \
  --query 'Changes[*].[Type, ResourceChange.Action, ResourceChange.LogicalResourceId, ResourceChange.ResourceType]' \
  --output table

# Review parameter changes
aws cloudformation describe-change-set \
  --change-set-name $CHANGE_SET_NAME \
  --stack-name $STACK_NAME \
  --query 'Parameters[*].[ParameterKey, ParameterValue]' \
  --output table
```

Alternatively, review in AWS Console: **CloudFormation → Stack → Change Sets**

##### Step 5: Execute Change Set

After confirming the changes are acceptable:

```bash
aws cloudformation execute-change-set \
  --change-set-name $CHANGE_SET_NAME \
  --stack-name $STACK_NAME

# Wait for update to complete
aws cloudformation wait stack-update-complete \
  --stack-name $STACK_NAME
```

##### Important Considerations

1. **Parameter Preservation**: The commands above preserve your existing parameter values. Without this, parameters would revert to defaults.

2. **Common Parameters to Check**:
   - `Environment` (prod, staging, dev)
   - `DefaultRetentionDays` (your retention policy)
   - `CreateConfigService` (true/false)
   - `CreateKMSKey` (true/false)
   - `LogLevel` (ERROR, WARN, INFO, DEBUG)
   - `LambdaMemorySize` (128-3008)
   - Custom schedule expressions

3. **Parameter Override if Needed**:
   ```bash
   # To override specific parameters during upgrade
   aws cloudformation create-change-set \
     --stack-name $STACK_NAME \
     --change-set-name $CHANGE_SET_NAME \
     --template-url $TEMPLATE_URL \
     --parameters file://current-params.json \
     --parameter-overrides \
       LogLevel=ERROR \
       DefaultRetentionDays=90 \
     --capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND
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