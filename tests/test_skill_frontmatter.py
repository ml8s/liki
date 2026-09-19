"""Contract: the unified skill has one valid SKILL.md and bounded app cards."""
import unittest

import yaml

from _helpers import SKILL_ROOT, ROOT


class TestSkillFrontmatter(unittest.TestCase):
    def test_unified_skill_frontmatter_is_valid_yaml(self):
        txt = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")
        self.assertTrue(txt.startswith("---\n"))
        meta = yaml.safe_load(txt.split("---\n")[1])
        self.assertIsInstance(meta, dict)
        self.assertEqual(meta.get("name"), "liki")
        self.assertIsInstance(meta.get("description"), str)
        self.assertTrue(meta["description"].strip())
        self.assertIs(meta.get("agent_created"), True)

    def test_skill_has_no_nested_skill_entries(self):
        skills = list(SKILL_ROOT.rglob("SKILL.md"))
        self.assertEqual([p.relative_to(SKILL_ROOT) for p in skills], [__import__("pathlib").Path("SKILL.md")])

    def test_app_cards_frontmatter_is_valid_yaml(self):
        cards = sorted((SKILL_ROOT / "bazi" / "app").glob("*.md"))
        self.assertGreater(len(cards), 0)
        for card in cards:
            with self.subTest(card=card.name):
                txt = card.read_text(encoding="utf-8")
                if not txt.startswith("---\n"):
                    continue
                meta = yaml.safe_load(txt.split("---\n")[1])
                self.assertIsInstance(meta, dict)

    def test_bazi_app_cards_keep_required_reading_bounded(self):
        cards = sorted((SKILL_ROOT / "bazi" / "app").glob("*.md"))
        for card in cards:
            with self.subTest(card=card.name):
                count = card.read_text(encoding="utf-8").count("[必读]")
                self.assertLessEqual(count, 3)


if __name__ == "__main__":
    unittest.main()
