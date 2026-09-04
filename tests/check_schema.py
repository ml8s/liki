"""断言库 schema 校验（长表、条件组、术数/作用域与生产数据纯度）。"""
from __future__ import annotations

import csv
import json
import os
import re
import sys
from collections import defaultdict

_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, os.path.join(_ROOT, "skills", "liki-bazi", "tools"))

from factor_tables import load_long_rows

TOOLS = os.path.join(_ROOT, "skills", "liki-bazi", "tools")
ASSERTIONS_PATH = os.path.join(TOOLS, "assertions", "assertions.csv")
CONDITIONS_PATH = os.path.join(TOOLS, "assertions", "assertion_conditions.csv")
CONTEXT_KEYS = {"性别"}

PROCESS_RE = re.compile(
    r"(mingli-skills|phase\d+|step\d+|硬标注|fangfa/|app/|domains/|[A-Za-z_-]+\.md|factors\.csv|"
    r"dayun\.md|hehui\.md|(?<!\d)0\d{3}(?!\d)|\bpan\d+\b|\b[a-z]+(?:_[a-z0-9]+)+\b|MingLi(?:-| )?Bench|benchmark|考时|回填|选项|题目|评测|合并原|根因|用户\s*\d{4}|"
    r"\b[MFSH]\d+[A-Za-z]?\b)", re.I
)


def _load_csv(path: str) -> tuple[list[str], list[dict]]:
    with open(path, encoding="utf-8-sig", newline="") as fh:
        reader = csv.DictReader(fh)
        return list(reader.fieldnames or []), list(reader)


def _factor_rows(filename: str) -> list[dict]:
    return load_long_rows(os.path.join(TOOLS, "factors", filename))


def _sides_by_factor(rows: list[dict]) -> dict[str, set[str]]:
    out: dict[str, set[str]] = defaultdict(set)
    for row in rows:
        out[row["因子"]].add((row.get("术数") or "bazi").strip())
    return out


def main() -> int:
    constants = json.load(open(os.path.join(TOOLS, "constants.json"), encoding="utf-8"))
    natal_rows = _factor_rows("factors.csv")
    flow_rows = _factor_rows("factors_liunian.csv")
    natal_sides = _sides_by_factor(natal_rows)
    flow_sides = _sides_by_factor(flow_rows)
    flow_trigger_factors = {
        row["因子"] for row in flow_rows
        if row.get("直通")
        or any(not key.startswith("引用本命[") for key in row.get("conds", {}))
    }

    direct_string: dict[str, bool] = defaultdict(bool)
    for filename in ("factors.csv", "factors_liunian.csv"):
        for row in _factor_rows(filename):
            direct_string[row["因子"]] = bool("任意" in (row.get("直通") or ""))

    assertion_fields, assertions = _load_csv(ASSERTIONS_PATH)
    condition_fields, conditions = _load_csv(CONDITIONS_PATH)
    expected_assertion_fields = [
        "assertion_id", "rule", "side", "领域", "事件类型", "时间层",
        "事件", "结论", "依据", "经典依据",
    ]
    expected_condition_fields = [
        "assertion_id", "condition_group_id", "factor", "expected"
    ]
    errors: list[str] = []
    warnings: list[str] = []

    if assertion_fields != expected_assertion_fields:
        errors.append(f"assertions.csv 表头不符: {assertion_fields}")
    if condition_fields != expected_condition_fields:
        errors.append(f"assertion_conditions.csv 表头不符: {condition_fields}")

    assertion_by_id = {row["assertion_id"]: row for row in assertions}
    if len(assertion_by_id) != len(assertions):
        errors.append("assertion_id 重复")

    groups: dict[tuple[str, str], dict[str, str]] = defaultdict(dict)
    for number, row in enumerate(conditions, 1):
        aid = row["assertion_id"]
        gid = row["condition_group_id"]
        factor = row["factor"]
        expected = row["expected"]
        if aid not in assertion_by_id:
            errors.append(f"条件第 {number} 行引用未知断语: {aid}")
            continue
        if not gid.isdigit() or int(gid) <= 0:
            errors.append(f"[{aid}] condition_group_id 无效: {gid}")
            continue
        if not factor or not expected:
            errors.append(f"[{aid}/{gid}] factor/expected 不能为空")
            continue
        if PROCESS_RE.search(f"{factor} {expected}"):
            errors.append(f"[{aid}/{gid}] 条件含评测/内部过程残留")
        key = (aid, gid)
        if key in groups and factor in groups[key]:
            errors.append(f"[{aid}/{gid}] 条件因子重复: {factor}")
        groups[key][factor] = expected

    # Rebuild evaluator-shaped rows and enforce scope/side reachability.
    grouped_assertions: dict[tuple[str, str], list[dict]] = defaultdict(list)
    for assertion in assertions:
        grouped_assertions[(assertion["side"], assertion["rule"])].append(assertion)

        aid = assertion["assertion_id"]
        side = assertion["side"]
        rule = assertion["rule"]
        if assertion["领域"] not in {
            "婚姻", "家庭", "子女", "学业", "事业", "财运", "健康", "官非",
            "灾劫", "性格", "精神", "外貌", "人际", "迁移", "房产", "出身",
            "玄学", "大限", "应期", "结构", "综合",
        }:
            errors.append(f"[{aid}] 领域不在受控闭集: {assertion['领域']}")
        if assertion["事件类型"] not in {
            "状态", "关系", "结构", "凶象", "吉象", "引动", "变动", "取象",
        }:
            errors.append(f"[{aid}] 事件类型不在受控闭集: {assertion['事件类型']}")
        expected_layer = "流年" if rule.startswith("年") else "大限" if rule == "大限" else "本命"
        if assertion["时间层"] != expected_layer:
            errors.append(
                f"[{aid}] 时间层 {assertion['时间层']} 与 rule {rule} 不符"
            )
        text_fields = ("领域", "事件类型", "时间层", "事件", "结论", "依据", "经典依据")
        for field in text_fields:
            if not assertion[field].strip():
                errors.append(f"[{aid}] 必填字段 {field} 为空")
            if assertion[field].count("（") != assertion[field].count("）"):
                errors.append(f"[{aid}] {field} 中文括号不平衡")
            if PROCESS_RE.search(assertion[field]):
                errors.append(f"[{aid}] {field} 含评测/内部过程残留")
        if "《" not in assertion["经典依据"] or "》" not in assertion["经典依据"]:
            errors.append(f"[{aid}] 经典依据缺少可核验书名")

        assertion_groups = [groups[key] for key in sorted(groups, key=lambda item: int(item[1])) if key[0] == aid]
        if not assertion_groups:
            errors.append(f"[{side}/{rule}/{aid}] 断言无条件组（恒不命中）")
            continue
        ordered_numbers = sorted(int(key[1]) for key in groups if key[0] == aid)
        if ordered_numbers != list(range(1, len(ordered_numbers) + 1)):
            errors.append(f"[{aid}] condition_group_id 不连续: {ordered_numbers}")

        selected_sides = set()
        has_flow_trigger = False
        for conditions_in_group in assertion_groups:
            # Current mechanical exclusivity guards. These are schema facts, not命理裁决.
            if sum(factor in {
                "大运印星运", "大运官杀运", "大运财星运", "大运食伤运", "大运比劫运"
            } for factor in conditions_in_group) > 1:
                errors.append(f"[{aid}] 同组同时要求多个互斥当前大运类")
            if sum(factor.startswith("年柱") for factor in conditions_in_group) > 1:
                errors.append(f"[{aid}] 同组同时要求多个年柱天干十神")
            for factor, expected in conditions_in_group.items():
                if factor in CONTEXT_KEYS:
                    if factor not in natal_sides and factor not in flow_sides:
                        selected_sides.add("context")
                    continue
                pool = flow_sides if rule.startswith("年") else natal_sides
                factor_sides = pool.get(factor)
                if factor_sides is None:
                    errors.append(f"[{side}/{rule}/{aid}] 因子 '{factor}' 不在对应作用域因子表")
                    continue
                if rule.startswith("年") and factor in flow_trigger_factors:
                    has_flow_trigger = True
                selected_sides.update(factor_sides)
                if side in {"bazi", "ziwei"} and factor_sides != {side} and "common" not in factor_sides:
                    errors.append(
                        f"[{side}/{rule}/{aid}] 引用异侧因子 '{factor}'（{sorted(factor_sides)}）"
                    )
                if expected not in ("0", "1") and not direct_string.get(factor):
                    errors.append(f"[{aid}/{factor}] 非二值因子使用了字符串期望值")
        if side == "common":
            if not {"bazi", "ziwei"} <= selected_sides:
                errors.append(f"[common/{rule}/{aid}] 合参断语未同时覆盖八字与紫微因子")
        if rule.startswith("年") and not has_flow_trigger:
            errors.append(
                f"[{side}/{rule}/{aid}] 流年断语缺少当年引动因子，仅本命体质会逐年重复命中"
            )

    # Duplicate condition signature + event is redundant; different events are projections.
    seen_signature: dict[tuple, str] = {}
    for (side, rule), rows in grouped_assertions.items():
        for row in rows:
            aid = row["assertion_id"]
            signatures = []
            for key, conditions_in_group in groups.items():
                if key[0] != aid:
                    continue
                signatures.append(tuple(sorted(conditions_in_group.items())))
            signature = tuple(sorted(signatures))
            dedup_key = (side, rule, row["事件"], signature)
            if dedup_key in seen_signature:
                warnings.append(
                    f"[{side}/{rule}] 重复条件+事件: {seen_signature[dedup_key]} 与 {aid}"
                )
            else:
                seen_signature[dedup_key] = aid

    # Scalar value closures.
    enum_sources = {
        "月令格": "月令格局", "扶抑从格": "扶抑从格", "身强弱": "身强弱状态",
        "调候季节": "调候季节", "日主": "天干", "日主五行": "五行",
        "日主长生状态": "十二长生", "日支神煞类型": "日支神煞",
        "月令本气十神": "十神", "大运十神类": "十神大类",
        "流年日主长生状态": "十二长生",
    }
    for (aid, _gid), conditions_in_group in groups.items():
        for factor, expected in conditions_in_group.items():
            source = enum_sources.get(factor)
            if source and expected not in constants.get(source, []):
                errors.append(f"[{aid}] {factor}={expected!r} 不在 constants.{source} 闭集")

    # Factor-table production purity and empty definitions.
    for filename, rows in (("factors.csv", natal_rows), ("factors_liunian.csv", flow_rows)):
        for row in rows:
            if not row.get("直通") and not row.get("conds"):
                errors.append(f"[{filename}] 因子 {row['因子']} 空定义")
            if row.get("术数") not in {"bazi", "ziwei", "common"}:
                errors.append(f"[{filename}] 因子 {row['因子']} 术数无效: {row.get('术数')}")
            if PROCESS_RE.search(row.get("依据") or ""):
                errors.append(f"[{filename}] 因子 {row['因子']} 含过程残留")

    print(f"因子全集: 本命 {len(natal_sides)} / 流年 {len(flow_sides)} 个")
    print(f"断语: {len(assertions)} 条 / 条件组 {len(groups)} 组 / 条件行 {len(conditions)} 行")
    print(f"错误: {len(errors)} 个")
    for error in errors:
        print("  ✗", error)
    print(f"警告: {len(warnings)} 个")
    for warning in warnings[:40]:
        print("  ⚠", warning)
    if len(warnings) > 40:
        print(f"  ... 共 {len(warnings)} 条警告")
    return 1 if errors or warnings else 0


if __name__ == "__main__":
    sys.exit(main())
