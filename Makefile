# LogGuardian Go Lambda Function Makefile

# Variables
FUNCTION_NAME := logguardian-compliance
BINARY_NAME := main
BUILD_DIR := build
DIST_DIR := dist
GO_VERSION := 1.24

# AWS Lambda requires GOOS=linux and GOARCH=amd64 for Go runtime
GOOS := linux
GOARCH := amd64

# Default target
.PHONY: all
all: clean build test

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	go clean -cache
	go clean -testcache

# Build the Lambda function binary
.PHONY: build
build: clean
	@echo "Building Lambda function..."
	mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build \
		-ldflags="-s -w" \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		./cmd/lambda

# Create deployment package
.PHONY: package
package: build
	@echo "Creating deployment package..."
	mkdir -p $(DIST_DIR)
	cd $(BUILD_DIR) && zip -r ../$(DIST_DIR)/$(FUNCTION_NAME).zip $(BINARY_NAME)
	@echo "Deployment package created: $(DIST_DIR)/$(FUNCTION_NAME).zip"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
.PHONY: test-coverage
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

# Run security scanner
.PHONY: security
security:
	@echo "Running security scan..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Run: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Check dependencies for vulnerabilities
.PHONY: vuln-check
vuln-check:
	@echo "Checking for vulnerabilities..."
	@if command -v govulncheck > /dev/null; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Run: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	go mod verify

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download

# Generate mocks (if mockgen is available)
.PHONY: mocks
mocks:
	@echo "Generating mocks..."
	@if command -v mockgen > /dev/null; then \
		go generate ./...; \
	else \
		echo "mockgen not installed. Run: go install github.com/golang/mock/mockgen@latest"; \
	fi

# Run all quality checks
.PHONY: check
check: fmt lint test security vuln-check
	@echo "All quality checks completed"

# Development build (faster, no optimizations)
.PHONY: dev-build
dev-build:
	@echo "Building for development..."
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/lambda

# Local testing with sample event
.PHONY: test-local
test-local: dev-build
	@echo "Running local test..."
	@if [ -f "testdata/sample-event.json" ]; then \
		$(BUILD_DIR)/$(BINARY_NAME) < testdata/sample-event.json; \
	else \
		echo "No sample event file found at testdata/sample-event.json"; \
	fi

# Memory usage analysis
.PHONY: memory-profile
memory-profile:
	@echo "Running memory profile tests..."
	go test -memprofile=mem.prof -bench=. ./...
	@echo "Memory profile saved to mem.prof. View with: go tool pprof mem.prof"

# CPU usage analysis
.PHONY: cpu-profile
cpu-profile:
	@echo "Running CPU profile tests..."
	go test -cpuprofile=cpu.prof -bench=. ./...
	@echo "CPU profile saved to cpu.prof. View with: go tool pprof cpu.prof"

# Get binary size information
.PHONY: size
size: build
	@echo "Binary size information:"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Deployment package size:"
	@if [ -f "$(DIST_DIR)/$(FUNCTION_NAME).zip" ]; then \
		ls -lh $(DIST_DIR)/$(FUNCTION_NAME).zip; \
	else \
		echo "Package not found. Run 'make package' first."; \
	fi

# Install development tools
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/golang/mock/mockgen@latest

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Clean, build, and test"
	@echo "  build         - Build Lambda function binary"
	@echo "  package       - Create deployment ZIP package"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  bench         - Run benchmarks"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  security      - Run security scan"
	@echo "  vuln-check    - Check for vulnerabilities"
	@echo "  check         - Run all quality checks"
	@echo "  clean         - Clean build artifacts"
	@echo "  tidy          - Tidy dependencies"
	@echo "  deps          - Download dependencies"
	@echo "  size          - Show binary and package size"
	@echo "  install-tools - Install development tools"
	@echo "  help          - Show this help message"