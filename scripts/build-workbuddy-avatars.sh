#!/usr/bin/env bash
# Rasterize the per-domain expert-pack avatars from their SVG sources.
#
# 权威源是包内 `avatars/expert.svg`（矢量、按域区分），产物是同目录 512×512 的
# `avatars/expert.png`（WorkBuddy 专家头像规范：PNG、512×512、单张 ≤500KB）。
# 中文字形依赖 `AR PL UKai CN`（起名专家头像的田字格用字），缺失时直接失败，
# 避免静默换字形。
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
FONT="AR PL UKai CN"

if ! command -v inkscape >/dev/null 2>&1; then
    echo "[workbuddy-avatars] error: inkscape not found" >&2
    exit 1
fi
if ! fc-list :lang=zh family | grep -qF "$FONT"; then
    echo "[workbuddy-avatars] error: 缺少中文字体 $FONT（avatar 用字依赖它）" >&2
    exit 1
fi

found=0
for svg in "$ROOT"/expert-packs/liki-*/avatars/expert.svg; do
    [ -f "$svg" ] || continue
    found=1
    png="${svg%.svg}.png"
    inkscape "$svg" -w 512 -h 512 -o "$png"
    echo "  ✓ ${png#"$ROOT/"} ($(du -h "$png" | cut -f1))"
done

if [ "$found" -eq 0 ]; then
    echo "[workbuddy-avatars] error: no expert-packs/*/avatars/expert.svg found" >&2
    exit 1
fi
