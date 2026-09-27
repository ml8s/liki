#!/bin/bash
# Sync expert-pack methodology cards from the authoritative root skill (liki).
#
# 设计决策：方法论卡的唯一权威源是 skills/liki（各域 domains/ 下）。独立专家包
# （liki-bazi 等）是根 skill 的对外发布形态，其方法论卡由此脚本从根同步，避免
# 手维护双份造成漂移。优化 skill 只改根 liki，再运行本脚本同步专家包。
#
# 同步策略：只覆盖"根 domains 中存在的 .md 方法论卡"到专家包 skills/<skill>/，
# 不删除专家包中额外的文件（如编排卡副本），也不覆盖专家包自己的 SKILL.md。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# <root 相对路径>:<expert_pack>:<expert_skill_dir>
# 显式列出根中的方法论卡源（可能分散在不同域的 domains/ 下）
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

for entry in "${MAPPINGS[@]}"; do
    root_rel="${entry%%:*}"
    rest="${entry#*:}"
    pack="${rest%%:*}"
    skill="${rest#*:}"

    root_src="$ROOT/skills/liki/$root_rel"
    if [ ! -d "$root_src" ]; then
        echo "[sync-expert] ✗ 根中未找到方法论源：skills/liki/$root_rel" >&2
        exit 1
    fi

    dest="$ROOT/skills/$pack/skills/$skill"
    if [ ! -d "$dest" ]; then
        echo "[sync-expert] ✗ 目标专家包目录不存在：$dest" >&2
        exit 1
    fi

    # 只复制根中存在的 .md 方法论卡（覆盖专家包同名文件），保留 SKILL.md 与专家包额外文件
    # shellcheck disable=SC2086
    rsync -a \
        --exclude 'SKILL.md' \
        --include '*/' \
        --include '*.md' \
        --exclude '*' \
        "$root_src/" "$dest/"
    echo "[sync-expert] ✓ $root_rel → $pack/skills/$skill"
done

echo "[sync-expert] 全部专家包方法论卡已从根 liki 同步。"