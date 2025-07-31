# Parameter Files

This directory contains CloudFormation parameter files for different environments.

## ⚠️ IMPORTANT: ALL VALUES ARE EXAMPLES ONLY

**🔴 DO NOT USE THESE PARAMETER VALUES IN PRODUCTION**

All parameter values in these files are **examples only** and must be replaced with your actual values before deployment:

- **Account IDs**: Replace `123456789012` with your actual AWS account ID
- **Bucket Names**: Replace example bucket names with your actual S3 buckets  
- **Region Names**: Verify regions match your deployment targets
- **Environment Names**: Adjust to match your naming conventions
- **Resource Names**: Update to follow your organization's naming standards

## Important Security Notes

⚠️ **Account ID Placeholders**: The bucket names in these parameter files contain example AWS account IDs (123456789012). Replace these with your actual AWS account ID before deployment.

⚠️ **S3 Bucket Names**: Update the `DeploymentBucket` and `TemplatesBucket` parameter values to use your actual S3 bucket names.

## Parameter Files

### `sandbox-parameters.json`
- **⚠️ ALL VALUES ARE EXAMPLES** - Replace with your actual values
- Environment: sandbox
- Retention: 7 days (short for testing)
- Memory: 256 MB (minimal for cost)
- All features enabled for testing

### `prod-parameters.json`
- **⚠️ ALL VALUES ARE EXAMPLES** - Replace with your actual values
- Environment: production
- Retention: 365 days (long-term)
- Memory: 256 MB (optimized)
- Scheduled every 6 hours

## Usage

```bash
# Deploy with parameter file
aws cloudformation deploy \
  --template-file ../01-logguardian-main.yaml \
  --stack-name logguardian-sandbox \
  --parameter-overrides file://sandbox-parameters.json \
  --capabilities CAPABILITY_NAMED_IAM
```

## Customization

Before deploying:
1. Replace `123456789012` with your AWS account ID
2. Update bucket names to match your S3 buckets
3. Adjust parameter values for your requirements
