package main

import "os"

// getExecutionMode reports whether we're running in AWS Lambda.
// If forceLambda is true, it returns true (useful for tests).
func getExecutionMode(forceLambda bool) bool {
	if forceLambda {
		return true
	}
	return os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""
}
