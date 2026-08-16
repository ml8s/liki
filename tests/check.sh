#!/usr/bin/env bash
# 改表/改码后自动验证：
#   默认（快）  ：check_schema（断语表质量）+ check_docs（文档引用契约）+ 版本一致性
#   --full（慢）：另加 eval_hybrid（160 题断语覆盖/零命中——32 命例全量排盘，数分钟级）
# 用法：bash tests/check.sh [--full]
set -e
cd "$(dirname "$0")/.."

echo "=== check_schema（断语表质量：因子引用/跨术数/死列/结论标签/必填/经典原文）==="
python3 tests/check_schema.py

echo "=== check_docs（文档契约：断语 id / 文件路径 / 方法名 / RPC 调用 引用可解析——4 skill）==="
for SK in liki-bazi liki-divination liki-fengshui liki-naming; do
  python3 tests/check_docs.py "skills/$SK"
done

echo "=== 版本一致性（skill-tools.json info.version == VERSION——单一事实源）==="
python3 - <<'PYEOF'
import json, os
p = "skills/liki-bazi/tools/skill-tools.json"
if not os.path.exists(p):
    print("  （无 skill-tools.json——子流程 skill 跳过）")
else:
    version = open("skills/liki-bazi/VERSION", encoding="utf-8").read().strip()
    info = json.load(open(p, encoding="utf-8")).get("info", {})
    got = info.get("version")
    if got != version:
        print(f"  ✗ info.version={got} ≠ VERSION={version}——跑 make build-archive 自动注入")
        exit(1)
    print(f"  ✓ info.version == VERSION == {version}")
PYEOF

if [ "${1:-}" = "--full" ]; then
  echo "=== 数据检查（eval_hybrid——160 题断语覆盖/零命中，数分钟级）==="
  python3 tests/eval_hybrid.py
fi
echo "✅ check 通过"
