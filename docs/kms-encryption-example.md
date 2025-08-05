# KMS Encryption Example

This document demonstrates the comprehensive KMS encryption functionality implemented in LogGuardian.

## Features

The Lambda now safely applies KMS encryption using pre-configured keys with the following capabilities:

✅ **Validates KMS key existence and accessibility**
✅ **Uses customer-managed keys via alias resolution**  
✅ **Verifies key policies allow CloudWatch Logs service**
✅ **Applies encryption with proper error handling**
✅ **Logs all encryption operations for audit trail**
✅ **Supports cross-region key management**

## Example Lambda Request

### Config Rule Evaluation Request

```json
{
  "type": "config-rule-evaluation",
  "configRuleName": "cloudwatch-logs-compliance",
  "region": "ca-central-1",
  "batchSize": 5
}
```

### Individual Config Event Request

```json
{
  "type": "config-event",
  "configEvent": {
    "configRuleName": "cloudwatch-logs-compliance",
    "configurationItem": {
      "resourceType": "AWS::Logs::LogGroup",
      "resourceId": "/aws/lambda/my-function",
      "resourceName": "/aws/lambda/my-function",
      "awsRegion": "ca-central-1",
      "accountId": "123456789012"
    }
  }
}
```

## Environment Variables

Configure the following environment variables for KMS encryption:

```bash
# KMS Configuration
KMS_KEY_ALIAS=alias/cloudwatch-logs-compliance  # Default KMS key alias
DEFAULT_RETENTION_DAYS=365                       # Default retention in days
DRY_RUN=false                                    # Set to true for testing

# AWS Configuration  
AWS_REGION=ca-central-1                             # AWS region
AWS_DEFAULT_REGION=ca-central-1                     # Fallback region
```

## KMS Key Policy Requirements

Your KMS key policy must allow the CloudWatch Logs service to use the key:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowCloudWatchLogsEncryption",
      "Effect": "Allow", 
      "Principal": {
        "Service": [
          "logs.amazonaws.com",
          "logs.ca-central-1.amazonaws.com"
        ]
      },
      "Action": [
        "kms:Encrypt",
        "kms:Decrypt", 
        "kms:ReEncrypt*",
        "kms:GenerateDataKey*",
        "kms:DescribeKey"
      ],
      "Resource": "*",
      "Condition": {
        "ArnEquals": {
          "kms:EncryptionContext:aws:logs:arn": "arn:aws:logs:ca-central-1:123456789012:log-group:*"
        }
      }
    }
  ]
}
```

## Cross-Region KMS Keys

The Lambda supports cross-region KMS key management:

1. **Key Discovery**: Automatically detects if a KMS key is in a different region
2. **Access Validation**: Verifies cross-region access permissions
3. **Audit Logging**: Logs cross-region key usage for compliance
4. **Error Handling**: Provides clear error messages for cross-region issues

## Audit Trail Example

The Lambda provides comprehensive audit logging for all KMS operations:

```json
{
  "level": "INFO",
  "msg": "Successfully applied KMS encryption",
  "log_group": "/aws/lambda/my-function",
  "kms_key_id": "12345678-1234-1234-1234-123456789012",
  "kms_key_arn": "arn:aws:kms:ca-central-1:123456789012:key/12345678-1234-1234-1234-123456789012",
  "operation": "associate_kms_key",
  "timestamp": "2025-07-30T12:00:00Z"
}
```

## Error Handling

The Lambda handles various KMS-related errors gracefully:

### Key Not Found
```json
{
  "level": "ERROR",
  "msg": "KMS key not found during validation",
  "kms_key_alias": "alias/nonexistent-key",
  "current_region": "ca-central-1",
  "audit_action": "key_validation_failed",
  "failure_reason": "key_not_found"
}
```

### Access Denied
```json
{
  "level": "ERROR", 
  "msg": "KMS key access denied during validation",
  "kms_key_alias": "alias/restricted-key",
  "current_region": "ca-central-1",
  "audit_action": "key_validation_failed",
  "failure_reason": "access_denied"
}
```

### Disabled Key
```json
{
  "level": "ERROR",
  "msg": "KMS key is not in usable state",
  "kms_key_alias": "alias/disabled-key", 
  "key_state": "Disabled",
  "audit_action": "key_validation_failed",
  "failure_reason": "unusable_key_state"
}
```

## Retry Logic

The Lambda includes exponential backoff retry logic for transient failures:

- **Maximum Retries**: 3 attempts
- **Backoff Strategy**: Exponential (1s, 2s, 4s)
- **Rate Limiting**: Automatic detection and handling
- **Non-Retryable Errors**: Key not found, access denied, invalid log group

## Multi-Region Support

For multi-region deployments, the Lambda can be configured with region-specific KMS keys:

```bash
# Region-specific KMS keys
KMS_KEY_ALIAS_CA_CENTRAL_1=alias/logs-ca-central-1
KMS_KEY_ALIAS_CA_WEST_1=alias/logs-ca-west-1
KMS_KEY_ALIAS_EU_CENTRAL_1=alias/logs-eu-central-1
```

The Lambda will automatically select the appropriate key based on the resource region.
