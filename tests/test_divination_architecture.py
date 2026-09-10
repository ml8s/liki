from __future__ import annotations

import json
import sys
from pathlib import Path

import pytest


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))
EXPECTED_TOOLS = {
    "liuyao_snapshot",
    "liuyao_ask",
    "qimen_snapshot",
    "qimen_ask",
    "huangli_days",
}
FORBIDDEN_TOOLS = {
    "divination_route",
    "liuyao_qigua",
    "liuyao_read",
    "liuyao_report",
    "liuyao_audit",
    "liuyao_session",
    "liuyao_topic_guidance",
    "liuyao_timing",
    "liuyao_conditions",
    "qimen_read",
    "qimen_report",
    "qimen_session",
}


def test_deprecated_rpc_wrapper_is_gone():
    assert not (TOOLS / "liuyao_rpc.py").exists()
    for path in TOOLS.glob("*.py"):
        assert "liuyao_rpc" not in path.name
        assert "liuyao_rpc" not in path.read_text(encoding="utf-8")


def test_removed_orchestration_modules_are_gone():
    for name in (
        "divination_route.py",
        "divination_router.py",
        "liuyao_cast.py",
        "liuyao_read.py",
        "liuyao_projection.py",
        "liuyao_report.py",
        "liuyao_session.py",
        "qimen_read.py",
        "qimen_report.py",
        "qimen_session.py",
        "liuyao_report_contract.json",
        "liuyao_session_contract.json",
        "qimen_report_contract.json",
        "qimen_session_contract.json",
        "app/auspicious.md",
    ):
        assert not (TOOLS / name if not name.startswith("app/") else ROOT / "skills/liki-divination" / name).exists(), name


def test_only_common_rpc_module_touches_urllib():
    rpc_files = []
    for path in TOOLS.glob("*.py"):
        text = path.read_text(encoding="utf-8")
        if "urllib.request" in text or "urlopen(" in text:
            rpc_files.append(path.name)
    assert rpc_files == ["divination_rpc.py"]


def test_contract_registry_has_exact_domain_contracts():
    from divination_contracts import CONTRACT_FILES

    assert set(CONTRACT_FILES) == {
        "divination_blocked",
        "liuyao_snapshot",
        "qimen_snapshot",
        "huangli_days",
        "liuyao_answer",
        "qimen_answer",
    }
    for filename in CONTRACT_FILES.values():
        assert (TOOLS / filename).is_file()


def test_snapshot_app_cards_replace_chart_cards():
    assert (ROOT / "skills/liki-divination/app/liuyao-snapshot.md").is_file()
    assert (ROOT / "skills/liki-divination/app/qimen-snapshot.md").is_file()
    assert not (ROOT / "skills/liki-divination/app/liuyao-chart.md").exists()
    assert not (ROOT / "skills/liki-divination/app/qimen-chart.md").exists()


def test_primary_entries_use_shared_safety():
    for filename in (
        "liuyao_snapshot.py",
        "liuyao_ask.py",
        "qimen_snapshot.py",
        "qimen_ask.py",
        "huangli_days.py",
    ):
        text = (TOOLS / filename).read_text(encoding="utf-8")
        assert "divination_safety" in text, f"{filename} must use shared safety"


def test_divination_rules_are_table_driven_not_python_literals():
    huangli = (TOOLS / "huangli_days.py").read_text(encoding="utf-8")
    timing = (TOOLS / "liuyao_timing.py").read_text(encoding="utf-8")
    matters = (TOOLS / "liuyao_matters.py").read_text(encoding="utf-8")

    assert "EVENT_RULES" not in huangli
    assert "def _classify" not in huangli
    assert "BASE_PRIORITY" not in timing
    assert "HORIZONS" not in timing
    assert "TIMING_FOCUS_ID_PREFIXES" not in timing
    assert "MATTERS: dict" not in matters
    assert (TOOLS / "data/liuyao_matters.csv").is_file()
    assert (TOOLS / "liuyao_timing_rules.json").is_file()


def test_liuyao_question_uses_matter_not_domain():
    contract = json.loads((TOOLS / "liuyao_snapshot_contract.json").read_text(encoding="utf-8"))
    question = contract["properties"]["question"]
    assert "domain" not in question["properties"]
    assert {"matter", "explicit_yong_shen", "perspective"} <= set(question["properties"])


def _load_agent_cli():
    import importlib.util
    if str(TOOLS) not in sys.path:
        sys.path.insert(0, str(TOOLS))
    spec = importlib.util.spec_from_file_location("divination_agent_cli", TOOLS / "agent_cli.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def test_toolface_is_exact_and_cli_matches_schema():
    schema = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    names = {item["function"]["name"] for item in schema["tools"]}
    assert names == EXPECTED_TOOLS
    assert names.isdisjoint(FORBIDDEN_TOOLS)
    assert len(schema["tools"]) == 5

    agent_cli = _load_agent_cli()
    assert set(agent_cli._DISPATCH) == EXPECTED_TOOLS
    assert set(agent_cli._DISPATCH).isdisjoint(FORBIDDEN_TOOLS)


def test_cli_rejects_unknown_and_schema_invalid_tools_before_dispatch():
    agent_cli = _load_agent_cli()
    with pytest.raises(ValueError, match="unknown tool"):
        agent_cli._dispatch("liuyao_read", {})
    with pytest.raises(ValueError, match="invalid liuyao_snapshot args"):
        agent_cli._dispatch("liuyao_snapshot", {
            "question": "这次面试能不能通过？",
            "matter": "career",
            "yong_shen": "官鬼",
        })


def test_cli_gates_dispatch_on_engine_version(monkeypatch):
    agent_cli = _load_agent_cli()
    calls = []
    outputs = []

    def incompatible_engine():
        calls.append(True)
        raise ValueError("engine incompatible")

    monkeypatch.setattr(agent_cli, "ensure_engine_compatible", incompatible_engine)
    monkeypatch.setattr(agent_cli, "_emit", outputs.append)
    monkeypatch.setattr("sys.stdin", type("Stdin", (), {"read": lambda self: '{"fn":"huangli_days","args":{}}'})())
    assert agent_cli.main() == 0
    assert calls == [1]
    assert outputs[0]["ok"] is False
    assert "engine incompatible" in outputs[0]["error"]


def test_huangli_python_consumes_engine_event_classification():
    source = (TOOLS / "huangli_days.py").read_text(encoding="utf-8")
    assert "event" in source
    assert "_project_day" in source
    assert "suitability" in source


def test_liuyao_conflicts_come_from_engine():
    source = (TOOLS / "liuyao_factors.py").read_text(encoding="utf-8")
    assert 'chart.get("conflicts", [])' in source
    assert '"生用" in relations' not in source
    assert '"克用" in relations' not in source
    assert '"旺", "相"' not in source


def test_liuyao_condition_state_classes_are_table_driven():
    source = (TOOLS / "liuyao_conditions.py").read_text(encoding="utf-8")
    rules = json.loads((TOOLS / "liuyao_condition_rules.json").read_text(encoding="utf-8"))
    assert set(rules["state_classes"]) == {"strong", "weak"}
    assert "state_classes" in source
    assert '"休", "囚", "死"' not in source


def test_liuyao_timing_blocking_and_target_rules_are_table_driven():
    source = (TOOLS / "liuyao_timing.py").read_text(encoding="utf-8")
    rules = json.loads((TOOLS / "liuyao_timing_rules.json").read_text(encoding="utf-8"))
    assert set(rules["blocked_topics"]) == {"health_context"}
    assert set(rules["target_indicators"]["basis_tokens"]) == {"yong_shen"}
    assert '"health_context"' not in source
    assert '"yong_shen" in text' not in source


def test_divination_safety_boundary_is_table_driven():
    source = (TOOLS / "divination_safety.py").read_text(encoding="utf-8")
    rules = json.loads((TOOLS / "divination_safety_rules.json").read_text(encoding="utf-8"))
    assert rules["schema_version"] == "divination-safety-rules-v1"
    assert {rule["category"] for rule in rules["rules"]} >= {
        "self_harm", "emergency_medical", "serious_medical", "domestic_violence",
    }
    assert "RULES = [" not in source
    assert "GUIDANCE = {" not in source


def test_liuyao_chart_rpc_has_no_legacy_yaos_input():
    source = (ROOT / "engine/internal/agent/tools_other.go").read_text(encoding="utf-8")
    chart_start = source.index('Name: "liuyao.chart"')
    chart_end = source.index('Name: "huangli.days"', chart_start)
    chart_source = source[chart_start:chart_end]
    assert '"required":["solar_time","casting"]' in chart_source
    assert '"additionalProperties":false' in chart_source
    assert "兼容旧" not in chart_source
    assert "NewValuesCasting" not in chart_source
    assert "CastingMode" not in source


def test_liuyao_qigua_rpc_returns_single_casting_receipt():
    source = (ROOT / "engine/internal/agent/tools_other.go").read_text(encoding="utf-8")
    qigua_start = source.index('Name: "liuyao.qigua"')
    qigua_end = source.index('Name: "liuyao.chart"', qigua_start)
    qigua_source = source[qigua_start:qigua_end]
    assert '"required":["casting"]' in qigua_source
    assert '"additionalProperties":false' in qigua_source
    assert '"required":["yaos","dong_yao","casting"]' not in qigua_source
    assert '"casting_mode"' not in source
    assert "Seed" not in qigua_source
    assert "QiguaWithSeed" not in source


def test_time_now_has_no_test_only_seed_input():
    source = (ROOT / "engine/internal/agent/tools_other.go").read_text(encoding="utf-8")
    start = source.index('Name: "time.now"')
    end = source.index('Name: "tianwen.time"', start)
    assert '"seed"' not in source[start:end]
    assert '"additionalProperties":false' in source[start:end]


def test_qimen_standard_and_specialized_projections_are_split():
    assert (TOOLS / "qimen_projection.py").is_file()
    assert (TOOLS / "qimen_specialized.py").is_file()
    assert not (TOOLS / "qimen_factors.py").exists()
    projection = (TOOLS / "qimen_projection.py").read_text(encoding="utf-8")
    specialized = (TOOLS / "qimen_specialized.py").read_text(encoding="utf-8")
    route_names = (
        "lost_context", "thief_context", "tianwang_context",
        "geng_ge", "geng_ge_levels", "geng_ge_status",
    )
    assert not any(f'projection == "{name}"' in projection for name in route_names)
    assert all(f'projection == "{name}"' in specialized for name in route_names)


def test_qimen_specialized_core_is_engine_owned():
    source = (TOOLS / "qimen_specialized.py").read_text(encoding="utf-8")
    assert 'f"specialized.{key}"' in source
    assert "_load_jia_dun" not in source
    for forbidden in ("天蓬", "玄武", "庚", "癸"):
        assert forbidden not in source
    assert not (TOOLS / "qimen_specialized_rules.json").exists()
