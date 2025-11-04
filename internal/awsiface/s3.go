// internal/awsiface/s3.go
//
// PURPOSE:
// Defines a lightweight interface for the S3 operations used by our Lambda functions.
//
// WHY:
// - Allows unit tests to mock AWS S3 without making real network calls.
// - Keeps tests isolated, fast, and deterministic.
//
// CURRENT USAGE:
// - Process handler (handler.go) uses only PutObject for uploads.
// - Integration tests (handler_integration_test.go) perform Put + Get
//   using the real S3 client (LocalStack or AWS).
//
// FUTURE USAGE:
// - GetObject is included to support future Lambdas that download or
//   read from S3. (e.g., "FetchReportLambda" or "ReadFileHandler").
// - If you add new S3 operations (e.g., DeleteObject, ListObjects),
//   extend this interface and regenerate mocks.
//
// REAL VALUES:
// - Real region, bucket, and endpoint are never hardcoded here.
// - They’re supplied at runtime through environment variables or dependency wiring.
//
// MOCKS:
// - Unit tests use a generated mock (mocks/mock_s3.go).
// - Run `mingw32-make mocks` anytime this interface changes.

package awsiface

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3PutObject defines the minimal methods our Lambda handlers might call.
// In production: satisfied by *s3.Client (real AWS SDK client).
// In unit tests: satisfied by mocks.NewMockS3PutObject.
type S3PutObject interface {
	// Used by the current Process Lambda for uploads.
	PutObject(ctx context.Context, in *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)

	// Reserved for future Lambdas that need to read or verify stored data.
	GetObject(ctx context.Context, in *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// Compile-time check: ensures the real *s3.Client implements this interface.
// (Prevents mismatch errors if AWS SDK updates method signatures.)
var _ S3PutObject = (*s3.Client)(nil)
