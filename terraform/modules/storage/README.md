# Storage Module

Creates S3 buckets for AWS Config data storage with security and lifecycle policies.

## Usage

### Create New Bucket
```hcl
module "storage" {
  source = "./modules/storage"

  environment            = "prod"
  create_config_bucket   = true
  enable_lifecycle_rules = true
  s3_expiration_days     = 90
}
```

### Skip Bucket Creation (Use Existing)
```hcl
module "storage" {
  source = "./modules/storage"

  environment          = "prod"
  create_config_bucket = false
}
```

## Features

- ✅ Secure by default (all public access blocked)
- ✅ Server-side encryption (AES256)
- ✅ Versioning support
- ✅ Lifecycle rules for cost optimization
- ✅ Proper bucket policy for AWS Config service
- ✅ Account-specific access controls

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| environment | Environment name | `string` | n/a | yes |
| create_config_bucket | Create Config bucket | `bool` | `false` | no |
| enable_lifecycle_rules | Enable lifecycle rules | `bool` | `true` | no |
| s3_expiration_days | Data expiration days | `number` | `90` | no |
| product_name | Product name | `string` | `"LogGuardian"` | no |
| tags | Additional tags | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| config_bucket_name | Name of Config bucket |
| config_bucket_arn | ARN of Config bucket |
| config_bucket_region | Region of Config bucket |
| config_bucket_created | Whether bucket was created |

## Security

- All public access blocked
- Encryption at rest (AES256)
- Bucket policy restricts access to AWS Config service only
- Account-specific conditions prevent cross-account access

## Cost Optimization

- Lifecycle rules automatically delete old data (configurable)
- Versioning can be disabled if not needed
- Old versions deleted after 7 days

## Cost

- **S3 Storage**: ~$0.023/GB/month (Standard)
- **With 10GB Config data**: ~$0.23/month
- **With lifecycle (90 days)**: ~$0.07/month average
