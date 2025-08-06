# LogGuardian - Simplified Makefile for Go Development
# Deployment is handled by SAM (see sam.mk)

# Variables
BINARY_NAME := bootstrap
BUILD_DIR := build
GOOS := linux
GOARCH := amd64

# Default target - essential development workflow
.PHONY: all
all: clean build test

# Clean build artifacts and caches
.PHONY: clean
clean:
	@echo "Cleaning build artifacts and caches..."
	rm -rf $(BUILD_DIR) .aws-sam dist
	go clean -cache -testcache -modcache

# Build Lambda function binary for AWS
.PHONY: build
build:
	@echo "Building Lambda function for AWS..."
	mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build \
		-ldflags="-s -w" \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		./cmd/lambda

# Development build (faster, local testing)
.PHONY: dev-build
dev-build:
	@echo "Building for local development..."
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME)-dev ./cmd/lambda

# Run tests with coverage
.PHONY: test
test:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Format, lint, and validate code
.PHONY: check
check: fmt lint vet security vuln-check

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

.PHONY: security
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

.PHONY: vuln-check
vuln-check:
	@echo "Checking for vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "⚠️  govulncheck not found. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Dependency management
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	go mod verify

.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download

# Install development tools
.PHONY: install-tools
install-tools:
	@echo "Installing Go development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

# Performance profiling
.PHONY: profile-memory
profile-memory:
	@echo "Running memory profiling..."
	go test -memprofile=mem.prof -bench=. ./...
	@echo "Profile saved: mem.prof (view with 'go tool pprof mem.prof')"

.PHONY: profile-cpu
profile-cpu:
	@echo "Running CPU profiling..."
	go test -cpuprofile=cpu.prof -bench=. ./...
	@echo "Profile saved: cpu.prof (view with 'go tool pprof cpu.prof')"

# Show binary size
.PHONY: size
size: build
	@echo "Binary size:"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)

# Include SAM deployment targets
include sam.mk

# Help
.PHONY: help
help:
	@echo "LogGuardian - Go Lambda Development"
	@echo "=================================="
	@echo ""
	@echo "🚀 Primary deployment method: AWS SAM"
	@echo ""
	@echo "Development Targets:"
	@echo "  all              - Clean, build, and test"
	@echo "  build            - Build Lambda binary for AWS"
	@echo "  dev-build        - Build for local development"
	@echo "  test             - Run tests with coverage"
	@echo "  bench            - Run benchmarks"
	@echo "  check            - Format, lint, vet, security scan"
	@echo "  clean            - Clean all build artifacts"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt              - Format code"
	@echo "  lint             - Lint with golangci-lint"
	@echo "  vet              - Run go vet"
	@echo "  security         - Security scan with gosec"
	@echo "  vuln-check       - Check for vulnerabilities"
	@echo ""
	@echo "Utilities:"
	@echo "  tidy             - Tidy dependencies"
	@echo "  deps             - Download dependencies"
	@echo "  install-tools    - Install development tools"
	@echo "  size             - Show binary size"
	@echo "  profile-memory   - Memory profiling"
	@echo "  profile-cpu      - CPU profiling"
	@echo ""
	@echo "📦 SAM Deployment (from sam.mk):"
	@echo "  sam-build        - Build for SAM"
	@echo "  sam-test-quick   - Quick local testing"
	@echo "  sam-deploy-dev   - Deploy to development"
	@echo "  sam-deploy-prod  - Deploy to production"
	@echo ""
	@echo "For complete SAM targets, see: sam.mk"