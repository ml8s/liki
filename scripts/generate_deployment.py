#!/usr/bin/env python3
"""Generate the single Liki expert AgentDeployment artifact."""

from __future__ import annotations

import argparse
import json
import shutil
import subprocess
import sys
from pathlib import Path
from typing import Any

import yaml

ROOT = Path(__file__).resolve().parents[1]
CONFIG_PATH = ROOT / "distribution" / "liki-expert" / "deployment.yaml"
HOST_ADAPTER_PATH = ROOT / "distribution" / "liki-expert" / "host-adapter.md"
CANONICAL_SKILL = ROOT / "skills" / "liki" / "SKILL.md"
PROFILE = "liki-expert"
AGENT_NAME = "liki"
SKILL_TOOLSET_TOOLS = {"list_skills", "load_skill", "load_skill_resource"}


def _read(path: Path) -> str:
    if not path.is_file():
        raise SystemExit(f"required file missing: {path.relative_to(ROOT)}")
    return path.read_text(encoding="utf-8")


def _split_skill() -> tuple[dict[str, Any], str]:
    text = _read(CANONICAL_SKILL)
    if not text.startswith("---\n"):
        raise SystemExit(f"canonical identity missing frontmatter: {CANONICAL_SKILL}")
    end = text.find("\n---\n")
    if end < 0:
        raise SystemExit(f"canonical identity has unterminated frontmatter: {CANONICAL_SKILL}")
    metadata = yaml.safe_load(text[4:end])
    if not isinstance(metadata, dict):
        raise SystemExit("canonical identity frontmatter must be an object")
    body = text[end + 5 :]
    return metadata, (body[1:] if body.startswith("\n") else body)


def git_sha() -> str:
    return subprocess.check_output(
        ["git", "-C", str(ROOT), "rev-parse", "--short=12", "HEAD"], text=True
    ).strip()


def load_config() -> dict[str, Any]:
    config = yaml.safe_load(_read(CONFIG_PATH))
    if not isinstance(config, dict):
        raise SystemExit(f"deployment config must be an object: {CONFIG_PATH}")
    for key in ("description", "mcpServers", "agent"):
        if key not in config:
            raise SystemExit(f"deployment config missing {key}")
    agent = config["agent"]
    if not isinstance(agent, dict) or agent.get("name") != AGENT_NAME:
        raise SystemExit(f"deployment config must define agent {AGENT_NAME}")
    if agent.get("mode") != "chat":
        raise SystemExit("the Liki expert must be a single chat agent")
    return config


def validate_agent(config: dict[str, Any]) -> dict[str, Any]:
    agent = config["agent"]
    server_names = {server["name"] for server in config["mcpServers"]}
    allow = agent.get("tools", {}).get("allow", {})
    unknown_servers = set(allow) - server_names - {"skilltoolset"}
    if unknown_servers:
        raise SystemExit(f"tool allowlist references unknown servers: {sorted(unknown_servers)}")
    if set(allow.get("skilltoolset", [])) != SKILL_TOOLSET_TOOLS:
        raise SystemExit(f"skilltoolset allowlist must be exactly {sorted(SKILL_TOOLSET_TOOLS)}")
    skills = agent.get("skills", {})
    if skills.get("root") != "/skills":
        raise SystemExit("the expert agent must bind skills.root=/skills")
    if skills.get("preload") not in (None, "", "frontmatter", "complete"):
        raise SystemExit(f"invalid skills.preload: {skills.get('preload')!r}")
    return agent


def build_instruction() -> str:
    _metadata, body = _split_skill()
    return body.rstrip() + "\n\n" + _read(HOST_ADAPTER_PATH).strip() + "\n"


def generate(out: Path) -> Path:
    config = load_config()
    agent_config = validate_agent(config)
    version = git_sha()
    out_dir = out / PROFILE
    shutil.rmtree(out_dir, ignore_errors=True)
    agent_dir = out_dir / "agents" / AGENT_NAME
    agent_dir.mkdir(parents=True, exist_ok=True)

    instruction_path = agent_dir / "instruction.md"
    instruction = build_instruction()
    instruction_path.write_text(instruction, encoding="utf-8")
    agent = {
        "name": AGENT_NAME,
        "version": version,
        "description": config["description"],
        "mode": config.get("mode", "chat"),
        "sub_agents": [],
        "instruction": {"path": f"agents/{AGENT_NAME}/instruction.md"},
        "tools": agent_config["tools"],
        "skills": agent_config["skills"],
    }
    deployment = {
        "metadata": {"name": PROFILE, "version": version},
        "spec": {"mcpServers": config["mcpServers"], "agents": [agent]},
    }
    manifest = out_dir / "deployment.json"
    manifest.write_text(json.dumps(deployment, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"generated {manifest} ({manifest.stat().st_size} bytes)")
    return manifest


def main() -> int:
    parser = argparse.ArgumentParser(description="Generate the Liki expert deployment")
    parser.add_argument("--out", type=Path, default=Path("dist"))
    args = parser.parse_args()
    generate(args.out)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
