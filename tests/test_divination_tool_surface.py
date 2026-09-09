from __future__ import annotations

import importlib.util
import json
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))


def _load_agent_cli():
    spec = importlib.util.spec_from_file_location("divination_agent_cli", TOOLS / "agent_cli.py")
    module = importlib.util.module_from_spec(spec)
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module


def test_llm_toolface_uses_qimen_read_not_internal_chart_tools():
    schema = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    names = {item["function"]["name"] for item in schema["tools"]}
    assert "qimen_read" in names
    assert "qimen_chart" not in names
    assert "query" not in names
    assert "liuyao_chart" not in names

    agent_cli = _load_agent_cli()
    assert "qimen_read" in agent_cli._DISPATCH
    assert {"qimen_chart", "query", "liuyao_chart"}.isdisjoint(agent_cli._DISPATCH)


def test_internal_qimen_orchestration_still_uses_chart_layer(monkeypatch):
    import qimen_paipan
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

    monkeypatch.setattr(qimen_paipan, "engine_data", fake_engine_data)
    monkeypatch.setattr(qimen_paipan, "city_coords", fake_city)
    monkeypatch.setattr(qimen_paipan, "solar_time", fake_solar)
    monkeypatch.setattr(qimen_paipan, "qimen_chart", fake_chart)
    qimen_read = _load_module("qimen_read")
    monkeypatch.setattr(qimen_read, "project_qimen_snapshot", lambda pan: {})

    result = qimen_read.read(
        question="当前态势",
        city="上海",
        time="2026-09-09T12:00:00+08:00",
    )
    assert "qimen.chart" in calls
    assert result["input"]["solar_time"] == "2026-09-09T11:57:00+08:00"


def _load_module(name: str):
    spec = importlib.util.spec_from_file_location(name, TOOLS / f"{name}.py")
    module = importlib.util.module_from_spec(spec)
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module
