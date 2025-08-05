# LogGuardian Go Lambda Function Makefile

# Variables
FUNCTION_NAME := logguardian-compliance
# The binary is named 'bootstrap' for AWS Lambda AL2023 runtime compatibility
BINARY_NAME := bootstrap
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

# CloudFormation template validation
.PHONY: validate-templates
validate-templates:
	@echo "Validating CloudFormation templates..."
	@if command -v aws >/dev/null 2>&1; then \
		cd templates && ./90-validate-templates.sh; \
	else \
		echo "AWS CLI not found. Skipping template validation."; \
	fi

# Deploy to development environment
.PHONY: deploy-dev
deploy-dev: package validate-templates
	@echo "Deploying to development environment..."
	@if [ -z "$(DEPLOYMENT_BUCKET)" ]; then \
		echo "Error: DEPLOYMENT_BUCKET environment variable is required"; \
		exit 1; \
	fi
	@if [ -z "$(AWS_REGION)" ]; then \
		echo "Error: AWS_REGION environment variable is required"; \
		exit 1; \
	fi
	aws s3 cp $(DIST_DIR)/$(FUNCTION_NAME).zip s3://$(DEPLOYMENT_BUCKET)/
	aws cloudformation deploy \
		--template-file templates/00-logguardian-simple.yaml \
		--stack-name logguardian-dev-$(AWS_REGION) \
		--parameter-overrides Environment=dev DeploymentBucket=$(DEPLOYMENT_BUCKET) \
		--capabilities CAPABILITY_NAMED_IAM \
		--region $(AWS_REGION)

# Deploy to staging environment
.PHONY: deploy-staging
deploy-staging: package validate-templates
	@echo "Deploying to staging environment..."
	@if [ -z "$(DEPLOYMENT_BUCKET)" ]; then \
		echo "Error: DEPLOYMENT_BUCKET environment variable is required"; \
		exit 1; \
	fi
	@if [ -z "$(AWS_REGION)" ]; then \
		echo "Error: AWS_REGION environment variable is required"; \
		exit 1; \
	fi
	aws s3 cp $(DIST_DIR)/$(FUNCTION_NAME).zip s3://$(DEPLOYMENT_BUCKET)/
	aws s3 sync templates/ s3://$(DEPLOYMENT_BUCKET)/templates/
	aws cloudformation deploy \
		--template-file templates/01-logguardian-main.yaml \
		--stack-name logguardian-staging-$(AWS_REGION) \
		--parameter-overrides Environment=staging DeploymentBucket=$(DEPLOYMENT_BUCKET) \
		--capabilities CAPABILITY_NAMED_IAM \
		--region $(AWS_REGION)

# Deploy to production environment
.PHONY: deploy-prod
deploy-prod: package validate-templates
	@echo "Deploying to production environment..."
	@if [ -z "$(DEPLOYMENT_BUCKET)" ]; then \
		echo "Error: DEPLOYMENT_BUCKET environment variable is required"; \
		exit 1; \
	fi
	@if [ -z "$(AWS_REGION)" ]; then \
		echo "Error: AWS_REGION environment variable is required"; \
		exit 1; \
	fi
	@echo "WARNING: Deploying to production environment!"
	@read -p "Are you sure you want to deploy to production? (y/N): " confirm && [ "$$confirm" = "y" ]
	aws s3 cp $(DIST_DIR)/$(FUNCTION_NAME).zip s3://$(DEPLOYMENT_BUCKET)/
	aws s3 sync templates/ s3://$(DEPLOYMENT_BUCKET)/templates/
	aws cloudformation deploy \
		--template-file templates/01-logguardian-main.yaml \
		--stack-name logguardian-prod-$(AWS_REGION) \
		--parameter-overrides Environment=prod DeploymentBucket=$(DEPLOYMENT_BUCKET) \
		--capabilities CAPABILITY_NAMED_IAM \
		--region $(AWS_REGION)

# Upload templates to S3 for nested stack deployment
.PHONY: upload-templates
upload-templates:
	@echo "Uploading CloudFormation templates to S3..."
	@if [ -z "$(DEPLOYMENT_BUCKET)" ]; then \
		echo "Error: DEPLOYMENT_BUCKET environment variable is required"; \
		exit 1; \
	fi
	aws s3 sync templates/ s3://$(DEPLOYMENT_BUCKET)/templates/
	@echo "Templates uploaded to s3://$(DEPLOYMENT_BUCKET)/templates/"
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

# Show help information
.PHONY: help
help:
	@echo "LogGuardian Go Lambda Function - Available Make Targets"
	@echo "======================================================"
	@echo ""
	@echo "Build Targets:"
	@echo "  build           - Build the Lambda function binary"
	@echo "  package         - Create deployment ZIP package"
	@echo "  clean           - Clean build artifacts"
	@echo ""
	@echo "Testing Targets:"
	@echo "  test            - Run unit tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  bench           - Run benchmarks"
	@echo ""
	@echo "Code Quality Targets:"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code with golangci-lint"
	@echo "  security        - Run security scan with gosec"
	@echo "  vuln-check      - Check for vulnerabilities"
	@echo "  check           - Run all quality checks"
	@echo ""
	@echo "CloudFormation Targets:"
	@echo "  validate-templates - Validate CloudFormation templates"
	@echo "  upload-templates   - Upload templates to S3"
	@echo ""
	@echo "Deployment Targets:"
	@echo "  deploy-dev      - Deploy to development environment"
	@echo "  deploy-staging  - Deploy to staging environment"
	@echo "  deploy-prod     - Deploy to production environment"
	@echo ""
	@echo "Utility Targets:"
	@echo "  deps            - Download dependencies"
	@echo "  tidy            - Tidy and verify dependencies"
	@echo "  mocks           - Generate test mocks"
	@echo "  memory-profile  - Run memory profiling"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "Environment Variables:"
	@echo "  DEPLOYMENT_BUCKET - S3 bucket for deployment artifacts (required for deployment)"
	@echo "  AWS_REGION        - AWS region for deployment (required for deployment)"
	@echo ""
	@echo "Examples:"
	@echo "  make build test                           # Build and test"
	@echo "  make package                              # Create deployment package"
	@echo "  DEPLOYMENT_BUCKET=my-bucket AWS_REGION=ca-central-1 make deploy-dev"
	@echo "  make validate-templates                   # Validate CloudFormation templates"

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