# Basic LogGuardian Deployment

Complete deployment of LogGuardian with all new resources.

## What Gets Created

- ✅ New KMS key for encryption
- ✅ New AWS Config service (recorder, delivery channel, S3 bucket)
- ✅ New Config rules (encryption & retention)
- ✅ Lambda function
- ✅ EventBridge scheduled triggers
- ✅ CloudWatch dashboard
- ✅ IAM roles and policies

## Prerequisites

1. **AWS Credentials** configured (`aws configure`)
2. **Terraform** >= 1.5.0 installed
3. **Lambda Binary** built:
```bash
   cd ../../../  # Go to logguardian root
   make build    # Or: GOOS=linux GOARCH=amd64 go build -o build/bootstrap ./cmd/lambda
```

## Usage
```bash
# 1. Copy example config
cp terraform.tfvars.example terraform.tfvars

# 2. Edit terraform.tfvars with your values
nano terraform.tfvars

# 3. Initialize Terraform
terraform init

# 4. Review the plan
terraform plan

# 5. Deploy
terraform apply

# 6. Test Lambda function
# Copy the command from the output:
terraform output -raw manual_invocation_command | bash
```

## Configuration

### Minimal Config
```hcl
environment = "dev"
aws_region  = "ca-central-1"
```

### Custom Retention
```hcl
environment            = "prod"
default_retention_days = 90
lambda_log_retention_days = 30
```

### Aggressive Scheduling
```hcl
encryption_schedule = "rate(12 hours)"
retention_schedule  = "rate(12 hours)"
```

### Dry-Run Mode
```hcl
dry_run = true  # Preview changes without applying
```

## Testing
```bash
# Manual Lambda invocation
aws lambda invoke \
  --function-name logguardian-compliance-dev \
  --payload '{"type":"config-rule-evaluation","configRuleName":"logguardian-encryption-dev","region":"ca-central-1"}' \
  response.json

# Check response
cat response.json

# View logs
aws logs tail /aws/lambda/logguardian-compliance-dev --follow

# View dashboard
terraform output dashboard_url
```

## Cost Estimate

**Monthly Cost (ca-central-1):**
- KMS Key: ~$1.00
- Lambda (weekly schedule): ~$0.01
- Config Service: ~$2.00
- Config Rules (2): ~$4.00
- S3 Storage (10GB): ~$0.23
- EventBridge: FREE (within free tier)
- CloudWatch Dashboard: FREE (first 3 dashboards)

**Total: ~$7-8/month**

## Cleanup
```bash
# Destroy all resources
terraform destroy

# Confirm when prompted
```

## Troubleshooting

**Issue: "No such file: bootstrap"**
```bash
# Build the Lambda binary
cd ../../..
make build
ls -lh build/bootstrap
```

**Issue: "AWS Config already enabled"**
```bash
# Use existing-config example instead
cd ../existing-config
```

**Issue: Permission errors**
```bash
# Check AWS credentials
aws sts get-caller-identity

# Check IAM permissions
aws iam get-user
```
