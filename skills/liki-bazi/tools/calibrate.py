"""考时定盘 — 多候选生日 × 人生事件交叉校验。

命理服务：当出生时辰不确定时，用已知人生大事校验多个候选盘，
找出与实际经历最吻合的盘。

数据流：N 个候选生日 → 排盘 → 流年排盘 → 因子 → 断语 → 对比矩阵
不做命中判断——信号方向/强度的解读由 LLM 完成。
"""

from __future__ import annotations

from errors import AssertionRuleError
from factors import evaluate_liunian_snap_from_pan, prepare_natal_context
from factor_constants import load_constants
from pan_schema import GENDERS
from paipan import full_paipan, liunian
from duanyu import (
    SCENE_ALIASES, YEARLY_RULES, brief, flow_factor_names,
    natal_factors_for_flow, query_yearly,
)
from yearly_eval import query_year_rules, yearly_snapshot


def calibrate(candidates: list, events: list, detail: bool = False) -> dict:
    valid = set(YEARLY_RULES) | set(SCENE_ALIASES)
    for e in events:
        if e.get("rule") not in valid:
            raise AssertionRuleError(
                f"calibrate events.rule 必须是流年命理域或场景别名，收到: '{e.get('rule')}'。"
                f"有效: {sorted(YEARLY_RULES)} + {sorted(SCENE_ALIASES)}")
        if "year" not in e or "label" not in e:
            raise ValueError("calibrate events 每项必须含 year、rule、label")
    labels = [c.get("label", "") for c in candidates]
    if len(labels) != len(set(labels)):
        dupes = [l for l in labels if labels.count(l) > 1]
        raise ValueError(
            f"calibrate candidates label 必须唯一，重复: {set(dupes)}。"
            f"重复 label 会静默覆盖前一个候选的结果。")
    if not candidates:
        raise ValueError("calibrate candidates 不能为空")
    if not events:
        raise ValueError("calibrate events 不能为空")
    if not 2 <= len(candidates) <= 3:
        raise ValueError(f"calibrate candidates 必须 2-3 个，收到 {len(candidates)} 个")
    if not 3 <= len(events) <= 5:
        raise ValueError(f"calibrate events 必须 3-5 件，收到 {len(events)} 件")
    flow_factors = flow_factor_names(list({e["rule"] for e in events}))
    natal_factors = natal_factors_for_flow(flow_factors)
    side_config = load_constants()["命理侧"]
    side_labels = side_config["标签"]
    results = {}
    for c in candidates:
        label = c.get("label", "")
        if not label:
            raise ValueError("calibrate candidates 每项必须含 label")
        correct = c.get("correct", True)
        if not isinstance(correct, bool):
            raise ValueError(f"candidate '{label}' correct 必须为 boolean")
        if correct and c.get("longitude") is None:
            raise ValueError(
                f"candidate '{label}' 缺少 longitude，禁止静默降级")
        if "gregorian" not in c or not c.get("gregorian"):
            raise ValueError(f"candidate '{label}' 缺少 gregorian（出生公历时间）")
        if "gender" not in c or c.get("gender") not in GENDERS:
            raise ValueError(
                f"candidate '{label}' 缺少 gender 或值不是 {'/'.join(GENDERS)}"
            )
        pan = full_paipan(c["gregorian"], c["gender"],
                  longitude=c.get("longitude"), correct=correct)
        event_results = []
        natal_context = prepare_natal_context(pan, factor_names=natal_factors)
        year_cache = {}
        for e in events:
            year = e["year"]
            if year not in year_cache:
                year_cache[year] = yearly_snapshot(
                    pan, year, natal_context,
                    liunian=liunian,
                    evaluate_liunian_snap_from_pan=evaluate_liunian_snap_from_pan,
                    factor_names=flow_factors,
                )
            snapshot = year_cache[year]
            rules = SCENE_ALIASES.get(e["rule"], (e["rule"],))
            grouped = query_year_rules(
                snapshot, rules, detail=True,
                query_yearly=query_yearly,
                brief=brief,
            )
            r = {side_labels[side]: [] for side in side_config["断言代码"]}
            evidence = {}
            for er in rules:
                qr = grouped[er]
                for side in side_config["断言代码"]:
                    side_label = side_labels[side]
                    r[side_label] += qr.get(side_label, [])
                evidence.update(qr.get("evidence", {}) or {})
            if not detail:
                r = {side: brief(items) for side, items in r.items()}
            else:
                r["evidence"] = evidence
            event_results.append({
                "year": year, "label": e["label"], "rule": e["rule"],
                **r,
            })
        results[label] = event_results
    return results
