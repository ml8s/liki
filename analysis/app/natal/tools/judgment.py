"""judgment 判断层：compute_factors / natal_query / period_query。

正交化后 engine-pro 的补集能力（判断，不排盘）：
- compute_factors(chart)    ：engine 排盘结果 → 因子快照（内部调 engine fullchart）
- natal_query(factors, ...) ：因子快照 → 本命断语
- period_query(factors, ...)：因子快照 → 应期断语（内部调 engine 流年/大限）

输入 chart 为 engine 分领域排盘输出（bazi_chart / ziwei_chart，公共契约）。
"""
from __future__ import annotations

import hashlib
import json

from duanyu import query, yearly_range
from factors import evaluate_factors
from paipan import _bazi_fullchart, _ziwei_daxian, _ziwei_fullchart


def _factors_digest(factors: dict) -> str:
    raw = json.dumps(factors, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()


def _is_bazi_chart(chart: dict) -> bool:
    return "ri" in chart  # 八字盘含日柱；紫微盘为宫位结构


def compute_factors(chart: dict) -> dict:
    """engine 排盘结果 → { factors, factors_digest }。

    chart 为 bazi_chart（八字）或 ziwei_chart（紫微）输出。内部调 engine
    fullchart 取因子所需详细字段，再求因子快照（与正交化基线一致）。
    """
    gender = chart.get("gender", "")
    if _is_bazi_chart(chart):
        full = _bazi_fullchart(chart)
        pan = {"chart": chart, "full": full, "gender": gender}
        factors = evaluate_factors(gender, pan, shushi="bazi")
    else:
        zw = _ziwei_fullchart(chart)
        daxian = _ziwei_daxian(zw)
        pan = {"chart": chart, "ziwei": zw, "ziwei_daxian": daxian, "gender": gender}
        factors = evaluate_factors(gender, pan, shushi="ziwei")
    return {"factors": factors, "factors_digest": _factors_digest(factors)}


def natal_query(factors: dict, topics: list[str]) -> dict:
    """因子快照 + topics → 本命断语。

    TODO(orthogonal): 断语层从 pan 输入改造为 factors 快照输入后实现。
    """
    raise NotImplementedError("natal_query 待正交化断语层改造后实现")


def period_query(factors: dict, time_scope: dict, topics: list[str]) -> dict:
    """因子快照 + 时间层 + topics → 应期断语。

    TODO(orthogonal): 断语层从 pan 输入改造为 factors 快照输入后实现。
    """
    raise NotImplementedError("period_query 待正交化断语层改造后实现")