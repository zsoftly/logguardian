# Development Guide

## Getting Started

### Prerequisites

- Go 1.24 or later
- AWS CLI configured
- Required AWS permissions for CloudWatch Logs, KMS, and Config

### Project Setup

```bash
# Clone the repository
git clone https://github.com/zsoftly/logguardian.git
cd logguardian

# Install dependencies
go mod download

# Run tests
make test
```

## Build System

The project uses a comprehensive Makefile for all development tasks:

### Core Commands

```bash
make build             # Build Lambda binary
make package           # Create deployment ZIP
make test              # Run tests
make test-coverage     # Generate coverage report
make clean             # Clean build artifacts
```

### Quality Assurance

```bash
make lint              # Run linter (uses .golangci.yml)
make security          # Run security scan (gosec)
make vuln-check        # Check for vulnerabilities
make check             # Run all quality checks
```

### Development Tools

```bash
make dev-build         # Fast build for development
make memory-profile    # Memory usage analysis
make cpu-profile       # CPU usage analysis
make bench             # Run benchmarks
make size              # Show binary and package size
```

## Code Standards

### Go Version

- Use Go 1.24 (latest stable version)
- Take advantage of new language features
- Follow Go 1.24 best practices

### Code Quality

The project enforces strict code quality standards:

- **Linting**: golangci-lint with comprehensive rule set
- **Security**: gosec security scanner
- **Testing**: >90% test coverage required
- **Documentation**: All public functions documented

### Architecture Principles

1. **Dependency Injection**: AWS clients injected for testability
2. **Error Handling**: All errors explicitly handled
3. **Context Propagation**: context.Context used throughout
4. **Structured Logging**: slog for all logging
5. **Memory Efficiency**: Optimized for Lambda constraints

## Testing Strategy

### Unit Tests

```bash
# Run all tests
make test

# Run tests with race detection
go test -race ./...

# Run specific test
go test -run TestComplianceHandler ./internal/handler
```

### Test Structure

- **Mock Services**: AWS services mocked for unit tests
- **Table-Driven Tests**: Comprehensive test cases
- **Error Testing**: Both success and failure paths tested
- **Benchmarks**: Performance regression prevention

### Coverage Requirements

- Minimum 90% test coverage
- All error paths tested
- Edge cases covered

## CI/CD Pipeline

### GitHub Actions Workflows

The project uses two main workflows:

1. **CI Workflow** (`.github/workflows/ci.yml`):
   - Linting and code quality
   - Security scanning
   - Multi-platform testing
   - Lambda function build
   - Benchmark testing

2. **Release Workflow** (`.github/workflows/release.yml`):
   - Automated releases on version tags
   - Binary artifacts generation
   - Changelog generation
   - GitHub release creation

### Quality Gates

All PRs must pass:
- Linting (golangci-lint)
- Security scanning (gosec, govulncheck)
- Tests on multiple Go versions (1.23, 1.24)
- Tests on multiple platforms (Linux, macOS, Windows)

## Local Development

### Environment Setup

```bash
# Install development tools
make install-tools

# This installs:
# - golangci-lint (linting)
# - gosec (security scanning)
# - govulncheck (vulnerability scanning)
# - mockgen (mock generation)
```

### Development Workflow

1. **Create Feature Branch**:
   ```bash
   git checkout -b feature/new-feature
   ```

2. **Make Changes**: Follow code standards

3. **Test Locally**:
   ```bash
   make check  # Run all quality checks
   ```

4. **Commit Changes**:
   ```bash
   git commit -m "feat: add new feature"
   ```

5. **Push and Create PR**:
   ```bash
   git push origin feature/new-feature
   ```

### Debugging

#### Local Testing

```bash
# Build development version
make dev-build

# Test with sample event
make test-local
```

#### Memory Debugging

```bash
# Generate memory profile
make memory-profile

# View profile
go tool pprof mem.prof
```

#### CPU Profiling

```bash
# Generate CPU profile
make cpu-profile

# View profile
go tool pprof cpu.prof
```

## AWS Integration

### Local AWS Testing

```bash
# Set up local AWS credentials
aws configure

# Test AWS connectivity
aws sts get-caller-identity

# Test Lambda deployment
aws lambda update-function-code \
  --function-name logguardian-compliance \
  --zip-file fileb://dist/logguardian-compliance.zip
```

### Environment Variables

For local development, set these environment variables:

```bash
export KMS_KEY_ALIAS="alias/cloudwatch-logs-compliance"
export DEFAULT_RETENTION_DAYS="365"
export SUPPORTED_REGIONS="us-east-1,us-west-2"
export DRY_RUN="true"  # For safe testing
```

## Code Organization

### Directory Structure

```
├── cmd/lambda/              # Application entry points
├── internal/               # Private application code
│   ├── handler/           # HTTP/event handlers
│   ├── service/           # Business logic
│   └── types/             # Data structures and models
├── testdata/              # Test fixtures and sample data
├── docs/                  # Documentation
├── .github/               # GitHub workflows and templates
└── dist/                  # Build artifacts (generated)
```

### Package Guidelines

- **cmd/**: Application entry points only
- **internal/**: Private application logic
- **internal/handler/**: Event processing logic
- **internal/service/**: Core business logic
- **internal/types/**: Shared data structures

### Import Organization

Follow Go conventions for import grouping:

```go
import (
    // Standard library
    "context"
    "fmt"
    "log/slog"

    // Third-party packages
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go-v2/config"

    // Local packages
    "github.com/zsoftly/logguardian/internal/service"
    "github.com/zsoftly/logguardian/internal/types"
)
```

## Performance Considerations

### Lambda Optimization

- **Binary Size**: Keep deployment package < 50MB
- **Cold Start**: Optimize for fast initialization
- **Memory Usage**: Monitor heap allocation
- **Execution Time**: Target < 15 minutes

### Memory Management

- **Client Pooling**: Reuse AWS SDK clients
- **Buffer Pooling**: Reuse byte buffers
- **Garbage Collection**: Minimize allocations
- **Memory Monitoring**: Track usage patterns

## Security Guidelines

### Code Security

- Never log sensitive information
- Use context for request cancellation
- Validate all input data
- Handle errors gracefully

### AWS Security

- Follow principle of least privilege
- Use customer-managed KMS keys
- Enable CloudTrail logging
- Monitor function metrics

## Troubleshooting

### Common Issues

1. **Build Failures**: Check Go version compatibility
2. **Test Failures**: Verify AWS credentials for integration tests
3. **Lint Errors**: Run `make fmt` to fix formatting
4. **Memory Issues**: Use profiling tools to identify leaks

### Getting Help

1. Check existing issues on GitHub
2. Review logs in CloudWatch
3. Use debugging tools for profiling
4. Consult AWS documentation for service limits