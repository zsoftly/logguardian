# KMS Encryption Validation

LogGuardian applies KMS encryption to CloudWatch log groups with comprehensive safety checks to prevent breaking application access while ensuring compliance.

## Why KMS Validation Matters

CloudWatch log group encryption with KMS keys can fail catastrophically if:
- Keys don't exist or are inaccessible in the target region
- Key policies don't allow CloudWatch Logs service access  
- Keys are disabled or pending deletion
- Cross-region permissions are misconfigured

LogGuardian prevents these failures through comprehensive pre-validation.

## Validation Process

**Implementation**: See `validateKMSKeyAccessibility()` and related functions in [`internal/service/compliance.go`](../internal/service/compliance.go)

### Key Validation Steps
1. **Existence Check** - Verifies key exists and is accessible
2. **State Validation** - Ensures key is enabled and usable
3. **Policy Verification** - Confirms CloudWatch Logs service permissions
4. **Cross-Region Detection** - Identifies and warns about cross-region usage
5. **Retry Logic** - Handles transient failures with exponential backoff

### Cross-Region Support

LogGuardian automatically detects cross-region KMS keys and provides appropriate warnings while still allowing the operation.

**Environment Configuration**:
```bash
# Global fallback key
KMS_KEY_ALIAS="alias/cloudwatch-logs-global"

# Region-specific keys (recommended)  
KMS_KEY_ALIAS_us_east_1="alias/cloudwatch-logs-east"
KMS_KEY_ALIAS_us_west_2="alias/cloudwatch-logs-west"
```

## Audit Trail

LogGuardian provides comprehensive structured logging for all KMS operations for compliance and troubleshooting.

**Implementation**: All audit logging is handled in the service layer with structured JSON output.

### Key Audit Events
- **Encryption Success** - Successful KMS key association
- **Validation Failures** - Key not found, access denied, policy issues
- **Cross-Region Usage** - Warnings when using keys across regions
- **Retry Operations** - Exponential backoff retry attempts

### Example Log Entries

**Successful Encryption**:
```json
{
  "level": "INFO",
  "msg": "Successfully applied KMS encryption", 
  "log_group": "/aws/lambda/my-function",
  "kms_key_id": "12345678-1234-1234-1234-123456789012",
  "operation": "associate_kms_key",
  "timestamp": "2025-07-30T12:00:00Z"
}
```

**Cross-Region Warning**:
```json
{
  "level": "WARN",
  "msg": "KMS key is in different region than current",
  "kms_key_alias": "alias/cloudwatch-logs-compliance", 
  "key_region": "ca-west-1",
  "current_region": "ca-central-1",
  "audit_action": "cross_region_key_usage"
}
```

## Prerequisites

### Required KMS Key Policy

Your KMS key policy must include CloudWatch Logs service permissions:

```json
{
  "Effect": "Allow",
  "Principal": {"Service": "logs.amazonaws.com"},
  "Action": [
    "kms:Encrypt", "kms:Decrypt", "kms:ReEncrypt*", 
    "kms:GenerateDataKey*", "kms:DescribeKey"
  ],
  "Resource": "*"
}
```

### Lambda IAM Permissions

The Lambda execution role needs:
- `kms:Describe*`, `kms:GetKeyPolicy` for validation
- `logs:AssociateKmsKey` for applying encryption

**Reference**: See [Go Lambda Function - IAM Permissions](go-lambda-function.md#iam-permissions) for complete IAM configuration.

## Best Practices

1. **Use Region-Specific Keys** - Avoid cross-region charges and latency
2. **Pre-validate Keys** - Test KMS key access before deploying LogGuardian  
3. **Monitor Failures** - Set up CloudWatch alarms for KMS validation failures
4. **Audit Regularly** - Review KMS usage logs for compliance

**Troubleshooting**: Common issues and solutions are logged with actionable error messages. Check CloudWatch Logs for detailed failure reasons.
