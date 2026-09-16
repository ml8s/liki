"""Contract test: the canonical Liki slogan stays consistent across public windows."""
import yaml

from helpers import ROOT, SKILL_ROOT, SLOGAN

FORBIDDEN_VARIANTS = [
    "Liki 灵机",
    "灵机 Skill",
    "灵机命书",
    "灵机合盘",
    "灵机命理",
    "问命理，用灵机",
    "命理事，问灵机",
    "灵机，懂命理",
    "懂命理，用灵机",
]
PUBLIC_WINDOWS = [ROOT / "README.md", ROOT / "README.en.md", SKILL_ROOT / "SKILL.md"]
WEBAPP_WINDOWS = sorted((ROOT / "webapp").glob("*/**/*.md"))
STALE_BRAND_CONTRACT = [
    "The canonical slogan is Chinese-only",
    "命理师的 Skill",
]


def test_root_readmes_repeat_canonical_slogan():
    for path in [ROOT / "README.md", ROOT / "README.en.md"]:
        text = path.read_text(encoding="utf-8")
        assert text.count(SLOGAN) >= 2

    en_text = (ROOT / "README.en.md").read_text(encoding="utf-8")
    assert "For Chinese Metaphysics, use Liki." in en_text


def test_skill_repeats_canonical_slogan_in_description_and_intro():
    path = SKILL_ROOT / "SKILL.md"
    text = path.read_text(encoding="utf-8")
    meta = yaml.safe_load(text.split("---\n")[1])
    assert SLOGAN in meta["description"]
    assert text.count(SLOGAN) >= 2


def test_public_windows_do_not_use_retired_variants():
    for path in PUBLIC_WINDOWS:
        text = path.read_text(encoding="utf-8")
        for variant in FORBIDDEN_VARIANTS:
            assert variant not in text, f"{path}: retired slogan variant {variant!r}"


def test_webapp_report_windows_use_liki_brand():
    assert WEBAPP_WINDOWS
    for path in WEBAPP_WINDOWS:
        text = path.read_text(encoding="utf-8")
        for variant in FORBIDDEN_VARIANTS:
            assert variant not in text, f"{path}: retired brand variant {variant!r}"


def test_brand_doc_is_the_canonical_truth_source():
    path = ROOT / "docs" / "brand.md"
    text = path.read_text(encoding="utf-8")
    active_text = text.split("## Governance Log（治理记录）", 1)[0]

    assert "Liki Canonical Brand Definition (v4.0)" in text
    assert "Liki 命理" in active_text
    assert "Liki has no separate Chinese brand name" in active_text
    assert "Slogan（品牌口号）" in active_text
    assert "«懂命理，用 Liki»" in active_text
    assert "「懂命理的人，都在用 Liki」" in active_text
    assert "这是 Liki 唯一对外 slogan" in active_text
    assert "Liki — 专业命理 Skill" in active_text

    for variant in FORBIDDEN_VARIANTS:
        assert variant not in active_text, f"{path}: retired brand variant {variant!r}"
    for phrase in STALE_BRAND_CONTRACT:
        assert phrase not in active_text, f"{path}: stale v3 brand contract {phrase!r}"
