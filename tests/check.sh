#!/usr/bin/env bash
# 改表/改码后自动验证：
#   默认（快）  ：check_schema + check_docs + unified skill 结构/版本契约
#   --full（慢）：另加 eval_hybrid（160 题断语覆盖/零命中）
set -e
cd "$(dirname "$0")/.."

echo "=== check_schema（断语表质量：条件组/作用域/跨术数/互斥条件/生产纯度/必填/经典依据）==="
python3 tests/check_schema.py

echo "=== check_docs（统一 Liki skill 文档契约）==="
python3 tests/check_docs.py skills/liki

echo "=== tool result contracts（LLM-facing schema 防漂移）==="
python3 scripts/generate_bazi_tool_contracts.py --check

echo "=== unified skill structure & version consistency ===="
python3 - <<'PYEOF'
import json
from pathlib import Path
root = Path('skills/liki')
version = (root / 'VERSION.txt').read_text(encoding='utf-8').strip()
skills = list(root.rglob('SKILL.md'))
entries = [root / d / 'ENTRY.md' for d in ('bazi', 'divination', 'fengshui', 'naming')]
assert [p.relative_to(root) for p in skills] == [Path('SKILL.md')], skills
assert all(p.exists() for p in entries), entries
assert len(list(root.rglob('VERSION.txt'))) == 1
assert not list(root.rglob('VERSION'))
assert len(list(root.rglob('feedback.py'))) == 1
assert len(list(root.rglob('feedback.schema.json'))) == 1
manifests = sorted(root.glob('*/tools/skill-tools.json'))
for p in manifests:
    got = json.loads(p.read_text(encoding='utf-8')).get('info', {}).get('version')
    assert got == version, f'{p}: {got} != {version}'
print(f'  ✓ unique SKILL.md + 4 ENTRY.md')
print(f'  ✓ {len(manifests)} domain manifests == VERSION.txt == {version}')
PYEOF

if [ "${1:-}" = "--full" ]; then
  echo "=== 数据检查（eval_hybrid——160 题断语覆盖/零命中，输出 stdout）==="
  python3 tests/eval_hybrid.py
fi
echo "✅ check 通过"
