"""Contract tests for channel-specific Liki distribution packages."""

import json
import struct
import sys
from pathlib import Path

import yaml

from _helpers import ROOT

sys.path.insert(0, str(ROOT / "scripts"))
import build_channel_package as gen  # noqa: E402


def test_catalog_declares_four_packages():
    for channel, variants in gen.load_catalog()["packages"].items():
        assert set(variants) - {"maxTurns", "keywords", "plugin"} == {"free", "pro"}, channel
    assert gen.load_package("skillhub", "free")["top"] == "liki-free"
    assert gen.load_package("skillhub", "pro")["top"] == "liki-pro"
    assert gen.load_package("workbuddy", "free")["top"] == "liki-free"
    assert gen.load_package("workbuddy", "pro")["top"] == "liki-pro"
    assert gen.load_catalog()["packages"]["workbuddy"]["maxTurns"] == 40

    free = gen.load_variant("free")
    pro = gen.load_variant("pro")
    assert set(free["mcpServers"]) == {"engine", "counsel"}
    assert set(pro["mcpServers"]) == {"liki"}
    assert free["payment"]["enabled"] is False
    assert pro["payment"]["enabled"] is True


def test_workbuddy_pro_manifest_is_generated_without_source_mirror():
    files = gen.package_manifest("workbuddy", "pro")
    assert files["liki-pro/.mcp.json"]
    assert files["liki-pro/agents/liki-pro.md"]
    assert files["liki-pro/skills/liki/SKILL.md"]
    assert files["liki-pro/.codebuddy-plugin/plugin.json"]
    mcp = json.loads(files["liki-pro/.mcp.json"])
    assert set(mcp["mcpServers"]) == {"liki"}
    assert mcp["mcpServers"]["liki"]["type"] == "http"
    assert mcp["mcpServers"]["liki"]["x-workbuddy"]["auth"]["type"] == "none"
    pro_text = (
        files["liki-pro/agents/liki-pro.md"].decode()
        + files["liki-pro/skills/liki/SKILL.md"].decode()
    )
    assert "Payment-Needed" in pro_text
    assert "is_calibrated" not in pro_text
    assert "sync_ui" not in pro_text

    references = {name for name in files if name.startswith("liki-pro/skills/liki/references/")}
    canonical = {
        f"liki-pro/skills/liki/references/{path.name}"
        for path in (ROOT / "skills/liki/references").glob("*.md")
    }
    assert references == canonical
    assert not (ROOT / "distribution/workbuddy/liki-pro/skills").exists()


def test_workbuddy_free_manifest_has_no_payment_flow():
    files = gen.package_manifest("workbuddy", "free")
    for required in (
        "liki-free/.mcp.json",
        "liki-free/agents/liki-free.md",
        "liki-free/skills/liki/SKILL.md",
        "liki-free/.codebuddy-plugin/plugin.json",
        "liki-free/avatars/expert.png",
        "liki-free/skills/liki/feedback.schema.json",
        "liki-free/skills/liki/scripts/feedback.py",
    ):
        assert files[required]
    mcp = json.loads(files["liki-free/.mcp.json"])
    assert set(mcp["mcpServers"]) == {"engine", "counsel"}
    assert all(server["type"] == "http" for server in mcp["mcpServers"].values())
    rendered = files["liki-free/agents/liki-free.md"] + files["liki-free/skills/liki/SKILL.md"]
    assert b"Payment-Needed" not in rendered
    assert b"Payment-Proof" not in rendered

def test_skillhub_frontmatter_uses_platform_dialects():
    for variant, slug in (("free", "liki"), ("pro", "liki-pro")):
        package = gen.load_package("skillhub", variant)
        payload = gen.package_manifest("skillhub", variant)[f"{package['top']}/SKILL.md"]
        text = payload.decode()
        assert text.startswith("---\n")
        metadata = yaml.safe_load(text.split("---\n")[1])
        assert metadata["name"] == slug
        assert metadata["slug"] == slug
        assert metadata["displayName"] == metadata["display_name"]
        assert metadata["summary"] == metadata["description_zh"]
        assert "channel:free" not in text


def test_workbuddy_plugin_is_projected_from_market_metadata():
    for variant in ("free", "pro"):
        package = gen.load_package("workbuddy", variant)
        assets = ROOT / "distribution/workbuddy" / package["top"]
        plugin = gen.plugin_doc(variant, package["top"], assets)
        expected_display = "灵机（专业版）" if variant == "pro" else "灵机"
        expected_profession = "Liki Astrologer (Pro)" if variant == "pro" else "Liki Astrologer"
        assert plugin["displayName"]["zh"] == expected_display
        assert plugin["profession"]["en"] == expected_profession
        assert len(plugin["quickPrompts"]) == 3
        assert plugin["defaultInitPrompt"] == plugin["quickPrompts"][0]


def test_workbuddy_avatars_are_platform_safe():
    for package_id in ("liki-free", "liki-pro"):
        svg = ROOT / "distribution/workbuddy" / package_id / "avatars/expert.svg"
        png = ROOT / "distribution/workbuddy" / package_id / "avatars/expert.png"
        svg_text = svg.read_text(encoding="utf-8")
        assert "<text" not in svg_text
        assert "font-family" not in svg_text
        payload = png.read_bytes()
        assert payload[:8] == b"\x89PNG\r\n\x1a\n"
        assert struct.unpack(">II", payload[16:24]) == (512, 512)


def test_builder_writes_and_verifies_zip(tmp_path: Path):
    for channel in ("skillhub", "workbuddy"):
        for variant in ("free", "pro"):
            target = tmp_path / channel / f"liki-{variant}.zip"
            generated = gen.build(channel, variant, target)
            assert generated == target
            assert gen.check(channel, variant, target) == 0
