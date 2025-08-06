# Rule Classification System

## Overview

LogGuardian uses a simple and reliable rule classification system to categorize AWS Config rules and apply appropriate compliance remediation. The system uses straightforward substring matching for maximum simplicity and reliability, avoiding the complexity of regex patterns or configurable mappings.

## Architecture

### RuleClassifier

The `RuleClassifier` (`internal/types/rules.go`) is the central component responsible for:

- Classifying Config rule names into specific types (Encryption, Retention, Unknown)
- Using simple, case-insensitive substring matching
- Providing reliable classification without complex configuration

### Rule Types

```go
type RuleType int

const (
    RuleTypeUnknown RuleType = iota
    RuleTypeEncryption
    RuleTypeRetention
)
```

## Classification Logic

### Simple Substring Matching

The classifier uses straightforward substring matching against rule names:

```go
func (rc *RuleClassifier) ClassifyRule(configRuleName string) RuleType {
    if configRuleName == "" {
        return RuleTypeUnknown
    }

    // Normalize for case-insensitive matching
    normalizedName := strings.ToLower(configRuleName)

    // Simple substring matching - rules have descriptive names
    if strings.Contains(normalizedName, "encryption") || strings.Contains(normalizedName, "encrypted") {
        return RuleTypeEncryption
    }

    if strings.Contains(normalizedName, "retention") {
        return RuleTypeRetention
    }

    return RuleTypeUnknown
}
```

### Supported Rule Examples

#### Encryption Rules (contain "encryption" or "encrypted")
- `cloudwatch-log-group-encrypted` - AWS standard encryption rule
- `logguardian-encryption-check` - Custom encryption rule
- `my-org-log-group-encrypted` - Organization-specific rule
- `kms-encryption-rule` - KMS-focused rule
- `CLOUDWATCH-LOG-GROUP-ENCRYPTED` - Case insensitive matching

#### Retention Rules (contain "retention")
- `cw-loggroup-retention-period-check` - AWS standard retention rule
- `logguardian-retention-check` - Custom retention rule
- `audit-retention-period-check` - Audit-focused rule
- `log-retention-policy` - Policy-focused rule

#### Unknown Rules (no matching substrings)
- `s3-backup-policy-check` - Different service
- `vpc-security-group-check` - Different compliance area
- `lambda-environment-check` - Different resource type

## Usage

### Basic Classification

```go
classifier := types.NewRuleClassifier()

// Classify a rule
ruleType := classifier.ClassifyRule("cloudwatch-log-group-encrypted")
// Returns: RuleTypeEncryption

// Check specific types
isEncryption := classifier.IsEncryptionRule("cloudwatch-log-group-encrypted")
// Returns: true

isRetention := classifier.IsRetentionRule("cw-loggroup-retention-period-check")
// Returns: true
```

### String Representation

```go
ruleType := classifier.ClassifyRule("logguardian-encryption-check")
fmt.Println(ruleType.String()) // Output: "encryption"
```

## Integration

### Handler Integration

The `ComplianceHandler` uses rule classification to ensure each Config rule only evaluates its specific requirement:

```go
ruleType := h.ruleClassifier.ClassifyRule(configRuleName)

switch ruleType {
case types.RuleTypeEncryption:
    // Only evaluate encryption compliance
    result.MissingEncryption = config.KmsKeyId == ""
    result.MissingRetention = false
    
case types.RuleTypeRetention:
    // Only evaluate retention compliance
    result.MissingRetention = config.RetentionInDays == nil
    result.MissingEncryption = false
    
default:
    // Unsupported rule - no remediation
    slog.Warn("Unsupported Config rule - no compliance evaluation performed",
        "config_rule", configRuleName,
        "rule_type", "unknown")
    result.MissingEncryption = false
    result.MissingRetention = false
}
```

### Service Integration

The `ComplianceService` uses the same classification for batch processing:

```go
ruleType := s.ruleClassifier.ClassifyRule(configRuleName)

switch ruleType {
case types.RuleTypeEncryption:
    result.MissingEncryption = true  // Non-compliant for encryption
    result.MissingRetention = false  // Not this rule's concern
    
case types.RuleTypeRetention:
    result.MissingRetention = true   // Non-compliant for retention
    result.MissingEncryption = false // Not this rule's concern
}
```

## Benefits

### 1. Maximum Simplicity

- No complex regex patterns to maintain
- No external configuration required
- Straightforward substring matching logic
- Easy to understand and debug

### 2. Reliable Classification

- Case-insensitive matching handles naming variations
- Descriptive rule names make classification obvious
- No false positives from complex pattern matching
- Predictable behavior across different rule naming conventions

### 3. Zero Configuration

- No environment variables to manage
- No CloudFormation parameters for patterns
- Works out-of-the-box with standard naming conventions
- Self-contained logic with no external dependencies

### 4. Performance and Maintainability

- Fast string operations (no regex compilation)
- No external configuration to load or parse
- Simple logic that's easy to test and maintain
- Thread-safe by default

## Testing

The rule classification system includes comprehensive tests covering:

- Standard AWS Config rule names
- Custom organization rule names
- Case sensitivity handling
- Edge cases (empty names, unknown rules)
- Performance benchmarking

### Test Examples

```go
func TestRuleClassifier_ClassifyRule(t *testing.T) {
    tests := []struct {
        name         string
        configRule   string
        expectedType RuleType
    }{
        {
            name:         "AWS standard encryption rule",
            configRule:   "cloudwatch-log-group-encrypted",
            expectedType: RuleTypeEncryption,
        },
        {
            name:         "Case sensitivity test - uppercase",
            configRule:   "CLOUDWATCH-LOG-GROUP-ENCRYPTED",
            expectedType: RuleTypeEncryption,
        },
        {
            name:         "Unrelated backup rule",
            configRule:   "s3-backup-policy-check",
            expectedType: RuleTypeUnknown,
        },
    }
    // ... test implementation
}
```

## AWS Config Compliance Model

This system ensures correct AWS Config compliance by:

1. **Independent Rule Evaluation**: Each rule only evaluates its specific requirement
2. **Complete Resource Coverage**: Every rule evaluates ALL applicable resources
3. **No Resource Skipping**: No coordination logic that causes compliance gaps
4. **Proper Audit Trails**: Clear logging of rule type and evaluation results

## Design Philosophy

The simplified approach prioritizes:

- **Reliability over Flexibility**: Simple substring matching is more reliable than complex patterns
- **Maintainability over Features**: Fewer moving parts mean less maintenance overhead
- **Clarity over Configurability**: Self-documenting code is better than external configuration
- **Performance over Complexity**: String operations are faster than regex matching

This design recognizes that AWS Config rule names are inherently descriptive and self-documenting, making complex classification logic unnecessary.

## Migration from Complex Systems

This system replaces complex pattern-based approaches with simple, reliable logic:

### Previous Complex Approach (Removed)
- Regex patterns requiring maintenance
- Environment variables for pattern configuration
- CloudFormation parameters for rule mapping
- Complex validation and error handling

### Current Simple Approach
```go
// Simple and reliable
normalizedName := strings.ToLower(configRuleName)
if strings.Contains(normalizedName, "encryption") || strings.Contains(normalizedName, "encrypted") {
    return RuleTypeEncryption
}
```

## Performance

- **Fast Operations**: String contains operations are highly optimized
- **No Compilation Overhead**: No regex patterns to compile
- **Thread-Safe**: Simple operations are inherently thread-safe
- **Predictable Performance**: O(n) substring search with small n

The rule classification system provides a simple, reliable, and maintainable foundation for AWS Config compliance evaluation in LogGuardian.