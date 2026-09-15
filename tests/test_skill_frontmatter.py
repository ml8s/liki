"""契约：SKILL.md frontmatter 必须是合法 YAML。

feedback a9c24b71：liki-divination description 含未加引号的 ASCII ': '，
npx skills add 解析 frontmatter 即报 'Nested mappings are not allowed'，安装失败。
"""
import unittest

import yaml

from helpers import ROOT, SKILL_NAMES

SKILLS_DIR = ROOT / "skills"


class TestSkillFrontmatter(unittest.TestCase):
    def test_all_skill_frontmatter_is_valid_yaml(self):
        for s in SKILL_NAMES:
            with self.subTest(skill=s):
                txt = (SKILLS_DIR / s / "SKILL.md").read_text(encoding="utf-8")
                self.assertTrue(txt.startswith("---\n"), f"{s}: 缺 frontmatter")
                fm = txt.split("---\n")[1]
                meta = yaml.safe_load(fm)  # 含未引号 ': ' 会在此抛 ParserError
                self.assertIsInstance(meta, dict)
                self.assertEqual(meta.get("name"), s)
                self.assertIsInstance(meta.get("description"), str)
                self.assertTrue(meta["description"].strip())
                # WorkBuddy SkillManage 要求 agent_created: true 才能修改/删除技能
                self.assertIs(meta.get("agent_created"), True)

    def test_app_cards_frontmatter_is_valid_yaml(self):
        cards = sorted((SKILLS_DIR / "liki-bazi" / "app").glob("*.md"))
        self.assertGreater(len(cards), 0)
        for card in cards:
            with self.subTest(card=card.name):
                txt = card.read_text(encoding="utf-8")
                if not txt.startswith("---\n"):
                    continue  # 无 frontmatter 的卡（如 README）跳过
                meta = yaml.safe_load(txt.split("---\n")[1])
                self.assertIsInstance(meta, dict)

    def test_bazi_app_cards_keep_required_reading_bounded(self):
        cards = sorted((SKILLS_DIR / "liki-bazi" / "app").glob("*.md"))
        for card in cards:
            with self.subTest(card=card.name):
                count = card.read_text(encoding="utf-8").count("[必读]")
                self.assertLessEqual(count, 3)


if __name__ == "__main__":
    unittest.main()
