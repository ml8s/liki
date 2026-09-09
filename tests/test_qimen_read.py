from __future__ import annotations

import json
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import qimen_read  # noqa: E402


def test_read_resolves_city_time_chart_and_snapshot(monkeypatch):
    calls = []

    def fake_engine_data(method, params):
        calls.append((method, params))
        if method == "time.now":
            return {"cst": "2026-09-08T12:00:00+08:00"}
        raise AssertionError(method)

    def fake_city(city):
        calls.append(("city.coords", {"city": city}))
        return {"name": city, "longitude": 121.47, "latitude": 31.23}

    def fake_solar(value, longitude):
        calls.append(("tianwen.time", {"time": value, "longitude": longitude}))
        return {"solar": "2026-09-08T11:57:00+08:00"}

    def fake_chart(solar, **kwargs):
        calls.append(("qimen.chart", {"solar_time": solar, **kwargs}))
        return {"matter": {"matter": "wealth"}, "chart": {"method": {"scope": "hour", "school": "zhuanpan"}}}

    monkeypatch.setattr(qimen_read, "engine_data", fake_engine_data)
    monkeypatch.setattr(qimen_read, "city_coords", fake_city)
    monkeypatch.setattr(qimen_read, "solar_time", fake_solar)
    monkeypatch.setattr(qimen_read, "qimen_chart", fake_chart)
    monkeypatch.setattr(qimen_read, "project_qimen_snapshot", lambda pan: {"snapshot": True})

    result = qimen_read.read(question="该往哪里推进？", city="上海", matter="wealth")
    assert result["input"]["city"] == "上海"
    assert result["input"]["solar_time"] == "2026-09-08T11:57:00+08:00"
    assert result["chart"]["method"]["scope"] == "hour"
    assert result["snapshot"] == {"snapshot": True}
    assert [item[0] for item in calls] == [
        "time.now", "city.coords", "tianwen.time", "qimen.chart",
    ]


def test_read_requires_location():
    import pytest

    with pytest.raises(ValueError, match="city 或 longitude"):
        qimen_read.read(question="现在该不该行动？", time="2026-09-08T12:00:00+08:00")


def test_qimen_read_schema_accepts_primary_call():
    from jsonschema import validate

    schema = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    tool = next(
        item["function"]
        for item in schema["tools"]
        if item["function"]["name"] == "qimen_read"
    )
    validate(
        {
            "question": "该往哪里推进？",
            "city": "上海",
            "matter": "wealth",
        },
        tool["parameters"],
    )
