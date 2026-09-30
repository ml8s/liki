"""Contract: AgentDeployment artifacts are generated, validated, and complete."""
import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from _helpers import ROOT


class TestDeploymentGeneration(unittest.TestCase):
    def setUp(self):
        self._tmp = tempfile.mkdtemp(prefix="liki-agents-")
        self.out = Path(self._tmp) / "out"
        self.addCleanup(shutil.rmtree, self._tmp, ignore_errors=True)

    def _generate(self, profile):
        script = ROOT / "scripts" / "generate_deployment.py"
        subprocess.run(
            [sys.executable, str(script), "--profile", profile, "--out", str(self.out)],
            check=True,
        )
        return self.out / profile / "deployment.json"

    def test_generates_valid_deployment_for_all_profiles(self):
        runtime_version = (ROOT / "skills" / "liki" / "VERSION.txt").read_text(
            encoding="utf-8"
        ).strip()
        manifest = self._generate("experts")
        deployment = json.loads(manifest.read_text(encoding="utf-8"))
        self.assertEqual(deployment["apiVersion"], "agent.liki/v1")
        self.assertEqual(deployment["kind"], "AgentDeployment")
        self.assertEqual(deployment["metadata"]["name"], "experts")
        self.assertEqual(deployment["metadata"]["version"], runtime_version)
        agents = deployment["spec"]["agents"]
        self.assertEqual(
            [a["name"] for a in agents],
            ["router", "bazi", "ziwei", "liuyao", "qimen", "fengshui", "naming"],
        )
        self.assertTrue(all(agent["version"] == runtime_version for agent in agents))

    def test_single_root_and_reachable_tree(self):
        manifest = self._generate("experts")
        deployment = json.loads(manifest.read_text(encoding="utf-8"))
        agents = {a["name"]: a for a in deployment["spec"]["agents"]}
        self.assertEqual(
            agents["router"]["sub_agents"],
            [
                {"name": "bazi"},
                {"name": "ziwei"},
                {"name": "liuyao"},
                {"name": "qimen"},
                {"name": "fengshui"},
                {"name": "naming"},
            ],
        )
        for expert in ["bazi", "ziwei", "liuyao", "qimen", "fengshui", "naming"]:
            self.assertEqual(agents[expert]["sub_agents"], [])

    def test_instructions_are_thin_and_delegate_to_skilltoolset(self):
        """2b 去烘焙：instruction 只留骨架，方法论唯一权威源是 skill（渐进加载）。"""
        for profile in ("experts", "single"):
            manifest = self._generate(profile)
            root = manifest.parent
            deployment = json.loads(manifest.read_text(encoding="utf-8"))
            for agent in deployment["spec"]["agents"]:
                with self.subTest(profile=profile, agent=agent["name"]):
                    path = root / agent["instruction"]["path"]
                    text = path.read_text(encoding="utf-8")
                    self.assertLess(path.stat().st_size, 4_096, f"{path} 超过骨架上限")
                    self.assertNotIn("方法论卡", text)
                    # 骨架仍须保留角色与硬边界
                    self.assertIn("边界", text)

    def test_instruction_paths_resolve_within_deployment_dir(self):
        manifest = self._generate("experts")
        deployment = json.loads(manifest.read_text(encoding="utf-8"))
        root = manifest.parent
        for agent in deployment["spec"]["agents"]:
            path = root / agent["instruction"]["path"]
            self.assertTrue(path.exists(), f"missing {path}")
            self.assertLess(path.stat().st_size, 2 << 20)

    def test_mcp_servers_use_contract_endpoint_env_names(self):
        manifest = self._generate("experts")
        deployment = json.loads(manifest.read_text(encoding="utf-8"))
        servers = {s["name"]: s for s in deployment["spec"]["mcpServers"]}
        self.assertEqual(servers["engine"]["endpointEnv"], "LIKI_MCP_ENGINE_URL")
        self.assertEqual(servers["counsel"]["endpointEnv"], "LIKI_MCP_COUNSEL_URL")

    def test_single_profile_exposes_complete_catalogued_tool_surface(self):
        catalog = json.loads(
            (ROOT / "contracts" / "mcp-tool-catalog.json").read_text(encoding="utf-8")
        )
        catalog_tools = {
            server: set(definition["tools"])
            for server, definition in catalog["servers"].items()
        }
        deployment = json.loads(self._generate("experts").read_text(encoding="utf-8"))
        agents = {agent["name"]: agent for agent in deployment["spec"]["agents"]}
        exposed = {server: set() for server in catalog_tools}
        for expert in ["bazi", "ziwei", "liuyao", "qimen", "fengshui", "naming"]:
            for server, allowed in agents[expert]["tools"]["allow"].items():
                allowed = set(allowed)
                if server == "skilltoolset":
                    # builtin（非 MCP server）：契约由 test_skills_binding.py 断言
                    self.assertEqual(
                        allowed, {"list_skills", "load_skill", "load_skill_resource"}
                    )
                    continue
                self.assertTrue(allowed <= catalog_tools[server])
                exposed[server].update(allowed)
        self.assertEqual(exposed, {server: set(tools) for server, tools in catalog_tools.items()})

    def test_expected_profiles_exist(self):
        self.assertEqual(
            sorted(path.stem for path in (ROOT / "profiles").glob("*.json")),
            ["experts", "single"],
        )

    def test_single_profile_is_one_root_agent_with_full_tools(self):
        manifest = self._generate("single")
        deployment = json.loads(manifest.read_text(encoding="utf-8"))
        agents = deployment["spec"]["agents"]
        self.assertEqual([a["name"] for a in agents], ["single_expert"])
        self.assertEqual(agents[0]["sub_agents"], [])
        # 全能力：工具覆盖 engine + counsel 全部领域
        self.assertGreater(len(agents[0]["tools"]["allow"]["engine"]), 20)
        self.assertGreater(len(agents[0]["tools"]["allow"]["counsel"]), 8)


if __name__ == "__main__":
    unittest.main()
