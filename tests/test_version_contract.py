"""分发版本与契约版本的一致性契约。"""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
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
        (ROOT / "counsel/app/natal/tools/skill-tools.json").read_text(encoding="utf-8")
    )
    domain_contract = json.loads(
        (ROOT / "counsel/app/natal/tools/natal_projection_contract.json").read_text(encoding="utf-8")
    )
    assert tools["info"]["version"] == version
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
