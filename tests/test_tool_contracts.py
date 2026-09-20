"""Bazi tool parameter contract tests."""
import json
from pathlib import Path

import pytest
import _helpers  # noqa: F401 —— 注入 tools 路径
from agent_cli import _REQUIRED_ARGS

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = json.loads(
    (ROOT / "skills/liki/bazi/tools/skill-tools.json").read_text(encoding="utf-8")
)
FUNCTIONS = {tool["function"]["name"]: tool["function"] for tool in MANIFEST["tools"]}


def test_all_tools_have_closed_args():
    required = {
        "city_coords": ["city"],
        "full_paipan": ["gregorian", "gender"],
        "query": ["rule", "pan", "domains"],
        "yearly_range": ["pan", "start", "end", "rules", "domains"],
        "calibrate": ["candidates", "events"],
        "bond": ["pan_a", "pan_b"],
    }
    for name, fn in FUNCTIONS.items():
        assert fn["parameters"]["additionalProperties"] is False, name
        assert fn["parameters"]["required"] == required[name], name
        assert fn["parameters"]["type"] == "object", name
