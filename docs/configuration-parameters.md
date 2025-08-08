# LogGuardian Configuration Parameters

## Quick Reference - Essential Parameters

**Infrastructure Control:**
- `CreateKMSKey` - Create new KMS key (`true`) or use existing (`false`)
- `CreateConfigService` - Set up AWS Config service (`true`) or use existing (`false`)
- `CreateConfigRules` - Create Config rules (`true`) or use existing (`false`)
- `CreateEventBridgeRules` - Schedule automated checks (`true`) or manual only (`false`)

**Key Settings:**
- `DefaultRetentionDays` - Log retention period (e.g., `30`, `365`)
- `LambdaMemorySize` - Lambda memory in MB (`256` small, `512` enterprise)
- `Environment` - Deployment environment (`dev`, `prod`, `staging`)

## Core Parameters

### Environment Configuration
| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `Environment` | String | Deployment environment name | `prod`, `dev`, `staging` |
| `ProductName` | String | Product name for resource tagging | `MyCompany-LogGuardian` |
| `Owner` | String | Owner/Team responsible | `DevOps-Team` |
| `ManagedBy` | String | Management tool used | `SAM`, `CloudFormation`, `Terraform` |

### KMS Configuration
| Parameter | Type | Description | Default | Enterprise Use |
|-----------|------|-------------|---------|----------------|
| `CreateKMSKey` | String | Create new KMS key | `true` | `false` (use existing) |
| `ExistingKMSKeyArn` | String | ARN of existing KMS key | - | `arn:aws:kms:region:account:key/id` |
| `KMSKeyAlias` | String | KMS key alias | - | `alias/company-logs-key` |

### AWS Config Configuration
| Parameter | Type | Description | Default | Enterprise Use |
|-----------|------|-------------|---------|----------------|
| `CreateConfigService` | String | Set up AWS Config service | `true` | `false` (use existing) |
| `ExistingConfigBucket` | String | Existing Config S3 bucket | - | `company-config-bucket` |
| `ExistingConfigServiceRoleArn` | String | Existing Config service role | - | `arn:aws:iam::account:role/ConfigRole` |

### Config Rules Configuration
| Parameter | Type | Description | Default | Enterprise Use |
|-----------|------|-------------|---------|----------------|
| `CreateConfigRules` | String | Create Config rules | `true` | `false` (use existing) |
| `ExistingEncryptionConfigRule` | String | Existing encryption rule | - | `company-encryption-rule` |
| `ExistingRetentionConfigRule` | String | Existing retention rule | - | `company-retention-rule` |

### EventBridge Scheduling
| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `CreateEventBridgeRules` | String | Schedule automated checks | `true`, `false` |
| `EncryptionScheduleExpression` | String | Encryption check schedule | `cron(0 2 ? * * *)` (daily 2 AM) |
| `RetentionScheduleExpression` | String | Retention check schedule | `cron(0 3 ? * * *)` (daily 3 AM) |

### Lambda Configuration
| Parameter | Type | Range | Description | Recommendation |
|-----------|------|-------|-------------|----------------|
| `DefaultRetentionDays` | Number | 1-3653 | Default log retention period | `30` (dev), `365` (prod) |
| `LambdaMemorySize` | Number | 128-3008 | Lambda memory allocation (MB) | `256` (small), `512` (enterprise) |
| `LambdaTimeout` | Number | 1-900 | Lambda timeout (seconds) | `300` (typical), `900` (large accounts) |

### S3 Lifecycle Configuration
| Parameter | Type | Range | Description |
|-----------|------|-------|-------------|
| `S3ExpirationDays` | Number | 1-3653 | Config data retention period |
| `EnableS3LifecycleRules` | String | - | Enable S3 lifecycle for cost optimization |

### Monitoring Configuration
| Parameter | Type | Description |
|-----------|------|-------------|
| `CreateMonitoringDashboard` | String | Create CloudWatch dashboard |

## Parameter Examples by Use Case

### Development Environment
```yaml
Environment: dev
DefaultRetentionDays: 7
LambdaMemorySize: 256
CreateMonitoringDashboard: false
CreateKMSKey: true
CreateConfigService: true
```

### Production Environment
```yaml
Environment: prod
DefaultRetentionDays: 365
LambdaMemorySize: 512
CreateMonitoringDashboard: true
CreateKMSKey: true
CreateConfigService: true
S3ExpirationDays: 2555  # 7 years
```

### Enterprise with Existing Infrastructure
```yaml
Environment: prod
CreateKMSKey: false
ExistingKMSKeyArn: arn:aws:kms:ca-central-1:123456789012:key/abcd1234-5678-90ef-ghij-klmnopqrstuv
CreateConfigService: false
ExistingConfigBucket: enterprise-config-bucket-ca-central-1
ExistingConfigServiceRoleArn: arn:aws:iam::123456789012:role/EnterpriseConfigRole
CreateConfigRules: false
ExistingEncryptionConfigRule: enterprise-log-encryption-rule
ExistingRetentionConfigRule: enterprise-log-retention-rule
ProductName: MyCompany-LogGuardian
Owner: Platform-Engineering
CustomerTagPrefix: MYCO
```

### Manual Invocation Only
```yaml
Environment: compliance-audit
CreateEventBridgeRules: false
DefaultRetentionDays: 90
LambdaMemorySize: 512
LambdaTimeout: 900
```

## Schedule Expression Examples

### EventBridge Cron Format
EventBridge uses 6-field cron expressions: `minute hour day-of-month month day-of-week year`

#### Common Schedules
| Schedule | Expression | Description |
|----------|------------|-------------|
| Daily at 2 AM | `cron(0 2 ? * * *)` | Every day at 2:00 AM |
| Daily at 3 AM | `cron(0 3 ? * * *)` | Every day at 3:00 AM |
| Weekly Sunday 2 AM | `cron(0 2 ? * SUN *)` | Every Sunday at 2:00 AM |
| Weekly Monday 2 AM | `cron(0 2 ? * MON *)` | Every Monday at 2:00 AM |
| Bi-weekly | `cron(0 2 ? * SUN#2,SUN#4 *)` | 2nd and 4th Sunday |
| Monthly 1st | `cron(0 2 1 * ? *)` | 1st day of every month |
| Quarterly | `cron(0 2 1 1,4,7,10 ? *)` | 1st day of quarters |

#### Rate Expressions
| Schedule | Expression | Description |
|----------|------------|-------------|
| Every hour | `rate(1 hour)` | Every 60 minutes |
| Every 6 hours | `rate(6 hours)` | Four times per day |
| Every day | `rate(1 day)` | Once per day |
| Every week | `rate(7 days)` | Once per week |

## Advanced Configuration

### Staggered Scheduling Example
To avoid AWS API throttling with large numbers of log groups:

```yaml
# Check encryption at 2 AM
EncryptionScheduleExpression: "cron(0 2 ? * * *)"

# Check retention at 3 AM (1 hour later)
RetentionScheduleExpression: "cron(0 3 ? * * *)"
```

### Custom Tagging Strategy
```yaml
CustomerTagPrefix: "ACME"
ProductName: "LogCompliance"
Owner: "CloudOps"
ManagedBy: "Terraform"

# Results in tags like:
# Product: ACME-LogCompliance
# Owner: CloudOps
# Environment: prod
# ManagedBy: Terraform
```
