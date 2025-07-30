package service

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/zsoftly/logguardian/internal/types"
)

// ComplianceServiceInterface defines the interface for compliance operations
type ComplianceServiceInterface interface {
	RemediateLogGroup(ctx context.Context, compliance types.ComplianceResult) (*types.RemediationResult, error)
}

// CloudWatchLogsClientInterface defines the interface for CloudWatch Logs operations
type CloudWatchLogsClientInterface interface {
	AssociateKmsKey(ctx context.Context, params *cloudwatchlogs.AssociateKmsKeyInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.AssociateKmsKeyOutput, error)
	PutRetentionPolicy(ctx context.Context, params *cloudwatchlogs.PutRetentionPolicyInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutRetentionPolicyOutput, error)
	DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
}

// KMSClientInterface defines the interface for KMS operations
type KMSClientInterface interface {
	DescribeKey(ctx context.Context, params *kms.DescribeKeyInput, optFns ...func(*kms.Options)) (*kms.DescribeKeyOutput, error)
}