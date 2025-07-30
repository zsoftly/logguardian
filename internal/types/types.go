package types

import (
	"encoding/json"
	"time"
)

// ConfigEvent represents AWS Config compliance evaluation event
type ConfigEvent struct {
	ConfigRuleInvokingEvent ConfigRuleInvokingEvent `json:"configRuleInvokingEvent"`
	ConfigRuleName          string                  `json:"configRuleName"`
	ResultToken             string                  `json:"resultToken"`
	EventLeftScope          bool                    `json:"eventLeftScope"`
	RuleParameters          map[string]string       `json:"ruleParameters"`
	AccountId               string                  `json:"accountId"`
	ConfigRuleArn           string                  `json:"configRuleArn"`
	ExecutionRoleArn        string                  `json:"executionRoleArn"`
}

// ConfigRuleInvokingEvent contains the resource information
type ConfigRuleInvokingEvent struct {
	ConfigurationItem        ConfigurationItem `json:"configurationItem"`
	NotificationCreationTime time.Time         `json:"notificationCreationTime"`
	MessageType              string            `json:"messageType"`
	RecordVersion            string            `json:"recordVersion"`
}

// ConfigurationItem represents a CloudWatch Log Group
type ConfigurationItem struct {
	ConfigurationItemVersion     string                `json:"configurationItemVersion"`
	ConfigurationItemCaptureTime time.Time             `json:"configurationItemCaptureTime"`
	ConfigurationStateId         int64                 `json:"configurationStateId"`
	AwsAccountId                 string                `json:"awsAccountId"`
	ConfigurationItemStatus      string                `json:"configurationItemStatus"`
	ResourceType                 string                `json:"resourceType"`
	ResourceId                   string                `json:"resourceId"`
	ResourceName                 string                `json:"resourceName"`
	AwsRegion                    string                `json:"awsRegion"`
	AvailabilityZone             string                `json:"availabilityZone"`
	ConfigurationStateMd5Hash    string                `json:"configurationStateMd5Hash"`
	ResourceCreationTime         time.Time             `json:"resourceCreationTime"`
	Configuration                LogGroupConfiguration `json:"configuration"`
}

// LogGroupConfiguration represents CloudWatch Log Group configuration
type LogGroupConfiguration struct {
	LogGroupName         string `json:"logGroupName"`
	RetentionInDays      *int32 `json:"retentionInDays"`
	KmsKeyId             string `json:"kmsKeyId"`
	CreationTime         int64  `json:"creationTime"`
	MetricFilterCount    int32  `json:"metricFilterCount"`
	DataProtectionStatus string `json:"dataProtectionStatus"`
	LogGroupClass        string `json:"logGroupClass"`
}

// ComplianceResult represents the result of compliance checking
type ComplianceResult struct {
	LogGroupName      string
	Region            string
	AccountId         string
	MissingEncryption bool
	MissingRetention  bool
	CurrentRetention  *int32
	CurrentKmsKeyId   string
}

// RemediationResult represents the result of applying remediation
type RemediationResult struct {
	LogGroupName      string
	Region            string
	EncryptionApplied bool
	RetentionApplied  bool
	Success           bool
	Error             error
}

// ConfigRuleEvaluationResults represents AWS Config rule evaluation results
type ConfigRuleEvaluationResults struct {
	EvaluationResults []EvaluationResult `json:"evaluationResults"`
	NextToken         string             `json:"nextToken,omitempty"`
}

// EvaluationResult represents a single Config rule evaluation result
type EvaluationResult struct {
	EvaluationResultIdentifier EvaluationResultIdentifier `json:"evaluationResultIdentifier"`
	ComplianceType             string                     `json:"complianceType"`
	ResultRecordedTime         time.Time                  `json:"resultRecordedTime"`
	ConfigRuleInvokedTime      time.Time                  `json:"configRuleInvokedTime"`
	Annotation                 string                     `json:"annotation,omitempty"`
	ResultToken                string                     `json:"resultToken,omitempty"`
}

// EvaluationResultIdentifier identifies a Config evaluation result
type EvaluationResultIdentifier struct {
	EvaluationResultQualifier EvaluationResultQualifier `json:"evaluationResultQualifier"`
	OrderingTimestamp         time.Time                 `json:"orderingTimestamp"`
}

// EvaluationResultQualifier qualifies a Config evaluation result
type EvaluationResultQualifier struct {
	ConfigRuleName string `json:"configRuleName"`
	ResourceType   string `json:"resourceType"`
	ResourceId     string `json:"resourceId"`
	EvaluationMode string `json:"evaluationMode,omitempty"`
}

// BatchComplianceRequest represents a request to process multiple non-compliant resources
type BatchComplianceRequest struct {
	ConfigRuleName      string                 `json:"configRuleName"`
	NonCompliantResults []NonCompliantResource `json:"nonCompliantResults"`
	Region              string                 `json:"region"`
	BatchSize           int                    `json:"batchSize"`
}

// NonCompliantResource represents a non-compliant resource from Config
type NonCompliantResource struct {
	ResourceId     string    `json:"resourceId"`
	ResourceType   string    `json:"resourceType"`
	ResourceName   string    `json:"resourceName"`
	Region         string    `json:"region"`
	AccountId      string    `json:"accountId"`
	ComplianceType string    `json:"complianceType"`
	Annotation     string    `json:"annotation"`
	LastEvaluated  time.Time `json:"lastEvaluated"`
}

// BatchRemediationResult represents the result of batch remediation
type BatchRemediationResult struct {
	TotalProcessed     int                 `json:"totalProcessed"`
	SuccessCount       int                 `json:"successCount"`
	FailureCount       int                 `json:"failureCount"`
	Results            []RemediationResult `json:"results"`
	ProcessingDuration time.Duration       `json:"processingDuration"`
	RateLimitHits      int                 `json:"rateLimitHits"`
}

// LambdaRequest represents the unified request format for the Lambda
type LambdaRequest struct {
	Type           string          `json:"type"`                     // "config-event" or "config-rule-evaluation"
	ConfigEvent    json.RawMessage `json:"configEvent,omitempty"`    // Contains Config event payload for individual Config events
	ConfigRuleName string          `json:"configRuleName,omitempty"` // For rule evaluation requests
	Region         string          `json:"region,omitempty"`         // For rule evaluation requests
	BatchSize      int             `json:"batchSize,omitempty"`      // For rule evaluation requests
}

// KMSEncryptionResult represents the result of KMS encryption operations
type KMSEncryptionResult struct {
	LogGroupName      string    `json:"logGroupName"`
	KMSKeyAlias       string    `json:"kmsKeyAlias"`
	KMSKeyId          string    `json:"kmsKeyId"`
	KMSKeyArn         string    `json:"kmsKeyArn"`
	KeyRegion         string    `json:"keyRegion"`
	CurrentRegion     string    `json:"currentRegion"`
	IsCrossRegion     bool      `json:"isCrossRegion"`
	EncryptionApplied bool      `json:"encryptionApplied"`
	Success           bool      `json:"success"`
	Error             string    `json:"error,omitempty"`
	ValidationSteps   []string  `json:"validationSteps"`
	AuditTimestamp    time.Time `json:"auditTimestamp"`
}

// KMSValidationReport provides comprehensive KMS key validation information
type KMSValidationReport struct {
	KeyAlias             string    `json:"keyAlias"`
	KeyId                string    `json:"keyId"`
	KeyArn               string    `json:"keyArn"`
	KeyState             string    `json:"keyState"`
	KeyRegion            string    `json:"keyRegion"`
	CurrentRegion        string    `json:"currentRegion"`
	IsCrossRegion        bool      `json:"isCrossRegion"`
	KeyExists            bool      `json:"keyExists"`
	KeyAccessible        bool      `json:"keyAccessible"`
	PolicyAccessible     bool      `json:"policyAccessible"`
	CloudWatchLogsAccess bool      `json:"cloudWatchLogsAccess"`
	ValidationErrors     []string  `json:"validationErrors,omitempty"`
	ValidationWarnings   []string  `json:"validationWarnings,omitempty"`
	RecommendedActions   []string  `json:"recommendedActions,omitempty"`
	ValidationTimestamp  time.Time `json:"validationTimestamp"`
}
