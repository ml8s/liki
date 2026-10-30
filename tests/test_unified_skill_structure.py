"""Structural contract for the Liki skill system（综合 liki + 专家包）。

覆盖：
1. liki 综合：根 SKILL.md + 平铺四域卡（references/<域>-ENTRY.md 等）+ 共享运行时文件
2. 专家包（四个顶层领域）：WorkBuddy 标准（SKILL.md 身份源 / plugin.json（含连接器依赖）/ skills / avatars / README）
3. MCP 连接器命名（engine/counsel 与专家端点）
"""
import json
import re
import sys
from pathlib import Path

import yaml
from _helpers import ROOT, SKILL_ROOT

sys.path.insert(0, str(ROOT / "scripts"))
import sync_expert_packs as gen  # noqa: E402

DOMAINS = ("natal", "divination", "fengshui", "naming")
# 顶层领域 = 专家包（1:1）：由生成器 MAPPINGS 派生，杜绝第二份包名清单。
EXPERT_PACKS = tuple(pack for _domain, pack in gen.MAPPINGS)


def test_liki_root_has_one_entry_and_four_domain_entries():
    assert (SKILL_ROOT / "SKILL.md").is_file()
    for domain in DOMAINS:
        assert (SKILL_ROOT / "references" / f"{domain}-ENTRY.md").is_file(), domain


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
        assert "FAQ.md" not in (SKILL_ROOT / "references" / f"{entry}-ENTRY.md").read_text(encoding="utf-8")


def test_app_cards_use_standard_contract_sections():
    required = {"## 流程", "## 边界条件", "## 输出模板"}
    cards = list(SKILL_ROOT.glob("references/*-app-*.md"))
    assert len(cards) >= 10
    for path in cards:
        lines = {line.strip() for line in path.read_text(encoding="utf-8").splitlines()}
        assert required <= lines, (path, required - lines)


def test_yongshen_decision_has_one_domain_authority():
    """用神裁决只写在八字域；起名域只消费结果。"""
    natal = (SKILL_ROOT / "references/natal-bazi-yongshen.md").read_text(encoding="utf-8")
    naming = (SKILL_ROOT / "references/naming-bazi-yongshen.md").read_text(encoding="utf-8")
    assert "## 裁决规则" in natal
    assert "YS-CG-FAKE" in natal
    assert "## 裁决规则" not in naming
    assert "八字域" in naming


def test_skill_docs_do_not_duplicate_mcp_method_signatures():
    """工具与参数由 MCP 自省；技能文档只保留领域能力与判断口径。"""
    forbidden = (
        "compute_factors",
        "naming_check",
        "period_query",
        "bazi_yongshen",
        "bazi_chart",
    )
    offenders = []
    for path in SKILL_ROOT.rglob("*.md"):
        text = path.read_text(encoding="utf-8")
        if any(name in text for name in forbidden):
            offenders.append(path.relative_to(SKILL_ROOT).as_posix())
    assert offenders == []


def test_app_cards_declare_output_template_body():
    """`## 输出模板` 必须给出模板体（代码块 / 表 / 列表），不能只有示例。"""
    for card in sorted(SKILL_ROOT.glob("references/*-app-*.md")):
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


def test_natal_layers_are_orthogonal():
    """三层正交：根入口不索引侧层文件；侧入口只索引本层卡，且本层卡齐备。"""
    root_entry = (SKILL_ROOT / "references" / "natal-ENTRY.md").read_text(encoding="utf-8")
    assert "八字侧" in root_entry and "紫微侧" in root_entry
    assert "编排" in root_entry
    assert "专家" not in root_entry
    # 跨层索引禁止：根入口不得出现侧层文件名
    assert "natal-bazi-" not in root_entry and "natal-ziwei-" not in root_entry
    for side, other in (("bazi", "ziwei"), ("ziwei", "bazi")):
        entry = (SKILL_ROOT / "references" / f"natal-{side}-ENTRY.md").read_text(encoding="utf-8")
        assert "## 任务路由" in entry, side
        assert f"natal-{other}-" not in entry, side
        cards = {
            card.name
            for card in (SKILL_ROOT / "references").glob(f"natal-{side}-*.md")
            if card.name != f"natal-{side}-ENTRY.md"
        }
        routed = set(re.findall(rf"references/(natal-{side}-[A-Za-z0-9_-]+\.md)", entry))
        assert cards == routed, (side, sorted(cards - routed), sorted(routed - cards))


def test_natal_side_entries_declare_assertion_contract():
    """断语公共字段、side 取值与受控闭集是硬契约，落在两侧入口（侧是断言消费者）。"""
    routes = json.loads(
        (ROOT / "counsel" / "app" / "natal" / "tools" / "topic_routes.json").read_text(encoding="utf-8")
    )
    for side in ("bazi", "ziwei"):
        text = (SKILL_ROOT / "references" / f"natal-{side}-ENTRY.md").read_text(encoding="utf-8")
        for field in ("side", "topic", "time_scope"):
            assert f"`{field}`" in text, (side, field)
        assert "`side` 是 `bazi` / `ziwei`" in text, side
        for value in ("career", "marriage", "study", "wealth", "chart_structure", "relocation"):
            assert f"`{value}`" in text, (side, value)
        for scope in ("current_year", "year_range", "current_decade", "decade"):
            assert f"`{scope}`" in text, (side, scope)
        # topic 键＝标签与服务器路由表单一源一致（卡内 `本命判断（事业）` 靠它对上 topic）
        for key, spec in routes["topics"].items():
            assert f"`{key}`＝{spec['label']}" in text, (side, key, spec["label"])


def test_naming_does_not_silently_default_missing_hour():
    app = (SKILL_ROOT / "references" / "naming-app-naming.md").read_text(encoding="utf-8")
    calibration = (SKILL_ROOT / "references" / "naming-bazi-calibration.md").read_text(encoding="utf-8")
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
    for card in sorted((SKILL_ROOT / "references").glob("natal-app-*.md")):
        for token in qualifier.findall(card.read_text(encoding="utf-8")):
            if token not in labels:
                offenders.append(f"{card.name}: {token}")
    assert offenders == []


def test_natal_card_actions_name_the_side():
    """卡内动作的每个能力调用都要显式带侧（counsel side 默认 bazi，省略即静默降级）。"""
    capability = re.compile(r"`(?:本命判断|应期判断)")
    offenders = []
    for card in sorted((SKILL_ROOT / "references").glob("natal-app-*.md")):
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

    Expert packs do not ship their own MCP manifest: they depend on the two
    published WorkBuddy connectors (plugin.json.dependencies.connectors). The
    sub-domain endpoints (/engine/mcp/{domain}, /counsel/mcp/{domain}) stay a
    self-hosted per-domain deployment knob, not a submission artifact.
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
    (references/natal-{bazi,ziwei}-*.md 平铺卡)，且该卡真实存在于根 skill。
    独立专家包由同步脚本保持一致。
    """
    pattern = re.compile(r"(references/natal-(?:bazi|ziwei)-[A-Za-z0-9_.-]+\.md)")
    for card in sorted((SKILL_ROOT / "references").glob("natal-*.md")):
        for reference in pattern.findall(card.read_text(encoding="utf-8")):
            assert (SKILL_ROOT / reference).is_file(), (card, reference)


def test_expert_packs_follow_workbuddy_standard():
    connector_ids, id_errors = gen.connector_ids()
    assert not id_errors, id_errors
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
        # 运行时依赖：只依赖已上架连接器（ID 单一源 workbuddy/connector-ids.json），
        # 专家不自带 MCP —— 自带会让同一次召唤出现两套工具面与两组引导卡片。
        deps = d.get("dependencies", {})
        assert "mcpServers" not in deps, (pack, deps)
        assert deps.get("connectors") == connector_ids, (pack, deps)
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
        # 依赖连接器时包内不得有 .mcp.json（平台会作为兜底自动加载，形成双重工具面）
        assert not (root / ".mcp.json").exists(), pack


def test_expert_pack_methodology_matches_index():
    # pack -> (skill dir, 镜像卡总数)。包内 references/ 是根平铺卡的按域前缀逐字节视图。
    for pack, skill, expect in (
        ("liki", "liki", 72),
                            ):
        skill_dir = ROOT / "expert-packs" / pack / "skills" / "liki"
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
    """Root SKILL 路由引用的各方法论域，其平铺卡必须存在于根 skill。

    根 liki 是方法论权威源（自包含）：路由表引用的每个领域，对应方法论卡
    必须存在于根 skill 内，保证对外归档全能力（评审 P1）。
    """
    root_skill = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")
    referenced = {domain for domain in DOMAINS if domain in root_skill}
    assert referenced, "root SKILL.md 未引用任何方法论域"

    for domain in referenced:
        cards = sorted((SKILL_ROOT / "references").glob(f"{domain}-*.md"))
        assert cards, f"{domain}: 方法论卡为空"

    # 全部四个方法论域必须完整（根自包含全能力）
    assert referenced == set(DOMAINS), (
        f"root 引用域与根方法论域清单不一致: {referenced}"
    )


CARD_TOKEN = re.compile(r"`([A-Za-z0-9_./-]+\.md)`")


def _resolve_card_token(skill_root: Path, token: str) -> Path | None:
    """把卡内 `.md` 引用解析成根 skill 内的唯一文件；解析不出返回 None。

    规范形态只有一种：完整 `references/...` 平铺全路径（另允许 `SKILL.md` 入口自述）。
    """
    if token.startswith("references/"):
        candidate = skill_root / token
        return candidate if candidate.is_file() else None
    if token == "SKILL.md":
        return skill_root / "SKILL.md"
    return None


def test_side_entry_contract_sections_match():
    """两侧入口的断语契约段逐字一致（仅年号界按体系不同）。"""
    normalized = {}
    for side, boundary in (("bazi", "八字公历立春界"), ("ziwei", "紫微农历春节界")):
        text = (SKILL_ROOT / "references" / f"natal-{side}-ENTRY.md").read_text(encoding="utf-8")
        start = text.index("## 断语契约")
        end = text.find("\n## ", start + 1)
        section = text[start:end if end >= 0 else len(text)]
        normalized[side] = section.replace(boundary, "<BOUNDARY>")
    assert normalized["bazi"] == normalized["ziwei"]


def test_app_info_words_are_routable_by_side_entries():
    """app 卡的信息需求词必须被两侧入口的任务路由覆盖（拓扑透明的前提）。"""
    info = re.compile(r"(?:八字 \+ 紫微|八字|紫微)：([\u4e00-\u9fff\s/]+?)(?:本命|应期)")
    side_text = "\n".join(
        (SKILL_ROOT / "references" / f"natal-{side}-ENTRY.md").read_text(encoding="utf-8")
        for side in ("bazi", "ziwei")
    )
    missing = []
    for card in sorted((SKILL_ROOT / "references").glob("natal-app-*.md")):
        for captured in info.findall(card.read_text(encoding="utf-8")):
            for word in re.split(r"[\s/]+", captured):
                if word and word not in side_text:
                    missing.append(f"{card.name}: {word}")
    assert missing == [], f"app 信息词未被侧入口路由覆盖: {missing}"


def test_domain_app_cards_are_reachable_from_entry():
    """每个领域 ENTRY 都能到达该域全部 app 卡（无孤儿卡、无重复路由表）。"""
    for domain in DOMAINS:
        cards = sorted((SKILL_ROOT / "references").glob(f"{domain}-app-*.md"))
        assert cards, domain
        entry = SKILL_ROOT / "references" / f"{domain}-ENTRY.md"
        seeds = [entry]
        seen: set[Path] = set()
        queue = list(seeds)
        while queue:
            doc = queue.pop()
            if doc in seen:
                continue
            seen.add(doc)
            for token in CARD_TOKEN.findall(doc.read_text(encoding="utf-8")):
                target = _resolve_card_token(SKILL_ROOT, token)
                if target is not None and target not in seen:
                    queue.append(target)
        orphans = [c.relative_to(SKILL_ROOT) for c in cards if c not in seen]
        assert orphans == [], f"{domain}: app 卡未被 ENTRY 路由（孤儿）: {orphans}"


def test_card_references_are_unambiguous_paths():
    """卡内 `.md` 引用必须使用完整 `references/` 平铺路径（或 `SKILL.md` 自述）。"""
    offenders = []
    docs = [SKILL_ROOT / "SKILL.md", SKILL_ROOT / "references" / "FAQ.md"]
    docs += sorted((SKILL_ROOT / "references").glob("*.md"))
    for doc in docs:
        for number, line in enumerate(doc.read_text(encoding="utf-8").splitlines(), 1):
            for token in CARD_TOKEN.findall(line):
                if _resolve_card_token(SKILL_ROOT, token) is None:
                    offenders.append(f"{doc.relative_to(SKILL_ROOT)}:{number}: {token}")
    assert offenders == [], offenders


def test_domain_lists_stay_in_sync():
    """域清单三处定义（check_docs / 生成器 MAPPINGS / 本测试）必须一致。"""
    import check_docs

    assert set(check_docs.DOMAINS) == {domain for domain, _pack in gen.MAPPINGS} == {"liki"}
