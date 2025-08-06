# AWS Marketplace SAM Deployment Guide for LogGuardian

## Overview

This document provides step-by-step instructions for deploying LogGuardian to the AWS Marketplace using AWS SAM (Serverless Application Model). The SAM approach is the recommended pattern for AWS Marketplace serverless applications.

## Prerequisites

### 1. Install Required Tools

```bash
# Install AWS SAM CLI
pip install aws-sam-cli

# Install AWS CLI
pip install awscli

# Verify installations
sam --version
aws --version
```

### 2. Configure AWS Credentials

```bash
# Configure AWS CLI with marketplace publisher credentials
aws configure
```

### 3. Set Environment Variables

```bash
export MARKETPLACE_BUCKET="your-marketplace-bucket"
export AWS_REGION="us-east-1"  # AWS Marketplace requires us-east-1
```

## AWS Marketplace SAM Pattern

### Template Structure

The `template.yaml` follows AWS Marketplace best practices:

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Transform: AWS::Serverless-2016-10-31  # Required for SAM

Metadata:
  AWS::ServerlessRepo::Application:    # Required for marketplace
    Name: LogGuardian
    Description: 'Enterprise CloudWatch compliance automation'
    Author: 'ZSoftly Technologies Inc'
    # ... other metadata

Resources:
  LogGuardianFunction:
    Type: AWS::Serverless::Function     # SAM resource type
    Properties:
      CodeUri: build/                   # Directory containing binary
      Handler: bootstrap                # Go Lambda handler
      Runtime: provided.al2023          # Go custom runtime
      # ... other properties
```

### Binary Packaging

For Go Lambda functions, AWS Marketplace expects:

1. **Binary name**: `bootstrap` (AWS Lambda AL2023 requirement)
2. **Location**: `build/bootstrap` 
3. **Architecture**: `x86_64` (linux/amd64)
4. **CGO**: Disabled (`CGO_ENABLED=0`)

## Deployment Steps

### 1. Build and Package

```bash
# Build the Go binary
make build

# Prepare SAM build structure
make sam-build

# Validate SAM template
make sam-validate
```

### 2. Local Testing (Optional)

```bash
# Test locally with sample event
make sam-local-invoke

# Start local API for testing
make sam-local-start
```

### 3. Package for Marketplace

```bash
# Package application for marketplace
MARKETPLACE_BUCKET=your-bucket make sam-package-marketplace
```

This creates `packaged-template.yaml` with S3 references.

### 4. Publish to AWS Serverless Application Repository

```bash
# Publish to AWS SAR (required for marketplace)
make sam-publish
```

### 5. Submit to AWS Marketplace

1. **AWS Marketplace Management Portal**:
   - Log into AWS Marketplace Management Portal
   - Create new product listing
   - Select "Serverless Application" product type

2. **Upload SAM Template**:
   - Upload the `packaged-template.yaml`
   - AWS will validate the template structure

3. **Configure Marketplace Listing**:
   - Product title: "LogGuardian - CloudWatch Compliance Automation"
   - Short description: "Enterprise-grade automation for CloudWatch log group encryption, retention, and compliance monitoring"
   - Pricing model: Choose appropriate pricing

## SAM vs Traditional CloudFormation

### Benefits of SAM for Marketplace:

| Feature | Traditional CF | SAM |
|---------|----------------|-----|
| **Lambda Packaging** | Manual ZIP creation | Automatic with `sam build` |
| **Local Testing** | Not supported | `sam local` commands |
| **Marketplace Integration** | Manual process | Built-in SAR publishing |
| **Template Validation** | Basic CF validation | SAM-specific validation |
| **Binary Handling** | Manual S3 upload | Automatic with `CodeUri` |

### SAM Advantages:

1. **Simplified Lambda Deployment**: SAM automatically handles binary packaging and S3 upload
2. **Built-in Marketplace Support**: `AWS::ServerlessRepo::Application` metadata
3. **Local Development**: Test Lambda functions locally before deployment
4. **Automatic IAM**: SAM can generate IAM policies from function requirements
5. **Event Source Mapping**: Simplified EventBridge/CloudWatch Events integration

## Template Migration

Your existing CloudFormation templates can be easily converted:

```yaml
# Before (CloudFormation)
LogGuardianLambda:
  Type: AWS::Lambda::Function
  Properties:
    Code:
      S3Bucket: !Ref DeploymentBucket
      S3Key: !Ref LambdaCodeKey
    Handler: main
    Runtime: provided.al2023

# After (SAM)
LogGuardianFunction:
  Type: AWS::Serverless::Function
  Properties:
    CodeUri: build/        # Local directory
    Handler: bootstrap     # Binary name
    Runtime: provided.al2023
    Events:               # Simplified event configuration
      Schedule:
        Type: Schedule
        Properties:
          Schedule: !Ref ScheduleExpression
```

## Marketplace Package Contents

Your marketplace package will contain:

```
logguardian-marketplace.zip
├── template.yaml          # SAM template
├── build/
│   └── bootstrap         # Go binary
├── README.md            # Marketplace description
├── LICENSE              # License file
└── USAGE.md            # Usage instructions
```

## Best Practices

### 1. Template Organization
- Use `Globals` section for common Lambda properties
- Leverage SAM `Events` for simplified event source mapping
- Include comprehensive `Metadata` section

### 2. Binary Optimization
- Compile with `-ldflags="-s -w"` to reduce binary size
- Use `provided.al2023` runtime for optimal performance
- Ensure binary is statically linked (`CGO_ENABLED=0`)

### 3. Marketplace Requirements
- Include detailed `README.md` with usage instructions
- Provide clear parameter descriptions
- Set appropriate default values
- Include comprehensive output values

### 4. Testing Strategy
```bash
# Local testing workflow
make sam-build
make sam-local-invoke
make test
make security
make sam-validate
```

## Troubleshooting

### Common Issues:

1. **Binary Architecture Mismatch**
   ```bash
   # Ensure Linux x86_64 compilation
   GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build
   ```

2. **SAM Template Validation Errors**
   ```bash
   # Validate template structure
   sam validate --template template.yaml --region us-east-1
   ```

3. **Marketplace Submission Rejections**
   - Ensure all required metadata fields are populated
   - Test template deployment in multiple regions
   - Verify IAM permissions are minimal but functional

### Getting Help:

- **SAM CLI Issues**: [AWS SAM GitHub](https://github.com/aws/aws-sam-cli)
- **Marketplace Issues**: AWS Marketplace Partner Support
- **LogGuardian Issues**: [GitHub Issues](https://github.com/zsoftly/logguardian/issues)

## Conclusion

Using AWS SAM for marketplace deployment provides:

✅ **Simplified Packaging**: Automatic binary handling and S3 upload  
✅ **Better Testing**: Local development and testing capabilities  
✅ **Marketplace Integration**: Built-in AWS Serverless Application Repository support  
✅ **Standardized Pattern**: Following AWS best practices for serverless marketplace apps  

The SAM pattern is definitely the recommended approach for your LogGuardian AWS Marketplace deployment.
