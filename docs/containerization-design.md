# LogGuardian Containerization Design Document

## Implementation Status

### Completed
- Multi-stage Dockerfile with Alpine base image
- Docker Compose configuration
- CLI interface (`cmd/container/main.go`)
- AWS authentication chain (`internal/container/auth.go`) 
- Config rule processor (`internal/container/processor.go`)
- Dry-run mode (`internal/container/dryrun.go`)
- Unit tests for all components

### TBD
- ECS task definitions and EventBridge integration
- CloudWatch Metrics and distributed tracing
- ECR push workflows and CI/CD pipeline
- Kubernetes manifests and Helm charts
- Enhanced health checks and graceful shutdown
- Correlation IDs and structured logging
- Lambda-to-container migration tools
- Main README updates with container usage

## Executive Summary

This document outlines the design principles, architectural patterns, and implementation strategy for containerizing the LogGuardian Lambda application. The containerized version will maintain functional parity with the Lambda implementation while providing deployment flexibility across container orchestration platforms, starting with Amazon ECS.

## Design Goals

### Primary Objectives
- **Functional Parity**: Maintain identical business logic and compliance remediation capabilities
- **Platform Agnostic**: Design for portability across ECS, Kubernetes, and other container platforms
- **Ephemeral Execution**: Run-to-completion pattern with automatic resource cleanup
- **Configuration Flexibility**: Support multiple parameter injection methods
- **Authentication Abstraction**: Seamless credential handling across environments

### Non-Functional Requirements
- **Performance**: Sub-second container startup time
- **Security**: Principle of least privilege, no embedded credentials
- **Observability**: Structured logging with correlation capabilities
- **Cost Optimization**: Minimal resource allocation, pay-per-execution model
- **Maintainability**: Single codebase for multiple deployment targets

## Architectural Patterns

### 1. Command Pattern Implementation

The container will implement the Command Pattern, treating each execution as a discrete command with:
- **Command Interface**: CLI arguments or environment variables as input
- **Command Processor**: Core business logic from Lambda handler
- **Command Result**: Exit codes and structured logs as output

This pattern enables:
- Decoupling of trigger mechanisms from business logic
- Easy testing through command simulation
- Clear separation of concerns

### 2. Adapter Pattern for AWS Services

An adapter layer will abstract AWS service interactions:
- **Authentication Adapter**: Handles multiple credential sources
- **Service Adapter**: Wraps AWS SDK calls with retry logic
- **Configuration Adapter**: Normalizes configuration from various sources

Benefits:
- Simplified testing with mock adapters
- Consistent error handling
- Platform-specific optimizations

### 3. Strategy Pattern for Authentication

Multiple authentication strategies will be supported through a common interface:
- **Task Role Strategy**: For ECS/Fargate deployments
- **Profile Strategy**: For local development
- **Environment Strategy**: For CI/CD pipelines
- **Instance Profile Strategy**: For EC2 deployments

The appropriate strategy is selected at runtime based on environment detection.

## Design Decisions

### Container vs Serverless Trade-offs

| Aspect | Lambda (Current) | Container (Proposed) | Decision Rationale |
|--------|------------------|---------------------|-------------------|
| **Startup Time** | Cold start (~1s) | Container pull + start (~3-5s) | Acceptable for batch processing |
| **Execution Limit** | 15 minutes | Unlimited | Enables future long-running tasks |
| **Deployment** | Function code only | Full container image | Better dependency control |
| **Scaling** | Automatic | Orchestrator-dependent | Predictable workload doesn't need auto-scaling |
| **Cost Model** | Per invocation | Per execution time | Similar costs for short tasks |

### Ephemeral vs Long-Running Design

**Decision: Ephemeral (Run-to-Completion)**

Rationale:
- Matches current Lambda execution model
- Simplifies state management (stateless)
- Reduces attack surface (short-lived containers)
- Optimizes costs (no idle resources)
- Enables easy migration between platforms

### Configuration Injection Methods

**Primary Method: CLI Arguments**
- Clear contract definition
- Easy to test and debug
- Platform agnostic
- Supports complex parameter structures
- Includes --dry-run flag for safe testing

**Secondary Method: Environment Variables**
- AWS service configuration (regions, endpoints)
- Feature flags and operational settings
- Secrets from orchestrator secret management

**Avoided: Configuration Files**
- Adds complexity to image building
- Reduces deployment flexibility
- Complicates secret management

## Authentication Architecture

### Credential Chain Design

The container will implement a hierarchical credential resolution pattern:

1. **Explicit Credentials** (highest priority)
   - Command-line specified profiles
   - Directly provided credentials

2. **Implicit Container Credentials**
   - ECS task roles via metadata service
   - Kubernetes service accounts via IRSA
   
3. **Environment Credentials**
   - Standard AWS environment variables
   - Temporary session tokens

4. **Default Credentials** (lowest priority)
   - EC2 instance profiles
   - Default profile from credentials file

### Security Principles

- **No Embedded Credentials**: Images never contain AWS credentials
- **Temporary Credentials Only**: Use STS tokens with time limits
- **Least Privilege**: Task roles with minimal required permissions
- **Credential Isolation**: Separate roles for different environments

## Container Image Design

### Base Image Selection

**Decision: Alpine Linux**

Rationale:
- Minimal attack surface (~5MB base)
- Fast download and startup
- Sufficient for Go static binaries
- Active security maintenance

### Build Strategy

**Multi-Stage Build Pattern**
- **Build Stage**: Full development environment
- **Runtime Stage**: Minimal runtime dependencies

Benefits:
- Smaller final images (~20MB total)
- No build tools in production
- Consistent build environment
- Cached layer optimization

### User and Permission Model

- Run as non-root user (UID 1000)
- Read-only root filesystem capability
- Explicit directory permissions for required paths
- No sudo or privilege escalation

## Orchestration Integration

### ECS/Fargate Design

**Task Definition Pattern**
- Single container per task
- Task-level IAM roles
- CloudWatch Logs integration
- Health check endpoints

**Triggering Mechanisms**
1. EventBridge → ECS RunTask
2. Lambda Orchestrator → ECS API
3. Step Functions → ECS Task State
4. Manual invocation via CLI/Console

### Future Kubernetes Design

**Job/CronJob Pattern**
- Kubernetes Job for one-time execution
- CronJob for scheduled runs
- ConfigMap for configuration
- ServiceAccount for AWS permissions (IRSA)

**Key Differences from ECS**
- Pod security policies
- Network policies for isolation
- Prometheus metrics vs CloudWatch

## Monitoring and Observability

### Logging Strategy

**Structured Logging**
- JSON format for machine parsing
- Correlation IDs for request tracing
- Log levels: ERROR, WARN, INFO, DEBUG
- Contextual fields for filtering

**Log Aggregation**
- CloudWatch Logs (ECS)
- Fluentd/Elasticsearch (Kubernetes)
- Consistent format across platforms

### Metrics Collection

**Key Metrics**
- Task execution count
- Processing duration
- Success/failure rates
- Resource utilization

**Collection Methods**
- CloudWatch Metrics (AWS)
- Prometheus (Kubernetes)
- StatsD (platform agnostic)

### Distributed Tracing

- AWS X-Ray integration for ECS
- OpenTelemetry for platform independence
- Trace context propagation

## Migration Strategy

### Phase 1: Parallel Development
- Maintain Lambda function unchanged
- Develop container version with same inputs/outputs
- Establish testing parity benchmarks

### Phase 2: Shadow Testing
- Deploy container to non-production
- Run both implementations in parallel
- Compare results and performance
- Validate compliance outcomes

### Phase 3: Gradual Rollout
- Start with non-critical Config rules
- Monitor for 1-2 weeks per rule
- Progressive migration based on success criteria
- Maintain rollback capability

### Phase 4: Decommissioning
- Remove Lambda infrastructure
- Archive serverless templates
- Update documentation and runbooks

## Testing Strategy

### Unit Testing
- Mock AWS service calls
- Test command parsing logic
- Validate error handling

### Integration Testing
- Local container execution
- AWS service integration with test accounts
- End-to-end compliance scenarios

### Performance Testing
- Container startup time benchmarks
- Memory and CPU profiling
- Concurrent execution limits

### Security Testing
- Container vulnerability scanning
- IAM permission validation
- Network isolation verification

## Cost Optimization

### Resource Sizing
- Start with minimal allocations (256 CPU, 512MB RAM)
- Monitor actual usage patterns
- Right-size based on p95 metrics

### Execution Optimization
- Batch processing to reduce invocations
- Parallel processing within limits
- Early termination on errors

### Platform Selection
- Fargate Spot for non-critical workloads
- Reserved capacity for predictable schedules
- Consider EC2 for high-volume scenarios

## Risk Analysis

### Technical Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Container startup delays | Missed SLA | Pre-pulled images, warm pools |
| Authentication failures | Service disruption | Multiple auth methods, fallbacks |
| Resource exhaustion | Task failures | Limits, monitoring, auto-scaling |
| Network issues | API call failures | Retries, circuit breakers |

### Operational Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Complex debugging | Longer MTTR | Comprehensive logging, tracing |
| Deployment failures | Service unavailability | Blue-green deployments, rollback |
| Configuration drift | Inconsistent behavior | GitOps, infrastructure as code |

## Success Criteria

### Functional Success
- 100% parity with Lambda functionality
- Same compliance remediation outcomes
- No regression in error rates
- Idempotent operations (safe to re-run)
- Execution reports generated for each run

### Performance Success
- Container startup < 5 seconds
- Processing time within 10% of Lambda
- Memory usage < 256MB
- Handles AWS API rate limiting gracefully

### Operational Success
- Deployment time < 5 minutes
- Clear error notifications to ops team
- Audit trail of all compliance changes
- Dry-run mode validates changes before applying

## Future Enhancements

### Short Term (3-6 months)
- Kubernetes deployment support
- Multi-region container registries
- Enhanced observability dashboards

### Medium Term (6-12 months)
- Sidecar pattern for monitoring
- Service mesh integration
- Automated performance tuning

### Long Term (12+ months)
- Multi-cloud support (Azure, GCP)
- Native cloud service integrations
- AI-driven optimization

## Decision Log

| Date | Decision | Rationale | Alternatives Considered |
|------|----------|-----------|------------------------|
| 2024-09 | CLI-based interface | Simplicity, testability | HTTP API, Message queue |
| 2024-09 | Alpine base image | Security, size | Distroless, Ubuntu |
| TBD | ECS Fargate | Serverless containers | EC2, Kubernetes |
| 2024-09 | Multi-stage builds | Image optimization | Single stage, BuildKit |
| 2024-09 | Idempotent operations | Safe re-runs for weekly jobs | Stateful tracking |
| 2024-09 | Dry-run mode | Safe testing of compliance changes | Direct execution only |

## Appendices

### A. IAM Permission Matrix

Define minimum required permissions for:
- ECS Task Role
- EventBridge Execution Role
- Container Repository Access
- CloudWatch Logs Access

### B. Network Architecture

- VPC requirements
- Security group rules
- PrivateLink endpoints
- Internet Gateway needs

### C. Disaster Recovery

- Backup strategies
- Recovery time objectives
- Failover procedures
- Data consistency guarantees

## Key Implementation Requirements

### Essential Features for Weekly Compliance Jobs
1. **Idempotency**: All operations must be safe to re-run without side effects
2. **Dry-run Mode**: `--dry-run` flag to preview changes without applying them
3. **Execution Reporting**: Generate JSON reports of processed/compliant/failed resources
4. **Rate Limiting**: Implement exponential backoff for AWS API throttling
5. **Error Notifications**: SNS/Slack integration for failure alerts
6. **Audit Logging**: Track all compliance changes with execution ID and timestamp

## Approval and Sign-off

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Technical Lead | | | |
| Security Architect | | | |
| DevOps Lead | | | |
| Product Owner | | | |

---

**Document Version**: 1.0.1  
**Last Updated**: 2024-09-10  
**Status**: Draft - Partial Implementation  
**Classification**: Internal  
**Review Cycle**: Quarterly