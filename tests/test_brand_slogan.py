"""Contract test: the canonical Liki slogan stays consistent across public windows."""

import yaml

from helpers import ROOT, SKILL_NAMES, skill_dir, SLOGAN
FORBIDDEN_VARIANTS = [
    "问命理，用灵机",
    "命理事，问灵机",
    "灵机，懂命理",
]
PUBLIC_WINDOWS = [ROOT / "README.md", ROOT / "README.en.md"] + [
    skill_dir(name) / "SKILL.md" for name in SKILL_NAMES
]


def test_root_readmes_repeat_canonical_slogan():
    for path in [ROOT / "README.md", ROOT / "README.en.md"]:
        text = path.read_text(encoding="utf-8")
        assert text.count(SLOGAN) >= 2, f"{path}: canonical slogan must appear in hero and summary"

    en_text = (ROOT / "README.en.md").read_text(encoding="utf-8")
    assert "Trusted by those who understand Chinese Metaphysics" in en_text


def test_each_skill_repeats_canonical_slogan_in_description_and_intro():
    for name in SKILL_NAMES:
        path = skill_dir(name) / "SKILL.md"
        text = path.read_text(encoding="utf-8")
        meta = yaml.safe_load(text.split("---\n")[1])
        assert SLOGAN in meta["description"], f"{name}: discovery description lacks canonical slogan"
        assert text.count(SLOGAN) >= 2, f"{name}: visible intro lacks canonical slogan"


def test_public_windows_do_not_use_retired_variants():
    for path in PUBLIC_WINDOWS:
        text = path.read_text(encoding="utf-8")
        for variant in FORBIDDEN_VARIANTS:
            assert variant not in text, f"{path}: retired slogan variant {variant!r}"


def test_brand_doc_is_the_canonical_truth_source():
    text = (ROOT / "docs" / "brand.md").read_text(encoding="utf-8")
    assert "Slogan（品牌口号）" in text
    assert "«懂命理，用灵机»" in text
    assert "「懂命理的人，都在用灵机」" in text
    assert "这是 Liki 唯一对外 slogan" in text
