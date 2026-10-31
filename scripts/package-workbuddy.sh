#!/usr/bin/env bash
# Package every WorkBuddy submission source into a zip under dist/workbuddy/.
#
# 官网只收 zip，且 zip 顶层必须是提交单元目录本身：
#   连接器  workbuddy/<source>-connector/  → <source>-connector.zip
#           （内含 connector-meta.json、mcp.json 或 cli.json、icon.*）
#   技能    skills/liki/                   → liki.zip
#           （内含 SKILL.md + references/、scripts/ 等子资源）
#   专家    expert-packs/liki/             → liki-expert.zip
#           （内含 .codebuddy-plugin/、agents/、avatars/、skills/、README.md）
#
# 产物写进 dist/workbuddy/（dist/ 不入库），源码目录只保留提交源。
# 专家包的包根 SKILL.md 是仓库里的身份手写源，平台侧入口是 agents/<pack>.md，故不随包提交。
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC_DIR="$ROOT/workbuddy"
SKILL_DIR="$ROOT/skills/liki"
PACKS_DIR="$ROOT/expert-packs"
OUT_DIR="$ROOT/dist/workbuddy"
# 只入包提交源：系统垃圾与本地运行残渣（字节码、虚拟环境、缓存）一律排除。
EXCLUDES=(-x '*.DS_Store' -x '__MACOSX/*' -x '*/__pycache__/*' -x '*.pyc' -x '*.pyo'
          -x '*/.venv/*' -x '*/.pytest_cache/*' -x '*/.ruff_cache/*' -x '*.egg-info/*')

if [ ! -d "$SRC_DIR" ] || [ ! -d "$PACKS_DIR" ]; then
    echo "[workbuddy-package] error: $SRC_DIR 或 $PACKS_DIR 不存在" >&2
    exit 1
fi

mkdir -p "$OUT_DIR"
found=0

for dir in "$SRC_DIR"/*-connector; do
    [ -d "$dir" ] || continue
    found=$((found + 1))
    name="$(basename "$dir")"
    out="$OUT_DIR/$name.zip"
    rm -f "$out"
    (cd "$SRC_DIR" && zip -q -X -r "$out" "$name" "${EXCLUDES[@]}")
    echo "  ✓ ${out#"$ROOT/"} ($(du -h "$out" | cut -f1))"
done

if [ -f "$SKILL_DIR/SKILL.md" ]; then
    found=$((found + 1))
    out="$OUT_DIR/liki.zip"
    rm -f "$out"
    (cd "$(dirname "$SKILL_DIR")" && zip -q -X -r "$out" "$(basename "$SKILL_DIR")" "${EXCLUDES[@]}")
    echo "  ✓ ${out#"$ROOT/"} ($(du -h "$out" | cut -f1))"
fi

if [ -d "$PACKS_DIR/liki" ]; then
    found=$((found + 1))
    out="$OUT_DIR/liki-expert.zip"
    rm -f "$out"
    (cd "$PACKS_DIR" && zip -q -X -r "$out" liki "${EXCLUDES[@]}" -x 'liki/SKILL.md')
    echo "  ✓ ${out#"$ROOT/"} ($(du -h "$out" | cut -f1))"
fi

if [ "$found" -eq 0 ]; then
    echo "[workbuddy-package] error: 没找到任何提交源（连接器 / 技能 / 专家包）" >&2
    exit 1
fi

python3 "$ROOT/scripts/check_workbuddy_zips.py" "$OUT_DIR"
