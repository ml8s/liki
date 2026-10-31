#!/usr/bin/env bash
# Rasterize the per-domain expert-pack avatars from their SVG sources.
#
# 权威源是包内 `avatars/expert.svg`（矢量、按域区分），产物是同目录 512×512 的
# `avatars/expert.png`（WorkBuddy 专家头像规范：PNG、512×512、单张 ≤500KB）。
# 市场宿主的字体与 `dominant-baseline` 支持都不可控，因此源里只允许路径字形；
# 含 `<text>`/`font-family`/`dominant-baseline` 的源在这里直接失败。
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

if ! command -v inkscape >/dev/null 2>&1; then
    echo "[workbuddy-avatars] error: inkscape not found" >&2
    exit 1
fi

svg="$ROOT/expert-packs/liki/avatars/expert.svg"
if [ ! -f "$svg" ]; then
    echo "[workbuddy-avatars] error: expert-packs/liki/avatars/expert.svg not found" >&2
    exit 1
fi
if grep -qE '<text|font-family|dominant-baseline' "$svg"; then
    echo "[workbuddy-avatars] error: ${svg#"$ROOT/"} 含文本/字体依赖，字形须先转路径" >&2
    exit 1
fi
png="${svg%.svg}.png"
inkscape "$svg" -w 512 -h 512 -o "$png"
echo "  ✓ ${png#"$ROOT/"} ($(du -h "$png" | cut -f1))"
