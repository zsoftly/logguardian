package testkit

import "testing"

// MustEq is a lightweight helper for quick equality assertions.
//
// Usage in small tests or benchmarks (without testify):
//   testkit.MustEq(t, gotValue, expectedValue)
//
// In larger tests, you can still use testify's require/assert if preferred.
// This helper just provides a minimal built-in option when you don’t want extra dependencies.
func MustEq[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v want %v", got, want)
	}
}
