# Config Rule Evaluation: Smart Resource Processing

## Core Value Proposition

**Process only non-compliant resources from AWS Config rules instead of scanning all log groups, achieving 90%+ cost reduction.**

## Why This Approach?

| Traditional Approach | Config Rule Evaluation |
|---------------------|------------------------|
| Scan ALL log groups (~1000s) | Process only NON_COMPLIANT (~50-100) |
| 100% API cost | ~5-10% API cost |
| High latency | Fast, targeted processing |
| Resource-intensive | Memory-efficient |

## Design Decision: Trust AWS Config

We **trust AWS Config evaluations** and skip redundant validation:

- ✅ **Config Rules are authoritative** - Recent evaluations from AWS
- ✅ **Auto-cleanup** - Deleted resources won't appear in next run
- ✅ **Graceful failure** - APIs handle non-existent resources safely
- ✅ **Performance** - No unnecessary `DescribeLogGroups` calls

## Implementation Details

- **Request formats**: See [examples.md](examples.md)
- **Batch optimization**: See [kms-batch-optimization.md](kms-batch-optimization.md) 
- **Rule classification**: See [rule-classification.md](rule-classification.md)
- **Lambda architecture**: See [architecture-overview.md](architecture-overview.md)
