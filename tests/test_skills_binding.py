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
EXPERTS = ["liki", "liki", "liki", "liki"]
PROFILES = ("liki-expert",)
ALLOWED_ROOTS = {"/skills"}


def _repo_path(root: str) -> Path:
    if root == "/skills":
        return ROOT / "skills"
    return ROOT / root.lstrip("/")


class TestAgentYamlSkillsBindings(unittest.TestCase):
    def test_every_deployment_yaml_binds_skilltoolset(self):
        deployment_yamls = sorted((ROOT / "distribution").glob("*/deployment.yaml"))
        self.assertEqual(len(deployment_yamls), len(PROFILES))
        for deployment_yaml in deployment_yamls:
            with self.subTest(deployment_yaml=deployment_yaml.name):
                document = yaml.safe_load(deployment_yaml.read_text(encoding="utf-8"))
                skills = document.get("skills")
                self.assertIsInstance(skills, dict, f"{deployment_yaml} missing skills")
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
        for profile in PROFILES:
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
        self.assertNotIn("/expert-packs", dockerfile)


if __name__ == "__main__":
    unittest.main()
