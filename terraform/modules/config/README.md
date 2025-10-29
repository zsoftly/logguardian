# Config Module

Sets up AWS Config service and compliance rules for CloudWatch log group monitoring.

## Usage

### Create Everything New
```hcl
module "config" {
  source = "./modules/config"

  environment            = "prod"
  create_config_service  = true
  config_bucket_name     = "my-config-bucket"
  create_encryption_rule = true
  create_retention_rule  = true
  default_retention_days = 90
}
```

### Use Existing Config, Create Rules
```hcl
module "config" {
  source = "./modules/config"

  environment            = "prod"
  create_config_service  = false
  create_encryption_rule = true
  create_retention_rule  = true
  default_retention_days = 90
}
```

### Use Everything Existing
```hcl
module "config" {
  source = "./modules/config"

  environment                = "prod"
  create_config_service      = false
  create_encryption_rule     = false
  existing_encryption_rule   = "my-encryption-rule"
  create_retention_rule      = false
  existing_retention_rule    = "my-retention-rule"
}
```

## Features

- ✅ Optional Config service creation (most enterprises have this already)
- ✅ Independent control of encryption and retention rules
- ✅ Uses AWS managed Config rules (no Lambda needed)
- ✅ Scoped to CloudWatch log groups only
- ✅ Configurable retention threshold
- ✅ 24-hour evaluation frequency

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| environment | Environment name | `string` | n/a | yes |
| create_config_service | Create Config service | `bool` | `false` | no |
| config_bucket_name | S3 bucket for Config | `string` | `null` | no |
| existing_config_role_arn | Existing Config role ARN | `string` | `null` | no |
| create_encryption_rule | Create encryption rule | `bool` | `true` | no |
| existing_encryption_rule | Existing encryption rule name | `string` | `null` | no |
| create_retention_rule | Create retention rule | `bool` | `true` | no |
| existing_retention_rule | Existing retention rule name | `string` | `null` | no |
| default_retention_days | Min retention days | `number` | `1` | no |
| product_name | Product name | `string` | `"LogGuardian"` | no |
| tags | Additional tags | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| encryption_config_rule_name | Encryption rule name |
| retention_config_rule_name | Retention rule name |
| encryption_config_rule_arn | Encryption rule ARN |
| retention_config_rule_arn | Retention rule ARN |
| config_recorder_name | Config recorder name |
| config_service_created | Whether Config was created |

## Config Rules Used

### Encryption Rule
- **AWS Managed**: `CLOUDWATCH_LOG_GROUP_ENCRYPTED`
- **Checks**: Log groups have KMS encryption enabled
- **Frequency**: Every 24 hours

### Retention Rule
- **AWS Managed**: `CW_LOGGROUP_RETENTION_PERIOD_CHECK`
- **Checks**: Log groups have retention >= specified days
- **Frequency**: Every 24 hours
- **Configurable**: `default_retention_days` parameter

## Cost

- **Config Recorder**: ~$2.00/month
- **Config Rules**: ~$2.00/month per rule
- **Total**: ~$4-6/month if creating Config service + rules
- **Free**: If using existing Config infrastructure

## Notes

- Most enterprise AWS accounts already have Config enabled
- Set `create_config_service=false` to use existing Config
- Rules can be created/used independently
- Recorder only tracks CloudWatch log groups (cost optimized)
