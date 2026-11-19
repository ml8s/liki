#!/usr/bin/env python3
"""Build channel-specific distribution packages from the canonical Liki skill.

输入:
- `skills/liki/`：唯一专家身份和唯一方法论源；
- `distribution/variants/*.json`：free / pro 商业变体；
- `distribution/channels/*.json`：SkillHub / WorkBuddy 渠道规则；
- `distribution/workbuddy/liki-pro/`：WorkBuddy 静态渠道资产。

输出:
- `dist/skillhub/liki-free.zip`
- `dist/skillhub/liki-pro.zip`
- `dist/workbuddy/liki-free.zip`
- `dist/workbuddy/liki-pro.zip`

不生成第二份 `references/` 源码镜像。
"""

import argparse
import json
import sys
import zipfile
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
from version_projection import package_version  # noqa: E402

SKILL_DIR = ROOT / "skills" / "liki"
REFERENCES_DIR = SKILL_DIR / "references"
CANONICAL_SKILL = SKILL_DIR / "SKILL.md"

DISTRIBUTION_DIR = ROOT / "distribution"
VARIANTS_DIR = DISTRIBUTION_DIR / "variants"
CHANNELS_DIR = DISTRIBUTION_DIR / "channels"
MARKET_DIR = DISTRIBUTION_DIR / "metadata"
COMMON_MARKET = MARKET_DIR / "common.json"
WORKBUDDY_DIR = DISTRIBUTION_DIR / "workbuddy"
PRO_ASSETS_DIR = WORKBUDDY_DIR / "liki-pro"
PRO_PACKAGE_ID = "liki-pro"

WORKBUDDY_MCP_META = {
    "engine": {
        "displayName": {"zh": "Liki 排盘引擎", "en": "Liki Chart Engine"},
        "description": {
            "zh": "连接后可排八字、紫微、六爻、奇门、黄历与风水盘。",
            "en": "Connect to cast BaZi, Zi Wei, Liu Yao, Qi Men, almanac and Feng Shui charts.",
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


def workbuddy_mcp_servers(variant: str) -> dict:
    servers = load_variant(variant)["mcpServers"]
    rendered: dict = {}
    for name, config in servers.items():
        item = dict(config)
        if item.get("type") == "streamableHttp":
            item["type"] = "http"
        item["x-workbuddy"] = WORKBUDDY_MCP_META[name]
        rendered[name] = item
    return rendered


def load_json(path: Path, label: str) -> dict:
    if not path.is_file():
        raise SystemExit(f"{label} missing: {path.relative_to(ROOT)}")
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        raise SystemExit(f"{label} invalid JSON: {path.relative_to(ROOT)}: {exc}") from exc
    if not isinstance(value, dict):
        raise SystemExit(f"{label} must be a JSON object: {path.relative_to(ROOT)}")
    return value


def load_channel(channel: str, variant: str) -> tuple[dict, dict]:
    channel_config = load_json(CHANNELS_DIR / f"{channel}.json", f"channel {channel}")
    packages = channel_config.get("packages")
    if not isinstance(packages, dict) or variant not in packages:
        raise SystemExit(f"channel {channel} does not define variant {variant}")
    package = packages[variant]
    if not isinstance(package, dict):
        raise SystemExit(f"channel {channel} package {variant} must be an object")
    return channel_config, package


def load_market(channel: str, variant: str, assets_dir: Path | None = None) -> dict:
    common = load_json(COMMON_MARKET, "common market metadata")
    if channel == "skillhub":
        data = load_json(MARKET_DIR / "skillhub.json", "skillhub market metadata")
        market = data.get(variant)
    elif channel == "workbuddy":
        if assets_dir is None:
            raise SystemExit("workbuddy market metadata requires assets_dir")
        market = load_json(assets_dir / "skill-meta.json", f"WorkBuddy {variant} market metadata")
    else:
        raise SystemExit(f"unsupported channel: {channel}")
    if not isinstance(market, dict):
        raise SystemExit(f"market metadata for {channel}/{variant} must be an object")

    source_metadata, _body = split_frontmatter(CANONICAL_SKILL)
    source_meta = source_metadata.get("metadata", {})
    if not isinstance(source_meta, dict):
        raise SystemExit("canonical identity metadata must be an object")
    merged = dict(common)
    merged.update(market)
    merged["author"] = source_meta.get("author")
    merged["license"] = source_metadata.get("license")
    display_name = merged.get("display_name")
    display_name_en = merged.get("display_name_en")
    for key, value in (
        ("display_name", display_name),
        ("display_name_en", display_name_en),
    ):
        if not isinstance(value, str) or not value.strip():
            raise SystemExit(f"market metadata {channel}/{variant} missing {key}")
    merged["description_zh"] = str(merged.get("description_zh", "")).replace(
        "{display_name}", display_name
    )
    merged["description_en"] = str(merged.get("description_en", "")).replace(
        "{display_name_en}", display_name_en
    )
    merged.setdefault("profession_zh", display_name)
    merged.setdefault("profession_en", display_name_en)
    if channel == "skillhub":
        if not isinstance(merged.get("slug"), str) or not merged["slug"].strip():
            raise SystemExit(f"market metadata {channel}/{variant} missing slug")
        for key in ("description_zh", "description_en"):
            if len(str(merged.get(key, ""))) > 200:
                raise SystemExit(f"SkillHub/{variant} {key} exceeds 200 chars")
    if channel == "workbuddy":
        for key in ("agent_display_name_zh", "agent_display_name_en"):
            value = merged.get(key)
            if not isinstance(value, str) or not value.strip():
                raise SystemExit(f"market metadata {channel}/{variant} missing {key}")
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
        value = merged.get(key)
        if not isinstance(value, str) or not value.strip():
            raise SystemExit(f"market metadata {channel}/{variant} missing {key}")
    examples: dict[str, list] = {}
    for key in ("examples_zh", "examples_en"):
        value = merged.get(key)
        if not isinstance(value, list) or not value:
            raise SystemExit(f"market metadata {channel}/{variant} missing {key}")
        if not all(isinstance(item, str) and item.strip() for item in value):
            raise SystemExit(f"market metadata {channel}/{variant} has invalid {key}")
        examples[key] = value
    if len(examples["examples_zh"]) != len(examples["examples_en"]):
        raise SystemExit(f"market metadata {channel}/{variant} example locales differ")
    if len(examples["examples_zh"]) > 3:
        raise SystemExit(f"market metadata {channel}/{variant} exceeds 3 examples")
    return merged


def load_variant(variant: str) -> dict:
    config = load_json(VARIANTS_DIR / f"{variant}.json", f"variant {variant}")
    for key in ("sku", "top", "mcpServers", "payment"):
        if key not in config:
            raise SystemExit(f"variant {variant} missing {key}")
    return config


def split_frontmatter(path: Path) -> tuple[dict, str]:
    text = CANONICAL_SKILL.read_text(encoding="utf-8")
    if not text.startswith("---\n"):
        raise SystemExit(f"canonical identity missing frontmatter: {CANONICAL_SKILL}")
    end = text.find("\n---\n")
    if end < 0:
        raise SystemExit(f"canonical identity has unterminated frontmatter: {CANONICAL_SKILL}")
    frontmatter = yaml.safe_load(text[4:end])
    if not isinstance(frontmatter, dict):
        raise SystemExit(f"canonical identity frontmatter must be an object: {CANONICAL_SKILL}")
    body = text[end + len("\n---\n"):]
    return frontmatter, body[1:] if body.startswith("\n") else body


def canonical_identity(*, pro: bool = False) -> tuple[dict, str]:
    metadata, body = split_frontmatter(CANONICAL_SKILL)
    channel_meta = metadata.get("metadata", {})
    if not isinstance(channel_meta, dict):
        raise SystemExit("canonical identity metadata must be an object")

    if pro:
        overlay = PRO_ASSETS_DIR / "overlays" / "payment.md"
        if not overlay.is_file():
            raise SystemExit(f"pro payment overlay missing: {overlay}")
        body += "\n" + overlay.read_text(encoding="utf-8").rstrip() + "\n"

    return metadata, body.rstrip() + "\n"


def methodology_index() -> list[str]:
    cards = sorted(REFERENCES_DIR.glob("*.md"))
    titles = []
    for card in cards:
        title = card.stem
        for line in card.read_text(encoding="utf-8").splitlines():
            if line.startswith("# "):
                title = line[2:].strip()
                break
        titles.append(f"- `references/{card.name}` {title}")
    return titles


def _frontmatter_block(metadata: dict) -> str:
    dumped = yaml.dump(
        metadata,
        allow_unicode=True,
        sort_keys=False,
        width=4096,
        default_flow_style=False,
    )
    return "---\n" + dumped + "---\n\n"


def _render_identity(body: str, market: dict, *, heading: str | None = None) -> str:
    persona = market["agent_name_zh"]
    persona_en = market["agent_name_en"]
    role = market["profession_zh"]
    heading = heading or persona
    body = body.replace("# Liki 命理师", f"# {heading}", 1)
    body = body.replace(
        "你是 Liki 命理师，基于确定性引擎和可回溯规则辅助命理分析。",
        f"你是{persona}（中文）/ {persona_en}（英文），{role}。你基于确定性引擎和可回溯规则辅助命理分析。",
        1,
    )
    return body


def skill_doc(
    variant: str,
    skill_name: str,
    market: dict,
    *,
    channel: str,
    include_index: bool = False,
) -> str:
    _metadata, body = canonical_identity(pro=(variant == "pro"))
    version = package_version()
    if channel == "skillhub":
        doc_metadata = {
            "name": skill_name,
            "slug": market["slug"],
            "display_name": market["display_name"],
            "display_name_en": market["display_name_en"],
            "displayName": market["display_name"],
            "description": market["description_zh"],
            "description_zh": market["description_zh"],
            "description_en": market["description_en"],
            "summary": market["description_zh"],
            "version": version,
            "author": market["author"],
            "license": market["license"],
            "examples_zh": market["examples_zh"],
            "examples_en": market["examples_en"],
        }
    elif channel == "workbuddy":
        doc_metadata = {
            "name": skill_name,
            "display_name": market["display_name"],
            "display_name_en": market["display_name_en"],
            "description": market["description_zh"],
            "description_zh": market["description_zh"],
            "description_en": market["description_en"],
            "examples_zh": market["examples_zh"],
            "examples_en": market["examples_en"],
            "version": version,
            "author": market["author"],
            "license": market["license"],
            "agent_created": True,
        }
    else:
        raise SystemExit(f"unsupported channel: {channel}")
    rendered = _frontmatter_block(doc_metadata) + _render_identity(body, market)
    if include_index:
        index = methodology_index()
        rendered += f"\n## 方法论（{len(index)} 卡）\n\n" + "\n".join(index) + "\n"
    return rendered


def agent_doc(variant: str, package_id: str, assets_dir: Path) -> str:
    market = load_market("workbuddy", variant, assets_dir)
    _metadata, body = canonical_identity(pro=(variant == "pro"))
    max_turns = load_json(
        assets_dir / "channel.json",
        f"WorkBuddy {variant} channel config",
    )["maxTurns"]
    display_name = {
        "en": market["agent_display_name_en"],
        "zh": market["agent_display_name_zh"],
    }
    profession = {
        "en": market["profession_en"],
        "zh": market["profession_zh"],
    }
    body = _render_identity(body, market, heading=display_name["zh"])
    doc_metadata = {
        "name": package_id,
        "description": market["description_en"],
        "displayName": display_name,
        "profession": profession,
        "maxTurns": max_turns,
    }
    return _frontmatter_block(doc_metadata) + body


def plugin_doc(variant: str, package_id: str, assets_dir: Path) -> dict:
    market = load_market("workbuddy", variant, assets_dir)
    version = package_version()
    plugin = load_json(
        assets_dir / ".codebuddy-plugin" / "plugin.json",
        f"WorkBuddy {variant} plugin template",
    )
    keywords = market.get("keywords") or []
    skill_name = SKILL_DIR.name
    examples_zh = market["examples_zh"]
    examples_en = market["examples_en"]
    if len(examples_zh) != len(examples_en):
        raise SystemExit(f"market metadata WorkBuddy/{variant} example locales differ")
    quick = [{"zh": zh, "en": en} for zh, en in zip(examples_zh, examples_en)]
    if len(quick) != 3:
        raise SystemExit(f"WorkBuddy/{variant} requires exactly 3 quickPrompts")
    tags = plugin.get("tags")
    if not isinstance(tags, list) or len(tags) != 3:
        raise SystemExit(f"WorkBuddy/{variant} requires exactly 3 tags")
    display_description = plugin.get("displayDescription", {})
    if not isinstance(display_description, dict):
        raise SystemExit(f"WorkBuddy/{variant} displayDescription must be an object")
    description_zh = display_description.get("zh") or ""
    description_size = len([char for char in description_zh if not char.isspace()])
    if not 40 <= description_size <= 50:
        raise SystemExit(
            f"WorkBuddy/{variant} displayDescription.zh must be 40-50 chars "
            f"(got {description_size})"
        )
    plugin.update(
        {
            "name": package_id,
            "version": version,
            "description": market["description_en"],
            "agents": [f"./agents/{package_id}.md"],
            "agentName": package_id,
            "keywords": keywords,
            "skills": [f"./skills/{skill_name}"],
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


def _ignored_source(relative: str) -> bool:
    parts = relative.split("/")
    ignored_parts = {"__pycache__", ".pytest_cache", ".ruff_cache", ".venv", ".mypy_cache"}
    return (
        relative.endswith((".pyc", ".pyo", ".DS_Store", ".egg-info"))
        or any(part in ignored_parts for part in parts)
    )


def canonical_files(top: str, variant: str) -> dict[str, bytes]:
    files: dict[str, bytes] = {}
    for path in sorted(p for p in SKILL_DIR.rglob("*") if p.is_file()):
        relative = path.relative_to(SKILL_DIR).as_posix()
        if _ignored_source(relative):
            continue
        if relative == ".mcp.json":
            files[f"{top}/.mcp.json"] = json.dumps(
                {"mcpServers": load_variant(variant)["mcpServers"]},
                ensure_ascii=False,
                indent=2,
            ).encode("utf-8") + b"\n"
        elif relative == "SKILL.md":
            market = load_market("skillhub", variant)
            files[f"{top}/SKILL.md"] = skill_doc(
                variant,
                market["slug"],
                market,
                channel="skillhub",
            ).encode("utf-8")
        else:
            files[f"{top}/{relative}"] = path.read_bytes()
    return files


def static_expert_files(top: str, assets_dir: Path) -> dict[str, bytes]:
    files: dict[str, bytes] = {}
    include = {
        "README.md",
        ".codebuddy-plugin/plugin.json",
        "avatars/expert.png",
        "avatars/expert.svg",
    }
    for relative in sorted(include):
        path = assets_dir / relative
        if not path.is_file():
            raise SystemExit(f"WorkBuddy expert asset missing: {path}")
        files[f"{top}/{relative}"] = path.read_bytes()
    return files


def package_manifest(channel: str, variant: str) -> dict[str, bytes]:
    _channel_config, package = load_channel(channel, variant)
    top = package["top"]

    if channel == "workbuddy" and package.get("expert"):
        package_id = package["top"]
        assets_dir = WORKBUDDY_DIR / package.get("source", package_id)
        files: dict[str, bytes] = {}
        files.update(static_expert_files(top, assets_dir))
        files[f"{top}/.mcp.json"] = json.dumps(
            {"mcpServers": workbuddy_mcp_servers(variant)},
            ensure_ascii=False,
            indent=2,
        ).encode("utf-8") + b"\n"
        agent_path = f"{package_id}.md"
        files[f"{top}/agents/{agent_path}"] = agent_doc(
            variant,
            package_id,
            assets_dir,
        ).encode("utf-8")
        market = load_market("workbuddy", variant, assets_dir)
        files[f"{top}/skills/liki/SKILL.md"] = skill_doc(
            variant,
            "liki",
            market,
            channel="workbuddy",
            include_index=True,
        ).encode("utf-8")
        plugin = plugin_doc(variant, package_id, assets_dir)
        files[f"{top}/.codebuddy-plugin/plugin.json"] = json.dumps(
            plugin,
            ensure_ascii=False,
            indent=2,
        ).encode("utf-8") + b"\n"
        for path in sorted(REFERENCES_DIR.glob("*.md")):
            files[f"{top}/skills/liki/references/{path.name}"] = path.read_bytes()
        files[f"{top}/skills/liki/feedback.schema.json"] = (
            SKILL_DIR / "feedback.schema.json"
        ).read_bytes()
        files[f"{top}/skills/liki/scripts/feedback.py"] = (
            SKILL_DIR / "scripts" / "feedback.py"
        ).read_bytes()
    elif channel == "workbuddy" or channel == "skillhub":
        files = canonical_files(top, variant)
    else:
        raise SystemExit(f"unsupported channel: {channel}")

    artifact = package.get("artifact", f"{top}.zip")
    files["__artifact__"] = artifact.encode("utf-8")
    return files


def _display_path(path: Path) -> Path:
    try:
        return path.relative_to(ROOT)
    except ValueError:
        return path


def artifact_path(channel: str, variant: str) -> Path:
    _channel_config, package = load_channel(channel, variant)
    out_dir = ROOT / "dist" / channel
    out_dir.mkdir(parents=True, exist_ok=True)
    return out_dir / package.get("artifact", f"{package['top']}.zip")


def build(channel: str, variant: str, out: Path | None = None) -> Path:
    manifest = package_manifest(channel, variant)
    payload = manifest.pop("__artifact__").decode("utf-8")
    if out:
        target = Path(out)
    else:
        artifact = artifact_path(channel, variant)
        target = artifact.parent / payload
    target.parent.mkdir(parents=True, exist_ok=True)

    with zipfile.ZipFile(target, "w", compression=zipfile.ZIP_DEFLATED) as archive:
        for name, data in sorted(manifest.items()):
            info = zipfile.ZipInfo(name, date_time=(1980, 1, 1, 0, 0, 0))
            info.compress_type = zipfile.ZIP_DEFLATED
            info.external_attr = 0o644 << 16
            archive.writestr(info, data)
    try:
        shown = target.relative_to(ROOT)
    except ValueError:
        shown = target
    print(f"[build-channel-package] ✓ {shown} ({target.stat().st_size} bytes)")
    return target


def check(channel: str, variant: str, out: Path | None = None) -> int:
    manifest = package_manifest(channel, variant)
    artifact = out or artifact_path(channel, variant)
    payload = manifest.pop("__artifact__").decode("utf-8")
    target = Path(out) if out else artifact.parent / payload
    if not target.is_file():
        print(f"[build-channel-package] ✗ artifact missing: {_display_path(target)}")
        return 1

    top = load_channel(channel, variant)[1]["top"]
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

    expected = manifest
    for name in sorted(set(expected) - set(actual)):
        errors.append(f"缺失: {name}")
    for name in sorted(set(actual) - set(expected)):
        errors.append(f"多余: {name}")
    for name in sorted(set(expected) & set(actual)):
        if expected[name] != actual[name]:
            errors.append(f"内容漂移: {name}")

    if errors:
        print(f"[build-channel-package] ✗ {_display_path(target)} 校验失败：")
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1
    print(f"[build-channel-package] ✓ {_display_path(target)} 校验通过")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description="Build channel-specific Liki packages")
    parser.add_argument("--channel", required=True, choices=("skillhub", "workbuddy"))
    parser.add_argument("--variant", required=True, choices=("free", "pro"))
    parser.add_argument("--out", help="output zip path; defaults to the channel artifact path")
    parser.add_argument("--check", action="store_true", help="verify an existing package")
    args = parser.parse_args()
    if args.check:
        return check(args.channel, args.variant, args.out)
    build(args.channel, args.variant, args.out)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
