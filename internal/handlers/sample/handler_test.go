package sample

import (
	"context"
	"errors"
	"testing"

	"unit-test/mocks"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

// Unit tests for the Process function.
// These tests use MOCKED AWS services (not real AWS calls).
// In real usage (production or integration tests), the mock client
// would be replaced by an actual AWS SDK client (e.g., s3.NewFromConfig(cfg)).
func TestProcess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Mock bucket name for local testing.
	// In production, this would be replaced by the actual S3 bucket name,
	// e.g. "myapp-prod-bucket"
	const bucket = "test-bucket"

	tests := []struct {
		name      string
		req       Request
		mockS3Err error
		wantErr   bool
	}{
		// Success case
		{"ok", Request{ID: "123", Data: "hello"}, nil, false},

		// Missing ID should fail validation before S3 is called
		{"missing id", Request{Data: "no-id"}, nil, true},

		// Simulated S3 failure using mock error
		{"s3 failure", Request{ID: "999", Data: "x"}, errors.New("boom"), true},
		
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Initialize the mock controller for this test case
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create a mock S3 client
			ms3 := mocks.NewMockS3PutObject(ctrl)

			// Only expect PutObject when the request is valid (ID present)
			if tc.req.ID != "" {
				ms3.EXPECT().
					PutObject(gomock.Any(), gomock.AssignableToTypeOf(&s3.PutObjectInput{})).
					Return(nil, tc.mockS3Err)
			}

			// Inject mocked dependencies into the function under test
			deps := Deps{
				S3:     ms3,     // In production, replace with real S3 client
				Bucket: bucket,  // In production, replace with actual bucket
			}

			got, err := Process(ctx, deps, tc.req)

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "ok", got)
		})
	}
}
