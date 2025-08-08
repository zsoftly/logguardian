# LogGuardian - AWS CloudWatch Log Compliance Automation

**Enterprise-grade automation for CloudWatch log group encryption, retention, and compliance monitoring**

**License:** MIT | **Language:** Go 1.24+ | **Platform:** AWS SAM

## What LogGuardian Does

LogGuardian automatically ensures your AWS CloudWatch log groups meet enterprise compliance standards:

- **Encryption**: Enforces KMS encryption on all log groups
- **Retention**: Sets appropriate log retention policies  
- **Compliance**: Continuous monitoring via AWS Config rules
- **Automation**: Zero-touch remediation and reporting

## Quick Deploy

**AWS Console:**
1. Click "Deploy" button above
2. Configure parameters as needed
3. Deploy to your AWS account

**AWS CLI:**
```bash
# Get the application
aws serverlessrepo create-cloud-formation-template \
  --application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
  --semantic-version 1.0.1 \
  --region ca-central-1

# Deploy the template
aws cloudformation deploy \
  --template-file template.yaml \
  --stack-name logguardian \
  --capabilities CAPABILITY_NAMED_IAM
```

## Key Features

- **Flexible Infrastructure**: Works with new or existing AWS Config, KMS keys, and rules
- **Automated Scheduling**: EventBridge-based compliance checks
- **Enterprise Ready**: Supports existing infrastructure and custom policies
- **Multi-Region**: Deploy across multiple AWS regions
- **Monitoring**: CloudWatch metrics and dashboards

**Configuration**: See [Parameter Guide](https://github.com/zsoftly/logguardian/blob/main/docs/configuration-parameters.md) for all deployment options

## Documentation & Support

**📚 Comprehensive Documentation:**
- **[Architecture & Workflow](https://github.com/zsoftly/logguardian/blob/main/docs/architecture-diagrams.md)** - Mermaid diagrams showing how LogGuardian works
- **[Configuration Parameters](https://github.com/zsoftly/logguardian/blob/main/docs/configuration-parameters.md)** - Complete parameter guide with enterprise examples
- **[Deployment Examples](https://github.com/zsoftly/logguardian/blob/main/docs/deployment-examples.md)** - AWS CLI, CloudFormation, and Terraform examples
- **[Full Documentation](https://github.com/zsoftly/logguardian)** - Complete project documentation

**🆘 Support:**
- [Report Issues](https://github.com/zsoftly/logguardian/issues)
- [Discussions](https://github.com/zsoftly/logguardian/discussions)

## License

MIT License - see the [LICENSE](https://github.com/zsoftly/logguardian/blob/main/LICENSE) file for details.

---

**Built by ZSoftly Technologies Inc | Made in Canada**
