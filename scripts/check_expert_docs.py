#!/usr/bin/env python3
"""Guard the expert identity cards and their generated expert-pack surfaces.

身份文件 `expert-packs/liki-<domain>/SKILL.md` 是专家身份的唯一手写源（正文＝部署
instruction 骨架，也是包内域 SKILL.md 的正文）。包内生成物的字节一致性与
版本面由 `scripts/sync_expert_packs.py --check`（`make check` 漂移门）与
`tests/test_version_surfaces.py` 守护；本脚本只守护那些门看不见的东西：手写身份文件的
契约与边界语义、手写连接器/市场描述的有效性，以及**全部描述面不得硬编码 MCP 方法名**。
另外守护头像矢量源不得夹带文本节点（字形一律转路径，市场宿主的字体与
`dominant-baseline` 支持都不可控）。
"""
from __future__ import annotations

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
ROUTER_IDENTITY = ROOT / "agents" / "router" / "router.md"
IDENTITY_SECTIONS = ("## 工具自省", "## 硬边界", "## 隐私与合规")
IDENTITY_PHRASES = ("条件性解读", "仅保留在当前会话")
AVATAR_TEXT_MARKERS = ("<text", "font-family", "dominant-baseline")


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
    for hit in _method_hits(path.read_text(encoding="utf-8"), method_names):
        errors.append(f"{path.relative_to(ROOT)}: hard-coded MCP method name `{hit}`")


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


def _check_router(method_names: set[str], errors: list[str]) -> None:
    """router 无专家包，身份文件就在 agents/router（部署 instruction 来源）。"""
    if not ROUTER_IDENTITY.is_file():
        errors.append(f"missing router identity: {ROUTER_IDENTITY.relative_to(ROOT)}")
        return
    try:
        metadata = _frontmatter(ROUTER_IDENTITY)
    except Exception as exc:
        errors.append(str(exc))
        return
    for key in ("name", "description"):
        if key not in metadata:
            errors.append(f"{ROUTER_IDENTITY.relative_to(ROOT)}: router frontmatter missing {key}")
    _check_sections(ROUTER_IDENTITY, ("## 专家路由", "## 隐私与合规"), errors)
    _check_no_methods(ROUTER_IDENTITY, method_names, errors)


def _check_card_icon(
    pack_dir: Path, icon: object, mcp_manifest: Path, errors: list[str]
) -> None:
    """引导卡片的 `x-workbuddy.icon` 必须落在包内且真实存在（相对 .mcp.json 所在目录解析）。"""
    if icon is None:
        return
    rel = mcp_manifest.relative_to(ROOT)
    if not isinstance(icon, str) or not icon.startswith("./"):
        errors.append(f"{rel}: x-workbuddy.icon 须为包内相对路径（如 ./avatars/expert.svg）")
        return
    target = (pack_dir / icon[2:]).resolve()
    if pack_dir.resolve() not in target.parents:
        errors.append(f"{rel}: x-workbuddy.icon `{icon}` 逃出包目录")
    elif not target.is_file():
        errors.append(f"{rel}: x-workbuddy.icon `{icon}` 不存在")


def _check_pack(domain: str, method_names: set[str], errors: list[str]) -> None:
    """包内全部描述面（生成的 SKILL 入口 + 手写 plugin/.mcp.json）都不得泄漏 MCP 方法名。"""
    pack = packs.DOMAIN_PACKS[domain]
    pack_dir = PACKS_DIR / pack
    identity_doc = pack_dir / "SKILL.md"
    nested_skill = pack_dir / "skills" / domain / "SKILL.md"
    plugin = pack_dir / ".codebuddy-plugin" / "plugin.json"
    mcp_manifest = pack_dir / ".mcp.json"

    for path in (identity_doc, nested_skill, plugin, mcp_manifest):
        if not path.is_file():
            errors.append(f"missing expert-pack file: {path.relative_to(ROOT)}")
        _check_no_methods(path, method_names, errors)

    if not mcp_manifest.is_file():
        return
    try:
        mcp = json.loads(mcp_manifest.read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"{mcp_manifest.relative_to(ROOT)}: invalid JSON: {exc}")
        return
    descriptions: list[str] = []
    for server in mcp.get("mcpServers", {}).values():
        metadata = server.get("x-workbuddy", {}).get("description", {})
        descriptions.extend(
            value for key, value in metadata.items() if key in ("zh", "en") and isinstance(value, str)
        )
        _check_card_icon(pack_dir, server.get("x-workbuddy", {}).get("icon"), mcp_manifest, errors)
    for text in descriptions:
        for hit in _method_hits(text, method_names):
            errors.append(f"{mcp_manifest.relative_to(ROOT)}: connector description hard-codes `{hit}`")


def _check_avatar(domain: str, errors: list[str]) -> None:
    """头像矢量源只允许纯图形：带字形的源由 `scripts/build-workbuddy-avatars.sh` 转路径后再提交。"""
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

    _check_router(method_names, errors)
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
