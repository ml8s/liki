#!/usr/bin/env bash
# Install the locked Markdown linter and run it without floating npx resolution.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MDL="$ROOT/node_modules/.bin/markdownlint-cli2"

node_major() {
    "${NODE:-node}" --version | sed 's/^v\([0-9][0-9]*\).*/\1/'
}

if [ "$(node_major)" -lt 22 ]; then
    for candidate in "$HOME"/.nvm/versions/node/v22.*/bin; do
        [ -x "$candidate/node" ] || continue
        export PATH="$candidate:$PATH"
        break
    done
fi

if [ "$(node_major)" -lt 22 ]; then
    echo "markdownlint-cli2 0.23.x requires Node >= 22; current: $(${NODE:-node} --version)" >&2
    exit 1
fi

if [ ! -x "$MDL" ]; then
    (cd "$ROOT" && npm ci --ignore-scripts --no-audit --no-fund)
fi

exec "$MDL" "$ROOT/skills/**/*.md"
