from __future__ import annotations

import json
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
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
        "divination_router.py",
        "liuyao_read.py",
        "liuyao_session.py",
        "qimen_read.py",
        "qimen_session.py",
    ):
        assert not (TOOLS / name).exists(), name


def test_only_common_rpc_module_touches_urllib():
    rpc_files = []
    for path in TOOLS.glob("*.py"):
        text = path.read_text(encoding="utf-8")
        if "urllib.request" in text or "urlopen(" in text:
            rpc_files.append(path.name)
    assert rpc_files == ["divination_rpc.py"]


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
