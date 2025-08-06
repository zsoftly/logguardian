# Local Lambda Testing with SAM

This document explains how to test the LogGuardian Lambda function locally using AWS SAM.

## Prerequisites

- AWS SAM CLI installed (`sam --version`)
- Docker running (SAM uses containers for local execution)
- Go 1.24+ for building the Lambda function

## Quick Start

```bash
# Build and test the Lambda function
make sam-build
make sam-validate
make sam-test-all-events
```

## Test Events

LogGuardian supports comprehensive testing with multiple event scenarios:

### 1. Config Rule Evaluation Events
Batch processing requests to evaluate non-compliant resources:

**Encryption Rule Evaluation** (`testdata/config-rule-evaluation-event.json`):
```json
{
  "type": "config-rule-evaluation",
  "configRuleName": "logguardian-encryption-dev",
  "region": "ca-central-1",
  "batchSize": 10
}
```
**Test with**: `make sam-local-invoke`

**Retention Rule Evaluation** (`testdata/retention-rule-evaluation-event.json`):
```json
{
  "type": "config-rule-evaluation", 
  "configRuleName": "logguardian-retention-dev",
  "batchSize": 15
}
```
**Test with**: `make sam-local-invoke-retention`

**Large Batch Evaluation** (`testdata/large-batch-evaluation-event.json`):
```json
{
  "type": "config-rule-evaluation",
  "configRuleName": "logguardian-encryption-prod",
  "batchSize": 50
}
```
**Test with**: `make sam-local-invoke-large-batch`

### 2. Individual Config Events
Individual Config resource change notifications:

**Missing Encryption** (`testdata/unified-config-event.json`):
- Log group: `/aws/lambda/test-function`
- Missing: KMS encryption
- Has: No retention policy
**Test with**: `make sam-local-invoke-config`

**Missing Retention** (`testdata/retention-config-event.json`):
- Log group: `/aws/apigateway/my-api`  
- Missing: Retention policy (has 30 days, needs 365)
- Has: No KMS encryption
**Test with**: `make sam-local-invoke-retention-config`

**Compliant Log Group** (`testdata/compliant-log-group-event.json`):
- Log group: `/aws/ecs/my-service`
- Has: KMS encryption + retention policy  
- Should: No remediation needed
**Test with**: `make sam-local-invoke-compliant`

**Both Missing** (`testdata/both-missing-config-event.json`):
- Log group: `/aws/rds/instance/my-db/error`
- Missing: Both KMS encryption + retention policy
- Should: Apply both remediations
**Test with**: `make sam-local-invoke-both-missing`

### 3. Error Handling Tests

**Invalid Event Type** (`testdata/invalid-event-type.json`):
```json
{
  "type": "invalid-type-test",
  "invalidData": {"someField": "someValue"}
}
```
**Test with**: `make sam-local-invoke-invalid`

## Local Testing Commands

### Comprehensive Testing
```bash
# Test all scenarios (recommended for development)
make sam-test-all-events

# Quick test of common scenarios
make sam-test-quick

# Test specific rule types
make sam-test-encryption    # Test encryption scenarios
make sam-test-retention     # Test retention scenarios
make sam-test-errors        # Test error handling
```

### Individual Tests
```bash
# Config rule evaluation events
make sam-local-invoke           # Encryption rule evaluation
make sam-local-invoke-retention # Retention rule evaluation  
make sam-local-invoke-large-batch # Large batch processing

# Individual config events
make sam-local-invoke-config    # Missing encryption
make sam-local-invoke-retention-config # Missing retention
make sam-local-invoke-compliant # Compliant log group
make sam-local-invoke-both-missing # Both issues
make sam-local-invoke-invalid   # Invalid event type
```

### Build and Validate
```bash
# Build Go binary and prepare SAM structure
make sam-build

# Validate SAM template
make sam-validate
```

### Local API Server
```bash
# Start local API Gateway
make sam-local-start

# In another terminal, send POST request:
curl -X POST http://127.0.0.1:3000/2015-03-31/functions/function/invocations \
  -d @testdata/config-rule-evaluation-event.json
```

## Environment Variables

Local testing uses `env.json` with these variables:

```json
{
  "LogGuardianFunction": {
    "KMS_KEY_ALIAS": "alias/logguardian-logs",
    "DEFAULT_RETENTION_DAYS": "365", 
    "ENVIRONMENT": "dev",
    "LOG_LEVEL": "DEBUG",
    "DRY_RUN": "true",
    "BATCH_LIMIT": "10"
  }
}
```

**Key Settings**:
- `DRY_RUN=true`: Prevents actual AWS API calls
- `LOG_LEVEL=DEBUG`: Detailed logging output

## Expected Behavior

### Config Rule Evaluation Event
```bash
❯ make sam-local-invoke
{"level":"INFO","msg":"Received Lambda request","type":"config-rule-evaluation"}
{"level":"INFO","msg":"Processing Config rule evaluation request","config_rule":"logguardian-encryption-dev"}
{"level":"ERROR","msg":"Failed to get compliance details","error":"...invalid security token..."}
```

**Expected**: Fails with invalid credentials (normal in local testing)

### Individual Config Event
```bash
❯ make sam-local-invoke-config
{"level":"INFO","msg":"Received Lambda request","type":"config-event"}
{"level":"INFO","msg":"Processing compliance event","log_group":"/aws/lambda/test-function"}
{"level":"INFO","msg":"DRY RUN: Would apply KMS encryption"}
{"level":"INFO","msg":"Remediation completed","success":true}
```

**Expected**: Succeeds with dry-run remediation

## Troubleshooting

### Common Issues

1. **SAM CLI not found**
   ```bash
   # Install SAM CLI
   pip install aws-sam-cli
   # or use official installer
   ```

2. **Docker not running**
   ```bash
   # Start Docker service
   sudo systemctl start docker
   ```

3. **Template validation errors**
   ```bash
   # Check template syntax
   make sam-validate
   ```

4. **Binary architecture mismatch**
   ```bash
   # Ensure Linux x86_64 compilation
   GOOS=linux GOARCH=amd64 make build
   ```

### Debug Mode

For detailed debugging:

```bash
# Enable debug logging
sam local invoke LogGuardianFunction \
  --event testdata/config-rule-evaluation-event.json \
  --env-vars env.json \
  --debug
```

## Next Steps

After local testing succeeds:

1. **Deploy to AWS**: `make sam-deploy-dev`
2. **Package for Marketplace**: `make sam-package-marketplace`
3. **Publish to SAR**: `make sam-publish`

## File Structure

```
testdata/
├── config-rule-evaluation-event.json  # Batch processing event
└── unified-config-event.json          # Individual Config event

env.json                               # Local environment variables
template.yaml                         # SAM template
build/bootstrap                       # Compiled Go binary
```
