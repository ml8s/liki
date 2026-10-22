#!/usr/bin/env python3
"""Guard the expert identity cards and their generated expert-pack surfaces.

身份文件 `expert-packs/liki-<domain>/SKILL.md` 是顶层专家身份的唯一手写源（正文＝部署
instruction 骨架，也是包内域 SKILL.md 的正文）；部署内子代理（如 natal 的
bazi/ziwei）身份在 `agents/<name>/<name>.md`，不进市场包。包内生成物的字节一致性
与版本面由 `scripts/sync_expert_packs.py --check`（`make check` 漂移门）与
`tests/test_version_surfaces.py` 守护；本脚本只守护那些门看不见的东西：手写身份文件的
契约与边界语义、`plugin.json.dependencies` 的连接器依赖（专家不自带 MCP）、
以及**全部描述面不得硬编码 MCP 方法名**。
另外守护头像矢量源不得夹带文本节点（字形一律转路径，市场宿主的字体与
`dominant-baseline` 支持都不可控）。
"""

import json
import re
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
import sync_expert_packs as packs  # noqa: E402  身份文件契约的唯一来源

CATALOG_PATH = ROOT / "tools" / "mcp-tool-catalog.json"
PACKS_DIR = ROOT / "expert-packs"
IDENTITY_SECTIONS = ("## 工具自省", "## 硬边界", "## 隐私与合规")
IDENTITY_PHRASES = ("条件性解读", "仅保留在当前会话")
AVATAR_TEXT_MARKERS = ("<text", "font-family", "dominant-baseline")
# 引擎方法名（点号形式，如 bazi.fullchart）：身份面与部署子代理身份同样不得硬编码；
# 能力发现一律走 tools/list 自举（字段名不在此列，另由 schema 同步测试守护）。
ENGINE_METHOD_RE = re.compile(
    r"(?<![A-Za-z0-9_])(?:bazi|ziwei|liuyao|qimen|huangli|bazhai|xuankong|time|city|tianwen)\.[a-z_]+"
)


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


def _check_no_methods(path: Path, method_names: set[str], errors: list[str]) -> None:
    if not path.is_file():
        return
    text = path.read_text(encoding="utf-8")
    for hit in _method_hits(text, method_names):
        errors.append(f"{path.relative_to(ROOT)}: hard-coded MCP method name `{hit}`")
    for hit in ENGINE_METHOD_RE.findall(text):
        errors.append(f"{path.relative_to(ROOT)}: hard-coded engine method name `{hit}`")


def _check_sections(path: Path, sections: tuple[str, ...], errors: list[str]) -> str:
    text = path.read_text(encoding="utf-8") if path.is_file() else ""
    for section in sections:
        if section not in text:
            errors.append(f"{path.relative_to(ROOT)}: missing {section}")
    return text


def _check_identity(domain: str, method_names: set[str], errors: list[str]) -> dict:
    """身份文件＝专家身份唯一手写源：契约（生成器）+ 边界语义（本脚本）。"""
    loaded = packs.load_identity(domain)
    if isinstance(loaded, list):
        errors.extend(loaded)
        return {}
    metadata, _body = loaded
    path = packs.identity_path(domain)
    text = _check_sections(path, IDENTITY_SECTIONS, errors)
    for phrase in IDENTITY_PHRASES:
        if phrase not in text:
            errors.append(f"{path.relative_to(ROOT)}: missing boundary phrase «{phrase}»")
    _check_no_methods(path, method_names, errors)
    return metadata


def _check_deployment_sub_agents(method_names: set[str], errors: list[str]) -> None:
    """部署内子代理（非顶层领域）的身份文件在 agents/<name>/<name>.md。

    子代理不进市场包：身份契约（name==目录名、description 驱动委派决策、
    边界语义、不得硬编码 MCP 方法名）由本脚本守护，与域专家同标准。
    """
    agents_dir = ROOT / "agents"
    if not agents_dir.is_dir():
        return
    for agent_dir in sorted(path for path in agents_dir.iterdir() if path.is_dir()):
        if agent_dir.name in packs.DOMAIN_PACKS:
            continue
        if not (agent_dir / "agent.yaml").is_file():
            continue
        identity = agent_dir / f"{agent_dir.name}.md"
        if not identity.is_file():
            errors.append(f"missing sub-agent identity: {identity.relative_to(ROOT)}")
            continue
        try:
            metadata = _frontmatter(identity)
        except Exception as exc:
            errors.append(str(exc))
            continue
        if metadata.get("name") != agent_dir.name:
            errors.append(
                f"{identity.relative_to(ROOT)}: name 须与目录名一致（{agent_dir.name}）"
            )
        description = metadata.get("description")
        if not isinstance(description, str) or not description.strip():
            errors.append(f"{identity.relative_to(ROOT)}: description 必填（驱动委派决策）")
        text = _check_sections(identity, IDENTITY_SECTIONS, errors)
        for phrase in IDENTITY_PHRASES:
            if phrase not in text:
                errors.append(f"{identity.relative_to(ROOT)}: missing boundary phrase «{phrase}»")
        _check_no_methods(identity, method_names, errors)


AGENT_FRONTMATTER_KEYS = ("name", "description", "displayName", "profession", "maxTurns")


def _check_agent_doc(pack: str, pack_dir: Path, plugin: dict, errors: list[str]) -> None:
    """WorkBuddy agent md（`agents/<pack>.md`）的路径约定与 frontmatter 契约。"""
    agents_dir = pack_dir / "agents"
    agent_doc = agents_dir / f"{pack}.md"
    if not agent_doc.is_file():
        errors.append(f"missing agent definition: {agent_doc.relative_to(ROOT)}")
        return
    extras = sorted(path for path in agents_dir.glob("*.md") if path != agent_doc)
    if extras:
        errors.append(
            f"{agents_dir.relative_to(ROOT)}: agents/ 只放身份文件投影的 {agent_doc.name}，"
            f"多出 {', '.join(path.name for path in extras)}"
        )
    if plugin.get("agents") != [f"./agents/{pack}.md"] or plugin.get("agentName") != pack:
        errors.append(
            f"{agent_doc.relative_to(ROOT)}: plugin.json 的 agents/agentName 须为 "
            f'["./agents/{pack}.md"] / {pack}'
        )
    try:
        metadata = _frontmatter(agent_doc)
    except ValueError as exc:
        errors.append(str(exc))
        return
    missing = [key for key in AGENT_FRONTMATTER_KEYS if key not in metadata]
    if missing:
        errors.append(f"{agent_doc.relative_to(ROOT)}: frontmatter 缺少 {', '.join(missing)}")
    if metadata.get("name") != pack:
        errors.append(f"{agent_doc.relative_to(ROOT)}: name 须与文件名一致（{pack}）")
    for key in ("displayName", "profession"):
        value = metadata.get(key)
        if not isinstance(value, dict) or set(value) != {"en", "zh"}:
            errors.append(f"{agent_doc.relative_to(ROOT)}: {key} 须为 en/zh 双语映射")
    if not isinstance(metadata.get("maxTurns"), int):
        errors.append(f"{agent_doc.relative_to(ROOT)}: maxTurns 须为整数")


def _check_dependencies(pack: str, pack_dir: Path, plugin_doc: dict, errors: list[str]) -> None:
    """专家依赖已上架连接器：`dependencies.connectors` 是投影值，且包内不得自带 MCP。

    平台在 `plugin.json` 未声明 `dependencies.mcpServers` 时会自动读取插件根 `.mcp.json`
    作为兜底；两个来源同时存在会让同一次召唤出现两套工具面与两组引导卡片，故 `.mcp.json`
    必须缺席。
    """
    rel = (pack_dir / ".codebuddy-plugin" / "plugin.json").relative_to(ROOT)
    ids, id_errors = packs.connector_ids()
    errors.extend(id_errors)
    dependencies = plugin_doc.get("dependencies")
    if not isinstance(dependencies, dict):
        errors.append(f'{rel}: 缺 `dependencies`（须为 {{"connectors": [...]}}）')
    else:
        if "mcpServers" in dependencies:
            errors.append(f"{rel}: `dependencies.mcpServers` 与连接器依赖互斥，专家不自带 MCP")
        if dependencies.get("connectors") != ids:
            errors.append(
                f"{rel}: `dependencies.connectors` 须等于 {ids}（单一源 workbuddy/connector-ids.json）"
            )
    if (pack_dir / ".mcp.json").is_file():
        errors.append(f"{pack}: 专家依赖连接器时不得带 `.mcp.json`（平台会作为兜底自动加载）")


def _check_pack(domain: str, method_names: set[str], errors: list[str]) -> None:
    """包内全部描述面（生成的 SKILL 入口 + 手写 plugin.json）都不得泄漏 MCP 方法名。"""
    pack = packs.DOMAIN_PACKS[domain]
    pack_dir = PACKS_DIR / pack
    identity_doc = pack_dir / "SKILL.md"
    nested_skill = pack_dir / "skills" / domain / "SKILL.md"
    plugin = pack_dir / ".codebuddy-plugin" / "plugin.json"

    agent_doc = pack_dir / "agents" / f"{pack}.md"
    for path in (identity_doc, agent_doc, nested_skill, plugin):
        if not path.is_file():
            errors.append(f"missing expert-pack file: {path.relative_to(ROOT)}")
        _check_no_methods(path, method_names, errors)

    if not plugin.is_file() or not agent_doc.is_file():
        return
    try:
        plugin_doc = json.loads(plugin.read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"{plugin.relative_to(ROOT)}: invalid JSON: {exc}")
        return
    _check_agent_doc(pack, pack_dir, plugin_doc, errors)
    _check_dependencies(pack, pack_dir, plugin_doc, errors)


def _check_avatar(domain: str, errors: list[str]) -> None:
    """头像：`expert.svg` 是 `expert.png` 的栅格化源（`make workbuddy-avatars`），只允许纯图形。

    带字形的源由 `scripts/gen_brand_marks.py` 直接写出路径字形，不允许 `<text>` / 字体依赖。
    """
    pack_dir = PACKS_DIR / packs.DOMAIN_PACKS[domain]
    svg = pack_dir / "avatars" / "expert.svg"
    png = pack_dir / "avatars" / "expert.png"
    for path in (svg, png):
        if not path.is_file():
            errors.append(f"missing expert-pack avatar: {path.relative_to(ROOT)}")
    if not svg.is_file():
        return
    text = svg.read_text(encoding="utf-8")
    if "<svg" not in text:
        errors.append(f"{svg.relative_to(ROOT)}: missing <svg> root element")
    for marker in AVATAR_TEXT_MARKERS:
        if marker in text:
            errors.append(f"{svg.relative_to(ROOT)}: avatar 矢量源只允许路径字形，出现 `{marker}`")


def main() -> int:
    errors: list[str] = []
    method_names = _method_names()

    _check_deployment_sub_agents(method_names, errors)
    for domain in packs.DOMAIN_PACKS:
        _check_identity(domain, method_names, errors)
        _check_pack(domain, method_names, errors)
        _check_avatar(domain, errors)

    if errors:
        for error in errors:
            print(f"✗ {error}", file=sys.stderr)
        return 1

    print("✓ 专家身份文件、expert-pack 描述与连接器契约符合 MCP 自省边界")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
