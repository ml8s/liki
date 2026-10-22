#!/usr/bin/env python3
"""Bump every Liki runtime/compatibility CalVer surface atomically."""

import json
import sys
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))
from version_projection import workbuddy_version  # noqa: E402  制品版本的唯一投影

VERSION_FILES = (
    ROOT / "skills/liki/VERSION.txt",
    ROOT / "engine/cmd/engine-mcp/VERSION",
    ROOT / "counsel/app/VERSION.txt",
)
PROFILE_FILES = tuple((ROOT / "profiles").glob("*.json"))
# Counsel tool contracts and projection tables carry no version:
# counsel/app/VERSION.txt is the runtime surface, and the MCP servers advertise
# it in the initialize result.
SKILL_FRONTMATTER = ROOT / "skills/liki/SKILL.md"
EXPERT_PACKS = ROOT / "expert-packs"
CONNECTOR_METAS = tuple(sorted((ROOT / "workbuddy").glob("*-connector/connector-meta.json")))


def current_versions() -> list[str]:
    return [
        path.read_text(encoding="utf-8").strip()
        for path in (*VERSION_FILES, *PROFILE_FILES)
    ]


def next_version() -> str:
    today = datetime.now(ZoneInfo("Asia/Shanghai")).strftime("%Y.%m.%d")
    serial = -1
    found = False
    for version in current_versions():
        if version.startswith(today + "."):
            found = True
            serial = max(serial, int(version.rsplit(".", 1)[1]))
    if not found:
        serial = 0
    else:
        serial += 1
    return f"{today}.{serial}"


def replace_frontmatter_version(path: Path, version: str) -> None:
    lines = path.read_text(encoding="utf-8").splitlines(keepends=True)
    if not lines or lines[0].rstrip("\r\n") != "---":
        raise SystemExit(f"missing frontmatter in {path.relative_to(ROOT)}")
    for index in range(1, len(lines)):
        if lines[index].rstrip("\r\n") == "---":
            raise SystemExit(
                f"frontmatter version missing in {path.relative_to(ROOT)}"
            )
        stripped = lines[index].lstrip()
        indent = lines[index][: len(lines[index]) - len(stripped)]
        if stripped.startswith("version:"):
            lines[index] = f"{indent}version: {version}\n"
            path.write_text("".join(lines), encoding="utf-8")
            return


def set_expert_pack_versions(version: str) -> None:
    for skill_path in sorted(EXPERT_PACKS.glob("*/SKILL.md")):
        replace_frontmatter_version(skill_path, version)
    for skill_path in sorted(EXPERT_PACKS.glob("*/skills/*/SKILL.md")):
        replace_frontmatter_version(skill_path, version)
    for plugin_path in sorted(EXPERT_PACKS.glob("*/.codebuddy-plugin/plugin.json")):
        document = json.loads(plugin_path.read_text(encoding="utf-8"))
        document["version"] = version
        plugin_path.write_text(
            json.dumps(document, ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )


def set_connector_meta_versions(version: str) -> None:
    """连接器元信息的 `version` 同样是制品面：与技能/专家同源，不手写。"""
    for meta_path in CONNECTOR_METAS:
        document = json.loads(meta_path.read_text(encoding="utf-8"))
        document["version"] = version
        meta_path.write_text(
            json.dumps(document, ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )


def set_pyproject(version: str) -> None:
    for relative in ("pyproject.toml", "counsel/pyproject.toml"):
        path = ROOT / relative
        lines = path.read_text(encoding="utf-8").splitlines()
        changed = False
        for index, line in enumerate(lines):
            if line.startswith("version ="):
                lines[index] = f'version = "{version}"'
                changed = True
                break
        if changed:
            path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def main() -> None:
    version = next_version()
    for path in VERSION_FILES:
        path.write_text(version + "\n", encoding="utf-8")
    for path in PROFILE_FILES:
        document = json.loads(path.read_text(encoding="utf-8"))
        document.setdefault("metadata", {})["version"] = version
        path.write_text(
            json.dumps(document, ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )
    # 平台按 semver 解析制品版本：技能/专家/连接器统一写投影值。
    artifact_version = workbuddy_version(version)
    replace_frontmatter_version(SKILL_FRONTMATTER, artifact_version)
    set_expert_pack_versions(artifact_version)
    set_connector_meta_versions(artifact_version)
    set_pyproject(version)
    print(f"✅ version → {version}（WorkBuddy 制品 {artifact_version}）")


if __name__ == "__main__":
    main()
