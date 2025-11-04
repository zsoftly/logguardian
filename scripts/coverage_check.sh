#!/usr/bin/env bash
set -eo pipefail
THRESH=${1:-90}
FILE=${2:-coverage.out}

if [ ! -f "$FILE" ]; then
  echo "Coverage file $FILE not found" >&2
  exit 1
fi

TOTAL=$(go tool cover -func="$FILE" | awk '/total:/ {print substr($3, 1, length($3)-1)}')
FLOOR=${TOTAL%.*}
if [ "$FLOOR" -lt "$THRESH" ]; then
  echo "Coverage $TOTAL% is below threshold ${THRESH}%" >&2
  exit 1
fi
echo "Coverage ${TOTAL}% ≥ ${THRESH}% ✔"
