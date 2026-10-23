"""校验分发版本面保持一致。"""
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
from version_projection import workbuddy_version  # noqa: E402


def _runtime_calver() -> str:
    return (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()


def _artifact_version() -> str:
    return workbuddy_version(_runtime_calver())
VERSION_FILES = (
    ROOT / "skills/liki/VERSION.txt",
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


def test_deployment_surfaces_use_distributed_version():
    version = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()
    for path in sorted((ROOT / "profiles").glob("*.json")):
        document = json.loads(path.read_text(encoding="utf-8"))
        assert document["metadata"]["version"] == version, path
