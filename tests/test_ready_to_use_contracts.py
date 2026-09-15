"""Ready-to-use payload contracts: agent should copy, not assemble."""
import json
import re
from pathlib import Path

import pytest

from helpers import SKILL_ROOT

TOOL_FILES = {
    "bazi": {
        "city_coords", "full_paipan", "query", "yearly_range",
        "calibrate", "bond",
    },
    "divination": {
        "liuyao_snapshot", "liuyao_ask", "qimen_snapshot",
        "qimen_ask", "huangli_days",
    },
}
DIRECT_RPC = {
    "naming": (
        "qiming.surname", "qiming.pick", "qiming.compose",
        "qiming.check", "qiming.char", "bazi.chart",
        "bazi.fullchart", "city.coords", "tianwen.time",
    ),
    "fengshui": (
        "time.now", "bazhai.chart", "bazhai.layout",
        "xuankong.chart", "xuankong.liunian",
    ),
}


def _sections(text: str, heading_prefix: str = "## "):
    matches = []
    lines = text.splitlines()
    starts = [i for i, line in enumerate(lines) if line.startswith(heading_prefix)]
    for index, start in enumerate(starts):
        end = starts[index + 1] if index + 1 < len(starts) else len(lines)
        title = lines[start].removeprefix(heading_prefix).strip()
        matches.append((title, "\n".join(lines[start:end])))
    return matches


def test_python_tool_domains_have_ready_stdin_payloads():
    for domain, expected_tools in TOOL_FILES.items():
        path = SKILL_ROOT / domain / "TOOLS.md"
        assert path.is_file(), domain
        text = path.read_text(encoding="utf-8")
        sections = dict(_sections(text, "## "))
        for tool in expected_tools:
            matches = [title for title in sections if re.sub(r"^\d+\.\s*", "", title) == tool]
            assert matches, (domain, tool)
        assert "python3" in text
        assert '"fn"' in text


def test_python_tool_payload_names_match_manifests():
    for domain, expected_tools in TOOL_FILES.items():
        manifest = json.loads(
            (SKILL_ROOT / domain / "tools" / "skill-tools.json").read_text(encoding="utf-8")
        )
        actual_tools = {
            tool["function"]["name"] for tool in manifest.get("tools", [])
        }
        assert expected_tools == actual_tools


def test_direct_rpc_domains_have_a_payload_for_every_method():
    for domain, methods in DIRECT_RPC.items():
        text = (SKILL_ROOT / domain / "RPC.md").read_text(encoding="utf-8")
        for method in methods:
            assert f'"method": "{method}"' in text, (domain, method)


def test_app_cards_reference_payload_libraries():
    mapping = {
        "bazi": "bazi/TOOLS.md",
        "divination": "divination/TOOLS.md",
        "naming": "naming/RPC.md",
        "fengshui": "fengshui/RPC.md",
    }
    for domain, library in mapping.items():
        cards = list((SKILL_ROOT / domain / "app").glob("*.md"))
        cards = [path for path in cards if path.name != "README.md"]
        assert cards, domain
        for path in cards:
            text = path.read_text(encoding="utf-8")
            assert library in text, (path, library)


def test_ambiguous_tool_triggers_are_absent():
    offenders = []
    for domain in ("bazi", "divination"):
        for path in (SKILL_ROOT / domain / "app").glob("*.md"):
            text = path.read_text(encoding="utf-8")
            if "必要时 `" in text or "必要时传" in text:
                offenders.append(str(path))
    assert not offenders, offenders


def test_root_has_exact_version_and_feedback_commands():
    text = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")
    assert "curl -fsS https://liki.hk/skills/liki/VERSION" in text
    assert "按点号整数逐段比较" in text
    assert "python3 feedback.py --payload-file" in text
    assert '"schema_version": "feedback-v1"' in text
    assert '"skill": "liki"' in text


def test_bazi_query_matrix_covers_manifest_rule_enum():
    manifest = json.loads(
        (SKILL_ROOT / "bazi" / "tools" / "skill-tools.json").read_text(encoding="utf-8")
    )
    query = next(
        tool["function"] for tool in manifest["tools"]
        if tool["function"]["name"] == "query"
    )
    rules = query["parameters"]["properties"]["rule"]["enum"]
    tools = (SKILL_ROOT / "bazi" / "TOOLS.md").read_text(encoding="utf-8")
    for rule in rules:
        assert f'#### query.{rule}' in tools, rule
        assert f'"rule": "{rule}"' in tools, rule


def test_bazi_app_cards_do_not_require_call_assembly():
    for path in (SKILL_ROOT / "bazi" / "app").glob("*.md"):
        text = path.read_text(encoding="utf-8")
        assert "query(rule=" not in text, path
        assert "必要时 `" not in text, path
        assert "bazi/TOOLS.md" in text, path
