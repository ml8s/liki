"""AgentDeployment 工件生成器。

读唯一专家身份 + `distribution/liki-expert/` 部署配置，组装 `liki-agent-runtime`
直接消费的 deployment 工件（deployment.json + instruction + output schema）。

身份单一源：`skills/liki/SKILL.md`。内部部署只追加
`distribution/liki-expert/host-adapter.md` 中的宿主协议；公开渠道不会得到该协议。
单 chat agent 直调本域工具；单侧/双侧执行路由由域卡（natal-ENTRY）承载，域边界由
工具白名单（profile + deployment.yaml）强制。

用法:
    python3 scripts/generate_deployment.py --profile liki-expert --out dist

产物:
    dist/<profile>/deployment.json
    dist/<profile>/agents/<name>/instruction.md
    dist/<profile>/agents/<name>/output-schema.json   (若定义)

方法论去重: instruction 只保留骨架（角色/能力/边界/路由）。方法论卡
的唯一权威源是 `skills/liki/references/`，
由 ADK skilltoolset 按需 load_skill / load_skill_resource 渐进加载，
不再烘焙进工件（避免双源与 ~245KB/轮 的 prompt 膨胀）。
"""

import argparse
import json
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

import yaml

_ROOT = Path(__file__).resolve().parent.parent
DISTRIBUTION_DIR = _ROOT / "distribution"
AGENTS_DIR = _ROOT / "agents"
EXPERT_PROFILE = "liki-expert"
EXPERT_CONFIG_DIR = DISTRIBUTION_DIR / EXPERT_PROFILE
TOOL_CATALOG_PATH = _ROOT / "tools" / "mcp-tool-catalog.json"
CANONICAL_SKILL = _ROOT / "skills" / "liki" / "SKILL.md"
FREE_CHANNEL_PATTERN = re.compile(
    r"<!-- channel:free -->.*?<!-- /channel:free -->\n?",
    re.S,
)


def identity_path(name: str) -> Path:
    """身份文件：聚合专家＝唯一根 Skill；子代理保留扩展位。"""
    if name == "liki":
        return CANONICAL_SKILL
    return AGENTS_DIR / name / f"{name}.md"


def load_identity(name: str) -> tuple[dict, str]:
    """唯一身份源 → (frontmatter dict, 部署 instruction 正文)。"""
    path = identity_path(name)
    if not path.is_file():
        sys.exit(f"identity source not found: {name} ({path})")
    text = path.read_text(encoding="utf-8")
    if not text.startswith("---\n"):
        sys.exit(f"identity source missing frontmatter start: {path.relative_to(_ROOT)}")
    end = text.find("\n---\n")
    if end < 0:
        sys.exit(f"identity source missing frontmatter end: {path.relative_to(_ROOT)}")
    frontmatter = yaml.safe_load(text[4:end])
    body = text[end + len("\n---\n"):]
    if body.startswith("\n"):
        body = body[1:]
    if name == "liki":
        body = FREE_CHANNEL_PATTERN.sub("", body)
        host_adapter = EXPERT_CONFIG_DIR / "host-adapter.md"
        if not host_adapter.is_file():
            sys.exit(f"internal host adapter not found: {host_adapter.relative_to(_ROOT)}")
        body += "\n" + host_adapter.read_text(encoding="utf-8").rstrip() + "\n"
        description_en = frontmatter.get("metadata", {}).get("description_en")
        if isinstance(description_en, str) and description_en.strip():
            frontmatter.setdefault("description_en", description_en)
    return frontmatter, body.rstrip() + "\n"


def load_profile(name: str) -> dict:
    if name != EXPERT_PROFILE:
        sys.exit(f"unsupported deployment profile: {name} (expected {EXPERT_PROFILE})")
    path = EXPERT_CONFIG_DIR / "profile.json"
    if not path.exists():
        sys.exit(f"profile not found: {name} ({path})")
    with open(path, encoding="utf-8") as fh:
        return json.load(fh)


def load_agent_yaml(name: str) -> dict:
    path = EXPERT_CONFIG_DIR / "deployment.yaml"
    if not path.exists():
        sys.exit(f"agent definition not found: {name} ({path})")
    with open(path, encoding="utf-8") as fh:
        return yaml.safe_load(fh)


def git_sha() -> str:
    """Return the short commit SHA used as the deployment build identity."""
    configured = os.environ.get("LIKI_GIT_SHA", "").strip()
    if configured:
        return configured
    return subprocess.check_output(
        ["git", "-C", str(_ROOT), "rev-parse", "--short=12", "HEAD"],
        text=True,
    ).strip()


def validate_profile_tool_catalog(profile: dict, catalog: dict) -> None:
    """Fail closed when a profile references a tool outside the MCP catalog."""
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


def validate_profile_graph(profile: dict, definitions: dict[str, dict]) -> None:
    """每个 profile 必须构成单根委派树（与运行时契约一致，先于生成 fail-closed）。"""
    names = [entry.get("name", "") for entry in profile.get("agents", [])]
    if any(not name for name in names) or len(set(names)) != len(names):
        sys.exit("profile agents must have unique non-empty names")
    declared = set(names)
    incoming = {name: 0 for name in names}
    callers: dict[str, str] = {}
    for name in names:
        seen: set[str] = set()
        for reference in definitions[name].get("sub_agents", []):
            target = reference.get("name", "")
            if target not in declared:
                sys.exit(
                    f"agent {name} references sub-agent {target!r} outside profile {profile['metadata'].get('name')}"
                )
            if target in seen:
                sys.exit(f"agent {name} declares duplicate sub-agent {target}")
            seen.add(target)
            if target in callers:
                sys.exit(
                    f"agent {target} has multiple parents ({callers[target]} and {name})"
                )
            callers[target] = name
            incoming[target] += 1
    roots = [name for name, count in incoming.items() if count == 0]
    if len(roots) != 1:
        sys.exit(f"profile must declare exactly one root agent, got {roots}")


SKILL_TOOLSET_TOOLS = ["list_skills", "load_skill", "load_skill_resource"]
# 镜像内合法 skill rootFS（assembly 镜像 COPY：/skills）。
# /skills 是聚合 skill 树；部署 agent 绑定该唯一 skill 根。
def validate_skills_binding(agent_name: str, skills: dict) -> None:
    root = (skills or {}).get("root", "")
    allowed_roots = {"/skills"}
    if root not in allowed_roots:
        sys.exit(
            f"agent {agent_name} skills.root {root!r} not in allowed roots: "
            + ", ".join(sorted(allowed_roots))
        )
    preload = skills.get("preload", "")
    if preload not in ("", "frontmatter", "complete"):
        sys.exit(f"agent {agent_name} skills.preload invalid: {preload!r}")


def build_agent(
    name: str,
    profile_agent: dict,
    version: str,
    out_dir: Path,
    definitions: dict[str, dict],
    identities: dict[str, tuple[dict, str]],
) -> dict:
    definition = definitions[name]
    if definition.get("name") != name:
        sys.exit(f"agent definition name mismatch: profile={name} yaml={definition.get('name')}")
    config_dir = EXPERT_CONFIG_DIR

    identity, instruction_text = identities[name]
    sub_agents = definition.get("sub_agents", [])
    instruction_path = out_dir / "agents" / name / "instruction.md"
    instruction_path.parent.mkdir(parents=True, exist_ok=True)
    instruction_path.write_text(instruction_text, encoding="utf-8")

    agent = {
        "name": definition["name"],
        "version": version,
        # 部署 agent 描述用英文面（plugin.json.description 同源）；缺省回落中文描述
        "description": identity.get("description_en") or identity.get("description") or definition["name"],
        "mode": definition.get("mode", "chat"),
        "sub_agents": sub_agents,
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

    output_schema = config_dir / "output-schema.json"
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
    parser.add_argument("--profile", required=True, help=f"deployment 名（{EXPERT_PROFILE}）")
    parser.add_argument("--out", default="dist", help="输出目录")
    args = parser.parse_args()

    profile = load_profile(args.profile)
    with open(TOOL_CATALOG_PATH, encoding="utf-8") as fh:
        tool_catalog = json.load(fh)
    validate_profile_tool_catalog(profile, tool_catalog)
    definitions = {
        entry["name"]: load_agent_yaml(entry["name"])
        for entry in profile["agents"]
    }
    identities = {
        entry["name"]: load_identity(entry["name"])
        for entry in profile["agents"]
    }
    validate_profile_graph(profile, definitions)
    metadata = profile["metadata"]
    metadata["version"] = git_sha()
    out_dir = Path(args.out) / args.profile
    shutil.rmtree(out_dir, ignore_errors=True)
    out_dir.mkdir(parents=True, exist_ok=True)

    agents = [
        build_agent(
            profile_agent["name"], profile_agent, git_sha(), out_dir,
            definitions, identities,
        )
        for profile_agent in profile["agents"]
    ]
    deployment = {
        "schema_version": "agent.deployment.v1",
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
