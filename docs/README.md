# LogGuardian Documentation

## 📚 Documentation Index

### Getting Started
- [**Architecture Overview**](architecture-overview.md) 🏗️ - System design and components
- [**Problem Statement & Solution**](problem-statement-solution.md) - Why LogGuardian exists
- [**Configuration Parameters**](configuration-parameters.md) - All available parameters
- [**Deployment Examples**](deployment-examples.md) - Real-world deployment scenarios

### Deployment Guides
- [**AWS Marketplace SAM**](aws-marketplace-sam.md) - Publishing to AWS Marketplace
- [**Upgrade Guide**](upgrade-guide.md) - Version upgrade instructions

### Technical Deep Dives
- [**Config Rule Evaluation**](config-rule-evaluation.md) - How Config rules work
- [**Rule Classification**](rule-classification.md) - Encryption vs Retention logic
- [**KMS Batch Optimization**](kms-batch-optimization.md) - Performance optimizations
- [**KMS Encryption Validation**](kms-encryption-validation.md) - Security validation process

### Configuration & Customization
- [**LogLevel Configuration**](loglevel-configuration.md) - Logging configuration
- [**Resource Tagging Strategy**](resource-tagging-strategy.md) - Tagging best practices
- [**Customer Infrastructure Integration**](customer-infrastructure-integration.md) - Enterprise integration
- [**Template Optimization Guide**](template-optimization-guide.md) - CloudFormation optimization

### Development & Testing
- [**Development Guide**](development.md) - Local development setup
- [**Local Testing**](local-testing.md) - Testing strategies
- [**Branch Protection**](branch-protection.md) - Git workflow

### Examples
- [**Examples**](examples.md) - Common use cases

## 🚀 Quick Start

### Deploy from SAR (Recommended)
1. Go to [AWS Serverless Application Repository](https://console.aws.amazon.com/serverlessrepo)
2. Search for "LogGuardian"
3. Click Deploy
4. Configure parameters
5. Deploy stack

### Deploy with SAM CLI
```bash
sam deploy \
  --template-file template.yaml \
  --stack-name logguardian \
  --parameter-overrides \
    Environment=prod \
    CreateConfigService=false \
    DefaultRetentionDays=30 \
  --capabilities CAPABILITY_IAM
```

## 🏛️ Architecture Highlights

```
EventBridge → Lambda → CloudWatch Logs
     ↓          ↓            ↓
  Schedule    Config      KMS/Retention
```

- **Event-driven**: Responds to Config compliance events
- **Scheduled**: Batch processing via EventBridge
- **Optimized**: KMS validation caching, parallel processing
- **Secure**: All logs encrypted with KMS
- **Compliant**: Tracks remediation via Config

## 📊 Key Features

| Feature | Description |
|---------|-------------|
| **Auto-Encryption** | Automatically encrypt unencrypted log groups |
| **Retention Management** | Enforce retention policies |
| **Batch Processing** | Process multiple resources efficiently |
| **Config Integration** | Track compliance status |
| **Cost Optimization** | S3 lifecycle, efficient Lambda sizing |
| **Multi-Region** | Deploy across multiple regions |

## 🔧 Configuration Options

### Essential Parameters
- `Environment`: prod, staging, dev
- `DefaultRetentionDays`: 1-3653 days
- `CreateConfigService`: true/false (default: false)
- `CreateKMSKey`: true/false (default: true)

### Advanced Parameters
- `LogLevel`: ERROR, WARN, INFO, DEBUG
- `LambdaMemorySize`: 128-3008 MB (default: 128)
- `LambdaTimeout`: 1-900 seconds (default: 60)
- `EncryptionScheduleExpression`: Cron/rate expression
- `RetentionScheduleExpression`: Cron/rate expression

## 📈 Monitoring

- **CloudWatch Dashboard**: Real-time metrics
- **Lambda Logs**: Structured JSON logging
- **Config Dashboard**: Compliance tracking
- **Cost Explorer**: Track spending

## 🔒 Security

- **Encryption**: AES-256 with AWS KMS
- **IAM**: Least privilege access
- **Tagging**: Comprehensive resource tagging
- **Audit**: All actions logged

## 🎯 Use Cases

1. **Compliance**: Meet regulatory requirements (SOC2, HIPAA, PCI-DSS)
2. **Security**: Ensure all logs are encrypted
3. **Cost Management**: Enforce retention to reduce storage costs
4. **Governance**: Centralized log management policy
5. **Automation**: Hands-off compliance management

## 📝 Latest Updates

### v1.2.6 (Current)
- Performance: Skip KMS validation for retention rules
- Fix: Proper code formatting with go fmt

### v1.2.5
- UX: Improved parameter descriptions for SAR

### v1.2.4
- Removed CustomerTagPrefix parameter
- Fixed ConfigBucket export issue
- Set CreateConfigService default to false

## 🤝 Contributing

See [Development Guide](development.md) for local setup and contribution guidelines.

## 📄 License

MIT License - See LICENSE file for details

## 🏢 Publisher

**ZSoftly Technologies Inc**
- GitHub: https://github.com/zsoftly/logguardian
- SAR: Search "LogGuardian" in AWS Serverless Application Repository