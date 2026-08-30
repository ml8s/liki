"""考时定盘 — 多候选生日 × 人生事件交叉校验。

命理服务：当出生时辰不确定时，用已知人生大事校验多个候选盘，
找出与实际经历最吻合的盘。

数据流：N 个候选生日 → 排盘 → 流年排盘 → 因子 → 断语 → 对比矩阵
不做命中判断——信号方向/强度的解读由 LLM 完成。
"""

from __future__ import annotations

from paipan import full_paipan, liunian
from factors import make_liunian_factors
from duanyu import query_yearly, _YEARLY_RULES

def calibrate(candidates: list, events: list, detail: bool = False) -> dict:
    for e in events:
        if e.get("rule") not in _YEARLY_RULES:
            raise ValueError(
                f"calibrate events.rule 必须是流年域，收到: '{e.get('rule')}'。"
                f"有效域: {sorted(_YEARLY_RULES)}")
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
    results = {}
    for c in candidates:
        label = c.get("label", "")
        if not label:
            raise ValueError("calibrate candidates 每项必须含 label")
        if "longitude" not in c or c.get("longitude") is None:
            raise ValueError(
                f"candidate '{label}' 缺少 longitude，禁止静默降级")
        if "gregorian" not in c or not c.get("gregorian"):
            raise ValueError(f"candidate '{label}' 缺少 gregorian（出生公历时间）")
        if "gender" not in c or c.get("gender") not in ("male", "female"):
            raise ValueError(f"candidate '{label}' 缺少 gender 或值不是 male/female")
        pan = full_paipan(c["gregorian"], c["gender"],
                  longitude=c["longitude"], correct=c.get("correct", True))
        event_results = []
        year_cache = {}
        for e in events:
            year = e["year"]
            if year not in year_cache:
                lnp = liunian(pan, year)
                snap = make_liunian_factors(pan, lnp, year=year)
                year_cache[year] = snap
            else:
                snap = year_cache[year]
            r = query_yearly(e["rule"], snap)
            if not detail:
                r = {
                    "八字": [{k: item[k] for k in ("事件", "结论") if k in item} for item in r.get("八字", [])],
                    "紫微": [{k: item[k] for k in ("事件", "结论") if k in item} for item in r.get("紫微", [])],
                }
            event_results.append({
                "year": year, "label": e["label"], "rule": e["rule"],
                "八字": r.get("八字", []), "紫微": r.get("紫微", []),
            })
        results[label] = event_results
    return results
