from __future__ import annotations

import json
import sys
from pathlib import Path

import pytest


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import qimen_read  # noqa: E402


def _load_agent_cli():
    if str(TOOLS) not in sys.path:
        sys.path.insert(0, str(TOOLS))
    import importlib.util
    spec = importlib.util.spec_from_file_location("divination_agent_cli", TOOLS / "agent_cli.py")
    module = importlib.util.module_from_spec(spec)
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module

def test_llm_toolface_uses_qimen_read_not_internal_chart_tools():
    schema = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    names = {item["function"]["name"] for item in schema["tools"]}
    assert "qimen_read" in names
    assert {"qimen_chart", "query", "liuyao_chart"}.isdisjoint(names)

    agent_cli = _load_agent_cli()
    assert "qimen_read" in agent_cli._DISPATCH
    assert {"qimen_chart", "query", "liuyao_chart"}.isdisjoint(agent_cli._DISPATCH)


def test_internal_qimen_orchestration_still_uses_chart_layer(monkeypatch):
    calls = []

    def fake_engine_data(method, params):
        calls.append(method)
        if method == "time.now":
            return {"cst": "2026-09-09T12:00:00+08:00"}
        raise AssertionError(method)

    def fake_city(city):
        return {"longitude": 121.47}

    def fake_solar(value, longitude):
        return {"solar": "2026-09-09T11:57:00+08:00"}

    def fake_chart(solar, **kwargs):
        calls.append("qimen.chart")
        return {"matter": None, "chart": {}}

    monkeypatch.setattr(qimen_read, "engine_data", fake_engine_data)
    monkeypatch.setattr(qimen_read, "city_coords", fake_city)
    monkeypatch.setattr(qimen_read, "solar_time", fake_solar)
    monkeypatch.setattr(qimen_read, "qimen_chart", fake_chart)
    monkeypatch.setattr(qimen_read, "project_qimen_snapshot", lambda pan: {})

    result = qimen_read.read(
        question="当前态势",
        longitude=121.47,
        time="2026-09-09T12:00:00+08:00",
    )
    assert "qimen.chart" in calls
    assert result["input"]["solar_time"] == "2026-09-09T11:57:00+08:00"


def test_internal_qimen_special_envelope_is_not_double_wrapped(monkeypatch):
    monkeypatch.setattr(qimen_read, "engine_data", lambda method, params: {"cst": "2026-09-09T12:00:00+08:00"})
    monkeypatch.setattr(qimen_read, "solar_time", lambda time, longitude: {"solar": "2026-09-09T11:57:00+08:00"})
    monkeypatch.setattr(qimen_read, "qimen_chart", lambda solar, **kwargs: {"matter": None, "chart": {"method": {"scope": "hour", "school": "zhuanpan"}}})
    monkeypatch.setattr(qimen_read, "project_qimen_snapshot", lambda pan: {"method": {"scope": "hour", "school": "zhuanpan"}, "ying_qi": []})
    def fake_query(rule, snapshot):
        assert rule == "lost_property"
        return {
            "rule": rule,
            "assertions": [{"id": "qimen_lost_property_direction", "basis": "table"}],
        }

    monkeypatch.setattr(qimen_read, "query", fake_query)
    monkeypatch.setattr(qimen_read, "report_template", lambda read_result: {})

    result = qimen_read.read(
        question="钥匙还能找到吗", longitude=121.47,
        time="2026-09-09T12:00:00+08:00", rule="lost_property",
    )
    assert result["special"]["rule"] == "lost_property"
    assert result["special"]["assertions"][0]["id"] == "qimen_lost_property_direction"
