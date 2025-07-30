# LogGuardian Go Lambda Function

A high-performance Go Lambda function that automatically remediates CloudWatch Log Groups compliance issues by applying KMS encryption and retention policies.

## Features

- **AWS Config Integration**: Processes Config rule evaluation events
- **KMS Encryption**: Automatically applies customer-managed KMS keys to unencrypted log groups
- **Retention Policies**: Sets configurable retention periods on log groups
- **Multi-Region Support**: Handles log groups across multiple AWS regions
- **Memory Optimized**: Efficient memory usage for large-scale processing
- **Structured Logging**: JSON structured logging with Go's slog package
- **Comprehensive Testing**: Unit tests with mocked AWS services
- **Error Handling**: Graceful error handling with detailed logging

## Architecture

```
AWS Config Rule → EventBridge → Lambda Function → CloudWatch Logs/KMS
                                      ↓
                               Structured Logging
```

## Quick Start

### Prerequisites

- Go 1.24 or later
- AWS CLI configured
- Required AWS permissions for CloudWatch Logs, KMS, and Config

### Build and Deploy

```bash
# Build the Lambda function
make build

# Create deployment package
make package

# Run tests
make test

# Deploy using AWS CLI
aws lambda create-function \
  --function-name logguardian-compliance \
  --runtime provided.al2023 \
  --role arn:aws:iam::YOUR-ACCOUNT:role/lambda-execution-role \
  --handler main \
  --zip-file fileb://dist/logguardian-compliance.zip
```

### Configuration

Set environment variables for the Lambda function:

```bash
# Required
KMS_KEY_ALIAS=alias/cloudwatch-logs-compliance
DEFAULT_RETENTION_DAYS=365

# Optional
SUPPORTED_REGIONS=ca-central-1,ca-west-1,us-east-2
DRY_RUN=false
BATCH_LIMIT=100

# Region-specific overrides
KMS_KEY_ALIAS_us_west_2=alias/cloudwatch-logs-west
DEFAULT_RETENTION_DAYS_us_west_2=180
```

### Environment Variable Details

- **KMS_KEY_ALIAS**: KMS key alias for encrypting log groups
- **DEFAULT_RETENTION_DAYS**: Default retention period for log groups (in days)
- **DRY_RUN**: When true, logs actions without making changes
- **BATCH_LIMIT**: Maximum number of resources to retrieve per Config API call (default: 100)
- **SUPPORTED_REGIONS**: Comma-separated list of regions to process

## Usage

The Lambda function automatically processes AWS Config compliance events:

1. **Event Processing**: Receives Config rule evaluation events
2. **Compliance Analysis**: Determines what remediation is needed
3. **KMS Encryption**: Applies encryption using specified KMS key
4. **Retention Policy**: Sets retention period in days
5. **Logging**: Records all actions with structured logging

## Development

### Project Structure

```
├── cmd/lambda/          # Lambda entry point
├── internal/
│   ├── handler/         # Event handlers
│   ├── service/         # Business logic
│   └── types/           # Data structures
├── testdata/           # Test fixtures
├── Makefile           # Build automation
└── go.mod            # Go module definition
```

### Available Commands

```bash
make help               # Show all available commands
make build             # Build Lambda binary
make test              # Run tests
make test-coverage     # Generate coverage report
make lint              # Run linter
make security          # Security scan
make package           # Create deployment ZIP
make clean             # Clean build artifacts
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make bench

# Memory profiling
make memory-profile
```

### Code Quality

The project includes comprehensive quality checks:

- **Linting**: golangci-lint with security-focused rules
- **Security**: gosec security scanner
- **Vulnerability**: govulncheck for dependency vulnerabilities
- **Testing**: Unit tests with mocked AWS services
- **Coverage**: Code coverage reporting

## Memory Optimization

The function includes several memory optimization features:

- **Client Pooling**: Reuses AWS SDK clients across invocations
- **String Pooling**: Efficient string buffer management
- **Garbage Collection**: Proactive memory cleanup
- **Memory Monitoring**: Runtime memory usage tracking

## Configuration Examples

### Basic Configuration

```json
{
  "KMS_KEY_ALIAS": "alias/cloudwatch-logs-compliance",
  "DEFAULT_RETENTION_DAYS": "365",
  "SUPPORTED_REGIONS": "ca-central-1,ca-west-1"
}
```

### Multi-Region Configuration

```json
{
  "KMS_KEY_ALIAS": "alias/cloudwatch-logs-compliance",
  "DEFAULT_RETENTION_DAYS": "365",
  "SUPPORTED_REGIONS": "ca-central-1,ca-west-1,us-east-2",
  "KMS_KEY_ALIAS_eu_west_1": "alias/cloudwatch-logs-eu",
  "DEFAULT_RETENTION_DAYS_eu_west_1": "180"
}
```

## Error Handling

The function provides comprehensive error handling:

- **AWS API Errors**: Graceful handling of throttling and permission errors
- **Validation Errors**: Input validation with detailed error messages
- **Resource Errors**: Handling of missing or deleted resources
- **Timeout Handling**: Context-aware timeouts for all operations

## Monitoring

### Structured Logging

All operations are logged with structured JSON:

```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "Remediation completed",
  "log_group": "/aws/lambda/my-function",
  "encryption_applied": true,
  "retention_applied": true,
  "success": true
}
```

### Metrics

Monitor key metrics:

- Function duration and memory usage
- Success/failure rates
- Remediation actions performed
- AWS API call patterns

## Security

### IAM Permissions

Required IAM permissions for the Lambda execution role:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:AssociateKmsKey",
        "logs:PutRetentionPolicy",
        "logs:DescribeLogGroups"
      ],
      "Resource": "arn:aws:logs:*:*:log-group:*"
    },
    {
      "Effect": "Allow", 
      "Action": [
        "kms:DescribeKey"
      ],
      "Resource": "arn:aws:kms:*:*:key/*"
    }
  ]
}
```

### KMS Key Policy

Ensure the KMS key policy allows the Lambda role:

```json
{
  "Effect": "Allow",
  "Principal": {
    "AWS": "arn:aws:iam::ACCOUNT:role/lambda-execution-role"
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
    "StringEquals": {
      "kms:ViaService": "logs.REGION.amazonaws.com"
    }
  }
}
```

## Performance

### Benchmarks

Typical performance characteristics:

- **Cold Start**: < 2 seconds
- **Warm Execution**: < 100ms per log group
- **Memory Usage**: 64-128 MB depending on batch size
- **Throughput**: 100+ log groups per invocation

### Optimization Tips

1. **Pre-warm** functions with EventBridge scheduled events
2. **Batch processing** multiple compliance events
3. **Region optimization** by grouping by region
4. **Memory allocation** tuning based on workload

## Contributing

1. Fork the repository
2. Create a feature branch
3. Run tests: `make test`
4. Run quality checks: `make check`
5. Submit a pull request