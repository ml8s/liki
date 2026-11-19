"""Core structure and safety contracts for the unified Liki skill."""

import json
import re
import sys
from pathlib import Path

from _helpers import ROOT, SKILL_ROOT

sys.path.insert(0, str(ROOT / "scripts"))
import build_channel_package as gen  # noqa: E402

DOMAINS = ("natal", "divination", "fengshui", "naming")
CARD_TOKEN = re.compile(r"`((?:references/)?[A-Za-z0-9_/-]+\.md|SKILL\.md)`")


def _resolve_card(skill_root: Path, token: str) -> Path | None:
    if token.startswith("references/"):
        candidate = skill_root / token
        return candidate if candidate.is_file() else None
    if token == "SKILL.md":
        return skill_root / "SKILL.md"
    return None


def test_liki_root_has_one_entry_and_domain_routing_entries():
    assert (SKILL_ROOT / "SKILL.md").is_file()
    for domain in DOMAINS:
        assert (SKILL_ROOT / "references" / f"{domain}-ENTRY.md").is_file(), domain


def test_shared_runtime_files_are_not_duplicated():
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("feedback.py")] == [
        Path("scripts/feedback.py")
    ]
    assert [
        p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("feedback.schema.json")
    ] == [Path("feedback.schema.json")]
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("FAQ.md")] == [
        Path("references/FAQ.md")
    ]


def test_skill_has_no_nested_python_manifests():
    assert [p.relative_to(SKILL_ROOT) for p in SKILL_ROOT.rglob("requirements*.txt")] == []


def test_root_entry_is_a_lightweight_router():
    lines = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8").splitlines()
    assert len(lines) <= 120
    assert "## MCP 前置" in lines
    assert "## 统一硬边界" in lines
    assert "## 输出契约" in lines
    assert "## 语言协议" in lines


def test_natal_layers_are_orthogonal():
    root_entry = (SKILL_ROOT / "references" / "natal-ENTRY.md").read_text(encoding="utf-8")
    assert "八字侧" in root_entry and "紫微侧" in root_entry
    assert "references/natal-bazi-ENTRY.md" in root_entry
    assert "references/natal-ziwei-ENTRY.md" in root_entry
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


def test_naming_does_not_silently_default_missing_hour():
    naming = (SKILL_ROOT / "references" / "naming-app-naming.md").read_text(encoding="utf-8")
    calibration = (SKILL_ROOT / "references" / "naming-bazi-calibration.md").read_text(encoding="utf-8")
    assert "默认取午时" not in naming + calibration
    assert "默认时辰" not in naming + calibration
    assert "不排八字" in naming


def test_root_mcp_uses_explicit_free_surface():
    text = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")
    assert all(domain in text for domain in DOMAINS)
    assert not re.search(r"liki-(?:natal|divination|fengshui|naming)", text)
    assert json.loads((SKILL_ROOT / ".mcp.json").read_text(encoding="utf-8")) == {
        "mcpServers": {
            "engine": {
                "type": "streamableHttp",
                "url": "https://mcp.tokflux.com/engine/mcp",
            },
            "counsel": {
                "type": "streamableHttp",
                "url": "https://mcp.tokflux.com/counsel/mcp",
            },
        }
    }


def test_root_skill_references_no_expert_pack_ids():
    pack_id = re.compile(r"liki-(?:bazi|ziwei|liuyao|qimen|fengshui|naming)")
    offenders = []
    for path in sorted(SKILL_ROOT.rglob("*")):
        if not path.is_file() or "__pycache__" in path.parts or path.suffix not in {
            ".py",
            ".md",
            ".json",
            ".yaml",
            ".yml",
            ".txt",
        }:
            continue
        for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            if pack_id.search(line):
                offenders.append(f"{path.relative_to(SKILL_ROOT)}:{number}")
    assert offenders == []


def test_channel_packages_do_not_expose_internal_host_protocol():
    for channel in ("skillhub", "workbuddy"):
        for variant in ("free", "pro"):
            files = gen.package_manifest(channel, variant)
            for name, payload in files.items():
                if not name.endswith((".md", ".json", ".py")):
                    continue
                text = payload.decode("utf-8")
                assert "## 内部档案同步" not in text, (channel, variant, name)
                assert "sync_ui" not in text, (channel, variant, name)
                assert "is_calibrated" not in text, (channel, variant, name)


def test_channel_manifests_use_only_canonical_methodology_files():
    canonical = {path.name for path in (SKILL_ROOT / "references").glob("*.md")}
    for channel in ("skillhub", "workbuddy"):
        for variant in ("free", "pro"):
            files = gen.package_manifest(channel, variant)
            top = gen.load_package(channel, variant)["top"]
            prefix = f"{top}/references/" if channel == "skillhub" else f"{top}/skills/liki/references/"
            names = {
                Path(name).name
                for name in files
                if prefix in name
            }
            assert names == canonical, (channel, variant)


def test_domain_app_cards_are_reachable_from_entry():
    for domain in DOMAINS:
        cards = sorted((SKILL_ROOT / "references").glob(f"{domain}-app-*.md"))
        assert cards, domain
        entry = SKILL_ROOT / "references" / f"{domain}-ENTRY.md"
        seen: set[Path] = set()
        queue = [entry]
        while queue:
            doc = queue.pop()
            if doc in seen:
                continue
            seen.add(doc)
            for token in CARD_TOKEN.findall(doc.read_text(encoding="utf-8")):
                target = _resolve_card(SKILL_ROOT, token)
                if target is not None and target not in seen:
                    queue.append(target)
        orphans = [card.relative_to(SKILL_ROOT) for card in cards if card not in seen]
        assert orphans == [], f"{domain}: orphan app cards: {orphans}"


def test_card_references_are_unambiguous_paths():
    offenders = []
    docs = [SKILL_ROOT / "SKILL.md", SKILL_ROOT / "references" / "FAQ.md"]
    docs.extend(sorted((SKILL_ROOT / "references").glob("*.md")))
    for doc in docs:
        for number, line in enumerate(doc.read_text(encoding="utf-8").splitlines(), 1):
            for token in CARD_TOKEN.findall(line):
                if _resolve_card(SKILL_ROOT, token) is None:
                    offenders.append(f"{doc.relative_to(SKILL_ROOT)}:{number}: {token}")
    assert offenders == []
