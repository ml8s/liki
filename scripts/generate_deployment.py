"""AgentDeployment 工件生成器。

读 agents/ 发布结构 + profiles/ 定义，组装符合 agent.liki/v1 契约的
deployment 工件（deployment.json + instruction + output schema）。

用法:
    python3 scripts/generate_deployment.py --profile experts --out dist/agents

产物:
    dist/agents/<profile>/deployment.json
    dist/agents/<profile>/agents/<name>/instruction.md
    dist/agents/<profile>/agents/<name>/output-schema.json   (若定义)

方法论去重: instruction 只保留骨架（角色/能力/边界/路由）。方法论卡
的唯一权威源是 skill（liki / expert-packs 域 skill 的 references/），
由 ADK skilltoolset 按需 load_skill / load_skill_resource 渐进加载，
不再烘焙进工件（避免双源与 ~245KB/轮 的 prompt 膨胀）。
"""
from __future__ import annotations

import argparse
import json
import shutil
import sys
from pathlib import Path

import yaml

_ROOT = Path(__file__).resolve().parent.parent
AGENTS_DIR = _ROOT / "agents"
PROFILES_DIR = _ROOT / "profiles"
TOOL_CATALOG_PATH = _ROOT / "contracts" / "mcp-tool-catalog.json"
RUNTIME_VERSION_PATH = _ROOT / "skills" / "liki" / "VERSION.txt"

def load_profile(name: str) -> dict:
    path = PROFILES_DIR / f"{name}.json"
    if not path.exists():
        sys.exit(f"profile not found: {name} ({path})")
    with open(path, encoding="utf-8") as fh:
        return json.load(fh)


def load_agent_yaml(name: str) -> dict:
    path = AGENTS_DIR / name / "agent.yaml"
    if not path.exists():
        sys.exit(f"agent definition not found: {name} ({path})")
    with open(path, encoding="utf-8") as fh:
        return yaml.safe_load(fh)


def runtime_version() -> str:
    return RUNTIME_VERSION_PATH.read_text(encoding="utf-8").strip()


def validate_profile_tool_catalog(profile: dict, catalog: dict) -> None:
    """Fail closed when a profile references a tool outside the MCP catalog."""
    runtime_version = profile.get("metadata", {}).get("version", "")
    if runtime_version != catalog.get("runtime_version"):
        sys.exit(
            "profile/runtime tool catalog mismatch: "
            f"profile={runtime_version} catalog={catalog.get('runtime_version')}"
        )
    servers = catalog.get("servers", {})
    declared_servers: set[str] = set()
    for server in profile.get("mcpServers", []):
        name = server.get("name", "")
        if name not in servers:
            sys.exit(f"profile MCP server is not in tool catalog: {name}")
        if name in declared_servers:
            sys.exit(f"profile declares duplicate MCP server: {name}")
        declared_servers.add(name)

    for profile_agent in profile.get("agents", []):
        name = profile_agent.get("name", "")
        tools = (profile_agent.get("tools") or {}).get("allow", {})
        for server, allowed in tools.items():
            if server == "skilltoolset":
                builtin = {"list_skills", "load_skill", "load_skill_resource"}
                extra = sorted(set(allowed) - builtin)
                if extra:
                    sys.exit(
                        f"agent {name} allowlists unknown skilltoolset tools: "
                        + ", ".join(extra)
                    )
                continue
            if server not in servers:
                sys.exit(f"agent {name} references unknown MCP server: {server}")
            catalog_tools = set(servers[server].get("tools", []))
            unknown = sorted(set(allowed) - catalog_tools)
            if unknown:
                sys.exit(
                    f"agent {name} allowlists tools absent from {server} catalog: "
                    + ", ".join(unknown)
                )


SKILL_TOOLSET_TOOLS = ["list_skills", "load_skill", "load_skill_resource"]
# 镜像内合法 skill rootFS（assembly 镜像 COPY：/skills、/expert-packs）。
# /skills 的直接子目录是 liki/（router、single_expert 绑定后发现 skill "liki"）；
# /expert-packs/liki-<x>/skills 的直接子目录是域 skill（name==目录名，ADK S2）。
def validate_skills_binding(agent_name: str, skills: dict) -> None:
    root = (skills or {}).get("root", "")
    allowed_roots = {"/skills"} | {
        f"/expert-packs/liki-{n}/skills"
        for n in ("bazi", "ziwei", "liuyao", "qimen", "fengshui", "naming")
    }
    if root not in allowed_roots:
        sys.exit(
            f"agent {agent_name} skills.root {root!r} not in allowed roots: "
            + ", ".join(sorted(allowed_roots))
        )
    preload = skills.get("preload", "")
    if preload not in ("", "frontmatter", "complete"):
        sys.exit(f"agent {agent_name} skills.preload invalid: {preload!r}")


def build_agent(name: str, profile_agent: dict, version: str, out_dir: Path) -> dict:
    definition = load_agent_yaml(name)
    if definition.get("name") != name:
        sys.exit(f"agent definition name mismatch: profile={name} yaml={definition.get('name')}")
    agent_dir = AGENTS_DIR / name

    instruction_text = (agent_dir / "instruction.md").read_text(encoding="utf-8").rstrip() + "\n"
    instruction_path = out_dir / "agents" / name / "instruction.md"
    instruction_path.parent.mkdir(parents=True, exist_ok=True)
    instruction_path.write_text(instruction_text, encoding="utf-8")

    agent = {
        "name": definition["name"],
        "version": version,
        "description": definition["description"].strip(),
        "mode": definition.get("mode", "chat"),
        "sub_agents": definition.get("sub_agents", []),
        "instruction": {"path": f"agents/{name}/instruction.md"},
        "tools": profile_agent.get("tools") or definition.get("tools", {"allow": {}}),
    }
    skills = profile_agent.get("skills") or definition.get("skills")
    if not skills:
        sys.exit(f"agent {name} has no skills binding (agent.yaml skills.root)")
    validate_skills_binding(name, skills)
    agent["skills"] = skills
    allow = agent["tools"].get("allow", {})
    toolset_tools = allow.get("skilltoolset")
    if sorted(toolset_tools or []) != sorted(SKILL_TOOLSET_TOOLS):
        sys.exit(
            f"agent {name} tools.allow.skilltoolset must be exactly {SKILL_TOOLSET_TOOLS}"
        )

    output_schema = agent_dir / "output-schema.json"
    if output_schema.exists():
        schema_path = out_dir / "agents" / name / "output-schema.json"
        shutil.copyfile(output_schema, schema_path)
        agent["output"] = {
            "schema": {"path": f"agents/{name}/output-schema.json"},
            "textPointer": definition.get("textPointer", "/"),
        }
    return agent


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--profile", required=True, help="profile 名（profiles/<name>.json）")
    parser.add_argument("--out", default="dist/agents", help="输出目录")
    parser.add_argument(
        "--version",
        default=runtime_version(),
        help="AgentDeployment runtime CalVer（默认读取 skills/liki/VERSION.txt）",
    )
    args = parser.parse_args()

    profile = load_profile(args.profile)
    with open(TOOL_CATALOG_PATH, encoding="utf-8") as fh:
        tool_catalog = json.load(fh)
    validate_profile_tool_catalog(profile, tool_catalog)
    metadata = profile["metadata"]
    metadata["version"] = args.version
    out_dir = Path(args.out) / args.profile
    shutil.rmtree(out_dir, ignore_errors=True)
    out_dir.mkdir(parents=True, exist_ok=True)

    agents = [
        build_agent(profile_agent["name"], profile_agent, args.version, out_dir)
        for profile_agent in profile["agents"]
    ]
    deployment = {
        "apiVersion": "agent.liki/v1",
        "kind": "AgentDeployment",
        "metadata": metadata,
        "spec": {
            "mcpServers": profile["mcpServers"],
            "agents": agents,
        },
    }
    manifest_path = out_dir / "deployment.json"
    manifest_path.write_text(
        json.dumps(deployment, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    print(f"generated {manifest_path} ({manifest_path.stat().st_size} bytes)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
