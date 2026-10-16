#!/usr/bin/env python3
"""校验 dist/workbuddy/ 下的提交 zip 与仓库源码逐字节一致、且结构符合平台解析要求。

官网只收 zip，且 zip 顶层必须是提交单元目录本身。本脚本把「打包」这一步的可信度钉死：
每个包都与源码目录（按同一套排除规则）逐文件比对 sha256，任何缺失、多余、内容漂移、
顶层目录名不符、绝对路径或 `..` 条目都判失败。同时检查平台侧的目录约定：

* 连接器 `<source>-connector/`：`connector-meta.json` + `mcp.json`/`cli.json` + `icon.*`；
* 技能 `liki/`：`SKILL.md` + 子资源；
* 专家 `liki-<domain>/`：`.codebuddy-plugin/plugin.json`、`agents/<pack>.md`、`avatars/expert.png`、
  `README.md`，且不含包根身份手写源 `SKILL.md`（platform 侧入口是 agents/<pack>.md）。

由 `scripts/package-workbuddy.sh` 打包后调用；`dist/workbuddy/` 不存在（干净 checkout）时跳过。
"""
from __future__ import annotations

import json
import sys
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
OUT_DIR = ROOT / "dist" / "workbuddy"
EXCLUDED_NAMES = {".DS_Store"}
EXCLUDED_PREFIXES = ("__MACOSX/",)
EXCLUDED_PARTS = ("__pycache__", ".venv", ".pytest_cache", ".ruff_cache")
EXCLUDED_SUFFIXES = (".pyc", ".pyo")
ICON_NAMES = ("icon.svg", "icon.png", "icon.jpg")
CONFIG_NAMES = ("mcp.json", "cli.json")


def ignored(relative: str) -> bool:
    """打包装箱规则（与 `scripts/package-workbuddy.sh` 的 -x 一致）。"""
    if relative.startswith(EXCLUDED_PREFIXES):
        return True
    parts = relative.split("/")
    if Path(relative).name in EXCLUDED_NAMES or parts[0] == "__MACOSX":
        return True
    if any(part in EXCLUDED_PARTS for part in parts[:-1]):
        return True
    return relative.endswith(EXCLUDED_SUFFIXES)


def source_files(root: Path, skip: frozenset[str] = frozenset()) -> dict[str, bytes]:
    return {
        path.relative_to(root).as_posix(): path.read_bytes()
        for path in sorted(root.rglob("*"))
        if path.is_file()
        and path.relative_to(root).as_posix() not in skip
        and not ignored(path.relative_to(root).as_posix())
    }


def zip_files(path: Path, top: str, errors: list[str]) -> dict[str, bytes]:
    payload: dict[str, bytes] = {}
    with zipfile.ZipFile(path) as archive:
        for info in archive.infolist():
            name = info.filename
            if name.endswith("/"):
                continue
            if not name.startswith(f"{top}/") or ".." in name.split("/") or name.startswith("/"):
                errors.append(f"{path.name}: 条目 `{name}` 不在顶层目录 `{top}/` 内")
                continue
            if ignored(name) or ignored(name[len(top) + 1:]):
                errors.append(f"{path.name}: 条目 `{name}` 是不该入包的本地残渣")
                continue
            payload[name[len(top) + 1:]] = archive.read(name)
    return payload


def compare(label: str, expected: dict[str, bytes], actual: dict[str, bytes], errors: list[str]) -> None:
    if expected == actual:
        return
    for name in sorted(set(expected) - set(actual)):
        errors.append(f"{label}: 缺文件 `{name}`")
    for name in sorted(set(actual) - set(expected)):
        errors.append(f"{label}: 多余文件 `{name}`")
    for name in sorted(set(expected) & set(actual)):
        if expected[name] != actual[name]:
            errors.append(f"{label}: 内容漂移 `{name}`")


def check_connector(zip_path: Path, name: str, errors: list[str]) -> None:
    files = zip_files(zip_path, name, errors)
    source = ROOT / "workbuddy" / name
    compare(f"{zip_path.name}", source_files(source), files, errors)
    if "connector-meta.json" not in files:
        errors.append(f"{zip_path.name}: 缺 connector-meta.json")
    if not any(config in files for config in CONFIG_NAMES):
        errors.append(f"{zip_path.name}: 缺 mcp.json / cli.json")
    if not any(icon in files for icon in ICON_NAMES):
        errors.append(f"{zip_path.name}: 缺图标（{'/'.join(ICON_NAMES)}）")


def check_skill(zip_path: Path, errors: list[str]) -> None:
    files = zip_files(zip_path, "liki", errors)
    compare(f"{zip_path.name}", source_files(ROOT / "skills" / "liki"), files, errors)
    if "SKILL.md" not in files:
        errors.append(f"{zip_path.name}: 缺 SKILL.md")


def check_expert(zip_path: Path, name: str, errors: list[str]) -> None:
    files = zip_files(zip_path, name, errors)
    source = ROOT / "expert-packs" / name
    compare(f"{zip_path.name}", source_files(source, skip=frozenset({"SKILL.md"})), files, errors)
    if "SKILL.md" in files:
        errors.append(f"{zip_path.name}: 包根身份手写源 SKILL.md 不随包提交（平台入口是 agents/<pack>.md）")
    plugin_rel = ".codebuddy-plugin/plugin.json"
    if plugin_rel not in files:
        errors.append(f"{zip_path.name}: 缺 {plugin_rel}")
        return
    plugin = json.loads(files[plugin_rel])
    if plugin.get("name") != name:
        errors.append(f"{zip_path.name}: plugin.json.name `{plugin.get('name')}` 与顶层目录 `{name}` 不一致")
    if plugin.get("agents") != [f"./agents/{name}.md"]:
        errors.append(f"{zip_path.name}: plugin.json.agents 须指向 [\"./agents/{name}.md\"]")
    for required in (f"agents/{name}.md", "avatars/expert.png", "README.md", ".mcp.json"):
        if required not in files:
            errors.append(f"{zip_path.name}: 缺 {required}")


def main() -> int:
    if not OUT_DIR.is_dir():
        print("[check-workbuddy-zips] dist/workbuddy/ 不存在（未打包），跳过")
        return 0
    zips = sorted(OUT_DIR.glob("*.zip"))
    if not zips:
        print("[check-workbuddy-zips] dist/workbuddy/ 里没有 zip，跳过")
        return 0

    errors: list[str] = []
    checked = 0
    for zip_path in zips:
        name = zip_path.stem
        if name == "liki":
            check_skill(zip_path, errors)
        elif (ROOT / "workbuddy" / name).is_dir():
            check_connector(zip_path, name, errors)
        elif (ROOT / "expert-packs" / name).is_dir():
            check_expert(zip_path, name, errors)
        else:
            errors.append(f"{zip_path.name}: 找不到对应源码目录（不是连接器 / 技能 / 专家包）")
            continue
        checked += 1

    if errors:
        for line in errors:
            print(f"[check-workbuddy-zips] ✗ {line}", file=sys.stderr)
        return 1
    print(f"[check-workbuddy-zips] ✓ {checked} 个提交包与源码逐字节一致，结构符合平台解析要求")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
