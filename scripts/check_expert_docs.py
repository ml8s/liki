#!/usr/bin/env python3
"""Guard agent and expert-pack descriptions against MCP implementation leakage."""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
CATALOG_PATH = ROOT / "tools" / "mcp-tool-catalog.json"
PERSONAS_DIR = ROOT / "agents" / "personas"
PACKS_DIR = ROOT / "expert-packs"
RUNTIME_VERSION_PATH = ROOT / "skills" / "liki" / "VERSION.txt"
DOMAINS = ("bazi", "ziwei", "liuyao", "qimen", "fengshui", "naming")


def _frontmatter(path: Path) -> dict:
    text = path.read_text(encoding="utf-8")
    if not text.startswith("---\n"):
        raise ValueError(f"missing frontmatter: {path.relative_to(ROOT)}")
    end = text.find("\n---\n")
    if end < 0:
        raise ValueError(f"unterminated frontmatter: {path.relative_to(ROOT)}")
    metadata = yaml.safe_load(text[4:end])
    if not isinstance(metadata, dict):
        raise ValueError(f"frontmatter must be an object: {path.relative_to(ROOT)}")
    return metadata


def _method_names() -> set[str]:
    catalog = json.loads(CATALOG_PATH.read_text(encoding="utf-8"))
    return {
        name
        for server in catalog.get("servers", {}).values()
        for name in server.get("tools", [])
    }


def _method_hits(text: str, method_names: set[str]) -> list[str]:
    hits: list[str] = []
    for name in sorted(method_names):
        pattern = r"(?<![A-Za-z0-9_])" + re.escape(name) + r"(?![A-Za-z0-9_])"
        if re.search(pattern, text):
            hits.append(name)
    return hits


def _check_persona(path: Path, method_names: set[str], errors: list[str]) -> dict:
    rel = path.relative_to(ROOT)
    try:
        metadata = _frontmatter(path)
    except Exception as exc:
        errors.append(str(exc))
        return {}
    for key in ("name", "description", "displayName", "profession"):
        if key not in metadata:
            errors.append(f"{rel}: persona frontmatter missing {key}")
    text = path.read_text(encoding="utf-8")
    if "## 工具自省" not in text:
        errors.append(f"{rel}: missing 工具自省 section")
    if "## 隐私与合规" not in text:
        errors.append(f"{rel}: missing 隐私与合规 section")
    for hit in _method_hits(text, method_names):
        errors.append(f"{rel}: hard-coded MCP method name `{hit}`")
    return metadata


def _expected_agent_metadata(name: str) -> dict:
    return {
        "name": f"{name}-expert",
        "skills": [name],
    }


def _check_pack(pack: str, method_names: set[str], errors: list[str]) -> None:
    pack_dir = PACKS_DIR / pack
    root_skill = pack_dir / "SKILL.md"
    nested_skill = pack_dir / "skills" / pack.removeprefix("liki-") / "SKILL.md"
    agent_copy = pack_dir / "agents" / f"{pack.removeprefix('liki-')}-expert.md"
    plugin = pack_dir / ".codebuddy-plugin" / "plugin.json"
    mcp_manifest = pack_dir / ".mcp.json"
    rel = lambda path: str(path.relative_to(ROOT))  # noqa: E731

    for path in (root_skill, nested_skill, agent_copy, plugin, mcp_manifest):
        if not path.is_file():
            errors.append(f"missing expert-pack file: {rel(path)}")

    runtime_version = RUNTIME_VERSION_PATH.read_text(encoding="utf-8").strip()
    root_metadata: dict = {}
    nested_metadata: dict = {}
    persona_path = PERSONAS_DIR / f"{pack.removeprefix('liki-')}.md"
    persona_metadata: dict = {}

    try:
        root_metadata = _frontmatter(root_skill)
    except Exception as exc:
        errors.append(str(exc))
    try:
        nested_metadata = _frontmatter(nested_skill)
    except Exception as exc:
        errors.append(str(exc))
    try:
        persona_metadata = _frontmatter(persona_path)
    except Exception as exc:
        errors.append(str(exc))

    for label, metadata in (("pack root", root_metadata), ("nested", nested_metadata)):
        version = metadata.get("version")
        if label == "nested":
            version = metadata.get("metadata", {}).get("version")
        if version != runtime_version:
            errors.append(
                f"{label} skill version is {version!r}, expected {runtime_version}"
            )

    root_text = root_skill.read_text(encoding="utf-8") if root_skill.is_file() else ""
    nested_text = nested_skill.read_text(encoding="utf-8") if nested_skill.is_file() else ""
    for path, text in ((root_skill, root_text), (nested_skill, nested_text)):
        for hit in _method_hits(text, method_names):
            errors.append(f"{rel(path)}: hard-coded MCP method name `{hit}`")

    if root_skill.is_file():
        for section in ("## 工具自举", "## 定位与边界"):
            if section not in root_text:
                errors.append(f"{rel(root_skill)}: missing {section}")
        if "出生信息仅保留在当前会话" not in root_text:
            errors.append(f"{rel(root_skill)}: missing privacy/session boundary")
        if "条件性解读" not in root_text:
            errors.append(f"{rel(root_skill)}: missing conditional-interpretation boundary")

    expected_prefixes = {
            "bazi": "八字方法论卡——",
            "ziwei": "紫微方法论卡——",
            "liuyao": "六爻方法论卡——",
            "qimen": "奇门方法论卡——",
            "fengshui": "风水方法论卡——",
            "naming": "起名方法论卡——",
        }
    if nested_metadata:
        domain = pack.removeprefix("liki-")
        if not nested_metadata.get("description", "").startswith(expected_prefixes[domain]):
            errors.append(f"{rel(nested_skill)}: nested description is not a methodology card")

    plugin_doc = {}
    try:
        plugin_doc = json.loads(plugin.read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"{rel(plugin)}: invalid JSON: {exc}")

    if plugin_doc.get("version") != runtime_version:
        errors.append(f"{rel(plugin)}: plugin version must be {runtime_version}")
    if persona_metadata:
        for key in ("description", "profession"):
            if plugin_doc.get(key) != persona_metadata.get(key):
                errors.append(
                    f"{rel(plugin)}: plugin {key} does not match persona {key}"
                )

    for path in (agent_copy, plugin, nested_skill):
        if path.is_file():
            for hit in _method_hits(path.read_text(encoding="utf-8"), method_names):
                errors.append(f"{rel(path)}: hard-coded MCP method name `{hit}`")

    try:
        mcp = json.loads(mcp_manifest.read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"{rel(mcp_manifest)}: invalid JSON: {exc}")
        return

    descriptions: list[str] = []
    for server in mcp.get("mcpServers", {}).values():
        metadata = server.get("x-workbuddy", {}).get("description", {})
        descriptions.extend(
            value for key, value in metadata.items() if key in ("zh", "en") and isinstance(value, str)
        )
    for text in descriptions:
        for hit in _method_hits(text, method_names):
            errors.append(f"{rel(mcp_manifest)}: connector description hard-codes `{hit}`")


def main() -> int:
    errors: list[str] = []
    method_names = _method_names()

    for domain in DOMAINS:
        _check_persona(PERSONAS_DIR / f"{domain}.md", method_names, errors)

    for domain in DOMAINS:
        _check_pack(f"liki-{domain}", method_names, errors)

    if errors:
        for error in errors:
            print(f"✗ {error}", file=sys.stderr)
        return 1

    print("✓ agent / expert-pack descriptions follow the MCP introspection boundary")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
