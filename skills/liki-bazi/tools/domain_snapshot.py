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
    contract = load_contract()
    pillar_suffix = contract["柱字段后缀"]
    for field, fact in contract["柱值字段"].items():
        if p.get(field):
            extras[f"{pillar}{pillar_suffix}{fact}"] = p[field]
    for field, fact in contract["柱存在字段"].items():
        if p.get(field) is not None:
            extras[f"{pillar}{pillar_suffix}{fact}"] = p[field]
    return extras


def _project_bazi_facts(pan: dict) -> dict:
    contract = load_contract()
    full = pan.get("full", {}) or {}
    chart = pan.get("chart", {}) or {}
    facts = {}
    for pillar in contract["四柱"]:
        facts.update(_pillar_extras(full, pillar))
    for source_key, fact_key in contract["八字字段映射"]["full"].items():
        if full.get(source_key):
            facts[fact_key] = full[source_key]
    for source_key, fact_key in contract["八字字段映射"]["chart"].items():
        if chart.get(source_key):
            facts[fact_key] = chart[source_key]
    return facts


def _project_ziwei_facts(pan: dict) -> dict:
    contract = load_contract()
    zw = pan.get("ziwei", {}) or {}
    facts = {}
    for source_key, fact_key in contract["紫微字段映射"]["ziwei"].items():
        value = zw.get(source_key)
        if value:
            facts[fact_key] = value
    for source_key, fact_key in contract["紫微字段映射"]["pan"].items():
        value = pan.get(source_key)
        if value:
            facts[fact_key] = value
    return facts


def project_domain_facts(pan: dict) -> dict:
    """从 pan 投影稳定领域事实；只读 pan，不写任何私有缓存。"""
    projectors = {"bazi": _project_bazi_facts, "ziwei": _project_ziwei_facts}
    return {
        label: projectors[code](pan)
        for code, label in load_contract()["投影侧代码"].items()
    }
