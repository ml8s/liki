"""仓库 CalVer 契约：分发版本与契约版本保持一致。"""
import json
from datetime import date
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
VERSION_FILES = (
    ROOT / "skills/liki-bazi/VERSION",
    ROOT / "skills/liki-divination/VERSION",
    ROOT / "skills/liki-fengshui/VERSION",
    ROOT / "skills/liki-naming/VERSION",
    ROOT / "engine/cmd/liki/VERSION",
)


def test_all_distributed_versions_are_synchronized():
    versions = {path.read_text(encoding="utf-8").strip() for path in VERSION_FILES}
    assert len(versions) == 1, versions
    assert next(iter(versions)).startswith("20")


def test_bazi_tool_and_domain_contracts_use_distributed_version():
    version = (ROOT / "skills/liki-bazi/VERSION").read_text(encoding="utf-8").strip()
    tools = json.loads(
        (ROOT / "skills/liki-bazi/tools/skill-tools.json").read_text(encoding="utf-8")
    )
    domain_contract = json.loads(
        (ROOT / "skills/liki-bazi/tools/domain_snapshot_contract.json").read_text(encoding="utf-8")
    )
    assert tools["info"]["version"] == version
    assert domain_contract["version"] == version


def test_changelog_is_project_level_and_current_version_is_calver():
    version = (ROOT / "skills/liki-bazi/VERSION").read_text(encoding="utf-8").strip()
    assert not list((ROOT / "skills").rglob("CHANGELOG.md"))
    changelog = (ROOT / "CHANGELOG.md").read_text(encoding="utf-8")
    first_heading = next(
        line for line in changelog.splitlines() if line.startswith("## ")
    )
    assert version in first_heading
    assert version.startswith(date.today().strftime("%Y.%m.%d."))


def test_project_changelog_has_no_duplicate_release_headings():
    headings = [
        line.strip()
        for line in (ROOT / "CHANGELOG.md").read_text(encoding="utf-8").splitlines()
        if line.startswith("## ")
    ]
    assert len(headings) == len(set(headings))
