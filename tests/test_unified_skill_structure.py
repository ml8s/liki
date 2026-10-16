"""Structural contract for the Liki skill system（综合 liki + 专家包）。

覆盖：
1. liki 综合：根 SKILL.md + 四域（references/natal/references/divination/references/fengshui/naming）+ 共享运行时文件
2. 专家包（四个顶层领域）：WorkBuddy 标准（SKILL.md 身份源 / plugin.json / skills / .mcp.json / avatars / README）
3. MCP 连接器命名（engine/counsel 与专家端点）
"""
import json
import re
from pathlib import Path

import yaml
from _helpers import ROOT, SKILL_ROOT

DOMAINS = ("natal", "divination", "fengshui", "naming")
# 顶层领域 = 专家包（1:1）：包内 references/ 是本域根子树（ENTRY/TOOLS/app/domains）的镜像。
EXPERT_PACKS = ("liki-natal", "liki-divination", "liki-fengshui", "liki-naming")


def test_liki_root_has_one_entry_and_four_domain_entries():
    assert (SKILL_ROOT / "SKILL.md").is_file()
    for domain in DOMAINS:
        assert (SKILL_ROOT / "references" / domain / "ENTRY.md").is_file(), domain


def test_shared_runtime_files_are_not_duplicated():
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("VERSION.txt")] == [Path("VERSION.txt")]
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("feedback.py")] == [Path("scripts/feedback.py")]
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("feedback.schema.json")] == [Path("feedback.schema.json")]


def test_python_dependency_has_one_skill_root_manifest():
    manifests = [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("requirements*.txt")]
    assert manifests == []


def test_faq_is_the_single_runtime_failure_entry():
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("FAQ.md")] == [Path("references/FAQ.md")]
    faq = (SKILL_ROOT / "references" / "FAQ.md").read_text(encoding="utf-8")
    assert "不可用" in faq
    assert "⛔" in faq and "💬" in faq
    for entry in DOMAINS:
        assert "FAQ.md" not in (SKILL_ROOT / "references" / entry / "ENTRY.md").read_text(encoding="utf-8")


def test_app_cards_use_standard_contract_sections():
    required = {"## 流程", "## 边界条件", "## 输出模板"}
    cards = list(SKILL_ROOT.glob("references/*/app/*.md"))
    assert len(cards) >= 10
    for path in cards:
        lines = {line.strip() for line in path.read_text(encoding="utf-8").splitlines()}
        assert required <= lines, (path, required - lines)


def test_shared_yongshen_priority_order_matches():
    """用神取舍优先序在 natal / naming 两域同源：专家包自包含 → 双份必须同字。"""
    line = "顺序：从格 → 调候 → 格局清透 → 普通命局 → 合并。"
    for rel in (
        "references/natal/domains/bazi/yongshen.md",
        "references/naming/domains/bazi/yongshen.md",
    ):
        assert line in (SKILL_ROOT / rel).read_text(encoding="utf-8"), rel


def test_app_cards_declare_output_template_body():
    """`## 输出模板` 必须给出模板体（代码块 / 表 / 列表），不能只有示例。"""
    for card in sorted(SKILL_ROOT.glob("references/*/app/*.md")):
        text = card.read_text(encoding="utf-8")
        match = re.search(r"^## 输出模板\s*$(.*?)(?=^## |\Z)", text, re.M | re.S)
        assert match, card
        before = match.group(1).split("### 示例")[0]
        assert (
            "```" in before
            or re.search(r"^\| ?", before, re.M)
            or re.search(r"^- ", before, re.M)
        ), card


def test_root_entry_is_a_lightweight_router():
    lines = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8").splitlines()
    assert len(lines) <= 120
    assert "## 统一硬边界" in lines


def test_natal_is_self_contained_orchestration_layer():
    """natal 自举八字侧/紫微侧：域内路由只按侧编排。"""
    entry = (SKILL_ROOT / "references" / "natal" / "ENTRY.md").read_text(encoding="utf-8")
    assert "八字侧" in entry and "紫微侧" in entry
    assert "编排" in entry
    assert "专家" not in entry
    tools = (SKILL_ROOT / "references" / "natal" / "TOOLS.md").read_text(encoding="utf-8")
    assert "## 输出契约（断语公共字段）" in tools
    # 域内路由只由 ENTRY.md 定义一次（TOOLS.md 不重复编排表）
    assert "## 路由" in entry and "合盘（双人）" not in tools
    # 根 liki 是方法论权威源：references/natal/domains/ 含 bazi/ziwei 方法论卡
    assert (SKILL_ROOT / "references" / "natal" / "domains" / "bazi").is_dir()
    assert (SKILL_ROOT / "references" / "natal" / "domains" / "ziwei").is_dir()


def test_natal_tools_declares_assertion_output_contract():
    """natal 断语公共字段、side 取值与受控闭集是硬契约，不得回退。"""
    tools = (SKILL_ROOT / "references" / "natal" / "TOOLS.md").read_text(encoding="utf-8")
    for field in (
        "assertion_id", "side", "topic", "method", "time_scope",
        "event_type", "event", "conclusion", "source", "evidence",
    ):
        assert f"`{field}`" in tools, field
    assert "`side` 是 `bazi` / `ziwei`" in tools
    for value in ("career", "marriage", "study", "wealth", "chart_structure", "relocation"):
        assert f"`{value}`" in tools, value
    for scope in ("current_year", "year_range", "current_decade", "decade"):
        assert f"`{scope}`" in tools, scope
    # topic 键＝标签与服务器路由表单一源一致（卡内 `本命判断（事业）` 靠它对上 topic）
    routes = json.loads(
        (ROOT / "counsel" / "app" / "natal" / "tools" / "topic_routes.json").read_text(encoding="utf-8")
    )
    for key, spec in routes["topics"].items():
        assert f"`{key}`＝{spec['label']}" in tools, (key, spec["label"])


def test_naming_does_not_silently_default_missing_hour():
    app = (SKILL_ROOT / "references" / "naming" / "app" / "naming.md").read_text(encoding="utf-8")
    calibration = (SKILL_ROOT / "references" / "naming" / "domains" / "bazi" / "calibration.md").read_text(encoding="utf-8")
    for text in (app, calibration):
        assert "默认取午时" not in text
        assert "默认时辰" not in text
    assert "不排八字" in app


def test_natal_card_qualifiers_are_topic_labels():
    """卡内 `本命判断（X）` / `应期判断（X）` 的 X 必须是受控 topic 标签或十年期层。"""
    routes = json.loads(
        (ROOT / "counsel" / "app" / "natal" / "tools" / "topic_routes.json").read_text(encoding="utf-8")
    )
    labels = {spec["label"] for spec in routes["topics"].values()} | {"十年期"}
    qualifier = re.compile(r"`(?:本命判断|应期判断)（([^）]+)）`")
    offenders = []
    for card in sorted((SKILL_ROOT / "references" / "natal" / "app").glob("*.md")):
        for token in qualifier.findall(card.read_text(encoding="utf-8")):
            if token not in labels:
                offenders.append(f"{card.name}: {token}")
    assert offenders == []


def test_natal_card_actions_name_the_side():
    """卡内动作的每个能力调用都要显式带侧（counsel side 默认 bazi，省略即静默降级）。"""
    capability = re.compile(r"`(?:本命判断|应期判断)")
    offenders = []
    for card in sorted((SKILL_ROOT / "references" / "natal" / "app").glob("*.md")):
        for number, line in enumerate(card.read_text(encoding="utf-8").splitlines(), 1):
            if not line.startswith("|") or line.startswith("| ---"):
                continue
            for cell in line.split("|")[1:-1]:
                for clause in cell.split("；"):
                    if capability.search(clause) and "八字侧" not in clause and "紫微侧" not in clause:
                        offenders.append(f"{card.name}:{number}")
    assert offenders == []


def test_skill_docs_use_mcp_result_shape():
    """MCP 结果形态：成功读工具内容 JSON、失败读 isError；不得回退到已退役的 data/error 信封。"""
    offenders = []
    for path in sorted((SKILL_ROOT / "references").rglob("*.md")):
        for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            if "响应的 `data`" in line or "成功读 `data`" in line or "失败读 `error`" in line:
                offenders.append(f"{path.relative_to(SKILL_ROOT)}:{number}")
    assert offenders == []


def test_root_mcp_uses_counsel_and_engine():
    text = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")
    assert "counsel" in text
    assert "engine" in text
    # 路由按领域（natal/divination/naming/fengshui），不按专家包 id
    for domain in DOMAINS:
        assert domain in text
    # 路由按领域（natal/divination/fengshui/naming），不按专家包 id
    assert not re.search(r"liki-(?:natal|divination|fengshui|naming)", text)
    mcp = json.loads((SKILL_ROOT / ".mcp.json").read_text(encoding="utf-8"))
    assert any("counsel" in name for name in mcp["mcpServers"])
    # engine 域连接器在专家包（本命包同时挂八字侧与紫微侧排盘端点）
    for pack in ("liki-natal",):
        expert_mcp = json.loads((ROOT / "expert-packs" / pack / ".mcp.json").read_text(encoding="utf-8"))
        assert any("engine" in name for name in expert_mcp["mcpServers"])


def test_root_skill_references_no_expert_pack_ids():
    """自包含契约：根 skill 文本不得引用专家包 id（路由只按领域 + 侧）。"""
    pack_id = re.compile(r"liki-(?:bazi|ziwei|liuyao|qimen|fengshui|naming)")
    offenders = []
    for path in sorted(SKILL_ROOT.rglob("*")):
        if not path.is_file() or "__pycache__" in path.parts:
            continue
        if path.suffix not in {".py", ".md", ".json", ".yaml", ".yml", ".txt"}:
            continue
        for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            if pack_id.search(line):
                offenders.append(f"{path.relative_to(SKILL_ROOT)}:{number}")
    assert offenders == []


def test_root_mcp_declares_manual_install_surface():
    """Embedded and manual clients get the two aggregate MCP surfaces.

    Sub-domain endpoints (/engine/mcp/{domain}, /counsel/mcp/{domain}) are
    shipped per expert pack (expert-packs/*/.mcp.json) for multi-expert
    per-domain deployment, not in the root skill.
    """
    mcp = json.loads((SKILL_ROOT / ".mcp.json").read_text(encoding="utf-8"))
    assert mcp["mcpServers"] == {
        "counsel-mcp": {
            "type": "streamableHttp",
            "url": "https://liki.hk/counsel/mcp",
        },
        "engine-mcp": {
            "type": "streamableHttp",
            "url": "https://liki.hk/engine/mcp",
        },
    }


def test_natal_cards_reference_packaged_expert_methodology():
    """Natal orchestration remains self-contained inside the unified archive.

    根 liki 是方法论权威源：natal 卡引用的方法论卡路径必须解析到根内
    (references/natal/domains/)，且该卡真实存在于根 skill。独立专家包由同步脚本保持一致。
    """
    pattern = re.compile(r"(references/natal/domains/(?:bazi|ziwei)/[A-Za-z0-9_.-]+\.md)")
    for card in (SKILL_ROOT / "references" / "natal").rglob("*.md"):
        for reference in pattern.findall(card.read_text(encoding="utf-8")):
            assert (SKILL_ROOT / reference).is_file(), (card, reference)


def test_expert_packs_follow_workbuddy_standard():
    for pack in EXPERT_PACKS:
        root = ROOT / "expert-packs" / pack
        assert (root / "SKILL.md").is_file(), pack
        assert (root / ".codebuddy-plugin" / "plugin.json").is_file(), pack
        assert (root / "README.md").is_file(), pack
        assert (root / "avatars").is_dir(), pack
        # plugin.json 必填
        d = json.loads((root / ".codebuddy-plugin" / "plugin.json").read_text(encoding="utf-8"))
        for f in ["name", "expertType", "version", "description", "author", "agents",
                  "agentName", "displayName", "profession", "displayDescription",
                  "avatar", "categoryId", "defaultInitPrompt", "plugin", "tags", "quickPrompts",
                  "keywords", "skills"]:
            assert f in d, (pack, f)
        assert d["expertType"] == "agent", pack
        agent_docs = sorted((root / "agents").glob("*.md"))
        assert agent_docs == [root / "agents" / f"{pack}.md"], f"{pack}: agents/ 只有身份文件投影的那一份"
        assert d["agentName"] == d["name"], pack
        assert d["agents"] == [f"./agents/{pack}.md"], pack
        for declared in d["agents"]:
            assert (root / declared.removeprefix("./")).is_file(), (pack, declared)
        assert d["plugin"] == d["name"], pack
        assert len(d["tags"]) == 3 and len(d["quickPrompts"]) == 3, pack
        assert d["quickPrompts"][0]["zh"] == d["defaultInitPrompt"]["zh"], pack
        n = len([c for c in d["displayDescription"]["zh"] if not c.isspace()])
        assert 40 <= n <= 50, (pack, n)
        # 运行时依赖：声明包内 .mcp.json（域级端点），不由 connectors 重复声明端点
        deps = d.get("dependencies", {})
        assert deps.get("mcpServers") == "./.mcp.json", (pack, deps)
        # agent md（平台入口）frontmatter：平台字段名 + 与文件名一致的 name
        agent_meta = yaml.safe_load(agent_docs[0].read_text(encoding="utf-8").split("---\n", 2)[1])
        assert agent_meta["name"] == pack, (pack, agent_meta)
        assert set(agent_meta) >= {"name", "description", "displayName", "profession", "maxTurns"}, pack
        assert set(agent_meta["displayName"]) == {"en", "zh"}, pack
        assert set(agent_meta["profession"]) == {"en", "zh"}, pack
        assert isinstance(agent_meta["maxTurns"], int), pack
        # 身份文件 frontmatter（内部部署与市场包共用同一份）
        identity_text = (root / "SKILL.md").read_text(encoding="utf-8")
        identity_meta = yaml.safe_load(identity_text.split("---\n", 2)[1])
        for f in ["name", "description", "description_en", "display_name",
                  "display_name_en", "profession", "maxTurns", "keywords", "skills"]:
            assert f in identity_meta, (pack, f)
        # 身份文件是单一源：plugin.json 的 keywords/skills 由它投影而来
        assert d["keywords"] == identity_meta["keywords"], pack
        assert d["skills"] == [f"./skills/{name}" for name in identity_meta["skills"]], pack
        # .mcp.json 连专家端点（正交化：排盘 engine + 判断 counsel 分离）
        mcp = json.loads((root / ".mcp.json").read_text(encoding="utf-8"))
        assert len(mcp["mcpServers"]) >= 1, pack
        assert all("engine" in name or "counsel" in name for name in mcp["mcpServers"]), (pack, mcp["mcpServers"])


def test_expert_pack_methodology_matches_index():
    # pack -> (skill dir, 镜像卡总数)。包内 references/ 是根子树逐字节镜像（含 ENTRY/TOOLS）。
    for pack, skill, expect in (
        ("liki-natal", "natal", 34),
        ("liki-divination", "divination", 25),
        ("liki-fengshui", "fengshui", 5),
        ("liki-naming", "naming", 7),
    ):
        skill_dir = ROOT / "expert-packs" / pack / "skills" / skill
        assert (skill_dir / "SKILL.md").is_file(), pack
        cards = [p for p in (skill_dir / "references").rglob("*.md") if p.name != "SKILL.md"]
        assert len(cards) == expect, (pack, len(cards), expect)


def test_no_legacy_naming_in_skills():
    legacy = ("liki-analysis", "liki-engine", "liki-master", "liki-usage", "RPC.md")
    roots = [ROOT / "skills" / "liki"] + [ROOT / "expert-packs" / pack for pack in EXPERT_PACKS]
    for root in roots:
        for path in root.rglob("*"):
            if not path.is_file() or "__pycache__" in path.parts or path.suffix == ".png":
                continue
            if path.suffix not in {".py", ".md", ".json", ".yaml", ".yml", ".txt"}:
                continue
            text = path.read_text(encoding="utf-8", errors="ignore")
            for old in legacy:
                assert old not in text, f"{path}: {old}"


def test_root_skill_routes_close_with_archive_sources():
    """Root SKILL 路由引用的各方法论域，其 domains/ 卡目录必须存在于根 skill。

    根 liki 是方法论权威源（自包含）：路由表引用的每个领域，对应方法论卡目录
    必须存在于根 skill 内，保证对外归档全能力（评审 P1）。
    """
    # root SKILL.md 路由的领域 → 根内方法论卡目录（domains/ 相对 skills/liki/）
    route_to_domain = {
        "natal": "references/natal/domains",
        "divination": "references/divination/domains",
        "fengshui": "references/fengshui/domains",
        "naming": "references/naming/domains",
    }
    root_skill = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")
    referenced = {domain for domain in route_to_domain if domain in root_skill}
    assert referenced, "root SKILL.md 未引用任何方法论域"

    for domain in referenced:
        domain_dir = SKILL_ROOT / route_to_domain[domain]
        assert domain_dir.is_dir(), f"{domain}: 方法论卡目录缺失 {domain_dir}"
        cards = [p for p in domain_dir.rglob("*.md") if p.name != "SKILL.md"]
        assert cards, f"{domain}: 方法论卡为空"

    # 全部四个方法论域必须完整（根自包含全能力）
    assert referenced == set(route_to_domain), (
        f"root 引用域与根方法论域清单不一致: {referenced}"
    )


CARD_TOKEN = re.compile(r"`([A-Za-z0-9_./-]+\.md)`")


def _resolve_card_token(skill_root: Path, doc: Path, token: str) -> Path | None:
    """把卡内 `.md` 引用解析成根 skill 内的唯一文件；解析不出或多义返回 None。

    规范形态：完整 `references/...` 路径；或相对本卡目录 / 本领域根的唯一短名。
    """
    if token.startswith("references/"):
        candidate = skill_root / token
        return candidate if candidate.is_file() else None
    if token == "SKILL.md":
        return skill_root / "SKILL.md"
    try:
        domain_root = skill_root / doc.relative_to(skill_root).parts[0] / doc.relative_to(skill_root).parts[1]
    except IndexError:  # pragma: no cover - 防御
        return None
    candidates = [doc.parent / token, domain_root / token]
    hits = sorted({c.resolve() for c in candidates if c.is_file()})
    return hits[0] if len(hits) == 1 else None


def test_domain_app_cards_are_reachable_from_entry():
    """每个领域 ENTRY 都能到达该域全部 app 卡（无孤儿卡、无重复路由表）。"""
    for domain in DOMAINS:
        domain_dir = SKILL_ROOT / "references" / domain
        cards = sorted((domain_dir / "app").glob("*.md"))
        assert cards, domain
        entry = domain_dir / "ENTRY.md"
        seeds = [entry]
        seen: set[Path] = set()
        queue = list(seeds)
        while queue:
            doc = queue.pop()
            if doc in seen:
                continue
            seen.add(doc)
            for token in CARD_TOKEN.findall(doc.read_text(encoding="utf-8")):
                target = _resolve_card_token(SKILL_ROOT, doc, token)
                if target is not None and target not in seen:
                    queue.append(target)
        orphans = [c.relative_to(SKILL_ROOT) for c in cards if c not in seen]
        assert orphans == [], f"{domain}: app 卡未被 ENTRY 路由（孤儿）: {orphans}"


def test_card_references_are_unambiguous_paths():
    """卡内 `.md` 引用必须能唯一解析：完整 `references/` 路径，或本目录 / 本域唯一短名。"""
    offenders = []
    docs = [SKILL_ROOT / "SKILL.md", SKILL_ROOT / "references" / "FAQ.md"]
    for domain in DOMAINS:
        docs += [SKILL_ROOT / "references" / domain / "ENTRY.md"]
        docs += sorted((SKILL_ROOT / "references" / domain).rglob("*.md"))
    for doc in docs:
        for number, line in enumerate(doc.read_text(encoding="utf-8").splitlines(), 1):
            for token in CARD_TOKEN.findall(line):
                if _resolve_card_token(SKILL_ROOT, doc, token) is None:
                    offenders.append(f"{doc.relative_to(SKILL_ROOT)}:{number}: {token}")
    assert offenders == [], offenders
