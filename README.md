# LogGuardian - AWS CloudWatch Log Compliance Automation

**Enterprise-grade automation for CloudWatch log group encryption, retention, and compliance monitoring**

<!-- Core Status Badges -->
[![Build Status](https://github.com/zsoftly/logguardian/workflows/CI/badge.svg)](https://github.com/zsoftly/logguardian/actions)
[![Security Scan](https://img.shields.io/badge/Security-GoSec%20%E2%9C%93-green.svg)](https://github.com/zsoftly/logguardian/actions)
[![Vulnerabilities](https://img.shields.io/badge/Vulnerabilities-0-brightgreen.svg)](https://github.com/zsoftly/logguardian/actions)

<!-- Technology & Platform -->
[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8.svg)](https://golang.org/)
[![AWS](https://img.shields.io/badge/AWS-CloudWatch-orange.svg)](https://aws.amazon.com/cloudwatch/)

<!-- Legal & Compliance -->
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Security](https://img.shields.io/badge/Security-Compliance-blue.svg)]()

## What LogGuardian Does

LogGuardian automatically ensures your AWS CloudWatch log groups meet enterprise compliance standards:

- **Encryption**: Enforces KMS encryption on all log groups
- **Retention**: Sets appropriate log retention policies  
- **Compliance**: Continuous monitoring via AWS Config rules
- **Automation**: Zero-touch remediation and reporting

## Quick Deploy

### AWS Serverless Application Repository (Recommended)
[![Deploy from AWS SAR](https://img.shields.io/badge/Deploy-AWS%20SAR-FF9900?style=for-the-badge&logo=amazon-aws)](https://serverlessrepo.aws.amazon.com/applications/ca-central-1/410129828371/LogGuardian)

**AWS Console:**
1. Click "Deploy" button above
2. Configure parameters as needed
3. Deploy to your AWS account

**AWS CLI:**
```bash
# Get the application (using latest version)
aws serverlessrepo create-cloud-formation-template \
  --application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
  --region ca-central-1

# Deploy the template
aws cloudformation deploy \
  --template-file template.yaml \
  --stack-name logguardian \
  --capabilities CAPABILITY_NAMED_IAM
```

### Manual Deployment
```bash
# Clone the repository
git clone https://github.com/zsoftly/logguardian.git
cd logguardian

# Build and package Lambda
make build && make package

# Deploy using AWS SAM
make sam-deploy-dev
```

## Key Features

- **Flexible Infrastructure**: Works with new or existing AWS Config, KMS keys, and rules
- **Automated Scheduling**: EventBridge-based compliance checks
- **Enterprise Ready**: Supports existing infrastructure and custom policies
- **Multi-Region**: Deploy across multiple AWS regions
- **Monitoring**: CloudWatch metrics and dashboards

## � Documentation & Support

**📚 Comprehensive Documentation:**
- **[Problem Statement & Solution](docs/problem-statement-solution.md)** - Detailed problem analysis and solution overview
- **[Architecture & Workflow](docs/architecture-diagrams.md)** - Mermaid diagrams showing how LogGuardian works
- **[Configuration Parameters](docs/configuration-parameters.md)** - Complete parameter guide with enterprise examples
- **[Deployment Examples](docs/deployment-examples.md)** - AWS CLI, CloudFormation, and Terraform examples
- **[Go Lambda Function](docs/go-lambda-function.md)** - Lambda function implementation details
- **[Config Rule Evaluation](docs/config-rule-evaluation.md)** - Batch processing and compliance analysis
- **[KMS Encryption Validation](docs/kms-encryption-validation.md)** - KMS key validation and cross-region support
- **[Local Testing](docs/local-testing.md)** - Comprehensive local Lambda testing with 9+ test scenarios
- **[Development Guide](docs/development.md)** - Development setup and guidelines
- **[🚀 Complete Deployment Guide](DEPLOYMENT.md)** - Full deployment instructions

**🆘 Support:**
- [Report Issues](https://github.com/zsoftly/logguardian/issues)
- [Discussions](https://github.com/zsoftly/logguardian/discussions)

## Contributing

We welcome contributions! Please see our [Development Guide](docs/development.md) for details.

```bash
# Clone and setup
git clone https://github.com/zsoftly/logguardian.git
cd logguardian

# Install dependencies and run tests
make test
```

## Professional Services

Need help with enterprise deployment? **ZSoftly Technologies Inc** provides professional AWS consulting services.

**🌐 [ZSoftly Cloud Services](https://cloud.zsoftly.com/)**

**📞 Contact Information:**
- **Phone:** +1 (343) 503-0513
- **Email:** info@zsoftly.com
- **Address:** 116 Albert Street, Suite 300, Ottawa, Ontario K1P 5G3
- **Business Hours:** Mon–Fri: 6 AM–10 PM EST
- **[Book Online Consultation](https://cloud.zsoftly.com/)**

## License

MIT License - see the [LICENSE](LICENSE) file for details.

---

**Built by [ZSoftly Technologies Inc](https://zsoftly.com) | Made in Canada | [Professional Services](https://cloud.zsoftly.com/)**