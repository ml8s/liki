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


def test_readmes_explain_qimen_default_and_advanced_usage() -> None:
    zh = (ROOT / "README.md").read_text(encoding="utf-8")
    en = (ROOT / "README.en.md").read_text(encoding="utf-8")

    assert "默认使用时家奇门、转盘法、拆补定局" in zh
    assert "无需选择排盘方法" in zh
    assert "用置闰盘看现在" in zh
    assert "用十二分钟十分局看看" in zh
    assert "用金函玉镜看今天" in zh

    assert "defaults to hour-scope Qimen" in en
    assert "no chart-method selection" in en.replace("**", "")
    assert "Use the zhirun chart" in en
    assert "Use the twelve-minute ten-division chart" in en
    assert "Use Golden Mirror for today" in en


def test_readmes_do_not_expose_qimen_project_names() -> None:
    for name in ("README.md", "README.en.md"):
        text = (ROOT / name).read_text(encoding="utf-8").lower()
        for banned in ("horosa", "kinqimen", "atopx", "potuo", "mingyu", "deminzhang"):
            assert banned not in text, (name, banned)


def test_generated_reports_and_duplicate_architecture_docs_stay_out_of_source() -> None:
    assert not (ROOT / "docs/FACTOR_LAYER_DESIGN.md").exists()
    assert not (ROOT / "tests/schema.md").exists()
    assert not (ROOT / "tests/RESULTS.md").exists()

    evaluator = (ROOT / "tests/eval_hybrid.py").read_text(encoding="utf-8")
    assert "RESULTS.md" not in evaluator
    assert "--output" in evaluator
