#!/usr/bin/env bash
# 改表/改文档后自动验证：check_schema + check_docs
set -e
cd "$(dirname "$0")/.."

echo "=== check_schema（断语表质量：条件组/作用域/跨术数/互斥条件/生产纯度/必填/经典依据）==="
python3 scripts/check_schema.py

echo "=== check_docs（统一 Liki skill 文档契约）==="
python3 scripts/check_docs.py skills/liki

if [[ "${1:-}" == "--full" ]]; then
    if [[ -z "${LIKI_MCP_URL:-}" ]]; then
        echo "LIKI_MCP_URL is required for --full (start engine MCP first)" >&2
        exit 1
    fi
    echo "=== eval_hybrid（160 题规则覆盖 / 零命中检查）==="
    python3 scripts/eval_hybrid.py
fi

echo "✅ check 通过"
