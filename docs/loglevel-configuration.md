# LogLevel Configuration

## Parameter
- **Name**: `LogLevel`
- **Type**: String
- **Default**: `INFO`
- **Allowed Values**: `ERROR`, `WARN`, `INFO`, `DEBUG`

## Usage by Environment

| Environment | Recommended | Purpose |
|------------|-------------|---------|
| Production | `ERROR` | Minimize CloudWatch costs, only critical issues |
| Staging | `INFO` | Standard logging for validation |
| Development | `DEBUG` | Full visibility for troubleshooting |

## Cost Impact

| Level | Relative Log Volume | CloudWatch Cost |
|-------|-------------------|-----------------|
| ERROR | 1x (baseline) | Lowest |
| WARN | ~5x | Low |
| INFO | ~20x | Moderate |
| DEBUG | ~100x | Highest |

## Runtime Override

For temporary debugging without redeployment:

```bash
aws lambda update-function-configuration \
  --function-name logguardian-compliance-${ENV} \
  --environment Variables="{LOG_LEVEL=DEBUG}"
```

## Log Examples

### ERROR
```
ERROR: KMS key not found: alias/missing-key
```

### INFO  
```
INFO: Processing compliance event
INFO: Applied KMS encryption to log-group-name
```

### DEBUG
```
DEBUG: Validating KMS key accessibility
DEBUG: Key state: Enabled, Region: ca-central-1
```