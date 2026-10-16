#!/usr/bin/env bash
# 本地 node/python 必须等于声明版本：CI 从 .nvmrc/.python-version 安装，
# 本地不符 = 本地跑出的结果 CI 无法复现。与 go-version-check 同一原则。
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
fail=0

want_node="$(cat "$ROOT/.nvmrc")"
if ! command -v node >/dev/null 2>&1; then
    echo "node $want_node is required; install it and put it on PATH (nvm install $want_node && nvm use)." >&2
    fail=1
else
    have_node="$(node -v | sed 's/^v//')"
    if [ "$have_node" != "$want_node" ]; then
        echo "node $want_node is required; found $have_node at $(command -v node). Run: nvm install $want_node && nvm use." >&2
        fail=1
    fi
fi

want_py="$(cat "$ROOT/.python-version")"
if ! command -v python3 >/dev/null 2>&1; then
    echo "python $want_py is required; see .python-version." >&2
    fail=1
else
    have_py="$(python3 -V | sed 's/^Python //')"
    if [ "$have_py" != "$want_py" ]; then
        echo "python $want_py is required; found $have_py at $(command -v python3)." >&2
        echo "  CI installs $want_py from .python-version; point python3 at a $want_py interpreter." >&2
        fail=1
    fi
fi

exit "$fail"
