#!/usr/bin/env python3
"""Guard the single expert identity and its generated channel surfaces.

唯一专家身份是 `skills/liki/SKILL.md`；`distribution/liki-expert/host-adapter.md` 只承载内部
宿主协议，不进市场包。WorkBuddy 包内生成物的字节一致性与版本面由
`scripts/build_channel_package.py --check`（`make check` 漂移门）守护；本脚本守护渠道
适配层契约：身份边界语义、付费 overlay、`.mcp.json` 依赖边界，以及
**全部描述面不得硬编码 MCP 方法名**。
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
import build_channel_package as packs  # noqa: E402  身份文件契约的唯一来源

CATALOG_PATH = ROOT / "tools" / "mcp-tool-catalog.json"
PACKS_DIR = ROOT / "distribution" / "workbuddy"
IDENTITY_SECTIONS = ("## MCP 前置", "## 统一硬边界", "## 输出契约")
IDENTITY_PHRASES = ("条件性解读", "仅保留在当前会话")
AVATAR_TEXT_MARKERS = ("<text", "font-family", "dominant-baseline")
# 引擎方法名（点号形式，如 bazi.chart）：身份面与部署子代理身份同样不得硬编码；
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
    """唯一专家身份：契约（生成器）+ 边界语义（本脚本）。"""
    metadata, _body, _skill_name = packs.canonical_identity("pro")
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
    边界语义、不得硬编码 MCP 方法名）由本脚本守护。
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
    """付费专家包通过 `.mcp.json` 声明 MCP 端点（liki-mcp paywall）。

    WorkBuddy pro 由 variant 配置声明 `liki-mcp` 网关；zip 构建器会生成包内
    `.mcp.json`。源码渠道目录不需要保存第二份 `.mcp.json`。
    """
    rel = (pack_dir / ".codebuddy-plugin" / "plugin.json").relative_to(ROOT)
    dependencies = plugin_doc.get("dependencies")
    if dependencies is not None:
        if "connectors" in dependencies:
            errors.append(f"{rel}: `dependencies.connectors` 已废弃（连接器不再支持个人注册）")
    try:
        variant = packs.load_variant("pro")
    except SystemExit as exc:
        errors.append(str(exc))
        return
    if not variant.get("payment", {}).get("enabled"):
        errors.append(f"{pack}: pro variant must enable payment")
    if set(variant.get("mcpServers", {})) != {"liki"}:
        errors.append(f"{pack}: pro variant must use only the liki-mcp gateway")


def _check_pack(domain: str, method_names: set[str], errors: list[str]) -> None:
    """包内全部描述面（生成的入口 + plugin.json）都不得泄漏 MCP 方法名。"""
    pack = packs.DOMAIN_PACKS[domain]
    pack_dir = PACKS_DIR / pack
    identity_doc = packs.identity_path(domain)
    payment_overlay = packs.overlay_path(domain)
    plugin = pack_dir / ".codebuddy-plugin" / "plugin.json"

    for path in (identity_doc, payment_overlay, plugin):
        if not path.is_file():
            errors.append(f"missing expert-pack file: {path.relative_to(ROOT)}")
        _check_no_methods(path, method_names, errors)

    if not plugin.is_file():
        return
    try:
        plugin_doc = json.loads(plugin.read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"{plugin.relative_to(ROOT)}: invalid JSON: {exc}")
        return
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

    print("✓ 唯一专家身份、WorkBuddy 渠道配置与 MCP 自省边界符合契约")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
