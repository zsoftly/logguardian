# LogGuardian - Production Makefile
# Focus: Build, Test, Package, Publish, Local Development

# Variables
BINARY_NAME := bootstrap
BUILD_DIR := build
GOOS := linux
GOARCH := amd64

# Include SAM targets
include sam.mk

# Default target - complete workflow
.PHONY: all
all: clean build test package

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) .aws-sam packaged-template*.yaml coverage.out coverage.html
	rm -f logguardian-*-template.yaml *.zip response*.json test-*.json
	go clean -cache -testcache

# Build Lambda function binary for AWS
.PHONY: build
build:
	@echo "Building LogGuardian Lambda function for AWS..."
	mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build \
		-ldflags="-s -w" \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		./cmd/lambda
	@echo "Building Custom Config Rule Lambda function for AWS..."
	mkdir -p $(BUILD_DIR)/config-rule
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build \
		-ldflags="-s -w" \
		-o $(BUILD_DIR)/config-rule/$(BINARY_NAME) \
		./cmd/config-rule

# Build main LogGuardian Lambda only
.PHONY: build-main
build-main:
	@echo "Building main LogGuardian Lambda function for AWS..."
	mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build \
		-ldflags="-s -w" \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		./cmd/lambda

# Build custom config rule Lambda only
.PHONY: build-config-rule
build-config-rule:
	@echo "Building Custom Config Rule Lambda function for AWS..."
	mkdir -p $(BUILD_DIR)/config-rule
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build \
		-ldflags="-s -w" \
		-o $(BUILD_DIR)/config-rule/$(BINARY_NAME) \
		./cmd/config-rule

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
package: package-main package-config-rule

# Package main LogGuardian Lambda
.PHONY: package-main
package-main: build-main
	@echo "Validating main LogGuardian SAM template..."
	sam validate --template template-main.yaml --region ca-central-1
	@echo "Packaging main LogGuardian for AWS Serverless Application Repository..."
	sam package \
		--template-file template-main.yaml \
		--resolve-s3 \
		--output-template-file packaged-template-main.yaml
	@echo "✅ Main LogGuardian packaged: packaged-template-main.yaml"

# Package custom config rule Lambda
.PHONY: package-config-rule
package-config-rule: build-config-rule
	@echo "Validating custom config rule SAM template..."
	sam validate --template template-config-rule.yaml --region ca-central-1
	@echo "Packaging custom config rule for AWS Serverless Application Repository..."
	sam package \
		--template-file template-config-rule.yaml \
		--resolve-s3 \
		--output-template-file packaged-template-config-rule.yaml
	@echo "✅ Custom config rule packaged: packaged-template-config-rule.yaml"

# Package combined template (legacy)
.PHONY: package-combined
package-combined: build
	@echo "Validating combined SAM template..."
	sam validate --template template.yaml --region ca-central-1
	@echo "Packaging combined template for AWS Serverless Application Repository..."
	sam package \
		--template-file template.yaml \
		--resolve-s3 \
		--output-template-file packaged-template.yaml
	@echo "✅ Combined template packaged: packaged-template.yaml"

# SAM publish to AWS Serverless Application Repository (public)
.PHONY: publish
publish: publish-main publish-config-rule

# Publish main LogGuardian to SAR
.PHONY: publish-main
publish-main: package-main
	@echo "Publishing main LogGuardian to AWS Serverless Application Repository..."
	sam publish \
		--template packaged-template-main.yaml \
		--region ca-central-1 \
		--semantic-version 1.0.3
	@echo "✅ Main LogGuardian published to SAR"

# Publish custom config rule to SAR
.PHONY: publish-config-rule
publish-config-rule: package-config-rule
	@echo "Publishing custom config rule to AWS Serverless Application Repository..."
	sam publish \
		--template packaged-template-config-rule.yaml \
		--region ca-central-1 \
		--semantic-version 1.0.3
	@echo "✅ Custom config rule published to SAR"

# Publish combined template (legacy)
.PHONY: publish-combined
publish-combined: package-combined
	@echo "Publishing combined template to AWS Serverless Application Repository..."
	sam publish \
		--template packaged-template.yaml \
		--region ca-central-1 \
		--semantic-version 1.0.3
	@echo "✅ Combined application published to SAR"

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
	TEMPLATE_URL=$$(aws serverlessrepo create-cloud-formation-template \
		--application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
		--semantic-version 1.0.2 \
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
			CreateConfigRules=true \
			CreateEventBridgeRules=true \
			CreateMonitoringDashboard=true \
			UseCustomRetentionRule=true \
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
	TEMPLATE_URL=$$(aws serverlessrepo create-cloud-formation-template \
		--application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian \
		--semantic-version 1.0.2 \
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
			CreateConfigRules=true \
			CreateEventBridgeRules=false \
			CreateMonitoringDashboard=false \
			UseCustomRetentionRule=true \
			EncryptionScheduleExpression="cron(0 2 ? * SUN *)" \
			RetentionScheduleExpression="cron(0 3 ? * SUN *)" \
			ManagedBy=CloudFormation \
		--capabilities CAPABILITY_NAMED_IAM \
		--region ca-central-1 && \
	echo "✅ DEVELOPMENT Deployment successful!"

# === SEPARATE DEPLOYMENT TARGETS ===

# Deploy main LogGuardian only to production
.PHONY: deploy-main-prod
deploy-main-prod:
	@echo "Deploying main LogGuardian to PRODUCTION account..."
	@echo "⚠️  WARNING: This will deploy main LogGuardian to your PRODUCTION environment!"
	@read -p "Are you sure you want to deploy main LogGuardian to PRODUCTION? (y/N): " confirm && [ "$$confirm" = "y" ]
	@echo "Step 1: Get main LogGuardian SAR template..."
	TEMPLATE_URL=$$(aws serverlessrepo create-cloud-formation-template \
		--application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian-Main \
		--semantic-version 1.0.3 \
		--region ca-central-1 \
		--query 'TemplateUrl' --output text) && \
	echo "Template URL: $$TEMPLATE_URL" && \
	echo "Step 2: Download template..." && \
	wget -O logguardian-main-template.yaml "$$TEMPLATE_URL" && \
	echo "Step 3: Deploy main LogGuardian to PRODUCTION..." && \
	aws cloudformation deploy \
		--template-file logguardian-main-template.yaml \
		--stack-name logguardian-main-prod \
		--parameter-overrides \
			Environment=prod \
			ProductName=LogGuardian-Main \
			Owner=DevOps-Team \
			KMSKeyAlias=alias/logguardian-prod \
			DefaultRetentionDays=365 \
			LambdaMemorySize=128 \
			LambdaTimeout=30 \
			ManagedBy=CloudFormation \
		--capabilities CAPABILITY_NAMED_IAM \
		--region ca-central-1 && \
	echo "✅ Main LogGuardian PRODUCTION Deployment successful!"

# Deploy custom config rule only to production
.PHONY: deploy-config-rule-prod
deploy-config-rule-prod:
	@echo "Deploying custom config rule to PRODUCTION account..."
	@echo "⚠️  WARNING: This will deploy custom config rule to your PRODUCTION environment!"
	@read -p "Are you sure you want to deploy custom config rule to PRODUCTION? (y/N): " confirm && [ "$$confirm" = "y" ]
	@echo "Step 1: Get custom config rule SAR template..."
	TEMPLATE_URL=$$(aws serverlessrepo create-cloud-formation-template \
		--application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian-CustomConfigRule \
		--semantic-version 1.0.3 \
		--region ca-central-1 \
		--query 'TemplateUrl' --output text) && \
	echo "Template URL: $$TEMPLATE_URL" && \
	echo "Step 2: Download template..." && \
	wget -O logguardian-config-rule-template.yaml "$$TEMPLATE_URL" && \
	echo "Step 3: Deploy custom config rule to PRODUCTION..." && \
	aws cloudformation deploy \
		--template-file logguardian-config-rule-template.yaml \
		--stack-name logguardian-config-rule-prod \
		--parameter-overrides \
			Environment=prod \
			ProductName=LogGuardian-CustomRule \
			Owner=DevOps-Team \
			DefaultRetentionDays=30 \
			CustomRetentionRuleMemorySize=128 \
			ManagedBy=CloudFormation \
		--capabilities CAPABILITY_NAMED_IAM \
		--region ca-central-1 && \
	echo "✅ Custom config rule PRODUCTION Deployment successful!"

# Deploy both main and config rule to production
.PHONY: deploy-separate-prod
deploy-separate-prod: deploy-main-prod deploy-config-rule-prod
	@echo "✅ Both main LogGuardian and custom config rule deployed to PRODUCTION!"

# Deploy main LogGuardian only to development
.PHONY: deploy-main-dev
deploy-main-dev:
	@echo "Deploying main LogGuardian to DEVELOPMENT account..."
	TEMPLATE_URL=$$(aws serverlessrepo create-cloud-formation-template \
		--application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian-Main \
		--semantic-version 1.0.3 \
		--region ca-central-1 \
		--query 'TemplateUrl' --output text) && \
	echo "Template URL: $$TEMPLATE_URL" && \
	wget -O logguardian-main-template.yaml "$$TEMPLATE_URL" && \
	aws cloudformation deploy \
		--template-file logguardian-main-template.yaml \
		--stack-name logguardian-main-dev \
		--parameter-overrides \
			Environment=dev \
			ProductName=LogGuardian-Main-Dev \
			Owner=DevOps-Team \
			KMSKeyAlias=alias/logguardian-dev \
			DefaultRetentionDays=90 \
			LambdaMemorySize=128 \
			LambdaTimeout=30 \
			ManagedBy=CloudFormation \
		--capabilities CAPABILITY_NAMED_IAM \
		--region ca-central-1 && \
	echo "✅ Main LogGuardian DEVELOPMENT Deployment successful!"

# Deploy custom config rule only to development
.PHONY: deploy-config-rule-dev
deploy-config-rule-dev:
	@echo "Deploying custom config rule to DEVELOPMENT account..."
	TEMPLATE_URL=$$(aws serverlessrepo create-cloud-formation-template \
		--application-id arn:aws:serverlessrepo:ca-central-1:410129828371:applications/LogGuardian-CustomConfigRule \
		--semantic-version 1.0.3 \
		--region ca-central-1 \
		--query 'TemplateUrl' --output text) && \
	echo "Template URL: $$TEMPLATE_URL" && \
	wget -O logguardian-config-rule-template.yaml "$$TEMPLATE_URL" && \
	aws cloudformation deploy \
		--template-file logguardian-config-rule-template.yaml \
		--stack-name logguardian-config-rule-dev \
		--parameter-overrides \
			Environment=dev \
			ProductName=LogGuardian-CustomRule-Dev \
			Owner=DevOps-Team \
			DefaultRetentionDays=30 \
			CustomRetentionRuleMemorySize=128 \
			ManagedBy=CloudFormation \
		--capabilities CAPABILITY_NAMED_IAM \
		--region ca-central-1 && \
	echo "✅ Custom config rule DEVELOPMENT Deployment successful!"

# Deploy both main and config rule to development
.PHONY: deploy-separate-dev
deploy-separate-dev: deploy-main-dev deploy-config-rule-dev
	@echo "✅ Both main LogGuardian and custom config rule deployed to DEVELOPMENT!"

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
	@echo "  build            - Build both Lambda functions"
	@echo "  build-main       - Build main LogGuardian Lambda only"
	@echo "  build-config-rule - Build custom config rule Lambda only"
	@echo "  dev-build        - Build both for local development"
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
	@echo "Packaging:"
	@echo "  package          - Package both applications"
	@echo "  package-main     - Package main LogGuardian only"
	@echo "  package-config-rule - Package custom config rule only"
	@echo "  package-combined - Package legacy combined template"
	@echo ""
	@echo "Publishing:"
	@echo "  publish          - Publish both to AWS SAR"
	@echo "  publish-main     - Publish main LogGuardian to SAR"
	@echo "  publish-config-rule - Publish custom config rule to SAR"
	@echo "  publish-combined - Publish legacy combined template"
	@echo ""
	@echo "Combined Deployment (Legacy):"
	@echo "  deploy-prod      - Deploy combined template to PRODUCTION"
	@echo "  deploy-dev       - Deploy combined template to DEVELOPMENT"
	@echo ""
	@echo "Separate Deployment (Recommended):"
	@echo "  deploy-separate-prod     - Deploy both separately to PRODUCTION"
	@echo "  deploy-separate-dev      - Deploy both separately to DEVELOPMENT"
	@echo "  deploy-main-prod         - Deploy main LogGuardian to PRODUCTION"
	@echo "  deploy-main-dev          - Deploy main LogGuardian to DEVELOPMENT"
	@echo "  deploy-config-rule-prod  - Deploy custom config rule to PRODUCTION"
	@echo "  deploy-config-rule-dev   - Deploy custom config rule to DEVELOPMENT"
	@echo ""
	@echo "Cleanup:"
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