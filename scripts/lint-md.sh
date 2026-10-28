#!/usr/bin/env bash
# Run the Markdown linter from the repository lockfile. Gate commands must not
# implicitly download tools: a missing local dependency fails fast instead.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

NODE_BIN="${NODE:-$(command -v node 2>/dev/null || true)}"
if [ -z "$NODE_BIN" ]; then
    for candidate in "$HOME"/.nvm/versions/node/v22.*/bin/node; do
        [ -x "$candidate" ] || continue
        NODE_BIN="$candidate"
        break
    done
fi

node_major() {
    "$NODE_BIN" --version | sed 's/^v\([0-9][0-9]*\).*/\1/'
}

if [ -z "$NODE_BIN" ] || [ ! -x "$NODE_BIN" ] || [ "$(node_major)" -lt 22 ]; then
    for candidate in "$HOME"/.nvm/versions/node/v22.*/bin/node; do
        [ -x "$candidate" ] || continue
        NODE_BIN="$candidate"
        break
    done
fi

if [ -z "$NODE_BIN" ] || [ ! -x "$NODE_BIN" ] || [ "$(node_major)" -lt 22 ]; then
    echo "markdownlint-cli2 0.23.x requires Node >= 22" >&2
    exit 1
fi

export PATH="$(dirname "$NODE_BIN"):$PATH"

MARKDOWNLINT="$ROOT/node_modules/.bin/markdownlint-cli2"
if [ ! -x "$MARKDOWNLINT" ]; then
    echo "markdownlint-cli2 is not installed; run npm ci once, then retry" >&2
    exit 127
fi

# No explicit targets: use the globs declared in .markdownlint-cli2.jsonc
# (README.md, README.en.md, docs/USER_GUIDE*.md, skills/liki/**/*.md).
exec "$MARKDOWNLINT" "$@"
