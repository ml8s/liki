#!/usr/bin/env python3
"""【Mingli160 校准 / 评估报告】

角色：
- 对 160 个命盘样本执行本地规则链：出生解析 → 排盘 → 因子生成 → 断语查询，
  统计各域覆盖与零命中题，供人工校准 skill 或比较改版前后表现。
- 【不判题】——本脚本不输出准确率，也不消费模型结论。
- 不包含评测标签、族逻辑或紫微铁断；它是手动评估入口，不属于测试门禁。

用法：python3 scripts/eval_hybrid.py
输出：stdout（断语覆盖统计——非正确率）；需要留档时显式传 --output FILE
"""
import argparse
import json
import os
import sys
from pathlib import Path

_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
# Rule-engine tools live with the unified Liki bazi domain.
_COUNSEL = os.path.join(_ROOT, "counsel")
_LOCAL = os.path.dirname(os.path.abspath(__file__))   # scripts/birth.py 在此
_BENCHMARK = os.path.join(_ROOT, "evals", "mingli160")
for _p in (_COUNSEL, _LOCAL):
    if _p not in sys.path:
        sys.path.insert(0, _p)

from app.natal.tools.charts import build_natal_chart, liunian
from app.natal.tools.duanyu import NATAL_RULES, YEARLY_RULES, match_rule, query_yearly, resolve_current_year
from app.natal.tools.factors import evaluate_liunian_snap_from_natal_chart, evaluate_snap_from_natal_chart
from birth import parse_birth

BENCHMARK = _BENCHMARK
with open(os.path.join(BENCHMARK, "groups.json"), encoding="utf-8") as _file:
    GROUPS = json.load(_file)
with open(os.path.join(BENCHMARK, "births.json"), encoding="utf-8") as _file:
    BIRTHS = json.load(_file)


def query_all(natal_chart: dict) -> dict:
    """断语查询（数据检查用）——本命域查本命快照，流年域用当前年流年盘采样。"""
    snap = evaluate_snap_from_natal_chart(natal_chart)
    domains = {}
    # 本命域
    for rule in sorted(NATAL_RULES):
        domains[rule] = match_rule(rule, snap)
    # 流年域——用当前年采样（完整流年覆盖需多年扫描，此处仅验证规则表不崩/有产出）
    cur_year, _ = resolve_current_year()
    lnp = liunian(natal_chart, cur_year)
    snap_y = evaluate_liunian_snap_from_natal_chart(natal_chart, lnp, year=cur_year)
    for rule in sorted(YEARLY_RULES):
        domains[rule] = query_yearly(rule, snap_y)
    return {"domains": domains}


def load_case_birth(case_id: str) -> str:
    return BIRTHS[case_id]


def build_report(total: int, zero: list[str], dom_hits: dict[str, int]) -> str:
    lines = []
    lines.append("# Mingli160 校准 / 评估报告（非判题）")
    lines.append("")
    lines.append(f"- 总题数：{total}")
    lines.append(f"- 零命中（全断语域无覆盖）：{len(zero)}")
    lines.append(f"- 零命中题：{zero}")
    lines.append("")
    lines.append("## 各域断语覆盖（命中题数）")
    lines.append("")
    for rule, count in sorted(dom_hits.items(), key=lambda item: -item[1]):
        lines.append(f"- {rule}：{count} 题有断语")
    lines.append("")
    lines.append("> 本报告只度量断语覆盖，供人工校准参考，不作为准确率评测。")
    return "\n".join(lines) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(description="规则表断语覆盖数据检查")
    parser.add_argument("--output", type=Path, help="可选报告文件；缺省打印到 stdout")
    args = parser.parse_args()

    total = 0
    zero = []                       # 全断语域零命中的题
    dom_hits = {}                   # 域 → 命中次数（覆盖）
    cache = {}
    for case_id, qids in sorted(GROUPS.items()):
        birth = load_case_birth(case_id)
        if not birth:
            continue
        if case_id not in cache:
            civil_time, gender, lon, _corr = parse_birth(birth)
            # 经度未知时用北京默认（与产品默认一致）；题干时间即 Gregorian time事实。
            natal_chart = build_natal_chart(
                civil_time[:19], "Asia/Shanghai", gender,
                lon if lon is not None else 116.4,
            )
            cache[case_id] = query_all(natal_chart)
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

    report = build_report(total, zero, dom_hits)
    if args.output is None:
        print(report, end="")
    else:
        args.output.write_text(report, encoding="utf-8")
        print(f"完整统计已写 {args.output}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
