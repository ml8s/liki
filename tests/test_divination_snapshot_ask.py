from __future__ import annotations

import sys
from pathlib import Path

import pytest


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import divination_snapshot  # noqa: E402
import liuyao_ask  # noqa: E402
import liuyao_snapshot  # noqa: E402
import qimen_ask  # noqa: E402
import qimen_snapshot  # noqa: E402
from qimen_projection import project as project_chart  # noqa: E402


def _liuyao_projected():
    return {
        "question": {"text": "这次面试能不能通过？", "matter": "career", "explicit_yong_shen": None, "perspective": None, "solar_time": "2026-09-08T12:00:00+08:00"},
        "casting": {"mode": "coins", "order": "bottom_up", "casting_id": "a" * 64, "yaos": [7] * 6, "dong_yao": []},
        "board": {
            "name": "乾为天", "ben_gua": "乾", "palace": "乾", "palace_wuxing": "金",
            "lines": [{
                "position": index + 1, "type": 7, "gan_zhi": "甲子", "wuxing": "水",
                "liu_qin": "子孙", "liu_shou": "螣蛇", "shi_ying": "",
                "chang_sheng_yue": "病",
                "flags": {"moving": False, "yue_po": False, "xun_kong": False,
                          "mu_ku": False, "mu_ku_branch": None, "mu_ku_element": None,
                          "dong_sheng": False, "dong_ke": False},
            } for index in range(6)],
        },
        "focus": {"yong_shen": {"name": "官鬼", "position": 4, "wang_shuai": "相"}},
        "evidence": {
            "primary": [{"id": "yong-shen-state", "fact": {}}],
            "secondary": [],
            "conflicts": [], "reference": [], "ignored_scope": [],
        },
        "facts": {
            "hidden_lines": [], "branch_relation_facts": [], "day_clash_facts": [],
            "moving_transformations": [], "san_he_candidates": [],
            "force_chain": {"yong_shen": "官鬼", "yong_element": "火", "position": 4, "entries": []},
            "yong_shen_candidates": [],
        },
        "timing_candidates": [{"id": "timing-1", "mechanism": "动爻逢值"}],
        "policy": {"may_output": [], "must_cover_conflicts": True, "forbidden": []},
    }


def _liuyao_factors(**kwargs):
    return {
        "matter": {"matter": kwargs["matter"]},
        "factors": _liuyao_projected(),
    }


def test_snapshot_envelope_fields_cannot_be_overridden():
    with pytest.raises(ValueError, match="cannot override envelope"):
        divination_snapshot.build_snapshot(
            method="liuyao",
            schema_version="liuyao-snapshot-v4",
            payload={"method": "qimen"},
        )


def test_snapshot_validate_rejects_wrong_schema_version():
    snapshot = divination_snapshot.build_snapshot(
        method="liuyao",
        schema_version="liuyao-snapshot-v4",
        payload={"question": {"text": "测试"}},
    )
    with pytest.raises(ValueError, match="schema_version"):
        divination_snapshot.validate_snapshot(
            snapshot,
            method="liuyao",
            schema_version="liuyao-snapshot-v5",
        )


def test_liuyao_snapshot_is_immutable_envelope(monkeypatch):
    monkeypatch.setattr(liuyao_snapshot, "qigua", lambda **_: {"mode": "coins", "casting_id": "a" * 64, "yaos": [7] * 6, "dong_yao": []})
    monkeypatch.setattr(liuyao_snapshot, "build_liuyao_factors", _liuyao_factors)
    result = liuyao_snapshot.create(
        question="这次面试能不能通过？", mode="coins", matter="career"
    )
    assert result["method"] == "liuyao"
    assert result["schema_version"] == "liuyao-snapshot-v4"
    assert result["snapshot_digest"]
    assert result["policy"]["immutable"] is True
    assert "chart" not in result
    assert "solar_time" not in result
    assert result["snapshot_digest"] == divination_snapshot.canonical_digest(result)


def test_liuyao_ask_rejects_tampered_snapshot(monkeypatch):
    monkeypatch.setattr(liuyao_snapshot, "qigua", lambda **_: {"mode": "coins", "casting_id": "a" * 64, "yaos": [7] * 6, "dong_yao": []})
    monkeypatch.setattr(liuyao_snapshot, "build_liuyao_factors", _liuyao_factors)
    snapshot = liuyao_snapshot.create(question="这次面试能不能通过？", matter="career")
    snapshot["question"]["text"] = "篡改后的问题"
    with pytest.raises(ValueError, match="digest mismatch"):
        liuyao_ask.ask(snapshot, message="现在该注意什么？")


def test_liuyao_ask_returns_structured_answer(monkeypatch):
    monkeypatch.setattr(liuyao_snapshot, "qigua", lambda **_: {"mode": "coins", "casting_id": "a" * 64, "yaos": [7] * 6, "dong_yao": []})
    monkeypatch.setattr(liuyao_snapshot, "build_liuyao_factors", _liuyao_factors)
    snapshot = liuyao_snapshot.create(question="这次面试能不能通过？", matter="career")
    answer = liuyao_ask.ask(snapshot, message="现在该注意什么？")
    assert answer["method"] == "liuyao"
    assert answer["snapshot_digest"] == snapshot["snapshot_digest"]
    assert answer["primary_evidence_refs"] == ["yong-shen-state"]
    assert answer["audit"]["accepted"] is True
    assert "必然" not in answer["verdict"]


def _pan():
    from tests.test_qimen_tools import _pan as qimen_pan
    return qimen_pan()


def test_qimen_snapshot_is_immutable_envelope(monkeypatch):
    monkeypatch.setattr(qimen_snapshot, "server_time", lambda: "2026-09-08T12:00:00+08:00")
    monkeypatch.setattr(qimen_snapshot, "_resolve_location", lambda city, longitude: (city, 121.47))
    monkeypatch.setattr(qimen_snapshot, "solar_time", lambda *_: {"solar": "2026-09-08T11:57:00+08:00"})
    monkeypatch.setattr(qimen_snapshot, "qimen_chart", lambda *_, **__: {"matter": None, "chart": _pan()["chart"]})
    result = qimen_snapshot.create(question="该往哪里推进？", city="上海", matter="wealth")
    assert result["method"] == "qimen"
    assert result["schema_version"] == "qimen-snapshot-v3"
    assert result["snapshot_digest"] == divination_snapshot.canonical_digest(result)
    assert "chart" not in result
    assert result["method_context"]["scope"] == "hour"
    assert result["factors"]


def test_qimen_ask_rejects_wrong_method():
    wrong = {"method": "liuyao", "schema_version": "qimen-snapshot-v3", "snapshot_digest": "x"}
    with pytest.raises(ValueError, match="snapshot method|qimen_snapshot contract failed"):
        qimen_ask.ask(wrong, message="现在适合行动吗？")


def test_qimen_ask_returns_structured_answer(monkeypatch):
    monkeypatch.setattr(qimen_snapshot, "server_time", lambda: "2026-09-08T12:00:00+08:00")
    monkeypatch.setattr(qimen_snapshot, "_resolve_location", lambda city, longitude: (city, 121.47))
    monkeypatch.setattr(qimen_snapshot, "solar_time", lambda *_: {"solar": "2026-09-08T11:57:00+08:00"})
    monkeypatch.setattr(qimen_snapshot, "qimen_chart", lambda *_, **__: {"matter": None, "chart": _pan()["chart"]})
    snapshot = qimen_snapshot.create(question="该往哪里推进？", city="上海", matter="wealth")
    answer = qimen_ask.ask(snapshot, message="现在适合行动吗？")
    assert answer["method"] == "qimen"
    assert answer["snapshot_digest"] == snapshot["snapshot_digest"]
    assert answer["audit"]["accepted"] is True
    assert answer["evidence_refs"]


def test_liuyao_snapshot_matches_contract(monkeypatch):
    import json
    from jsonschema import validate

    monkeypatch.setattr(liuyao_snapshot, "qigua", lambda **_: {
        "mode": "coins",
        "order": "bottom_up",
        "casting_id": "a" * 64,
        "rounds": [
            {"position": i, "coins": ["正", "正", "反"], "value": 8, "label": "少阴", "changing": False}
            for i in range(1, 7)
        ],
        "yaos": [8] * 6,
        "dong_yao": [],
    })
    monkeypatch.setattr(liuyao_snapshot, "build_liuyao_factors", _liuyao_factors)
    snapshot = liuyao_snapshot.create(question="这次面试能不能通过？", matter="career")
    contract = json.loads((TOOLS / "liuyao_snapshot_contract.json").read_text(encoding="utf-8"))
    validate(snapshot, contract)


def test_qimen_snapshot_matches_contract(monkeypatch):
    import json
    from jsonschema import validate

    monkeypatch.setattr(qimen_snapshot, "server_time", lambda: "2026-09-08T12:00:00+08:00")
    monkeypatch.setattr(qimen_snapshot, "_resolve_location", lambda city, longitude: (city, 121.47))
    monkeypatch.setattr(qimen_snapshot, "solar_time", lambda *_: {"solar": "2026-09-08T11:57:00+08:00"})
    monkeypatch.setattr(qimen_snapshot, "qimen_chart", lambda *_, **__: {"matter": None, "chart": _pan()["chart"]})
    snapshot = qimen_snapshot.create(question="该往哪里推进？", city="上海", matter="wealth")
    contract = json.loads((TOOLS / "qimen_snapshot_contract.json").read_text(encoding="utf-8"))
    validate(snapshot, contract)


def test_qimen_snapshot_rejects_incompatible_rule_before_rpc(monkeypatch):
    monkeypatch.setattr(qimen_snapshot, "server_time", lambda: "2026-09-08T12:00:00+08:00")
    monkeypatch.setattr(qimen_snapshot, "_resolve_location", lambda city, longitude: (city, 121.47))
    monkeypatch.setattr(qimen_snapshot, "solar_time", lambda *_: (_ for _ in ()).throw(AssertionError("engine called")))
    with pytest.raises(ValueError, match="lost_property requires scope"):
        qimen_snapshot.create(
            question="该往哪里推进？", city="上海", matter="wealth",
            rule="lost_property", scope="day", school="zhuanpan",
        )


def test_qimen_answer_timing_ids_are_collision_resistant(monkeypatch):
    from tests.test_divination_snapshot_ask import _pan

    monkeypatch.setattr(qimen_snapshot, "server_time", lambda: "2026-09-08T12:00:00+08:00")
    monkeypatch.setattr(qimen_snapshot, "_resolve_location", lambda city, longitude: (city, 121.47))
    monkeypatch.setattr(qimen_snapshot, "solar_time", lambda *_: {"solar": "2026-09-08T11:57:00+08:00"})
    chart = _pan()["chart"]
    chart["factors"] if False else None
    factors = project_chart(chart)
    factors["ying_qi"].extend([
        {"type": "ma_xing", "branch": "子", "gong": "坎", "related_to": []},
        {"type": "ma_xing", "branch": "子", "gong": "坎", "related_to": [{"symbol": "日干"}]},
    ])
    snapshot = qimen_snapshot.build_snapshot(
        method="qimen",
        schema_version=qimen_snapshot.SCHEMA_VERSION,
        payload={
            "snapshot_kind": "standard",
            "question": "该往哪里推进？",
            "input": {"city": "上海", "longitude": 121.47, "local_time": "2026-09-08T12:00:00+08:00", "solar_time": "2026-09-08T11:57:00+08:00"},
            "matter": None,
            "method_context": {"scope": "hour", "school": "zhuanpan"},
            "factors": factors,
            "special": None,
            "policy": {"immutable": True, "no_rechart_without_new_event": True},
        },
    )
    answer = qimen_ask.ask(snapshot, message="现在适合行动吗？")
    assert len(answer["timing_refs"]) == len(factors["ying_qi"])
