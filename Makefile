# Use bash everywhere (CI-safe)
SHELL := bash

# List only handler packages for tests & coverage
TEST_PKGS := $(shell go list ./internal/handlers/...)
# coverpkg must be comma-separated for multiple packages
COVERPKG  := $(shell echo $(TEST_PKGS) | tr ' ' ',')

COVERFILE := coverage.out
COVERHTML := coverage.html

# Common test flags (tweak as needed)
GOTESTFLAGS ?= -race

.PHONY: all tools mocks unit cover bench clean integration

all: unit

tools:
	@command -v mockgen >/dev/null 2>&1 || go install github.com/golang/mock/mockgen@v1.6.0

mocks: tools
	mockgen -destination=mocks/mock_s3.go -package=mocks \
		unit-test/internal/awsiface S3PutObject

unit:
	# Run tests for handlers and generate coverage over the same set
	go test $(GOTESTFLAGS) -covermode=atomic -coverpkg=$(COVERPKG) \
		-coverprofile=$(COVERFILE) $(TEST_PKGS)
	bash ./scripts/coverage_check.sh 90 $(COVERFILE)

cover: unit
	go tool cover -func=$(COVERFILE)
	go tool cover -html=$(COVERFILE) -o $(COVERHTML)

bench:
	go test -run=^$ -bench=. -benchmem $(TEST_PKGS)

# Runs integration tests (LocalStack or real AWS based on .env)
integration:
	# godotenv in TestMain loads .env; just pass the tag
	go test -tags=integration ./...

clean:
	rm -f $(COVERFILE) $(COVERHTML)
