# LogGuardian Project Instructions for Claude

## Critical Requirements - DO NOT CHANGE

### Go Version Requirements
- **MINIMUM Go Version: 1.24** 
- **Current go.mod: Go 1.24**
- **Dockerfile Base Image: golang:1.24-alpine**
- **NEVER downgrade Go version below 1.24**
- **Always use the latest stable Go version available**

### Version Compatibility Matrix
| Component | Required Version | Reason |
|-----------|-----------------|---------|
| Go | 1.24 | Project standard - DO NOT DOWNGRADE |
| Alpine | 3.19+ | Security and compatibility |
| AWS SDK v2 | v1.38.0+ | Latest features and security |

## Container Implementation Status

### Completed Features
✅ Multi-stage Dockerfile with Alpine base image  
✅ Docker Compose configuration  
✅ CLI interface (`cmd/container/main.go`)  
✅ AWS authentication chain (`internal/container/auth.go`)  
✅ Config rule processor (`internal/container/processor.go`)  
✅ Service adapter with retry logic (`internal/container/service_adapter.go`)  
✅ Dry-run mode (`internal/container/dryrun.go`)  
✅ Unit tests for all components  

### Authentication Strategies (Priority Order)
1. Explicit credentials (AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY)
2. Profile (--profile flag or AWS_PROFILE env)
3. Assume Role (--assume-role flag)
4. ECS Task Role (auto-detected)
5. Environment variables
6. EC2 Instance Profile
7. Default credential chain

## Local Testing Commands

### Build Commands
```bash
# Build the binary locally
go build -o ./build/logguardian-container ./cmd/container/main.go

# Build Docker image (ensure Dockerfile uses golang:1.24-alpine)
docker build -t logguardian:latest .
```

### Test with AWS Profile (Local Binary)
```bash
# ALWAYS use ca-central-1 region for all testing

# Basic dry-run test
./build/logguardian-container \
  --config-rule cw-lg-retention-min \
  --region ca-central-1 \
  --profile logdev \
  --dry-run

# With verbose logging
./build/logguardian-container \
  --config-rule cw-lg-retention-min \
  --region ca-central-1 \
  --profile logdev \
  --dry-run \
  --verbose

# Text output format
./build/logguardian-container \
  --config-rule cw-lg-retention-min \
  --region ca-central-1 \
  --profile logdev \
  --dry-run \
  --output text
```

### Docker Testing Commands
```bash
# ALWAYS use ca-central-1 region for all testing

# Run with AWS credentials from host
docker run --rm \
  -v ~/.aws:/home/logguardian/.aws:ro \
  -e AWS_PROFILE=logdev \
  logguardian:latest \
  --config-rule cw-lg-retention-min \
  --region ca-central-1 \
  --dry-run

# With explicit environment variables
docker run --rm \
  -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
  -e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
  -e AWS_SESSION_TOKEN="$AWS_SESSION_TOKEN" \
  -e AWS_DEFAULT_REGION=ca-central-1 \
  logguardian:latest \
  --config-rule cw-lg-retention-min \
  --region ca-central-1 \
  --dry-run

# Using Docker Compose
docker-compose run --rm logguardian \
  --config-rule cw-lg-retention-min \
  --region ca-central-1 \
  --dry-run
```

## Linting and Testing

### Required Checks Before Committing
**IMPORTANT**: Always run these checks locally to avoid pipeline failures

```bash
# 1. Format code
go fmt ./...

# 2. Run golangci-lint (REQUIRED - catches security issues)
# Install if needed: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run --timeout=5m

# 3. Run tests
go test ./internal/container/... -v

# 4. Run tests with race detector
go test ./internal/container/... -race

# 5. Build to verify compilation
go build ./...

# 6. Run with dry-run to verify functionality (ALWAYS use ca-central-1)
./build/logguardian-container --config-rule cw-lg-retention-min --region ca-central-1 --profile logdev --dry-run
```

### Quick Pre-Commit Check Script
```bash
#!/bin/bash
# Save as pre-commit.sh and run before committing

echo "🔍 Running pre-commit checks..."

echo "📝 Formatting code..."
go fmt ./...

echo "🔒 Running security linter..."
golangci-lint run --timeout=5m || exit 1

echo "🧪 Running tests..."
go test ./... -v || exit 1

echo "🏁 Running race detector..."
go test ./... -race || exit 1

echo "🔨 Building..."
go build ./... || exit 1

echo "✅ All checks passed!"
```

## Common Issues and Solutions

### Issue: Docker build fails with Go version error
**Solution**: Ensure Dockerfile uses `golang:1.24-alpine`, NEVER use lower versions

### Issue: Authentication fails in container
**Solution**: Mount AWS credentials directory or pass environment variables:
```bash
-v ~/.aws:/home/logguardian/.aws:ro
```

### Issue: No EC2 IMDS role found error
**Solution**: Use explicit profile with `--profile` flag or set AWS_PROFILE environment variable

## Environment Variables

| Variable | Description | Example/Default |
|----------|-------------|-----------------|
| AWS_PROFILE | AWS profile to use | logdev |
| AWS_REGION | AWS region (ALWAYS use ca-central-1) | ca-central-1 |
| AWS_DEFAULT_REGION | Fallback region (ALWAYS use ca-central-1) | ca-central-1 |
| CONFIG_RULE_NAME | Config rule name | cw-lg-retention-min |
| BATCH_SIZE | Resources per batch | 10 |
| DRY_RUN | Enable dry-run mode | true |

## Architecture Notes

### Container vs Lambda Parity
The container implementation maintains 100% functional parity with the Lambda version:
- Same business logic in `internal/service` and `internal/handler`
- Same AWS Config rule evaluation
- Same CloudWatch Logs remediation capabilities
- Additional features: dry-run mode, multiple auth strategies, CLI interface

### Service Adapter Pattern
- Provides abstraction layer for AWS services
- Built-in retry logic with exponential backoff
- Rate limiting to prevent API throttling
- Circuit breaker pattern for fault tolerance
- Metrics collection for monitoring

## Future Enhancements (Not Yet Implemented)
- ECS task definitions and EventBridge integration
- CloudWatch Metrics and distributed tracing
- ECR push workflows and CI/CD pipeline
- Kubernetes manifests and Helm charts
- Enhanced health checks and graceful shutdown
- Correlation IDs and structured logging
- Lambda-to-container migration tools

## Important Reminders
1. **NEVER downgrade Go version below 1.24 - Project uses Go 1.24 as standard**
2. **ALWAYS use ca-central-1 region for all AWS operations**
3. **ALWAYS run golangci-lint locally before committing to catch security issues**
4. **Always test with --dry-run flag first**
5. **Run tests before committing code**
6. **Maintain functional parity between Lambda and container versions**
7. **Document any new authentication strategies or configuration options**
8. **Use crypto/rand for randomness, NEVER math/rand (security requirement)**

## Default AWS Configuration
**IMPORTANT**: Always use `ca-central-1` as the default AWS region for all operations:
- **Region**: `ca-central-1` (Canada Central)
- **Profile**: `logdev` (for development/testing)
- **Config Rule**: `cw-lg-retention-min`

## Contact and Support
- Repository: github.com/zsoftly/logguardian
- Primary Config Rule: cw-lg-retention-min
- Test Environment: logdev profile, ca-central-1 region

## Test File Management Guidelines
**IMPORTANT**: When creating test files or scripts during development:

### Temporary Files to Clean Up
1. **One-time test files**: Delete after use
   - `*_repair_test.go`, `*_init_test.go` (unless part of permanent suite)
   - `test_coverage.sh`, `run_tests.sh` (temporary scripts)
   
2. **Coverage artifacts**: Remove after review
   ```bash
   rm -rf coverage/ *.out *.html
   ```

3. **Test binaries and cache**: Clean after testing
   ```bash
   go clean -testcache
   rm -f *.test
   ```

### Permanent Test Files (Keep These)
- `*_test.go` files that test actual functionality
- Test fixtures in `testdata/` directories
- Benchmark tests for performance validation

### Best Practices
- **Before committing**: Clean up all temporary test artifacts
- **After debugging**: Remove one-off test files
- **Duplicate tests**: Update existing tests rather than creating new files
- **Use .gitignore**: Ensure coverage/, *.out, *.test are ignored

## Code Maintenance Guidelines
**CRITICAL**: When removing deprecated, redundant, or obsolete code:
- **NO COMMENTS**: Never leave comments explaining what was removed or why
- **CLEAN REMOVAL**: Completely remove all traces of deprecated functionality
- **NO DEAD CODE**: Remove entire functions, variables, and imports that are no longer needed
- **NO EXPLANATORY COMMENTS**: Never add comments like "// Removed X", "// Deprecated", or "// No longer needed"
- **COMPLETE CLEANUP**: If removing a feature, remove ALL related code including:
  - Function definitions
  - Variable declarations
  - Import statements
  - Test functions
  - Documentation references
  - Configuration options
- **Examples of what NOT to do**:
  ```go
  // BAD - Don't do this:
  // Removed deprecated auth method
  // func oldAuthMethod() { } // Deprecated
  
  // GOOD - Just remove it completely with no trace
  ```
- **Principal**: When deprecating/removing code, act as a principal engineer - leave the codebase cleaner with no remnants of removed functionality