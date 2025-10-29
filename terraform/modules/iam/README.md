# IAM Module

Creates IAM roles and policies for Lambda execution with least-privilege permissions.

## Usage
```hcl
module "iam" {
  source = "./modules/iam"

  environment = "prod"
  kms_key_arn = "arn:aws:kms:ca-central-1:123456789012:key/..."
  
  tags = {
    Team = "DevOps"
  }
}
```

## Permissions Granted

- **AWS Config**: Read compliance data
- **CloudWatch Logs**: Create/update log groups, apply retention and encryption
- **KMS**: Encrypt/decrypt with specified key
- **CloudWatch Metrics**: Publish custom metrics (LogGuardian namespace only)

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| environment | Environment name | `string` | n/a | yes |
| kms_key_arn | KMS key ARN | `string` | n/a | yes |
| product_name | Product name | `string` | `"LogGuardian"` | no |
| owner | Owner/Team | `string` | `"DevOps"` | no |
| managed_by | Management tool | `string` | `"Terraform"` | no |
| tags | Additional tags | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| lambda_execution_role_arn | ARN of Lambda execution role |
| lambda_execution_role_name | Name of Lambda execution role |
| lambda_execution_role_id | ID of Lambda execution role |

## Security

- Least-privilege IAM policies
- Scoped permissions (no wildcard resources where possible)
- KMS access limited to specified key only
- CloudWatch Metrics restricted to LogGuardian namespace

## Cost

IAM roles and policies are free.
