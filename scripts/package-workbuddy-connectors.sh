#!/usr/bin/env bash
# Package every WorkBuddy connector directory into a submission zip.
#
# 提审要求提交「连接器目录」本身：zip 的顶层必须是 <source>-connector/，内含
# connector-meta.json、mcp.json（或 cli.json）、icon.*。产物写进 dist/workbuddy/
# （dist/ 不入库），保持源码目录只有提交源。
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC_DIR="$ROOT/workbuddy"
OUT_DIR="$ROOT/dist/workbuddy"

if [ ! -d "$SRC_DIR" ]; then
    echo "[workbuddy-package] error: $SRC_DIR not found" >&2
    exit 1
fi

mkdir -p "$OUT_DIR"

found=0
for dir in "$SRC_DIR"/*-connector; do
    [ -d "$dir" ] || continue
    found=1
    name="$(basename "$dir")"
    out="$OUT_DIR/$name.zip"
    rm -f "$out"
    (cd "$SRC_DIR" && zip -q -X -r "$out" "$name" -x '*.DS_Store' -x '__MACOSX/*')
    echo "  ✓ ${out#"$ROOT/"} ($(du -h "$out" | cut -f1))"
done

if [ "$found" -eq 0 ]; then
    echo "[workbuddy-package] error: no *-connector directory under $SRC_DIR" >&2
    exit 1
fi
