"""Contract: the unified skill has one valid SKILL.md and bounded app cards."""
import unittest

import yaml
from _helpers import ROOT, SKILL_ROOT


class TestSkillFrontmatter(unittest.TestCase):
    def test_unified_skill_frontmatter_is_valid_yaml(self):
        txt = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")
        self.assertTrue(txt.startswith("---\n"))
        meta = yaml.safe_load(txt.split("---\n")[1])
        self.assertIsInstance(meta, dict)
        self.assertEqual(meta.get("name"), "liki")
        self.assertIsInstance(meta.get("description"), str)
        self.assertTrue(meta["description"].strip())
        self.assertEqual(meta["metadata"].get("agent_created"), "true")

    def test_skill_has_no_nested_skill_entries(self):
        skills = list(SKILL_ROOT.rglob("SKILL.md"))
        self.assertEqual([p.relative_to(SKILL_ROOT) for p in skills], [__import__("pathlib").Path("SKILL.md")])

    def test_skills_dir_holds_single_installable_skill(self):
        """npx skills add ml8s/liki 只装 liki：skills/ 下有且仅有一个 SKILL.md。
        专家包必须位于 expert-packs/，否则会被 CLI 误装进任何宿主。"""
        roots = list(SKILL_ROOT.parent.rglob("SKILL.md"))
        self.assertEqual([str(p.relative_to(SKILL_ROOT.parent)) for p in roots],
                         ["liki/SKILL.md"], roots)

    def test_description_within_adk_skilltoolset_limit(self):
        """ADK skilltoolset: description 1..1024 字节，超限则 ProcessRequest 每轮失败。"""
        meta = yaml.safe_load((SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8").split("---")[1])
        n = len(meta["description"].encode("utf-8"))
        self.assertGreater(n, 0)
        self.assertLessEqual(n, 1024, f"description {n}B > ADK 1024B")

    def test_frontmatter_is_adk_skilltoolset_compatible(self):
        """Runtime skilltoolset 只接受 agentskills.io 标准字段；平台元数据必须收进 metadata。"""
        meta = yaml.safe_load((SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8").split("---")[1])
        self.assertEqual(
            set(meta),
            {"name", "license", "description", "metadata"},
        )
        for key in ("description_zh", "description_en", "version", "author"):
            value = meta["metadata"].get(key)
            self.assertIsInstance(value, str, key)
            self.assertTrue(value.strip(), key)

    def test_frontmatter_version_is_platform_semver_projection(self):
        """SKILL.md metadata.version 是平台投影：取发行 CalVer 的 semver 投影（平台拒 4 段）。"""
        import sys

        sys.path.insert(0, str(ROOT / "scripts"))
        from version_projection import workbuddy_version

        meta = yaml.safe_load((SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8").split("---")[1])
        version = meta["metadata"]["version"]
        runtime_version = (ROOT / "skills" / "liki" / "VERSION.txt").read_text(encoding="utf-8").strip()
        self.assertEqual(str(version), workbuddy_version(runtime_version))
        self.assertRegex(str(version), r"^\d+\.\d+\.\d+$", version)

    def test_skill_readable_markdown_lives_in_references(self):
        """ADK LoadResource 只认 references/assets/scripts：references 外不得有可读 md。"""
        outside = [
            p.relative_to(SKILL_ROOT)
            for p in SKILL_ROOT.rglob("*.md")
            if "references" not in p.parts
        ]
        self.assertEqual(outside, [__import__("pathlib").Path("SKILL.md")], outside)

    def test_app_cards_frontmatter_is_valid_yaml(self):
        cards = sorted(SKILL_ROOT.glob("references/natal-app-*.md"))
        self.assertGreater(len(cards), 0)
        for card in cards:
            with self.subTest(card=card.name):
                txt = card.read_text(encoding="utf-8")
                if not txt.startswith("---\n"):
                    continue
                meta = yaml.safe_load(txt.split("---\n")[1])
                self.assertIsInstance(meta, dict)

    def test_bazi_app_cards_keep_required_reading_bounded(self):
        cards = sorted(SKILL_ROOT.glob("references/natal-app-*.md"))
        for card in cards:
            with self.subTest(card=card.name):
                count = card.read_text(encoding="utf-8").count("[必读]")
                self.assertLessEqual(count, 3)


if __name__ == "__main__":
    unittest.main()
