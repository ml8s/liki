"""Contract tests for channel-specific Liki distribution packages."""
import json
import sys
import zipfile

import yaml
from _helpers import ROOT

sys.path.insert(0, str(ROOT / "scripts"))
import build_channel_package as gen


def test_channel_and_variant_configs_are_declared():
    assert gen.load_channel("skillhub", "free")[1]["artifact"] == "liki-free.zip"
    assert gen.load_channel("skillhub", "pro")[1]["artifact"] == "liki-pro.zip"
    assert gen.load_channel("workbuddy", "free")[1]["top"] == "liki-free"
    assert gen.load_channel("workbuddy", "free")[1]["expert"] is True
    assert gen.load_channel("workbuddy", "pro")[1]["expert"] is True

    free = gen.load_variant("free")
    pro = gen.load_variant("pro")
    assert free["sku"] == "liki-free"
    assert pro["sku"] == "liki-pro"
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
    assert "Payment-Needed" in files["liki-pro/agents/liki-pro.md"].decode("utf-8")
    assert "Payment-Needed" in files["liki-pro/skills/liki/SKILL.md"].decode("utf-8")
    assert "channel:free" not in files["liki-pro/skills/liki/SKILL.md"].decode("utf-8")
    assert "is_calibrated" not in files["liki-pro/agents/liki-pro.md"].decode("utf-8")
    assert "sync_ui" not in files["liki-pro/agents/liki-pro.md"].decode("utf-8")

    references = {name for name in files if name.startswith("liki-pro/skills/liki/references/")}
    canonical = {
        f"liki-pro/skills/liki/references/{path.name}"
        for path in (ROOT / "skills/liki/references").glob("*.md")
    }
    assert references == canonical
    assert not (ROOT / "distribution/workbuddy/liki-pro/skills").exists()


def test_workbuddy_free_manifest_is_an_expert_package():
    files = gen.package_manifest("workbuddy", "free")
    assert files["liki-free/.mcp.json"]
    assert files["liki-free/agents/liki-free.md"]
    assert files["liki-free/skills/liki/SKILL.md"]
    assert files["liki-free/.codebuddy-plugin/plugin.json"]
    assert files["liki-free/avatars/expert.png"]
    assert files["liki-free/skills/liki/references/ops-FEEDBACK.md"]
    assert files["liki-free/skills/liki/feedback.schema.json"]
    assert files["liki-free/skills/liki/scripts/feedback.py"]
    mcp = json.loads(files["liki-free/.mcp.json"])
    assert set(mcp["mcpServers"]) == {"engine", "counsel"}
    assert all(server["type"] == "http" for server in mcp["mcpServers"].values())
    assert all(server["x-workbuddy"]["auth"]["type"] == "none" for server in mcp["mcpServers"].values())
    assert "Payment-Needed" not in files["liki-free/agents/liki-free.md"].decode("utf-8")
    assert "Payment-Needed" not in files["liki-free/skills/liki/SKILL.md"].decode("utf-8")
    assert "channel:free" not in files["liki-free/skills/liki/SKILL.md"].decode("utf-8")


def test_packages_do_not_include_local_build_residue():
    for channel in ("skillhub", "workbuddy"):
        for variant in ("free", "pro"):
            files = gen.package_manifest(channel, variant)
            for name in files:
                assert "__pycache__/" not in name, (channel, variant, name)
                assert not name.endswith((".pyc", ".pyo")), (channel, variant, name)


def test_free_packages_do_not_include_payment_overlay():
    for channel in ("skillhub", "workbuddy"):
        files = gen.package_manifest(channel, "free")
        rendered = b"\n".join(
            payload
            for name, payload in files.items()
            if name.endswith("SKILL.md") or name.endswith("liki-free.md")
        )
        assert "Payment-Needed" not in rendered.decode("utf-8")
        assert "Payment-Proof" not in rendered.decode("utf-8")


def _frontmatter(payload: bytes) -> dict:
    text = payload.decode("utf-8")
    assert text.startswith("---\n")
    return yaml.safe_load(text.split("---\n")[1])


def test_channel_frontmatter_uses_platform_dialects():
    for variant, slug in (("free", "liki"), ("pro", "liki-pro")):
        top = gen.load_channel("skillhub", variant)[1]["top"]
        payload = gen.package_manifest("skillhub", variant)[f"{top}/SKILL.md"]
        meta = _frontmatter(payload)
        text = payload.decode("utf-8")
        assert meta["name"] == slug
        assert meta["slug"] == slug
        assert "channel:free" not in text
        assert "ops-FEEDBACK.md" in text
        if variant == "pro":
            assert "Payment-Needed" in text
        else:
            assert "Payment-Needed" not in text
        assert meta["displayName"] == meta["display_name"]
        assert meta["summary"] == meta["description_zh"]
        assert meta["author"] == "ml8s"
        assert meta["license"] == "MIT"
        assert "metadata" not in meta
        assert "agent_created" not in meta

        top = gen.load_channel("workbuddy", variant)[1]["top"]
        meta = _frontmatter(
            gen.package_manifest("workbuddy", variant)[f"{top}/skills/liki/SKILL.md"]
        )
        assert meta["name"] == "liki"
        assert meta["display_name"] == "Liki 命理师" + ("" if variant == "free" else "（专业版）")
        assert meta["agent_created"] is True
        assert meta["author"] == "ml8s"
        assert "slug" not in meta
        assert "displayName" not in meta
        assert "summary" not in meta
        assert "metadata" not in meta


def test_workbuddy_plugin_is_projected_from_market_metadata():
    for variant in ("free", "pro"):
        top = gen.load_channel("workbuddy", variant)[1]["top"]
        assets_dir = gen.WORKBUDDY_DIR / f"liki-{variant}"
        market = gen.load_market("workbuddy", variant, assets_dir)
        files = gen.package_manifest("workbuddy", variant)
        plugin = json.loads(files[f"{top}/.codebuddy-plugin/plugin.json"])
        assert plugin["displayName"] == {
            "en": market["agent_display_name_en"],
            "zh": market["agent_display_name_zh"],
        }
        assert plugin["profession"] == {
            "en": market["profession_en"],
            "zh": market["profession_zh"],
        }
        assert plugin["keywords"] == market["keywords"]
        assert len(plugin["quickPrompts"]) == 3
        assert plugin["quickPrompts"][0] == plugin["defaultInitPrompt"]
        assert plugin["quickPrompts"][0]["zh"] == "帮我看八字命盘"


def test_builder_writes_and_verifies_zip(tmp_path):
    for variant in ("free", "pro"):
        target = gen.build("workbuddy", variant, tmp_path / f"liki-{variant}.zip")
        assert target.is_file()
        assert gen.check("workbuddy", variant, target) == 0
        with zipfile.ZipFile(target) as archive:
            assert archive.testzip() is None
            assert f"liki-{variant}/SKILL.md" not in archive.namelist()
            assert f"liki-{variant}/skills/liki/SKILL.md" in archive.namelist()
