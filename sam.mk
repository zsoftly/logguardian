# Build for SAM deployment
.PHONY: sam-build
sam-build: build
	@echo "Preparing SAM build directory..."
	mkdir -p .aws-sam/build/LogGuardianFunction
	cp $(BUILD_DIR)/$(BINARY_NAME) .aws-sam/build/LogGuardianFunction/
	@echo "SAM build ready"

# SAM local testing with environment
.PHONY: sam-local-start
sam-local-start: sam-build
	@echo "Starting SAM local API..."
	@echo "Loading local environment variables..."
	@if [ -f ".env.local" ]; then \
		export $$(grep -v '^#' .env.local | xargs) && sam local start-api --env-vars .env.local; \
	else \
		sam local start-api; \
	fi

# SAM local invoke with different event types
.PHONY: sam-local-invoke
sam-local-invoke: sam-build
	@echo "Invoking function locally with config-rule-evaluation event..."
	@if [ -f "testdata/config-rule-evaluation-event.json" ]; then \
		sam local invoke LogGuardianFunction --event testdata/config-rule-evaluation-event.json --env-vars env.json; \
	else \
		echo "No config-rule-evaluation event file found"; \
	fi

.PHONY: sam-local-invoke-config
sam-local-invoke-config: sam-build
	@echo "Invoking function locally with individual config event..."
	@if [ -f "testdata/unified-config-event.json" ]; then \
		sam local invoke LogGuardianFunction --event testdata/unified-config-event.json --env-vars env.json; \
	else \
		echo "No unified config event file found"; \
	fi

.PHONY: sam-local-invoke-retention
sam-local-invoke-retention: sam-build
	@echo "Invoking function locally with retention rule evaluation event..."
	@if [ -f "testdata/retention-rule-evaluation-event.json" ]; then \
		sam local invoke LogGuardianFunction --event testdata/retention-rule-evaluation-event.json --env-vars env.json; \
	else \
		echo "No retention rule evaluation event file found"; \
	fi

.PHONY: sam-local-invoke-retention-config
sam-local-invoke-retention-config: sam-build
	@echo "Invoking function locally with retention config event..."
	@if [ -f "testdata/retention-config-event.json" ]; then \
		sam local invoke LogGuardianFunction --event testdata/retention-config-event.json --env-vars env.json; \
	else \
		echo "No retention config event file found"; \
	fi

.PHONY: sam-local-invoke-compliant
sam-local-invoke-compliant: sam-build
	@echo "Invoking function locally with compliant log group event..."
	@if [ -f "testdata/compliant-log-group-event.json" ]; then \
		sam local invoke LogGuardianFunction --event testdata/compliant-log-group-event.json --env-vars env.json; \
	else \
		echo "No compliant log group event file found"; \
	fi

.PHONY: sam-local-invoke-both-missing
sam-local-invoke-both-missing: sam-build
	@echo "Invoking function locally with both encryption and retention missing event..."
	@if [ -f "testdata/both-missing-config-event.json" ]; then \
		sam local invoke LogGuardianFunction --event testdata/both-missing-config-event.json --env-vars env.json; \
	else \
		echo "No both missing config event file found"; \
	fi

.PHONY: sam-local-invoke-invalid
sam-local-invoke-invalid: sam-build
	@echo "Invoking function locally with invalid event type..."
	@if [ -f "testdata/invalid-event-type.json" ]; then \
		sam local invoke LogGuardianFunction --event testdata/invalid-event-type.json --env-vars env.json; \
	else \
		echo "No invalid event type file found"; \
	fi

.PHONY: sam-local-invoke-large-batch
sam-local-invoke-large-batch: sam-build
	@echo "Invoking function locally with large batch evaluation event..."
	@if [ -f "testdata/large-batch-evaluation-event.json" ]; then \
		sam local invoke LogGuardianFunction --event testdata/large-batch-evaluation-event.json --env-vars env.json; \
	else \
		echo "No large batch evaluation event file found"; \
	fi

.PHONY: sam-test-all-events
sam-test-all-events: sam-build
	@echo "Testing all event types locally..."
	@echo ""
	@echo "=== 1. Config Rule Evaluation Events ==="
	@echo "1a. Testing encryption rule evaluation..."
	@make sam-local-invoke
	@echo ""
	@echo "1b. Testing retention rule evaluation..."
	@make sam-local-invoke-retention
	@echo ""
	@echo "1c. Testing large batch evaluation..."
	@make sam-local-invoke-large-batch
	@echo ""
	@echo "=== 2. Individual Config Events ==="
	@echo "2a. Testing missing encryption..."
	@make sam-local-invoke-config
	@echo ""
	@echo "2b. Testing missing retention..."
	@make sam-local-invoke-retention-config
	@echo ""
	@echo "2c. Testing compliant log group..."
	@make sam-local-invoke-compliant
	@echo ""
	@echo "2d. Testing both encryption and retention missing..."
	@make sam-local-invoke-both-missing
	@echo ""
	@echo "=== 3. Error Handling Tests ==="
	@echo "3a. Testing invalid event type..."
	@make sam-local-invoke-invalid || true
	@echo ""
	@echo "=== Test Summary Complete ==="

# Quick test with most common scenarios
.PHONY: sam-test-quick
sam-test-quick: sam-build
	@echo "Quick test of common scenarios..."
	@echo "1. Config rule evaluation (encryption)..."
	@make sam-local-invoke
	@echo ""
	@echo "2. Individual config event (missing encryption)..."
	@make sam-local-invoke-config
	@echo ""
	@echo "3. Retention issue..."
	@make sam-local-invoke-retention-config
	@echo ""
	@echo "Quick test complete!"

# Test specific scenarios
.PHONY: sam-test-encryption
sam-test-encryption: sam-build
	@echo "Testing encryption scenarios..."
	@make sam-local-invoke
	@make sam-local-invoke-config

.PHONY: sam-test-retention  
sam-test-retention: sam-build
	@echo "Testing retention scenarios..."
	@make sam-local-invoke-retention
	@make sam-local-invoke-retention-config

.PHONY: sam-test-errors
sam-test-errors: sam-build
	@echo "Testing error handling..."
	@make sam-local-invoke-invalid || true

# SAM validate template
.PHONY: sam-validate
sam-validate:
	@echo "Validating SAM template..."
	sam validate --template template.yaml --region ca-central-1

# SAM deploy for development
.PHONY: sam-deploy-dev
sam-deploy-dev: sam-build sam-validate
	@echo "Deploying with SAM (development)..."
	sam deploy \
		--template-file template.yaml \
		--stack-name logguardian-dev \
		--capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM \
		--parameter-overrides Environment=dev \
		--resolve-s3

# SAM deploy for production
.PHONY: sam-deploy-prod
sam-deploy-prod: sam-build sam-validate
	@echo "Deploying with SAM (production)..."
	@echo "WARNING: Deploying to production environment!"
	@read -p "Are you sure you want to deploy to production? (y/N): " confirm && [ "$$confirm" = "y" ]
	sam deploy \
		--template-file template.yaml \
		--stack-name logguardian-prod \
		--capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM \
		--parameter-overrides Environment=prod \
		--resolve-s3

# SAM package for AWS Marketplace
.PHONY: sam-package-marketplace
sam-package-marketplace: sam-build sam-validate
	@echo "Packaging for AWS Marketplace..."
	@if [ -z "$(MARKETPLACE_BUCKET)" ]; then \
		echo "Error: MARKETPLACE_BUCKET environment variable is required"; \
		exit 1; \
	fi
	sam package \
		--template-file template.yaml \
		--s3-bucket $(MARKETPLACE_BUCKET) \
		--output-template-file packaged-template.yaml
	@echo "Packaged template ready: packaged-template.yaml"

# SAM publish to AWS Serverless Application Repository
.PHONY: sam-publish
sam-publish: sam-package-marketplace
	@echo "Publishing to AWS Serverless Application Repository..."
	sam publish \
		--template packaged-template.yaml \
		--region us-east-1
	@echo "Application published to SAR"

# Clean SAM artifacts
.PHONY: sam-clean
sam-clean:
	@echo "Cleaning SAM artifacts..."
	rm -rf .aws-sam/
	rm -f packaged-template.yaml

# Complete SAM workflow for marketplace
.PHONY: sam-marketplace-ready
sam-marketplace-ready: clean sam-build sam-validate test security sam-package-marketplace
	@echo "LogGuardian is ready for AWS Marketplace deployment!"
	@echo "Next steps:"
	@echo "1. Review packaged-template.yaml"
	@echo "2. Run 'make sam-publish' to publish to AWS Serverless Application Repository"
	@echo "3. Submit to AWS Marketplace Partner Portal"
