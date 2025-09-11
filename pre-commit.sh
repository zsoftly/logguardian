#!/bin/bash
# Pre-commit checks for LogGuardian
# Run this before committing to avoid pipeline failures

set -e

echo "🔍 Running pre-commit checks for LogGuardian..."
echo ""

echo "📝 Step 1: Formatting code..."
go fmt ./...
echo "✅ Code formatted"
echo ""

echo "🔒 Step 2: Running security linter..."
# Run without config due to version issues, but enable key security checks
if command -v golangci-lint &> /dev/null; then
    golangci-lint run --no-config --enable gosec,errcheck,govet,staticcheck --timeout=5m ./... || {
        echo "❌ Linter found issues. Please fix them before committing."
        exit 1
    }
    echo "✅ Security checks passed"
else
    echo "⚠️  golangci-lint not installed. Install with:"
    echo "    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi
echo ""

echo "🧪 Step 3: Running tests..."
go test ./... -v || {
    echo "❌ Tests failed. Please fix them before committing."
    exit 1
}
echo "✅ All tests passed"
echo ""

echo "🏁 Step 4: Running race detector..."
go test ./... -race || {
    echo "❌ Race conditions detected. Please fix them before committing."
    exit 1
}
echo "✅ No race conditions detected"
echo ""

echo "🔨 Step 5: Building..."
go build ./... || {
    echo "❌ Build failed. Please fix compilation errors before committing."
    exit 1
}
echo "✅ Build successful"
echo ""

echo "🎉 All pre-commit checks passed!"
echo ""
echo "📌 Remember:"
echo "   - Always use ca-central-1 region for testing"
echo "   - Use crypto/rand for randomness (never math/rand)"
echo "   - Run with --dry-run flag first"
echo ""
echo "Ready to commit! 🚀"