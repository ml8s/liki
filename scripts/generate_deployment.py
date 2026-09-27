"""AgentDeployment 工件生成器。

读 agents/ 发布结构 + profiles/ 定义，组装符合 agent.liki/v1 契约的
deployment 工件（deployment.json + instruction + output schema）。

用法:
    python3 scripts/generate_deployment.py --profile experts --out dist/agents

产物:
    dist/agents/<profile>/deployment.json
    dist/agents/<profile>/agents/<name>/instruction.md
    dist/agents/<profile>/agents/<name>/output-schema.json   (若定义)

方法论卡合并: 各 agent 的 instruction.md 为骨架，构建时把对应
skills/<family>/skills/<skill>/ 下除 SKILL.md 外的方法论 md 追加到末尾，
使工件自包含、可回溯。
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

# agent name -> methodology source directory. 权威源在根 skill（liki）各域
# domains/ 下；独立专家包由 scripts/sync-expert-methodology.sh 从根同步。
METHODOLOGY_MAP = {
    "bazi": _ROOT / "skills" / "liki" / "natal" / "domains" / "bazi",
    "ziwei": _ROOT / "skills" / "liki" / "natal" / "domains" / "ziwei",
    "liuyao": _ROOT / "skills" / "liki" / "divination" / "domains" / "liuyao",
    "qimen": _ROOT / "skills" / "liki" / "divination" / "domains" / "qimen",
    "fengshui": _ROOT / "skills" / "liki" / "fengshui" / "domains",
    "naming": _ROOT / "skills" / "liki" / "naming" / "domains",
}


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
            if server not in servers:
                sys.exit(f"agent {name} references unknown MCP server: {server}")
            catalog_tools = set(servers[server].get("tools", []))
            unknown = sorted(set(allowed) - catalog_tools)
            if unknown:
                sys.exit(
                    f"agent {name} allowlists tools absent from {server} catalog: "
                    + ", ".join(unknown)
                )


def merge_methodology(agent_name: str, instruction_path: Path) -> str:
    """把方法论 md 追加到 instruction 骨架后，返回合并后的完整指令文本。"""
    skeleton = instruction_path.read_text(encoding="utf-8").rstrip()

    # single_expert 是全能力单根 agent：直接合并根 skill（liki）全部 md
    # （SKILL.md 路由 + 各域 app 编排卡 + 全部方法论卡），自包含全能力。
    if agent_name == "single_expert":
        root_skill = _ROOT / "skills" / "liki"
        cards = sorted(
            p for p in root_skill.rglob("*.md")
            if p.name != "SKILL.md" and "__pycache__" not in p.parts
        )
        blocks = [skeleton]
        for card in cards:
            blocks.append(
                f"\n## 方法论卡：{card.relative_to(root_skill)}\n\n"
                f"{card.read_text(encoding='utf-8').strip()}"
            )
        return "\n\n".join(blocks) + "\n"

    source_dir = METHODOLOGY_MAP.get(agent_name)
    if source_dir is None or not source_dir.is_dir():
        return skeleton + "\n"
    cards = sorted(
        p for p in source_dir.rglob("*.md") if p.name != "SKILL.md"
    )
    if not cards:
        return skeleton + "\n"
    blocks = [skeleton]
    for card in cards:
        blocks.append(
            f"\n## 方法论卡：{card.stem}\n\n{card.read_text(encoding='utf-8').strip()}"
        )
    return "\n\n".join(blocks) + "\n"


def build_agent(name: str, profile_agent: dict, version: str, out_dir: Path) -> dict:
    definition = load_agent_yaml(name)
    if definition.get("name") != name:
        sys.exit(f"agent definition name mismatch: profile={name} yaml={definition.get('name')}")
    agent_dir = AGENTS_DIR / name

    instruction_text = merge_methodology(name, agent_dir / "instruction.md")
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
