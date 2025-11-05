package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"unit-test/mocks"
)

// Deps mirrors your handler’s external dependencies (S3, config, etc.)
// In production, these come from main.go — for tests, we inject mocks.
type Deps struct {
	S3     *mocks.MockS3PutObject
	Bucket string
}

// Sample Lambda request input
type Request struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

// Mocked process logic for testing (replace with actual handler logic if available)
func Process(ctx context.Context, d Deps, r Request) (string, error) {
	if r.ID == "" {
		return "", errors.New("missing id")
	}
	if _, err := d.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &d.Bucket,
		Key:    &r.ID,
		Body:   strings.NewReader(r.Data),
	}); err != nil {
		return "", err
	}
	return "ok", nil
}

func TestLambdaProcess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name      string
		req       Request
		mockErr   error
		wantErr   bool
	}{
		{"✅ success", Request{ID: "123", Data: "hello"}, nil, false},
		{"❌ missing id", Request{Data: "no-id"}, nil, true},
		{"❌ s3 failure", Request{ID: "x99", Data: "error"}, errors.New("boom"), true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ms3 := mocks.NewMockS3PutObject(ctrl)

			if tc.req.ID != "" {
				ms3.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Return(nil, tc.mockErr)
			}

			deps := Deps{S3: ms3, Bucket: "test-bucket"}

			got, err := Process(ctx, deps, tc.req)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, "ok", got)
			}
		})
	}
}

// Benchmark for performance profiling
func BenchmarkLambdaProcess(b *testing.B) {
	ctx := context.Background()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	ms3 := mocks.NewMockS3PutObject(ctrl)
	ms3.EXPECT().PutObject(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)

	deps := Deps{S3: ms3, Bucket: "bench-bucket"}
	req := Request{ID: "bench-001", Data: strings.Repeat("x", 2<<20)} // ~2MB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Process(ctx, deps, req)
	}
}
