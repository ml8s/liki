"""根文档禁止回退到已移除的排盘流程或暴露内部项目名。"""
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


def test_readmes_do_not_expose_internal_qimen_project_names() -> None:
    for name in ("README.md", "README.en.md"):
        text = (ROOT / name).read_text(encoding="utf-8").lower()
        for banned in ("horosa", "kinqimen", "atopx", "potuo", "mingyu", "deminzhang"):
            assert banned not in text, (name, banned)
