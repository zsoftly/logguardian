# LogGuardian Config Rule Evaluation Examples

This file contains examples of how to invoke the LogGuardian Lambda function to consume AWS Config rule evaluation results.

## Example 1: Process All Non-Compliant Resources from a Config Rule

```json
{
  "type": "config-rule-evaluation",
  "configRuleName": "cloudwatch-log-group-encrypted",
  "region": "ca-central-1",
  "batchSize": 20
}
```

**Expected Behavior:**
1. Lambda retrieves all NON_COMPLIANT resources from the Config rule
2. Validates that log groups still exist
3. Processes resources in batches of 20
4. Applies encryption and retention policies as needed
5. Returns detailed results including success/failure counts

## Example 2: Process Individual Config Event (Original Mode)

```json
{
  "type": "config-event",
  "configEvent": {
    "configRuleName": "cloudwatch-log-group-encrypted",
    "accountId": "123456789012",
    "configRuleInvokingEvent": {
      "configurationItem": {
        "resourceType": "AWS::Logs::LogGroup",
        "resourceName": "/aws/lambda/my-function",
        "awsRegion": "ca-central-1",
        "awsAccountId": "123456789012",
        "configurationItemStatus": "ResourceDiscovered",
        "configuration": {
          "logGroupName": "/aws/lambda/my-function",
          "retentionInDays": null,
          "kmsKeyId": ""
        }
      }
    }
  }
}
```

## Example 3: Backward Compatibility (Legacy Format)

```json
{
  "configRuleName": "cloudwatch-log-group-encrypted",
  "accountId": "123456789012",
  "configRuleInvokingEvent": {
    "configurationItem": {
      "resourceType": "AWS::Logs::LogGroup",
      "resourceName": "/aws/lambda/legacy-function",
      "awsRegion": "ca-central-1",
      "awsAccountId": "123456789012",
      "configurationItemStatus": "ResourceDiscovered",
      "configuration": {
        "logGroupName": "/aws/lambda/legacy-function",
        "retentionInDays": null,
        "kmsKeyId": ""
      }
    }
  }
}
```

## Example 4: AWS CLI Invocation

```bash
# Process all non-compliant resources from a Config rule
aws lambda invoke \
  --function-name logguardian-compliance \
  --payload '{
    "type": "config-rule-evaluation",
    "configRuleName": "cloudwatch-log-group-encrypted",
    "region": "ca-central-1",
    "batchSize": 15
  }' \
  response.json

# Check the response
cat response.json
```

## Example 5: EventBridge Rule for Scheduled Processing

```json
{
  "Rules": [
    {
      "Name": "LogGuardianScheduledCompliance",
      "ScheduleExpression": "cron(0 2 * * ? *)",
      "State": "ENABLED",
      "Targets": [
        {
          "Id": "1",
          "Arn": "arn:aws:lambda:ca-central-1:123456789012:function:logguardian-compliance",
          "Input": "{\"type\":\"config-rule-evaluation\",\"configRuleName\":\"cloudwatch-log-group-encrypted\",\"region\":\"ca-central-1\",\"batchSize\":25}"
        }
      ]
    }
  ]
}
```

## Example 6: Multi-Region Processing

```bash
# Process different regions with separate invocations
for region in ca-central-1 ca-west-1 us-east-2; do
  aws lambda invoke \
    --function-name logguardian-compliance \
    --payload "{
      \"type\": \"config-rule-evaluation\",
      \"configRuleName\": \"cloudwatch-log-group-encrypted\",
      \"region\": \"$region\",
      \"batchSize\": 10
    }" \
    "response-$region.json"
done
```

## Expected Response Format

### Successful Config Rule Evaluation Response

```json
{
  "statusCode": 200,
  "body": {
    "message": "Config rule evaluation processing completed",
    "configRule": "cloudwatch-log-group-encrypted",
    "region": "ca-central-1",
    "totalProcessed": 45,
    "successCount": 43,
    "failureCount": 2,
    "duration": "2m15s",
    "rateLimitHits": 1
  }
}
```

### Error Response

```json
{
  "errorType": "ValidationError",
  "errorMessage": "configRuleName is required for type 'config-rule-evaluation'"
}
```

## Performance Considerations

### Batch Size Guidelines

- **Small Environments (< 100 log groups)**: Use batch size 5-10
- **Medium Environments (100-1000 log groups)**: Use batch size 10-20  
- **Large Environments (> 1000 log groups)**: Use batch size 20-50

### Cost Optimization Examples

**Before (scanning all log groups):**
- 1,000 log groups × 100% scanned = 1,000 API calls
- Lambda execution time: ~5 minutes
- Cost: ~$0.50 per run

**After (processing only non-compliant):**
- 1,000 log groups × 5% non-compliant = 50 resources processed
- Lambda execution time: ~30 seconds
- Cost: ~$0.05 per run
- **Savings: 90% reduction in cost and execution time**

## Monitoring and Alerting

### CloudWatch Metrics to Monitor

```bash
# Lambda execution metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/Lambda \
  --metric-name Duration \
  --dimensions Name=FunctionName,Value=logguardian-compliance \
  --start-time 2024-01-01T00:00:00Z \
  --end-time 2024-01-02T00:00:00Z \
  --period 3600 \
  --statistics Average,Maximum

# Error rate monitoring  
aws cloudwatch get-metric-statistics \
  --namespace AWS/Lambda \
  --metric-name Errors \
  --dimensions Name=FunctionName,Value=logguardian-compliance \
  --start-time 2024-01-01T00:00:00Z \
  --end-time 2024-01-02T00:00:00Z \
  --period 3600 \
  --statistics Sum
```

### Log Insights Queries

```sql
-- Find all Config rule evaluation processing
fields @timestamp, @message
| filter @message like /Config rule evaluation processing completed/
| sort @timestamp desc
| limit 20

-- Monitor rate limiting issues
fields @timestamp, @message
| filter @message like /Rate limit encountered/
| stats count() by bin(5m)
```
