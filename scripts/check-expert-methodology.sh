#!/bin/bash
# Verify expert-pack methodology cards match the authoritative root skill.
#
# 设计决策：方法论卡唯一权威源在 skills/liki（各域 domains/）。本脚本校验
# 独立专家包（liki-bazi 等）的方法论卡与根一致，防止双份漂移。
# 修复：运行 scripts/sync-expert-methodology.sh 重新同步。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# <root 相对路径>:<expert_pack>:<expert_skill_dir>（与 sync 脚本一致）
MAPPINGS=(
    "natal/domains/bazi:liki-bazi:bazi"
    "natal/domains/ziwei:liki-ziwei:ziwei"
    "divination/domains/liuyao:liki-liuyao:liuyao"
    "divination/domains/qimen:liki-qimen:qimen"
    "divination/domains/huangli:liki-qimen:huangli"
    "fengshui/domains/bazhai:liki-fengshui:fengshui"
    "fengshui/domains/xuankong:liki-fengshui:fengshui"
    "naming/domains/qiming:liki-naming:naming"
    "naming/domains/bazi:liki-naming:naming"
)

fail=0
for entry in "${MAPPINGS[@]}"; do
    root_rel="${entry%%:*}"
    rest="${entry#*:}"
    pack="${rest%%:*}"
    skill="${rest#*:}"

    root_src="$ROOT/skills/liki/$root_rel"
    dest="$ROOT/skills/$pack/skills/$skill"

    # 根中的每张方法论卡（除 SKILL.md）必须在专家包中一致存在
    for card in "$root_src"/*.md; do
        [ -f "$card" ] || continue
        name="$(basename "$card")"
        [ "$name" = "SKILL.md" ] && continue
        if [ ! -f "$dest/$name" ]; then
            echo "❌ $pack/skills/$skill 缺 $name（根有）" >&2
            fail=1
        elif ! diff -q "$card" "$dest/$name" >/dev/null 2>&1; then
            echo "❌ $pack/skills/$skill/$name 与根不一致（运行 sync-expert-methodology.sh 修复）" >&2
            fail=1
        fi
    done
done

if [ "$fail" -ne 0 ]; then
    echo "专家包方法论卡与根权威源不一致，请运行 scripts/sync-expert-methodology.sh" >&2
    exit 1
fi
echo "✓ 全部专家包方法论卡与根 liki 一致。"