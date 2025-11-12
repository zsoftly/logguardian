#!/usr/bin/env bash
#
set -euo pipefail
# LogGuardian Deployment
# Validates and deploys LogGuardian infrastructure
#

set -euo pipefail

readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m'

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly TF_DIR="${SCRIPT_DIR}/.."

log_info() { echo -e "${GREEN}[INFO]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

# Pre-flight validation
validate() {
    log_info "Running pre-flight checks..."
    
    command -v terraform >/dev/null 2>&1 || log_error "terraform not installed"
    command -v aws >/dev/null 2>&1 || log_error "aws cli not installed"
    aws sts get-caller-identity >/dev/null 2>&1 || log_error "AWS credentials not configured"
    
    local env="$1"
    local bucket="$2"
    
    [[ "${env}" =~ ^(dev|staging|prod|sandbox)$ ]] || log_error "Invalid environment: ${env}"
    aws s3 ls "s3://${bucket}" >/dev/null 2>&1 || log_error "S3 bucket not accessible: ${bucket}"
    
    log_info "Pre-flight checks passed ✓"
}

create_config() {
    cat > "${TF_DIR}/terraform.tfvars" <<EOF
environment      = "$1"
lambda_s3_bucket = "$2"

create_kms_key            = true
create_config_service     = false
create_config_rules       = true
create_eventbridge_rules  = true

default_retention_days = 30
lambda_memory_size     = 128
lambda_log_level       = "INFO"

create_monitoring_dashboard = true
enable_cloudwatch_alarms   = true
