package sample

import (
	"bytes" // build request body as io.Reader
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"unit-test/internal/awsiface"
)

// Deps are injected for testability.
// - UNIT TESTS: S3 is a mock, Bucket is a test value.
// - INTEGRATION/PROD: S3 is a real *s3.Client and Bucket comes from env/wiring (not hardcoded here).
type Deps struct {
	S3     awsiface.S3PutObject
	Bucket string // e.g., set via BUCKET_NAME in wiring code (NewDepsFromEnv)
}

type Request struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

// Process validates input, marshals it, and writes to S3.
// Real AWS credentials/region/bucket are provided by the caller via Deps.
// This function should remain unaware of env vars or endpoints.
func Process(ctx context.Context, d Deps, r Request) (string, error) {
	if r.ID == "" {
		return "", errors.New("missing id")
	}

	b, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	_, err = d.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &d.Bucket,                                   // PROD: real bucket injected in Deps
		Key:    ptr(fmt.Sprintf("items/%s.json", r.ID)),     // consider making the prefix configurable if needed
		Body:   bytes.NewReader(b),
		// Optional (common in prod):
		// ContentType:         ptr("application/json"),
		// ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		// SSEKMSKeyId:          ptr(os.Getenv("KMS_KEY_ID")), // from wiring, not here
	})
	if err != nil {
		return "", fmt.Errorf("s3 put: %w", err)
	}
	return "ok", nil
}

// Tiny generic pointer helper
func ptr[T any](v T) *T { return &v }
