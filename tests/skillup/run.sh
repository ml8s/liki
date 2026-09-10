#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."

DOMAINS=(bazi divination fengshui naming)
SELECTED=()
VALIDATE=0
PARALLELISM="${SKILLUP_PARALLELISM:-4}"
while [ "$#" -gt 0 ]; do
  case "$1" in
    --validate)
      VALIDATE=1
      ;;
    --parallelism)
      PARALLELISM="${2:?--parallelism requires a value}"
      shift
      ;;
    bazi|divination|fengshui|naming)
      SELECTED+=("$1")
      ;;
    *)
      echo "用法: $0 [--validate] [--parallelism N] [bazi|divination|fengshui|naming ...]" >&2
      exit 2
      ;;
  esac
  shift
done
if [ "${#SELECTED[@]}" -eq 0 ]; then
  SELECTED=("${DOMAINS[@]}")
fi

if ! command -v skill-up >/dev/null 2>&1; then
  echo "错误：未找到 skill-up 命令" >&2
  exit 1
fi

LOCAL_MODEL_ENV="${LOCAL_MODEL_ENV:-tests/skillup/evals/.local.env}"
if [ -f "$LOCAL_MODEL_ENV" ]; then
  set -a
  # shellcheck disable=SC1090
  source "$LOCAL_MODEL_ENV"
  set +a
fi
KEY_SOURCE=openai
if [ -n "${ZHIPU_API_KEY:-}" ] || [ -n "${ZHIPUAI_API_KEY:-}" ]; then
  export OPENAI_API_KEY="${ZHIPU_API_KEY:-${ZHIPUAI_API_KEY:-}}"
  KEY_SOURCE=zhipu
elif [ -z "${OPENAI_API_KEY:-}" ] && [ -f "../liki-web/.env" ]; then
  export OPENAI_API_KEY="$(grep '^DEEPSEEK_API_KEY=' "../liki-web/.env" | head -1 | cut -d'=' -f2- | tr -d '"')"
  KEY_SOURCE=deepseek
fi
if [ "$VALIDATE" -eq 1 ]; then
  export OPENAI_API_KEY="${OPENAI_API_KEY:-validate-only}"
  export OPENAI_BASE_URL="${OPENAI_BASE_URL:-http://localhost}"
  export OPENAI_MODEL="${OPENAI_MODEL:-validate-only}"
elif [ -z "${OPENAI_API_KEY:-}" ]; then
  echo "错误：未找到模型 key。可 export OPENAI_API_KEY=... 或 ZHIPU_API_KEY=..." >&2
  exit 1
fi
case "$KEY_SOURCE" in
  zhipu)
    export OPENAI_BASE_URL="${OPENAI_BASE_URL:-https://open.bigmodel.cn/api/paas/v4}"
    export OPENAI_MODEL="${OPENAI_MODEL:-glm-5.3-flash}"
    ;;
  deepseek)
    export OPENAI_BASE_URL="${OPENAI_BASE_URL:-https://api.deepseek.com/v1}"
    export OPENAI_MODEL="${OPENAI_MODEL:-deepseek-v4-flash}"
    ;;
esac

LIKI_RPC_MODE=docker
source scripts/local-engine.sh
cleanup() {
  rm -f "${RENDERED_YAMLS[@]:-}"
  stop_local_engine
}
trap cleanup EXIT

RENDERED_YAMLS=()
for domain in "${SELECTED[@]}"; do
  rendered="$(mktemp "tests/skillup/evals/.run-${domain}.XXXXXX.yaml")"
  RENDERED_YAMLS+=("$rendered")
  sed -e "s|\${OPENAI_API_KEY}|${OPENAI_API_KEY}|g" \
      -e "s|\${OPENAI_BASE_URL}|${OPENAI_BASE_URL:-https://api.openai.com/v1}|g" \
      -e "s|\${OPENAI_MODEL}|${OPENAI_MODEL:-glm-5.3-flash}|g" \
      -e "s|\${LIKI_RPC_URL}|pending-local-engine|g" \
      "tests/skillup/evals/${domain}.yaml" > "$rendered"
  echo "=== validate ${domain} ==="
  skill-up validate "$rendered"
done

if [ "$VALIDATE" -eq 1 ]; then
  echo "✓ skill-up smoke 配置校验通过"
  exit 0
fi

ensure_local_engine
for i in "${!SELECTED[@]}"; do
  domain="${SELECTED[$i]}"
  rendered="${RENDERED_YAMLS[$i]}"
  sed -i "s|pending-local-engine|$LOCAL_RPC|g" "$rendered"
  echo "=== skill-up smoke: ${domain} ==="
  skill-up run "$rendered" \
    --parallelism "$PARALLELISM" \
    --output-dir "$(pwd)/tests/skillup/workspace/${domain}" \
    --no-delete
done
