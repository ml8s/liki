#!/usr/bin/env bash
# Reject Go toolchains older than the security-supported engine toolchain.
set -eo pipefail

REQUIRED="${LIKI_REQUIRED_GO_VERSION:-1.26.6}"
command -v go >/dev/null 2>&1 || {
    echo "Go >= ${REQUIRED} is required (install it and put it on PATH)." >&2
    exit 1
}

ACTUAL="$(go env GOVERSION | sed 's/^go//')"
LOWEST="$(printf '%s\n%s\n' "$REQUIRED" "$ACTUAL" | sort -V | head -1)"
if [ "$LOWEST" != "$REQUIRED" ]; then
    echo "Go >= ${REQUIRED} is required; found ${ACTUAL} at $(command -v go)." >&2
    exit 1
fi
