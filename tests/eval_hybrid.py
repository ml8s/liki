#!/usr/bin/env python3
"""【规则表数据检查】（非判题——评测唯一走 skill-up agent）

角色（2026 去异化后定稿）：
- 本脚本只做【规则层数据检查】：对 160 题跑 skill 断语（evaluate_from_chart——
  排盘/因子/断语全调 skill API）——统计断语覆盖（各域命中数/零命中题）——
  验证规则表改动不崩、断语覆盖正常。
- 【不判题】——判题（题目→skill→答案→对比）唯一走 skill-up agent 评测
  （tests/run-qwen.sh：agent 读 SKILL.md → evaluate_from_chart → 综合判题 → grade-grouped 判分）。
- 无 _STATUS_MAP/族逻辑/紫微铁断（评测标签/判题逻辑已全部移出 skill 与工具脚本）。

用法：python3 tests/eval_hybrid.py
输出：tests/RESULTS.md（断语覆盖统计——非正确率）
"""
import json
import os
import re
import sys

_LOCAL = os.path.dirname(os.path.abspath(__file__))
if _LOCAL not in sys.path:
    sys.path.insert(0, _LOCAL)
_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
_TOOLS = os.path.join(_ROOT, "tools")
if _TOOLS not in sys.path:
    sys.path.insert(0, _TOOLS)

from birth import parse_birth
from client import full_panchang
from duanyu import evaluate_from_chart

BASE = _LOCAL
GROUPS = json.load(open(os.path.join(BASE, "groups.json"), encoding="utf-8"))


def load_case_birth(case_id: str) -> str:
    s = open(os.path.join(BASE, "cases-grouped", f"{case_id}.yaml"), encoding="utf-8").read()
    m = re.search(r"出生信息：([^\n]+)", s)
    return m.group(1) if m else ""


def main():
    total = 0
    zero = []                       # 全断语域零命中的题
    dom_hits = {}                   # 域 → 命中次数（覆盖）
    cache = {}
    for case_id, qids in sorted(GROUPS.items()):
        birth = load_case_birth(case_id)
        if not birth:
            continue
        if case_id not in cache:
            solar, gender, lon, corr = parse_birth(birth)
            chart = full_panchang(solar, gender, lon, correct=corr)
            cache[case_id] = evaluate_from_chart(chart)
        r = cache[case_id]
        for qid in qids:
            total += 1
            n = 0
            for rule, entry in r["domains"].items():
                hits = entry.get("八字", []) + entry.get("紫微", [])
                if hits:
                    dom_hits[rule] = dom_hits.get(rule, 0) + 1
                    n += len(hits)
            if n == 0:
                zero.append(qid)

    lines = []
    lines.append("# 规则表数据检查（非判题——评测唯一走 skill-up agent）")
    lines.append("")
    lines.append(f"- 总题数：{total}")
    lines.append(f"- 零命中（全断语域无覆盖）：{len(zero)}")
    lines.append(f"- 零命中题：{zero}")
    lines.append("")
    lines.append("## 各域断语覆盖（命中题数）")
    lines.append("")
    for rule, n in sorted(dom_hits.items(), key=lambda x: -x[1]):
        lines.append(f"- {rule}：{n} 题有断语")
    lines.append("")
    lines.append("> 判题（题目→skill→答案→对比）请跑 skill-up agent 评测：`bash tests/run-qwen.sh`")
    out = os.path.join(BASE, "RESULTS.md")
    open(out, "w", encoding="utf-8").write("\n".join(lines) + "\n")
    print("\n".join(lines[:6]))
    print(f"完整统计已写 {out}")


if __name__ == "__main__":
    main()
