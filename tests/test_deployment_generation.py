"""Generate and inspect AgentDeployment artifacts."""
import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from _helpers import ROOT

PROFILES = ("liki-expert",)
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
        build_id = subprocess.check_output(
            ["git", "-C", str(ROOT), "rev-parse", "--short=12", "HEAD"], text=True
        ).strip()
        for profile in PROFILES:
            with self.subTest(profile=profile):
                deployment = self._deployment(profile)
                self.assertNotIn("apiVersion", deployment)
                self.assertEqual(deployment["schema_version"], "agent.deployment.v1")
                self.assertNotIn("kind", deployment)
                self.assertEqual(deployment["metadata"]["name"], profile)
                self.assertEqual(deployment["metadata"]["version"], build_id)
                agents = deployment["spec"]["agents"]
                self.assertTrue(all(agent["version"] == build_id for agent in agents))

    def test_single_root_and_reachable_tree(self):
        deployment = self._deployment("liki-expert")
        agents = {a["name"]: a for a in deployment["spec"]["agents"]}
        self.assertEqual(agents["liki"]["sub_agents"], [])
        for profile in PROFILES:
            with self.subTest(profile=profile):
                deployment = self._deployment(profile)
                self.assertEqual(
                    [a["name"] for a in deployment["spec"]["agents"]], ["liki"]
                )

    def test_every_expert_is_a_single_chat_agent(self):
        deployment = self._deployment("liki-expert")
        agents = {a["name"]: a for a in deployment["spec"]["agents"]}
        self.assertEqual(agents["liki"]["mode"], "chat")
        for profile in PROFILES:
            deployment = self._deployment(profile)
            self.assertEqual(deployment["spec"]["agents"][0]["mode"], "chat")

    def test_natal_owns_the_full_domain_surface(self):
        """聚合专家拥有全部工具面（六域全集），无域排除。"""
        deployment = self._deployment("liki-expert")
        agents = {a["name"]: a for a in deployment["spec"]["agents"]}
        allow = agents["liki"]["tools"]["allow"]
        assert len(allow["engine"]) == 14, f"ENGINE tools = {len(allow['engine'])}"
        assert len(allow["counsel"]) == 10, f"COUNSEL tools = {len(allow['counsel'])}"
        assert "liki" not in allow
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
                    self.assertLess(path.stat().st_size, 6_144, f"{path} 超过骨架上限")
                    # 路由表引用入口卡名是设计意图（非烘焙内容）；方法论文本由 skilltoolset 按需加载
                    self.assertNotIn("## 流程", text)
                    self.assertNotIn("## 边界条件", text)
                    # 骨架仍须保留角色与硬边界
                    self.assertIn("边界", text)
                    # 语言协议是全局执行前提，必须随骨架进入 deployment。
                    self.assertIn("## 语言协议", text)
                    self.assertIn("推理、规划、规则比对和自我校验一律使用简体中文", text)

    def test_instructions_carry_no_delegation_section(self):
        """单 agent 拓扑：任何 instruction 不得残留子代理 / 委派语义。"""
        deployment = self._deployment("liki-expert")
        root = self.out / "liki-expert" / "agents"
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
        deployment = self._deployment("liki-expert")
        servers = {s["name"]: s for s in deployment["spec"]["mcpServers"]}
        self.assertEqual(servers["engine"]["endpointEnv"], "LIKI_ENGINE_URL")
        self.assertEqual(servers["counsel"]["endpointEnv"], "LIKI_COUNSEL_URL")

    def test_profiles_expose_complete_catalogued_tool_surface(self):
        catalog = json.loads(
            (ROOT / "tools" / "mcp-tool-catalog.json").read_text(encoding="utf-8")
        )
        catalog_tools = {
            server: set(definition["tools"])
            for server, definition in catalog["servers"].items()
        }
        assert set(catalog_tools) == {"engine", "counsel", "liki"}
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
        # ENGINE and COUNSEL are the product surfaces for expert; LIKI is not used.
        self.assertEqual(exposed["engine"], set(catalog_tools["engine"]) - {"liuyao_chart", "liuyao_qigua", "qimen_chart"})
        self.assertEqual(exposed["counsel"], set(catalog_tools["counsel"]))

    def test_instruction_body_comes_from_identity_source(self):
        """聚合专家 instruction＝唯一根身份正文 + 内部 host adapter。"""
        sys.path.insert(0, str(ROOT / "scripts"))
        import generate_deployment as gen  # noqa: E402

        for profile in PROFILES:
            self._generate(profile)
        identity = gen.identity_path("liki")
        self.assertEqual(gen.identity_path("liki"), identity)
        loaded_metadata, body = gen.load_identity("liki")
        instruction = (
            self.out / "liki-expert" / "agents" / "liki" / "instruction.md"
        ).read_text(encoding="utf-8")
        self.assertEqual(instruction, body)
        self.assertIn("内部档案同步", instruction)
        self.assertIn("sync_ui", instruction)
        self.assertIn("MCP", instruction)

    def test_single_expert_profile_is_the_only_release_profile(self):
        self.assertEqual(
            sorted(json.loads(path.read_text())["metadata"]["name"] for path in (ROOT / "distribution").glob("*/profile.json")),
            sorted(PROFILES),
        )


if __name__ == "__main__":
    unittest.main()
