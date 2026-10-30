"""Natal prompts describe capabilities, not concrete MCP methods or schemas."""

from pathlib import Path

import pytest
from _helpers import ROOT

SOURCES = (
    ROOT / "expert-packs/liki/SKILL.md",
    ROOT / "skills/liki/references/natal-ENTRY.md",
    ROOT / "skills/liki/references/natal-app-calibration.md",
    ROOT / "skills/liki/references/natal-bazi-ENTRY.md",
    ROOT / "skills/liki/references/natal-bazi-calibration.md",
    ROOT / "skills/liki/references/natal-ziwei-ENTRY.md",
    ROOT / "skills/liki/references/natal-ziwei-calibration.md",
)

FORBIDDEN = (
    "tools/list",
    "inputSchema",
    "outputSchema",
    "time_now",
    "city_coords",
    "tianwen_resolve",
    "bazi_chart",
    "bazi_fullchart",
    "ziwei_chart",
    "ziwei_fullchart",
)


@pytest.mark.parametrize("path", SOURCES, ids=lambda path: path.name)
def test_natal_prompts_do_not_name_methods_or_wire_fields(path: Path):
    text = path.read_text(encoding="utf-8")
    for token in FORBIDDEN:
        assert token not in text, f"{path.name} mentions {token!r}"


def test_root_identity_describes_capability_based_routing() -> None:
    text = (ROOT / "expert-packs/liki/SKILL.md").read_text(encoding="utf-8")
    assert "排盘和断语来自确定性工具" in text


def test_calibration_prompt_keeps_semantic_impact_mapping() -> None:
    text = (
        ROOT / "skills/liki/references/natal-bazi-calibration.md"
    ).read_text(encoding="utf-8")
    assert "候选原因域" in text
    for domain in ("时辰", "日界", "交节", "时制"):
        assert domain in text
