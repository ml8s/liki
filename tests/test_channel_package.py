"""Contract tests for channel-specific Liki distribution packages."""
import sys
import zipfile

from _helpers import ROOT

sys.path.insert(0, str(ROOT / "scripts"))
import build_channel_package as gen


def test_channel_and_variant_configs_are_declared():
    assert gen.load_channel("skillhub", "free")[1]["artifact"] == "liki-free.zip"
    assert gen.load_channel("skillhub", "pro")[1]["artifact"] == "liki-pro.zip"
    assert gen.load_channel("workbuddy", "free")[1]["top"] == "liki-free"
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
    assert "Payment-Needed" in files["liki-pro/agents/liki-pro.md"].decode("utf-8")
    assert "is_calibrated" not in files["liki-pro/agents/liki-pro.md"].decode("utf-8")
    assert "sync_ui" not in files["liki-pro/agents/liki-pro.md"].decode("utf-8")

    references = {name for name in files if name.startswith("liki-pro/skills/liki/references/")}
    shared = {"FAQ.md", "ops-FEEDBACK.md"}
    canonical = {
        f"liki-pro/skills/liki/references/{path.name}"
        for path in (ROOT / "skills/liki/references").glob("*.md")
        if path.name not in shared
    }
    assert references == canonical
    assert not (ROOT / "distribution/workbuddy/liki-pro/skills").exists()


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


def test_builder_writes_and_verifies_zip(tmp_path):
    target = gen.build("workbuddy", "pro", tmp_path / "liki-pro.zip")
    assert target.is_file()
    assert gen.check("workbuddy", "pro", target) == 0
    with zipfile.ZipFile(target) as archive:
        assert archive.testzip() is None
        assert "liki-pro/SKILL.md" not in archive.namelist()
        assert "liki-pro/skills/liki/SKILL.md" in archive.namelist()
