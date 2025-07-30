package types

import "time"

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
