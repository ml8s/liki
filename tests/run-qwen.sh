#!/usr/bin/env bash
# Liki 评测运行脚本（qwen_code 标准模式）
# 隔离原理：运行前把答案文件移出 skill 目录（agent 容器挂载不到）→ 评测 → 恢复 → 自动判分
# 用法：bash tests/run-qwen.sh [--parallelism N] [--resume]
set -euo pipefail
cd "$(dirname "$0")/.."

PARALLELISM=16
RESUME=0
for a in "$@"; do
  [[ "$a" == "--parallelism" ]] && { PARALLELISM="${2:-16}"; shift; }
  [[ "$a" == "--resume" ]] && RESUME=1
done

# 模型 key（不硬编码）：优先用已 export 的 OPENAI_API_KEY，否则从 liki-web/.env 或 ~/.reasonix/.env 读 DEEPSEEK_API_KEY
if [ -z "${OPENAI_API_KEY:-}" ] && [ -f "../liki-web/.env" ]; then
  export OPENAI_API_KEY="$(grep '^DEEPSEEK_API_KEY=' "../liki-web/.env" | head -1 | cut -d= -f2- | tr -d '"')"
fi
if [ -z "${OPENAI_API_KEY:-}" ] && [ -f "$HOME/.reasonix/.env" ]; then
  export OPENAI_API_KEY="$(grep '^DEEPSEEK_API_KEY=' "$HOME/.reasonix/.env" | head -1 | cut -d= -f2- | tr -d '"')"
fi
if [ -z "${OPENAI_API_KEY:-}" ]; then
  echo "错误：未找到模型 key。请先 export OPENAI_API_KEY=sk-...（或配置 ../liki-web/.env 的 DEEPSEEK_API_KEY）" >&2
  exit 1
fi

# 答案文件（运行期间必须不在 skill 目录内——stash 用固定路径 + 启动自愈：SIGKILL(137) 后残留自动恢复）
ANSWER_FILES=(tests/answers.json tests/groups.json tests/cats.json)
STASH_DIR="/tmp/liki-answers"

# 判分前答案自愈：答案缺失（/tmp 被清/stash 丢失）→ 从 git 恢复，防丢答案
restore_answers() {
  for f in "${ANSWER_FILES[@]}"; do
    if [ -f "$STASH_DIR/$(basename "$f")" ]; then
      mv "$STASH_DIR/$(basename "$f")" "$f" 2>/dev/null || true
    fi
  done
  rmdir "$STASH_DIR" 2>/dev/null || true
  for f in "${ANSWER_FILES[@]}"; do
    [ -f "$f" ] || git checkout -- "$f" 2>/dev/null || true
  done
}
trap restore_answers EXIT

# 自愈：上次评测中断（SIGKILL 等无法捕获）残留的 stash——启动即恢复
if [ -d "$STASH_DIR" ] && [ -f "$STASH_DIR/answers.json" ]; then
  echo "检测到上次评测残留 stash——自动恢复答案"
  restore_answers
fi

# 1. 移走答案（确保容器 agent 物理读不到）——stash 目录先建（防 /tmp 残留被清后 mv 失败）
mkdir -p "$STASH_DIR"
for f in "${ANSWER_FILES[@]}"; do
  [ -f "$f" ] && mv "$f" "$STASH_DIR/"
done
echo "答案已移出 skill 目录（隔离生效）: ${ANSWER_FILES[*]}"

# 2. 评测（标准 skill-up run；key 经 sed 注入 yaml 的 environment.env——宿主 export 通道对
#    qwen_code 引擎不生效，environment.env 是容器级 env，qwen 必然读到；占位符防 key 入库）
RUN_YAML="$(mktemp tests/evals/.run-eval.XXXXXX.yaml)"
sed "s|\${OPENAI_API_KEY}|$OPENAI_API_KEY|g" tests/evals/eval.yaml > "$RUN_YAML"
echo "key 已渲染进临时配置: $RUN_YAML"
trap 'rm -f "$RUN_YAML"; restore_answers' EXIT
RESUME_ARGS=()
if [ "$RESUME" = 1 ]; then
  # 断点续跑：找最新 iteration 已完成 case（有 response/stdout 且可判）——排除重跑
  LATEST_DONE="$(ls -dt tests/liki-workspace/iteration-* 2>/dev/null | head -1 || true)"
  if [ -n "$LATEST_DONE" ]; then
    DONE_GLOBS=()
    for d in "$LATEST_DONE"/pan*; do
      [ -d "$d" ] || continue
      if [ -f "$d/with_skill/outputs/response.md" ] || grep -q "题1" "$d/with_skill/outputs/agent/run/stdout.txt" 2>/dev/null; then
        DONE_GLOBS+=("$(basename "$d")")
      fi
    done
    if [ ${#DONE_GLOBS[@]} -gt 0 ]; then
      echo "断点续跑：跳过已完成 $(IFS=,; echo "${DONE_GLOBS[*]}")（共 ${#DONE_GLOBS[@]} case）"
      for g in "${DONE_GLOBS[@]}"; do RESUME_ARGS+=(--exclude-case-name "$g"); done
    fi
  fi
fi
skill-up run "$RUN_YAML" --parallelism "$PARALLELISM" --output-dir "$(pwd)/tests/liki-workspace" --no-delete "${RESUME_ARGS[@]}"

# 3. 恢复答案（trap 兜底，这里显式再调一次）
restore_answers

# 4. 判分（skill-up script judge 已随评测完成——答案正确率在报告的 grading rationale 里；
#    此处仅提示结果位置；--resume 时先合并旧 iteration 已完成 case 的输出）
LATEST_ITER="$(ls -dt tests/liki-workspace/iteration-* 2>/dev/null | head -1)"
if [ -n "$LATEST_ITER" ]; then
  if [ "$RESUME" = 1 ] && [ -n "${LATEST_DONE:-}" ] && [ "$LATEST_DONE" != "$LATEST_ITER" ]; then
    echo "合并旧 iteration 已完成 case 输出: $LATEST_DONE"
    for d in "$LATEST_DONE"/pan*; do
      [ -d "$d" ] || continue
      base="$(basename "$d")"
      if [ ! -e "$LATEST_ITER/$base/with_skill/outputs" ] || [ -z "$(ls -A "$LATEST_ITER/$base/with_skill/outputs" 2>/dev/null)" ]; then
        mkdir -p "$LATEST_ITER/$base/with_skill"
        cp -r "$d/with_skill/outputs" "$LATEST_ITER/$base/with_skill/" 2>/dev/null || true
      fi
    done
  fi
  echo "=== 评测完成: $LATEST_ITER ==="
  echo "（答案判分已由 skill-up script judge（tests/grade-case.py）随评测完成——每盘正确率/错题见 $LATEST_ITER/benchmark.md 与 result.json 的 grading 段）"
else
  echo "未找到评测输出目录（tests/liki-workspace/iteration-*），跳过判分"
fi
