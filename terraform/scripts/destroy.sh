#!/usr/bin/env bash
#
set -euo pipefail
# LogGuardian Destroy Script
#
# Safely removes all LogGuardian infrastructure
#

set -euo pipefail

readonly RED='\033[0;31m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m'

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly TF_DIR="${SCRIPT_DIR}/.."

log_warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

main() {
    cd "${TF_DIR}"
    
    echo ""
    log_warn "This will DESTROY all LogGuardian resources!"
    echo ""
    
    read -p "Type 'destroy' to confirm: " confirm
    [[ "${confirm}" == "destroy" ]] || log_error "Cancelled"
    
    terraform plan -destroy -out=destroy.tfplan
    
    echo ""
    read -p "Proceed with destruction? (yes/no): " final_confirm
    [[ "${final_confirm}" == "yes" ]] || log_error "Cancelled"
    
    terraform apply destroy.tfplan
    rm -f destroy.tfplan
    
    echo ""
    echo "Resources destroyed successfully"
}

main "$@"
