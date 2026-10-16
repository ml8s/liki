"""校验分发版本面保持一致。"""
import json
import sys
from pathlib import Path

import pytest
import yaml

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
from version_projection import workbuddy_version  # noqa: E402


def _runtime_calver() -> str:
    return (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()


def _artifact_version() -> str:
    return workbuddy_version(_runtime_calver())
VERSION_FILES = (
    ROOT / "skills/liki/VERSION.txt",
    ROOT / "engine/cmd/engine-mcp/VERSION",
    ROOT / "counsel/app/VERSION.txt",
)


def test_all_distributed_versions_are_synchronized():
    versions = {path.read_text(encoding="utf-8").strip() for path in VERSION_FILES}
    assert len(versions) == 1, versions
    version = next(iter(versions))
    assert version.startswith("20")
    parts = version.split(".")
    assert len(parts) == 4
    assert all(part.isdigit() for part in parts)
    for relative in ("pyproject.toml", "counsel/pyproject.toml"):
        text = (ROOT / relative).read_text(encoding="utf-8")
        assert f'version = "{version}"' in text, relative


def test_counsel_contracts_use_distributed_version():
    from app.contracts import VERSION

    version = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()
    assert VERSION == version


def test_bazi_domain_contract_uses_distributed_version():
    version = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()
    domain_contract = json.loads(
        (ROOT / "counsel/app/natal/tools/natal_projection_contract.json").read_text(encoding="utf-8")
    )
    assert domain_contract["version"] == version


def test_divination_projection_contract_uses_distributed_version():
    version = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()
    projection_contract = json.loads(
        (ROOT / "counsel/app/divination/tools/qimen_projection_contract.json").read_text(encoding="utf-8")
    )
    assert projection_contract["version"] == version


def test_deployment_surfaces_use_distributed_version():
    version = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()
    for path in sorted((ROOT / "profiles").glob("*.json")):
        document = json.loads(path.read_text(encoding="utf-8"))
        assert document["metadata"]["version"] == version, path


def test_readme_image_identity_uses_immutable_digest():
    for relative in ("README.md", "README.en.md"):
        text = (ROOT / relative).read_text(encoding="utf-8")
        assert "liki-engine@sha256:" in text, relative
        assert "liki-engine:2026." not in text, relative


def test_expert_pack_plugin_versions_use_platform_projection():
    """专家包是平台制品：包内版本面全是发行 CalVer 的 semver 投影（不手写、不与 CalVer 混用）。"""
    artifact_version = _artifact_version()
    for plugin in sorted((ROOT / "expert-packs").glob("*/.codebuddy-plugin/plugin.json")):
        pack = plugin.parents[1]
        plugin_version = json.loads(plugin.read_text(encoding="utf-8"))["version"]
        skill_text = (pack / "SKILL.md").read_text(encoding="utf-8")
        skill_version = yaml.safe_load(skill_text.split("---")[1])["version"]
        assert skill_version == artifact_version, pack
        assert plugin_version == skill_version, pack
        for domain_skill in sorted((pack / "skills").glob("*/SKILL.md")):
            text = domain_skill.read_text(encoding="utf-8")
            domain_version = yaml.safe_load(text.split("---")[1])["metadata"]["version"]
            assert domain_version == plugin_version, domain_skill


def test_skill_and_connector_versions_use_platform_projection():
    artifact_version = _artifact_version()
    skill_text = (ROOT / "skills/liki/SKILL.md").read_text(encoding="utf-8")
    assert yaml.safe_load(skill_text.split("---\n")[1])["version"] == artifact_version
    connectors = sorted((ROOT / "workbuddy").glob("*-connector/connector-meta.json"))
    assert connectors, "缺少连接器提交源"
    for meta_path in connectors:
        assert json.loads(meta_path.read_text(encoding="utf-8"))["version"] == artifact_version, meta_path


def test_workbuddy_version_projection_is_semver_and_monotonic():
    """投影必须是合法 semver、跨日跨年严格递增、且能还原回月日。"""
    semver = __import__("re").compile(r"^\d+\.\d+\.\d+$")
    assert semver.match(_artifact_version())
    for calver in ("2026.10.16.2", "2026.01.01.0", "2026.02.01.0", "2027.01.01.0"):
        assert semver.match(workbuddy_version(calver)), calver
    ordered = [
        workbuddy_version(v)
        for v in ("2026.10.16.0", "2026.10.16.5", "2026.10.17.0", "2026.11.01.0", "2027.01.01.0")
    ]
    assert ordered == sorted(ordered, key=lambda v: tuple(int(x) for x in v.split("."))), ordered


def test_workbuddy_version_projection_rejects_non_calver():
    with pytest.raises(ValueError):
        workbuddy_version("1.0.0")
