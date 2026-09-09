from __future__ import annotations

import sys
from pathlib import Path

import pytest


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import divination_safety  # noqa: E402
import huangli_days  # noqa: E402
import liuyao_read  # noqa: E402
import qimen_read  # noqa: E402
from divination_contracts import validate_document  # noqa: E402


def test_shared_safety_blocks_direct_entries():
    question = "医生说我病情很重，还能活多久？"
    for result in (
        liuyao_read.read(casting={"mode": "auto"}, question=question),
        qimen_read.read(question=question, city="上海"),
        huangli_days.days(question=question, event="medical"),
    ):
        assert result["blocked"] is True
        assert result["route"] == "blocked"
        assert result["safety"]["category"] in {"serious_medical", "emergency_medical"}


def test_shared_rpc_module_is_single_source():
    import divination_rpc
    import qimen_paipan

    assert qimen_paipan.shared_call is divination_rpc.call


def test_huangli_normalizes_range_and_event(monkeypatch):
    calls = []

    def fake_engine_data(method, params):
        calls.append((method, params))
        if method == "time.now":
            return {"cst": "2026-09-08T12:00:00+08:00"}
        assert method == "huangli.days"
        return [{
            "date": "2026-09-09",
            "jian_chu": "成",
            "gan_ji": "甲不开仓财物耗散",
            "zhi_ji": "巳不远行财物伏藏",
        }]

    monkeypatch.setattr(huangli_days, "engine_data", fake_engine_data)
    result = huangli_days.days(
        question="哪天适合开业？", event="开业",
        start_date="2026-09-09", end_date="2026-09-09",
    )
    assert calls[-1] == ("huangli.days", {"start_date": "2026-09-09", "count": 1})
    assert result["event"] == "opening"
    assert result["recommended"][0]["date"] == "2026-09-09"
    assert result["recommended"][0]["suitability"] == "recommended"
    assert result["candidates"][0]["warnings"]


def test_huangli_rejects_over_range(monkeypatch):
    monkeypatch.setattr(huangli_days, "_server_date", lambda: __import__("datetime").date(2026, 9, 8))
    with pytest.raises(ValueError, match="1-30 days"):
        huangli_days.days(question="哪天签约？", event="sign", days=31)


def test_central_contracts_validate_sample_documents():
    snapshot = {
        "schema_version": "liuyao-snapshot-v2",
        "question": {"text": "测试", "domain": "career"},
        "casting": {"mode": "coins", "fingerprint": "abc", "yaos": [7] * 6, "dong_yao": []},
        "board": {"name": "乾为天", "ben_gua": "乾", "bian_gua": None, "palace": "乾", "palace_wuxing": "金", "lines": [{}, {}, {}, {}, {}, {}]},
        "focus": {"yong_shen": {}, "yong_line": None, "world": None, "other": None},
        "evidence": {"primary": [], "secondary": [], "reference": [], "conflicts": [], "ignored_scope": []},
        "facts": {
            "hidden_lines": [], "branch_relation_facts": [], "day_clash_facts": [],
            "moving_transformations": [], "san_he_candidates": [],
            "force_chain": {}, "yong_shen_candidates": [],
        },
        "timing_candidates": [],
        "policy": {"may_output": [], "must_cover_conflicts": True, "forbidden": []},
        "followup": {"locked": True, "casting_fingerprint": "abc", "yaos": [7] * 6, "dong_yao": [], "rules": []},
    }
    validate_document("liuyao_snapshot", snapshot)

    read = {
        "schema_version": "qimen-read-v1",
        "snapshot": {"method": {"scope": "hour", "school": "zhuanpan"}, "ying_qi": []},
        "special": {"assertions": [{"id": "qimen_lost_property_direction"}]},
    }
    from qimen_report import template
    report = template(read)
    report.update({
        "headline": "方向候选可用，需现实核验",
        "verdict": "当前宫位支持推进候选，不承诺结果。",
        "action": "先完成现实核查，再按候选方向推进。",
    })
    validate_document("qimen_report", report)
    with pytest.raises(ValueError, match="contract failed"):
        validate_document("liuyao_snapshot", {"schema_version": "wrong"})
