# LogGuardian - Production Makefile
# Focus: Build, Test, Package, Publish, Local Development

# Variables
BINARY_NAME := bootstrap
BUILD_DIR := build
GOOS := linux
GOARCH := amd64
VERSION := $(shell cat VERSION || (echo "Error: VERSION file not found" >&2; exit 1))

# Include SAM targets
include sam.mk

# Default target - complete workflow
.PHONY: all
all: clean build test package

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) .aws-sam packaged-template.yaml coverage.out coverage.html
	go clean -cache -testcache

# Build Lambda function binary for AWS
.PHONY: build
build:
	@echo "Building Lambda function for AWS..."
	mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build \
		-ldflags="-s -w" \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		./cmd/lambda

# Build for local development (current OS)
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

# Code quality checks
.PHONY: check
check: fmt lint vet

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Security scanning
.PHONY: security
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed, skipping"; \
	fi

# Vulnerability check
.PHONY: vuln-check
vuln-check:
	@echo "Checking for vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed, skipping"; \
	fi

# Dependency management
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download

# Install development tools
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

# Binary size analysis
.PHONY: size
size: build
	@echo "Binary size analysis:"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Detailed size breakdown:"
	@if command -v nm >/dev/null 2>&1; then \
		nm -S --size-sort $(BUILD_DIR)/$(BINARY_NAME) | tail -10; \
	else \
		echo "nm not available for size analysis"; \
	fi

# SAM package for AWS Serverless Application Repository
.PHONY: package
package: build
	@echo "Validating SAM template..."
	sam validate --template template.yaml --region ca-central-1
	@echo "Packaging for AWS Serverless Application Repository..."
	sam package \
		--template-file template.yaml \
		--resolve-s3 \
		--output-template-file packaged-template.yaml
	@echo "✅ Packaged template ready: packaged-template.yaml"

# SAM publish to AWS Serverless Application Repository (public)
# Note: VERSION file contains "vX.Y.Z" format for git tags and general use,
# but AWS SAR requires semantic version without "v" prefix (X.Y.Z).
# We extract and strip the "v" prefix using shell parameter expansion.
.PHONY: publish
publish: package
	@echo "Publishing to AWS Serverless Application Repository..."
	@SEMANTIC_VERSION=$$(cat VERSION | sed 's/^v//'); \
	echo "VERSION file: $$(cat VERSION)"; \
	echo "Semantic version for SAR: $$SEMANTIC_VERSION"; \
	sam publish \
		--template packaged-template.yaml \
		--region ca-central-1 \
		--semantic-version $$SEMANTIC_VERSION
	@echo "✅ Application published to SAR"

# Make SAR application public
.PHONY: publish-public
publish-public:
	@echo "Making SAR application public..."
	@echo "⚠️  WARNING: This will make LogGuardian publicly accessible to all AWS users"
	@read -p "Are you sure you want to make this public? (y/N): " confirm && [ "$$confirm" = "y" ]
	aws serverlessrepo put-application-policy \
		--application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
		--statements '[{"Sid":"PublicReadPolicy","Effect":"Allow","Principal":"*","Action":["serverlessrepo:GetApplication","serverlessrepo:CreateCloudFormationTemplate","serverlessrepo:GetCloudFormationTemplate"]}]' \
		--region ca-central-1
	@echo "✅ LogGuardian is now publicly available in AWS SAR!"
	@echo "🔗 Share this URL: https://console.aws.amazon.com/serverlessrepo/home#/published-applications/arn:aws:serverlessrepo:ca-central-1:410129828371:applications~LogGuardian"

# Deploy to production account
.PHONY: deploy-prod
deploy-prod:
	@echo "Deploying LogGuardian to PRODUCTION account..."
	@echo "⚠️  WARNING: This will deploy to your PRODUCTION environment!"
	@read -p "Are you sure you want to deploy to PRODUCTION? (y/N): " confirm && [ "$$confirm" = "y" ]
	@echo "Step 1: Get SAR template..."
	@SEMANTIC_VERSION=$$(cat VERSION | sed 's/^v//'); \
	TEMPLATE_URL=$$(aws serverlessrepo create-cloud-formation-template \
		--application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
		--semantic-version $$SEMANTIC_VERSION \
		--region ca-central-1 \
		--query 'TemplateUrl' --output text) && \
	echo "Template URL: $$TEMPLATE_URL" && \
	echo "Step 2: Download template..." && \
	wget -O logguardian-template.yaml "$$TEMPLATE_URL" && \
	echo "Step 3: Deploy to PRODUCTION..." && \
	aws cloudformation deploy \
		--template-file logguardian-template.yaml \
		--stack-name logguardian-prod \
		--parameter-overrides \
			Environment=prod \
			ProductName=LogGuardian \
			Owner=DevOps-Team \
			KMSKeyAlias=alias/logguardian-prod \
			DefaultRetentionDays=365 \
			LambdaMemorySize=128 \
			LambdaTimeout=30 \
			S3ExpirationDays=90 \
			EnableS3LifecycleRules=true \
			CreateKMSKey=true \
			CreateConfigService=true \
			CreateEncryptionConfigRule=true \
		CreateRetentionConfigRule=true \
			CreateEventBridgeRules=true \
			CreateMonitoringDashboard=true \
			EncryptionScheduleExpression="cron(0 2 ? * SUN *)" \
			RetentionScheduleExpression="cron(0 3 ? * SUN *)" \
			ManagedBy=CloudFormation \
		--capabilities CAPABILITY_NAMED_IAM \
		--region ca-central-1 && \
	echo "✅ PRODUCTION Deployment successful!"

# Deploy to development account
.PHONY: deploy-dev
deploy-dev:
	@echo "Deploying LogGuardian to DEVELOPMENT account..."
	@echo "Step 1: Get SAR template..."
	@SEMANTIC_VERSION=$$(cat VERSION | sed 's/^v//'); \
	TEMPLATE_URL=$$(aws serverlessrepo create-cloud-formation-template \
		--application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
		--semantic-version $$SEMANTIC_VERSION \
		--region ca-central-1 \
		--query 'TemplateUrl' --output text) && \
	echo "Template URL: $$TEMPLATE_URL" && \
	echo "Step 2: Download template..." && \
	wget -O logguardian-template.yaml "$$TEMPLATE_URL" && \
	echo "Step 3: Deploy to DEVELOPMENT..." && \
	aws cloudformation deploy \
		--template-file logguardian-template.yaml \
		--stack-name logguardian-dev \
		--parameter-overrides \
			Environment=dev \
			ProductName=LogGuardian-Dev \
			Owner=DevOps-Team \
			KMSKeyAlias=alias/logguardian-dev \
			DefaultRetentionDays=90 \
			LambdaMemorySize=128 \
			LambdaTimeout=30 \
			S3ExpirationDays=30 \
			EnableS3LifecycleRules=true \
			CreateKMSKey=true \
			CreateConfigService=true \
			CreateEncryptionConfigRule=true \
		CreateRetentionConfigRule=true \
			CreateEventBridgeRules=false \
			CreateMonitoringDashboard=false \
			EncryptionScheduleExpression="cron(0 2 ? * SUN *)" \
			RetentionScheduleExpression="cron(0 3 ? * SUN *)" \
			ManagedBy=CloudFormation \
		--capabilities CAPABILITY_NAMED_IAM \
		--region ca-central-1 && \
	echo "✅ DEVELOPMENT Deployment successful!"

# Clean up development deployment
.PHONY: cleanup-dev
cleanup-dev:
	@echo "Cleaning up DEVELOPMENT deployment..."
	aws cloudformation delete-stack --stack-name logguardian-dev --region ca-central-1
	rm -f logguardian-template.yaml
	@echo "✅ DEVELOPMENT cleanup complete"

# Clean up production deployment
.PHONY: cleanup-prod
cleanup-prod:
	@echo "Cleaning up PRODUCTION deployment..."
	@echo "⚠️  WARNING: This will DELETE the LogGuardian stack from PRODUCTION!"
	@read -p "Are you sure you want to delete from PRODUCTION? (y/N): " confirm && [ "$$confirm" = "y" ]
	aws cloudformation delete-stack --stack-name logguardian-prod --region ca-central-1
	rm -f logguardian-template.yaml
	@echo "✅ PRODUCTION cleanup complete"

# Help
.PHONY: help
help:
	@echo "LogGuardian - Production Build System"
	@echo "===================================="
	@echo ""
	@echo "Development:"
	@echo "  all              - Clean, build, test, package"
	@echo "  build            - Build Lambda binary for AWS"
	@echo "  dev-build        - Build for local development"
	@echo "  test             - Run tests with coverage"
	@echo "  bench            - Run benchmarks"
	@echo "  check            - Format, lint, vet, security scan"
	@echo "  clean            - Clean all build artifacts"
	@echo "  tidy             - Tidy dependencies"
	@echo "  deps             - Download dependencies"
	@echo "  install-tools    - Install development tools"
	@echo "  size             - Show binary size analysis"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt              - Format code"
	@echo "  lint             - Lint with golangci-lint"
	@echo "  vet              - Run go vet"
	@echo "  security         - Security scan with gosec"
	@echo "  vuln-check       - Check for vulnerabilities"
	@echo ""
	@echo "Production:"
	@echo "  package          - Package for AWS SAR"
	@echo "  publish          - Publish to AWS SAR (public)"
	@echo "  deploy-prod      - Deploy to PRODUCTION account"
	@echo "  deploy-dev       - Deploy to DEVELOPMENT account"
	@echo "  cleanup-prod     - Clean up PRODUCTION deployment"
	@echo "  cleanup-dev      - Clean up DEVELOPMENT deployment"
	@echo ""
	@echo "📦 SAM Local Testing:"
	@echo "  sam-build        - Build for SAM"
	@echo "  sam-test-quick   - Quick local testing"
	@echo "  sam-test-all-events - Test all event scenarios"
	@echo "  sam-local-invoke - Test with config rule event"
	@echo "  sam-validate     - Validate SAM template"
	@echo ""
	@echo "Usage Example:"
	@echo "  make all         # Development workflow"
	@echo "  make publish     # Production release"
	@echo "  make deploy-dev  # Deploy to development account"
	@echo "  make deploy-prod # Deploy to production account"
	@echo ""
	@echo "For complete SAM targets, see: sam.mk"

# =============================================================================
# Container Build Targets
# =============================================================================

# Container variables
CONTAINER_IMAGE := logguardian
CONTAINER_TAG := $(VERSION)
CONTAINER_REGISTRY := 
CONTAINER_BINARY := logguardian-container

# Build container binary
.PHONY: container-build
container-build:
	@echo "Building container binary..."
	mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build \
		-ldflags="-s -w -X main.version=$(VERSION)" \
		-o $(BUILD_DIR)/$(CONTAINER_BINARY) \
		./cmd/container

# Build Docker image
.PHONY: docker-build
docker-build:
	@echo "Building Docker image $(CONTAINER_IMAGE):$(CONTAINER_TAG)..."
	docker build -t $(CONTAINER_IMAGE):$(CONTAINER_TAG) \
		-t $(CONTAINER_IMAGE):latest \
		--build-arg VERSION=$(VERSION) \
		.

# Run container locally with dry-run
.PHONY: docker-run-dryrun
docker-run-dryrun: docker-build
	@echo "Running container in dry-run mode..."
	docker run --rm \
		-e AWS_REGION=ca-central-1 \
		-e CONFIG_RULE_NAME=test-rule \
		-e DRY_RUN=true \
		-v ~/.aws:/home/logguardian/.aws:ro \
		$(CONTAINER_IMAGE):$(CONTAINER_TAG) \
		--dry-run --verbose

# Run container with AWS credentials
.PHONY: docker-run
docker-run: docker-build
	@echo "Running container..."
	docker run --rm \
		-e AWS_REGION=$(AWS_REGION) \
		-e CONFIG_RULE_NAME=$(CONFIG_RULE_NAME) \
		-v ~/.aws:/home/logguardian/.aws:ro \
		$(CONTAINER_IMAGE):$(CONTAINER_TAG)

# Push container to registry
.PHONY: docker-push
docker-push:
	@if [ -z "$(CONTAINER_REGISTRY)" ]; then \
		echo "Error: CONTAINER_REGISTRY not set"; \
		exit 1; \
	fi
	@echo "Tagging image for registry..."
	docker tag $(CONTAINER_IMAGE):$(CONTAINER_TAG) $(CONTAINER_REGISTRY)/$(CONTAINER_IMAGE):$(CONTAINER_TAG)
	docker tag $(CONTAINER_IMAGE):$(CONTAINER_TAG) $(CONTAINER_REGISTRY)/$(CONTAINER_IMAGE):latest
	@echo "Pushing to registry..."
	docker push $(CONTAINER_REGISTRY)/$(CONTAINER_IMAGE):$(CONTAINER_TAG)
	docker push $(CONTAINER_REGISTRY)/$(CONTAINER_IMAGE):latest

# Clean container artifacts
.PHONY: docker-clean
docker-clean:
	@echo "Cleaning container images..."
	docker rmi $(CONTAINER_IMAGE):$(CONTAINER_TAG) $(CONTAINER_IMAGE):latest 2>/dev/null || true
	rm -f $(BUILD_DIR)/$(CONTAINER_BINARY)

# Run container tests
.PHONY: container-test
container-test:
	@echo "Running container tests..."
	go test -v ./cmd/container/... ./internal/container/...

# Validate container build
.PHONY: container-validate
container-validate: container-build container-test docker-build
	@echo "Container validation complete"

# Scan container for vulnerabilities
.PHONY: docker-scan
docker-scan: docker-build
	@echo "Scanning container for vulnerabilities..."
	@which trivy > /dev/null 2>&1 || (echo "Trivy not installed. Install from https://github.com/aquasecurity/trivy" && exit 1)
	trivy image $(CONTAINER_IMAGE):$(CONTAINER_TAG)