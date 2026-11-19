"""Contracts for the single Liki expert deployment and assembly image."""

import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from _helpers import ROOT


class ExpertDeploymentTest(unittest.TestCase):
    def setUp(self):
        self._tmp = tempfile.mkdtemp(prefix="liki-expert-")
        self.out = Path(self._tmp)
        self.addCleanup(shutil.rmtree, self._tmp, ignore_errors=True)

    def generate(self):
        script = ROOT / "scripts/generate_deployment.py"
        subprocess.run(
            [sys.executable, str(script), "--profile", "liki-expert", "--out", str(self.out)],
            check=True,
        )
        manifest = self.out / "liki-expert" / "deployment.json"
        return json.loads(manifest.read_text(encoding="utf-8")), manifest.parent

    def test_generates_single_chat_expert(self):
        deployment, _root = self.generate()
        build_id = subprocess.check_output(
            ["git", "-C", str(ROOT), "rev-parse", "--short=12", "HEAD"], text=True
        ).strip()
        self.assertEqual(deployment["metadata"], {"name": "liki-expert", "version": build_id})
        agents = deployment["spec"]["agents"]
        self.assertEqual(len(agents), 1)
        agent = agents[0]
        self.assertEqual(agent["name"], "liki")
        self.assertEqual(agent["version"], build_id)
        self.assertEqual(agent["mode"], "chat")
        self.assertEqual(agent["sub_agents"], [])
        self.assertTrue(agent["description"].startswith("Liki Astrologer (Pro) covers"))

    def test_expert_has_the_complete_skill_surface(self):
        deployment, _root = self.generate()
        agent = deployment["spec"]["agents"][0]
        allow = agent["tools"]["allow"]
        self.assertEqual(len(allow["engine"]), 15)
        self.assertEqual(len(allow["counsel"]), 10)
        self.assertNotIn("liki", allow)
        self.assertIn("liuyao_qigua", allow["engine"])
        self.assertNotIn("liuyao_chart", allow["engine"])
        self.assertIn("liuyao_ask", allow["counsel"])
        self.assertNotIn("liuyao_chart", allow["counsel"])

    def test_expert_binds_the_packaged_skill_root(self):
        deployment, _root = self.generate()
        agent = deployment["spec"]["agents"][0]
        self.assertEqual(agent["skills"], {"root": "/skills"})
        self.assertEqual(
            set(agent["tools"]["allow"]["skilltoolset"]),
            {"list_skills", "load_skill", "load_skill_resource"},
        )

    def test_instruction_is_canonical_skill_plus_internal_host_adapter(self):
        deployment, root = self.generate()
        agent = deployment["spec"]["agents"][0]
        path = root / agent["instruction"]["path"]
        text = path.read_text(encoding="utf-8")
        self.assertLess(path.stat().st_size, 6_144)
        self.assertIn("## 统一硬边界", text)
        self.assertIn("## 内部档案同步", text)
        self.assertIn("sync_ui", text)
        self.assertIn("standard_time", text)
        self.assertNotIn("## 流程", text)
        self.assertNotIn("## 边界条件", text)

    def test_assembly_dockerfile_copies_generated_deployment_and_skills(self):
        dockerfile = (ROOT / "assembly/Dockerfile").read_text(encoding="utf-8")
        self.assertIn("ARG BASE_IMAGE", dockerfile)
        self.assertIn("COPY dist/deployment/ ${DEPLOYMENT_DIR}/", dockerfile)
        self.assertIn("COPY dist/skills/ /skills/", dockerfile)


if __name__ == "__main__":
    unittest.main()
