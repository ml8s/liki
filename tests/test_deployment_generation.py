"""Generate and inspect AgentDeployment artifacts."""
import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from _helpers import ROOT

PROFILES = ("natal", "divination", "fengshui", "naming")
EXPERT_DOMAINS = ("natal", "divination", "fengshui", "naming")
SUB_AGENTS = ("bazi", "ziwei")


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
                self.assertEqual(deployment["schema_version"], "liki.agent-deployment.v1")
                self.assertNotIn("kind", deployment)
                self.assertEqual(deployment["metadata"]["name"], profile)
                self.assertEqual(deployment["metadata"]["version"], runtime_version)
                agents = deployment["spec"]["agents"]
                self.assertTrue(all(agent["version"] == runtime_version for agent in agents))

    def test_single_root_and_reachable_tree(self):
        deployment = self._deployment("natal")
        agents = {a["name"]: a for a in deployment["spec"]["agents"]}
        self.assertEqual(
            agents["natal"]["sub_agents"],
            [{"name": "bazi"}, {"name": "ziwei"}],
        )
        self.assertEqual(agents["bazi"]["sub_agents"], [])
        self.assertEqual(agents["ziwei"]["sub_agents"], [])
        for profile in ("divination", "fengshui", "naming"):
            with self.subTest(profile=profile):
                deployment = self._deployment(profile)
                self.assertEqual(
                    [a["name"] for a in deployment["spec"]["agents"]], [profile]
                )

    def test_natal_is_chat_root_with_task_sub_agents(self):
        deployment = self._deployment("natal")
        agents = {a["name"]: a for a in deployment["spec"]["agents"]}
        self.assertEqual(agents["natal"]["mode"], "chat")
        self.assertEqual(agents["bazi"]["mode"], "task")
        self.assertEqual(agents["ziwei"]["mode"], "task")

    def test_natal_exposes_time_fact_surface(self):
        """natal root can read clock facts; domain sub-agents can resolve them."""
        deployment = self._deployment("natal")
        agents = {a["name"]: a for a in deployment["spec"]["agents"]}
        self.assertEqual(agents["natal"]["tools"]["allow"]["engine"], ["time_now"])
        for name in ("bazi", "ziwei"):
            allowed = set(agents[name]["tools"]["allow"]["engine"])
            self.assertIn("time_now", allowed)
            self.assertIn("tianwen_resolve", allowed)

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
                    self.assertNotIn("方法论卡", text)
                    # 骨架仍须保留角色与硬边界
                    self.assertIn("边界", text)

    def test_instructions_carry_no_delegation_section(self):
        """三分拓扑不写进任何 instruction：根靠 task 工具自省（描述含
        「不要并行调用」，由 ADK 注入）、路由靠 natal-ENTRY 卡、委派义务由
        根的能力边界（无排盘工具）强制。"""
        deployment = self._deployment("natal")
        root = self.out / "natal" / "agents"
        natal = (root / "natal" / "instruction.md").read_text(encoding="utf-8")
        for marker in ("协作委派", "子专家职责内的工作一律委派", "同一轮只调用一个"):
            self.assertNotIn(marker, natal)
        for sub_agent in SUB_AGENTS:
            body = (root / sub_agent / "instruction.md").read_text(encoding="utf-8")
            self.assertNotIn("协作委派", body, sub_agent)
        self.assertEqual(
            [a["name"] for a in deployment["spec"]["agents"]],
            ["natal", "bazi", "ziwei"],
        )

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
        deployment = self._deployment("natal")
        servers = {s["name"]: s for s in deployment["spec"]["mcpServers"]}
        self.assertEqual(servers["engine"]["endpointEnv"], "LIKI_MCP_ENGINE_URL")
        self.assertEqual(servers["counsel"]["endpointEnv"], "LIKI_MCP_COUNSEL_URL")

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

    def test_instruction_bodies_come_from_identity_sources(self):
        """域专家 instruction＝包根 SKILL.md 正文；子代理＝agents/<name>/<name>.md。"""
        sys.path.insert(0, str(ROOT / "scripts"))
        import generate_deployment as gen  # noqa: E402

        for profile in PROFILES:
            self._generate(profile)
        for domain in EXPERT_DOMAINS:
            identity = ROOT / "expert-packs" / f"liki-{domain}" / "SKILL.md"
            self.assertEqual(gen.identity_path(domain), identity)
            body = identity.read_text(encoding="utf-8").split("\n---\n", 1)[1].lstrip("\n").rstrip() + "\n"
            instruction = (
                self.out / domain / "agents" / domain / "instruction.md"
            ).read_text(encoding="utf-8")
            # instruction 就是身份正文本身：不做拓扑增补（子代理靠 tools/list
            # 自省、路由在域卡、委派义务由能力边界强制）。
            self.assertEqual(instruction, body, domain)
            self.assertNotIn("references/", instruction, domain)
        for sub_agent in SUB_AGENTS:
            identity = ROOT / "agents" / sub_agent / f"{sub_agent}.md"
            self.assertEqual(gen.identity_path(sub_agent), identity)
            body = identity.read_text(encoding="utf-8").split("\n---\n", 1)[1].lstrip("\n").rstrip() + "\n"
            instruction = (
                self.out / "natal" / "agents" / sub_agent / "instruction.md"
            ).read_text(encoding="utf-8")
            self.assertEqual(instruction, body, sub_agent)

    def test_four_expert_profiles_are_the_only_release_profiles(self):
        self.assertEqual(
            sorted(path.stem for path in (ROOT / "profiles").glob("*.json")),
            sorted(PROFILES),
        )


if __name__ == "__main__":
    unittest.main()
