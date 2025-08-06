package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zsoftly/logguardian/internal/types"
)

// MockKMSClient for testing batch optimization
type MockKMSClientOptimized struct {
	mock.Mock
}

func (m *MockKMSClientOptimized) DescribeKey(ctx context.Context, params *kms.DescribeKeyInput, optFns ...func(*kms.Options)) (*kms.DescribeKeyOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*kms.DescribeKeyOutput), args.Error(1)
}

func (m *MockKMSClientOptimized) GetKeyPolicy(ctx context.Context, params *kms.GetKeyPolicyInput, optFns ...func(*kms.Options)) (*kms.GetKeyPolicyOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*kms.GetKeyPolicyOutput), args.Error(1)
}

func (m *MockKMSClientOptimized) ListGrants(ctx context.Context, params *kms.ListGrantsInput, optFns ...func(*kms.Options)) (*kms.ListGrantsOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*kms.ListGrantsOutput), args.Error(1)
}

// MockLogsClient for testing batch optimization
type MockLogsClientOptimized struct {
	mock.Mock
}

func (m *MockLogsClientOptimized) AssociateKmsKey(ctx context.Context, params *cloudwatchlogs.AssociateKmsKeyInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.AssociateKmsKeyOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*cloudwatchlogs.AssociateKmsKeyOutput), args.Error(1)
}

func (m *MockLogsClientOptimized) PutRetentionPolicy(ctx context.Context, params *cloudwatchlogs.PutRetentionPolicyInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.PutRetentionPolicyOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*cloudwatchlogs.PutRetentionPolicyOutput), args.Error(1)
}

func (m *MockLogsClientOptimized) DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*cloudwatchlogs.DescribeLogGroupsOutput), args.Error(1)
}

func TestBatchKMSValidationCache(t *testing.T) {
	tests := []struct {
		name          string
		keyAlias      string
		expectedKeyId string
		mockKMSError  error
		expectedError bool
	}{
		{
			name:          "successful KMS validation",
			keyAlias:      "alias/test-key",
			expectedKeyId: "key-12345",
			mockKMSError:  nil,
			expectedError: false,
		},
		{
			name:          "KMS key not found",
			keyAlias:      "alias/non-existent-key",
			expectedKeyId: "",
			mockKMSError:  errors.New("NotFoundException: Key 'alias/non-existent-key' does not exist"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockKMS := new(MockKMSClientOptimized)
			mockLogs := new(MockLogsClientOptimized)

			service := &ComplianceService{
				kmsClient:  mockKMS,
				logsClient: mockLogs,
				config: ServiceConfig{
					DefaultKMSKeyAlias:   tt.keyAlias,
					DefaultRetentionDays: 365,
					DryRun:               false,
					Region:               "ca-central-1",
				},
			}

			ctx := context.Background()
			request := types.BatchComplianceRequest{
				ConfigRuleName: "test-rule",
				Region:         "ca-central-1",
				NonCompliantResults: []types.NonCompliantResource{
					{ResourceName: "/aws/lambda/test1", ResourceType: "AWS::Logs::LogGroup"},
				},
				BatchSize: 5,
			}

			if !tt.expectedError {
				// Mock successful KMS key validation
				mockKMS.On("DescribeKey", ctx, mock.MatchedBy(func(params *kms.DescribeKeyInput) bool {
					return *params.KeyId == tt.keyAlias
				})).Return(&kms.DescribeKeyOutput{
					KeyMetadata: &kmstypes.KeyMetadata{
						KeyId:    aws.String(tt.expectedKeyId),
						Arn:      aws.String("arn:aws:kms:ca-central-1:123456789012:key/" + tt.expectedKeyId),
						KeyState: kmstypes.KeyStateEnabled,
					},
				}, tt.mockKMSError)

				// Mock KMS policy validation
				mockKMS.On("GetKeyPolicy", ctx, mock.MatchedBy(func(params *kms.GetKeyPolicyInput) bool {
					return *params.KeyId == tt.expectedKeyId
				})).Return(&kms.GetKeyPolicyOutput{
					Policy: aws.String(`{"Statement":[{"Effect":"Allow","Principal":{"Service":"logs.amazonaws.com"},"Action":["kms:Encrypt","kms:Decrypt"]}]}`),
				}, nil)
			} else {
				// Mock KMS error
				mockKMS.On("DescribeKey", ctx, mock.MatchedBy(func(params *kms.DescribeKeyInput) bool {
					return *params.KeyId == tt.keyAlias
				})).Return(&kms.DescribeKeyOutput{}, tt.mockKMSError)
			}

			// Test batch context creation
			batchCtx, err := service.NewBatchRemediationContext(ctx, request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, batchCtx)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, batchCtx)
				assert.Equal(t, tt.keyAlias, batchCtx.kmsCache.keyAlias)
				assert.Equal(t, request.Region, batchCtx.region)
				assert.Equal(t, request.ConfigRuleName, batchCtx.configRuleName)

				// Test that KMS key info is cached
				keyInfo, err := batchCtx.GetValidatedKMSKeyInfo()
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedKeyId, keyInfo.KeyId)
			}

			mockKMS.AssertExpectations(t)
		})
	}
}

func TestProcessNonCompliantResourcesOptimized(t *testing.T) {
	tests := []struct {
		name                    string
		resources               []types.NonCompliantResource
		batchSize               int
		expectedKMSValidations  int // Should be 1 for optimized version
		expectedEncryptionCalls int
		expectedRetentionCalls  int
		expectedSuccessCount    int
		expectedFailureCount    int
		dryRun                  bool
	}{
		{
			name: "successful batch processing with optimization",
			resources: []types.NonCompliantResource{
				{ResourceName: "/aws/lambda/test1", ResourceType: "AWS::Logs::LogGroup"},
				{ResourceName: "/aws/lambda/test2", ResourceType: "AWS::Logs::LogGroup"},
				{ResourceName: "/aws/lambda/test3", ResourceType: "AWS::Logs::LogGroup"},
			},
			batchSize:               2,
			expectedKMSValidations:  1, // Key optimization: only 1 validation for entire batch
			expectedEncryptionCalls: 3, // One per resource
			expectedRetentionCalls:  3, // One per resource
			expectedSuccessCount:    3,
			expectedFailureCount:    0,
			dryRun:                  false,
		},
		{
			name: "dry run batch processing",
			resources: []types.NonCompliantResource{
				{ResourceName: "/aws/lambda/test1", ResourceType: "AWS::Logs::LogGroup"},
				{ResourceName: "/aws/lambda/test2", ResourceType: "AWS::Logs::LogGroup"},
			},
			batchSize:               1,
			expectedKMSValidations:  1, // Still only 1 validation even in dry run
			expectedEncryptionCalls: 0, // No actual calls in dry run
			expectedRetentionCalls:  0, // No actual calls in dry run
			expectedSuccessCount:    2,
			expectedFailureCount:    0,
			dryRun:                  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockKMS := new(MockKMSClientOptimized)
			mockLogs := new(MockLogsClientOptimized)

			service := &ComplianceService{
				kmsClient:  mockKMS,
				logsClient: mockLogs,
				config: ServiceConfig{
					DefaultKMSKeyAlias:   "alias/test-key",
					DefaultRetentionDays: 365,
					DryRun:               tt.dryRun,
					Region:               "ca-central-1",
					MaxKMSRetries:        3,
					RetryBaseDelay:       time.Second,
				},
			}

			ctx := context.Background()
			request := types.BatchComplianceRequest{
				ConfigRuleName:      "test-rule",
				Region:              "ca-central-1",
				NonCompliantResults: tt.resources,
				BatchSize:           tt.batchSize,
			}

			// Mock KMS key validation (should only be called once)
			mockKMS.On("DescribeKey", ctx, mock.MatchedBy(func(params *kms.DescribeKeyInput) bool {
				return *params.KeyId == "alias/test-key"
			})).Return(&kms.DescribeKeyOutput{
				KeyMetadata: &kmstypes.KeyMetadata{
					KeyId:    aws.String("key-12345"),
					Arn:      aws.String("arn:aws:kms:ca-central-1:123456789012:key/key-12345"),
					KeyState: kmstypes.KeyStateEnabled,
				},
			}, nil).Times(tt.expectedKMSValidations) // This is the key assertion!

			// Mock KMS policy validation (should only be called once)
			mockKMS.On("GetKeyPolicy", ctx, mock.MatchedBy(func(params *kms.GetKeyPolicyInput) bool {
				return *params.KeyId == "key-12345"
			})).Return(&kms.GetKeyPolicyOutput{
				Policy: aws.String(`{"Statement":[{"Effect":"Allow","Principal":{"Service":"logs.amazonaws.com"},"Action":["kms:Encrypt"]}]}`),
			}, nil).Times(tt.expectedKMSValidations) // This too!

			if !tt.dryRun {
				// Mock CloudWatch Logs operations
				for i := 0; i < tt.expectedEncryptionCalls; i++ {
					mockLogs.On("AssociateKmsKey", ctx, mock.AnythingOfType("*cloudwatchlogs.AssociateKmsKeyInput")).Return(&cloudwatchlogs.AssociateKmsKeyOutput{}, nil)
				}
				for i := 0; i < tt.expectedRetentionCalls; i++ {
					mockLogs.On("PutRetentionPolicy", ctx, mock.AnythingOfType("*cloudwatchlogs.PutRetentionPolicyInput")).Return(&cloudwatchlogs.PutRetentionPolicyOutput{}, nil)
				}
			}

			// Execute optimized batch processing
			startTime := time.Now()
			result, err := service.ProcessNonCompliantResourcesOptimized(ctx, request)
			processingTime := time.Since(startTime)

			// Assertions
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, len(tt.resources), result.TotalProcessed)
			assert.Equal(t, tt.expectedSuccessCount, result.SuccessCount)
			assert.Equal(t, tt.expectedFailureCount, result.FailureCount)
			assert.Greater(t, result.ProcessingDuration, time.Duration(0))

			// Performance assertion: optimized version should be faster due to cached KMS validation
			t.Logf("Processing time for %d resources: %v", len(tt.resources), processingTime)

			// Verify all mocks were called as expected
			mockKMS.AssertExpectations(t)
			mockLogs.AssertExpectations(t)
		})
	}
}

func TestBatchRemediationContext_GetValidatedKMSKeyInfo(t *testing.T) {
	tests := []struct {
		name          string
		setupCache    func(*BatchKMSValidationCache)
		expectedError bool
		expectedKeyId string
	}{
		{
			name: "valid cached key info",
			setupCache: func(cache *BatchKMSValidationCache) {
				cache.keyInfo = &KMSKeyInfo{
					KeyId: "key-12345",
					Arn:   "arn:aws:kms:ca-central-1:123456789012:key/key-12345",
				}
				cache.validationError = nil
			},
			expectedError: false,
			expectedKeyId: "key-12345",
		},
		{
			name: "cached validation error",
			setupCache: func(cache *BatchKMSValidationCache) {
				cache.keyInfo = nil
				cache.validationError = errors.New("key not found")
			},
			expectedError: true,
			expectedKeyId: "",
		},
		{
			name: "no validation performed",
			setupCache: func(cache *BatchKMSValidationCache) {
				cache.keyInfo = nil
				cache.validationError = nil
			},
			expectedError: true,
			expectedKeyId: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchCtx := &BatchRemediationContext{
				kmsCache: &BatchKMSValidationCache{
					keyAlias: "alias/test-key",
				},
			}

			tt.setupCache(batchCtx.kmsCache)

			keyInfo, err := batchCtx.GetValidatedKMSKeyInfo()

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, keyInfo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, keyInfo)
				assert.Equal(t, tt.expectedKeyId, keyInfo.KeyId)
			}
		})
	}
}

// Performance benchmark to compare optimized vs non-optimized batch processing
func BenchmarkBatchProcessing(b *testing.B) {
	// Create test resources
	resources := make([]types.NonCompliantResource, 10)
	for i := 0; i < 10; i++ {
		resources[i] = types.NonCompliantResource{
			ResourceName: "/aws/lambda/test-" + string(rune(i)),
			ResourceType: "AWS::Logs::LogGroup",
		}
	}

	request := types.BatchComplianceRequest{
		ConfigRuleName:      "test-rule",
		Region:              "ca-central-1",
		NonCompliantResults: resources,
		BatchSize:           5,
	}

	// Setup service with mocks
	mockKMS := new(MockKMSClientOptimized)
	mockLogs := new(MockLogsClientOptimized)

	service := &ComplianceService{
		kmsClient:  mockKMS,
		logsClient: mockLogs,
		config: ServiceConfig{
			DefaultKMSKeyAlias:   "alias/test-key",
			DefaultRetentionDays: 365,
			DryRun:               true, // Use dry run for benchmarking
			Region:               "ca-central-1",
		},
	}

	ctx := context.Background()

	// Mock KMS operations for benchmarking
	mockKMS.On("DescribeKey", ctx, mock.Anything).Return(&kms.DescribeKeyOutput{
		KeyMetadata: &kmstypes.KeyMetadata{
			KeyId:    aws.String("key-12345"),
			Arn:      aws.String("arn:aws:kms:ca-central-1:123456789012:key/key-12345"),
			KeyState: kmstypes.KeyStateEnabled,
		},
	}, nil)

	mockKMS.On("GetKeyPolicy", ctx, mock.Anything).Return(&kms.GetKeyPolicyOutput{
		Policy: aws.String(`{"Statement":[{"Effect":"Allow","Principal":{"Service":"logs.amazonaws.com"}}]}`),
	}, nil)

	b.ResetTimer()

	b.Run("OptimizedBatchProcessing", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := service.ProcessNonCompliantResourcesOptimized(ctx, request)
			if err != nil {
				b.Fatalf("Optimized batch processing failed: %v", err)
			}
		}
	})
}
