import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))

from build_channel_package import skill_doc  # noqa: E402
from version_projection import workbuddy_version  # noqa: E402


def test_packaged_skill_projects_release_version() -> None:
    runtime = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()
    frontmatter = skill_doc("free", "liki").split("---\n")[1]
    assert f"version: {workbuddy_version(runtime)}" in frontmatter
