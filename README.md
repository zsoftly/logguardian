# LogGuardian

**Automated CloudWatch Log Groups Compliance Automation**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![AWS](https://img.shields.io/badge/AWS-CloudWatch-orange.svg)](https://aws.amazon.com/cloudwatch/)
[![Security](https://img.shields.io/badge/Security-Compliance-blue.svg)]()

> Enterprise-grade automation for CloudWatch log group encryption, retention, and compliance monitoring

## 🚀 Quick Start

### One-Click AWS Marketplace Deployment
[![Deploy from AWS Marketplace](https://img.shields.io/badge/Deploy-AWS%20Marketplace-FF9900?style=for-the-badge&logo=amazon-aws)](https://aws.amazon.com/marketplace/pp/prodview-logguardian)

**→ [Launch LogGuardian from AWS Marketplace](https://aws.amazon.com/marketplace/pp/prodview-logguardian)**

### Manual Deployment
```bash
# Clone the repository
git clone https://github.com/zsoftly/logguardian.git
cd logguardian

# Deploy using CloudFormation
aws cloudformation deploy \
  --template-file templates/logguardian.yaml \
  --stack-name logguardian \
  --capabilities CAPABILITY_IAM
```

### Go Lambda Function
For developers wanting to build and customize the Lambda function:

```bash
# Build the Go Lambda function
make build && make package

# Run tests and security scans
make test && make security
```

**📖 [Go Lambda Function Documentation](docs/go-lambda-function.md)**

## 🛠️ Implementation Status

**✅ Completed:**
- Go 1.24 Lambda function with AWS SDK v2
- AWS Config event processing and compliance analysis
- KMS encryption and retention policy remediation
- Multi-region support with memory optimization
- Comprehensive test suite with mocked AWS services
- CI/CD pipeline with security scanning (GoSec, govulncheck)
- Structured logging with Go's slog package

## 📋 Table of Contents
- [Problem Statement](#problem-statement)
- [Solution Overview](#solution-overview)
- [Features](#features)
- [Architecture](#architecture)
- [Deployment Options](#deployment-options)
- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
- [Usage](#usage)
- [Cost Analysis](#cost-analysis)
- [Contributing](#contributing)
- [License](#license)

## ❗ Problem Statement

AWS customers struggle with maintaining CloudWatch log group compliance across their organization due to:

### **Operational Challenges**
- **Manual Compliance Management**: Organizations must manually check hundreds or thousands of log groups for KMS encryption and retention policy compliance
- **Scale Challenges**: As organizations grow, manual compliance checking becomes impossible to maintain
- **Operational Overhead**: DevOps teams spend significant time on repetitive compliance tasks

### **Financial Impact**
- **Cost Inefficiency**: Log groups without retention policies accumulate indefinitely, leading to unexpected storage costs
- **Resource Waste**: Teams over-provision monitoring resources due to inefficient compliance checking

### **Security & Compliance Risks**
- **Security Gaps**: Unencrypted log groups fail compliance audits and security frameworks
- **Compliance Violations**: Inconsistent retention policies lead to regulatory compliance issues
- **Audit Failures**: Lack of systematic compliance tracking during security reviews

## 🎯 Solution Overview

LogGuardian transforms CloudWatch log group compliance from a manual, error-prone process into an automated, cost-effective, and reliable system that scales with organizational growth while maintaining security and compliance standards.

### **Key Differentiators**
- **Cost-Optimized**: Uses AWS Config Rules instead of expensive continuous Lambda scanning
- **Safe Automation**: Shared responsibility model prevents application disruptions
- **Enterprise-Ready**: Built for multi-account, multi-region AWS environments
- **Compliance-First**: Designed specifically for audit and regulatory requirements

## ✨ Features

### **🔍 Intelligent Compliance Discovery**
- Utilizes AWS Config Rules to efficiently identify non-compliant CloudWatch log groups
- Pre-built compliance rules for encryption and retention requirements
- Configurable compliance standards (365 days retention minimum, customer-managed KMS keys)
- Multi-region compliance monitoring from centralized deployment

### **🛡️ Safe Automated Remediation**
- Automated application of retention policies to non-compliant log groups
- Safe KMS encryption using pre-configured customer-managed keys
- Prerequisite validation to ensure service IAM roles have proper KMS permissions
- Rollback capabilities for failed remediation attempts

### **🤝 Shared Responsibility Model**
- Customer maintains control over KMS key creation and IAM permission management
- Product assumes keys and permissions are pre-configured and tested
- Clear separation of customer vs. automation responsibilities
- Fail-fast approach when prerequisites are not met

### **💰 Cost-Optimized Operations**
- Event-driven remediation based on Config Rule evaluations
- Process only non-compliant resources (typically 5-10% of total log groups)
- Configurable schedule options (daily, weekly, monthly) based on organizational requirements
- Elimination of continuous resource scanning

### **📊 Enterprise Governance**
- Comprehensive compliance reporting and dashboards
- Audit trail of all remediation activities
- Integration with AWS Organizations for multi-account deployments
- Customizable notification and alerting for compliance changes

### **🔧 Flexible Deployment Options**
- Single-region or multi-region deployment configurations
- Support for different compliance schedules per environment (prod vs. dev)
- Granular policy controls for different log group patterns
- Integration with existing CI/CD and infrastructure-as-code workflows

## 🏗️ Architecture

### **High-Level Architecture**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   AWS Config    │    │  EventBridge    │    │ Remediation     │
│   Rules         │────│  Scheduler      │────│ Lambda          │
│                 │    │                 │    │                 │
│ • Encryption    │    │ Day N-1: Config │    │ Day N: Process  │
│ • Retention     │    │ Day N: Lambda   │    │ Non-Compliant   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ CloudWatch Log  │    │   Compliance    │    │    Customer     │
│ Groups          │    │   Dashboard     │    │   KMS Keys      │
│                 │    │                 │    │                 │
│ • Target        │    │ • Reports       │    │ • Pre-created   │
│   Resources     │    │ • Metrics       │    │ • IAM Ready     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### **Process Flow**
1. **Day N-1**: AWS Config Rules evaluate all CloudWatch log groups for compliance
2. **Day N**: EventBridge triggers Lambda function with non-compliant resource list
3. **Remediation**: Lambda applies encryption and retention policies only to non-compliant resources
4. **Reporting**: Compliance dashboard updates with remediation results
5. **Monitoring**: Ongoing compliance monitoring and alerting

## 🚀 Deployment Options

### **Option 1: AWS Marketplace (Recommended)**

**One-click deployment with enterprise support**

[![Deploy Now](https://img.shields.io/badge/Deploy%20Now-AWS%20Marketplace-FF9900?style=for-the-badge)](https://aws.amazon.com/marketplace/pp/prodview-logguardian)

**Benefits:**
- ✅ One-click deployment
- ✅ Pre-configured best practices
- ✅ Enterprise support included
- ✅ Automatic updates
- ✅ 30-day free trial

**Pricing:** Starting at $99/month per AWS account

### **Option 2: Manual CloudFormation Deployment**

**Free open-source deployment**

```bash
# 1. Clone repository
git clone https://github.com/zsoftly/logguardian.git
cd logguardian

# 2. Configure parameters
cp config/example.yaml config/production.yaml
# Edit config/production.yaml with your settings

# 3. Deploy stack
aws cloudformation deploy \
  --template-file templates/logguardian.yaml \
  --stack-name logguardian-prod \
  --parameter-overrides file://config/production.yaml \
  --capabilities CAPABILITY_IAM
```

### **Option 3: Terraform Deployment**

```hcl
module "logguardian" {
  source = "github.com/zsoftly/logguardian//terraform"
  
  # Configuration
  retention_days = 365
  kms_key_alias = "alias/cloudwatch-logs-compliance"
  schedule = "weekly"
  
  # Multi-region support
  regions = ["us-east-1", "us-west-2", "eu-west-1"]
  
  # Notification settings
  notification_email = "compliance@yourcompany.com"
}
```

## 📚 Documentation

- [📘 Go Lambda Function](docs/go-lambda-function.md) - Complete Lambda function documentation
- [🔧 Development Guide](docs/development.md) - Development setup and guidelines
- [🏗️ Architecture Deep Dive](docs/architecture.md) - Technical architecture details
- [🔒 Security Best Practices](docs/security.md) - Security implementation guide
- [🚀 Deployment Guide](docs/deployment.md) - Deployment instructions
- [📊 Monitoring & Alerting](docs/monitoring.md) - Monitoring setup guide

## 🤝 Contributing

We welcome contributions! Please see our [Development Guide](docs/development.md) for details.

### **Quick Start**
```bash
# Clone and setup
git clone https://github.com/zsoftly/logguardian.git
cd logguardian

# Install dependencies and run tests
make test

# See development guide for more details
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Built with ❤️ by [ZSoftly Technologies Inc](https://zsoftly.com)**