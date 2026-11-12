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


def _check_identity(method_names: set[str], errors: list[str]) -> dict:
    """唯一专家身份：契约（生成器）+ 边界语义（本脚本）。"""
    metadata, _body = packs.canonical_identity(pro=True)
    path = packs.CANONICAL_SKILL
    text = _check_sections(path, IDENTITY_SECTIONS, errors)
    for phrase in IDENTITY_PHRASES:
        if phrase not in text:
            errors.append(f"{path.relative_to(ROOT)}: missing boundary phrase «{phrase}»")
    _check_no_methods(path, method_names, errors)
    return metadata


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


def _check_pack(method_names: set[str], errors: list[str]) -> None:
    """包内全部描述面（生成的入口 + plugin.json）都不得泄漏 MCP 方法名。"""
    pack = packs.PRO_PACKAGE_ID
    pack_dir = PACKS_DIR / pack
    identity_doc = packs.CANONICAL_SKILL
    payment_overlay = packs.PRO_ASSETS_DIR / "overlays" / "payment.md"
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


def _check_avatar(errors: list[str]) -> None:
    """头像：权威品牌图来自 liki-web C3 资产；SVG 必须自包含，PNG 必须是 512×512。
    """
    pack_dir = packs.PRO_ASSETS_DIR
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
    with png.open("rb") as fh:
        header = fh.read(24)
    if header[:8] != b"\x89PNG\r\n\x1a\n":
        errors.append(f"{png.relative_to(ROOT)}: avatar is not PNG")
    elif int.from_bytes(header[16:20], "big") != 512 or int.from_bytes(header[20:24], "big") != 512:
        errors.append(f"{png.relative_to(ROOT)}: avatar must be 512x512 PNG")


def main() -> int:
    errors: list[str] = []
    method_names = _method_names()

    _check_identity(method_names, errors)
    _check_pack(method_names, errors)
    _check_avatar(errors)


if __name__ == "__main__":
    raise SystemExit(main())
