#!/usr/bin/env bash
# 改表/改码后自动验证：schema 校验（快） + 数据检查（慢——--full）
# 用法：bash tests/check.sh [--full]
set -e
cd "$(dirname "$0")/.."

echo "=== check_schema（断语表质量：因子引用/跨术数/结论标签/必填/经典原文）==="
python3 tests/check_schema.py

if [ "${1:-}" = "--full" ]; then
  echo "=== 数据检查（eval_hybrid——160 题断语覆盖/零命中）==="
  python3 tests/eval_hybrid.py
fi
echo "✅ check 通过"
