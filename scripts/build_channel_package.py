#!/usr/bin/env python3
"""Build WorkBuddy and SkillHub packages from one distribution catalog."""

from __future__ import annotations

import argparse
import json
import sys
import zipfile
from pathlib import Path
from typing import Any

import yaml

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
from version_projection import package_version  # noqa: E402

CATALOG_PATH = ROOT / "distribution" / "catalog.json"
SKILL_DIR = ROOT / "skills" / "liki"
REFERENCES_DIR = SKILL_DIR / "references"
CANONICAL_SKILL = SKILL_DIR / "SKILL.md"
WORKBUDDY_DIR = ROOT / "distribution" / "workbuddy"

WORKBUDDY_MCP_META = {
    "engine": {
        "displayName": {"zh": "Liki 排盘引擎", "en": "Liki Chart Engine"},
        "description": {
            "zh": "连接后可排八字、紫微、六爻、奇门、黄历与风水盘。",
            "en": "Connect to cast BaZi, Zi Wei Dou Shu, Liu Yao, Qi Men, Feng Shui and almanac charts.",
        },
        "auth": {"type": "none"},
    },
    "counsel": {
        "displayName": {"zh": "Liki 断语引擎", "en": "Liki Reading Engine"},
        "description": {
            "zh": "连接后可按规则命中带出处的断语、因子与应期。",
            "en": "Connect to match sourced assertions, factors and timing.",
        },
        "auth": {"type": "none"},
    },
    "liki": {
        "displayName": {"zh": "Liki 专业版", "en": "Liki Pro"},
        "description": {
            "zh": "连接后可使用专业版排盘、断语与依据回溯。",
            "en": "Connect for pro charting, sourced readings and traceable evidence.",
        },
        "auth": {"type": "none"},
    },
}


def _load_json(path: Path, label: str) -> dict[str, Any]:
    if not path.is_file():
        raise SystemExit(f"{label} missing: {path.relative_to(ROOT)}")
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"{label} invalid JSON: {path.relative_to(ROOT)}: {exc}") from exc
    if not isinstance(value, dict):
        raise SystemExit(f"{label} must be an object: {path.relative_to(ROOT)}")
    return value


def load_catalog() -> dict[str, Any]:
    return _load_json(CATALOG_PATH, "distribution catalog")


def load_variant(variant: str) -> dict[str, Any]:
    try:
        value = load_catalog()["variants"][variant]
    except KeyError as exc:
        raise SystemExit(f"unknown package variant: {variant}") from exc
    for key in ("sku", "top", "mcpServers", "payment"):
        if key not in value:
            raise SystemExit(f"variant {variant} missing {key}")
    return value


def load_package(channel: str, variant: str) -> dict[str, Any]:
    try:
        package = load_catalog()["packages"][channel][variant]
    except KeyError as exc:
        raise SystemExit(f"unknown channel package: {channel}/{variant}") from exc
    if package.get("variant", variant) != variant:
        raise SystemExit(f"channel package {channel}/{variant} points to another variant")
    return package


def split_frontmatter() -> tuple[dict[str, Any], str]:
    text = CANONICAL_SKILL.read_text(encoding="utf-8")
    if not text.startswith("---\n"):
        raise SystemExit(f"canonical identity missing frontmatter: {CANONICAL_SKILL}")
    end = text.find("\n---\n")
    if end < 0:
        raise SystemExit(f"canonical identity has unterminated frontmatter: {CANONICAL_SKILL}")
    frontmatter = yaml.safe_load(text[4:end])
    if not isinstance(frontmatter, dict):
        raise SystemExit("canonical identity frontmatter must be an object")
    body = text[end + 5 :]
    return frontmatter, body[1:] if body.startswith("\n") else body


def canonical_identity(*, pro: bool = False) -> tuple[dict[str, Any], str]:
    metadata, body = split_frontmatter()
    if pro:
        overlay = WORKBUDDY_DIR / "liki-pro" / "overlays" / "payment.md"
        if not overlay.is_file():
            raise SystemExit(f"pro payment overlay missing: {overlay}")
        body += "\n" + overlay.read_text(encoding="utf-8").rstrip() + "\n"
    return metadata, body.rstrip() + "\n"


def load_market(channel: str, variant: str, assets_dir: Path | None = None) -> dict[str, Any]:
    catalog = load_catalog()
    package = load_package(channel, variant)
    common = catalog["commonMarket"]
    if channel == "skillhub":
        overrides = package["market"]
    elif channel == "workbuddy":
        assets_dir = assets_dir or WORKBUDDY_DIR / package["source"]
        overrides = _load_json(
            assets_dir / "skill-meta.json", f"WorkBuddy {variant} market metadata"
        )
    else:
        raise SystemExit(f"unsupported channel: {channel}")
    if not isinstance(overrides, dict):
        raise SystemExit(f"market metadata for {channel}/{variant} must be an object")

    frontmatter, _body = split_frontmatter()
    source_metadata = frontmatter.get("metadata", {})
    if not isinstance(source_metadata, dict):
        raise SystemExit("canonical identity metadata must be an object")
    market = dict(common)
    market.update(overrides)
    market["author"] = source_metadata.get("author")
    market["license"] = frontmatter.get("license")

    for key in ("display_name", "display_name_en"):
        if not isinstance(market.get(key), str) or not market[key].strip():
            raise SystemExit(f"market metadata {channel}/{variant} missing {key}")
    market["description_zh"] = market["description_zh"].replace("{display_name}", market["display_name"])
    market["description_en"] = market["description_en"].replace("{display_name_en}", market["display_name_en"])
    market.setdefault("profession_zh", market["display_name"])
    market.setdefault("profession_en", market["display_name_en"])
    for key in (
        "description_zh",
        "description_en",
        "profession_zh",
        "profession_en",
        "agent_name_zh",
        "agent_name_en",
        "author",
        "license",
    ):
        if not isinstance(market.get(key), str) or not market[key].strip():
            raise SystemExit(f"market metadata {channel}/{variant} missing {key}")
    examples: dict[str, list[Any]] = {}
    for key in ("examples_zh", "examples_en"):
        value = market.get(key)
        if not isinstance(value, list) or not value or not all(isinstance(item, str) and item.strip() for item in value):
            raise SystemExit(f"market metadata {channel}/{variant} has invalid {key}")
        examples[key] = value
    if len(examples["examples_zh"]) != len(examples["examples_en"]) or len(examples["examples_zh"]) > 3:
        raise SystemExit(f"market metadata {channel}/{variant} examples are invalid")
    if channel == "skillhub":
        if not market.get("slug") or any(len(market[key]) > 200 for key in ("description_zh", "description_en")):
            raise SystemExit(f"SkillHub metadata for {variant} is invalid")
    if channel == "workbuddy" and any(not market.get(key) for key in ("agent_display_name_zh", "agent_display_name_en")):
        raise SystemExit(f"WorkBuddy metadata for {variant} is invalid")
    return market


def _frontmatter_block(metadata: dict[str, Any]) -> str:
    dumped = yaml.dump(metadata, allow_unicode=True, sort_keys=False, width=4096, default_flow_style=False)
    return f"---\n{dumped}---\n\n"


def methodology_index() -> list[str]:
    titles: list[str] = []
    for card in sorted(REFERENCES_DIR.glob("*.md")):
        title = card.stem
        for line in card.read_text(encoding="utf-8").splitlines():
            if line.startswith("# "):
                title = line[2:].strip()
                break
        titles.append(f"- `references/{card.name}` {title}")
    return titles


def _render_identity(body: str, market: dict[str, Any], heading: str | None = None) -> str:
    body = body.replace("# Liki 命理师", f"# {heading or market['agent_name_zh']}", 1)
    return body.replace(
        "你是 Liki 命理师，基于确定性引擎和可回溯规则辅助命理分析。",
        (
            f"你是{market['agent_name_zh']}（中文）/ {market['agent_name_en']}（英文），"
            f"{market['profession_zh']}。你基于确定性引擎和可回溯规则辅助命理分析。"
        ),
        1,
    )


def skill_doc(
    variant: str,
    skill_name: str,
    market: dict[str, Any],
    *,
    channel: str,
    include_index: bool = False,
) -> str:
    _frontmatter, body = canonical_identity(pro=variant == "pro")
    doc: dict[str, Any] = {
        "name": skill_name,
        "display_name": market["display_name"],
        "display_name_en": market["display_name_en"],
        "description": market["description_zh"],
        "description_zh": market["description_zh"],
        "description_en": market["description_en"],
        "examples_zh": market["examples_zh"],
        "examples_en": market["examples_en"],
        "version": package_version(),
        "author": market["author"],
        "license": market["license"],
    }
    if channel == "skillhub":
        doc.update(
            {
                "slug": market["slug"],
                "displayName": market["display_name"],
                "summary": market["description_zh"],
            }
        )
    elif channel == "workbuddy":
        doc["agent_created"] = True
    else:
        raise SystemExit(f"unsupported channel: {channel}")
    rendered = _frontmatter_block(doc) + _render_identity(body, market)
    if include_index:
        index = methodology_index()
        rendered += f"\n## 方法论（{len(index)} 卡）\n\n" + "\n".join(index) + "\n"
    return rendered


def agent_doc(variant: str, package_id: str, assets_dir: Path) -> str:
    market = load_market("workbuddy", variant, assets_dir)
    _frontmatter, body = canonical_identity(pro=variant == "pro")
    channel = _load_json(assets_dir / "channel.json", f"WorkBuddy {variant} channel config")
    metadata = {
        "name": package_id,
        "description": market["description_en"],
        "displayName": {
            "en": market["agent_display_name_en"],
            "zh": market["agent_display_name_zh"],
        },
        "profession": {
            "en": market["profession_en"],
            "zh": market["profession_zh"],
        },
        "maxTurns": channel["maxTurns"],
    }
    heading = market["agent_display_name_zh"]
    return _frontmatter_block(metadata) + _render_identity(body, market, heading=heading)


def plugin_doc(variant: str, package_id: str, assets_dir: Path) -> dict[str, Any]:
    market = load_market("workbuddy", variant, assets_dir)
    plugin = _load_json(
        assets_dir / ".codebuddy-plugin" / "plugin.json",
        f"WorkBuddy {variant} plugin template",
    )
    examples = list(zip(market["examples_zh"], market["examples_en"]))
    if len(examples) != 3:
        raise SystemExit(f"WorkBuddy/{variant} requires exactly 3 quickPrompts")
    tags = plugin.get("tags")
    if not isinstance(tags, list) or len(tags) != 3:
        raise SystemExit(f"WorkBuddy/{variant} requires exactly 3 tags")
    description_zh = plugin.get("displayDescription", {}).get("zh", "")
    width = len([char for char in description_zh if not char.isspace()])
    if not 40 <= width <= 50:
        raise SystemExit(f"WorkBuddy/{variant} displayDescription.zh must be 40-50 chars (got {width})")
    quick = [{"zh": zh, "en": en} for zh, en in examples]
    plugin.update(
        {
            "name": package_id,
            "version": package_version(),
            "description": market["description_en"],
            "agents": [f"./agents/{package_id}.md"],
            "agentName": package_id,
            "keywords": market.get("keywords", []),
            "skills": [f"./skills/{SKILL_DIR.name}"],
            "plugin": package_id,
            "displayName": {
                "en": market["agent_display_name_en"],
                "zh": market["agent_display_name_zh"],
            },
            "profession": {
                "en": market["profession_en"],
                "zh": market["profession_zh"],
            },
            "quickPrompts": quick,
            "defaultInitPrompt": quick[0],
        }
    )
    return plugin


def canonical_files(top: str, variant: str) -> dict[str, bytes]:
    files: dict[str, bytes] = {}
    for path in sorted(item for item in SKILL_DIR.rglob("*") if item.is_file()):
        relative = path.relative_to(SKILL_DIR).as_posix()
        if "__pycache__" in relative.split("/") or relative.endswith((".pyc", ".pyo", ".DS_Store")):
            continue
        if relative == ".mcp.json":
            servers = load_variant(variant)["mcpServers"]
            files[f"{top}/.mcp.json"] = json.dumps(
                {"mcpServers": servers}, ensure_ascii=False, indent=2
            ).encode() + b"\n"
        elif relative == "SKILL.md":
            market = load_market("skillhub", variant)
            files[f"{top}/SKILL.md"] = skill_doc(
                variant, market["slug"], market, channel="skillhub"
            ).encode()
        else:
            files[f"{top}/{relative}"] = path.read_bytes()
    return files


def package_manifest(channel: str, variant: str) -> dict[str, bytes]:
    package = load_package(channel, variant)
    top = package["top"]
    if channel == "skillhub":
        files = canonical_files(top, variant)
    elif channel == "workbuddy":
        assets_dir = WORKBUDDY_DIR / package["source"]
        files = {
            f"{top}/{relative}": (assets_dir / relative).read_bytes()
            for relative in (
                "README.md",
                ".codebuddy-plugin/plugin.json",
                "avatars/expert.png",
                "avatars/expert.svg",
            )
        }
        rendered = {}
        for name, config in load_variant(variant)["mcpServers"].items():
            if config.get("type") == "streamableHttp":
                config = {**config, "type": "http"}
            rendered[name] = {**config, "x-workbuddy": WORKBUDDY_MCP_META[name]}
        files[f"{top}/.mcp.json"] = json.dumps(
            {"mcpServers": rendered}, ensure_ascii=False, indent=2
        ).encode() + b"\n"
        files[f"{top}/agents/{package['top']}.md"] = agent_doc(
            variant, package["top"], assets_dir
        ).encode()
        market = load_market("workbuddy", variant, assets_dir)
        files[f"{top}/skills/liki/SKILL.md"] = skill_doc(
            variant, "liki", market, channel="workbuddy", include_index=True
        ).encode()
        files[f"{top}/.codebuddy-plugin/plugin.json"] = json.dumps(
            plugin_doc(variant, package["top"], assets_dir), ensure_ascii=False, indent=2
        ).encode() + b"\n"
        for path in sorted(REFERENCES_DIR.glob("*.md")):
            files[f"{top}/skills/liki/references/{path.name}"] = path.read_bytes()
        files[f"{top}/skills/liki/feedback.schema.json"] = (SKILL_DIR / "feedback.schema.json").read_bytes()
        files[f"{top}/skills/liki/scripts/feedback.py"] = (SKILL_DIR / "scripts" / "feedback.py").read_bytes()
    else:
        raise SystemExit(f"unsupported channel: {channel}")
    files["__artifact__"] = package["artifact"].encode()
    return files


def artifact_path(channel: str, variant: str) -> Path:
    out = ROOT / "dist" / channel
    out.mkdir(parents=True, exist_ok=True)
    return out / load_package(channel, variant)["artifact"]


def _display_path(target: Path) -> Path:
    try:
        return target.relative_to(ROOT)
    except ValueError:
        return target


def _write_zip(target: Path, files: dict[str, bytes]) -> None:
    target.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(target, "w", compression=zipfile.ZIP_DEFLATED) as archive:
        for name, data in sorted(files.items()):
            info = zipfile.ZipInfo(name, date_time=(1980, 1, 1, 0, 0, 0))
            info.compress_type = zipfile.ZIP_DEFLATED
            info.external_attr = 0o644 << 16
            archive.writestr(info, data)


def build(channel: str, variant: str, out: Path | None = None) -> Path:
    files = package_manifest(channel, variant)
    artifact = files.pop("__artifact__").decode()
    target = out or artifact_path(channel, variant).parent / artifact
    _write_zip(target, files)
    print(f"[build-channel-package] ✓ {_display_path(target)} ({target.stat().st_size} bytes)")
    return target


def check(channel: str, variant: str, out: Path | None = None) -> int:
    expected = package_manifest(channel, variant)
    artifact = expected.pop("__artifact__").decode()
    target = out or artifact_path(channel, variant).parent / artifact
    if not target.is_file():
        print(f"[build-channel-package] ✗ artifact missing: {target}")
        return 1
    top = load_package(channel, variant)["top"]
    actual: dict[str, bytes] = {}
    errors: list[str] = []
    with zipfile.ZipFile(target) as archive:
        for info in archive.infolist():
            if info.filename.endswith("/"):
                continue
            if not info.filename.startswith(f"{top}/"):
                errors.append(f"条目不在 {top}/ 下: {info.filename}")
                continue
            actual[info.filename] = archive.read(info.filename)
    for name in sorted(set(expected) - set(actual)):
        errors.append(f"缺失: {name}")
    for name in sorted(set(actual) - set(expected)):
        errors.append(f"多余: {name}")
    errors.extend(
        f"内容漂移: {name}"
        for name in sorted(set(expected) & set(actual))
        if expected[name] != actual[name]
    )
    if errors:
        print(f"[build-channel-package] ✗ {target} 校验失败：")
        print("\n".join(f"  - {error}" for error in errors), file=sys.stderr)
        return 1
    print(f"[build-channel-package] ✓ {_display_path(target)} 校验通过")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description="Build channel-specific Liki packages")
    parser.add_argument("--channel", required=True, choices=("skillhub", "workbuddy"))
    parser.add_argument("--variant", required=True, choices=("free", "pro"))
    parser.add_argument("--out", type=Path)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    if args.check:
        return check(args.channel, args.variant, args.out)
    build(args.channel, args.variant, args.out)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
