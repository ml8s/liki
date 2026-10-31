"""Generate and inspect AgentDeployment artifacts."""
import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from _helpers import ROOT

PROFILES = ("liki",)
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

    def _deployment(self, profile):
        return json.loads(self._generate(profile).read_text(encoding="utf-8"))

    def test_generates_valid_deployment_for_all_profiles(self):
        runtime_version = (ROOT / "skills" / "liki" / "VERSION.txt").read_text(
            encoding="utf-8"
        ).strip()
        for profile in PROFILES:
            with self.subTest(profile=profile):
                deployment = self._deployment(profile)
                self.assertNotIn("apiVersion", deployment)
                self.assertEqual(deployment["schema_version"], "agent.deployment.v1")
                self.assertNotIn("kind", deployment)
                self.assertEqual(deployment["metadata"]["name"], profile)
                self.assertEqual(deployment["metadata"]["version"], runtime_version)
                agents = deployment["spec"]["agents"]
                self.assertTrue(all(agent["version"] == runtime_version for agent in agents))

    def test_single_root_and_reachable_tree(self):
        deployment = self._deployment("liki")
        agents = {a["name"]: a for a in deployment["spec"]["agents"]}
        self.assertEqual(agents["liki"]["sub_agents"], [])
        for profile in PROFILES:
            with self.subTest(profile=profile):
                deployment = self._deployment(profile)
                self.assertEqual(
                    [a["name"] for a in deployment["spec"]["agents"]], [profile]
                )

    def test_every_expert_is_a_single_chat_agent(self):
        deployment = self._deployment("liki")
        agents = {a["name"]: a for a in deployment["spec"]["agents"]}
        self.assertEqual(agents["liki"]["mode"], "chat")
        for profile in PROFILES:
            deployment = self._deployment(profile)
            self.assertEqual(deployment["spec"]["agents"][0]["mode"], "chat")

    def test_natal_owns_the_full_domain_surface(self):
        """聚合专家拥有全部工具面（六域全集），无域排除。"""
        deployment = self._deployment("liki")
        agents = {a["name"]: a for a in deployment["spec"]["agents"]}
        allow = agents["liki"]["tools"]["allow"]
        assert len(allow["engine"]) == 27, f"engine tools = {len(allow['engine'])}"
        assert len(allow["counsel"]) == 13, f"counsel tools = {len(allow['counsel'])}"
    def test_instructions_are_thin_and_delegate_to_skilltoolset(self):
        """2b 去烘焙：instruction 只留骨架，方法论唯一权威源是 skill（渐进加载）。"""
        for profile in PROFILES:
            manifest = self._generate(profile)
            root = manifest.parent
            deployment = json.loads(manifest.read_text(encoding="utf-8"))
            for agent in deployment["spec"]["agents"]:
                with self.subTest(profile=profile, agent=agent["name"]):
                    path = root / agent["instruction"]["path"]
                    text = path.read_text(encoding="utf-8")
                    self.assertLess(path.stat().st_size, 4_096, f"{path} 超过骨架上限")
                    # 路由表引用入口卡名是设计意图（非烘焙内容）；方法论文本由 skilltoolset 按需加载
                    self.assertNotIn("## 流程", text)
                    self.assertNotIn("## 边界条件", text)
                    # 骨架仍须保留角色与硬边界
                    self.assertIn("边界", text)

    def test_instructions_carry_no_delegation_section(self):
        """单 agent 拓扑：任何 instruction 不得残留子代理 / 委派语义。"""
        deployment = self._deployment("liki")
        root = self.out / "liki" / "agents"
        natal = (root / "liki" / "instruction.md").read_text(encoding="utf-8")
        for marker in ("委派", "子代理", "子能力", "协作委派", "同一轮只调用一个"):
            self.assertNotIn(marker, natal)
        self.assertEqual([a["name"] for a in deployment["spec"]["agents"]], ["liki"])

    def test_instruction_paths_resolve_within_deployment_dir(self):
        for profile in PROFILES:
            manifest = self._generate(profile)
            deployment = json.loads(manifest.read_text(encoding="utf-8"))
            root = manifest.parent
            for agent in deployment["spec"]["agents"]:
                with self.subTest(profile=profile, agent=agent["name"]):
                    path = root / agent["instruction"]["path"]
                    self.assertTrue(path.exists(), f"missing {path}")
                    self.assertLess(path.stat().st_size, 2 << 20)

    def test_mcp_servers_use_contract_endpoint_env_names(self):
        deployment = self._deployment("liki")
        servers = {s["name"]: s for s in deployment["spec"]["mcpServers"]}
        self.assertEqual(servers["engine"]["endpointEnv"], "LIKI_ENGINE_MCP_URL")
        self.assertEqual(servers["counsel"]["endpointEnv"], "LIKI_COUNSEL_MCP_URL")

    def test_profiles_expose_complete_catalogued_tool_surface(self):
        catalog = json.loads(
            (ROOT / "tools" / "mcp-tool-catalog.json").read_text(encoding="utf-8")
        )
        catalog_tools = {
            server: set(definition["tools"])
            for server, definition in catalog["servers"].items()
        }
        exposed = {server: set() for server in catalog_tools}
        for profile in PROFILES:
            deployment = self._deployment(profile)
            for agent in deployment["spec"]["agents"]:
                for server, allowed in agent["tools"]["allow"].items():
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

    def test_instruction_body_comes_from_identity_source(self):
        """聚合专家 instruction＝agents/liki/liki.md 正文。"""
        sys.path.insert(0, str(ROOT / "scripts"))
        import generate_deployment as gen  # noqa: E402

        for profile in PROFILES:
            self._generate(profile)
        identity = ROOT / "agents" / "liki" / "liki.md"
        self.assertEqual(gen.identity_path("liki"), identity)
        body = identity.read_text(encoding="utf-8").split("\n---\n", 1)[1].lstrip("\n").rstrip() + "\n"
        instruction = (
            self.out / "liki" / "agents" / "liki" / "instruction.md"
        ).read_text(encoding="utf-8")
        self.assertEqual(instruction, body)
        self.assertIn("定盘门", instruction)
        self.assertIn("MCP", instruction)

    def test_single_expert_profile_is_the_only_release_profile(self):
        self.assertEqual(
            sorted(path.stem for path in (ROOT / "profiles").glob("*.json")),
            sorted(PROFILES),
        )


if __name__ == "__main__":
    unittest.main()
