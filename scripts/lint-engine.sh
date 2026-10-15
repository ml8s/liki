#!/usr/bin/env bash
# Run the pinned golangci-lint, then go vet, over the engine module.
#
# The version is pinned in deps/golangci-lint.version and shared with CI's
# `engine-lint` job: a PATH binary of another version would make local lint and
# CI disagree, so a mismatch is an error rather than a silent fallback.
# Install with golangci-lint's own installer (CONTRIBUTING.md): its binaries are
# built against a specific Go toolchain, which `go install` does not reproduce.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PIN_FILE="$ROOT/deps/golangci-lint.version"
WANT="$(sed -n 's/^version:[[:space:]]*//p' "$PIN_FILE" | head -1)"
if [ -z "$WANT" ]; then
    echo "golangci-lint pin missing in $PIN_FILE" >&2
    exit 1
fi

INSTALL_HINT="  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b \"\$(go env GOPATH)/bin\" $WANT"

if ! command -v golangci-lint >/dev/null 2>&1; then
    echo "golangci-lint not found; install the pinned version:" >&2
    echo "$INSTALL_HINT" >&2
    exit 1
fi

HAVE="$(golangci-lint version 2>/dev/null | sed -n 's/.*version[[:space:]]\{1,\}\(v\?[0-9][^ ]*\).*/\1/p' | head -1)"
if [ "$HAVE" != "$WANT" ] && [ "${HAVE#v}" != "${WANT#v}" ]; then
    echo "golangci-lint $HAVE does not match the pinned $WANT" >&2
    echo "install the pinned version:" >&2
    echo "$INSTALL_HINT" >&2
    exit 1
fi

cd "$ROOT/engine"
golangci-lint run ./...
go vet ./...
