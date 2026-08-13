#!/usr/bin/env bash
# Liki 评测运行脚本（qwen_code 标准模式）
# 隔离原理：运行前把答案文件移出 skill 目录（agent 容器挂载不到）→ 评测 → 恢复 → 自动判分
# 用法：bash tests/run-qwen.sh [--parallelism N]
set -euo pipefail
cd "$(dirname "$0")/.."

PARALLELISM=16
[[ "${1:-}" == "--parallelism" ]] && PARALLELISM="${2:-16}"

# 模型 key（不硬编码）：优先用已 export 的 OPENAI_API_KEY，否则从 ~/.reasonix/.env 读 DEEPSEEK_API_KEY
if [ -z "${OPENAI_API_KEY:-}" ] && [ -f "$HOME/.reasonix/.env" ]; then
  export OPENAI_API_KEY="$(grep '^DEEPSEEK_API_KEY=' "$HOME/.reasonix/.env" | head -1 | cut -d= -f2- | tr -d '"')"
fi
if [ -z "${OPENAI_API_KEY:-}" ]; then
  echo "错误：未找到模型 key。请先 export OPENAI_API_KEY=sk-...（或配置 ~/.reasonix/.env 的 DEEPSEEK_API_KEY）" >&2
  exit 1
fi

# 答案文件（运行期间必须不在 skill 目录内）
ANSWER_FILES=(tests/answers.json tests/groups.json tests/cats.json)
STASH_DIR="$(mktemp -d /tmp/liki-answers.XXXXXX)"

restore_answers() {
  for f in "${ANSWER_FILES[@]}"; do
    if [ -f "$STASH_DIR/$(basename "$f")" ]; then
      mv "$STASH_DIR/$(basename "$f")" "$f" 2>/dev/null || true
    fi
  done
  rmdir "$STASH_DIR" 2>/dev/null || true
}
trap restore_answers EXIT

# 1. 移走答案（确保容器 agent 物理读不到）
for f in "${ANSWER_FILES[@]}"; do
  [ -f "$f" ] && mv "$f" "$STASH_DIR/"
done
echo "答案已移出 skill 目录（隔离生效）: ${ANSWER_FILES[*]}"

# 2. 评测（标准 skill-up run；key 经 sed 注入 yaml 的 environment.env——宿主 export 通道对
#    qwen_code 引擎不生效，environment.env 是容器级 env，qwen 必然读到；占位符防 key 入库）
RUN_YAML="$(mktemp .run-eval.XXXXXX.yaml)"
sed "s|\${OPENAI_API_KEY}|$OPENAI_API_KEY|g" tests/eval-grouped-qwen.yaml > "$RUN_YAML"
echo "key 已渲染进临时配置: $RUN_YAML"
trap 'rm -f "$RUN_YAML"; restore_answers' EXIT
skill-up run "$RUN_YAML" --parallelism "$PARALLELISM" --no-delete

# 3. 恢复答案（trap 兜底，这里显式再调一次）
restore_answers

# 4. 判分（取最新 iteration）
LATEST_ITER="$(ls -dt ../liki-workspace/iteration-* 2>/dev/null | head -1)"
if [ -n "$LATEST_ITER" ]; then
  echo "=== 判分: $LATEST_ITER ==="
  python3 tests/grade-grouped.py "$(realpath "$LATEST_ITER")"
else
  echo "未找到评测输出目录（../liki-workspace/iteration-*），跳过判分"
fi
