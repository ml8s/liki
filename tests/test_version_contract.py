"""分发版本与契约版本的一致性契约。"""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
VERSION_FILES = (
    ROOT / "skills/liki/VERSION.txt",
    ROOT / "engine/cmd/engine-mcp/VERSION",
    ROOT / "engine/cmd/engine-rpc/VERSION",
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


def test_counsel_manifests_use_distributed_version():
    version = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()
    counsel_tools = json.loads(
        (ROOT / "counsel/app/natal/tools/counsel-tools.json").read_text(encoding="utf-8")
    )
    naming_tools = json.loads(
        (ROOT / "counsel/app/naming/tools/skill-tools.json").read_text(encoding="utf-8")
    )
    assert counsel_tools["version"] == version
    assert naming_tools["info"]["version"] == version


def test_bazi_tool_and_domain_contracts_use_distributed_version():
    version = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()
    tools = json.loads(
        (ROOT / "counsel/app/natal/tools/counsel-tools.json").read_text(encoding="utf-8")
    )
    domain_contract = json.loads(
        (ROOT / "counsel/app/natal/tools/natal_projection_contract.json").read_text(encoding="utf-8")
    )
    assert tools["version"] == version
    assert domain_contract["version"] == version


def test_divination_tool_and_projection_contracts_use_distributed_version():
    version = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()
    tools = json.loads(
        (ROOT / "counsel/app/divination/tools/skill-tools.json").read_text(encoding="utf-8")
    )
    projection_contract = json.loads(
        (ROOT / "counsel/app/divination/tools/qimen_projection_contract.json").read_text(encoding="utf-8")
    )
    assert tools["info"]["version"] == version
    assert projection_contract["version"] == version


def test_deployment_and_catalog_surfaces_use_distributed_version():
    version = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()
    catalog = json.loads(
        (ROOT / "contracts/mcp-tool-catalog.json").read_text(encoding="utf-8")
    )
    assert catalog["runtime_version"] == version
    for path in sorted((ROOT / "profiles").glob("*.json")):
        document = json.loads(path.read_text(encoding="utf-8"))
        assert document["metadata"]["version"] == version, path


def test_readme_image_tags_use_distributed_version():
    version = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()
    for relative in ("README.md", "README.en.md"):
        text = (ROOT / relative).read_text(encoding="utf-8")
        assert f"liki-engine:{version}" in text, relative


def test_expert_pack_plugin_versions_internally_consistent():
    for plugin in sorted((ROOT / "expert-packs").glob("*/.codebuddy-plugin/plugin.json")):
        pack = plugin.parents[1]
        plugin_version = json.loads(plugin.read_text(encoding="utf-8"))["version"]
        skill_text = (pack / "SKILL.md").read_text(encoding="utf-8")
        skill_version = next(
            line.split(":", 1)[1].strip()
            for line in skill_text.splitlines()
            if line.startswith("version:")
        )
        assert plugin_version == skill_version, pack
        for domain_skill in sorted((pack / "skills").glob("*/SKILL.md")):
            text = domain_skill.read_text(encoding="utf-8")
            domain_version = next(
                line.split(":", 1)[1].strip()
                for line in text.splitlines()
                if line.strip().startswith("version:")
            )
            assert domain_version == plugin_version, domain_skill
