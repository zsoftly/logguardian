//go:build integration

package sample_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"unit-test/internal/testkit"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/joho/godotenv" // Loads .env automatically for flexibility
)

// TestMain runs once before all tests.
// Loads .env so tests can use dynamic values (e.g., bucket, region).
func TestMain(m *testing.M) {
	_ = godotenv.Load(".env") // Non-fatal if .env missing
	os.Exit(m.Run())
}

// Test_LocalStack_S3_RoundTrip verifies S3 upload/download using LocalStack.
// Uses environment variables — no hardcoded AWS details.
// For real AWS tests, switch endpoint in .env to AWS and ensure bucket exists.
func Test_LocalStack_S3_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create AWS SDK config (uses LOCALSTACK_ENDPOINT or real AWS)
	cfg, err := testkit.NewAWSConfigLocal(ctx)
	if err != nil {
		t.Fatal(err)
	}
	s3c := s3.NewFromConfig(cfg)

	// Pull configuration from .env or fall back to defaults
	bucket := getenvOr("BUCKET_NAME", "test-bucket")
	key := getenvOr("OBJECT_KEY", "test-object.txt")
	region := getenvOr("AWS_REGION", "ca-central-1")
	body := []byte(getenvOr("TEST_BODY", "hello, localstack"))

	// Create a temporary bucket (skip this step for real AWS)
	_, _ = s3c.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &bucket,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(regionToConstraint(region)),
		},
	})

	time.Sleep(200 * time.Millisecond) // Give LocalStack time to stabilize

	// Upload the object
	if _, err := s3c.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(body),
	}); err != nil {
		t.Fatalf("put object: %v", err)
	}

	// Retrieve and verify the object content
	out, err := s3c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		t.Fatalf("get object: %v", err)
	}
	got, _ := io.ReadAll(out.Body)
	_ = out.Body.Close()

	if string(got) != string(body) {
		t.Fatalf("body mismatch: got=%q want=%q", got, body)
	}
}

// getenvOr returns environment variable value or fallback if unset.
// Keeps tests flexible across local/dev/staging environments.
func getenvOr(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// regionToConstraint maps region strings to valid S3 bucket constraints.
// Allows easy switching between LocalStack and real AWS regions.
func regionToConstraint(region string) types.BucketLocationConstraint {
	switch region {
	case "us-east-1":
		return ""
	case "ca-central-1":
		return types.BucketLocationConstraintCaCentral1
	default:
		return types.BucketLocationConstraint(region)
	}
}
