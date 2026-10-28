#!/usr/bin/env bash
# Engine unit coverage floor: the single source of truth for both local
# `make test-engine-coverage` and the CI `engine-test` job. Keeping the floor
# here means a bump is one edit, not two that can drift apart.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
FLOOR="${COVERAGE_FLOOR:-85}"
PROFILE="${COVERAGE_PROFILE:-$ROOT/engine/coverage.out}"

"$ROOT/scripts/check-go-version.sh"

cd "$ROOT/engine"
go test -race -count=1 -short -coverprofile="$PROFILE" ./...

COV="$(go tool cover -func="$PROFILE" | tail -1 | awk '{print $3}' | sed 's/%//')"
echo "engine coverage: ${COV}% (floor ${FLOOR}%)"
if awk -v c="$COV" -v m="$FLOOR" 'BEGIN {exit !(c < m)}'; then
    echo "engine coverage ${COV}% is below the ${FLOOR}% floor" >&2
    exit 1
fi
