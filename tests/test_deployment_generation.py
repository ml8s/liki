"""Generate and inspect AgentDeployment artifacts."""
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
        self._tmp = tempfile.mkdtemp(prefix="liki-agent-runtime-")
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
        self.assertNotIn("apiVersion", deployment)
        self.assertEqual(deployment["schema_version"], "liki.agent-deployment.v1")
        self.assertNotIn("kind", deployment)
        self.assertEqual(deployment["metadata"]["name"], "experts")
        self.assertEqual(deployment["metadata"]["version"], runtime_version)
        agents = deployment["spec"]["agents"]
        self.assertEqual(
            [a["name"] for a in agents],
            ["router", "natal", "divination", "fengshui", "naming"],
        )
        self.assertTrue(all(agent["version"] == runtime_version for agent in agents))

    def test_single_root_and_reachable_tree(self):
        manifest = self._generate("experts")
        deployment = json.loads(manifest.read_text(encoding="utf-8"))
        agents = {a["name"]: a for a in deployment["spec"]["agents"]}
        self.assertEqual(
            agents["router"]["sub_agents"],
            [
                {"name": "natal"},
                {"name": "divination"},
                {"name": "fengshui"},
                {"name": "naming"},
            ],
        )
        for expert in ["natal", "divination", "fengshui", "naming"]:
            self.assertEqual(agents[expert]["sub_agents"], [])

    def test_instructions_are_thin_and_delegate_to_skilltoolset(self):
        """2b 去烘焙：instruction 只留骨架，方法论唯一权威源是 skill（渐进加载）。"""
        for profile in ("experts",):
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
            (ROOT / "tools" / "mcp-tool-catalog.json").read_text(encoding="utf-8")
        )
        catalog_tools = {
            server: set(definition["tools"])
            for server, definition in catalog["servers"].items()
        }
        deployment = json.loads(self._generate("experts").read_text(encoding="utf-8"))
        agents = {agent["name"]: agent for agent in deployment["spec"]["agents"]}
        exposed = {server: set() for server in catalog_tools}
        for expert in ["natal", "divination", "fengshui", "naming"]:
            for server, allowed in agents[expert]["tools"]["allow"].items():
                allowed = set(allowed)
                if server == "skilltoolset":
                    # builtin（非 MCP server）：绑定由 test_skills_binding.py 断言
                    self.assertEqual(
                        allowed, {"list_skills", "load_skill", "load_skill_resource"}
                    )
                    continue
                self.assertTrue(allowed <= catalog_tools[server])
                exposed[server].update(allowed)
        self.assertEqual(exposed, {server: set(tools) for server, tools in catalog_tools.items()})

    def test_expert_instructions_come_from_pack_identity_files(self):
        """域专家 instruction＝包根 SKILL.md 正文（唯一手写源）；router 取 agents/router。"""
        sys.path.insert(0, str(ROOT / "scripts"))
        import generate_deployment as gen  # noqa: E402

        self._generate("experts")
        for domain in ("natal", "divination", "fengshui", "naming"):
            identity = ROOT / "expert-packs" / f"liki-{domain}" / "SKILL.md"
            self.assertEqual(gen.identity_path(domain), identity)
            body = identity.read_text(encoding="utf-8").split("\n---\n", 1)[1].lstrip("\n").rstrip() + "\n"
            instruction = (
                self.out / "experts" / "agents" / domain / "instruction.md"
            ).read_text(encoding="utf-8")
            self.assertEqual(instruction, body, domain)
            self.assertNotIn("references/", instruction, domain)
        self.assertEqual(gen.identity_path("router"), ROOT / "agents" / "router" / "router.md")

    def test_expert_profile_is_the_only_release_profile(self):
        self.assertEqual(
            sorted(path.stem for path in (ROOT / "profiles").glob("*.json")),
            ["experts"],
        )


if __name__ == "__main__":
    unittest.main()
