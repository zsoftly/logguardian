# LogGuardian Dev Environment

See the main [Terraform README](../../README.md) for complete documentation.

## Quick Start

```bash
# Authenticate to AWS
ztictl auth login zsoftly
# Select: zsoftly dev logguardian
# Select: AdministratorAccess

export AWS_PROFILE=zsoftly
export AWS_REGION=ca-central-1

# Deploy
terraform init
terraform plan
terraform apply
```

## Environment-Specific Configuration

This environment uses the following configuration (see `env.auto.tfvars`):

- **Region:** ca-central-1
- **Schedule:** Weekly (Sunday at 3-4 AM UTC)
- **Network:** Default VPC with public subnets
- **Dry Run:** Enabled (set to `false` for production)

For detailed documentation on configuration options, deployment steps, and troubleshooting, refer to [../../README.md](../../README.md).
