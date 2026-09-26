"""考时定盘 — 多候选生日 × 人生事件交叉校验。

命理服务：当出生时辰不确定时，用已知人生大事校验多个候选盘，
找出与实际经历最吻合的盘。

数据流：N 个候选生日 → 排盘 → 流年排盘 → 因子 → 断语 → 对比矩阵
不做命中判断——信号方向/强度的解读由 LLM 完成。
"""

from __future__ import annotations

from duanyu import (
    SCENE_ALIASES,
    YEARLY_RULES,
    brief,
    default_scene_domains,
    filter_domains,
    flow_factor_names,
    natal_factors_for_flow,
    query_yearly,
)
from errors import AssertionRuleError
from factor_constants import load_constants
from factors import evaluate_liunian_snap_from_pan, prepare_natal_context
from paipan import full_paipan, liunian
from pan_schema import GENDERS
from yearly_eval import query_year_rules, yearly_snapshot


def calibrate(candidates: list, events: list, detail: bool = False) -> dict:
    valid = set(YEARLY_RULES) | set(SCENE_ALIASES)
    for e in events:
        rules = e.get("rule")
        rule_items = rules if isinstance(rules, list) else [rules]
        invalid = [rule for rule in rule_items if rule not in valid]
        if invalid:
            received = {key: e.get(key) for key in ("domain", "rules", "rule", "year", "label")}
            raise AssertionRuleError(
                f"calibrate events.rule 必须是流年命理域或场景别名，无效: {invalid}。"
                f"收到 event: {received}。有效: {sorted(YEARLY_RULES)} + {sorted(SCENE_ALIASES)}")
        if "year" not in e or "label" not in e:
            raise ValueError("calibrate events 每项必须含 year、rule、label")
    labels = [c.get("label", "") for c in candidates]
    if len(labels) != len(set(labels)):
        dupes = [label for label in labels if labels.count(label) > 1]
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
    all_rules: list[str] = []
    for e in events:
        rules = e["rule"] if isinstance(e["rule"], list) else [e["rule"]]
        for rule in rules:
            if rule not in all_rules:
                all_rules.append(rule)
    flow_factors = flow_factor_names(all_rules)
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
            raw_rules = e["rule"] if isinstance(e["rule"], list) else [e["rule"]]
            rules = []
            for raw_rule in raw_rules:
                for expanded in SCENE_ALIASES.get(raw_rule, (raw_rule,)):
                    if expanded not in rules:
                        rules.append(expanded)
            grouped = query_year_rules(
                snapshot, rules, detail=True,
                query_yearly=query_yearly,
                brief=brief,
            )
            r = {side_labels[side]: [] for side in side_config["断言代码"]}
            for er in rules:
                qr = grouped[er]
                domains = e.get("domains", default_scene_domains(raw_rules))
                if domains is not None:
                    qr = filter_domains(qr, domains)
                for side in side_config["断言代码"]:
                    side_label = side_labels[side]
                    r[side_label] += qr.get(side_label, [])
            if not detail:
                r = {side: brief(items) for side, items in r.items()}
            event_results.append({
                "year": year, "label": e["label"], "rule": raw_rules,
                **r,
            })
        results[label] = event_results
    return results
