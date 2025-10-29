# KMS Module

Manages KMS encryption keys for CloudWatch Logs. Can create a new key or use an existing one.

## Usage

### Create New Key
```hcl
module "kms" {
  source = "./modules/kms"

  environment    = "prod"
  create_kms_key = true
  kms_key_alias  = "alias/logguardian-prod"
}
```

### Use Existing Key
```hcl
module "kms" {
  source = "./modules/kms"

  environment          = "prod"
  create_kms_key       = false
  existing_kms_key_arn = "arn:aws:kms:ca-central-1:123456789012:key/..."
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| environment | Environment name | `string` | n/a | yes |
| create_kms_key | Create new KMS key | `bool` | `true` | no |
| existing_kms_key_arn | Existing KMS key ARN | `string` | `null` | no |
| kms_key_alias | KMS key alias | `string` | `"alias/logguardian-cloudwatch-logs"` | no |
| enable_key_rotation | Enable key rotation | `bool` | `true` | no |
| product_name | Product name | `string` | `"LogGuardian"` | no |
| tags | Additional tags | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| kms_key_arn | ARN of the KMS key |
| kms_key_id | ID of the KMS key |
| kms_key_alias | Alias of the KMS key |

## Security

- Automatic key rotation enabled by default
- 30-day deletion window
- Least-privilege key policy
- CloudWatch Logs service-specific access

## Cost

~$1-2/month per key
