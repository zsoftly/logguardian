# Config Rule Evaluation Processing

**Why:** Process only non-compliant resources from AWS Config rules instead of scanning all log groups, achieving 90%+ cost reduction.

## Request Format

### Batch Processing
```json
{
  "type": "config-rule-evaluation",
  "configRuleName": "cloudwatch-log-group-encrypted",
  "region": "us-east-1",
  "batchSize": 20
}
```

### Individual Events
```json
{
  "type": "config-event",
  "configEvent": { /* standard Config event */ }
}
```

## Benefits

- **Cost Efficiency**: Process ~5-10% of resources vs 100% scanning
- **Performance**: Batch processing with rate limiting prevents API throttling  
- **Reliability**: Trusts AWS Config's recent evaluation - deleted resources fail gracefully
- **Scalability**: Handles pagination for large Config rule result sets

## Design Decision: No Resource Validation

We trust AWS Config rule evaluations and skip resource existence validation because:

1. **Config Rules are authoritative** - Resources come from recent AWS Config evaluations
2. **Auto-cleanup** - Deleted resources won't appear in subsequent Config rule runs  
3. **Graceful failure** - Remediation APIs handle non-existent resources safely
4. **Performance** - Eliminates unnecessary DescribeLogGroups API calls
