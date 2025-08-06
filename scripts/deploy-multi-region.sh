#!/bin/bash

# Legacy wrapper script - redirects to new comprehensive deployment script
# This maintains backward compatibility

echo "⚠️  This script has been replaced by the comprehensive deployment script."
echo "📦 Using: scripts/logguardian-deploy.sh"
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NEW_SCRIPT="$SCRIPT_DIR/logguardian-deploy.sh"

# Convert old parameters to new format
ENVIRONMENT=${1:-sandbox}
shift

# Default regions if none provided
if [ $# -eq 0 ]; then
    REGIONS="ca-central-1,ca-west-1"
else
    # Convert space-separated to comma-separated
    REGIONS=$(IFS=','; echo "$*")
fi

echo "Converting legacy call:"
echo "  Old: deploy-multi-region.sh $ENVIRONMENT ${REGIONS//,/ }"
echo "  New: logguardian-deploy.sh deploy-multi -e $ENVIRONMENT -r $REGIONS"
echo ""

# Call new script
exec "$NEW_SCRIPT" deploy-multi -e "$ENVIRONMENT" -r "$REGIONS"
