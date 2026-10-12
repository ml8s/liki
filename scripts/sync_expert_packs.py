#!/usr/bin/env python3
"""Sync expert-pack methodology cards from the authoritative root skill (liki).

唯一权威源：skills/liki/references/（domains 真值卡 + app 编排卡）。
expert-packs/ 的方法论卡全部由此生成，不要手改包内卡。

    python3 scripts/sync_expert_packs.py            # 镜像生成（覆盖 + 清孤儿）
    python3 scripts/sync_expert_packs.py --check    # 双向漂移检测（make check / CI）

语义：每个目标目录（dest）聚合其全部源（多对一，如 naming ← qiming+bazi+app），
dest 的 *.md 必须与源并集逐名逐字节一致——缺失、多余（孤儿）、内容差异均判漂移。
"""
from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
SKILL = ROOT / "skills" / "liki" / "references"
PACKS = ROOT / "expert-packs"

# <根相对路径>:<包名>:<skill 目录>[:<子目录>]（domains 全集 + app 编排卡副本）
MAPPINGS: tuple[tuple[str, str, str, str | None], ...] = (
    ("natal/domains/bazi", "liki-bazi", "bazi", None),
    ("natal/domains/ziwei", "liki-ziwei", "ziwei", None),
    ("divination/domains/liuyao", "liki-liuyao", "liuyao", None),
    ("divination/domains/qimen", "liki-qimen", "qimen", None),
    ("divination/domains/huangli", "liki-qimen", "qimen", "huangli"),
    ("fengshui/domains/bazhai", "liki-fengshui", "fengshui", None),
    ("fengshui/domains/xuankong", "liki-fengshui", "fengshui", None),
    ("fengshui/domains/fengshui", "liki-fengshui", "fengshui", None),
    ("fengshui/app", "liki-fengshui", "fengshui", None),
    ("naming/domains/qiming", "liki-naming", "naming", None),
    ("naming/domains/bazi", "liki-naming", "naming", None),
    ("naming/app", "liki-naming", "naming", None),
)


# 包内卡片互引统一形如 references/<name>.md（ADK load_skill_resource 只认
# references/assets/scripts 前缀）；根 skill 的子树布局 token 由 sync rebase。
_REF_PATTERN = re.compile(r"references/[A-Za-z0-9_./-]+\.md")

# 专家 persona 单一源（agents/personas/<name>.md）→ 包内角色文件 <name>-expert.md。
# persona 正文＝部署 instruction 骨架（运营内容）；frontmatter 提供给包消费方。
PERSONAS_DIR = ROOT / "agents" / "personas"
PERSONA_PACKS: dict[str, str] = {
    "bazi": "liki-bazi",
    "ziwei": "liki-ziwei",
    "liuyao": "liki-liuyao",
    "qimen": "liki-qimen",
    "fengshui": "liki-fengshui",
    "naming": "liki-naming",
}

RUNTIME_VERSION_PATH = ROOT / "skills" / "liki" / "VERSION.txt"


def split_persona(path: Path) -> tuple[list[str], str] | list[str]:
    """解析 persona：返回 (frontmatter 行, 正文)；损坏时返回错误行列表。"""
    text = path.read_text(encoding="utf-8")
    if not text.startswith("---\n"):
        return [f"persona 缺少 frontmatter 起始标记: {path.relative_to(ROOT)}"]
    end = text.find("\n---\n")
    if end < 0:
        return [f"persona 缺少 frontmatter 结束标记: {path.relative_to(ROOT)}"]
    front = text[4:end].splitlines()
    body = text[end + len("\n---\n"):]
    if body.startswith("\n"):
        body = body[1:]
    return front, body


def persona_targets() -> tuple[dict[str, dict[str, bytes]], list[str]]:
    """persona 单一源 → {包: {角色文件名: 字节}}；错误累积返回。"""
    targets: dict[str, dict[str, bytes]] = {}
    errors: list[str] = []
    for name, pack in PERSONA_PACKS.items():
        src = PERSONAS_DIR / f"{name}.md"
        if not src.is_file():
            errors.append(f"persona 源缺失: {src.relative_to(ROOT)}")
            continue
        parsed = split_persona(src)
        if isinstance(parsed, list):
            errors.extend(parsed)
            continue
        front, body = parsed
        name_lines = [ln for ln in front if ln.startswith("name:")]
        if len(name_lines) != 1 or name_lines[0].strip() != f"name: {name}":
            errors.append(f"persona frontmatter name 与目录不符: {name}")
            continue
        renamed = [
            f"name: {name}-expert" if ln.startswith("name:") else ln for ln in front
        ]
        payload = "---\n" + "\n".join(renamed) + "\n---\n\n" + body
        payload = payload.rstrip() + "\n"
        targets.setdefault(pack, {})[f"{name}-expert.md"] = payload.encode("utf-8")
    return targets, errors


def _persona_metadata(name: str) -> tuple[dict, list[str]]:
    src = PERSONAS_DIR / f"{name}.md"
    parsed = split_persona(src)
    if isinstance(parsed, list):
        return {}, parsed
    front, _body = parsed
    try:
        metadata = yaml.safe_load("\n".join(front))
    except Exception as exc:
        return {}, [f"persona frontmatter YAML 无效: {src.relative_to(ROOT)}: {exc}"]
    if not isinstance(metadata, dict):
        return {}, [f"persona frontmatter必须是对象: {src.relative_to(ROOT)}"]
    return metadata, []


def plugin_metadata_updates() -> tuple[list[tuple[Path, dict, dict]], list[str]]:
    """Return plugin manifests and the persona-derived fields that must match."""
    runtime_version = RUNTIME_VERSION_PATH.read_text(encoding="utf-8").strip()
    updates: list[tuple[Path, dict, dict]] = []
    errors: list[str] = []
    for name, pack in PERSONA_PACKS.items():
        metadata, metadata_errors = _persona_metadata(name)
        errors.extend(metadata_errors)
        if metadata_errors:
            continue
        plugin_path = PACKS / pack / ".codebuddy-plugin" / "plugin.json"
        try:
            doc = json.loads(plugin_path.read_text(encoding="utf-8"))
        except Exception as exc:
            errors.append(f"plugin manifest读取失败: {plugin_path.relative_to(ROOT)}: {exc}")
            continue
        expected = {
            "description": metadata.get("description"),
            "profession": metadata.get("profession"),
            "version": runtime_version,
        }
        updates.append((plugin_path, doc, expected))
    return updates, errors


def _plugin_metadata_drift(
    updates: list[tuple[Path, dict, dict]]
) -> list[tuple[Path, dict, dict]]:
    drift: list[tuple[Path, dict, dict]] = []
    for path, doc, expected in updates:
        mismatched = {
            key: expected[key]
            for key, value in expected.items()
            if doc.get(key) != value
        }
        if mismatched:
            drift.append((path, doc, mismatched))
    return drift


def _source_names(srcs: list[Path]) -> set[str]:
    names: set[str] = set()
    for src in srcs:
        if not src.is_dir():
            continue
        for card in src.glob("*.md"):
            if card.name != "SKILL.md":
                names.add(card.name)
    return names


def _pack_name_maps(groups: dict[Path, list[Path]]) -> tuple[dict[str, dict[str, str]], list[str]]:
    """pack → {basename: 包内 references 路径}。同包同名异位 = 歧义，fail-closed。"""
    maps: dict[str, dict[str, str]] = {}
    errors: list[str] = []
    for dest, srcs in groups.items():
        pack = dest.relative_to(PACKS).parts[0]
        rel = dest.relative_to(PACKS).parts
        prefix = "references" if len(rel) <= 4 else f"references/{rel[4]}"
        name_map = maps.setdefault(pack, {})
        for name in _source_names(srcs):
            target = f"{prefix}/{name}"
            if name in name_map and name_map[name] != target:
                errors.append(f"包内同名卡歧义: {pack} {name} → {name_map[name]} vs {target}")
            else:
                name_map[name] = target
    return maps, errors


def _dest_groups() -> dict[Path, list[Path]]:
    groups: dict[Path, list[Path]] = {}
    for src_rel, pack, skill, sub in MAPPINGS:
        dest = PACKS / pack / "skills" / skill / "references"
        if sub:
            dest = dest / sub
        groups.setdefault(dest, []).append(SKILL / src_rel)
    return groups


def _rebase(payload: bytes, name_map: dict[str, str], errors: list[str], where: str) -> bytes:
    text = payload.decode("utf-8")

    def replace(match: re.Match[str]) -> str:
        token = match.group(0)
        target = name_map.get(Path(token).name)
        if target is None:
            errors.append(f"包内死链（该包不存在此卡）: {where}: {token}")
            return token
        return target

    return _REF_PATTERN.sub(replace, text).encode("utf-8")


def _expected(dest: Path, srcs: list[Path], name_map: dict[str, str]) -> dict[str, bytes] | list[str]:
    merged: dict[str, bytes] = {}
    errors: list[str] = []
    for src in srcs:
        if not src.is_dir():
            errors.append(f"源不存在: {src.relative_to(ROOT)}")
            continue
        for card in src.glob("*.md"):
            if card.name == "SKILL.md":
                continue
            payload = card.read_bytes()
            if card.name in merged and merged[card.name] != payload:
                errors.append(f"多源同名卡内容冲突: {card.name} ← {src.relative_to(ROOT)}")
            merged[card.name] = payload
    rebased: dict[str, bytes] = {}
    for name, payload in merged.items():
        rebased[name] = _rebase(payload, name_map, errors, f"{dest.relative_to(ROOT)}/{name}")
    return errors if errors else rebased


def sync() -> int:
    failed = False
    groups = _dest_groups()
    pack_maps, map_errors = _pack_name_maps(groups)
    for line in map_errors:
        print(f"[sync-expert-packs] ✗ {line}", file=sys.stderr)
    failed = bool(map_errors)
    for dest, srcs in groups.items():
        pack = dest.relative_to(PACKS).parts[0]
        result = _expected(dest, srcs, pack_maps.get(pack, {}))
        if isinstance(result, list):
            for line in result:
                print(f"[sync-expert-packs] ✗ {line}", file=sys.stderr)
            failed = True
            continue
        dest.mkdir(parents=True, exist_ok=True)
        for name, payload in result.items():
            target = dest / name
            if not target.exists() or target.read_bytes() != payload:
                target.write_bytes(payload)
        for orphan in dest.glob("*.md"):
            if orphan.name not in result:
                orphan.unlink()
        src_rel = ", ".join(str(s.relative_to(SKILL)) for s in srcs)
        print(f"[sync-expert-packs] ✓ {src_rel} → {dest.relative_to(ROOT)}")

    personas, persona_errors = persona_targets()
    for line in persona_errors:
        print(f"[sync-expert-packs] ✗ {line}", file=sys.stderr)
    failed = failed or bool(persona_errors)
    for pack, files in personas.items():
        agents_dir = PACKS / pack / "agents"
        agents_dir.mkdir(parents=True, exist_ok=True)
        for filename, payload in files.items():
            target = agents_dir / filename
            if not target.exists() or target.read_bytes() != payload:
                target.write_bytes(payload)
        for orphan in agents_dir.glob("*.md"):
            if orphan.name not in files:
                orphan.unlink()
        print(f"[sync-expert-packs] ✓ personas → {agents_dir.relative_to(ROOT)}")

    updates, plugin_errors = plugin_metadata_updates()
    for line in plugin_errors:
        print(f"[sync-expert-packs] ✗ {line}", file=sys.stderr)
    failed = failed or bool(plugin_errors)
    for path, doc, expected in updates:
        changed = any(doc.get(key) != value for key, value in expected.items())
        if changed:
            doc.update(expected)
            path.write_text(
                json.dumps(doc, ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
            print(f"[sync-expert-packs] ✓ plugin metadata → {path.relative_to(ROOT)}")

    if failed:
        return 1
    print("[sync-expert-packs] 完成。")
    return 0


def check() -> int:
    drift: list[str] = []
    groups = _dest_groups()
    pack_maps, map_errors = _pack_name_maps(groups)
    drift.extend(map_errors)
    for dest, srcs in groups.items():
        pack = dest.relative_to(PACKS).parts[0]
        result = _expected(dest, srcs, pack_maps.get(pack, {}))
        if isinstance(result, list):
            drift.extend(result)
            continue
        if not dest.is_dir():
            drift.append(f"目标不存在: {dest.relative_to(ROOT)}")
            continue
        actual = {c.name: c.read_bytes() for c in dest.glob("*.md")}
        for name in sorted(set(result) - set(actual)):
            drift.append(f"缺失: {dest.relative_to(ROOT)}/{name}")
        for name in sorted(set(actual) - set(result)):
            drift.append(f"多余（孤儿卡，双向漂移）: {dest.relative_to(ROOT)}/{name}")
        for name in sorted(set(result) & set(actual)):
            if result[name] != actual[name]:
                drift.append(f"内容漂移: {dest.relative_to(ROOT)}/{name}")

    personas, persona_errors = persona_targets()
    drift.extend(persona_errors)
    for pack, files in personas.items():
        agents_dir = PACKS / pack / "agents"
        if not agents_dir.is_dir():
            drift.append(f"目标不存在: {agents_dir.relative_to(ROOT)}")
            continue
        actual = {c.name: c.read_bytes() for c in agents_dir.glob("*.md")}
        for name in sorted(set(files) - set(actual)):
            drift.append(f"缺失: {agents_dir.relative_to(ROOT)}/{name}")
        for name in sorted(set(actual) - set(files)):
            drift.append(f"多余（孤儿卡，双向漂移）: {agents_dir.relative_to(ROOT)}/{name}")
        for name in sorted(set(files) & set(actual)):
            if files[name] != actual[name]:
                drift.append(f"内容漂移: {agents_dir.relative_to(ROOT)}/{name}")

    updates, plugin_errors = plugin_metadata_updates()
    drift.extend(plugin_errors)
    for path, _doc, expected in _plugin_metadata_drift(updates):
        for key, value in expected.items():
            drift.append(f"plugin metadata漂移: {path.relative_to(ROOT)}/{key}")

    if drift:
        print("[sync-expert-packs] ✗ expert-packs 与权威源存在漂移：", file=sys.stderr)
        for line in drift:
            print(f"  {line}", file=sys.stderr)
        print("运行 `make sync-expert-packs` 重新生成。", file=sys.stderr)
        return 1
    total = len(_dest_groups())
    print(f"[sync-expert-packs] ✓ 无漂移（{total} 目标 × 集合+内容双向一致）。")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser(description="Sync expert-pack cards from root skill")
    ap.add_argument("--check", action="store_true", help="只检测漂移，不写文件")
    args = ap.parse_args()
    return check() if args.check else sync()


if __name__ == "__main__":
    sys.exit(main())
