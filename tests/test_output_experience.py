"""Global output contracts are the standard way to cover every app card."""

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def test_all_skill_output_contracts_are_conclusion_first_and_reality_scoped():
    for skill in ("liki-naming", "liki-bazi", "liki-divination", "liki-fengshui"):
        text = (ROOT / "skills" / skill / "SKILL.md").read_text(encoding="utf-8")
        assert "首句" in text or "先给一句话判断" in text, skill
        assert "现实" in text or "专业" in text, skill
        assert "输出语言跟随用户" in text, skill
