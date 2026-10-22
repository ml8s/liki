"""Contract test: the canonical Liki slogan stays consistent across public windows."""
import yaml
from _helpers import ROOT, SKILL_ROOT, SLOGAN, SLOGAN_EN

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
STALE_BRAND_CONTRACT = [
    "The canonical slogan is Chinese-only",
    "命理师的 Skill",
]


def test_root_readmes_repeat_canonical_slogan():
    zh_text = (ROOT / "README.md").read_text(encoding="utf-8")
    assert zh_text.count(SLOGAN) >= 2

    # Brand rule B: English surfaces use the English explanation only.
    en_text = (ROOT / "README.en.md").read_text(encoding="utf-8")
    assert en_text.count(SLOGAN_EN) >= 2
    assert SLOGAN not in en_text


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


def test_brand_doc_points_to_canonical_truth_source():
    """真相源已迁至 liki-web 仓库 BRAND.md（v5.0）；本仓 docs/BRAND.md 仅作指针。"""
    path = ROOT / "docs" / "BRAND.md"
    text = path.read_text(encoding="utf-8")
    assert "liki-web" in text, "docs/BRAND.md must point to the canonical brand source in the liki-web repository"
    assert "BRAND.md" in text
