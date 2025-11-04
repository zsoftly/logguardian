// Package testkit provides helpers for integration tests using either
// LocalStack (for local AWS emulation) or real AWS.
//
// ENV BEHAVIOR (loaded externally via godotenv in tests):
//   - LOCALSTACK_ENDPOINT: when set (e.g., http://localhost:4566), routes SDK calls to LocalStack.
//   - AWS_REGION: AWS region (default: ca-central-1).
//
// LOCAL vs REAL AWS:
//   - If LOCALSTACK_ENDPOINT is set → uses LocalStack (no real AWS calls).
//   - If unset → uses real AWS (requires credentials).
package testkit

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// NewAWSConfigLocal returns an AWS config that automatically
// targets LocalStack if LOCALSTACK_ENDPOINT is defined.
func NewAWSConfigLocal(ctx context.Context) (aws.Config, error) {
	region := getenvOr("AWS_REGION", "ca-central-1")
	endpoint := os.Getenv("LOCALSTACK_ENDPOINT")

	// If no endpoint → connect to real AWS
	if endpoint == "" {
		return config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}

	// Otherwise → override S3 endpoint to point to LocalStack
	return config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, r string, _ ...interface{}) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:               endpoint,
						HostnameImmutable: true,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		),
	)
}

// NewS3Local returns an S3 client using the provided config.
// Automatically connects to LocalStack or AWS based on config.
func NewS3Local(cfg aws.Config) *s3.Client {
	return s3.NewFromConfig(cfg)
}

// EnsureS3Bucket creates a bucket if it doesn't exist.
// Safe to call repeatedly in LocalStack; in AWS, bucket must exist beforehand.
func EnsureS3Bucket(ctx context.Context, s3c *s3.Client, bucket, region string) error {
	input := &s3.CreateBucketInput{Bucket: &bucket}
	if region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: regionToConstraint(region),
		}
	}
	_, err := s3c.CreateBucket(ctx, input)
	return err
}

// regionToConstraint maps region strings to typed S3 constraints.
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

// getenvOr fetches an env var or returns a fallback.
func getenvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

/*
OPTIONAL:
Uncomment if you prefer automatic .env loading within this package.

import "github.com/joho/godotenv"
func init() { _ = godotenv.Load(".env") }
*/
