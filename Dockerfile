# Multi-stage build for LogGuardian container
# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the container binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s -X main.version=$(cat VERSION 2>/dev/null || echo 'unknown')" \
    -o logguardian-container \
    cmd/container/main.go

# Stage 2: Runtime stage
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 1000 -g 1000 logguardian

# Copy timezone data for proper timestamp handling
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary from builder
COPY --from=builder /build/logguardian-container /usr/local/bin/logguardian

# Set proper permissions
RUN chmod +x /usr/local/bin/logguardian

# Switch to non-root user
USER logguardian

# Set environment variables for container operation
ENV AWS_REGION=""
ENV CONFIG_RULE_NAME=""
ENV BATCH_SIZE="10"
ENV DRY_RUN="false"
ENV APP_VERSION="1.3.0"

# Health check endpoint (container exits after execution, so this is mainly for build verification)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=1 \
    CMD /usr/local/bin/logguardian --help || exit 1

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/logguardian"]

# Default command (show help)
CMD ["--help"]