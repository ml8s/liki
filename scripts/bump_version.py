#!/usr/bin/env python3
"""Bump every Liki runtime/compatibility CalVer surface atomically."""
from __future__ import annotations

import json
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo

ROOT = Path(__file__).resolve().parents[1]

VERSION_FILES = (
    ROOT / "skills/liki/VERSION.txt",
    ROOT / "engine/cmd/engine-mcp/VERSION",
    ROOT / "engine/cmd/engine-rpc/VERSION",
    ROOT / "counsel/app/VERSION.txt",
)
PROFILE_FILES = tuple((ROOT / "profiles").glob("*.json"))
JSON_SURFACES = (
    (ROOT / "counsel/app/divination/tools/skill-tools.json", ("info", "version")),
    (ROOT / "counsel/app/naming/tools/skill-tools.json", ("info", "version")),
    (ROOT / "counsel/app/natal/tools/counsel-tools.json", ("version",)),
    (ROOT / "counsel/app/natal/tools/natal_projection_contract.json", ("version",)),
    (ROOT / "counsel/app/divination/tools/qimen_projection_contract.json", ("version",)),
)
SKILL_FRONTMATTER = ROOT / "skills/liki/SKILL.md"
EXPERT_PACKS = ROOT / "expert-packs"


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


def set_json(path: Path, keys: tuple[str, ...], version: str) -> None:
    document = json.loads(path.read_text(encoding="utf-8"))
    node = document
    for key in keys[:-1]:
        node = node.setdefault(key, {})
    node[keys[-1]] = version
    path.write_text(
        json.dumps(document, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )


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
    for path, keys in JSON_SURFACES:
        set_json(path, keys, version)
    replace_frontmatter_version(SKILL_FRONTMATTER, version)
    set_expert_pack_versions(version)
    set_pyproject(version)
    print(f"✅ version → {version}")


if __name__ == "__main__":
    main()
