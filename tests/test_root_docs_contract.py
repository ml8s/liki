"""根文档必须描述当前单层 Python 工具架构。"""
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def test_root_docs_do_not_describe_removed_rpc_discover_workflow() -> None:
    for name in ("README.md", "README.en.md", "CONTRIBUTING.md"):
        text = (ROOT / name).read_text(encoding="utf-8")
        assert "rpc.discover" not in text, name
        assert "5 工具" not in text, name
        assert "5 tools" not in text, name


def test_readme_does_not_advertise_silent_default_birth_hour() -> None:
    text = (ROOT / "README.md").read_text(encoding="utf-8")
    assert "直接用默认时辰" not in text


def test_contributing_does_not_reference_missing_version_targets() -> None:
    text = (ROOT / "CONTRIBUTING.md").read_text(encoding="utf-8")
    for target in ("version-patch", "version-minor", "version-major"):
        assert target not in text
