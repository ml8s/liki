"""Structural contract for the unified Liki skill."""
import json
from pathlib import Path

from helpers import ROOT, SKILL_ROOT

OLD_SKILL_NAMES = {"liki-bazi", "liki-divination", "liki-fengshui", "liki-naming"}
DOMAINS = ("bazi", "divination", "fengshui", "naming")
DIRECT_RPC_SCOPES = {
    "fengshui": ("bazhai", "xuankong", "time"),
    "naming": ("qiming", "bazi.chart", "bazi.fullchart", "city", "tianwen"),
}
REQUIRED_RPC_METHODS = {
    "fengshui": (
        "bazhai.chart", "bazhai.layout", "xuankong.chart",
        "xuankong.liunian", "time.now",
    ),
    "naming": (
        "bazi.chart", "bazi.fullchart", "qiming.surname", "qiming.pick",
        "qiming.compose", "qiming.check", "qiming.char", "city.coords",
        "tianwen.time",
    ),
}


def test_skill_has_exactly_one_entry_and_four_domain_entries():
    entries = list(SKILL_ROOT.rglob("SKILL.md"))
    assert [path.relative_to(SKILL_ROOT) for path in entries] == [Path("SKILL.md")]
    for domain in DOMAINS:
        assert (SKILL_ROOT / domain / "ENTRY.md").is_file(), domain


def test_shared_runtime_files_are_not_duplicated():
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("VERSION")] == [Path("VERSION")]
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("feedback.py")] == [Path("feedback.py")]
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("feedback.schema.json")] == [Path("feedback.schema.json")]


def test_root_entry_is_a_lightweight_router():
    lines = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8").splitlines()
    assert len(lines) <= 120
    assert "## 领域路由" in lines
    assert "## 统一硬边界" in lines


def test_old_skill_names_are_absent_from_installed_content():
    for path in SKILL_ROOT.rglob("*"):
        if not path.is_file() or "__pycache__" in path.parts:
            continue
        if path.suffix not in {".py", ".md", ".json", ".yaml", ".yml", ".cmd", ".txt"}:
            continue
        text = path.read_text(encoding="utf-8", errors="ignore")
        for old in OLD_SKILL_NAMES:
            assert old not in text, f"{path}: {old}"


def test_old_skill_directories_are_gone():
    for old in OLD_SKILL_NAMES:
        assert not (ROOT / "skills" / old).exists(), old


def _rpc_contract(domain: str) -> dict:
    text = (SKILL_ROOT / domain / "RPC.md").read_text(encoding="utf-8")
    section = text.split("## 1. 固定 discover 报文", 1)[1]
    start = section.index("{")
    request, _ = json.JSONDecoder().raw_decode(section[start:])
    return request


def test_direct_rpc_domains_have_fixed_contracts():
    for domain, scopes in DIRECT_RPC_SCOPES.items():
        rpc_path = SKILL_ROOT / domain / "RPC.md"
        entry = (SKILL_ROOT / domain / "ENTRY.md").read_text(encoding="utf-8")
        assert rpc_path.is_file(), domain
        assert f"{domain}/RPC.md" in entry
        assert "不得推断其他方法或参数" in entry
        assert _rpc_contract(domain) == {
            "jsonrpc": "2.0",
            "id": f"discover-{domain}",
            "method": "rpc.discover",
            "params": {"methods": ",".join(scopes)},
        }


def test_direct_rpc_contracts_require_all_business_methods():
    for domain in DIRECT_RPC_SCOPES:
        text = (SKILL_ROOT / domain / "RPC.md").read_text(encoding="utf-8")
        for method in REQUIRED_RPC_METHODS[domain]:
            assert f'"method": "{method}"' in text, (domain, method)


def test_root_defers_discover_closure_to_domain_entries():
    text = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")
    assert "固定 discover" in text
    assert "固定 discover scope" in text
    assert "只复制领域 `RPC.md` 中的固定 discover scope 和完整 JSON-RPC 报文" in text
    assert "必须覆盖领域契约要求" in text


def test_runtime_discover_closures_match_domain_entries(monkeypatch):
    import sys

    bazi_tools = str(SKILL_ROOT / "bazi" / "tools")
    divination_tools = str(SKILL_ROOT / "divination" / "tools")
    monkeypatch.syspath_prepend(bazi_tools)
    monkeypatch.syspath_prepend(divination_tools)
    import paipan
    import divination_rpc

    assert tuple(paipan.DISCOVER_SCOPES) == ("bazi", "ziwei", "city", "tianwen", "time")
    assert tuple(paipan.REQUIRED_METHODS) == (
        "bazi.chart", "bazi.fullchart", "bazi.bond", "bazi.liunian",
        "ziwei.chart", "ziwei.fullchart", "ziwei.daxian", "ziwei.bond",
        "ziwei.liunian", "city.coords", "tianwen.time", "time.now",
    )
    assert tuple(divination_rpc.DISCOVER_SCOPES) == (
        "liuyao", "qimen", "huangli", "city", "tianwen", "time"
    )
    assert tuple(divination_rpc.REQUIRED_METHODS) == (
        "liuyao.qigua", "liuyao.chart", "qimen.chart", "huangli.days",
        "city.coords", "tianwen.time", "time.now",
    )


def test_discover_execution_ownership_is_explicit():
    for domain in ("bazi", "divination"):
        text = (SKILL_ROOT / domain / "ENTRY.md").read_text(encoding="utf-8")
        assert "不直接调用 RPC" in text
        assert "CLI 启动时校验 engine 版本和必需 RPC" in text
    for domain in ("fengshui", "naming"):
        assert (SKILL_ROOT / domain / "RPC.md").is_file()
        entry = (SKILL_ROOT / domain / "ENTRY.md").read_text(encoding="utf-8")
        assert "完整 JSON-RPC 报文" in entry


def test_naming_does_not_silently_default_missing_hour():
    app = (SKILL_ROOT / "naming" / "app" / "naming.md").read_text(encoding="utf-8")
    calibration = (SKILL_ROOT / "naming" / "domains" / "bazi" / "calibration.md").read_text(encoding="utf-8")
    for text in (app, calibration):
        assert "按午时排盘" not in text
        assert "默认取午时" not in text
        assert "默认时辰" not in text
    assert "不排八字" in app
    assert "未评估用神" in app
    assert "不得使用午时" in calibration
    assert "期望五行策略" in calibration


def test_docs_describe_domain_specific_rpc_boundaries():
    zh = (ROOT / "README.md").read_text(encoding="utf-8")
    en = (ROOT / "README.en.md").read_text(encoding="utf-8")
    package = (ROOT / "docs" / "SKILL_PACKAGE.md").read_text(encoding="utf-8")
    assert "RPC 方法对 LLM 不可见" not in zh
    assert "RPC methods remain invisible" not in en
    assert "无 `tools/` 的领域" in package
    assert "固定 discover scope" in package
