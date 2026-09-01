"""领域快照层 — pan 中稳定命理事实的只读投影。

snap 不是只为当前断语表服务的中间结构：这些 reserved domain facts 是领域模型
的一部分，当前未消费不代表可删。本层只做结构投影，不做因子判定。
"""
from __future__ import annotations

import json
import os

__all__ = ["project_domain_facts", "load_contract", "CONTRACT_PATH"]

CONTRACT_PATH = os.path.join(
    os.path.dirname(os.path.abspath(__file__)), "domain_snapshot_contract.json"
)
_CONTRACT = None


def load_contract() -> dict:
    """加载领域快照契约；契约是投影字段的唯一事实源。"""
    global _CONTRACT
    if _CONTRACT is None:
        with open(CONTRACT_PATH, encoding="utf-8") as fh:
            _CONTRACT = json.load(fh)
    return _CONTRACT


def _pillar_extras(full: dict, pillar: str) -> dict:
    p = full.get(pillar, {}) or {}
    extras = {}
    if p.get("na_yin"):
        extras[f"{pillar}柱纳音"] = p["na_yin"]
    if p.get("cang_gan"):
        extras[f"{pillar}柱藏干"] = p.get("cang_gan")
    if p.get("is_void") is not None:
        extras[f"{pillar}柱旬空"] = p["is_void"]
    if p.get("is_self_he") is not None:
        extras[f"{pillar}柱自合"] = p["is_self_he"]
    if p.get("is_kui_gang") is not None:
        extras[f"{pillar}柱魁罡"] = p["is_kui_gang"]
    if p.get("self_he_name"):
        extras[f"{pillar}柱自合名"] = p["self_he_name"]
    return extras


def _project_bazi_facts(pan: dict) -> dict:
    full = pan.get("full", {}) or {}
    chart = pan.get("chart", {}) or {}
    facts = {}
    for pillar in ("nian", "yue", "ri", "shi"):
        facts.update(_pillar_extras(full, pillar))
    for source_key, fact_key in (
        ("san_yuan", "三元"), ("xun_kong", "旬空"),
        ("san_qi_name", "三奇贵人"), ("gong_jia", "拱夹"),
        ("nayin_rel", "纳音生克"),
    ):
        if full.get(source_key):
            facts[fact_key] = full[source_key]
    # 大运完整透传：干支、五行、日期段、起止年与索引都是领域事实。
    if chart.get("da_yun"):
        facts["大运"] = chart["da_yun"]
    return facts


def _project_ziwei_facts(pan: dict) -> dict:
    zw = pan.get("ziwei", {}) or {}
    facts = {}
    mappings = (
        ("gong_wei", "宫位"), ("ju_shu", "局数"),
        ("ming_zhu", "命主"), ("shen_zhu", "身主"),
        ("ming_gong", "命宫"), ("shen_gong", "身宫"),
        ("kong_gong", "空宫"), ("nian_gan", "年干"),
        ("nian_zhi", "太岁"), ("shi_zhi", "时支"),
        ("ziwei_pos", "紫微星位"),
    )
    for source_key, fact_key in mappings:
        value = zw.get(source_key)
        if value:
            facts[fact_key] = value
    return facts


def project_domain_facts(pan: dict) -> dict:
    """从 pan 投影稳定领域事实；只读 pan，不写任何私有缓存。"""
    return {
        "八字": _project_bazi_facts(pan),
        "紫微": _project_ziwei_facts(pan),
    }
