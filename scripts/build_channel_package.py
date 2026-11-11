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
import re
import sys
import zipfile
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
from version_projection import workbuddy_version  # noqa: E402

SKILL_DIR = ROOT / "skills" / "liki"
REFERENCES_DIR = SKILL_DIR / "references"
CANONICAL_SKILL = SKILL_DIR / "SKILL.md"
RUNTIME_VERSION_PATH = ROOT / "skills" / "liki" / "VERSION.txt"

DISTRIBUTION_DIR = ROOT / "distribution"
VARIANTS_DIR = DISTRIBUTION_DIR / "variants"
CHANNELS_DIR = DISTRIBUTION_DIR / "channels"
WORKBUDDY_DIR = DISTRIBUTION_DIR / "workbuddy"
PRO_ASSETS_DIR = WORKBUDDY_DIR / "liki-pro"
DOMAINS = ("liki",)
DOMAIN_PACKS = {"liki": "liki-pro"}
EXPERT_PACKS = tuple(DOMAIN_PACKS.values())
PACK_AUTHOR = "Liki"

FREE_CHANNEL_PATTERN = re.compile(
    r"<!-- channel:free -->.*?<!-- /channel:free -->\n?",
    re.S,
)
IDENTITY_LOCALES = ("en", "zh")


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


def identity_path(variant: str) -> Path:
    return CANONICAL_SKILL


def overlay_path(variant: str) -> Path:
    return PRO_ASSETS_DIR / "overlays" / "payment.md"


def load_channel(channel: str, variant: str) -> tuple[dict, dict]:
    channel_config = load_json(CHANNELS_DIR / f"{channel}.json", f"channel {channel}")
    packages = channel_config.get("packages")
    if not isinstance(packages, dict) or variant not in packages:
        raise SystemExit(f"channel {channel} does not define variant {variant}")
    package = packages[variant]
    if not isinstance(package, dict):
        raise SystemExit(f"channel {channel} package {variant} must be an object")
    return channel_config, package


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


def canonical_identity(variant: str, *, workbuddy: bool = False) -> tuple[dict, str, str]:
    metadata, body = split_frontmatter(CANONICAL_SKILL)
    channel_meta = metadata.get("metadata", {})
    if not isinstance(channel_meta, dict):
        raise SystemExit("canonical identity metadata must be an object")
    variant_config = load_variant(variant)
    sku = variant_config["sku"]

    for key in ("description_en", "display_name", "display_name_en"):
        value = channel_meta.get(key)
        if isinstance(value, str) and value.strip():
            metadata.setdefault(key, value)

    if variant == "pro":
        body = FREE_CHANNEL_PATTERN.sub("", body)
        overlay = PRO_ASSETS_DIR / "overlays" / "payment.md"
        if not overlay.is_file():
            raise SystemExit(f"pro payment overlay missing: {overlay}")
        body += "\n" + overlay.read_text(encoding="utf-8").rstrip() + "\n"

    if workbuddy:
        skill_name = "liki"
    else:
        skill_name = sku
    return metadata, body.rstrip() + "\n", skill_name


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


def skill_doc(variant: str, *, workbuddy: bool = False, include_index: bool = False) -> str:
    metadata, body, skill_name = canonical_identity(variant, workbuddy=workbuddy)
    version = workbuddy_version(RUNTIME_VERSION_PATH.read_text(encoding="utf-8").strip())
    doc_metadata = {
        "name": skill_name,
        "description": metadata["description"],
        "metadata": {
            "display_name": metadata["display_name"],
            "display_name_en": metadata["display_name_en"],
            "version": version,
            "author": PACK_AUTHOR,
        },
    }
    rendered = _frontmatter_block(doc_metadata) + body
    if include_index:
        index = methodology_index()
        rendered += f"\n## 方法论（{len(index)} 卡）\n\n" + "\n".join(index) + "\n"
    return rendered


def agent_doc() -> str:
    metadata, body, _skill_name = canonical_identity("pro", workbuddy=True)
    channel = load_json(PRO_ASSETS_DIR / "channel.json", "WorkBuddy pro channel config")
    doc_metadata = {
        "name": "liki-pro",
        "description": metadata["description_en"],
        "displayName": {
            "en": metadata["display_name_en"],
            "zh": metadata["display_name"],
        },
        "profession": channel["profession"],
        "maxTurns": channel["maxTurns"],
    }
    return _frontmatter_block(doc_metadata) + body


def plugin_doc(variant: str) -> dict:
    if variant != "pro":
        raise SystemExit("WorkBuddy plugin packages use the pro variant")
    metadata, _body, _skill_name = canonical_identity("pro", workbuddy=True)
    channel = load_json(PRO_ASSETS_DIR / "channel.json", "WorkBuddy pro channel config")
    version = workbuddy_version(RUNTIME_VERSION_PATH.read_text(encoding="utf-8").strip())
    plugin = load_json(PRO_ASSETS_DIR / ".codebuddy-plugin" / "plugin.json", "WorkBuddy plugin template")
    plugin.update(
        {
            "name": "liki-pro",
            "version": version,
            "description": metadata["description_en"],
            "agents": ["./agents/liki-pro.md"],
            "agentName": "liki-pro",
            "displayName": {
                "en": metadata["display_name_en"],
                "zh": metadata["display_name"],
            },
            "profession": channel["profession"],
            "keywords": channel["keywords"],
            "skills": [f"./skills/{name}" for name in channel["skills"]],
            "plugin": "liki-pro",
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
            files[f"{top}/SKILL.md"] = skill_doc(variant, workbuddy=False).encode("utf-8")
        else:
            files[f"{top}/{relative}"] = path.read_bytes()
    return files


def static_pro_files(top: str) -> dict[str, bytes]:
    files: dict[str, bytes] = {}
    include = {
        "README.md",
        ".codebuddy-plugin/plugin.json",
        "avatars/expert.png",
        "avatars/expert.svg",
    }
    for relative in sorted(include):
        path = PRO_ASSETS_DIR / relative
        if not path.is_file():
            raise SystemExit(f"WorkBuddy pro asset missing: {path}")
        files[f"{top}/{relative}"] = path.read_bytes()
    return files


def package_manifest(channel: str, variant: str) -> dict[str, bytes]:
    channel_config, package = load_channel(channel, variant)
    variant_config = load_variant(variant)
    top = package["top"]

    if channel == "workbuddy" and variant == "pro":
        files: dict[str, bytes] = {}
        files.update(static_pro_files(top))
        files[f"{top}/.mcp.json"] = json.dumps(
            {"mcpServers": variant_config["mcpServers"]},
            ensure_ascii=False,
            indent=2,
        ).encode("utf-8") + b"\n"
        files[f"{top}/agents/liki-pro.md"] = agent_doc().encode("utf-8")
        files[f"{top}/skills/liki/SKILL.md"] = skill_doc(
            variant,
            workbuddy=True,
            include_index=True,
        ).encode("utf-8")
        plugin = plugin_doc(variant)
        files[f"{top}/.codebuddy-plugin/plugin.json"] = json.dumps(
            plugin,
            ensure_ascii=False,
            indent=2,
        ).encode("utf-8") + b"\n"
        for path in sorted(REFERENCES_DIR.glob("*.md")):
            if path.name in {"FAQ.md", "ops-FEEDBACK.md"}:
                continue
            files[f"{top}/skills/liki/references/{path.name}"] = path.read_bytes()
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
