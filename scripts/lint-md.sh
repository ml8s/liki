#!/usr/bin/env bash
# Run the Markdown linter from the repository lockfile. Gate commands must not
# implicitly download tools: a missing local dependency fails fast instead.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

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

MARKDOWNLINT="$ROOT/node_modules/.bin/markdownlint-cli2"
if [ ! -x "$MARKDOWNLINT" ]; then
    echo "markdownlint-cli2 is not installed; run npm ci once, then retry" >&2
    exit 127
fi

# No explicit targets: use the globs declared in .markdownlint-cli2.jsonc
# (README.md, README.en.md, docs/USER_GUIDE*.md, skills/liki/**/*.md).
exec "$MARKDOWNLINT" "$@"
