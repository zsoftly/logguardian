# EventBridge Module

Creates EventBridge (CloudWatch Events) rules to automatically trigger Lambda on a schedule.

## Usage

### Enable Scheduled Automation
```hcl
module "eventbridge" {
  source = "./modules/eventbridge"

  environment            = "prod"
  lambda_function_arn    = "arn:aws:lambda:ca-central-1:123456789012:function/logguardian-compliance-prod"
  lambda_function_name   = "logguardian-compliance-prod"
  encryption_config_rule = "logguardian-encryption-prod"
  retention_config_rule  = "logguardian-retention-prod"
  
  # Weekly on Sunday
  encryption_schedule = "cron(0 3 ? * SUN *)"
  retention_schedule  = "cron(0 4 ? * SUN *)"
}
```

### Disable Automation (Manual Invocation Only)
```hcl
module "eventbridge" {
  source = "./modules/eventbridge"

  environment              = "prod"
  create_eventbridge_rules = false
  lambda_function_arn      = "..."
  lambda_function_name     = "..."
  encryption_config_rule   = "..."
  retention_config_rule    = "..."
}
```

## Schedule Expressions

### Cron Format
```
cron(Minutes Hours Day-of-month Month Day-of-week Year)
```

**Examples:**
```hcl
# Daily at 2 AM UTC
encryption_schedule = "cron(0 2 * * ? *)"

# Weekly on Sunday at 3 AM UTC
encryption_schedule = "cron(0 3 ? * SUN *)"

# Weekdays at 6 AM UTC
encryption_schedule = "cron(0 6 ? * MON-FRI *)"

# First day of month at midnight
encryption_schedule = "cron(0 0 1 * ? *)"
```

### Rate Format
```
rate(value unit)
```

**Examples:**
```hcl
# Every hour
encryption_schedule = "rate(1 hour)"

# Every 12 hours
encryption_schedule = "rate(12 hours)"

# Every day
encryption_schedule = "rate(1 day)"
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| environment | Environment name | `string` | n/a | yes |
| create_eventbridge_rules | Create rules | `bool` | `true` | no |
| lambda_function_arn | Lambda ARN | `string` | n/a | yes |
| lambda_function_name | Lambda name | `string` | n/a | yes |
| encryption_config_rule | Encryption rule name | `string` | n/a | yes |
| retention_config_rule | Retention rule name | `string` | n/a | yes |
| encryption_schedule | Encryption check schedule | `string` | `"cron(0 3 ? * SUN *)"` | no |
| retention_schedule | Retention check schedule | `string` | `"cron(0 4 ? * SUN *)"` | no |
| tags | Additional tags | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| encryption_rule_arn | Encryption EventBridge rule ARN |
| retention_rule_arn | Retention EventBridge rule ARN |
| eventbridge_rules_created | Whether rules were created |

## Default Schedule

**Encryption checks:** Sunday at 3 AM UTC (Saturday 10 PM EST)  
**Retention checks:** Sunday at 4 AM UTC (Saturday 11 PM EST)

Staggered by 1 hour to avoid overwhelming AWS APIs.

## Recommended Schedules

### Production
```hcl
# Weekly during off-hours
encryption_schedule = "cron(0 3 ? * SUN *)"  # Sunday 3 AM UTC
retention_schedule  = "cron(0 4 ? * SUN *)"  # Sunday 4 AM UTC
```

### Development/Testing
```hcl
# Daily for faster feedback
encryption_schedule = "rate(1 day)"
retention_schedule  = "rate(1 day)"
```

### High-Compliance Environments
```hcl
# Every 12 hours
encryption_schedule = "rate(12 hours)"
retention_schedule  = "rate(12 hours)"
```

## Cost

EventBridge rules are free for:
- First 1M invocations/month (Free Tier)

After Free Tier:
- $1.00 per million custom events

**Weekly schedule:** ~8 invocations/month = FREE  
**Daily schedule:** ~60 invocations/month = FREE  
**Hourly schedule:** ~1,440 invocations/month = FREE

## Testing
```bash
# Manually trigger encryption rule
aws events put-events \
  --entries '[{
    "Source": "manual.test",
    "DetailType": "Manual Test",
    "Detail": "{\"test\": true}",
    "Resources": ["arn:aws:events:ca-central-1:123456789012:rule/logguardian-encryption-schedule-prod"]
  }]'

# Check rule is enabled
aws events describe-rule \
  --name logguardian-encryption-schedule-prod

# View recent invocations
aws cloudwatch get-metric-statistics \
  --namespace AWS/Events \
  --metric-name TriggeredRules \
  --dimensions Name=RuleName,Value=logguardian-encryption-schedule-prod \
  --start-time $(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 86400 \
  --statistics Sum
```

## Disabling vs Deleting

**Disable (Keep Rules):**
```hcl
# Rules exist but won't trigger Lambda
# Set in EventBridge console or:
aws events disable-rule --name logguardian-encryption-schedule-prod
```

**Delete (Remove Rules):**
```hcl
create_eventbridge_rules = false
# Then: terraform apply
```
