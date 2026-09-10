from __future__ import annotations

import sys
from pathlib import Path

import pytest
from jsonschema.exceptions import ValidationError


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import divination_contracts  # noqa: E402
import divination_rpc  # noqa: E402
import divination_snapshot  # noqa: E402
import divination_safety  # noqa: E402
import huangli_days  # noqa: E402
import json  # noqa: E402
from qimen_interpretations import load_rule_table  # noqa: E402
import liuyao_ask  # noqa: E402
import liuyao_snapshot  # noqa: E402
import qimen_ask  # noqa: E402
import qimen_paipan  # noqa: E402
import qimen_snapshot  # noqa: E402
from qimen_projection import validate as validate_standard_factors  # noqa: E402


def test_safety_block_matches_common_contract():
    safety = divination_safety.assess("我不想活了")
    payload = divination_safety.blocked_payload(
        question="我不想活了", method="liuyao", safety=safety,
    )
    divination_contracts.validate_document("divination_blocked", payload)
    assert payload["policy"]["no_casting"] is True


def test_safety_rules_support_multilingual_boundary():
    cases = {
        "kill myself": "self_harm",
        "I don't want to live": "self_harm",
        "自杀": "self_harm",
        "I can't breathe": "emergency_medical",
        " overdose": "emergency_medical",
        "How long do I have to live?": "serious_medical",
        "domestic violence": "domestic_violence",
        "all my money": "major_financial_risk",
    }
    for question, expected in cases.items():
        safety = divination_safety.assess(question)
        assert safety["status"] == "redirect", question
        assert safety["category"] == expected, question


def test_safety_rule_table_is_data_driven():
    source = (TOOLS / "divination_safety.py").read_text(encoding="utf-8")
    assert 'RULES = [' not in source
    assert 'GUIDANCE = {' not in source
    assert "divination_safety_rules.json" in source


def test_agent_cli_fails_closed_on_incompatible_engine(monkeypatch):
    monkeypatch.setattr(
        divination_rpc,
        "engine_version",
        lambda: "2026.09.02.1",
    )
    with pytest.raises(Exception, match="engine version .* incompatible"):
        divination_rpc.ensure_engine_compatible()


def test_engine_version_gate_accepts_minimum(monkeypatch):
    monkeypatch.setattr(
        divination_rpc,
        "engine_version",
        lambda: "2026.09.11.0",
    )
    divination_rpc.ensure_engine_compatible()


def test_liuyao_snapshot_rejects_conflicting_arguments_before_rpc(monkeypatch):
    monkeypatch.setattr(liuyao_snapshot, "qigua", lambda **_: (_ for _ in ()).throw(AssertionError("engine called")))
    with pytest.raises(ValueError, match="exactly one"):
        liuyao_snapshot.create(
            question="测试", matter="career", yong_shen="官鬼", mode="yaos", yaos=[7] * 6,
        )
    with pytest.raises(ValueError, match="perspective"):
        liuyao_snapshot.create(
            question="测试", matter="career", perspective="male", mode="yaos", yaos=[7] * 6,
        )


def test_qimen_snapshot_rejects_conflicting_location_before_rpc(monkeypatch):
    monkeypatch.setattr(qimen_snapshot, "server_time", lambda: "2026-09-09T12:00:00+08:00")
    monkeypatch.setattr(qimen_snapshot, "solar_time", lambda *_: (_ for _ in ()).throw(AssertionError("engine called")))
    with pytest.raises(ValueError, match="exactly one"):
        qimen_snapshot.create(
            question="测试", city="上海", longitude=121.47, matter="wealth",
        )


def test_huangli_days_rejects_ambiguous_range_before_rpc(monkeypatch):
    monkeypatch.setattr(huangli_days, "_server_date", lambda: (_ for _ in ()).throw(AssertionError("engine called")))
    with pytest.raises(ValueError, match="only one"):
        huangli_days.days(
            question="哪天适合签约？", event="sign",
            end_date="2026-10-31", days=7,
        )


def test_huangli_days_rejects_bool_days_before_rpc(monkeypatch):
    monkeypatch.setattr(huangli_days, "_server_date", lambda: (_ for _ in ()).throw(AssertionError("engine called")))
    with pytest.raises(ValueError, match="days must be an integer"):
        huangli_days.days(question="哪天适合签约？", event="sign", days=True)


def test_huangli_days_rejects_malformed_engine_day(monkeypatch):
    monkeypatch.setattr(huangli_days, "_server_date", lambda: __import__("datetime").date(2026, 9, 9))
    monkeypatch.setattr(huangli_days, "engine_items", lambda *_: [None])
    with pytest.raises(ValueError, match="non-object day"):
        huangli_days.days(question="哪天适合签约？", event="sign", days=1)


def test_huangli_days_strips_question(monkeypatch):
    monkeypatch.setattr(huangli_days, "_server_date", lambda: __import__("datetime").date(2026, 9, 9))
    monkeypatch.setattr(huangli_days, "engine_items", lambda *_: [{"date": "2026-09-10", "jian_chu": "定"}])
    result = huangli_days.days(question="  哪天适合签约？  ")
    assert result["question"] == "哪天适合签约？"


def test_huangli_days_projects_engine_event_classification(monkeypatch):
    monkeypatch.setattr(huangli_days, "_server_date", lambda: __import__("datetime").date(2026, 9, 9))
    monkeypatch.setattr(
        huangli_days,
        "engine_items",
        lambda *_: [{
            "date": "2026-09-10",
            "jian_chu": "定",
            "event": "sign",
            "event_label": "签约",
            "suitability": "recommended",
            "reason": "建除「定」适合签约。",
            "gan_ji": "甲不开仓",
            "zhi_ji": "子不问卜",
        }],
    )
    result = huangli_days.days(question="哪天适合签约？", event="sign", days=1)
    assert result["event"] == "sign"
    assert result["event_label"] == "签约"
    assert result["recommended"][0]["suitability"] == "recommended"
    assert result["recommended"][0]["warnings"] == ["甲不开仓", "子不问卜"]


def test_qimen_snapshot_rejects_malformed_factors(monkeypatch):
    from tests.test_divination_snapshot_ask import _pan

    monkeypatch.setattr(qimen_snapshot, "server_time", lambda: "2026-09-08T12:00:00+08:00")
    monkeypatch.setattr(qimen_snapshot, "_resolve_location", lambda city, longitude: (city, 121.47))
    monkeypatch.setattr(qimen_snapshot, "solar_time", lambda *_: {"solar": "2026-09-08T11:57:00+08:00"})
    monkeypatch.setattr(qimen_snapshot, "qimen_chart", lambda *_, **__: {"matter": None, "chart": _pan()["chart"]})
    snapshot = qimen_snapshot.create(question="该往哪里推进？", city="上海", matter="wealth")
    assert snapshot["snapshot_kind"] == "standard"
    snapshot["factors"].pop("scope")
    snapshot["snapshot_digest"] = divination_snapshot.canonical_digest(snapshot)
    with pytest.raises(ValueError, match="standard qimen factors invalid"):
        validate_standard_factors(snapshot["factors"])


def test_qimen_snapshot_rejects_method_context_mismatch(monkeypatch):
    from tests.test_divination_snapshot_ask import _pan

    valid_chart = _pan()["chart"]
    mismatch_chart = json.loads(json.dumps(valid_chart))
    mismatch_chart["method"]["scope"] = "day"
    from qimen_projection import project as project_chart

    monkeypatch.setattr(qimen_snapshot, "server_time", lambda: "2026-09-08T12:00:00+08:00")
    monkeypatch.setattr(qimen_snapshot, "_resolve_location", lambda city, longitude: (city, 121.47))
    monkeypatch.setattr(qimen_snapshot, "solar_time", lambda *_: {"solar": "2026-09-08T11:57:00+08:00"})
    monkeypatch.setattr(qimen_snapshot, "qimen_chart", lambda *_, **__: {"matter": None, "chart": mismatch_chart})
    monkeypatch.setattr(qimen_snapshot, "project_standard_factors", lambda _: project_chart(valid_chart))
    with pytest.raises(ValueError, match="method context does not match"):
        qimen_snapshot.create(question="该往哪里推进？", city="上海", matter="wealth")


def test_qimen_snapshot_rejects_whitespace_city_before_rpc(monkeypatch):
    monkeypatch.setattr(qimen_paipan, "city_coords", lambda *_: (_ for _ in ()).throw(AssertionError("engine called")))
    with pytest.raises(ValueError, match="city must be non-empty"):
        qimen_snapshot.create(question="该往哪里推进？", city="   ", matter="wealth")


def test_qimen_snapshot_rejects_empty_yong_shen_before_rpc(monkeypatch):
    monkeypatch.setattr(qimen_paipan, "city_coords", lambda *_: (_ for _ in ()).throw(AssertionError("engine called")))
    with pytest.raises(ValueError, match="yong_shen cannot be empty"):
        qimen_snapshot.create(question="该往哪里推进？", longitude=121.47, yong_shen=[])


def test_liuyao_snapshot_contract_rejects_invalid_yao_values():
    from jsonschema import validate
    from jsonschema.exceptions import ValidationError

    contract = divination_contracts.load_contract("liuyao_snapshot")
    with pytest.raises(ValidationError):
        validate(
            {"mode": "yaos", "order": "bottom_up", "casting_id": "a" * 64, "yaos": [1] * 6, "dong_yao": []},
            contract["properties"]["casting"],
        )


def test_liuyao_snapshot_contract_allows_mode_specific_casting_shapes():
    from jsonschema import validate

    contract = divination_contracts.load_contract("liuyao_snapshot")
    casting = contract["properties"]["casting"]
    validate({
        "mode": "coins",
        "order": "bottom_up",
        "casting_id": "a" * 64,
        "rounds": [
            {
                "position": index,
                "coins": ["正", "正", "反"],
                "value": 8,
                "label": "少阴",
                "changing": False,
            }
            for index in range(1, 7)
        ],
        "yaos": [8] * 6,
        "dong_yao": [],
    }, casting)
    validate({
        "mode": "yaos",
        "order": "bottom_up",
        "casting_id": "b" * 64,
        "rounds": [],
        "yaos": [7] * 6,
        "dong_yao": [],
    }, casting)

    with pytest.raises(ValidationError):
        validate({
            "mode": "yaos",
            "order": "bottom_up",
            "casting_id": "c" * 64,
            "rounds": [
                {
                    "position": 1,
                    "coins": ["正", "正", "反"],
                    "value": 8,
                    "label": "少阴",
                    "changing": False,
                }
            ],
            "yaos": [7] * 6,
            "dong_yao": [],
        }, casting)


def test_qimen_snapshot_special_rule_enum_matches_table():
    contract = divination_contracts.load_contract("qimen_snapshot")
    expected = sorted(load_rule_table())
    assert set(contract["properties"]["special"]["properties"]["rule"]["enum"]) == set(expected)


def test_contract_schema_versions_match_runtime_constants():
    import qimen_snapshot

    liuyao_contract = divination_contracts.load_contract("liuyao_snapshot")
    qimen_contract = divination_contracts.load_contract("qimen_snapshot")
    liuyao_answer_contract = divination_contracts.load_contract("liuyao_answer")
    qimen_answer_contract = divination_contracts.load_contract("qimen_answer")

    assert liuyao_contract["properties"]["schema_version"]["const"] == liuyao_snapshot.SCHEMA_VERSION
    assert qimen_contract["properties"]["schema_version"]["const"] == qimen_snapshot.SCHEMA_VERSION
    assert liuyao_answer_contract["properties"]["schema_version"]["const"] == liuyao_ask.ANSWER_SCHEMA_VERSION
    assert qimen_answer_contract["properties"]["schema_version"]["const"] == qimen_ask.ANSWER_SCHEMA_VERSION


def test_qimen_snapshot_rejects_jinhan_focus_before_rpc(monkeypatch):
    with pytest.raises(ValueError, match="jinhan"):
        qimen_snapshot.create(
            question="今天运势如何？", city="上海", scope="day",
            school="jinhan_yujing", matter="wealth",
        )


def test_qimen_snapshot_rejects_jinhan_missing_day_before_rpc(monkeypatch):
    with pytest.raises(ValueError, match="scope.*day|day.*scope"):
        qimen_snapshot.create(question="今天运势如何？", city="上海", school="jinhan_yujing")


def test_qimen_jinhan_projection_snapshot_and_answer(monkeypatch):
    import json
    from jsonschema import validate

    def fake_engine_data(method, params):
        assert method in {"city.coords", "tianwen.time", "qimen.chart"}
        if method == "city.coords":
            return {"longitude": 121.47}
        if method == "tianwen.time":
            return {"solar": "2026-09-08T11:57:00+08:00"}
        return {
            "method": {
                "scope": "day", "school": "jinhan_yujing",
                "dingju_method": "none", "dun_source": "winter_solstice_yang_summer_solstice_yin",
            },
            "pan": {
                "yin_dun": False, "ri_gan": "甲", "ri_zhi": "子",
                "gong_wei": [
                    {"gong": {"name": gong, "luoshu": index}, "xing": "太乙", "men": "休门", "men_present": True}
                    for index, gong in enumerate(["坎","坤","震","巽","中","乾","兑","艮","离"], 1)
                ],
                "day_spirits": [
                    {"zhi": f"zhi-{index}", "shen": f"shen-{index}"} for index in range(12)
                ],
            },
        }

    monkeypatch.setattr(qimen_paipan, "engine_data", fake_engine_data)
    monkeypatch.setattr(qimen_snapshot, "solar_time", lambda *_: {"solar": "2026-09-08T11:57:00+08:00"})
    snapshot = qimen_snapshot.create(
        question="今天整体如何？", longitude=121.47, time="2026-09-08T12:00:00+08:00",
        scope="day", school="jinhan_yujing",
    )
    assert snapshot["snapshot_kind"] == "jinhan"
    assert snapshot["factors"]["kind"] == "jinhan"
    assert "yong_shen" not in snapshot["factors"]
    divination_contracts.validate_document("qimen_snapshot", snapshot)

    answer = qimen_ask.ask(snapshot, message="今天整体态势如何？")
    assert answer["assertion_refs"] == []
    assert answer["timing_refs"] == []
    assert "snapshot:palaces" in answer["evidence_refs"]
    divination_contracts.validate_document("qimen_answer", answer)


def test_short_high_risk_question_is_blocked_before_length(monkeypatch):
    monkeypatch.setattr(liuyao_snapshot, "qigua", lambda **_: (_ for _ in ()).throw(AssertionError("engine called")))
    payload = liuyao_snapshot.create(question="我不想活", mode="yaos", yaos=[7] * 6, matter="career")
    assert payload["blocked"] is True
    assert payload["safety"]["status"] == "redirect"


def test_qimen_short_high_risk_message_is_blocked_before_length():
    payload = qimen_ask.ask(
        {
            "snapshot_kind": "jinhan",
            "schema_version": qimen_snapshot.SCHEMA_VERSION,
            "method": "qimen",
            "snapshot_digest": "x" * 16,
            "question": "今天整体态势如何？",
            "input": {"city": "上海", "longitude": 121.47, "local_time": "x", "solar_time": "x"},
            "matter": None,
            "method_context": {"scope": "day", "school": "jinhan_yujing"},
            "factors": {"kind": "jinhan", "school": "jinhan_yujing", "scope": "day"},
            "special": None,
            "policy": {"immutable": True, "no_rechart_without_new_event": True},
        },
        message="自杀",
    )
    assert payload["blocked"] is True


def test_qimen_snapshot_rejects_invalid_special_focus_before_rpc(monkeypatch):
    monkeypatch.setattr(qimen_snapshot, "server_time", lambda: (_ for _ in ()).throw(AssertionError("engine called")))

    with pytest.raises(ValueError, match="lost_property does not accept"):
        qimen_snapshot.create(
            question="钥匙丢了还能找到吗？", longitude=121.47,
            scope="hour", school="zhuanpan", rule="lost_property",
            matter="lost_item",
        )

    with pytest.raises(ValueError, match="missing_person requires matter"):
        qimen_snapshot.create(
            question="家人走失了", longitude=121.47,
            scope="hour", school="zhuanpan", rule="missing_person",
        )


def test_qimen_snapshot_accepts_missing_person_focus(monkeypatch):
    from tests.test_divination_snapshot_ask import _pan

    monkeypatch.setattr(qimen_snapshot, "server_time", lambda: "2026-09-08T12:00:00+08:00")
    monkeypatch.setattr(qimen_snapshot, "_resolve_location", lambda city, longitude: (city, 121.47))
    monkeypatch.setattr(qimen_snapshot, "solar_time", lambda *_: {"solar": "2026-09-08T11:57:00+08:00"})
    monkeypatch.setattr(
        qimen_snapshot,
        "qimen_chart",
        lambda *_, **__: {
            "matter": {"matter": "missing_person"},
            "chart": _pan()["chart"],
        },
    )
    snapshot = qimen_snapshot.create(
        question="家人走失了", longitude=121.47,
        scope="hour", school="zhuanpan",
        matter="missing_person", rule="missing_person",
    )
    assert snapshot["snapshot_kind"] == "standard"
    assert snapshot["special"]["rule"] == "missing_person"
