#!/usr/bin/env python3
"""【规则表数据检查】（非判题——评测唯一走 skill-up agent）

角色（2026 去异化后定稿）：
- 本脚本只做【规则层数据检查】：对 160 题跑 skill 断语（排盘(client) → 因子生成(evaluate_factors) → 断语查询(match)——全调 skill API）——统计断语覆盖（各域命中数/零命中题）——
  验证规则表改动不崩、断语覆盖正常。
- 【不判题】——判题（题目→skill→答案→对比）唯一走 skill-up agent 评测
  （tests/run-qwen.sh：agent 读 SKILL.md → 排盘(RPC)+因子生成+断语查询 → 综合判题 → grade-grouped 判分）。
- 无 _STATUS_MAP/族逻辑/紫微铁断（评测标签/判题逻辑已全部移出 skill 与工具脚本）。

用法：python3 tests/eval_hybrid.py
输出：tests/RESULTS.md（断语覆盖统计——非正确率）
"""
import json
import os
import re
import sys


_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
# 4.1.0 重构：tools 移至 skills/liki-bazi/tools（原 _ROOT/tools 已不存在）
_TOOLS = os.path.join(_ROOT, "skills", "liki-bazi", "tools")
_LOCAL = os.path.dirname(os.path.abspath(__file__))   # tests/（client/birth 排盘工具在此）
for _p in (_TOOLS, _LOCAL):
    if _p not in sys.path:
        sys.path.insert(0, _p)

from birth import parse_birth
from paipan import full_paipan
from duanyu import make_factors, query, query_yearly, _NATAL_RULES, _YEARLY_RULES

BASE = os.path.dirname(os.path.abspath(__file__))
GROUPS = json.load(open(os.path.join(BASE, "groups.json"), encoding="utf-8"))


def query_all(pan: dict) -> dict:
    """断语查询（数据检查用）——本命域查本命快照，流年域用当前年流年盘采样。"""
    snap = make_factors(pan)
    domains = {}
    # 本命域
    for rule in sorted(_NATAL_RULES):
        domains[rule] = query(rule, pan)
    # 流年域——用当前年采样（完整流年覆盖需多年扫描，此处仅验证规则表不崩/有产出）
    from datetime import datetime
    cur_year = datetime.now().year
    from paipan import liunian
    from duanyu import make_liunian_factors
    lnp = liunian(pan, cur_year)
    for rule in sorted(_YEARLY_RULES - {"yingqi"}):
        target = "配偶星"
        snap_y = make_liunian_factors(pan, lnp, target=target, year=cur_year)
        domains[rule] = query_yearly(rule, snap_y)
    # yingqi 同时有本命表和流年表——流年版
    domains["yingqi"] = query_yearly("yingqi", make_liunian_factors(pan, lnp, year=cur_year))
    return {"domains": domains}


def load_case_birth(case_id: str) -> str:
    s = open(os.path.join(BASE, "evals/cases", f"{case_id}.yaml"), encoding="utf-8").read()
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
            # longitude 未知时降级为 correct=False（无经度无法校正——宁可不校正也不用错误经度）
            eff_corr = corr and lon is not None
            pan = full_paipan(solar, gender, lon, correct=eff_corr)
            cache[case_id] = query_all(pan)
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
