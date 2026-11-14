"""Trusted skill persona and forbidden retired AI identity contract."""

import json
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
SKILL_MD = ROOT / "skills/liki/SKILL.md"
PLUGIN_JSON = (
    ROOT
    / "distribution/workbuddy/liki-pro/.codebuddy-plugin/plugin.json"
)
PRO_README = ROOT / "distribution/workbuddy/liki-pro/README.md"
PRODUCT_MCP_CONFIG = ROOT / "mcp/app/config.py"

FORBIDDEN_IDENTITY = (
    "AI 命理师",
    "AI Astrologer",
    "Professional AI Astrologer",
    "Liki Fortune Master",
    "Fortune Master",
    "命理专家",
)


def _skill_metadata() -> dict:
    text = SKILL_MD.read_text(encoding="utf-8")
    assert text.startswith("---\n")
    metadata = yaml.safe_load(text.split("---\n", 2)[1])
    assert isinstance(metadata, dict)
    return metadata


def test_skill_identity_is_trusted_astrologer_persona():
    metadata = _skill_metadata()
    description = metadata.get("description") or ""
    description_en = metadata["metadata"].get("description_en") or ""
    assert metadata["metadata"]["display_name"] == "Liki 命理师"
    assert metadata["metadata"]["display_name_en"] == "Liki Astrologer"
    assert description.startswith("Liki 命理师覆盖")
    assert description_en.startswith("Liki Astrologer covers")

    text = SKILL_MD.read_text(encoding="utf-8")
    assert "\n# Liki 命理师\n" in text
    assert "你是 Liki 命理师" in text


def test_active_surfaces_do_not_use_retired_ai_identity():
    surfaces = {
        SKILL_MD,
        PRO_README,
        PLUGIN_JSON,
        ROOT / "README.md",
        ROOT / "README.zh.md",
        PRODUCT_MCP_CONFIG,
    }
    for path in surfaces:
        text = path.read_text(encoding="utf-8")
        for phrase in FORBIDDEN_IDENTITY:
            assert phrase not in text, f"{path.relative_to(ROOT)}: {phrase}"


def test_workbuddy_plugin_uses_astrologer_persona():
    plugin = json.loads(PLUGIN_JSON.read_text(encoding="utf-8"))
    assert plugin["defaultInitPrompt"] == {
        "en": "Hello, I am Liki Astrologer.",
        "zh": "你好，我是 Liki 命理师。",
    }
    assert plugin["quickPrompts"][0]["zh"] == plugin["defaultInitPrompt"]["zh"]
    for prompt in plugin["quickPrompts"]:
        assert "我是 Liki Skill" not in prompt["zh"]


def test_product_mcp_default_title_is_liki_mcp():
    text = PRODUCT_MCP_CONFIG.read_text(encoding="utf-8")
    assert 'server_title=_env("LIKI_MCP_SERVER_TITLE", "Liki MCP")' in text
    assert "Professional AI Astrologer" not in text
