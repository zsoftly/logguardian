# Monitoring Module

Creates a CloudWatch dashboard for LogGuardian observability with Lambda metrics and compliance tracking.

## Usage

### Enable Dashboard
```hcl
module "monitoring" {
  source = "./modules/monitoring"

  environment          = "prod"
  lambda_function_name = "logguardian-compliance-prod"
  log_group_name       = "/aws/lambda/logguardian-compliance-prod"
  create_dashboard     = true
}
```

### Disable Dashboard
```hcl
module "monitoring" {
  source = "./modules/monitoring"

  environment          = "prod"
  lambda_function_name = "logguardian-compliance-prod"
  log_group_name       = "/aws/lambda/logguardian-compliance-prod"
  create_dashboard     = false
}
```

## Dashboard Widgets

### 1. Lambda Function Metrics
- **Average Duration** - How long Lambda takes to run
- **Max Duration** - Longest execution time
- **Errors** - Failed invocations
- **Invocations** - Total executions
- **Throttles** - Rate limit hits

### 2. Compliance Metrics
- **LogGroupsProcessed** - Total log groups evaluated
- **LogGroupsRemediated** - Log groups fixed
- **RemediationErrors** - Failed remediation attempts

### 3. Recent Errors and Warnings
- Live log query showing recent ERROR and WARN messages
- Last 20 entries
- Auto-refreshes

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| environment | Environment name | `string` | n/a | yes |
| create_dashboard | Create dashboard | `bool` | `true` | no |
| lambda_function_name | Lambda function name | `string` | n/a | yes |
| log_group_name | Lambda log group name | `string` | n/a | yes |
| product_name | Product name | `string` | `"LogGuardian"` | no |
| tags | Additional tags | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| dashboard_name | CloudWatch dashboard name |
| dashboard_arn | Dashboard ARN |
| dashboard_url | Direct URL to dashboard |
| dashboard_created | Whether dashboard was created |

## Accessing the Dashboard

**After deployment:**
```bash
# Get dashboard URL from Terraform output
terraform output dashboard_url

# Or via AWS CLI
aws cloudwatch get-dashboard \
  --dashboard-name LogGuardian-prod \
  --region ca-central-1
```

**AWS Console:**
1. Go to CloudWatch console
2. Click "Dashboards" in left menu
3. Select "LogGuardian-{environment}"

## Custom Metrics

LogGuardian publishes custom metrics to CloudWatch:

**Namespace:** `LogGuardian`

**Metrics:**
- `LogGroupsProcessed` - Count of log groups evaluated
- `LogGroupsRemediated` - Count of log groups fixed
- `RemediationErrors` - Count of remediation failures

**Dimensions:**
- `Environment` - Deployment environment (dev/staging/prod)

## Alarms (Optional Enhancement)

You can add alarms manually or extend this module:
```hcl
# High error rate alarm
resource "aws_cloudwatch_metric_alarm" "high_errors" {
  alarm_name          = "logguardian-high-errors-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 300
  statistic           = "Sum"
  threshold           = 5
  alarm_description   = "Lambda function has high error rate"
  
  dimensions = {
    FunctionName = var.lambda_function_name
  }
}
```

## Cost

CloudWatch dashboards:
- **First 3 dashboards:** FREE
- **Additional dashboards:** $3/month each
- **Custom metrics:** $0.30 per metric per month
- **Metric API calls:** $0.01 per 1,000 GetMetricData requests

**Typical cost:** FREE (within free tier) to $1-2/month

## Dashboard Customization

To customize the dashboard, modify the `dashboard_body` in `main.tf`:

**Add widget:**
```json
{
  "type": "metric",
  "properties": {
    "metrics": [["YourNamespace", "YourMetric"]],
    "period": 300,
    "stat": "Average"
  }
}
```

**Add alarm widget:**
```json
{
  "type": "alarm",
  "properties": {
    "alarms": ["arn:aws:cloudwatch:region:account:alarm:alarm-name"]
  }
}
```

## Troubleshooting

**Issue: No data in Compliance Metrics widget**
- Lambda hasn't run yet (check EventBridge schedule)
- Lambda isn't publishing metrics (check Lambda logs)

**Issue: Dashboard not found**
- Check `create_dashboard = true` in module
- Verify dashboard name matches environment

**Issue: Log query shows no results**
- Lambda hasn't logged any errors/warnings yet (good!)
- Check log group name is correct
