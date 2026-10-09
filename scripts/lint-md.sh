#!/usr/bin/env bash
# Run the pinned Markdown linter without coupling a development-only CLI
# vulnerability surface into every application dependency lockfile.
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

exec "${NPM:-npm}" exec --yes markdownlint-cli2@0.23.3 -- "$ROOT/skills/**/*.md"
