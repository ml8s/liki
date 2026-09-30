"""Contract: skilltoolset bindings are complete and resolve to real skill roots."""
import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

import yaml
from _helpers import ROOT

SKILL_TOOLSET_TOOLS = ["list_skills", "load_skill", "load_skill_resource"]
EXPERTS = ["bazi", "ziwei", "liuyao", "qimen", "fengshui", "naming"]
ALLOWED_ROOTS = {"/skills"} | {f"/expert-packs/liki-{name}/skills" for name in EXPERTS}


def _repo_path(root: str) -> Path:
    if root == "/skills":
        return ROOT / "skills"
    return ROOT / root.lstrip("/")


class TestAgentYamlSkillsBindings(unittest.TestCase):
    def test_every_agent_yaml_binds_skilltoolset(self):
        agent_dirs = sorted(path for path in (ROOT / "agents").iterdir() if path.is_dir())
        self.assertEqual(len(agent_dirs), 8)
        for agent_dir in agent_dirs:
            with self.subTest(agent=agent_dir.name):
                document = yaml.safe_load(
                    (agent_dir / "agent.yaml").read_text(encoding="utf-8")
                )
                skills = document.get("skills")
                self.assertIsInstance(skills, dict, f"{agent_dir.name} missing skills")
                root = skills.get("root")
                self.assertIn(root, ALLOWED_ROOTS)
                allow = document["tools"]["allow"].get("skilltoolset")
                self.assertEqual(sorted(allow), sorted(SKILL_TOOLSET_TOOLS))

    def test_every_root_resolves_to_a_skill_directory(self):
        for root in sorted(ALLOWED_ROOTS):
            with self.subTest(root=root):
                base = _repo_path(root)
                self.assertTrue(base.is_dir(), f"{base} missing")
                skill_mds = [p for p in base.iterdir() if p.is_dir() and (p / "SKILL.md").is_file()]
                self.assertGreaterEqual(len(skill_mds), 1, f"no SKILL.md under {base}")


class TestGeneratedDeploymentSkillsBinding(unittest.TestCase):
    def _generate(self, profile: str) -> dict:
        tmp = Path(tempfile.mkdtemp(prefix="liki-skills-"))
        self.addCleanup(shutil.rmtree, tmp, ignore_errors=True)
        subprocess.run(
            [
                sys.executable,
                str(ROOT / "scripts" / "generate_deployment.py"),
                "--profile",
                profile,
                "--out",
                str(tmp / "out"),
            ],
            check=True,
        )
        manifest = tmp / "out" / profile / "deployment.json"
        return json.loads(manifest.read_text(encoding="utf-8"))

    def test_profiles_bind_skilltoolset_for_every_agent(self):
        for profile in ("experts", "single"):
            with self.subTest(profile=profile):
                deployment = self._generate(profile)
                agents = deployment["spec"]["agents"]
                self.assertGreaterEqual(len(agents), 1)
                for agent in agents:
                    skills = agent.get("skills")
                    self.assertIsInstance(skills, dict, f"{profile}:{agent['name']} missing skills")
                    self.assertIn(skills["root"], ALLOWED_ROOTS)
                    self.assertNotIn("preload", skills)
                    allow = agent["tools"]["allow"].get("skilltoolset")
                    self.assertEqual(sorted(allow), sorted(SKILL_TOOLSET_TOOLS))


class TestAssemblyCopiesSkillRoots(unittest.TestCase):
    def test_dockerfile_copies_skill_roots(self):
        dockerfile = (ROOT / "assembly" / "Dockerfile").read_text(encoding="utf-8")
        self.assertIn("COPY dist/skills/ /skills/", dockerfile)
        self.assertIn("COPY dist/expert-packs/ /expert-packs/", dockerfile)


if __name__ == "__main__":
    unittest.main()
