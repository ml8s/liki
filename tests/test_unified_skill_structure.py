"""Structural contract for the Liki skill system（综合 liki + 专家包）。

覆盖：
1. liki 综合：根 SKILL.md + 四域（natal/divination/fengshui/naming）+ 共享运行时文件
2. 专家包（liki-bazi/liki-ziwei）：WorkBuddy 标准（plugin.json/agents/skills/.mcp.json/avatars/README）
3. MCP 连接器命名（engine-pro/engine 与专家端点）
"""
import json
from pathlib import Path

from _helpers import ROOT, SKILL_ROOT

DOMAINS = ("natal", "divination", "fengshui", "naming")
EXPERT_PACKS = ("liki-bazi", "liki-ziwei")


def test_liki_root_has_one_entry_and_four_domain_entries():
    assert (SKILL_ROOT / "SKILL.md").is_file()
    for domain in DOMAINS:
        assert (SKILL_ROOT / domain / "ENTRY.md").is_file(), domain


def test_shared_runtime_files_are_not_duplicated():
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("VERSION.txt")] == [Path("VERSION.txt")]
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("feedback.py")] == [Path("feedback.py")]
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("feedback.schema.json")] == [Path("feedback.schema.json")]


def test_python_dependency_has_one_skill_root_manifest():
    manifests = [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("requirements.txt")]
    assert manifests == [Path("requirements.txt")]


def test_faq_is_the_single_runtime_failure_entry():
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("FAQ.md")] == [Path("FAQ.md")]
    faq = (SKILL_ROOT / "FAQ.md").read_text(encoding="utf-8")
    assert "不可用" in faq
    assert "⛔" in faq and "💬" in faq
    for entry in DOMAINS:
        assert "FAQ.md" not in (SKILL_ROOT / entry / "ENTRY.md").read_text(encoding="utf-8")


def test_app_cards_use_standard_contract_sections():
    required = {"## 流程", "## 边界条件", "## 输出模板"}
    cards = [p for p in SKILL_ROOT.glob("*/app/*.md") if p.name != "README.md"]
    assert len(cards) >= 10
    for path in cards:
        lines = {line.strip() for line in path.read_text(encoding="utf-8").splitlines()}
        assert required <= lines, (path, required - lines)


def test_root_entry_is_a_lightweight_router():
    lines = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8").splitlines()
    assert len(lines) <= 160
    assert "## 统一硬边界" in lines


def test_natal_is_orchestration_layer_routing_to_experts():
    entry = (SKILL_ROOT / "natal" / "ENTRY.md").read_text(encoding="utf-8")
    assert "liki-bazi" in entry and "liki-ziwei" in entry
    assert "编排" in entry
    assert not (SKILL_ROOT / "natal" / "domains").exists()


def test_naming_does_not_silently_default_missing_hour():
    app = (SKILL_ROOT / "naming" / "app" / "naming.md").read_text(encoding="utf-8")
    calibration = (SKILL_ROOT / "naming" / "domains" / "bazi" / "calibration.md").read_text(encoding="utf-8")
    for text in (app, calibration):
        assert "默认取午时" not in text
        assert "默认时辰" not in text
    assert "不排八字" in app


def test_root_mcp_uses_counsel_and_engine():
    text = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")
    assert "counsel" in text
    assert "engine" in text
    assert "liki-bazi" in text and "liki-ziwei" in text
    mcp = json.loads((SKILL_ROOT / ".mcp.json").read_text(encoding="utf-8"))
    assert {"counsel", "engine"} <= set(mcp["mcpServers"])


def test_expert_packs_follow_workbuddy_standard():
    for pack in EXPERT_PACKS:
        root = ROOT / "skills" / pack
        assert (root / "SKILL.md").is_file(), pack
        assert (root / ".codebuddy-plugin" / "plugin.json").is_file(), pack
        assert (root / "README.md").is_file(), pack
        assert (root / "avatars").is_dir(), pack
        # plugin.json 必填
        d = json.loads((root / ".codebuddy-plugin" / "plugin.json").read_text(encoding="utf-8"))
        for f in ["name", "expertType", "version", "description", "author", "agents",
                  "agentName", "displayName", "profession", "displayDescription",
                  "avatar", "categoryId", "defaultInitPrompt", "plugin", "tags", "quickPrompts"]:
            assert f in d, (pack, f)
        assert d["expertType"] == "agent", pack
        agent_name = d["agentName"]
        assert (root / "agents" / f"{agent_name}.md").is_file(), pack
        assert d["plugin"] == d["name"], pack
        assert len(d["tags"]) == 3 and len(d["quickPrompts"]) == 3, pack
        assert d["quickPrompts"][0]["zh"] == d["defaultInitPrompt"]["zh"], pack
        n = len([c for c in d["displayDescription"]["zh"] if not c.isspace()])
        assert 40 <= n <= 50, (pack, n)
        # 连接器依赖
        deps = d.get("dependencies", {})
        assert deps.get("connectors"), (pack, deps)
        # agent frontmatter
        head = (root / "agents" / f"{agent_name}.md").read_text(encoding="utf-8")[:400]
        for f in ["name:", "description:", "displayName:", "profession:"]:
            assert f in head, (pack, agent_name, f)
        # .mcp.json 连专家端点（正交化：排盘 engine + 判断 counsel 分离）
        mcp = json.loads((root / ".mcp.json").read_text(encoding="utf-8"))
        assert len(mcp["mcpServers"]) >= 1, pack
        assert all("engine" in name or "counsel" in name for name in mcp["mcpServers"]), (pack, mcp["mcpServers"])


def test_expert_pack_methodology_matches_index():
    for pack, expect in (("liki-bazi", 16), ("liki-ziwei", 8)):
        skill = "bazi" if "bazi" in pack else "ziwei"
        cards = [p.name for p in (ROOT / "skills" / pack / "skills" / skill).glob("*.md")]
        assert "SKILL.md" in cards, pack
        # 方法论卡数量（除 SKILL.md）
        assert len(cards) - 1 == expect, (pack, len(cards) - 1, expect)


def test_no_legacy_naming_in_skills():
    legacy = ("liki-analysis", "liki-engine", "liki-master", "liki-usage", "RPC.md")
    for pack in ("liki", *EXPERT_PACKS):
        root = ROOT / "skills" / pack
        for path in root.rglob("*"):
            if not path.is_file() or "__pycache__" in path.parts or path.suffix == ".png":
                continue
            if path.suffix not in {".py", ".md", ".json", ".yaml", ".yml", ".txt"}:
                continue
            text = path.read_text(encoding="utf-8", errors="ignore")
            for old in legacy:
                assert old not in text, f"{path}: {old}"